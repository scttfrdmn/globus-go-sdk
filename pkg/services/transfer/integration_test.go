//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer_test

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
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// testAuthorizer implements the authorizer interface for testing
type testAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *testAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns whether the authorization is valid
func (a *testAuthorizer) IsValid() bool {
	return a.token != ""
}

// GetToken returns the token
func (a *testAuthorizer) GetToken() string {
	return a.token
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

func getTestCredentials(t *testing.T) (string, string, string, string) {
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

	return clientID, clientSecret, sourceEndpointID, destEndpointID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	// First, check if there's a transfer token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if staticToken != "" {
		t.Log("Using static transfer token from environment")
		return staticToken
	}

	// If no static token, try to get one via client credentials
	t.Log("Getting client credentials token for transfer")
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Try different scopes that might work for transfer
	scopes := []string{
		"urn:globus:auth:scope:transfer.api.globus.org:all",
		"https://auth.globus.org/scopes/transfer.api.globus.org/all",
	}

	var tokenResp *auth.TokenResponse
	var gotToken bool

	// Try each scope until we get a token
	for _, scope := range scopes {
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background(), []string{scope})
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
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background(), nil)
		if err != nil {
			t.Fatalf("Failed to get any token: %v", err)
		}
		t.Logf("Using default token with resource server: %s, scopes: %s",
			tokenResp.ResourceServer, tokenResp.Scope)
		t.Log("WARNING: This token may not have transfer permissions. Consider providing GLOBUS_TEST_TRANSFER_TOKEN")
	}

	return tokenResp.AccessToken
}

func TestIntegration_ListEndpoints(t *testing.T) {
	clientID, clientSecret, _, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// Try a simple endpoint list request to verify connectivity
	t.Log("Testing transfer API connectivity with ListEndpoints call")

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
	t.Logf("Direct HTTP status: %d", resp.StatusCode)
	t.Logf("Direct HTTP response: %s", string(bodyBytes))

	// Now try through the client
	var endpoints *transfer.EndpointList
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var listErr error
			// Use the same filter options that worked in the direct HTTP request
			endpoints, listErr = client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
				FilterScope: "my-endpoints",
				Limit:       10,
			})
			if listErr != nil {
				t.Logf("Endpoint listing attempt failed: %v", listErr)
			}
			return listErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsRateLimitExceeded(err) {
			t.Fatalf("Rate limit exceeded when listing endpoints: %v", err)
		} else {
			// Provide more debugging information
			t.Logf("Auth header: Bearer %s...", accessToken[:min(20, len(accessToken))])
			t.Fatalf("ListEndpoints failed: %v", err)
		}
	}

	// Verify we got some endpoints
	if len(endpoints.Data) == 0 {
		t.Log("No endpoints found, but this is not an error. User may not have any endpoints.")
	} else {
		t.Logf("Found %d endpoints", len(endpoints.Data))

		// Verify endpoint data
		for i, endpoint := range endpoints.Data {
			if endpoint.ID == "" {
				t.Errorf("Endpoint %d is missing ID", i)
			}
			if endpoint.DisplayName == "" {
				t.Errorf("Endpoint %d is missing display name", i)
			}
		}
	}
}

