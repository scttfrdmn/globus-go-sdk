// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
	// Get access token from environment
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_ACCESS_TOKEN environment variable is required")
	}

	// Create a flows client
	client := flows.NewClient(
		accessToken,
		core.WithLogging(true),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run the example
	if err := runFlowsExample(ctx, client); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runFlowsExample(ctx context.Context, client *flows.Client) error {
	fmt.Println("=== Globus Flows Client Example ===")

	// List flows using iterator
	fmt.Println("\n=== Listing Flows ===")
	if err := listFlowsWithIterator(ctx, client); err != nil {
		return fmt.Errorf("error listing flows: %w", err)
	}

	// Get all action providers
	fmt.Println("\n=== Listing Action Providers ===")
	if err := listActionProviders(ctx, client); err != nil {
		return fmt.Errorf("error listing action providers: %w", err)
	}

	// Create a new flow
	fmt.Println("\n=== Creating a Flow ===")
	flow, err := createFlow(ctx, client)
	if err != nil {
		return fmt.Errorf("error creating flow: %w", err)
	}

	fmt.Printf("Created flow: %s (%s)\n", flow.Title, flow.ID)

	// Run the flow
	fmt.Println("\n=== Running the Flow ===")
	run, err := runFlow(ctx, client, flow.ID)
	if err != nil {
		return fmt.Errorf("error running flow: %w", err)
	}

	fmt.Printf("Started flow run: %s (status: %s)\n", run.RunID, run.Status)

	// Wait for run completion
	fmt.Println("\n=== Waiting for Run Completion ===")
	finalRun, err := client.WaitForRun(ctx, run.RunID, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error waiting for run: %w", err)
	}

	fmt.Printf("Run completed with status: %s\n", finalRun.Status)

	// Get logs for the run
	fmt.Println("\n=== Getting Run Logs ===")
	if err := listRunLogs(ctx, client, run.RunID); err != nil {
		return fmt.Errorf("error listing run logs: %w", err)
	}

	// Batch operations example
	fmt.Println("\n=== Batch Operations Example ===")
	if err := batchOperationsExample(ctx, client); err != nil {
		return fmt.Errorf("error in batch operations: %w", err)
	}

	// Clean up by deleting the flow
	fmt.Println("\n=== Cleaning Up ===")
	if err := client.DeleteFlow(ctx, flow.ID); err != nil {
		return fmt.Errorf("error deleting flow: %w", err)
	}

	fmt.Printf("Deleted flow: %s\n", flow.ID)
	return nil
}

func listFlowsWithIterator(ctx context.Context, client *flows.Client) error {
	// Create an iterator
	iterator := client.GetFlowsIterator(&flows.ListFlowsOptions{
		Limit: 5, // Small limit to demonstrate pagination
	})

	// Iterate through flows
	count := 0
	for iterator.Next(ctx) {
		flow := iterator.Flow()
		fmt.Printf("%d. %s (ID: %s, Owner: %s)\n", count+1, flow.Title, flow.ID, flow.FlowOwner)
		count++

		// Just show the first 10 for brevity
		if count >= 10 {
			fmt.Println("... more flows available")
			break
		}
	}

	// Check for errors
	if err := iterator.Err(); err != nil {
		return err
	}

	if count == 0 {
		fmt.Println("No flows found")
	} else {
		fmt.Printf("Found %d flows\n", count)
	}

	return nil
}

func listActionProviders(ctx context.Context, client *flows.Client) error {
	// Get all action providers
	providers, err := client.ListAllActionProviders(ctx, &flows.ListActionProvidersOptions{
		FilterGlobus: true, // Only get Globus-managed providers
	})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d action providers:\n", len(providers))
	for i, provider := range providers {
		fmt.Printf("%d. %s (ID: %s, Type: %s)\n", i+1, provider.DisplayName, provider.ID, provider.Type)

		// For the first provider, list its roles
		if i == 0 {
			roles, err := client.ListAllActionRoles(ctx, provider.ID)
			if err != nil {
				fmt.Printf("   Error listing roles: %v\n", err)
				continue
			}

			fmt.Printf("   Roles (%d):\n", len(roles))
			for j, role := range roles {
				if j >= 3 {
					fmt.Printf("      ... %d more roles\n", len(roles)-j)
					break
				}
				fmt.Printf("      - %s (ID: %s)\n", role.Name, role.ID)
			}
		}
	}

	return nil
}

func createFlow(ctx context.Context, client *flows.Client) (*flows.Flow, error) {
	// Create a simple flow definition
	definition := map[string]interface{}{
		"Comment": "A simple example flow",
		"StartAt": "Echo",
		"States": map[string]interface{}{
			"Echo": map[string]interface{}{
				"Type":      "Action",
				"ActionUrl": "https://actions.globus.org/hello_world",
				"Parameters": map[string]interface{}{
					"echo_string": "$.input.message",
				},
				"ResultPath": "$.output",
				"End":        true,
			},
		},
	}

	// Input schema
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to echo",
			},
		},
		"required": []string{"message"},
	}

	// Create the flow
	createRequest := &flows.FlowCreateRequest{
		Title:       "Example Echo Flow",
		Description: "A simple flow that echoes a message",
		Definition:  definition,
		InputSchema: inputSchema,
		Keywords:    []string{"example", "echo"},
	}

	return client.CreateFlow(ctx, createRequest)
}

