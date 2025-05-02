// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestValidateToken(t *testing.T) {
	// Test with valid token
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Return mock response for a valid token
		response := TokenInfo{
			Active:      true,
			Scope:       "openid profile email",
			ClientID:    "test-client-id",
			Username:    "test-user",
			Exp:         time.Now().Add(time.Hour).Unix(),
			Subject:     "test-subject",
			SubjectType: "user",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test validation of valid token
	err := client.ValidateToken(context.Background(), "valid-token")
	if err != nil {
		t.Errorf("ValidateToken() with valid token should not return error, got: %v", err)
	}

	// Test with inactive token
	handler = func(w http.ResponseWriter, r *http.Request) {
		// Return mock response for an inactive token
		response := TokenInfo{
			Active:   false,
			Scope:    "openid profile email",
			ClientID: "test-client-id",
			Exp:      time.Now().Add(time.Hour).Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	// Test validation of inactive token
	err = client.ValidateToken(context.Background(), "inactive-token")
	if err != ErrTokenInvalid {
		t.Errorf("ValidateToken() with inactive token should return ErrTokenInvalid, got: %v", err)
	}

	// Test with expired token
	handler = func(w http.ResponseWriter, r *http.Request) {
		// Return mock response for an expired token
		response := TokenInfo{
			Active:   true, // API might still return active=true for recently expired tokens
			Scope:    "openid profile email",
			ClientID: "test-client-id",
			Exp:      time.Now().Add(-time.Hour).Unix(), // Expired an hour ago
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	// Test validation of expired token
	err = client.ValidateToken(context.Background(), "expired-token")
	if err != ErrTokenExpired {
		t.Errorf("ValidateToken() with expired token should return ErrTokenExpired, got: %v", err)
	}
}

func TestGetRemainingValidity(t *testing.T) {
	// One hour from now
	expiryTime := time.Now().Add(time.Hour).Unix()

	// Test with valid token
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Return mock response for a valid token
		response := TokenInfo{
			Active:      true,
			Scope:       "openid profile email",
			ClientID:    "test-client-id",
			Username:    "test-user",
			Exp:         expiryTime,
			Subject:     "test-subject",
			SubjectType: "user",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test getting remaining validity for valid token
	remaining, err := client.GetRemainingValidity(context.Background(), "valid-token")
	if err != nil {
		t.Errorf("GetRemainingValidity() with valid token should not return error, got: %v", err)
	}

	// Should be close to one hour, with some tolerance for test execution time
	expectedMin := time.Minute * 59 // Just under an hour
	if remaining < expectedMin {
		t.Errorf("GetRemainingValidity() remaining time = %v, want at least %v", remaining, expectedMin)
	}
	expectedMax := time.Hour + time.Second*5 // Slightly over an hour with tolerance
	if remaining > expectedMax {
		t.Errorf("GetRemainingValidity() remaining time = %v, want at most %v", remaining, expectedMax)
	}

	// Test with expired token
	handler = func(w http.ResponseWriter, r *http.Request) {
		// Return mock response for an expired token
		response := TokenInfo{
			Active:   true,
			Scope:    "openid profile email",
			ClientID: "test-client-id",
			Exp:      time.Now().Add(-time.Hour).Unix(), // Expired an hour ago
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	// Test getting remaining validity for expired token
	remaining, err = client.GetRemainingValidity(context.Background(), "expired-token")
	if err != nil {
		t.Errorf("GetRemainingValidity() with expired token should not return error, got: %v", err)
	}
	if remaining != 0 {
		t.Errorf("GetRemainingValidity() with expired token should return 0, got: %v", remaining)
	}
}