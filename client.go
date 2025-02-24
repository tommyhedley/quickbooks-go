// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.
package quickbooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"sync"
	"time"
)

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
	Client           *http.Client
	baseEndpoint     *url.URL
	discoveryAPI     *DiscoveryAPI
	clientId         string
	clientSecret     string
	minorVersion     string
	throttled        bool
	rateLimiter      *RateLimiterManager
	globalConcurrent chan struct{}
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
		Client:       req.Client,
		discoveryAPI: req.DiscoveryAPI,
		clientId:     req.ClientId,
		clientSecret: req.ClientSecret,
		minorVersion: req.MinorVersion,
		throttled:    false,
		rateLimiter:  NewRateLimiterManager(),
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
	ctx     context.Context
	realmId string
	token   *BearerToken
}

func (c *Client) req(params RequestParameters, method string, endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	// First, acquire the global concurrency limiter.
	c.globalConcurrent <- struct{}{}
	defer func() { <-c.globalConcurrent }()

	// Retrieve the per-realm limiter.
	limiter := c.rateLimiter.getRealmLimiter(params.realmId)

	// Wait for a token from the general rate limiter.
	if err := limiter.general.Wait(params.ctx); err != nil {
		return fmt.Errorf("rate limiter error: %v", err)
	}

	// Acquire a slot from the realm-specific concurrent limiter.
	limiter.concurrent <- struct{}{}
	defer func() { <-limiter.concurrent }()

	// Build the full endpoint URL including realmId.
	endpointUrl := *c.baseEndpoint
	endpointUrl.Path += params.realmId + "/" + endpoint

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

	req, err := http.NewRequestWithContext(params.ctx, method, endpointUrl.String(), bytes.NewBuffer(marshalledJson))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+params.token.AccessToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Successful response.
	case http.StatusTooManyRequests:
		c.throttled = true
		go func() {
			time.Sleep(1 * time.Minute)
			c.throttled = false
		}()
		return errors.New("rate limit exceeded")
	default:
		return parseFailure(resp)
	}

	if responseObject != nil {
		if err = json.NewDecoder(resp.Body).Decode(&responseObject); err != nil {
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
	limiter := c.rateLimiter.getRealmLimiter(params.realmId)
	if err := limiter.batch.Wait(params.ctx); err != nil {
		return fmt.Errorf("batch rate limiter error: %v", err)
	}
	return c.post(params, "batch", payloadData, responseObject, nil)
}
