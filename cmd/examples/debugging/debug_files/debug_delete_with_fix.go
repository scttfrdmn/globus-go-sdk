// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package debug_files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

// fixAuthorizer implements the authorizer interface for testing
type fixAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *fixAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns whether the authorization is valid
func (a *fixAuthorizer) IsValid() bool {
	return a.token != ""
}

// GetToken returns the token
func (a *fixAuthorizer) GetToken() string {
	return a.token
}

// This represents the DeleteItem structure with the DATA_TYPE field added
// Enhanced delete item struct with explicit DataType
// Note: The API does not support a "recursive" field for delete_item as of API v0.10
type EnhancedDeleteItem struct {
	Path     string `json:"path"`
	DataType string `json:"DATA_TYPE,omitempty"` // DATA_TYPE field is required
}

// Custom DeleteTaskRequest for our test
type EnhancedDeleteTaskRequest struct {
	DataType     string               `json:"DATA_TYPE,omitempty"`
	Label        string               `json:"label,omitempty"`
	EndpointID   string               `json:"endpoint"`
	SubmissionID string               `json:"submission_id,omitempty"`
	Items        []EnhancedDeleteItem `json:"DATA"`
}

func RunDeleteWithFix() {
	// Load environment variables from .env.test file
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load("../../../.env.test")

	// Get credentials from environment
	accessToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if accessToken == "" {
		fmt.Println("ERROR: GLOBUS_TEST_TRANSFER_TOKEN environment variable is required")
		os.Exit(1)
	}

	endpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	if endpointID == "" {
		fmt.Println("ERROR: GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable is required")
		os.Exit(1)
	}

	// Set HTTP_DEBUG environment variable to enable debugging
	os.Setenv("HTTP_DEBUG", "true")

	// Create Transfer client with debugging enabled
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&fixAuthorizer{token: accessToken}),
		transfer.WithHTTPDebugging(true),
		transfer.WithHTTPTracing(true),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create transfer client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Create a test directory to delete
	testPath := fmt.Sprintf("globus-test/debug_delete_fix_%s", time.Now().Format("20060102_150405"))

	fmt.Printf("Creating test directory: %s\n", testPath)
	err = client.Mkdir(ctx, endpointID, testPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to create test directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== MANUAL FIX: ADDING DATA_TYPE TO DELETE ITEMS ===")

	// Get submission ID
	submissionID, err := client.GetSubmissionID(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to get submission ID: %v\n", err)
		os.Exit(1)
	}

	// Create a custom delete task request with DATA_TYPE field for items
	customRequest := EnhancedDeleteTaskRequest{
		DataType:     "delete",
		Label:        "Debug delete with DataType for items",
		EndpointID:   endpointID,
		SubmissionID: submissionID,
		Items: []EnhancedDeleteItem{
			{
				Path:     testPath,
				DataType: "delete_item", // Set DATA_TYPE field for each item
			},
		},
	}

	// Marshal the custom request to JSON
	reqBody, err := json.Marshal(customRequest)
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	// Print the request JSON for inspection
	fmt.Println("\nRequest JSON:")
	fmt.Printf("%s\n", string(reqBody))

	// Create a custom HTTP request to send directly
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://transfer.api.globus.org/v0.10/delete", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("ERROR: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request using the client's transport
	fmt.Println("\nSending custom request with DATA_TYPE field for delete items...")
	resp, err := client.Client.Transport.RoundTrip(req)
	if err != nil {
		fmt.Printf("ERROR: Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR: Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Print response
	fmt.Printf("\nResponse status: %s\n", resp.Status)
	fmt.Printf("Response body: %s\n", string(respBody))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("\nSUCCESS: Delete task created with DATA_TYPE field for items")
	} else {
		fmt.Println("\nERROR: Delete task request failed")
	}

	fmt.Println("\n=== DEBUG COMPLETE ===")
}
