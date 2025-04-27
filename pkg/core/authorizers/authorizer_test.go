// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"testing"
	"time"
)

func TestNullAuthorizer(t *testing.T) {
	auth := &NullAuthorizer{}

	// Test GetAuthorizationHeader
	header, err := auth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("NullAuthorizer.GetAuthorizationHeader returned error: %v", err)
	}
	if header != "" {
		t.Errorf("NullAuthorizer.GetAuthorizationHeader returned non-empty header: %s", header)
	}

	// Test HandleMissingAuthorization
	if auth.HandleMissingAuthorization(context.Background()) {
		t.Error("NullAuthorizer.HandleMissingAuthorization returned true, expected false")
	}

	// Test IsExpired
	if auth.IsExpired() {
		t.Error("NullAuthorizer.IsExpired returned true, expected false")
	}
}

func TestStaticTokenAuthorizer(t *testing.T) {
	testToken := "test-token"
	auth := NewStaticTokenAuthorizer(testToken)

	// Test GetAuthorizationHeader
	header, err := auth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("StaticTokenAuthorizer.GetAuthorizationHeader returned error: %v", err)
	}
	if header != "Bearer "+testToken {
		t.Errorf("StaticTokenAuthorizer.GetAuthorizationHeader returned unexpected header: %s", header)
	}

	// Test with empty token
	emptyAuth := NewStaticTokenAuthorizer("")
	emptyHeader, err := emptyAuth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("StaticTokenAuthorizer(empty).GetAuthorizationHeader returned error: %v", err)
	}
	if emptyHeader != "" {
		t.Errorf("StaticTokenAuthorizer(empty).GetAuthorizationHeader returned non-empty header: %s", emptyHeader)
	}

	// Test HandleMissingAuthorization
	if auth.HandleMissingAuthorization(context.Background()) {
		t.Error("StaticTokenAuthorizer.HandleMissingAuthorization returned true, expected false")
	}

	// Test IsExpired
	if auth.IsExpired() {
		t.Error("StaticTokenAuthorizer.IsExpired returned true, expected false")
	}
}

func TestRefreshableTokenAuthorizer(t *testing.T) {
	testToken := "test-token"
	testRefreshToken := "test-refresh-token"
	expiresIn := 3600 // 1 hour

	// Mock refresh function
	refreshCalled := false
	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		refreshCalled = true
		if refreshToken != testRefreshToken {
			t.Errorf("RefreshFunc received unexpected refreshToken: %s", refreshToken)
		}
		return "new-access-token", "new-refresh-token", time.Now().Add(time.Hour), nil
	}

	auth := NewRefreshableTokenAuthorizer(testToken, testRefreshToken, expiresIn, refreshFunc)

	// Test GetAuthorizationHeader
	header, err := auth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("RefreshableTokenAuthorizer.GetAuthorizationHeader returned error: %v", err)
	}
	if header != "Bearer "+testToken {
		t.Errorf("RefreshableTokenAuthorizer.GetAuthorizationHeader returned unexpected header: %s", header)
	}

	// Test with empty token
	emptyAuth := NewRefreshableTokenAuthorizer("", testRefreshToken, expiresIn, refreshFunc)
	emptyHeader, err := emptyAuth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("RefreshableTokenAuthorizer(empty).GetAuthorizationHeader returned error: %v", err)
	}
	if emptyHeader != "" {
		t.Errorf("RefreshableTokenAuthorizer(empty).GetAuthorizationHeader returned non-empty header: %s", emptyHeader)
	}

	// Test IsExpired
	if auth.IsExpired() {
		t.Error("RefreshableTokenAuthorizer.IsExpired returned true for a non-expired token")
	}

	// Test with expired token
	expiredAuth := NewRefreshableTokenAuthorizer(testToken, testRefreshToken, -10, refreshFunc)
	if !expiredAuth.IsExpired() {
		t.Error("RefreshableTokenAuthorizer.IsExpired returned false for an expired token")
	}

	// Test HandleMissingAuthorization with non-expired token
	if auth.HandleMissingAuthorization(context.Background()) {
		t.Error("RefreshableTokenAuthorizer.HandleMissingAuthorization returned true for non-expired token")
	}
	if refreshCalled {
		t.Error("RefreshFunc was called for non-expired token")
	}

	// Test HandleMissingAuthorization with expired token
	refreshCalled = false
	expiredAuth.HandleMissingAuthorization(context.Background())
	if !refreshCalled {
		t.Error("RefreshFunc was not called for expired token")
	}

	// Check that tokens were updated
	if expiredAuth.AccessToken != "new-access-token" {
		t.Errorf("Access token not updated, got: %s", expiredAuth.AccessToken)
	}
	if expiredAuth.RefreshToken != "new-refresh-token" {
		t.Errorf("Refresh token not updated, got: %s", expiredAuth.RefreshToken)
	}

	// Test with no refresh token
	noRefreshAuth := NewRefreshableTokenAuthorizer(testToken, "", expiresIn, refreshFunc)
	refreshCalled = false
	if noRefreshAuth.HandleMissingAuthorization(context.Background()) {
		t.Error("RefreshableTokenAuthorizer.HandleMissingAuthorization returned true with no refresh token")
	}
	if refreshCalled {
		t.Error("RefreshFunc was called with no refresh token")
	}

	// Test with no refresh function
	noFuncAuth := NewRefreshableTokenAuthorizer(testToken, testRefreshToken, expiresIn, nil)
	if noFuncAuth.HandleMissingAuthorization(context.Background()) {
		t.Error("RefreshableTokenAuthorizer.HandleMissingAuthorization returned true with no refresh function")
	}
}

