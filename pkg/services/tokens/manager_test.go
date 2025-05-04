// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// TestGetTokenRefresh tests that the token manager refreshes tokens correctly
func TestGetTokenRefresh(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		return &auth.TokenResponse{
			AccessToken:  "refreshed-access-token",
			RefreshToken: refreshToken,
			ExpiresIn:    3600, // 1 hour
			Scope:        "test-scope",
		}, nil
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store an expired token
	expiredEntry := &Entry{
		Resource: "test-resource",
		TokenSet: &TokenSet{
			AccessToken:  "expired-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
			Scope:        "test-scope",
		},
	}

	err = storage.Store(expiredEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token, which should trigger a refresh
	refreshedEntry, err := manager.GetToken(ctx, "test-resource")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Verify token was refreshed
	if refreshedEntry.TokenSet.AccessToken != "refreshed-access-token" {
		t.Errorf("Expected refreshed token to be 'refreshed-access-token', got '%s'", refreshedEntry.TokenSet.AccessToken)
	}

	// Verify the refresh token was preserved
	if refreshedEntry.TokenSet.RefreshToken != "test-refresh-token" {
		t.Errorf("Refresh token should be preserved, got '%s'", refreshedEntry.TokenSet.RefreshToken)
	}
}

// TestGetTokenNearExpiry tests that tokens close to expiry are refreshed
func TestGetTokenNearExpiry(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		return &auth.TokenResponse{
			AccessToken:  "refreshed-near-expiry",
			RefreshToken: refreshToken,
			ExpiresIn:    3600, // 1 hour
			Scope:        "test-scope",
		}, nil
	})

	// Create a token manager with a high refresh threshold
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store a token close to expiry
	nearExpiryEntry := &Entry{
		Resource: "near-expiry",
		TokenSet: &TokenSet{
			AccessToken:  "near-expiry-token",
			RefreshToken: "test-refresh-token",
			// Close to expiry but not expired
			ExpiresAt: time.Now().Add(15 * time.Minute),
			Scope:     "test-scope",
		},
	}

	err = storage.Store(nearExpiryEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token, which should trigger a refresh since it's within threshold
	refreshedEntry, err := manager.GetToken(ctx, "near-expiry")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Verify token was refreshed
	if refreshedEntry.TokenSet.AccessToken != "refreshed-near-expiry" {
		t.Errorf("Expected refreshed token to be 'refreshed-near-expiry', got '%s'", refreshedEntry.TokenSet.AccessToken)
	}
}

