// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MFARequiredError represents an error indicating that MFA is required
type MFARequiredError struct {
	// The original error response from the API
	Response *ErrorResponse

	// The MFA challenge that needs to be satisfied
	Challenge *MFAChallenge
}

// Error returns the error message
func (e *MFARequiredError) Error() string {
	if e.Challenge != nil {
		return fmt.Sprintf("MFA required: %s (challenge ID: %s)",
			e.Response.ErrorDescription, e.Challenge.ChallengeID)
	}
	return fmt.Sprintf("MFA required: %s", e.Response.ErrorDescription)
}

// MFAChallenge represents an MFA challenge that needs to be satisfied
type MFAChallenge struct {
	// ChallengeID is the unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// Type indicates the type of MFA challenge (e.g., "totp", "webauthn", "backup_code")
	Type string `json:"type"`

	// Prompt is a human-readable prompt to display to the user
	Prompt string `json:"prompt"`

	// AllowedTypes contains all MFA types that can be used to satisfy this challenge
	AllowedTypes []string `json:"allowed_types"`

	// Additional information specific to the challenge type
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// MFAResponse represents a response to an MFA challenge
type MFAResponse struct {
	// ChallengeID is the unique identifier for the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// Type is the type of MFA being used to respond (e.g., "totp", "webauthn", "backup_code")
	Type string `json:"type"`

	// Value is the actual MFA code or response value
	Value string `json:"value"`
}

// IsMFAError checks if an error is an MFA required error
func IsMFAError(err error) bool {
	_, ok := err.(*MFARequiredError)
	return ok || (err != nil && strings.Contains(err.Error(), "MFA required"))
}

// GetMFAChallenge extracts the MFA challenge from an error
func GetMFAChallenge(err error) *MFAChallenge {
	if mfaErr, ok := err.(*MFARequiredError); ok {
		return mfaErr.Challenge
	}
	return nil
}

// CheckForMFARequired checks if a token response error indicates MFA is required
// and extracts the MFA challenge if present
func (c *Client) CheckForMFARequired(resp *http.Response) (*MFARequiredError, error) {
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read error response: %w", err)
	}

	// Try to parse as an error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return nil, fmt.Errorf("failed to parse error response: %w", err)
	}

	// Check if this is an MFA required error
	if errorResp.Error == "mfa_required" ||
		(errorResp.Error == "invalid_grant" &&
			strings.Contains(errorResp.ErrorDescription, "MFA")) {

		// Extract the challenge ID from the error description
		challengeID := extractChallengeID(errorResp.ErrorDescription)
		if challengeID == "" {
			return &MFARequiredError{
				Response: &errorResp,
			}, nil
		}

		// Get the MFA challenge details
		challenge, err := c.GetMFAChallenge(context.Background(), challengeID)
		if err != nil {
			return &MFARequiredError{
				Response: &errorResp,
			}, nil
		}

		return &MFARequiredError{
			Response:  &errorResp,
			Challenge: challenge,
		}, nil
	}

	// Not an MFA error
	return nil, fmt.Errorf("%s: %s", errorResp.Error, errorResp.ErrorDescription)
}

// extractChallengeID extracts the challenge ID from an error description
func extractChallengeID(description string) string {
	// Look for patterns like "challenge ID: abc123" in the error description
	prefix := "challenge ID: "
	if idx := strings.Index(description, prefix); idx >= 0 {
		// Get everything after the prefix
		suffix := description[idx+len(prefix):]

		// If there's a closing parenthesis, strip it
		if closingIdx := strings.Index(suffix, ")"); closingIdx >= 0 {
			return strings.TrimSpace(suffix[:closingIdx])
		}

		return strings.TrimSpace(suffix)
	}
	return ""
}

// GetMFAChallenge gets details about an MFA challenge
func (c *Client) GetMFAChallenge(ctx context.Context, challengeID string) (*MFAChallenge, error) {
	// Create the request URL
	reqURL := fmt.Sprintf("%soauth2/mfa/challenge/%s", c.Client.BaseURL, challengeID)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA challenge request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MFA challenge request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MFA challenge request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var challenge MFAChallenge
	if err := json.NewDecoder(resp.Body).Decode(&challenge); err != nil {
		return nil, fmt.Errorf("failed to parse MFA challenge response: %w", err)
	}

	return &challenge, nil
}

// RespondToMFAChallenge sends a response to an MFA challenge
func (c *Client) RespondToMFAChallenge(ctx context.Context, response *MFAResponse) (*TokenResponse, error) {
	// Create the request URL
	reqURL := fmt.Sprintf("%soauth2/mfa/response", c.Client.BaseURL)

	// Create the request body
	reqBody, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MFA response: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA response request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MFA response request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		// Check if this is another MFA challenge
		if resp.StatusCode == http.StatusBadRequest {
			mfaErr, parseErr := c.CheckForMFARequired(resp)
			if parseErr == nil && mfaErr != nil {
				return nil, mfaErr
			}
		}

		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MFA response request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response as a token response
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResponse, nil
}

// ExchangeAuthorizationCodeWithMFA exchanges an authorization code with MFA support
func (c *Client) ExchangeAuthorizationCodeWithMFA(
	ctx context.Context,
	code string,
	mfaHandler func(challenge *MFAChallenge) (*MFAResponse, error),
) (*TokenResponse, error) {
	if c.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required for code exchange")
	}

	// Build the request body
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURL)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	// Use MFA-enabled token request
	tokenResp, err := c.tokenRequestMFA(ctx, form)
	if err != nil {
		// Check if this is an MFA error
		if mfaErr, ok := err.(*MFARequiredError); ok && mfaHandler != nil {
			// Call the handler to get the MFA response
			mfaResponse, handlerErr := mfaHandler(mfaErr.Challenge)
			if handlerErr != nil {
				return nil, fmt.Errorf("MFA handler failed: %w", handlerErr)
			}

			// Send the MFA response
			return c.RespondToMFAChallenge(ctx, mfaResponse)
		}
		return nil, err
	}

	return tokenResp, nil
}

// RefreshTokenWithMFA refreshes a token with MFA support
func (c *Client) RefreshTokenWithMFA(
	ctx context.Context,
	refreshToken string,
	mfaHandler func(challenge *MFAChallenge) (*MFAResponse, error),
) (*TokenResponse, error) {
	// Build the request body
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	// Use MFA-enabled token request
	tokenResp, err := c.tokenRequestMFA(ctx, form)
	if err != nil {
		// Check if this is an MFA error
		if mfaErr, ok := err.(*MFARequiredError); ok && mfaHandler != nil {
			// Call the handler to get the MFA response
			mfaResponse, handlerErr := mfaHandler(mfaErr.Challenge)
			if handlerErr != nil {
				return nil, fmt.Errorf("MFA handler failed: %w", handlerErr)
			}

			// Send the MFA response
			return c.RespondToMFAChallenge(ctx, mfaResponse)
		}
		return nil, err
	}

	return tokenResp, nil
}

// tokenRequestMFA is a version of tokenRequest that supports MFA challenges
func (c *Client) tokenRequestMFA(ctx context.Context, form url.Values) (*TokenResponse, error) {
	// Set the headers
	headers := http.Header{}
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create the request body
	body := strings.NewReader(form.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL+"oauth2/token", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		// Check if this is an MFA required error
		if resp.StatusCode == http.StatusBadRequest {
			mfaErr, parseErr := c.CheckForMFARequired(resp)
			if parseErr == nil && mfaErr != nil {
				return nil, mfaErr
			}
		}

		// Not an MFA error, or error parsing the MFA error
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResponse, nil
}
