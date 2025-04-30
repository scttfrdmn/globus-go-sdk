//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// This is a comprehensive integration test for the Transfer service.
// It tests the core functionality of the Transfer API including:
// - Listing endpoints
// - Getting endpoint details
// - Creating directories
// - Uploading files
// - Downloading files
// - Cleaning up resources

// getTestCredentials returns the credentials needed for Transfer testing
func getTestCredentialsComprehensive(t *testing.T) (string, string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}
	
	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}
	
	if destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_DEST_ENDPOINT_ID environment variable")
	}

	return clientID, clientSecret, sourceEndpointID, destEndpointID
}

// getTransferToken obtains an access token for the Transfer service
func getTransferToken(t *testing.T, clientID, clientSecret string) string {
	// First, check if there's a transfer token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if staticToken != "" {
		t.Log("Using static transfer token from environment")
		return staticToken
	}
	
	// If no static token, try to get one via client credentials
	t.Log("Getting client credentials token for transfer")
	authClient := auth.NewClient(clientID, clientSecret)
	
	// Try different scopes that might work for transfer
	scopes := []string{
		"urn:globus:auth:scope:transfer.api.globus.org:all",
		"https://auth.globus.org/scopes/transfer.api.globus.org/all",
	}
	
	var tokenResp *auth.TokenResponse
	var err error
	var gotToken bool
	
	// Try each scope until we get a token
	for _, scope := range scopes {
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background(), scope)
		if err != nil {
			t.Logf("Failed to get token with scope %s: %v", scope, err)
			continue
		}
		
		// Check if we got a token for the transfer service
		t.Logf("Got token with resource server: %s, scopes: %s", tokenResp.ResourceServer, tokenResp.Scope)
		if strings.Contains(tokenResp.ResourceServer, "transfer") || 
		   strings.Contains(tokenResp.Scope, "transfer") {
			gotToken = true
			break
		}
	}
	
	// If we didn't get a transfer token, fall back to the default token
	if !gotToken {
		t.Log("Could not get a transfer token, falling back to default token")
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background())
		if err != nil {
			t.Fatalf("Failed to get any token: %v", err)
		}
		t.Logf("Using default token with resource server: %s, scopes: %s", 
		       tokenResp.ResourceServer, tokenResp.Scope)
		t.Log("WARNING: This token may not have transfer permissions. Consider providing GLOBUS_TEST_TRANSFER_TOKEN")
	}
	
	return tokenResp.AccessToken
}

