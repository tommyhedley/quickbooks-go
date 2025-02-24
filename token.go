package quickbooks

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type BearerToken struct {
	RefreshToken           string      `json:"refresh_token"`
	AccessToken            string      `json:"access_token"`
	TokenType              string      `json:"token_type"`
	IdToken                string      `json:"id_token"`
	ExpiresIn              json.Number `json:"expires_in"`
	ExpiresOn              time.Time   `json:"expires_on,omitempty"`
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

	bearerTokenResponse, err := getBearerTokenResponse(body)

	return bearerTokenResponse, err
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

	bearerTokenResponse, err := getBearerTokenResponse(body)

	return bearerTokenResponse, err
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

// CheckExpiration
// Check if the tokens ExpiresOn value is within the next s seconds
func (b *BearerToken) CheckExpiration(s int) bool {
	expirationCutoff := time.Now().Add(time.Duration(s) * time.Second)
	return b.ExpiresOn.Before(expirationCutoff)
}

func basicAuth(c *Client) string {
	return base64.StdEncoding.EncodeToString([]byte(c.clientId + ":" + c.clientSecret))
}

func getBearerTokenResponse(body []byte) (*BearerToken, error) {
	token := BearerToken{}

	if err := json.Unmarshal(body, &token); err != nil {
		return nil, errors.New(string(body))
	}

	expiresIn, err := token.ExpiresIn.Int64()
	if err != nil {
		return nil, errors.New("Unable to convert expires_in to int64")
	}

	token.ExpiresOn = time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)

	return &token, nil
}
