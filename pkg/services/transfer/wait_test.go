// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
)

func TestWaitForTaskCompletion(t *testing.T) {
	testCases := []struct {
		name           string
		statusSequence []string
		maxCalls       int
		expectError    bool
		finalStatus    string
	}{
		{
			name:           "Task succeeds immediately",
			statusSequence: []string{"SUCCEEDED"},
			maxCalls:       1,
			expectError:    false,
			finalStatus:    "SUCCEEDED",
		},
		{
			name:           "Task eventually succeeds",
			statusSequence: []string{"ACTIVE", "ACTIVE", "SUCCEEDED"},
			maxCalls:       3,
			expectError:    false,
			finalStatus:    "SUCCEEDED",
		},
		{
			name:           "Task fails",
			statusSequence: []string{"ACTIVE", "FAILED"},
			maxCalls:       2,
			expectError:    false,
			finalStatus:    "FAILED",
		},
		{
			name:           "Task is canceled",
			statusSequence: []string{"ACTIVE", "CANCELED"},
			maxCalls:       2,
			expectError:    false,
			finalStatus:    "CANCELED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var callCount int32

			// Create a test server that returns different task statuses
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&callCount, 1)
				index := count - 1
				if int(index) >= len(tc.statusSequence) {
					index = int32(len(tc.statusSequence) - 1)
				}

				// Return a task with the appropriate status
				task := Task{
					TaskID:         "task-123",
					Type:           "TRANSFER",
					Status:         tc.statusSequence[index],
					Label:          "Test Task",
					RequestTime:    time.Now().Add(-time.Minute),
					CompletionTime: timePtr(time.Now()),
					CreatorID:      "test-creator",
					OwnerID:        "test-owner",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(task)
			}))
			defer server.Close()

			// Create a client that uses the test server
			client, err := NewClient(
				WithAuthorizer(mockAuthorizer("test-token")),
				WithCoreOption(core.WithBaseURL(server.URL+"/")),
			)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Set a shorter poll interval for tests
			pollInterval := 10 * time.Millisecond

			// Create a context with a timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Wait for the task to complete
			task, err := client.WaitForTaskCompletion(ctx, "task-123", pollInterval)

			// Check the results
			if tc.expectError && err == nil {
				t.Errorf("Expected an error, but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if err == nil && task.Status != tc.finalStatus {
				t.Errorf("Expected final status %s, but got %s", tc.finalStatus, task.Status)
			}

			// Check that we made the expected number of calls
			if int(atomic.LoadInt32(&callCount)) != tc.maxCalls {
				t.Errorf("Expected %d calls to the server, but got %d", tc.maxCalls, callCount)
			}
		})
	}
}

// Helper function to return a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
