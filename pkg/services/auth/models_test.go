// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTokenResponse_HasRefreshToken(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		want         bool
	}{
		{
			name:         "With refresh token",
			refreshToken: "test-refresh-token",
			want:         true,
		},
		{
			name:         "Without refresh token",
			refreshToken: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TokenResponse{
				RefreshToken: tt.refreshToken,
			}
			if got := tr.HasRefreshToken(); got != tt.want {
				t.Errorf("TokenResponse.HasRefreshToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenResponse_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		expiryTime time.Time
		want       bool
	}{
		{
			name:       "Valid token",
			expiryTime: time.Now().Add(time.Hour),
			want:       true,
		},
		{
			name:       "Expired token",
			expiryTime: time.Now().Add(-time.Hour),
			want:       false,
		},
		{
			name:       "About to expire token",
			expiryTime: time.Now().Add(10 * time.Second),
			want:       false, // Should be false because of 30 second buffer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TokenResponse{
				ExpiryTime: tt.expiryTime,
			}
			if got := tr.IsValid(); got != tt.want {
				t.Errorf("TokenResponse.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenResponse_GetOtherTokens(t *testing.T) {
	// Create sample token responses
	token1 := TokenResponse{
		AccessToken: "token1",
		ExpiresIn:   3600,
	}
	token2 := TokenResponse{
		AccessToken: "token2",
		ExpiresIn:   7200,
	}

	// Marshal to JSON
	tokenJSON1, _ := json.Marshal(token1)
	tokenJSON2, _ := json.Marshal(token2)

	// Test parsing
	tr := &TokenResponse{
		OtherTokens: []json.RawMessage{tokenJSON1, tokenJSON2},
	}

	otherTokens, err := tr.GetOtherTokens()
	if err != nil {
		t.Fatalf("TokenResponse.GetOtherTokens() error = %v", err)
	}

	if len(otherTokens) != 2 {
		t.Fatalf("TokenResponse.GetOtherTokens() returned %d tokens, want 2", len(otherTokens))
	}

	if otherTokens[0].AccessToken != "token1" || otherTokens[1].AccessToken != "token2" {
		t.Errorf("TokenResponse.GetOtherTokens() returned incorrect tokens")
	}

	// Test with no other tokens
	tr = &TokenResponse{}
	otherTokens, err = tr.GetOtherTokens()
	if err != nil {
		t.Fatalf("TokenResponse.GetOtherTokens() with no tokens error = %v", err)
	}
	if otherTokens != nil {
		t.Errorf("TokenResponse.GetOtherTokens() with no tokens returned %v, want nil", otherTokens)
	}
}

func TestTokenResponse_GetDependentTokens(t *testing.T) {
	// Create sample token responses
	token1 := TokenResponse{
		AccessToken: "dep-token1",
		ExpiresIn:   3600,
	}
	token2 := TokenResponse{
		AccessToken: "dep-token2",
		ExpiresIn:   7200,
	}

	// Marshal to JSON
	tokenJSON1, _ := json.Marshal(token1)
	tokenJSON2, _ := json.Marshal(token2)

	// Test parsing
	tr := &TokenResponse{
		DependentTokens: []json.RawMessage{tokenJSON1, tokenJSON2},
	}

	dependentTokens, err := tr.GetDependentTokens()
	if err != nil {
		t.Fatalf("TokenResponse.GetDependentTokens() error = %v", err)
	}

	if len(dependentTokens) != 2 {
		t.Fatalf("TokenResponse.GetDependentTokens() returned %d tokens, want 2", len(dependentTokens))
	}

	if dependentTokens[0].AccessToken != "dep-token1" || dependentTokens[1].AccessToken != "dep-token2" {
		t.Errorf("TokenResponse.GetDependentTokens() returned incorrect tokens")
	}

	// Test with no dependent tokens
	tr = &TokenResponse{}
	dependentTokens, err = tr.GetDependentTokens()
	if err != nil {
		t.Fatalf("TokenResponse.GetDependentTokens() with no tokens error = %v", err)
	}
	if dependentTokens != nil {
		t.Errorf("TokenResponse.GetDependentTokens() with no tokens returned %v, want nil", dependentTokens)
	}
}

func TestTokenInfo_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		active bool
		want   bool
	}{
		{"Active token", true, true},
		{"Inactive token", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &TokenInfo{Active: tt.active}
			if got := info.IsActive(); got != tt.want {
				t.Errorf("TokenInfo.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenInfo_ExpiresAt(t *testing.T) {
	// Use a fixed timestamp for predictable testing
	timestamp := int64(1609459200) // 2021-01-01 00:00:00 UTC
	info := &TokenInfo{Exp: timestamp}

	expected := time.Unix(timestamp, 0)
	if got := info.ExpiresAt(); !got.Equal(expected) {
		t.Errorf("TokenInfo.ExpiresAt() = %v, want %v", got, expected)
	}
}

func TestTokenInfo_IsExpired(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name string
		exp  int64
		want bool
	}{
		{"Future expiry", now + 3600, false},
		{"Past expiry", now - 3600, true},
		{"Near expiry", now + 10, true}, // Should be true because of 30 second buffer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &TokenInfo{Exp: tt.exp}
			if got := info.IsExpired(); got != tt.want {
				t.Errorf("TokenInfo.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
