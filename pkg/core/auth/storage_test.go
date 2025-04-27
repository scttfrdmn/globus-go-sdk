// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenInfo(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(time.Hour)
	pastTime := now.Add(-time.Hour)

	// Test valid token
	validToken := TokenInfo{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    futureTime,
	}

	if !validToken.IsValid() {
		t.Error("Token with future expiry should be valid")
	}

	if !validToken.CanRefresh() {
		t.Error("Token with refresh token should be refreshable")
	}

	// Test expired token
	expiredToken := TokenInfo{
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    pastTime,
	}

	if expiredToken.IsValid() {
		t.Error("Token with past expiry should be invalid")
	}

	if !expiredToken.CanRefresh() {
		t.Error("Expired token with refresh token should still be refreshable")
	}

	// Test non-refreshable token
	nonRefreshableToken := TokenInfo{
		AccessToken: "non-refreshable-token",
		ExpiresAt:   futureTime,
	}

	if !nonRefreshableToken.IsValid() {
		t.Error("Non-refreshable token with future expiry should be valid")
	}

	if nonRefreshableToken.CanRefresh() {
		t.Error("Token without refresh token should not be refreshable")
	}
}

func TestMemoryTokenStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryTokenStorage()

	// Test storing and retrieving a token
	token1 := TokenInfo{
		AccessToken:  "test-token-1",
		RefreshToken: "refresh-token-1",
		ExpiresAt:    time.Now().Add(time.Hour),
		Scopes:       []string{"scope1", "scope2"},
		ResourceID:   "resource1",
	}

	// Store the token
	err := storage.StoreToken(ctx, "user1", token1)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Retrieve the token
	retrievedToken, err := storage.GetToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Check the token values
	if retrievedToken.AccessToken != token1.AccessToken {
		t.Errorf("Retrieved access token = %s, want %s", retrievedToken.AccessToken, token1.AccessToken)
	}
	if retrievedToken.RefreshToken != token1.RefreshToken {
		t.Errorf("Retrieved refresh token = %s, want %s", retrievedToken.RefreshToken, token1.RefreshToken)
	}
	if !retrievedToken.ExpiresAt.Equal(token1.ExpiresAt) {
		t.Errorf("Retrieved expiry time = %v, want %v", retrievedToken.ExpiresAt, token1.ExpiresAt)
	}
	if len(retrievedToken.Scopes) != len(token1.Scopes) {
		t.Errorf("Retrieved scopes count = %d, want %d", len(retrievedToken.Scopes), len(token1.Scopes))
	}
	if retrievedToken.ResourceID != token1.ResourceID {
		t.Errorf("Retrieved resource ID = %s, want %s", retrievedToken.ResourceID, token1.ResourceID)
	}

	// Store a second token
	token2 := TokenInfo{
		AccessToken:  "test-token-2",
		RefreshToken: "refresh-token-2",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	err = storage.StoreToken(ctx, "user2", token2)
	if err != nil {
		t.Fatalf("Failed to store second token: %v", err)
	}

	// List tokens
	keys, err := storage.ListTokens(ctx)
	if err != nil {
		t.Fatalf("Failed to list tokens: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Listed token count = %d, want 2", len(keys))
	}

	// Check if both keys are in the list
	foundUser1 := false
	foundUser2 := false
	for _, key := range keys {
		if key == "user1" {
			foundUser1 = true
		}
		if key == "user2" {
			foundUser2 = true
		}
	}
	if !foundUser1 || !foundUser2 {
		t.Errorf("Listed tokens should include both user1 and user2")
	}

	// Delete a token
	err = storage.DeleteToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify the token is deleted
	_, err = storage.GetToken(ctx, "user1")
	if err != ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound after deletion, got: %v", err)
	}

	// List tokens again
	keys, err = storage.ListTokens(ctx)
	if err != nil {
		t.Fatalf("Failed to list tokens after deletion: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Listed token count after deletion = %d, want 1", len(keys))
	}
	if keys[0] != "user2" {
		t.Errorf("Remaining token key = %s, want user2", keys[0])
	}
}

func TestFileTokenStorage(t *testing.T) {
	// Create a temporary directory for token storage
	tempDir, err := os.MkdirTemp("", "token-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	storage, err := NewFileTokenStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file token storage: %v", err)
	}

	// Test storing and retrieving a token
	token1 := TokenInfo{
		AccessToken:  "test-token-1",
		RefreshToken: "refresh-token-1",
		ExpiresAt:    time.Now().Add(time.Hour),
		Scopes:       []string{"scope1", "scope2"},
		ResourceID:   "resource1",
	}

	// Store the token
	err = storage.StoreToken(ctx, "user1", token1)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(filepath.Join(tempDir, "user1.json")); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}

	// Retrieve the token
	retrievedToken, err := storage.GetToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Check the token values
	if retrievedToken.AccessToken != token1.AccessToken {
		t.Errorf("Retrieved access token = %s, want %s", retrievedToken.AccessToken, token1.AccessToken)
	}
	if retrievedToken.RefreshToken != token1.RefreshToken {
		t.Errorf("Retrieved refresh token = %s, want %s", retrievedToken.RefreshToken, token1.RefreshToken)
	}
	// Allow small time differences due to serialization
	if retrievedToken.ExpiresAt.Sub(token1.ExpiresAt).Abs() > time.Second {
		t.Errorf("Retrieved expiry time = %v, want %v", retrievedToken.ExpiresAt, token1.ExpiresAt)
	}
	if len(retrievedToken.Scopes) != len(token1.Scopes) {
		t.Errorf("Retrieved scopes count = %d, want %d", len(retrievedToken.Scopes), len(token1.Scopes))
	}
	if retrievedToken.ResourceID != token1.ResourceID {
		t.Errorf("Retrieved resource ID = %s, want %s", retrievedToken.ResourceID, token1.ResourceID)
	}

	// Store a second token
	token2 := TokenInfo{
		AccessToken:  "test-token-2",
		RefreshToken: "refresh-token-2",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	err = storage.StoreToken(ctx, "user2", token2)
	if err != nil {
		t.Fatalf("Failed to store second token: %v", err)
	}

	// List tokens
	keys, err := storage.ListTokens(ctx)
	if err != nil {
		t.Fatalf("Failed to list tokens: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Listed token count = %d, want 2", len(keys))
	}

	// Check if both keys are in the list
	foundUser1 := false
	foundUser2 := false
	for _, key := range keys {
		if key == "user1" {
			foundUser1 = true
		}
		if key == "user2" {
			foundUser2 = true
		}
	}
	if !foundUser1 || !foundUser2 {
		t.Errorf("Listed tokens should include both user1 and user2")
	}

	// Delete a token
	err = storage.DeleteToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify the token is deleted
	_, err = storage.GetToken(ctx, "user1")
	if err != ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound after deletion, got: %v", err)
	}

	// List tokens again
	keys, err = storage.ListTokens(ctx)
	if err != nil {
		t.Fatalf("Failed to list tokens after deletion: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("Listed token count after deletion = %d, want 1", len(keys))
	}
}
