// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Set up the SDK configuration from environment variables
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create clients
	authClient, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Get access tokens (using client credentials for simplicity in this example)
	// In a real application, you would likely use the authorization code flow
	ctx := context.Background()

	// Get required scopes for all services
	allScopes := pkg.GetScopesByService("auth", "groups", "transfer", "search", "flows", "compute")

	// Get token using client credentials
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, allScopes...)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	accessToken := tokenResp.AccessToken
	fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)

	// Create service clients using the access token
	groupsClient, err := config.NewGroupsClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create groups client: %v", err)
	}
	transferClient, err := config.NewTransferClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create transfer client: %v", err)
	}
	searchClient, err := config.NewSearchClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create search client: %v", err)
	}
	flowsClient, err := config.NewFlowsClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create flows client: %v", err)
	}
	computeClient, err := config.NewComputeClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}

	// Demonstrate Groups API - List groups
	fmt.Println("\n=== Groups API ===")
	groupsList, err := groupsClient.ListGroups(ctx, &groups.ListGroupsOptions{
		MyGroups: true,
		PageSize: 5,
	})
	if err != nil {
		log.Printf("Failed to list groups: %v", err)
	} else {
		fmt.Printf("Found %d groups:\n", len(groupsList.Groups))
		for i, group := range groupsList.Groups {
			fmt.Printf("%d. %s (%s)\n", i+1, group.Name, group.ID)
		}
	}

	// Demonstrate Transfer API - List endpoints
	fmt.Println("\n=== Transfer API ===")
	endpoints, err := transferClient.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list endpoints: %v", err)
	} else {
		fmt.Printf("Found %d endpoints:\n", len(endpoints.Data))
		for i, endpoint := range endpoints.Data {
			fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
		}
	}

	// Demonstrate a file transfer (only if the user provided endpoint IDs)
	sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("DEST_ENDPOINT_ID")

	if sourceEndpointID != "" && destEndpointID != "" {
		fmt.Println("\n=== Transfer Demonstration ===")

		// Create a test file at the source
		sourcePath := "/~/test_transfer_" + time.Now().Format("20060102_150405") + ".txt"
		destPath := "/~/received_test_file.txt"

		fmt.Printf("Starting transfer from %s to %s\n", sourcePath, destPath)

		// NOTE: Explicit endpoint activation has been removed.
		// Modern Globus endpoints (v0.10+) automatically activate with properly scoped tokens.
		// Just ensure your token has the proper permissions for the endpoints you're using.

		// Submit the transfer
		options := map[string]interface{}{
			"notify_on_succeeded": true,
			"verify_checksum":     true,
		}

		taskResponse, err := transferClient.SubmitTransfer(
			ctx,
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			"SDK Example Transfer",
			options,
		)

		if err != nil {
			log.Printf("Failed to submit transfer: %v", err)
		} else {
			fmt.Printf("Transfer submitted successfully, task ID: %s\n", taskResponse.TaskID)
			fmt.Printf("You can monitor this task using the Globus web interface\n")
		}
	}

	// Demonstrate Search API - List indexes
	fmt.Println("\n=== Search API ===")
	indexes, err := searchClient.ListIndexes(ctx, &search.ListIndexesOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list search indexes: %v", err)
	} else {
		fmt.Printf("Found %d search indexes:\n", len(indexes.Indexes))
		for i, index := range indexes.Indexes {
			fmt.Printf("%d. %s (%s)\n", i+1, index.DisplayName, index.ID)
		}
	}

	// Demonstrate Flows API - List flows
	fmt.Println("\n=== Flows API ===")
	flowsList, err := flowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list flows: %v", err)
	} else {
		fmt.Printf("Found %d flows:\n", len(flowsList.Flows))
		for i, flow := range flowsList.Flows {
			fmt.Printf("%d. %s (%s)\n", i+1, flow.Title, flow.ID)
		}
	}

	// Demonstrate Compute API - List endpoints
	fmt.Println("\n=== Compute API ===")
	compEndpoints, err := computeClient.ListEndpoints(ctx, &compute.ListEndpointsOptions{
		PerPage: 5,
	})
	if err != nil {
		log.Printf("Failed to list compute endpoints: %v", err)
	} else {
		fmt.Printf("Found %d compute endpoints:\n", len(compEndpoints.Endpoints))
		for i, endpoint := range compEndpoints.Endpoints {
			fmt.Printf("%d. %s (%s) - Status: %s\n",
				i+1, endpoint.Name, endpoint.ID, endpoint.Status)
		}
	}

	fmt.Println("\nSDK showcase complete!")
}
