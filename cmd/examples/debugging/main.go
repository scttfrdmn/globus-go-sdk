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

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// customLogger adapts a standard log.Logger to the interfaces.Logger interface
type customLogger struct {
	logger *log.Logger
}

func (l *customLogger) Debug(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}

func (l *customLogger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

func (l *customLogger) Warn(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

func (l *customLogger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

func main() {
	// Load environment variables from .env.test file
	_ = godotenv.Load(".env.test")

	// Create a logger to capture debug output
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Get credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")
	
	if clientID == "" || clientSecret == "" {
		fmt.Println("ERROR: Missing client credentials. Set GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
		os.Exit(1)
	}
	
	if sourceEndpointID == "" || destEndpointID == "" {
		fmt.Println("ERROR: Missing endpoints. Set GLOBUS_TEST_SOURCE_ENDPOINT_ID and GLOBUS_TEST_DEST_ENDPOINT_ID")
		os.Exit(1)
	}
	
	// Get token
	fmt.Println("Getting transfer token...")
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create auth client: %v\n", err)
		os.Exit(1)
	}
	
	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), transfer.TransferScope)
	if err != nil {
		fmt.Printf("ERROR: Failed to get token: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Got token with resource server: %s\n", tokenResp.ResourceServer)
	
	// Set up auth
	authorizer := authorizers.StaticTokenCoreAuthorizer(tokenResp.AccessToken)

	// Create a transfer client with debugging enabled
	clientLogger := &customLogger{logger: logger}
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizer),
		transfer.WithHTTPDebugging(true),
		transfer.WithHTTPTracing(true), // Enable detailed tracing
		transfer.WithLogger(clientLogger),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Set a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Create unique test directories
	timestamp := time.Now().Format("20060102_150405")
	testDir := fmt.Sprintf("test-debug-%s", timestamp)
	sourceDir := testDir + "/source"
	destDir := testDir + "/dest"
	
	// Create source directory
	fmt.Printf("Creating source directory: %s\n", sourceDir)
	err = client.Mkdir(ctx, sourceEndpointID, sourceDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create source directory: %v\n", err)
		os.Exit(1)
	}
	
	// Create destination directory
	fmt.Printf("Creating destination directory: %s\n", destDir)
	err = client.Mkdir(ctx, destEndpointID, destDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create destination directory: %v\n", err)
		// Still continue
	}
	
	// Test JSON serialization
	fmt.Println("\nTesting JSON serialization for TransferTaskRequest")
	
	// Create transfer request with uppercase and lowercase JSON data fields
	transferRequestLower := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 fmt.Sprintf("Debug Lower Test %s", timestamp),
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		Items: []transfer.TransferItem{
			{
				SourcePath:      sourceDir,
				DestinationPath: destDir,
				Recursive:       true,
			},
		},
	}
	
	// Print JSON requests
	jsonLower, _ := json.MarshalIndent(transferRequestLower, "", "  ")
	fmt.Printf("Lower case JSON 'data': %s\n\n", string(jsonLower))
	
	// Submit transfer with proper case
	fmt.Println("Submitting transfer task...")
	resp, err := client.CreateTransferTask(ctx, transferRequestLower)
	if err != nil {
		fmt.Printf("ERROR: Transfer task failed: %v\n", err)
	} else {
		fmt.Printf("Transfer task submitted successfully. Task ID: %s\n", resp.TaskID)
	}
	
	// Clean up
	fmt.Println("Cleaning up test directories...")
	
	// Create delete tasks to clean up the directories
	sourceDeleteRequest := &transfer.DeleteTaskRequest{
		DataType:     "delete",
		Label:        "Cleanup Debug Test Directory - Source",
		EndpointID:   sourceEndpointID,
		Items: []transfer.DeleteItem{
			{
				Path:     testDir,
				DataType: "delete_item",
			},
		},
	}
	
	destDeleteRequest := &transfer.DeleteTaskRequest{
		DataType:     "delete",
		Label:        "Cleanup Debug Test Directory - Destination",
		EndpointID:   destEndpointID,
		Items: []transfer.DeleteItem{
			{
				Path:     testDir,
				DataType: "delete_item",
			},
		},
	}
	
	// Submit delete tasks - ignore errors as these are just cleanup operations
	_, _ = client.CreateDeleteTask(ctx, sourceDeleteRequest)
	_, _ = client.CreateDeleteTask(ctx, destDeleteRequest)
	
	fmt.Println("Done.")
}