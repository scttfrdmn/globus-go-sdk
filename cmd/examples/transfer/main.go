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
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Get access token from environment (in a real app, you would get this from auth flow)
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_ACCESS_TOKEN environment variable is required")
	}

	// Create a transfer client with the access token
	transferClient, err := pkg.NewConfig().
		NewTransferClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create transfer client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List endpoints
	fmt.Println("Listing your endpoints...")
	options := &transfer.ListEndpointsOptions{
		FilterScope: "my-endpoints",
		Limit:       10,
	}

	endpoints, err := transferClient.ListEndpoints(ctx, options)
	if err != nil {
		log.Fatalf("Failed to list endpoints: %v", err)
	}

	if len(endpoints.Data) == 0 {
		fmt.Println("No endpoints found.")
		return
	}

	// Print endpoints
	fmt.Printf("Found %d endpoints:\n", len(endpoints.Data))
	for i, ep := range endpoints.Data {
		fmt.Printf("%d. %s (%s) - %s\n", i+1, ep.DisplayName, ep.ID,
			map[bool]string{true: "Activated", false: "Not Activated"}[ep.Activated])
	}

	// Ask user to select source and destination endpoints
	var sourceIdx, destIdx int
	fmt.Print("\nSelect source endpoint (enter number): ")
	fmt.Scanln(&sourceIdx)
	sourceIdx-- // Adjust for 0-based indexing

	if sourceIdx < 0 || sourceIdx >= len(endpoints.Data) {
		log.Fatal("Invalid source endpoint selection")
	}

	fmt.Print("Select destination endpoint (enter number): ")
	fmt.Scanln(&destIdx)
	destIdx-- // Adjust for 0-based indexing

	if destIdx < 0 || destIdx >= len(endpoints.Data) {
		log.Fatal("Invalid destination endpoint selection")
	}

	sourceEndpoint := endpoints.Data[sourceIdx]
	destEndpoint := endpoints.Data[destIdx]

	// NOTE: Explicit endpoint activation has been removed.
	// Modern Globus endpoints (v0.10+) automatically activate with properly scoped tokens.
	// Just ensure your token has the proper permissions for the endpoints.

	fmt.Println("Using endpoints:")
	fmt.Printf("  - Source: %s (%s)\n", sourceEndpoint.DisplayName, sourceEndpoint.ID)
	fmt.Printf("  - Destination: %s (%s)\n", destEndpoint.DisplayName, destEndpoint.ID)

	// Get source and destination paths from user
	var sourcePath, destPath string
	fmt.Print("Enter source path: ")
	fmt.Scanln(&sourcePath)
	fmt.Print("Enter destination path: ")
	fmt.Scanln(&destPath)

	// Submit transfer
	fmt.Println("Submitting transfer task...")
	options2 := map[string]interface{}{
		"recursive":       true,
		"verify_checksum": true,
		"preserve_mtime":  true,
	}

	taskResponse, err := transferClient.SubmitTransfer(
		ctx,
		sourceEndpoint.ID, sourcePath,
		destEndpoint.ID, destPath,
		"SDK Example Transfer",
		options2,
	)

	if err != nil {
		log.Fatalf("Failed to submit transfer: %v", err)
	}

	fmt.Printf("Transfer submitted! Task ID: %s\n", taskResponse.TaskID)
	fmt.Println("You can monitor this transfer task in the Globus web interface.")
}
