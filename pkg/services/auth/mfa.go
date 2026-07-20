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

		// Extract the challenge ID from the error description. Globus Auth at
		// 3.65.0 has no /oauth2/mfa/challenge route to fetch further details;
		// the challenge ID from the token-endpoint error is all that's exposed.
		challengeID := extractChallengeID(errorResp.ErrorDescription)
		if challengeID == "" {
			return &MFARequiredError{Response: &errorResp}, nil
		}
		return &MFARequiredError{
			Response:  &errorResp,
			Challenge: &MFAChallenge{ChallengeID: challengeID},
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

// resubmitWithMFA answers an MFA challenge by resubmitting the original token
// request to the token endpoint with the MFA response fields attached. Globus
// Auth at 3.65.0 has no separate /oauth2/mfa/* routes; MFA is completed by
// re-POSTing to /v2/oauth2/token.
func (c *Client) resubmitWithMFA(ctx context.Context, form url.Values, resp *MFAResponse) (*TokenResponse, error) {
	if resp != nil {
		if resp.ChallengeID != "" {
			form.Set("mfa_challenge_id", resp.ChallengeID)
		}
		if resp.Type != "" {
			form.Set("mfa_type", resp.Type)
		}
		if resp.Value != "" {
			form.Set("mfa_value", resp.Value)
		}
	}
	return c.tokenRequestMFA(ctx, form)
}

// ExchangeAuthorizationCodeWithMFA exchanges an authorization code, completing an
// MFA challenge by resubmitting to the token endpoint when required.
func (c *Client) ExchangeAuthorizationCodeWithMFA(
	ctx context.Context,
	code string,
	mfaHandler func(challenge *MFAChallenge) (*MFAResponse, error),
) (*TokenResponse, error) {
	if c.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required for code exchange")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURL)
	form.Set("client_id", c.ClientID)
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	tokenResp, err := c.tokenRequestMFA(ctx, form)
	if err != nil {
		if mfaErr, ok := err.(*MFARequiredError); ok && mfaHandler != nil {
			mfaResponse, handlerErr := mfaHandler(mfaErr.Challenge)
			if handlerErr != nil {
				return nil, fmt.Errorf("MFA handler failed: %w", handlerErr)
			}
			return c.resubmitWithMFA(ctx, form, mfaResponse)
		}
		return nil, err
	}

	return tokenResp, nil
}

// RefreshTokenWithMFA refreshes a token, completing an MFA challenge by
// resubmitting to the token endpoint when required.
func (c *Client) RefreshTokenWithMFA(
	ctx context.Context,
	refreshToken string,
	mfaHandler func(challenge *MFAChallenge) (*MFAResponse, error),
) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	tokenResp, err := c.tokenRequestMFA(ctx, form)
	if err != nil {
		if mfaErr, ok := err.(*MFARequiredError); ok && mfaHandler != nil {
			mfaResponse, handlerErr := mfaHandler(mfaErr.Challenge)
			if handlerErr != nil {
				return nil, fmt.Errorf("MFA handler failed: %w", handlerErr)
			}
			return c.resubmitWithMFA(ctx, form, mfaResponse)
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
	defer func() { _ = resp.Body.Close() }()

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
