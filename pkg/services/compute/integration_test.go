//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

func getTestCredentials(t *testing.T) (string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	endpointID := os.Getenv("GLOBUS_TEST_COMPUTE_ENDPOINT_ID")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}

	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	return clientID, clientSecret, endpointID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	// Create auth client with client ID and secret
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), ComputeScope)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	return tokenResp.AccessToken
}

// The Globus Compute API takes and returns open-ended documents; the Go client
// mirrors that with map[string]interface{} bodies and results. These integration
// tests verify the client reaches the live endpoints and returns documents.

func newIntegrationClient(t *testing.T) *Client {
	clientID, clientSecret, _ := getTestCredentials(t)
	accessToken := getAccessToken(t, clientID, clientSecret)
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func TestIntegration_GetVersion(t *testing.T) {
	client := newIntegrationClient(t)

	version, err := client.GetVersion(context.Background(), "")
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	t.Logf("Compute service version document: %v", version)
}

func TestIntegration_GetEndpoints(t *testing.T) {
	client := newIntegrationClient(t)

	endpoints, err := client.GetEndpoints(context.Background(), &GetEndpointsOptions{Role: "owner"})
	if err != nil {
		if core.IsNotFound(err) || core.IsForbidden(err) || core.IsUnauthorized(err) {
			t.Logf("Request reached the service but returned an expected permissions error: %v", err)
			return
		}
		t.Fatalf("GetEndpoints failed with unexpected error: %v", err)
	}
	t.Logf("Endpoints document: %v", endpoints)
}

func TestIntegration_RegisterFunctionLifecycle(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := context.Background()

	// Register a simple function (passthrough document).
	fn, err := client.RegisterFunction(ctx, map[string]interface{}{
		"function_name": "integration_hello",
		"function_code": "def hello(name='World'):\n    return f'Hello, {name}!'\n",
	})
	if err != nil {
		if core.IsForbidden(err) || core.IsUnauthorized(err) {
			t.Skipf("Credentials lack function-registration scope: %v", err)
		}
		t.Fatalf("RegisterFunction failed: %v", err)
	}

	functionID, _ := fn["function_id"].(string)
	if functionID == "" {
		t.Fatalf("RegisterFunction returned no function_id: %v", fn)
	}
	t.Logf("Registered function: %s", functionID)

	// Clean up.
	defer func() {
		if _, err := client.DeleteFunction(ctx, functionID); err != nil {
			t.Logf("Warning: failed to delete test function %s: %v", functionID, err)
		}
	}()

	// Fetch it back.
	fetched, err := client.GetFunction(ctx, functionID)
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}
	t.Logf("Fetched function document: %v", fetched)
}

func TestIntegration_SubmitAndBatchStatus(t *testing.T) {
	client := newIntegrationClient(t)
	_, _, endpointID := getTestCredentials(t)
	if endpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_COMPUTE_ENDPOINT_ID environment variable")
	}
	ctx := context.Background()

	// Submit is an open-ended document (POST /v2/submit).
	result, err := client.Submit(ctx, map[string]interface{}{
		"tasks": map[string]interface{}{
			endpointID: []interface{}{},
		},
	})
	if err != nil {
		if core.IsForbidden(err) || core.IsUnauthorized(err) || core.IsNotFound(err) {
			t.Logf("Request reached the service but returned an expected error: %v", err)
			return
		}
		t.Fatalf("Submit failed with unexpected error: %v", err)
	}
	t.Logf("Submit result document: %v", result)

	// Query any returned task IDs via POST /v2/batch_status.
	if taskIDs, ok := result["task_ids"].([]interface{}); ok && len(taskIDs) > 0 {
		ids := make([]string, 0, len(taskIDs))
		for _, id := range taskIDs {
			if s, ok := id.(string); ok {
				ids = append(ids, s)
			}
		}
		status, err := client.GetBatchStatus(ctx, ids)
		if err != nil {
			t.Fatalf("GetBatchStatus failed: %v", err)
		}
		t.Logf("Batch status document: %v", status)
	}
}
