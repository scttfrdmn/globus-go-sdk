// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create a new Auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Get token using client credentials for simplicity
	// In a real application, you would likely use the authorization code flow
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.FlowsScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)
	accessToken := tokenResp.AccessToken

	// Create Flows client
	flowsClient := config.NewFlowsClient(accessToken)

	// Check if flow ID is provided
	flowID := os.Getenv("GLOBUS_FLOW_ID")
	if flowID == "" {
		// List available flows if no flow ID is provided
		fmt.Println("\n=== Available Flows ===")
		
		flowsList, err := flowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
			Limit: 5,
		})
		if err != nil {
			log.Fatalf("Failed to list flows: %v", err)
		}
		
		if len(flowsList.Flows) == 0 {
			fmt.Println("No flows found. Create a flow first.")
			
			// Create a simple flow for demonstration
			fmt.Println("\n=== Creating New Flow ===")
			
			timestamp := time.Now().Format("20060102_150405")
			flowTitle := fmt.Sprintf("SDK Example Flow %s", timestamp)
			
			// Simple flow definition that logs a message and returns it
			flowDefinition := map[string]interface{}{
				"title":       flowTitle,
				"description": "An example flow created by the Globus Go SDK",
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
					"Comment": "Simple example flow",
					"StartAt": "LogMessage",
					"States": map[string]interface{}{
						"LogMessage": map[string]interface{}{
							"Type":     "Pass",
							"Result":   "$.message",
							"ResultPath": "$.output",
							"End":      true,
						},
					},
				},
			}
			
			createRequest := &flows.FlowCreateRequest{
				Title:       flowTitle,
				Description: "An example flow created by the Globus Go SDK",
				Definition:  flowDefinition,
			}
			
			newFlow, err := flowsClient.CreateFlow(ctx, createRequest)
			if err != nil {
				log.Fatalf("Failed to create flow: %v", err)
			}
			
			fmt.Printf("Created new flow: %s (%s)\n", newFlow.Title, newFlow.ID)
			flowID = newFlow.ID
		} else {
			// Use the first available flow
			fmt.Printf("Found %d flows:\n", len(flowsList.Flows))
			for i, flow := range flowsList.Flows {
				fmt.Printf("%d. %s (%s)\n", i+1, flow.Title, flow.ID)
			}
			
			flowID = flowsList.Flows[0].ID
			fmt.Printf("\nUsing first flow: %s\n", flowID)
		}
	}

	// Get flow details
	flow, err := flowsClient.GetFlow(ctx, flowID)
	if err != nil {
		log.Fatalf("Failed to get flow: %v", err)
	}
	
	fmt.Printf("\n=== Flow Details ===\n")
	fmt.Printf("ID: %s\n", flow.ID)
	fmt.Printf("Title: %s\n", flow.Title)
	fmt.Printf("Description: %s\n", flow.Description)
	fmt.Printf("Owner: %s\n", flow.FlowOwner)
	fmt.Printf("Created: %s\n", flow.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Run Count: %d\n", flow.RunCount)
	
	// Print input schema if available
	if flow.InputSchema != nil {
		inputSchemaJSON, _ := json.MarshalIndent(flow.InputSchema, "", "  ")
		fmt.Printf("\nInput Schema:\n%s\n", inputSchemaJSON)
	}

	// Run the flow
	fmt.Println("\n=== Running Flow ===")
	
	// Create a run request
	runRequest := &flows.RunRequest{
		FlowID: flowID,
		Label:  "SDK Example Run " + time.Now().Format("20060102_150405"),
		Tags:   []string{"example", "sdk", "go"},
		Input: map[string]interface{}{
			"message": "Hello from Globus Go SDK!",
		},
	}
	
	run, err := flowsClient.RunFlow(ctx, runRequest)
	if err != nil {
		log.Fatalf("Failed to run flow: %v", err)
	}
	
	fmt.Printf("Flow run started with ID: %s\n", run.RunID)
	fmt.Printf("Status: %s\n", run.Status)
	fmt.Printf("Created at: %s\n", run.CreatedAt.Format(time.RFC3339))
	
	// Wait for a few seconds to let the flow run
	fmt.Println("\nWaiting for flow run to complete...")
	time.Sleep(3 * time.Second)
	
	// Get updated run status
	runStatus, err := flowsClient.GetRun(ctx, run.RunID)
	if err != nil {
		log.Fatalf("Failed to get run status: %v", err)
	}
	
	fmt.Printf("\n=== Run Status ===\n")
	fmt.Printf("Status: %s\n", runStatus.Status)
	if runStatus.CompletedAt.IsZero() {
		fmt.Println("Run is still in progress")
	} else {
		fmt.Printf("Completed at: %s\n", runStatus.CompletedAt.Format(time.RFC3339))
		fmt.Printf("Duration: %s\n", runStatus.CompletedAt.Sub(runStatus.CreatedAt))
	}
	
	// Show run output if available
	if runStatus.Output != nil {
		outputJSON, _ := json.MarshalIndent(runStatus.Output, "", "  ")
		fmt.Printf("\nRun Output:\n%s\n", outputJSON)
	}
	
	// Get run logs
	logs, err := flowsClient.GetRunLogs(ctx, run.RunID, 10, 0)
	if err != nil {
		log.Printf("Failed to get run logs: %v", err)
	} else {
		fmt.Printf("\n=== Run Logs (%d entries) ===\n", len(logs.Entries))
		for i, entry := range logs.Entries {
			fmt.Printf("%d. [%s] %s - %s\n", 
				i+1, 
				entry.CreatedAt.Format("15:04:05"),
				entry.Code, 
				entry.Description)
			
			if entry.Details != nil && len(entry.Details) > 0 {
				detailsJSON, _ := json.MarshalIndent(entry.Details, "", "  ")
				fmt.Printf("   Details: %s\n", detailsJSON)
			}
		}
	}
	
	// List action providers
	fmt.Println("\n=== Action Providers ===")
	providers, err := flowsClient.ListActionProviders(ctx, &flows.ListActionProvidersOptions{
		Limit:        5,
		FilterGlobus: true,
	})
	if err != nil {
		log.Printf("Failed to list action providers: %v", err)
	} else {
		fmt.Printf("Found %d Globus action providers:\n", len(providers.ActionProviders))
		for i, provider := range providers.ActionProviders {
			fmt.Printf("%d. %s (%s) - %s\n", 
				i+1, 
				provider.DisplayName, 
				provider.ID,
				provider.Type)
		}
		
		// Show roles for the first provider
		if len(providers.ActionProviders) > 0 {
			provider := providers.ActionProviders[0]
			fmt.Printf("\nRoles for %s:\n", provider.DisplayName)
			
			roles, err := flowsClient.ListActionRoles(ctx, provider.ID, 5, 0)
			if err != nil {
				log.Printf("Failed to list action roles: %v", err)
			} else {
				for i, role := range roles.ActionRoles {
					fmt.Printf("%d. %s (%s)\n", i+1, role.Name, role.ID)
					if role.Description != "" {
						fmt.Printf("   Description: %s\n", role.Description)
					}
				}
			}
		}
	}

	fmt.Println("\nFlows example complete!")
}