func TestClientCredentialsAuthorizer(t *testing.T) {
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	scopes := []string{"scope1", "scope2"}

	// Mock auth function
	authCalled := false
	authFunc := func(ctx context.Context, cID, cSecret string, s []string) (string, time.Time, error) {
		authCalled = true
		if cID != clientID {
			t.Errorf("AuthFunc received unexpected clientID: %s", cID)
		}
		if cSecret != clientSecret {
			t.Errorf("AuthFunc received unexpected clientSecret: %s", cSecret)
		}
		if len(s) != len(scopes) {
			t.Errorf("AuthFunc received unexpected scopes: %v", s)
		}
		return "new-access-token", time.Now().Add(time.Hour), nil
	}

	auth := NewClientCredentialsAuthorizer(clientID, clientSecret, scopes, authFunc)

	// Test IsExpired (should be true initially with no token)
	if !auth.IsExpired() {
		t.Error("ClientCredentialsAuthorizer.IsExpired returned false for initial state with no token")
	}

	// Test GetAuthorizationHeader (should call auth function)
	header, err := auth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("ClientCredentialsAuthorizer.GetAuthorizationHeader returned error: %v", err)
	}
	if !authCalled {
		t.Error("AuthFunc was not called during GetAuthorizationHeader")
	}
	if header != "Bearer new-access-token" {
		t.Errorf("ClientCredentialsAuthorizer.GetAuthorizationHeader returned unexpected header: %s", header)
	}

	// Test IsExpired after getting token
	if auth.IsExpired() {
		t.Error("ClientCredentialsAuthorizer.IsExpired returned true after getting a token")
	}

	// Test HandleMissingAuthorization
	authCalled = false
	if !auth.HandleMissingAuthorization(context.Background()) {
		t.Error("ClientCredentialsAuthorizer.HandleMissingAuthorization returned false")
	}
	if !authCalled {
		t.Error("AuthFunc was not called during HandleMissingAuthorization")
	}

	// Test with no auth function
	noFuncAuth := NewClientCredentialsAuthorizer(clientID, clientSecret, scopes, nil)
	noHeader, err := noFuncAuth.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("ClientCredentialsAuthorizer(no func).GetAuthorizationHeader returned error: %v", err)
	}
	if noHeader != "Bearer " {
		t.Errorf("ClientCredentialsAuthorizer(no func).GetAuthorizationHeader returned unexpected header: %s", noHeader)
	}

	if noFuncAuth.HandleMissingAuthorization(context.Background()) {
		t.Error("ClientCredentialsAuthorizer(no func).HandleMissingAuthorization returned true")
	}
}
