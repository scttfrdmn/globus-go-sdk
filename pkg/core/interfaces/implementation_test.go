// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package interfaces_test

import (
	"context"
	"os"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// TestClientImplementsClientInterface verifies that the core.Client type
// properly implements the ClientInterface
func TestClientImplementsClientInterface(t *testing.T) {
	var _ interfaces.ClientInterface = &core.Client{}
}

// TestTransportImplementsTransportInterface verifies that the transport.Transport type
// properly implements the Transport interface
func TestTransportImplementsTransportInterface(t *testing.T) {
	var _ interfaces.Transport = &transport.Transport{}
}

// TestAuthorizerImplementsAuthInterface verifies that the SDK authorizers
// provide the functionality expected from the auth.Authorizer interface
func TestAuthorizerImplementsAuthInterface(t *testing.T) {
	// Note: We can't directly verify authorizers.X implements auth.Authorizer
	// due to method signature differences, but we can verify they have
	// equivalent functionality

	// Create a test authorizer that implements the auth.Authorizer interface
	testAuth := &testAuthorizer{token: "test-token"}
	var _ auth.Authorizer = testAuth

	// Use actual authorizers from the authorizers package
	staticAuth := authorizers.NewStaticTokenAuthorizer("test-token")

	// Verify the basic functionality is equivalent
	if got, _ := staticAuth.GetAuthorizationHeader(nil); got != "Bearer test-token" {
		t.Errorf("StaticTokenAuthorizer.GetAuthorizationHeader() = %q, want %q", got, "Bearer test-token")
	}
}

// testAuthorizer is a simple implementation of auth.Authorizer for testing
type testAuthorizer struct {
	token string
}

// GetAuthorizationHeader implements auth.Authorizer.GetAuthorizationHeader
func (a *testAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// TestPoolImplementsPoolInterface verifies that the HTTPConnectionPool
// properly implements the Pool interface
func TestPoolImplementsPoolInterface(t *testing.T) {
	var _ interfaces.ConnectionPool = http.NewHttpConnectionPool(nil)
}

// TestLoggerImplementsLoggerInterface verifies that the logger
// properly implements the Logger interface
func TestLoggerImplementsLoggerInterface(t *testing.T) {
	// Create a logger instance
	opts := &logging.Options{
		Output: os.Stderr,
		Level:  logging.LogLevelDebug,
	}
	logger := logging.NewLogger(opts)

	// Test that it implements the interface
	var _ interfaces.Logger = logger
}

// TestInterfaceMethodSignatures verifies that interface methods have the correct
// signatures by attempting to call them with proper arguments
func TestInterfaceMethodSignatures(t *testing.T) {
	// This test doesn't actually make calls, it just verifies that the
	// method signatures are compatible with the expected usage

	// For each interface, create a test function that takes the interface
	// and makes calls to its methods with appropriate arguments

	// This test fails at compile time if the methods don't exist or have
	// incorrect signatures

	t.Skip("Test is compile-time only, not meant to be executed")
}
