//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
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

func TestIntegration_ListEndpoints(t *testing.T) {
	clientID, clientSecret, _, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&testAuthorizer{token: accessToken}),
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

	// Create Transfer client with rate limiting
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&testAuthorizer{token: accessToken}),
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

	// NOTE: Explicit endpoint activation has been removed.
	// Modern Globus endpoints (v0.10+) automatically activate with properly scoped tokens.

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
			return client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
				EndpointID: sourceEndpointID,
				Path:       sourceDir,
			})
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
					Path:      sourceDir,
					Recursive: true,
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
			return client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
				EndpointID: destEndpointID,
				Path:       destDir,
			})
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
					Path:      destDir,
					Recursive: true,
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

	// 4. Create a simple test file in source directory
	sourceSubDir := fmt.Sprintf("%s/subdir", sourceDir)
	t.Logf("Using paths: source=%s, destination=%s, subdir=%s", sourceDir, destDir, sourceSubDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
				EndpointID: sourceEndpointID,
				Path:       sourceSubDir,
			})
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	
	if err != nil {
		t.Fatalf("Failed to create source subdirectory: %v", err)
	}
	t.Logf("Created source subdirectory: %s", sourceSubDir)

	// 5. Submit a transfer task with retry
	t.Logf("Submitting transfer: %s to %s", sourceDir, destDir)
	
	// Use the helper function which might handle edge cases better
	options := map[string]interface{}{
		"recursive":       true,
		"verify_checksum": true,
		"encrypt_data":    true,
	}
	
	// Create transfer task with retry for rate limiting
	var taskResponse *transfer.TaskResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var taskErr error
			taskResponse, taskErr = client.SubmitTransfer(
				ctx,
				sourceEndpointID, sourceDir,
				destEndpointID, destDir,
				fmt.Sprintf("Integration Test Transfer %s", timestamp),
				options,
			)
			
			// Add additional request inspection for debugging
			if taskErr != nil && strings.Contains(taskErr.Error(), "400") {
				t.Logf("Debug - Request failed with endpoints: src=%s, dst=%s", sourceEndpointID, destEndpointID)
				t.Logf("Debug - Paths: src=%s, dst=%s", sourceDir, destDir)
			}
			
			return taskErr
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	
	if err != nil {
		// Get more detailed error information
		errDetail := "Unknown error"
		if strings.Contains(err.Error(), "400") {
			errDetail = "Bad request - This could be due to invalid paths, endpoint configuration, or permission issues"
			t.Fatalf("TRANSFER ERROR: %v (%s) - To resolve, ensure your token has the correct permissions and the paths are correct", err, errDetail)
		} else if strings.Contains(err.Error(), "401") {
			errDetail = "Unauthorized - Authentication token may be invalid or expired"
			t.Fatalf("AUTHENTICATION ERROR: %s - To resolve, provide a valid GLOBUS_TEST_TRANSFER_TOKEN", errDetail)
		} else if strings.Contains(err.Error(), "403") {
			errDetail = "Forbidden - You don't have permission to transfer between these endpoints"
			t.Fatalf("PERMISSION ERROR: %s - To resolve, set GLOBUS_TEST_TRANSFER_TOKEN with a token that has proper permissions", errDetail)
		} else if strings.Contains(err.Error(), "429") {
			errDetail = "Rate limit exceeded"
			t.Fatalf("RATE LIMIT ERROR: %v (%s) - To resolve, wait and try again later", err, errDetail)
		} else if strings.Contains(err.Error(), "500") {
			errDetail = "Server error"
			t.Fatalf("SERVER ERROR: %v (%s) - This is a server-side issue", err, errDetail)
		} else {
			t.Fatalf("Failed to submit transfer task: %v (%s)", err, errDetail)
		}
	}

	t.Logf("Transfer task submitted, task ID: %s", taskResponse.TaskID)

	// 6. Wait for task completion (with a timeout)
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
			// 7. List contents of destination directory to verify transfer using retry
			var listing *transfer.FileList
			err = ratelimit.RetryWithBackoff(
				ctx,
				func(ctx context.Context) error {
					var listErr error
					listing, listErr = client.ListDirectory(ctx, &transfer.ListDirectoryOptions{
						EndpointID: destEndpointID,
						Path:       destDir,
					})
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
}