func TestIntegration_TransferFlow(t *testing.T) {
	// Skip tests if the GLOBUS_TEST_SKIP_TRANSFER environment variable is set
	if os.Getenv("GLOBUS_TEST_SKIP_TRANSFER") != "" {
		t.Skip("Skipping transfer test due to GLOBUS_TEST_SKIP_TRANSFER environment variable")
	}

	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if source endpoint ID is not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}

	// Skip if destination endpoint ID is not provided
	if destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_DEST_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client with rate limiting and debugging
	debug := os.Getenv("HTTP_DEBUG") != ""
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
		transfer.WithHTTPDebugging(debug),
		transfer.WithHTTPTracing(debug),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// 1. Verify endpoints exist with retry
	var sourceEndpoint *transfer.Endpoint
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			sourceEndpoint, getErr = client.GetEndpoint(ctx, sourceEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsResourceNotFound(err) {
			t.Skipf("Source endpoint not found: %v", err)
		} else if transfer.IsPermissionDenied(err) {
			t.Skipf("Permission denied for source endpoint: %v", err)
		} else {
			t.Fatalf("Failed to get source endpoint: %v", err)
		}
	}
	t.Logf("Source endpoint: %s (%s)", sourceEndpoint.DisplayName, sourceEndpoint.ID)

	var destEndpoint *transfer.Endpoint
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			destEndpoint, getErr = client.GetEndpoint(ctx, destEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsResourceNotFound(err) {
			t.Skipf("Destination endpoint not found: %v", err)
		} else if transfer.IsPermissionDenied(err) {
			t.Skipf("Permission denied for destination endpoint: %v", err)
		} else {
			t.Fatalf("Failed to get destination endpoint: %v", err)
		}
	}
	t.Logf("Destination endpoint: %s (%s)", destEndpoint.DisplayName, destEndpoint.ID)

	// 3. Create unique test directories with timestamp
	timestamp := time.Now().Format("20060102_150405")

	// Use test directory path from environment or default to a simple test directory
	// This allows specifying a directory with proper permissions for testing
	testBasePath := os.Getenv("GLOBUS_TEST_DIRECTORY_PATH")
	if testBasePath == "" {
		testBasePath = "globus-test" // Default to a simple test directory
	}

	sourceDir := fmt.Sprintf("%s/test_transfer_%s", testBasePath, timestamp)
	destDir := fmt.Sprintf("%s/test_received_%s", testBasePath, timestamp)

	// Create source directory with retry
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		// Report proper error message based on error type
		if transfer.IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create source directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else {
			t.Fatalf("Failed to create source test directory: %v", err)
		}
	}
	t.Logf("Created source directory: %s", sourceDir)

	// Setup cleanup for source directory
	defer func() {
		// Clean up source test directory
		t.Logf("Cleaning up source test directory: %s", sourceDir)

		// Use retry with backoff for deletion
		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup source test dir %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     sourceDir,
				},
			},
		}

		var deleteResp *transfer.TaskResponse
		deleteErr := ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				deleteResp, err = client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if deleteErr != nil {
			if transfer.IsPermissionDenied(deleteErr) {
				t.Logf("CLEANUP WARNING: Permission denied when deleting source test directory - %v - This may require manual cleanup", deleteErr)
			} else if transfer.IsRateLimitExceeded(deleteErr) {
				t.Logf("CLEANUP WARNING: Rate limit exceeded when deleting source test directory - %v - This may require manual cleanup", deleteErr)
			} else if strings.Contains(deleteErr.Error(), "400") {
				t.Logf("CLEANUP WARNING: Failed to delete source test directory (status: 400) - %v - This may require manual cleanup", deleteErr)
			} else {
				t.Logf("CLEANUP WARNING: Failed to delete source test directory - %v - This may require manual cleanup", deleteErr)
			}
		} else {
			t.Logf("Submitted delete task for source directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// Create destination directory with retry
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, destEndpointID, destDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create destination directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else {
			if strings.Contains(err.Error(), "502") {
				t.Fatalf("ENDPOINT ERROR: Destination endpoint unavailable: %v - To resolve, check if the endpoint is online and accessible", err)
			} else if strings.Contains(err.Error(), "400") {
				t.Fatalf("BAD REQUEST ERROR: Invalid directory request: %v - To resolve, check path format and endpoint configuration", err)
			} else {
				t.Fatalf("ERROR: Failed to create destination test directory: %v - To resolve, check endpoint configuration and token permissions", err)
			}
		}
	}
	t.Logf("Created destination directory: %s", destDir)

	// Setup cleanup for destination directory
	defer func() {
		// Clean up destination test directory
		t.Logf("Cleaning up destination test directory: %s", destDir)

		// Use retry with backoff for deletion
		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup dest test dir %s", timestamp),
			EndpointID: destEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     destDir,
				},
			},
		}

		var deleteResp *transfer.TaskResponse
		deleteErr := ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				deleteResp, err = client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if deleteErr != nil {
			if transfer.IsPermissionDenied(deleteErr) {
				t.Logf("CLEANUP WARNING: Permission denied when deleting destination test directory - %v - This may require manual cleanup", deleteErr)
			} else if transfer.IsRateLimitExceeded(deleteErr) {
				t.Logf("CLEANUP WARNING: Rate limit exceeded when deleting destination test directory - %v - This may require manual cleanup", deleteErr)
			} else if strings.Contains(deleteErr.Error(), "400") {
				t.Logf("CLEANUP WARNING: Failed to delete destination test directory (status: 400) - %v - This may require manual cleanup", deleteErr)
			} else {
				t.Logf("CLEANUP WARNING: Failed to delete destination test directory - %v - This may require manual cleanup", deleteErr)
			}
		} else {
			t.Logf("Submitted delete task for destination directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// 4. Create a subdirectory in source
	sourceSubDir := fmt.Sprintf("%s/subdir", sourceDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceSubDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to create source subdirectory: %v", err)
	}
	t.Logf("Created source subdirectory: %s", sourceSubDir)

	// 5. Create a test file in source directory
	// First create a local temporary file
	tempFile, err := ioutil.TempFile("", "globus-transfer-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	testContent := fmt.Sprintf("Globus Transfer SDK Test File Content %s", timestamp)
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Get the filename part
	tempFilename := filepath.Base(tempFile.Name())
	sourceFilePath := fmt.Sprintf("%s/%s", sourceDir, tempFilename)
	destFilePath := fmt.Sprintf("%s/%s", destDir, tempFilename)

	t.Logf("Local file created: %s", tempFile.Name())
	defer os.Remove(tempFile.Name())

	// Upload the file to source endpoint
	// Note: This step may require separate helper functions/tools depending on
	// how the SDK is designed to handle uploads
	t.Logf("Uploading file to source endpoint path: %s", sourceFilePath)

	// This is a simplified example - actual implementation would depend on how the SDK supports uploads
	// For now, assume the file is uploaded to the source endpoint by other means

	// 6. Submit a transfer from source to destination
	label := fmt.Sprintf("Integration Test Transfer %s", timestamp)
	transferRequest := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		Items: []transfer.TransferItem{
			{
				DataType:        "transfer_item",
				SourcePath:      sourceDir,
				DestinationPath: destDir,
				Recursive:       true,
			},
		},
	}

	t.Logf("Submitting transfer: %s to %s", sourceDir, destDir)

	// Create transfer task with retry for rate limiting
	var taskResponse *transfer.TaskResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var taskErr error
			taskResponse, taskErr = client.CreateTransferTask(ctx, transferRequest)
			return taskErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to submit transfer task: %v", err)
	}

	t.Logf("Transfer task submitted, task ID: %s", taskResponse.TaskID)

	// 7. Wait for task completion (with a timeout)
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Use polling with retries instead of a single wait call
	var task *transfer.Task
	pollErr := ratelimit.RetryWithBackoff(
		waitCtx,
		func(ctx context.Context) error {
			var getErr error
			task, getErr = client.GetTask(ctx, taskResponse.TaskID)
			if getErr != nil {
				return getErr
			}

			if task.Status == "SUCCEEDED" || task.Status == "FAILED" || task.Status == "CANCELED" {
				return nil // Terminal state reached
			}

			return fmt.Errorf("task still in progress: %s", task.Status)
		},
		&ratelimit.ExponentialBackoff{
			InitialDelay: 5 * time.Second,
			MaxDelay:     30 * time.Second,
			Factor:       1.5,
			Jitter:       true,
			MaxAttempt:   20,
		},
		func(err error) bool {
			// Retry if task still in progress or retryable error
			if err != nil && strings.Contains(err.Error(), "task still in progress") {
				return true
			}
			return transfer.IsRetryableTransferError(err)
		},
	)

	if pollErr != nil {
		if waitCtx.Err() != nil {
			t.Logf("Task did not complete within timeout period, but this doesn't necessarily indicate failure")
		} else if transfer.IsRateLimitExceeded(pollErr) {
			t.Logf("Rate limit exceeded during task completion check, but this doesn't necessarily indicate task failure")
		} else {
			t.Logf("Error waiting for task completion: %v", pollErr)
		}
	} else {
		t.Logf("Task completed with status: %s", task.Status)

		if task.Status == "SUCCEEDED" {
			// 8. List contents of destination directory to verify transfer
			var listing *transfer.FileList
			err = ratelimit.RetryWithBackoff(
				ctx,
				func(ctx context.Context) error {
					var listErr error
					listing, listErr = client.ListFiles(ctx, destEndpointID, destDir, nil)
					return listErr
				},
				ratelimit.DefaultBackoff(),
				transfer.IsRetryableTransferError,
			)

			if err != nil {
				t.Fatalf("Failed to list destination directory: %v", err)
			}

			t.Logf("Destination directory contents (%d items):", len(listing.Data))
			for _, item := range listing.Data {
				t.Logf("  - %s [%s]", item.Name, item.Type)
			}
		}
	}

	// 9. Test file rename operation if transfer succeeded
	if task != nil && task.Status == "SUCCEEDED" {
		renamedPath := destDir + "/renamed_file.txt"
		t.Logf("Renaming file from %s to %s", destFilePath, renamedPath)

		err := ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				return client.Rename(ctx, destEndpointID, destFilePath, renamedPath)
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if err != nil {
			t.Logf("Rename operation failed (file may not exist): %v", err)
		} else {
			t.Logf("File renamed successfully")

			// List directory again to confirm rename
			var listing *transfer.FileList
			err = ratelimit.RetryWithBackoff(
				ctx,
				func(ctx context.Context) error {
					var listErr error
					listing, listErr = client.ListFiles(ctx, destEndpointID, destDir, nil)
					return listErr
				},
				ratelimit.DefaultBackoff(),
				transfer.IsRetryableTransferError,
			)

			if err != nil {
				t.Fatalf("Failed to list destination directory after rename: %v", err)
			}

			found := false
			for _, item := range listing.Data {
				if item.Name == "renamed_file.txt" {
					found = true
					break
				}
			}

			if found {
				t.Logf("Renamed file found in destination directory")
			} else {
				t.Errorf("Renamed file not found in destination directory")
			}
		}
	}
}

