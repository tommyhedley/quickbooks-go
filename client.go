// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.
package quickbooks

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/time/rate"
)

type rateLimitType struct {
	Name string
	Rate string
}

var (
	apiRl = rateLimitType{
		Name: "extenal api",
		Rate: "",
	}
	realmGeneralRL = rateLimitType{
		Name: "internal realm general",
		Rate: "500 req/min, burst to 10 req/sec",
	}
	realmConcurrentRL = rateLimitType{
		Name: "internal realm concurrent",
		Rate: "10 req/sec",
	}
	realmBatchRL = rateLimitType{
		Name: "internal realm batch",
		Rate: "40 req/min",
	}
	globalGeneralRL = rateLimitType{
		Name: "internal global general",
		Rate: "500 req/min, burst to 10 req/sec",
	}
	globalConcurrentRL = rateLimitType{
		Name: "internal global concurrent",
		Rate: "10 req/sec",
	}
)

type RateLimitError struct {
	Message   string
	LimitType rateLimitType
}

func (e *RateLimitError) Error() string {
	return e.Message
}

func NewRateLimitError(limitType rateLimitType) *RateLimitError {
	var message string
	if limitType.Rate == "" {
		message = fmt.Sprintf("%s rate limit exceeded", limitType.Name)
	} else {
		message = fmt.Sprintf("%s rate limit exceeded: %s", limitType.Name, limitType.Rate)
	}
	return &RateLimitError{
		Message:   message,
		LimitType: limitType,
	}
}

type RealmRateLimiters struct {
	// General limiter: 500 req/min = ~8.33 req/sec with a burst of 10.
	general *rate.Limiter
	// Semaphore limiting concurrent requests to 10.
	concurrent chan struct{}
	// Batch limiter: 40 batch req/min = ~0.67 req/sec with a burst of 5.
	batch *rate.Limiter
}

// RateLimiterManager manages rate limiters per realm.
type RateLimiterManager struct {
	mu       sync.Mutex
	limiters map[string]*RealmRateLimiters
}

// NewRateLimiterManager initializes a new RateLimiterManager.
func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]*RealmRateLimiters),
	}
}

// getRealmLimiter returns (or creates) the rate limiters for a given realm.
func (m *RateLimiterManager) getRealmLimiter(realmId string) *RealmRateLimiters {
	m.mu.Lock()
	defer m.mu.Unlock()
	if limiter, exists := m.limiters[realmId]; exists {
		return limiter
	}
	// Create a new set of limiters.
	limiter := &RealmRateLimiters{
		general:    rate.NewLimiter(rate.Limit(500.0/60.0), 10),
		concurrent: make(chan struct{}, 10),
		batch:      rate.NewLimiter(rate.Limit(40.0/60.0), 5),
	}
	m.limiters[realmId] = limiter
	return limiter
}

// Client is your handle to the QuickBooks API.
type Client struct {
	Client            *http.Client
	baseEndpoint      *url.URL
	discoveryAPI      *DiscoveryAPI
	clientId          string
	clientSecret      string
	minorVersion      string
	rateLimiter       *RateLimiterManager
	globalConcurrent  chan struct{}
	globalRateLimiter *rate.Limiter
}

type ClientRequest struct {
	Client       *http.Client
	DiscoveryAPI *DiscoveryAPI
	ClientId     string
	ClientSecret string
	Endpoint     string
	MinorVersion string
}

// NewClient initializes a new QuickBooks client for interacting with their Online API
func NewClient(req ClientRequest) (c *Client, err error) {
	if req.MinorVersion == "" {
		req.MinorVersion = "75"
	}

	client := Client{
		Client:            req.Client,
		discoveryAPI:      req.DiscoveryAPI,
		clientId:          req.ClientId,
		clientSecret:      req.ClientSecret,
		minorVersion:      req.MinorVersion,
		rateLimiter:       NewRateLimiterManager(),
		globalConcurrent:  make(chan struct{}, 10),
		globalRateLimiter: rate.NewLimiter(rate.Limit(500.0/60.0), 10),
	}

	client.baseEndpoint, err = url.Parse(req.Endpoint + "/v3/company/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API endpoint: %v", err)
	}

	return &client, nil
}

