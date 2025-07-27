// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
	// Load environment variables
	_ = godotenv.Load(".env.test")

	// Get credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")

	if clientID == "" || clientSecret == "" {
		fmt.Println("ERROR: Missing client credentials")
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

	// Create transfer client
	authorizer := authorizers.StaticTokenCoreAuthorizer(tokenResp.AccessToken)
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizer),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Create test directories
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("globus-test/test_transfer_%s", timestamp)
	destDir := fmt.Sprintf("globus-test/test_received_%s", timestamp)

	ctx := context.Background()

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
		os.Exit(1)
	}

	// Create a file in source directory to transfer
	subSourceDir := sourceDir + "/subdir"

	fmt.Printf("Creating subdirectory: %s\n", subSourceDir)
	err = client.Mkdir(ctx, sourceEndpointID, subSourceDir)
	if err != nil {
		fmt.Printf("ERROR: Failed to create source subdirectory: %v\n", err)
		os.Exit(1)
	}

	// Generate a valid UUID for submission ID
	submissionID := uuid.New().String()

	// Submit a transfer with a subdirectory that actually exists
	transferRequest := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 fmt.Sprintf("Test Transfer %s", timestamp),
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		SubmissionID:          submissionID,
		Items: []transfer.TransferItem{
			{
				DataType:        "transfer_item",
				SourcePath:      sourceDir,
				DestinationPath: destDir,
				Recursive:       true,
			},
		},
	}

	// Print the request
	jsonBytes, _ := json.MarshalIndent(transferRequest, "", "  ")
	fmt.Printf("Request: %s\n", string(jsonBytes))

	// Try with direct HTTP request to debug the issue
	fmt.Println("\nTrying with direct HTTP request for debugging...")
	directRequest := map[string]interface{}{
		"DATA_TYPE":            "transfer",
		"label":                fmt.Sprintf("Direct Test Transfer %s", timestamp),
		"source_endpoint":      sourceEndpointID,
		"destination_endpoint": destEndpointID,
		"verify_checksum":      true,
		"encrypt_data":         true,
		"submission_id":        submissionID,
		"DATA": []map[string]interface{}{
			{
				"DATA_TYPE":        "transfer_item",
				"source_path":      sourceDir,
				"destination_path": destDir,
				"recursive":        true,
			},
		},
	}

	directJSON, _ := json.Marshal(directRequest)

	req, err := http.NewRequest("POST", "https://transfer.api.globus.org/v0.10/transfer", bytes.NewBuffer(directJSON))
	if err != nil {
		fmt.Printf("ERROR: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	client2 := &http.Client{}
	resp2, err := client2.Do(req)
	if err != nil {
		fmt.Printf("ERROR: Direct HTTP request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	respBody, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Direct response status: %s\n", resp2.Status)
	fmt.Printf("Direct response body: %s\n", string(respBody))

	// Try with a direct submission
	fmt.Println("\nTrying direct submission without submission_id...")
	directReq2 := map[string]interface{}{
		"DATA_TYPE":            "transfer",
		"label":                fmt.Sprintf("Direct Test 2 %s", timestamp),
		"source_endpoint":      sourceEndpointID,
		"destination_endpoint": destEndpointID,
		"verify_checksum":      true,
		"encrypt_data":         true,
		"DATA": []map[string]interface{}{
			{
				"DATA_TYPE":        "transfer_item",
				"source_path":      sourceDir,
				"destination_path": destDir,
				"recursive":        true,
			},
		},
	}

	directJSON2, _ := json.Marshal(directReq2)
	req2, _ := http.NewRequest("POST", "https://transfer.api.globus.org/v0.10/transfer", bytes.NewBuffer(directJSON2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	resp3, _ := client2.Do(req2)
	respBody2, _ := io.ReadAll(resp3.Body)
	fmt.Printf("Second direct response status: %s\n", resp3.Status)
	fmt.Printf("Second direct response body: %s\n\n", string(respBody2))

	// Submit transfer with SDK
	fmt.Println("\nNow trying with SDK...")
	resp, err := client.CreateTransferTask(ctx, transferRequest)
	if err != nil {
		fmt.Printf("ERROR: Transfer failed: %v\n", err)
		fmt.Printf("Debug error type: %T\n", err)

		// Let's try to parse the error response
		if strings.Contains(err.Error(), "400") {
			fmt.Println("\nLooks like a 400 error. Here are possible reasons:")
			fmt.Println("1. Submission ID might be required (but not automatically generated)")
			fmt.Println("2. DATA_TYPE field might be missing or incorrect")
			fmt.Println("3. Source or destination path might be invalid")
			fmt.Println("4. Missing required fields")
		}
	} else {
		fmt.Printf("Transfer submitted successfully. Task ID: %s\n", resp.TaskID)
	}

	// Clean up
	fmt.Println("Cleaning up test directories...")
	// Deletion not actually needed as we'll see the error first
}