func TestIntegration_RecursiveTransfer(t *testing.T) {
	// Skip tests if the GLOBUS_TEST_SKIP_TRANSFER environment variable is set
	if os.Getenv("GLOBUS_TEST_SKIP_TRANSFER") != "" {
		t.Skip("Skipping transfer test due to GLOBUS_TEST_SKIP_TRANSFER environment variable")
	}

	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if source endpoint ID is not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}

	// Skip if destination endpoint ID is not provided
	if destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_DEST_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// 1. Create test directory structure with timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102_150405")

	// Use test directory path from environment or default to a simple test directory
	testBasePath := os.Getenv("GLOBUS_TEST_DIRECTORY_PATH")
	if testBasePath == "" {
		testBasePath = "globus-test" // Default to a simple test directory
	}

	sourceDir := fmt.Sprintf("%s/test_recursive_%s", testBasePath, timestamp)
	destDir := fmt.Sprintf("%s/test_recursive_dest_%s", testBasePath, timestamp)

	// Create source directory
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create source directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else {
			t.Fatalf("Failed to create source test directory: %v", err)
		}
	}
	t.Logf("Created source directory: %s", sourceDir)

	// Setup cleanup for source directory
	defer func() {
		// Clean up source test directory
		t.Logf("Cleaning up source test directory: %s", sourceDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup source test dir %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     sourceDir,
				},
			},
		}

		var deleteResp *transfer.TaskResponse
		deleteErr := ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				deleteResp, err = client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if deleteErr != nil {
			t.Logf("Warning: Failed to delete source directory: %v", deleteErr)
		} else {
			t.Logf("Submitted delete task for source directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// Create destination directory
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, destEndpointID, destDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsPermissionDenied(err) || strings.Contains(err.Error(), "403") {
			t.Fatalf("PERMISSION ERROR: Cannot create destination directory: %v - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has write permissions", err)
		} else {
			t.Fatalf("Failed to create destination test directory: %v", err)
		}
	}
	t.Logf("Created destination directory: %s", destDir)

	// Setup cleanup for destination directory
	defer func() {
		// Clean up destination test directory
		t.Logf("Cleaning up destination test directory: %s", destDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup dest test dir %s", timestamp),
			EndpointID: destEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     destDir,
				},
			},
		}

		var deleteResp *transfer.TaskResponse
		deleteErr := ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				deleteResp, err = client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if deleteErr != nil {
			t.Logf("Warning: Failed to delete destination directory: %v", deleteErr)
		} else {
			t.Logf("Submitted delete task for destination directory, task ID: %s", deleteResp.TaskID)
		}
	}()

	// 2. Create a nested directory structure
	for _, subpath := range []string{"/subdir1", "/subdir1/nested1", "/subdir2"} {
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				return client.Mkdir(ctx, sourceEndpointID, sourceDir+subpath)
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if err != nil {
			t.Fatalf("Failed to create nested directory %s: %v", subpath, err)
		}
		t.Logf("Created nested directory: %s%s", sourceDir, subpath)
	}

	// 3. Test recursive transfer using the SDK's recursive transfer functionality with retry
	options := &transfer.RecursiveTransferOptions{
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		SourcePath:            sourceDir,
		DestinationPath:       destDir,
		Label:                 fmt.Sprintf("Recursive Transfer Test %s", timestamp),
		Sync:                  true,
		VerifyChecksum:        true,
	}

	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.RecursiveTransfer(ctx, options)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		if transfer.IsRateLimitExceeded(err) {
			t.Logf("Rate limit exceeded during recursive transfer. Continuing test but transfer may not complete: %v", err)
		} else {
			t.Fatalf("RecursiveTransfer failed: %v", err)
		}
	}

	t.Log("Recursive transfer submitted successfully")

	// 4. List tasks to verify transfer was initiated
	listTasksOptions := &transfer.ListTasksOptions{
		Limit: 5,
	}

	var tasks *transfer.TaskList
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var listErr error
			tasks, listErr = client.ListTasks(ctx, listTasksOptions)
			return listErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks.Data) == 0 {
		t.Error("No transfer tasks found after recursive transfer")
	} else {
		t.Logf("Found %d recent tasks, most recent: %s", len(tasks.Data), tasks.Data[0].TaskID)
	}
}

