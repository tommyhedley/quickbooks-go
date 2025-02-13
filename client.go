// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.
package quickbooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is your handle to the QuickBooks API.
type Client struct {
	// Get this from oauth2.NewClient().
	Client *http.Client
	// Set to ProductionEndpoint or SandboxEndpoint.
	endpoint *url.URL
	// The set of quickbooks APIs
	discoveryAPI *DiscoveryAPI
	// The client Id
	clientId string
	// The client Secret
	clientSecret string
	// The minor version of the QB API
	minorVersion string
	// The account Id you're connecting to.
	realmId string
	// Flag set if the limit of 500req/s has been hit (source: https://developer.intuit.com/app/developer/qbo/docs/learn/rest-api-features#limits-and-throttles)
	throttled bool
}

type ClientRequest struct {
	DiscoveryAPI *DiscoveryAPI
	ClientId     string
	ClientSecret string
	RealmId      string
	Endpoint     EndpointUrl
	MinorVersion string
	Token        *BearerToken
}

// NewClient initializes a new QuickBooks client for interacting with their Online API
func NewClient(req ClientRequest) (c *Client, err error) {
	if req.MinorVersion == "" {
		req.MinorVersion = "75"
	}

	client := Client{
		discoveryAPI: req.DiscoveryAPI,
		clientId:     req.ClientId,
		clientSecret: req.ClientSecret,
		minorVersion: req.MinorVersion,
		realmId:      req.RealmId,
		throttled:    false,
	}

	client.endpoint, err = url.Parse(req.Endpoint.String() + "/v3/company/" + req.RealmId + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API endpoint: %v", err)
	}

	if req.Token != nil {
		client.Client = getHttpClient(req.Token)
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

func (c *Client) req(method string, endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	// TODO: possibly just wait until c.throttled is false, and continue the request?
	if c.throttled {
		return errors.New("waiting for rate limit")
	}

	endpointUrl := *c.endpoint
	endpointUrl.Path += endpoint
	urlValues := url.Values{}

	if len(queryParameters) > 0 {
		for param, value := range queryParameters {
			urlValues.Add(param, value)
		}
	}

	urlValues.Set("minorversion", c.minorVersion)
	urlValues.Encode()
	endpointUrl.RawQuery = urlValues.Encode()

	var err error
	var marshalledJson []byte

	if payloadData != nil {
		marshalledJson, err = json.Marshal(payloadData)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	req, err := http.NewRequest(method, endpointUrl.String(), bytes.NewBuffer(marshalledJson))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusTooManyRequests:
		c.throttled = true
		go func(c *Client) {
			time.Sleep(1 * time.Minute)
			c.throttled = false
		}(c)
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

func (c *Client) get(endpoint string, responseObject interface{}, queryParameters map[string]string) error {
	return c.req("GET", endpoint, nil, responseObject, queryParameters)
}

func (c *Client) post(endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	return c.req("POST", endpoint, payloadData, responseObject, queryParameters)
}

// query makes the specified QBO `query` and unmarshals the result into `responseObject`
func (c *Client) query(query string, responseObject interface{}) error {
	return c.get("query", responseObject, map[string]string{"query": query})
}
