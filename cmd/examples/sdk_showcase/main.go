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
)

func main() {
	// Set up the SDK configuration from environment variables
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create clients
	authClient := config.NewAuthClient()

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
	fmt.Printf("Obtained access token (expires in %d seconds)
", tokenResp.ExpiresIn)

	// Create service clients using the access token
	groupsClient := config.NewGroupsClient(accessToken)
	transferClient := config.NewTransferClient(accessToken)
	searchClient := config.NewSearchClient(accessToken)
	flowsClient := config.NewFlowsClient(accessToken)
	computeClient := config.NewComputeClient(accessToken)

	// Demonstrate Groups API - List groups
	fmt.Println("
=== Groups API ===")
	groups, err := groupsClient.ListGroups(ctx, &pkg.ListGroupsOptions{
		MyGroups: true,
		PageSize: 5,
	})
	if err != nil {
		log.Printf("Failed to list groups: %v", err)
	} else {
		fmt.Printf("Found %d groups:
", len(groups.Groups))
		for i, group := range groups.Groups {
			fmt.Printf("%d. %s (%s)
", i+1, group.Name, group.ID)
		}
	}

	// Demonstrate Transfer API - List endpoints
	fmt.Println("
=== Transfer API ===")
	endpoints, err := transferClient.ListEndpoints(ctx, &pkg.ListEndpointsOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list endpoints: %v", err)
	} else {
		fmt.Printf("Found %d endpoints:
", len(endpoints.DATA))
		for i, endpoint := range endpoints.DATA {
			fmt.Printf("%d. %s (%s)
", i+1, endpoint.DisplayName, endpoint.ID)
		}
	}

	// Demonstrate a file transfer (only if the user provided endpoint IDs)
	sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("DEST_ENDPOINT_ID")
	
	if sourceEndpointID != "" && destEndpointID != "" {
		fmt.Println("
=== Transfer Demonstration ===")
		
		// Create a test file at the source
		sourcePath := "/~/test_transfer_" + time.Now().Format("20060102_150405") + ".txt"
		destPath := "/~/received_test_file.txt"
		
		fmt.Printf("Starting transfer from %s to %s
", sourcePath, destPath)
		
		// Activate the endpoints
		if err := transferClient.ActivateEndpoint(ctx, sourceEndpointID); err != nil {
			log.Printf("Warning: Failed to activate source endpoint: %v", err)
		}
		
		if err := transferClient.ActivateEndpoint(ctx, destEndpointID); err != nil {
			log.Printf("Warning: Failed to activate destination endpoint: %v", err)
		}
		
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
			fmt.Printf("Transfer submitted successfully, task ID: %s
", taskResponse.TaskID)
			fmt.Printf("You can monitor this task using the Globus web interface
")
		}
	}
	
	// Demonstrate Search API - List indexes
	fmt.Println("
=== Search API ===")
	indexes, err := searchClient.ListIndexes(ctx, &pkg.ListIndexesOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list search indexes: %v", err)
	} else {
		fmt.Printf("Found %d search indexes:
", len(indexes.Indexes))
		for i, index := range indexes.Indexes {
			fmt.Printf("%d. %s (%s)
", i+1, index.DisplayName, index.ID)
		}
	}
	
	// Demonstrate Flows API - List flows
	fmt.Println("
=== Flows API ===")
	flows, err := flowsClient.ListFlows(ctx, &pkg.ListFlowsOptions{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list flows: %v", err)
	} else {
		fmt.Printf("Found %d flows:
", len(flows.Flows))
		for i, flow := range flows.Flows {
			fmt.Printf("%d. %s (%s)
", i+1, flow.Title, flow.ID)
		}
	}
	
	// Demonstrate Compute API - List endpoints
	fmt.Println("
=== Compute API ===")
	compEndpoints, err := computeClient.ListEndpoints(ctx, &pkg.ListEndpointsOptions{
		PerPage: 5,
	})
	if err != nil {
		log.Printf("Failed to list compute endpoints: %v", err)
	} else {
		fmt.Printf("Found %d compute endpoints:
", len(compEndpoints.Endpoints))
		for i, endpoint := range compEndpoints.Endpoints {
			fmt.Printf("%d. %s (%s) - Status: %s
", 
				i+1, endpoint.Name, endpoint.ID, endpoint.Status)
		}
	}
	
	fmt.Println("
SDK showcase complete!")
}