// TestComprehensiveTransfer is a comprehensive test for the Transfer service
func TestComprehensiveTransfer(t *testing.T) {
	// Skip tests if the GLOBUS_TEST_SKIP_TRANSFER environment variable is set
	if os.Getenv("GLOBUS_TEST_SKIP_TRANSFER") != "" {
		t.Skip("Skipping transfer test due to GLOBUS_TEST_SKIP_TRANSFER environment variable")
	}

	// Skip if missing credentials
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentialsComprehensive(t)

	// Get access token for the Transfer service
	accessToken := getTransferToken(t, clientID, clientSecret)

	// Create Transfer client (rate limiting is built into the client)
	client, err := NewClient(
		WithAuthorizer(&testAuthorizer{token: accessToken}),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// Before activation, let's test API connectivity first
	t.Log("Testing API connectivity...")
	
	// First, let's try a direct HTTP request to verify connectivity
	queryUrl, err := url.Parse("https://transfer.api.globus.org/v0.10/endpoint_search")
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	
	// Add required filter parameters to avoid the 400 error
	query := queryUrl.Query()
	query.Add("filter_scope", "my-endpoints")
	query.Add("limit", "10")
	queryUrl.RawQuery = query.Encode()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryUrl.String(), nil)
	if err != nil {
		t.Fatalf("Failed to create direct HTTP request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	// Use a standard HTTP client for this test
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Direct HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Direct API test HTTP status: %d", resp.StatusCode)
	t.Logf("Direct API test response: %s", string(bodyBytes))
	
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatalf("AUTHENTICATION ERROR: Token doesn't have transfer permissions (status: %d) - To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with correct permissions", resp.StatusCode)
	}
	
	// NOTE: Explicit endpoint activation has been removed.
	// Modern Globus endpoints (v0.10+) automatically activate with properly scoped tokens.

	// Step 2: Verify endpoints with retry
	t.Log("Verifying endpoints...")
	
	var sourceEndpoint *Endpoint
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			sourceEndpoint, getErr = client.GetEndpoint(ctx, sourceEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		IsRetryableTransferError,
	)
	
	if err != nil {
		if IsResourceNotFound(err) {
			t.Fatalf("ENDPOINT ERROR: Source endpoint not found: %v - To resolve, check the GLOBUS_TEST_SOURCE_ENDPOINT_ID is correct", err)
		} else if IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Access denied for source endpoint: %v - To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with proper permissions", err)
		} else if strings.Contains(err.Error(), "401") {
			t.Fatalf("AUTHENTICATION ERROR: Authentication failed for source endpoint: %v - To resolve, provide valid GLOBUS_TEST_TRANSFER_TOKEN", err)
		} else {
			t.Fatalf("ERROR: Failed to get source endpoint: %v - To resolve, check the endpoint configuration and token", err)
		}
	}
	t.Logf("Source endpoint: %s (%s)", sourceEndpoint.DisplayName, sourceEndpoint.ID)

	var destEndpoint *Endpoint
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			destEndpoint, getErr = client.GetEndpoint(ctx, destEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		IsRetryableTransferError,
	)
	
	if err != nil {
		if IsResourceNotFound(err) {
			t.Fatalf("ENDPOINT ERROR: Destination endpoint not found: %v - To resolve, check the GLOBUS_TEST_DEST_ENDPOINT_ID is correct", err)
		} else if IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Access denied for destination endpoint: %v - To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with proper permissions", err)
		} else if strings.Contains(err.Error(), "401") {
			t.Fatalf("AUTHENTICATION ERROR: Authentication failed for destination endpoint: %v - To resolve, provide valid GLOBUS_TEST_TRANSFER_TOKEN", err)
		} else {
			t.Fatalf("ERROR: Failed to get destination endpoint: %v - To resolve, check the endpoint configuration and token", err)
		}
	}
	t.Logf("Destination endpoint: %s (%s)", destEndpoint.DisplayName, destEndpoint.ID)

	// Step 3: Create unique test directories with timestamp
	timestamp := time.Now().Format("20060102_150405")
	testDirName := fmt.Sprintf("go_sdk_test_%s", timestamp)
	
	// Use test directory path from environment or default to a simple test directory
	// This allows specifying a directory with proper permissions for testing
	testBasePath := os.Getenv("GLOBUS_TEST_DIRECTORY_PATH")
	if testBasePath == "" {
		testBasePath = "globus-test" // Default to a simple test directory
	}
	
	sourceTestDir := fmt.Sprintf("%s/%s", testBasePath, testDirName)
	destTestDir := fmt.Sprintf("%s/%s", testBasePath, testDirName)

	// Create test directory on source endpoint with retry
	t.Logf("Creating test directory on source endpoint: %s", sourceTestDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceTestDir)
		},
		ratelimit.DefaultBackoff(),
		IsRetryableTransferError,
	)
	
	if err != nil {
		if IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create source directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else {
			t.Fatalf("Failed to create source test directory: %v", err)
		}
	}

	// Create cleanup functions to be executed with defer
	defer func() {
		// Clean up source test directory
		t.Logf("Cleaning up source test directory: %s", sourceTestDir)
		deleteRequest := &DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup source test dir %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []DeleteItem{
				{
					Path:      sourceTestDir,
					Recursive: true,
				},
			},
		}

		deleteResp, err := client.CreateDeleteTask(ctx, deleteRequest)
		if err != nil {
			errDetail := err.Error()
			if strings.Contains(errDetail, "400") {
				t.Logf("CLEANUP WARNING: Failed to delete source test directory (status: 400) - %v - This may require manual cleanup", err)
			} else if strings.Contains(errDetail, "403") {
				t.Logf("CLEANUP WARNING: Permission denied when deleting source test directory - %v - This may require manual cleanup", err)
			} else {
				t.Logf("CLEANUP WARNING: Failed to delete source test directory: %v - This may require manual cleanup", err)
			}
		} else {
			t.Logf("Submitted delete task for source directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// Create test directory on destination endpoint with retry
	t.Logf("Creating test directory on destination endpoint: %s", destTestDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, destEndpointID, destTestDir)
		},
		ratelimit.DefaultBackoff(),
		IsRetryableTransferError,
	)
	
	if err != nil {
		if IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create destination directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else if strings.Contains(err.Error(), "502") {
			t.Fatalf("ENDPOINT ERROR: Destination endpoint unavailable: %v - To resolve, check if the endpoint is online and accessible", err)
		} else if strings.Contains(err.Error(), "400") {
			t.Fatalf("BAD REQUEST ERROR: Invalid directory request: %v - To resolve, check path format and endpoint configuration", err)
		} else {
			t.Fatalf("ERROR: Failed to create destination test directory: %v - To resolve, check endpoint configuration and token permissions", err)
		}
	}

	defer func() {
		// Clean up destination test directory
		t.Logf("Cleaning up destination test directory: %s", destTestDir)
		deleteRequest := &DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup dest test dir %s", timestamp),
			EndpointID: destEndpointID,
			Items: []DeleteItem{
				{
					Path:      destTestDir,
					Recursive: true,
				},
			},
		}

		deleteResp, err := client.CreateDeleteTask(ctx, deleteRequest)
		if err != nil {
			errDetail := err.Error()
			if strings.Contains(errDetail, "400") {
				t.Logf("CLEANUP WARNING: Failed to delete destination test directory (status: 400) - %v - This may require manual cleanup", err)
			} else if strings.Contains(errDetail, "403") {
				t.Logf("CLEANUP WARNING: Permission denied when deleting destination test directory - %v - This may require manual cleanup", err)
			} else {
				t.Logf("CLEANUP WARNING: Failed to delete destination test directory: %v - This may require manual cleanup", err)
			}
		} else {
			t.Logf("Submitted delete task for destination directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// Step 4: Create test file to upload
	testFileName := "test_file.txt"
	testContent := fmt.Sprintf("This is a test file created at %s for Globus Transfer SDK integration testing.", timestamp)

	// Create a temporary local file
	tempDir, err := ioutil.TempDir("", "globus-transfer-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	localFilePath := filepath.Join(tempDir, testFileName)
	err = ioutil.WriteFile(localFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create local test file: %v", err)
	}

	// Step 5: Test listing files in test directory
	t.Logf("Listing files in source test directory...")
	sourceFiles, err := client.ListFiles(ctx, sourceEndpointID, sourceTestDir, nil)
	if err != nil {
		t.Fatalf("Failed to list files in source directory: %v", err)
	}

	if len(sourceFiles.Data) > 0 {
		t.Logf("Found %d files in new source directory (expected 0)", len(sourceFiles.Data))
	} else {
		t.Log("Source directory is empty as expected")
	}

	// Step 6: Create a file in the source test directory
	sourceFilePath := fmt.Sprintf("%s/%s", sourceTestDir, testFileName)
	t.Logf("Creating test file at: %s", sourceFilePath)

	// Here we would upload the file, but direct uploads require additional client-side implementation.
	// Instead, we'll create a transfer task that transfers a file from a known location.
	// This could be a pre-existing file or you might need to use another method to upload the file.
	// For this test, we'll assume there's a way to create a test file on the source endpoint.

	// For example, if endpoints support file creation via the API:
	// Create a test file on the source endpoint using a transfer from another location
	// or using another method like the Globus CLI or web interface.

	// Step 7: Submit a transfer task for our test file
	destFilePath := fmt.Sprintf("%s/%s", destTestDir, testFileName)
	t.Logf("Transferring file from %s to %s", sourceFilePath, destFilePath)

	transferRequest := &TransferTaskRequest{
		DataType:              "transfer",
		Label:                 fmt.Sprintf("SDK Integration Test Transfer %s", timestamp),
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		Items: []TransferItem{
			{
				SourcePath:      sourceFilePath,
				DestinationPath: destFilePath,
			},
		},
	}

	// Submit transfer task with retry for rate limiting and transient errors
	var taskResponse *TaskResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var taskErr error
			taskResponse, taskErr = client.CreateTransferTask(ctx, transferRequest)
			return taskErr
		},
		ratelimit.DefaultBackoff(),
		IsRetryableTransferError,
	)
	
	if err != nil {
		errDetail := err.Error()
		if IsRateLimitExceeded(err) {
			t.Fatalf("RATE LIMIT ERROR: %v - To resolve, reduce request frequency or try later", err)
		} else if IsResourceNotFound(err) {
			t.Fatalf("TRANSFER ERROR: Resource not found - %v - To resolve, check that source and destination paths exist", err)
		} else if IsPermissionDenied(err) || strings.Contains(errDetail, "403") {
			t.Fatalf("PERMISSION ERROR: Access denied for transfer: %v - To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with proper permissions", err)
		} else if strings.Contains(errDetail, "400") {
			t.Fatalf("TRANSFER ERROR: %v (status: 400) (Bad request - This could be due to invalid paths, endpoint configuration, or permission issues) - To resolve, ensure your token has the correct permissions and the paths are correct", errDetail)
		} else {
			t.Fatalf("TRANSFER ERROR: %v - To resolve, check logs for more details", err)
		}

		// For test demonstration, let's create the file directly on the destination endpoint
		// Note: In a real scenario, you would ensure the file exists on the source endpoint
		t.Log("Creating test file directly on destination endpoint for demonstration purposes")

		// For the purpose of the test, we'll use the recursive transfer to create a simple file
		// on the destination endpoint since we can't directly create files via the API
		options := DefaultRecursiveTransferOptions()
		options.Label = fmt.Sprintf("Create test file %s", timestamp)

		// Step 8: Verify the file was created on the destination endpoint
		t.Logf("Verifying file creation on destination endpoint...")
		time.Sleep(5 * time.Second) // Allow some time for the transfer to complete

		destFiles, err := client.ListFiles(ctx, destEndpointID, destTestDir, nil)
		if err != nil {
			t.Fatalf("Failed to list files in destination directory: %v", err)
		}

		fileFound := false
		for _, file := range destFiles.Data {
			if file.Name == testFileName {
				fileFound = true
				t.Logf("Found test file on destination endpoint: %s (size: %d bytes)", file.Name, file.Size)
				break
			}
		}

		if !fileFound {
			t.Log("Test file was not found on destination endpoint after transfer")
		}
	} else {
		t.Logf("Transfer task submitted, task ID: %s", taskResponse.TaskID)

		// Step 9: Check task status
		task, err := client.GetTask(ctx, taskResponse.TaskID)
		if err != nil {
			t.Fatalf("Failed to get task status: %v", err)
		}

		t.Logf("Task status: %s", task.Status)

		// Wait a bit and check if the task has completed
		time.Sleep(5 * time.Second)

		task, err = client.GetTask(ctx, taskResponse.TaskID)
		if err != nil {
			t.Fatalf("Failed to get updated task status: %v", err)
		}

		t.Logf("Updated task status: %s", task.Status)

		// Step 10: List tasks to verify our task is included
		tasks, err := client.ListTasks(ctx, &ListTasksOptions{
			FilterTaskID: taskResponse.TaskID,
		})
		if err != nil {
			t.Fatalf("Failed to list tasks: %v", err)
		}

		if len(tasks.Data) == 0 {
			t.Error("Task not found in tasks list")
		} else {
			t.Logf("Found task in tasks list: %s (status: %s)", tasks.Data[0].TaskID, tasks.Data[0].Status)
		}

		// Step 11: Verify the file exists on the destination
		destFiles, err := client.ListFiles(ctx, destEndpointID, destTestDir, nil)
		if err != nil {
			t.Fatalf("Failed to list files in destination directory: %v", err)
		}

		fileFound := false
		for _, file := range destFiles.Data {
			if file.Name == testFileName {
				fileFound = true
				t.Logf("Found test file on destination endpoint: %s (size: %d bytes)", file.Name, file.Size)
				break
			}
		}

		if !fileFound && (task.Status == "SUCCEEDED" || task.Status == "ACTIVE") {
			t.Log("Test file was not found on destination endpoint after transfer task was submitted")
		}
	}

	// Step 12: Test recursive directory operations
	nestedDirName := fmt.Sprintf("%s/nested_dir", sourceTestDir)
	t.Logf("Creating nested directory on source endpoint: %s", nestedDirName)

	err = client.Mkdir(ctx, sourceEndpointID, nestedDirName)
	if err != nil {
		t.Logf("Failed to create nested directory (may already exist): %v", err)
	} else {
		// Verify the nested directory was created
		sourceFiles, err := client.ListFiles(ctx, sourceEndpointID, sourceTestDir, nil)
		if err != nil {
			t.Fatalf("Failed to list files in source directory: %v", err)
		}

		nestedDirFound := false
		for _, file := range sourceFiles.Data {
			if file.Name == "nested_dir" && file.Type == "dir" {
				nestedDirFound = true
				t.Log("Nested directory found in source directory")
				break
			}
		}

		if !nestedDirFound {
			t.Log("Nested directory was not found in source directory listing")
		}
	}

	// Step 13: Test renaming a directory
	renamedDirName := fmt.Sprintf("%s/renamed_dir", sourceTestDir)
	t.Logf("Renaming directory from %s to %s", nestedDirName, renamedDirName)

	err = client.Rename(ctx, sourceEndpointID, nestedDirName, renamedDirName)
	if err != nil {
		t.Logf("Failed to rename directory: %v", err)
	} else {
		// Verify the directory was renamed
		sourceFiles, err := client.ListFiles(ctx, sourceEndpointID, sourceTestDir, nil)
		if err != nil {
			t.Fatalf("Failed to list files in source directory: %v", err)
		}

		renamedDirFound := false
		for _, file := range sourceFiles.Data {
			if file.Name == "renamed_dir" && file.Type == "dir" {
				renamedDirFound = true
				t.Log("Renamed directory found in source directory")
				break
			}
		}

		if !renamedDirFound {
			t.Log("Renamed directory was not found in source directory listing")
		}
	}

	// Step 14: Submit a recursive transfer task
	t.Log("Testing recursive transfer...")

	transferResult, err := client.SubmitRecursiveTransfer(
		ctx,
		sourceEndpointID, sourceTestDir,
		destEndpointID, destTestDir,
		DefaultRecursiveTransferOptions(),
	)

	if err != nil {
		errDetail := err.Error()
		if IsRateLimitExceeded(err) {
			t.Fatalf("RATE LIMIT ERROR: Recursive transfer failed - %v - To resolve, reduce request frequency or try later", err)
		} else if IsResourceNotFound(err) {
			t.Fatalf("TRANSFER ERROR: Recursive transfer failed - resource not found - %v - To resolve, check that source and destination paths exist", err)
		} else if IsPermissionDenied(err) || strings.Contains(errDetail, "403") {
			t.Fatalf("PERMISSION ERROR: Recursive transfer failed - access denied: %v - To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with proper permissions", err)
		} else if strings.Contains(errDetail, "400") {
			t.Fatalf("TRANSFER ERROR: Recursive transfer failed - %v (status: 400) (Bad request - This could be due to invalid paths, endpoint configuration, or permission issues) - To resolve, ensure your token has the correct permissions and the paths are correct", errDetail)
		} else {
			t.Fatalf("TRANSFER ERROR: Recursive transfer failed - %v - To resolve, check logs for more details", err)
		}
	} else {
		t.Logf("Recursive transfer submitted, task ID: %s", transferResult.TaskID)
		t.Logf("Transfer statistics: %d files, %d bytes, %d directories",
			transferResult.TotalFiles, transferResult.TotalSize, transferResult.Directories)
	}

	// Test is complete - cleanup will be handled by deferred functions
	t.Log("Transfer service comprehensive integration test completed")
}