// TestGetTokenNoRefresh tests that valid tokens far from expiry aren't refreshed
func TestGetTokenNoRefresh(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler that would fail if called
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		t.Error("Refresh handler called when it shouldn't have been")
		return nil, errors.New("refresh should not be called")
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store a valid token far from expiry
	validEntry := &Entry{
		Resource: "valid-token",
		TokenSet: &TokenSet{
			AccessToken:  "valid-access-token",
			RefreshToken: "test-refresh-token",
			// Far from expiry
			ExpiresAt: time.Now().Add(2 * time.Hour),
			Scope:     "test-scope",
		},
	}

	err = storage.Store(validEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token, which should not trigger a refresh
	retrievedEntry, err := manager.GetToken(ctx, "valid-token")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Verify token was not refreshed
	if retrievedEntry.TokenSet.AccessToken != "valid-access-token" {
		t.Errorf("Token was refreshed when it shouldn't have been")
	}
}

// TestGetTokenCannotRefresh tests handling tokens that can't be refreshed
func TestGetTokenCannotRefresh(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler that would fail if called
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		t.Error("Refresh handler called for token without refresh token")
		return nil, errors.New("refresh should not be called")
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store an expired token without a refresh token
	noRefreshEntry := &Entry{
		Resource: "no-refresh",
		TokenSet: &TokenSet{
			AccessToken: "expired-access-token",
			// No refresh token
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Scope:     "test-scope",
		},
	}

	err = storage.Store(noRefreshEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token, which should return the expired token (implementation doesn't error here)
	entry, err := manager.GetToken(ctx, "no-refresh")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Verify the token is still expired
	if !entry.TokenSet.IsExpired() {
		t.Error("Expected token to be expired")
	}
	
	// Verify it's the same token
	if entry.TokenSet.AccessToken != "expired-access-token" {
		t.Errorf("Expected original token, got %s", entry.TokenSet.AccessToken)
	}
}

// TestGetTokenValidButCannotRefresh tests handling tokens that are valid but can't be refreshed
func TestGetTokenValidButCannotRefresh(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler that would fail if called
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		t.Error("Refresh handler called for token without refresh token")
		return nil, errors.New("refresh should not be called")
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store a near-expiry token without a refresh token
	nearExpiryNoRefreshEntry := &Entry{
		Resource: "valid-no-refresh",
		TokenSet: &TokenSet{
			AccessToken: "near-expiry-access-token",
			// No refresh token
			ExpiresAt: time.Now().Add(15 * time.Minute), // Near expiry
			Scope:     "test-scope",
		},
	}

	err = storage.Store(nearExpiryNoRefreshEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token, which should succeed but with the same token
	retrievedEntry, err := manager.GetToken(ctx, "valid-no-refresh")
	if err != nil {
		t.Fatalf("Failed to get valid token without refresh token: %v", err)
	}

	// Verify token was not refreshed
	if retrievedEntry.TokenSet.AccessToken != "near-expiry-access-token" {
		t.Error("Token was refreshed when it shouldn't have been")
	}
}

// TestGetTokenRefreshFailed tests handling a refresh failure
func TestGetTokenRefreshFailed(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Create a mock refresh handler that fails
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		return nil, errors.New("refresh failed")
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Test cases
	tests := []struct {
		name        string
		tokenExpiry time.Duration
		expectError bool
	}{
		{
			name:        "expired token with refresh failure",
			tokenExpiry: -1 * time.Hour,
			expectError: true,
		},
		{
			name:        "valid token with refresh failure",
			tokenExpiry: 1 * time.Hour,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Store a token
			entry := &Entry{
				Resource: tc.name,
				TokenSet: &TokenSet{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					ExpiresAt:    time.Now().Add(tc.tokenExpiry),
					Scope:        "test-scope",
				},
			}

			err = storage.Store(entry)
			if err != nil {
				t.Fatalf("Failed to store token: %v", err)
			}

			// Get the token
			_, err = manager.GetToken(ctx, tc.name)
			if tc.expectError && err == nil {
				t.Error("Expected error, but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Did not expect error, but got: %v", err)
			}
		})
	}
}

// TestRefreshAllTokens tests refreshing all tokens
func TestRefreshAllTokens(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock storage
	storage := NewMemoryStorage()

	// Track refresh calls
	refreshCalls := make(map[string]int)

	// Create a mock refresh handler that tracks calls
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		// Extract resource from refresh token (for testing only)
		resource := refreshToken

		refreshCalls[resource]++

		return &auth.TokenResponse{
			AccessToken:  "refreshed-" + resource,
			RefreshToken: refreshToken,
			ExpiresIn:    3600, // 1 hour
			Scope:        "test-scope",
		}, nil
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store multiple tokens
	resources := []string{"resource1", "resource2", "resource3", "resource4"}
	for i, resource := range resources {
		var entry *Entry

		if i < 2 {
			// First two need refresh
			entry = &Entry{
				Resource: resource,
				TokenSet: &TokenSet{
					AccessToken:  "token-" + resource,
					RefreshToken: resource, // Use resource as refresh token for tracking
					ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
					Scope:        "test-scope",
				},
			}
		} else {
			// Last two don't need refresh
			entry = &Entry{
				Resource: resource,
				TokenSet: &TokenSet{
					AccessToken:  "token-" + resource,
					RefreshToken: resource, // Use resource as refresh token for tracking
					ExpiresAt:    time.Now().Add(2 * time.Hour), // Not expired
					Scope:        "test-scope",
				},
			}
		}

		err = storage.Store(entry)
		if err != nil {
			t.Fatalf("Failed to store token %s: %v", resource, err)
		}
	}

	// Call the refresh method
	manager.refreshAllTokens(ctx)

	// Check which tokens were refreshed
	for i, resource := range resources {
		calls := refreshCalls[resource]
		if i < 2 {
			// First two should be refreshed
			if calls != 1 {
				t.Errorf("Expected resource %s to be refreshed once, got %d refreshes", resource, calls)
			}

			// Verify token was updated in storage
			entry, err := storage.Lookup(resource)
			if err != nil {
				t.Fatalf("Failed to lookup token %s: %v", resource, err)
			}

			if entry.TokenSet.AccessToken != "refreshed-"+resource {
				t.Errorf("Expected token %s to be refreshed, got %s", resource, entry.TokenSet.AccessToken)
			}
		} else {
			// Last two should not be refreshed
			if calls != 0 {
				t.Errorf("Expected resource %s not to be refreshed, got %d refreshes", resource, calls)
			}
		}
	}
}

// TestBackgroundRefresh tests the background refresh functionality
func TestBackgroundRefresh(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create a mock storage
	storage := NewMemoryStorage()

	// Track refresh calls
	refreshCalls := make(map[string]int)
	refreshMutex := &sync.Mutex{}

	// Create a mock refresh handler that tracks calls
	mockHandler := NewMockRefreshHandler()
	mockHandler.SetRefreshFunc(func(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
		// Extract resource from refresh token (for testing only)
		resource := refreshToken

		refreshMutex.Lock()
		refreshCalls[resource]++
		refreshMutex.Unlock()

		return &auth.TokenResponse{
			AccessToken:  "refreshed-" + resource + "-" + time.Now().Format(time.RFC3339Nano),
			RefreshToken: refreshToken,
			ExpiresIn:    3600, // 1 hour
			Scope:        "test-scope",
		}, nil
	})

	// Create a token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithRefreshHandler(mockHandler),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store multiple tokens
	resources := []string{"bg-resource1", "bg-resource2"}
	for _, resource := range resources {
		entry := &Entry{
			Resource: resource,
			TokenSet: &TokenSet{
				AccessToken:  "token-" + resource,
				RefreshToken: resource, // Use resource as refresh token for tracking
				ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
				Scope:        "test-scope",
			},
		}

		err = storage.Store(entry)
		if err != nil {
			t.Fatalf("Failed to store token %s: %v", resource, err)
		}
	}

	// Start background refresh
	stop := manager.StartBackgroundRefresh(100 * time.Millisecond)
	defer stop()

	// Wait for at least one refresh cycle
	time.Sleep(300 * time.Millisecond)

	// Stop background refresh
	stop()

	// Check that tokens were refreshed at least once
	for _, resource := range resources {
		refreshMutex.Lock()
		calls := refreshCalls[resource]
		refreshMutex.Unlock()

		if calls < 1 {
			t.Errorf("Expected resource %s to be refreshed at least once, got %d refreshes", resource, calls)
		}

		// Verify tokens were updated in storage
		entry, err := storage.Lookup(resource)
		if err != nil {
			t.Fatalf("Failed to lookup token %s: %v", resource, err)
		}

		if entry.TokenSet.AccessToken == "token-"+resource {
			t.Errorf("Expected token %s to be refreshed, still has original token", resource)
		}
	}
}

