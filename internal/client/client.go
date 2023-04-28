package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.encore.dev/platform-sdk/pkg/auth"
)

// Client is the underlying raw client for communicating with the Encore Platform services.
//
// It is injected into each service struct by the main [platform] package.
type Client struct {
	cfg *Config
}

func New(cfg *Config) *Client {
	return &Client{cfg}
}

// SignedPost performs a signed POST request to the specified path.
func (c *Client) SignedPost(ctx context.Context, path string, object auth.ObjectType, action auth.ActionType, body auth.Payload, response any, additionalAuthContext ...[]byte) error {
	// Hash the request
	opHash, err := auth.NewOperationHash(
		object, action, body, additionalAuthContext...,
	)
	if err != nil {
		return fmt.Errorf("failed to hash request: %w", err)
	}

	// Sign the hash
	headers := auth.Sign(&c.cfg.LatestAuthKey, c.cfg.AppSlug, c.cfg.EnvName, c.cfg.Clock, opHash)

	// Create the request
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, fmt.Sprintf("%s%s", c.cfg.Host, path), bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("User-Agent", "Encore-Platform-SDK")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", headers.Authorization)
	req.Header.Set("Date", headers.Date)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status %s", resp.Status)
	}

	// Decode the response
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// VerifyAndDecodeRequest verifies the authenticity of the request and decodes the request into body.
func (c *Client) VerifyAndDecodeRequest(req *http.Request, object auth.ObjectType, action auth.ActionType, body auth.Payload, additionalAuthContext ...[]byte) error {
	opHash, err := auth.GetVerifiedOperationHash(req, c.cfg.AuthKeys, c.cfg.Clock)
	if err != nil {
		return fmt.Errorf("unable to verify operation hash: %w", err)
	}

	// Body bytes
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", err)
	}

	// Decode the payload
	if err := json.Unmarshal(bodyBytes, body); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	// Verify the operation hash is correct
	ok, err := opHash.Verify(
		object, action, body,
		additionalAuthContext...,
	)
	if err != nil {
		return fmt.Errorf("unable to verify operation hash: %w", err)
	}
	if !ok {
		return auth.ErrAuthenticationFailed
	}

	return nil
}
