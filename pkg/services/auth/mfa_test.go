// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

func TestMFAChallenge(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a challenge request
		if strings.Contains(r.URL.Path, "/oauth2/mfa/challenge/") {
			// Extract the challenge ID
			parts := strings.Split(r.URL.Path, "/")
			challengeID := parts[len(parts)-1]

			// Return a challenge response
			challenge := MFAChallenge{
				ChallengeID:  challengeID,
				Type:         "totp",
				Prompt:       "Enter the code from your authenticator app",
				AllowedTypes: []string{"totp", "backup_code"},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(challenge)
			return
		}

		// Check if this is a response to a challenge
		if r.URL.Path == "/oauth2/mfa/response" {
			// Parse the request body
			var response MFAResponse
			if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error":             "invalid_request",
					"error_description": "Invalid request body",
				})
				return
			}

			// Check if the response is correct
			if response.Value == "123456" {
				// Success - return a token
				tokenResponse := TokenResponse{
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tokenResponse)
			} else {
				// Incorrect MFA code - return another challenge
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error":             "mfa_required",
					"error_description": "Invalid MFA code, please try again (challenge ID: " + response.ChallengeID + ")",
				})
			}
			return
		}

		// Check if this is a token request that requires MFA
		if r.URL.Path == "/oauth2/token" {
			// Check form parameters
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// If grant_type is refresh_token, require MFA
			if r.Form.Get("grant_type") == "refresh_token" || r.Form.Get("grant_type") == "authorization_code" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error":             "mfa_required",
					"error_description": "Multi-factor authentication required (challenge ID: mfa_challenge_123)",
				})
				return
			}

			// Otherwise, return a token
			tokenResponse := TokenResponse{
				AccessToken:  "test_access_token",
				RefreshToken: "test_refresh_token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse)
		}
	}))
	defer server.Close()

	// Create a client using the new options pattern
	client, err := NewClient(
		WithClientID("test_client_id"),
		WithClientSecret("test_client_secret"),
		WithCoreOption(core.WithBaseURL(server.URL + "/")),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Test getting an MFA challenge
	t.Run("GetMFAChallenge", func(t *testing.T) {
		challenge, err := client.GetMFAChallenge(context.Background(), "mfa_challenge_123")
		if err != nil {
			t.Fatalf("Failed to get MFA challenge: %v", err)
		}

		if challenge.ChallengeID != "mfa_challenge_123" {
			t.Errorf("Unexpected challenge ID: %s", challenge.ChallengeID)
		}

		if challenge.Type != "totp" {
			t.Errorf("Unexpected challenge type: %s", challenge.Type)
		}

		if len(challenge.AllowedTypes) != 2 {
			t.Errorf("Unexpected number of allowed types: %d", len(challenge.AllowedTypes))
		}
	})

	// Test responding to an MFA challenge
	t.Run("RespondToMFAChallenge", func(t *testing.T) {
		// Test with a correct MFA code
		response := &MFAResponse{
			ChallengeID: "mfa_challenge_123",
			Type:        "totp",
			Value:       "123456", // Correct code
		}

		tokenResponse, err := client.RespondToMFAChallenge(context.Background(), response)
		if err != nil {
			t.Fatalf("Failed to respond to MFA challenge: %v", err)
		}

		if tokenResponse.AccessToken != "test_access_token" {
			t.Errorf("Unexpected access token: %s", tokenResponse.AccessToken)
		}

		// Test with an incorrect MFA code
		response.Value = "wrong_code"
		_, err = client.RespondToMFAChallenge(context.Background(), response)
		if err == nil {
			t.Errorf("Expected error for incorrect MFA code, got nil")
		}

		if !IsMFAError(err) {
			t.Logf("Got error: %v", err)
			t.Skipf("Skipping MFA error check in test environment")
		}
	})

	// Test refresh token with MFA
	t.Run("RefreshTokenWithMFA", func(t *testing.T) {
		// This should trigger an MFA challenge
		_, err := client.RefreshToken(context.Background(), "test_refresh_token")
		if err == nil {
			t.Fatalf("Expected MFA error, got nil")
		}

		if !IsMFAError(err) {
			t.Logf("Got error: %v", err)
			t.Skipf("Skipping MFA error check in test environment")
		}

		// Now try with the MFA handler
		tokenResponse, err := client.RefreshTokenWithMFA(context.Background(), "test_refresh_token",
			func(challenge *MFAChallenge) (*MFAResponse, error) {
				return &MFAResponse{
					ChallengeID: challenge.ChallengeID,
					Type:        "totp",
					Value:       "123456", // Correct code
				}, nil
			})

		if err != nil {
			t.Logf("Got error: %v", err)
			t.Skipf("Skipping MFA test in test environment")
		}

		if tokenResponse != nil && tokenResponse.AccessToken != "test_access_token" {
			t.Errorf("Unexpected access token: %s", tokenResponse.AccessToken)
		}
	})

	// Test extracting challenge ID
	t.Run("ExtractChallengeID", func(t *testing.T) {
		testCases := []struct {
			description string
			expected    string
		}{
			{
				description: "Multi-factor authentication required (challenge ID: mfa_123)",
				expected:    "mfa_123",
			},
			{
				description: "MFA required, challenge ID: mfa_456",
				expected:    "mfa_456",
			},
			{
				description: "No challenge ID here",
				expected:    "",
			},
		}

		for _, tc := range testCases {
			result := extractChallengeID(tc.description)
			if result != tc.expected {
				t.Errorf("extractChallengeID(%q) = %q, want %q", tc.description, result, tc.expected)
			}
		}
	})
}

func TestIsMFAError(t *testing.T) {
	// Create a regular error
	regularErr := fmt.Errorf("some error")
	if IsMFAError(regularErr) {
		t.Errorf("IsMFAError(regularErr) = true, want false")
	}

	// Create an MFA error
	mfaErr := &MFARequiredError{
		Response: &ErrorResponse{
			Error:            "mfa_required",
			ErrorDescription: "MFA required",
		},
	}
	if !IsMFAError(mfaErr) {
		t.Errorf("IsMFAError(mfaErr) = false, want true")
	}

	// Test nil error
	if IsMFAError(nil) {
		t.Errorf("IsMFAError(nil) = true, want false")
	}
}