func runFlow(ctx context.Context, client *flows.Client, flowID string) (*flows.RunResponse, error) {
	// Create a run request
	runRequest := &flows.RunRequest{
		FlowID: flowID,
		Label:  "Example run",
		Input: map[string]interface{}{
			"message": "Hello, Globus Flows!",
		},
	}

	return client.RunFlow(ctx, runRequest)
}

func listRunLogs(ctx context.Context, client *flows.Client, runID string) error {
	// Get all logs
	logs, err := client.ListAllRunLogs(ctx, runID)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d log entries:\n", len(logs))
	for i, entry := range logs {
		fmt.Printf("%d. [%s] %s\n", i+1, entry.Code, entry.Description)

		// If there are details, print them as JSON
		if len(entry.Details) > 0 {
			detailsJSON, err := json.MarshalIndent(entry.Details, "   ", "  ")
			if err == nil {
				fmt.Printf("   Details: %s\n", string(detailsJSON))
			}
		}
	}

	return nil
}

func batchOperationsExample(ctx context.Context, client *flows.Client) error {
	// First, get a list of flows to work with
	flows, err := client.ListAllFlows(ctx, &flows.ListFlowsOptions{
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to list flows: %w", err)
	}

	if len(flows) == 0 {
		fmt.Println("No flows found for batch operations example")
		return nil
	}

	// Extract flow IDs
	flowIDs := make([]string, len(flows))
	for i, flow := range flows {
		flowIDs[i] = flow.ID
	}

	// Example of BatchGetFlows
	fmt.Println("Batch retrieving flows...")
	batchFlowsResp := client.BatchGetFlows(ctx, &flows.BatchFlowsRequest{
		FlowIDs: flowIDs,
		Options: &flows.BatchOptions{
			Concurrency: 5,
		},
	})

	successCount := 0
	errorCount := 0
	for _, result := range batchFlowsResp.Responses {
		if result.Error != nil {
			errorCount++
			fmt.Printf("Error retrieving flow at index %d: %v\n", result.Index, result.Error)
		} else {
			successCount++
		}
	}
	fmt.Printf("Batch get flows complete: %d successes, %d errors\n", successCount, errorCount)

	// Example of BatchRunFlows (if we want to run multiple instances)
	if len(flows) > 0 {
		// Find an appropriate flow to run (assume the first one is runnable)
		flowID := flows[0].ID

		// Create multiple run requests
		runRequests := []*flows.RunRequest{
			{
				FlowID: flowID,
				Label:  "Batch run 1",
				Input: map[string]interface{}{
					"message": "Batch message 1",
				},
			},
			{
				FlowID: flowID,
				Label:  "Batch run 2",
				Input: map[string]interface{}{
					"message": "Batch message 2",
				},
			},
		}

		fmt.Println("Starting batch flow runs...")
		batchRunResp := client.BatchRunFlows(ctx, &flows.BatchRunFlowsRequest{
			Requests: runRequests,
			Options: &flows.BatchOptions{
				Concurrency: 2,
			},
		})

		runIDs := make([]string, 0)
		for i, result := range batchRunResp.Responses {
			if result.Error != nil {
				fmt.Printf("Error starting run at index %d: %v\n", i, result.Error)
			} else {
				fmt.Printf("Started run: %s (label: %s)\n", result.Response.RunID, result.Response.Label)
				runIDs = append(runIDs, result.Response.RunID)
			}
		}

		// If we started any runs, batch cancel them
		if len(runIDs) > 0 {
			fmt.Println("Batch canceling runs...")
			batchCancelResp := client.BatchCancelRuns(ctx, &flows.BatchCancelRunsRequest{
				RunIDs: runIDs,
			})

			for i, result := range batchCancelResp.Responses {
				if result.Error != nil {
					fmt.Printf("Error canceling run at index %d: %v\n", i, result.Error)
				} else {
					fmt.Printf("Canceled run: %s\n", result.RunID)
				}
			}
		}
	}

	return nil
}
