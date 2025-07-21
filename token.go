package quickbooks

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BearerToken struct {
	RefreshToken           string      `json:"refresh_token"`
	AccessToken            string      `json:"access_token"`
	TokenType              string      `json:"token_type"`
	IdToken                string      `json:"id_token"`
	ExpiresIn              json.Number `json:"expires_in"`
	XRefreshTokenExpiresIn json.Number `json:"x_refresh_token_expires_in"`
}

// RefreshToken
// Call the refresh endpoint to generate new tokens
func (c *Client) RefreshToken(refreshToken string) (*BearerToken, error) {
	urlValues := url.Values{}
	urlValues.Set("grant_type", "refresh_token")
	urlValues.Add("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", c.discoveryAPI.TokenEndpoint, bytes.NewBufferString(urlValues.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Authorization", "Basic "+basicAuth(c))

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var token BearerToken

	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &token, nil
}

// RetrieveBearerToken
// Method to retrieve access token (bearer token).
// This method can only be called once
func (c *Client) RetrieveBearerToken(authorizationCode, redirectURI string) (*BearerToken, error) {
	urlValues := url.Values{}
	// set parameters
	urlValues.Add("code", authorizationCode)
	urlValues.Set("grant_type", "authorization_code")
	urlValues.Add("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", c.discoveryAPI.TokenEndpoint, bytes.NewBufferString(urlValues.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Authorization", "Basic "+basicAuth(c))

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseFailure(resp)
	}

	var token BearerToken

	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &token, nil
}

// RevokeToken
// Call the revoke endpoint to revoke tokens
func (c *Client) RevokeToken(refreshToken string) error {
	urlValues := url.Values{}
	urlValues.Add("token", refreshToken)

	req, err := http.NewRequest("POST", c.discoveryAPI.RevocationEndpoint, bytes.NewBufferString(urlValues.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Authorization", "Basic "+basicAuth(c))

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	c.Client = nil

	return nil
}

func basicAuth(c *Client) string {
	return base64.StdEncoding.EncodeToString([]byte(c.clientId + ":" + c.clientSecret))
}