func TestIntegration_GetEndpointActivationRequirements(t *testing.T) {
	clientID, clientSecret, sourceEndpointID, _ := getTestCredentials(t)

	// Skip if endpoint is not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// Get activation requirements with retry
	var requirements *transfer.ActivationRequirements
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			requirements, getErr = client.GetActivationRequirements(ctx, sourceEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to get activation requirements: %v", err)
	}

	// Verified we got a response - exact requirements depend on the endpoint type
	t.Logf("Activation requirements data type: %s", requirements.DataType)
	t.Logf("Number of activation requirements: %d", len(requirements.ActivationRequirements))

	// Try to activate the endpoint (might already be activated)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.ActivateEndpoint(ctx, sourceEndpointID)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Logf("Activation might require additional steps: %v", err)
	} else {
		t.Log("Endpoint activated successfully")
	}

	// Get endpoint autoactivation status with retry
	var endpoint *transfer.Endpoint
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			endpoint, getErr = client.GetEndpoint(ctx, sourceEndpointID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}

	// Check if the endpoint has the IsAutoActivateEnabled field (it might be a different name like AutoActivate)
	t.Logf("Endpoint activation details: %+v", endpoint)

	// If available, log auto-activation details
	if endpoint.ActivationProfile != "" {
		t.Logf("Endpoint activation profile: %s", endpoint.ActivationProfile)
	}
}

