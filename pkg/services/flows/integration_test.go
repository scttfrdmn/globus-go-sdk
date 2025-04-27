// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

func getTestCredentials(t *testing.T) (string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	flowID := os.Getenv("GLOBUS_TEST_FLOW_ID")

	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}

	return clientID, clientSecret, flowID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	authClient := auth.NewClient(clientID, clientSecret)

	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), FlowsScope)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	return tokenResp.AccessToken
}

func TestIntegration_ListFlows(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client
	client := NewClient(accessToken)
	ctx := context.Background()

	// List flows
	flows, err := client.ListFlows(ctx, &ListFlowsOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListFlows failed: %v", err)
	}

	// Verify we got some data
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

	// Create Flows client
	client := NewClient(accessToken)
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
		t.Fatalf("Failed to create flow: %v", err)
	}

	// Make sure the flow gets deleted after the test
	defer func() {
		err := client.DeleteFlow(ctx, createdFlow.ID)
		if err != nil {
			t.Logf("Warning: Failed to delete test flow (%s): %v", createdFlow.ID, err)
		} else {
			t.Logf("Successfully deleted test flow (%s)", createdFlow.ID)
		}
	}()

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

func TestIntegration_ExistingFlow(t *testing.T) {
	clientID, clientSecret, flowID := getTestCredentials(t)

	// Skip if no existing flow ID is provided
	if flowID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_FLOW_ID for existing flow operations")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Flows client
	client := NewClient(accessToken)
	ctx := context.Background()

	// Verify we can get the flow
	flow, err := client.GetFlow(ctx, flowID)
	if err != nil {
		t.Fatalf("Failed to get flow: %v", err)
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

	// Create Flows client
	client := NewClient(accessToken)
	ctx := context.Background()

	// List action providers
	providers, err := client.ListActionProviders(ctx, &ListActionProvidersOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Failed to list action providers: %v", err)
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
