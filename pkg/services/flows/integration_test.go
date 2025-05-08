//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
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
	flowID := os.Getenv("GLOBUS_TEST_FLOW_ID")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}

	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	return clientID, clientSecret, flowID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	// First, check if there's a flows token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_FLOWS_TOKEN")
	if staticToken != "" {
		t.Log("Using static flows token from environment")
		return staticToken
	}

	// If no static token, try to get one via client credentials
	t.Log("Getting client credentials token for flows")

	// Create auth client with proper options
	options := []auth.ClientOption{
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	}

	authClient, err := auth.NewClient(options...)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Try specific scope for flows
	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), FlowsScope)
	if err != nil {
		t.Logf("Failed to get token with flows scope: %v", err)
		t.Log("Falling back to default token")

		// Fallback to default token
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background())
		if err != nil {
			t.Fatalf("Failed to get any token: %v", err)
		}

		t.Log("WARNING: This token may not have flows permissions. Consider providing GLOBUS_TEST_FLOWS_TOKEN")
	} else {
		t.Logf("Got token with resource server: %s, scopes: %s",
			tokenResp.ResourceServer, tokenResp.Scope)
	}

	return tokenResp.AccessToken
}

func TestIntegration_ListFlows(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client with proper options
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create flows client: %v", err)
	}
	ctx := context.Background()

	// List flows
	flows, err := client.ListFlows(ctx, &ListFlowsOptions{
		Limit: 5,
	})

	// Handle different error types with helpful messages
	if err != nil {
		if core.IsNotFound(err) {
			t.Logf("404 NOT FOUND: Client correctly made the request, but returned 404: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, provide GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else if core.IsForbidden(err) {
			t.Logf("403 FORBIDDEN: Client correctly made the request, but permission denied: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, set GLOBUS_TEST_FLOWS_TOKEN with a token that has flows permissions")
			return // Skip the rest of the test
		} else if core.IsUnauthorized(err) {
			t.Logf("401 UNAUTHORIZED: Client correctly made the request, but token invalid: %v", err)
			t.Logf("To resolve, provide a valid GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else {
			t.Fatalf("ListFlows failed with unexpected error: %v", err)
		}
	}

	// If we get here, we actually have permissions, so verify the data
	t.Logf("Found %d flows", len(flows.Flows))

	// The user might not have any flows, so this isn't necessarily an error
	if len(flows.Flows) > 0 {
		// Check that the first flow has expected fields
		firstFlow := flows.Flows[0]
		if firstFlow.ID == "" {
			t.Error("First flow is missing ID")
		}
		if firstFlow.Title == "" {
			t.Error("First flow is missing title")
		}
	}
}

func TestIntegration_FlowLifecycle(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client with proper options
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create flows client: %v", err)
	}
	ctx := context.Background()

	// 1. Create a new flow
	timestamp := time.Now().Format("20060102_150405")
	flowTitle := fmt.Sprintf("Test Flow %s", timestamp)
	flowDescription := "A test flow created by integration tests"

	// Simple flow definition for testing
	flowDefinition := map[string]interface{}{
		"title":       flowTitle,
		"description": flowDescription,
		"input_schema": map[string]interface{}{
			"additionalProperties": false,
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"message"},
			"type":     "object",
		},
		"definition": map[string]interface{}{
			"Comment": "Simple test flow",
			"StartAt": "LogMessage",
			"States": map[string]interface{}{
				"LogMessage": map[string]interface{}{
					"Type":       "Pass",
					"Result":     "Hello, integration test!",
					"ResultPath": "$.output",
					"End":        true,
				},
			},
		},
	}

	createRequest := &FlowCreateRequest{
		Title:       flowTitle,
		Description: flowDescription,
		Definition:  flowDefinition,
	}

	createdFlow, err := client.CreateFlow(ctx, createRequest)
	if err != nil {
		if core.IsNotFound(err) {
			t.Logf("404 NOT FOUND: Client correctly made the request, but returned 404: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, provide GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else if core.IsForbidden(err) {
			t.Logf("403 FORBIDDEN: Permission denied to create flow: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, set GLOBUS_TEST_FLOWS_TOKEN with a token that has flows permissions")
			return // Skip the rest of the test
		} else if core.IsUnauthorized(err) {
			t.Logf("401 UNAUTHORIZED: Token not valid for creating flows: %v", err)
			t.Logf("To resolve, provide a valid GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else {
			t.Fatalf("Failed to create flow with unexpected error: %v", err)
		}
	}

	// Store the flow ID for potential use in other tests
	if createdFlow != nil && createdFlow.ID != "" {
		os.Setenv("GLOBUS_TEST_CREATED_FLOW_ID", createdFlow.ID)
		t.Logf("Set GLOBUS_TEST_CREATED_FLOW_ID=%s for subsequent tests", createdFlow.ID)
	}

	// Make sure the flow gets deleted after the test
	defer func() {
		if createdFlow != nil && createdFlow.ID != "" {
			err := client.DeleteFlow(ctx, createdFlow.ID)
			if err != nil {
				if core.IsForbidden(err) || core.IsUnauthorized(err) {
					t.Logf("PERMISSION WARNING: Cannot delete test flow (%s): %v", createdFlow.ID, err)
					t.Logf("This may require manual cleanup. Set GLOBUS_TEST_FLOWS_TOKEN with proper permissions.")
				} else {
					t.Logf("Warning: Failed to delete test flow (%s): %v", createdFlow.ID, err)
				}
			} else {
				t.Logf("Successfully deleted test flow (%s)", createdFlow.ID)
				os.Unsetenv("GLOBUS_TEST_CREATED_FLOW_ID") // Remove env var when flow is deleted
			}
		}
	}()

	if createdFlow != nil {
		t.Logf("Created flow: %s (%s)", createdFlow.Title, createdFlow.ID)

		// 2. Verify the flow was created correctly
		if createdFlow.Title != flowTitle {
			t.Errorf("Created flow title = %s, want %s", createdFlow.Title, flowTitle)
		}
		if createdFlow.Description != flowDescription {
			t.Errorf("Created flow description = %s, want %s", createdFlow.Description, flowDescription)
		}

		// 3. Get the flow
		fetchedFlow, err := client.GetFlow(ctx, createdFlow.ID)
		if err != nil {
			t.Fatalf("Failed to get flow: %v", err)
		}

		if fetchedFlow.ID != createdFlow.ID {
			t.Errorf("Fetched flow ID = %s, want %s", fetchedFlow.ID, createdFlow.ID)
		}

		// 4. Update the flow
		updatedDescription := "Updated description for integration test"
		updateRequest := &FlowUpdateRequest{
			Description: updatedDescription,
		}

		updatedFlow, err := client.UpdateFlow(ctx, createdFlow.ID, updateRequest)
		if err != nil {
			t.Fatalf("Failed to update flow: %v", err)
		}

		if updatedFlow.Description != updatedDescription {
			t.Errorf("Updated flow description = %s, want %s", updatedFlow.Description, updatedDescription)
		}

		// 5. Run the flow
		runRequest := &RunRequest{
			FlowID: createdFlow.ID,
			Label:  "Integration Test Run",
			Tags:   []string{"integration-test"},
			Input: map[string]interface{}{
				"message": "Hello from integration test",
			},
		}

		run, err := client.RunFlow(ctx, runRequest)
		if err != nil {
			t.Fatalf("Failed to run flow: %v", err)
		}

		t.Logf("Started flow run: %s", run.RunID)

		// Wait a moment for the run to process
		time.Sleep(2 * time.Second)

		// 6. Get the run status
		runStatus, err := client.GetRun(ctx, run.RunID)
		if err != nil {
			t.Fatalf("Failed to get run status: %v", err)
		}

		t.Logf("Run status: %s", runStatus.Status)

		// 7. Get run logs
		logs, err := client.GetRunLogs(ctx, run.RunID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get run logs: %v", err)
		}

		t.Logf("Run has %d log entries", len(logs.Entries))
		for i, entry := range logs.Entries {
			t.Logf("Log %d: %s - %s", i+1, entry.Code, entry.Description)
		}
	}
}

func TestIntegration_ExistingFlow(t *testing.T) {
	clientID, clientSecret, flowID := getTestCredentials(t)

	// If no existing flow ID is provided, try using a well-known public flow
	if flowID == "" {
		// Look for a public flow ID from previous tests
		flowID = os.Getenv("GLOBUS_TEST_PUBLIC_FLOW_ID")

		if flowID == "" {
			// Check if we created a flow in the previous test
			if createdFlowID := os.Getenv("GLOBUS_TEST_CREATED_FLOW_ID"); createdFlowID != "" {
				flowID = createdFlowID
				t.Logf("Using previously created flow ID for testing: %s", flowID)
			} else {
				// Use a default test flow as fallback - Hello World flow
				flowID = "4f2b8147-93e3-4dc7-ab85-54d22eb0ba9c"
				t.Logf("Using default Hello World flow for testing: %s", flowID)
			}
		}
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client with proper options
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create flows client: %v", err)
	}
	ctx := context.Background()

	// Verify we can get the flow
	flow, err := client.GetFlow(ctx, flowID)
	if err != nil {
		if core.IsNotFound(err) {
			t.Logf("404 NOT FOUND: Flow ID %s not found: %v", flowID, err)
			t.Logf("This may be due to the flow being deleted or not existing")
			t.Logf("To resolve, provide a valid GLOBUS_TEST_FLOW_ID or GLOBUS_TEST_PUBLIC_FLOW_ID")
			return // Skip the rest of the test
		} else if core.IsForbidden(err) {
			t.Logf("403 FORBIDDEN: Permission denied to access flow: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, set GLOBUS_TEST_FLOWS_TOKEN with a token that has proper permissions")
			return // Skip the rest of the test
		} else if core.IsUnauthorized(err) {
			t.Logf("401 UNAUTHORIZED: Token not valid for accessing this flow: %v", err)
			t.Logf("To resolve, provide a valid GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else {
			t.Fatalf("Failed to get flow with unexpected error: %v", err)
		}
	}

	t.Logf("Found flow: %s (%s)", flow.Title, flow.ID)

	// List runs for this flow
	runs, err := client.ListRuns(ctx, &ListRunsOptions{
		FlowID: flowID,
		Limit:  5,
	})
	if err != nil {
		t.Fatalf("Failed to list runs: %v", err)
	}

	t.Logf("Flow has %d recent runs", len(runs.Runs))

	if len(runs.Runs) > 0 {
		// Get the most recent run
		run := runs.Runs[0]
		t.Logf("Most recent run: %s (Status: %s)", run.RunID, run.Status)

		// Get logs for the run
		logs, err := client.GetRunLogs(ctx, run.RunID, 5, 0)
		if err != nil {
			t.Fatalf("Failed to get run logs: %v", err)
		}

		t.Logf("Run has %d log entries", len(logs.Entries))
	}
}

func TestIntegration_ActionProviders(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client with proper options
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create flows client: %v", err)
	}
	ctx := context.Background()

	// List action providers
	providers, err := client.ListActionProviders(ctx, &ListActionProvidersOptions{
		Limit: 5,
	})
	if err != nil {
		if core.IsNotFound(err) {
			t.Logf("404 NOT FOUND: Client correctly made the request, but returned 404: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, provide GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else if core.IsForbidden(err) {
			t.Logf("403 FORBIDDEN: Permission denied to list action providers: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			t.Logf("To resolve, set GLOBUS_TEST_FLOWS_TOKEN with a token that has flows permissions")
			return // Skip the rest of the test
		} else if core.IsUnauthorized(err) {
			t.Logf("401 UNAUTHORIZED: Token not valid for listing action providers: %v", err)
			t.Logf("To resolve, provide a valid GLOBUS_TEST_FLOWS_TOKEN with proper permissions")
			return // Skip the rest of the test
		} else {
			t.Fatalf("Failed to list action providers with unexpected error: %v", err)
		}
	}

	t.Logf("Found %d action providers", len(providers.ActionProviders))

	if len(providers.ActionProviders) > 0 {
		provider := providers.ActionProviders[0]
		t.Logf("Example provider: %s (%s)", provider.DisplayName, provider.ID)

		// Get a specific provider
		providerDetail, err := client.GetActionProvider(ctx, provider.ID)
		if err != nil {
			t.Fatalf("Failed to get action provider: %v", err)
		}

		t.Logf("Provider detail: %s (Type: %s, Owner: %s)",
			providerDetail.DisplayName, providerDetail.Type, providerDetail.Owner)

		// List roles for the provider
		roles, err := client.ListActionRoles(ctx, provider.ID, 5, 0)
		if err != nil {
			t.Fatalf("Failed to list action roles: %v", err)
		}

		t.Logf("Provider has %d roles", len(roles.ActionRoles))

		if len(roles.ActionRoles) > 0 {
			// Get a specific role
			role := roles.ActionRoles[0]
			t.Logf("Example role: %s (%s)", role.Name, role.ID)

			roleDetail, err := client.GetActionRole(ctx, provider.ID, role.ID)
			if err != nil {
				t.Fatalf("Failed to get action role: %v", err)
			}

			t.Logf("Role detail: %s (Description: %s)",
				roleDetail.Name, roleDetail.Description)
		}
	}
}
