// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package auth

import (
	"context"
	"reflect"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/tests/compatibility"
)

// AuthClientTest verifies Auth client compatibility
type AuthClientTest struct{}

func (test *AuthClientTest) Name() string {
	return "AuthClient"
}

func (test *AuthClientTest) Setup(ctx context.Context) error {
	return nil
}

func (test *AuthClientTest) Run(ctx context.Context, version string, t *testing.T) error {
	// Create Auth client
	client, err := auth.NewClient()
	if err != nil {
		t.Fatalf("Failed to create Auth client: %v", err)
	}

	// Verify required methods exist
	t.Run("MethodsExist", func(t *testing.T) {
		// Get method type using reflection
		clientType := reflect.TypeOf(client)

		// List of required methods to verify
		requiredMethods := []string{
			"GetTokenInfo",
			"IntrospectToken",
			"RevokeToken",
			"ExchangeAuthorizationCode",
			"GetClientCredentialsGrant",
		}

		// Verify each method exists
		for _, methodName := range requiredMethods {
			method, found := clientType.MethodByName(methodName)
			if !found {
				t.Errorf("Required method %s not found", methodName)
				continue
			}

			// Log method signature for debugging
			t.Logf("Method %s exists with type %v", methodName, method.Type)
		}
	})

	// Check if client has the BaseClient field that implements interfaces.ClientInterface
	t.Run("HasBaseClient", func(t *testing.T) {
		clientType := reflect.TypeOf(client).Elem()
		_, found := clientType.FieldByName("BaseClient")
		if !found {
			t.Errorf("Auth client doesn't have BaseClient field")
		}
	})

	return nil
}

func (test *AuthClientTest) Teardown(ctx context.Context) error {
	return nil
}

func init() {
	compatibility.RegisterTest(&AuthClientTest{})
}