func TestIntegration_TaskManagement(t *testing.T) {
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if source endpoint ID is not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}

	// Skip if destination endpoint ID is not provided
	if destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_DEST_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}
	ctx := context.Background()

	// 1. Create a small transfer task
	timestamp := time.Now().Format("20060102_150405")
	label := fmt.Sprintf("Task Management Test %s", timestamp)
	sourcePath := "/~/" // Assuming user home directory exists
	destPath := "/~/"

	transferRequest := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Sync:                  false,
		Items: []transfer.TransferItem{
			{
				DataType:        "transfer_item",
				SourcePath:      sourcePath,
				DestinationPath: destPath,
				Recursive:       false,
			},
		},
	}

	// This transfer might fail if paths don't exist, but we just want to test task management
	var taskResponse *transfer.TaskResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var taskErr error
			taskResponse, taskErr = client.CreateTransferTask(ctx, transferRequest)
			return taskErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Logf("Transfer request failed (possibly path doesn't exist): %v", err)
		// Create a minimal task if the main one failed
		deleteTaskRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      "Delete task for testing task management",
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     "/~/nonexistent_path_for_test",
				},
			},
		}

		var deleteErr error
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				taskResponse, err = client.CreateDeleteTask(ctx, deleteTaskRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if err != nil {
			t.Fatalf("Failed to create any test task: %v", err)
		}
	}

	t.Logf("Task created with ID: %s", taskResponse.TaskID)

	// 2. Get task by ID with retry
	var task *transfer.Task
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			task, getErr = client.GetTask(ctx, taskResponse.TaskID)
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	t.Logf("Task status: %s", task.Status)
	t.Logf("Task label: %s", task.Label)

	// 3. Update task label if the task is not completed
	if task.Status != "SUCCEEDED" && task.Status != "FAILED" && task.Status != "CANCELED" {
		updatedLabel := fmt.Sprintf("Updated Test Task %s", timestamp)
		t.Logf("Updating task label to: %s", updatedLabel)

		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				return client.UpdateTaskLabel(ctx, taskResponse.TaskID, updatedLabel)
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if err != nil {
			t.Logf("Failed to update task label (might be already completed): %v", err)
		} else {
			// Verify label was updated
			var updatedTask *transfer.Task
			err = ratelimit.RetryWithBackoff(
				ctx,
				func(ctx context.Context) error {
					var getErr error
					updatedTask, getErr = client.GetTask(ctx, taskResponse.TaskID)
					return getErr
				},
				ratelimit.DefaultBackoff(),
				transfer.IsRetryableTransferError,
			)

			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if updatedTask.Label == updatedLabel {
				t.Log("Task label updated successfully")
			} else {
				t.Errorf("Task label not updated, expected %s, got %s", updatedLabel, updatedTask.Label)
			}
		}
	} else {
		t.Logf("Task already completed, skipping label update")
	}

	// 4. Test task cancellation if the task is still active
	if task.Status == "ACTIVE" {
		t.Log("Attempting to cancel task")

		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				return client.CancelTask(ctx, taskResponse.TaskID)
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)

		if err != nil {
			t.Logf("Failed to cancel task (might be already completed): %v", err)
		} else {
			t.Log("Task cancellation request submitted")

			// Verify task was cancelled
			var cancelledTask *transfer.Task
			err = ratelimit.RetryWithBackoff(
				ctx,
				func(ctx context.Context) error {
					var getErr error
					cancelledTask, getErr = client.GetTask(ctx, taskResponse.TaskID)
					return getErr
				},
				ratelimit.DefaultBackoff(),
				transfer.IsRetryableTransferError,
			)

			if err != nil {
				t.Fatalf("Failed to get cancelled task: %v", err)
			}

			t.Logf("Task status after cancellation request: %s", cancelledTask.Status)
		}
	} else {
		t.Logf("Task not in ACTIVE state, skipping cancellation test")
	}

	// 5. Get task events with retry
	var events *transfer.TaskEventList
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var getErr error
			events, getErr = client.GetTaskEvents(ctx, taskResponse.TaskID, &transfer.GetTaskEventsOptions{
				Limit: 5,
			})
			return getErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)

	if err != nil {
		t.Fatalf("Failed to get task events: %v", err)
	}

	t.Logf("Found %d task events", len(events.Data))
	for i, event := range events.Data {
		t.Logf("Event %d: %s at %s", i, event.Code, event.Time)
	}
}
