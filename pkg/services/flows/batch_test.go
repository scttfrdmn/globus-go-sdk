// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBatchRunFlows(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runs" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /runs, got %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Parse request body
		var request RunRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create a response with details from the request
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunResponse{
			RunID:     "run-id-" + request.Label, // Use label to create unique ID
			FlowID:    request.FlowID,
			Status:    "ACTIVE",
			CreatedAt: runTime,
			StartedAt: runTime,
			Label:     request.Label,
			Tags:      request.Tags,
			UserID:    "test-user",
			RunOwner:  "test-user",
			Input:     request.Input,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with trailing slash to avoid double-slashes
	serverURL := server.URL
	if serverURL[len(serverURL)-1] != '/' {
		serverURL += "/"
	}
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(serverURL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create batch request
	const batchSize = 5
	requests := make([]*RunRequest, batchSize)
	for i := 0; i < batchSize; i++ {
		requests[i] = &RunRequest{
			FlowID: "test-flow-id",
			Label:  fmt.Sprintf("batch-%d", i),
			Input: map[string]interface{}{
				"param": fmt.Sprintf("value-%d", i),
			},
		}
	}

	batchRequest := &BatchRunFlowsRequest{
		Requests: requests,
		Options: &BatchOptions{
			Concurrency: 2, // Low concurrency to test batching
		},
	}

	// Execute batch
	ctx := context.Background()
	response := client.BatchRunFlows(ctx, batchRequest)

	// Verify results
	if len(response.Responses) != batchSize {
		t.Errorf("Expected %d responses, got %d", batchSize, len(response.Responses))
	}

	for i, result := range response.Responses {
		if result.Error != nil {
			t.Errorf("Result %d had error: %v", i, result.Error)
			continue
		}

		if result.Response == nil {
			t.Errorf("Result %d had nil response", i)
			continue
		}

		expectedLabel := fmt.Sprintf("batch-%d", i)
		if result.Response.Label != expectedLabel {
			t.Errorf("Expected label %s, got %s", expectedLabel, result.Response.Label)
		}

		expectedRunID := "run-id-" + expectedLabel
		if result.Response.RunID != expectedRunID {
			t.Errorf("Expected run ID %s, got %s", expectedRunID, result.Response.RunID)
		}

		expectedParam := fmt.Sprintf("value-%d", i)
		if param, ok := result.Response.Input["param"].(string); !ok || param != expectedParam {
			t.Errorf("Expected param value %s, got %v", expectedParam, result.Response.Input["param"])
		}
	}
}

func TestBatchGetFlows(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For debugging
		t.Logf("Received request: %s %s", r.Method, r.URL.Path)

		// Extract flow ID from path - be more flexible with path matching
		flowID := ""
		path := r.URL.Path
		if path == "/flows/flow-1" {
			flowID = "flow-1"
		} else if path == "/flows/flow-2" {
			flowID = "flow-2"
		} else if path == "/flows/flow-3" {
			flowID = "flow-3"
		} else if path == "/flows/error-id" {
			flowID = "error-id"
		} else {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return flow information or error
		if flowID == "error-id" {
			w.WriteHeader(http.StatusNotFound)
			errorResp := ErrorResponse{
				Code:     "NotFound",
				Message:  "Flow not found",
				Resource: "flow",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(errorResp)
			return
		}

		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Flow{
			ID:          flowID,
			Title:       "Flow " + flowID,
			Description: "Test flow " + flowID,
			FlowOwner:   "test-user",
			CreatedAt:   flowTime,
			UpdatedAt:   flowTime,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with trailing slash to avoid double-slashes
	serverURL := server.URL
	if serverURL[len(serverURL)-1] != '/' {
		serverURL += "/"
	}
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(serverURL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create batch request
	flowIDs := []string{
		"flow-1",
		"flow-2",
		"error-id", // This will result in an error
		"flow-3",
	}

	batchRequest := &BatchFlowsRequest{
		FlowIDs: flowIDs,
		Options: &BatchOptions{
			Concurrency: 2,
		},
	}

	// Execute batch
	ctx := context.Background()
	response := client.BatchGetFlows(ctx, batchRequest)

	// Verify results
	if len(response.Responses) != len(flowIDs) {
		t.Errorf("Expected %d responses, got %d", len(flowIDs), len(response.Responses))
	}

	// Check successful responses
	for i, result := range response.Responses {
		if flowIDs[i] == "error-id" {
			if result.Error == nil {
				t.Errorf("Expected error for flow ID 'error-id', got nil")
			}
			if !IsFlowNotFoundError(result.Error) {
				t.Errorf("Expected FlowNotFoundError, got %T: %v", result.Error, result.Error)
			}
		} else {
			if result.Error != nil {
				t.Errorf("Result %d had unexpected error: %v", i, result.Error)
				continue
			}

			if result.Flow == nil {
				t.Errorf("Result %d had nil flow", i)
				continue
			}

			if result.Flow.ID != flowIDs[i] {
				t.Errorf("Expected flow ID %s, got %s", flowIDs[i], result.Flow.ID)
			}
		}
	}
}

func TestBatchCancelRuns(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract run ID from path
		const cancelPath = "/runs/"
		const cancelSuffix = "/cancel"

		if len(r.URL.Path) <= len(cancelPath)+len(cancelSuffix) ||
			r.URL.Path[:len(cancelPath)] != cancelPath ||
			r.URL.Path[len(r.URL.Path)-len(cancelSuffix):] != cancelSuffix {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		runID := r.URL.Path[len(cancelPath) : len(r.URL.Path)-len(cancelSuffix)]

		// Return success or error based on run ID
		if runID == "error-id" {
			w.WriteHeader(http.StatusNotFound)
			errorResp := ErrorResponse{
				Code:     "NotFound",
				Message:  "Run not found",
				Resource: "run",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(errorResp)
			return
		}

		// Return success
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client with trailing slash to avoid double-slashes
	serverURL := server.URL
	if serverURL[len(serverURL)-1] != '/' {
		serverURL += "/"
	}
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(serverURL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create batch request
	runIDs := []string{
		"run-1",
		"run-2",
		"error-id", // This will result in an error
		"run-3",
	}

	batchRequest := &BatchCancelRunsRequest{
		RunIDs: runIDs,
		Options: &BatchOptions{
			Concurrency: 2,
		},
	}

	// Execute batch
	ctx := context.Background()
	response := client.BatchCancelRuns(ctx, batchRequest)

	// Verify results
	if len(response.Responses) != len(runIDs) {
		t.Errorf("Expected %d responses, got %d", len(runIDs), len(response.Responses))
	}

	// Check responses
	for i, result := range response.Responses {
		if runIDs[i] == "error-id" {
			if result.Error == nil {
				t.Errorf("Expected error for run ID 'error-id', got nil")
			}
			if !IsRunNotFoundError(result.Error) {
				t.Errorf("Expected RunNotFoundError, got %T: %v", result.Error, result.Error)
			}
		} else {
			if result.Error != nil {
				t.Errorf("Result %d had unexpected error: %v", i, result.Error)
			}

			if result.RunID != runIDs[i] {
				t.Errorf("Expected run ID %s, got %s", runIDs[i], result.RunID)
			}
		}
	}
}