// FindAuthorizationUrl compiles the authorization url from the discovery api's auth endpoint.
//
// Example: qbClient.FindAuthorizationUrl("com.intuit.quickbooks.accounting", "security_token", "https://developer.intuit.com/v2/OAuth2Playground/RedirectUrl")
//
// You can find live examples from https://developer.intuit.com/app/developer/playground
func (c *Client) FindAuthorizationUrl(scope string, state string, redirectUri string) (string, error) {
	var authorizationUrl *url.URL

	authorizationUrl, err := url.Parse(c.discoveryAPI.AuthorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse auth endpoint: %v", err)
	}

	urlValues := url.Values{}
	urlValues.Add("client_id", c.clientId)
	urlValues.Add("response_type", "code")
	urlValues.Add("scope", scope)
	urlValues.Add("redirect_uri", redirectUri)
	urlValues.Add("state", state)
	authorizationUrl.RawQuery = urlValues.Encode()

	return authorizationUrl.String(), nil
}

type RequestParameters struct {
	Ctx     context.Context
	RealmId string
	Token   *BearerToken
}

func (c *Client) req(params RequestParameters, method string, endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	// Attempt to acquire the global concurrency slot non-blocking.
	select {
	case c.globalConcurrent <- struct{}{}:
		defer func() { <-c.globalConcurrent }()
	default:
		return NewRateLimitError(globalConcurrentRL)
	}

	// Check global rate limiter non-blocking.
	if !c.globalRateLimiter.Allow() {
		return NewRateLimitError(globalGeneralRL)
	}

	// Retrieve the per-realm limiter.
	limiter := c.rateLimiter.getRealmLimiter(params.RealmId)

	// Check realm-specific rate limiter non-blocking.
	if !limiter.general.Allow() {
		return NewRateLimitError(realmGeneralRL)
	}

	// Attempt to acquire the global concurrency slot non-blocking.
	select {
	case limiter.concurrent <- struct{}{}:
		defer func() { <-limiter.concurrent }()
	default:
		return NewRateLimitError(realmConcurrentRL)
	}

	// Build the full endpoint URL including realmId.
	endpointUrl := *c.baseEndpoint
	endpointUrl.Path += params.RealmId + "/" + endpoint

	// Build query parameters.
	urlValues := url.Values{}
	for param, value := range queryParameters {
		urlValues.Add(param, value)
	}
	urlValues.Set("minorversion", c.minorVersion)
	endpointUrl.RawQuery = urlValues.Encode()

	var marshalledJson []byte
	if payloadData != nil {
		var err error
		marshalledJson, err = json.Marshal(payloadData)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(params.Ctx, method, endpointUrl.String(), bytes.NewBuffer(marshalledJson))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+params.Token.AccessToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Successful response.
	case http.StatusTooManyRequests:
		return NewRateLimitError(apiRl)
	default:
		return parseFailure(resp)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	if responseObject != nil {
		if err = json.NewDecoder(reader).Decode(&responseObject); err != nil {
			return fmt.Errorf("failed to unmarshal response into object: %v", err)
		}
	}

	return nil
}

func (c *Client) get(params RequestParameters, endpoint string, responseObject interface{}, queryParameters map[string]string) error {
	return c.req(params, "GET", endpoint, nil, responseObject, queryParameters)
}

func (c *Client) post(params RequestParameters, endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	return c.req(params, "POST", endpoint, payloadData, responseObject, queryParameters)
}

// query makes the specified QBO query and unmarshals the result into responseObject.
func (c *Client) query(params RequestParameters, query string, responseObject interface{}) error {
	return c.get(params, "query", responseObject, map[string]string{"query": query})
}

// batch handles batch requests. It waits on the batch limiter before sending.
func (c *Client) batch(params RequestParameters, payloadData interface{}, responseObject interface{}) error {
	limiter := c.rateLimiter.getRealmLimiter(params.RealmId)
	if err := limiter.batch.Wait(params.Ctx); err != nil {
		return fmt.Errorf("batch rate limiter error: %v", err)
	}
	return c.post(params, "batch", payloadData, responseObject, nil)
}
