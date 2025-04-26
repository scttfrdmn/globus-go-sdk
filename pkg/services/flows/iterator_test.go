// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/core"
)

func TestFlowIterator(t *testing.T) {
	// Set up test server that returns paginated responses
	page := 0
	totalItems := 25
	pageSize := 10
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/flows" {
			t.Errorf("Expected path /flows, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Parse query parameters
		query := r.URL.Query()
		offset, _ := strconv.Atoi(query.Get("offset"))
		
		// Calculate items for this page
		start := offset
		end := start + pageSize
		if end > totalItems {
			end = totalItems
		}
		
		// Create response
		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		flows := make([]Flow, 0, end-start)
		
		for i := start; i < end; i++ {
			flows = append(flows, Flow{
				ID:          "flow-id-" + strconv.Itoa(i),
				Title:       "Flow " + strconv.Itoa(i),
				Description: "Test flow " + strconv.Itoa(i),
				FlowOwner:   "test-user",
				CreatedAt:   flowTime,
				UpdatedAt:   flowTime,
			})
		}
		
		hadMore := end < totalItems
		
		response := FlowList{
			Flows:   flows,
			Total:   totalItems,
			HadMore: hadMore,
			Offset:  offset,
			Limit:   pageSize,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		
		page++
	}))
	defer server.Close()
	
	// Create client
	client := NewClient("test-token", core.WithBaseURL(server.URL+"/"))
	
	// Create iterator
	iterator := client.GetFlowsIterator(&ListFlowsOptions{
		Limit: pageSize,
	})
	
	// Test iteration
	ctx := context.Background()
	count := 0
	
	for iterator.Next(ctx) {
		flow := iterator.Flow()
		if flow == nil {
			t.Errorf("Expected non-nil flow at position %d", count)
		} else {
			expectedID := "flow-id-" + strconv.Itoa(count)
			if flow.ID != expectedID {
				t.Errorf("Expected flow ID %s, got %s", expectedID, flow.ID)
			}
		}
		count++
	}
	
	// Check for errors
	if err := iterator.Err(); err != nil {
		t.Errorf("Iterator returned error: %v", err)
	}
	
	// Verify we got all items
	if count != totalItems {
		t.Errorf("Expected %d items, got %d", totalItems, count)
	}
}

func TestRunIterator(t *testing.T) {
	// Set up test server that returns paginated responses
	page := 0
	totalItems := 15
	pageSize := 5
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runs" {
			t.Errorf("Expected path /runs, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Parse query parameters
		query := r.URL.Query()
		offset, _ := strconv.Atoi(query.Get("offset"))
		
		// Calculate items for this page
		start := offset
		end := start + pageSize
		if end > totalItems {
			end = totalItems
		}
		
		// Create response
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		runs := make([]RunResponse, 0, end-start)
		
		for i := start; i < end; i++ {
			runs = append(runs, RunResponse{
				RunID:     "run-id-" + strconv.Itoa(i),
				FlowID:    "flow-id",
				Status:    "ACTIVE",
				CreatedAt: runTime,
				StartedAt: runTime,
				UserID:    "test-user",
				RunOwner:  "test-user",
			})
		}
		
		hadMore := end < totalItems
		
		response := RunList{
			Runs:    runs,
			Total:   totalItems,
			HadMore: hadMore,
			Offset:  offset,
			Limit:   pageSize,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		
		page++
	}))
	defer server.Close()
	
	// Create client
	client := NewClient("test-token", core.WithBaseURL(server.URL+"/"))
	
	// Create iterator
	iterator := client.GetRunsIterator(&ListRunsOptions{
		Limit: pageSize,
	})
	
	// Test iteration
	ctx := context.Background()
	count := 0
	
	for iterator.Next(ctx) {
		run := iterator.Run()
		if run == nil {
			t.Errorf("Expected non-nil run at position %d", count)
		} else {
			expectedID := "run-id-" + strconv.Itoa(count)
			if run.RunID != expectedID {
				t.Errorf("Expected run ID %s, got %s", expectedID, run.RunID)
			}
		}
		count++
	}
	
	// Check for errors
	if err := iterator.Err(); err != nil {
		t.Errorf("Iterator returned error: %v", err)
	}
	
	// Verify we got all items
	if count != totalItems {
		t.Errorf("Expected %d items, got %d", totalItems, count)
	}
}

func TestRunLogIterator(t *testing.T) {
	runID := "test-run-id"
	totalItems := 12
	pageSize := 5
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/runs/" + runID + "/log"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Parse query parameters
		query := r.URL.Query()
		offset, _ := strconv.Atoi(query.Get("offset"))
		limit, _ := strconv.Atoi(query.Get("limit"))
		
		if limit != pageSize {
			t.Errorf("Expected limit %d, got %d", pageSize, limit)
		}
		
		// Calculate items for this page
		start := offset
		end := start + limit
		if end > totalItems {
			end = totalItems
		}
		
		// Create response
		logTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		entries := make([]RunLogEntry, 0, end-start)
		
		for i := start; i < end; i++ {
			code := "CODE_" + strconv.Itoa(i)
			entries = append(entries, RunLogEntry{
				Code:        code,
				RunID:       runID,
				CreatedAt:   logTime.Add(time.Duration(i) * time.Second),
				Description: "Log entry " + strconv.Itoa(i),
				Details: map[string]interface{}{
					"index": i,
				},
			})
		}
		
		hadMore := end < totalItems
		
		response := RunLogList{
			Entries: entries,
			Total:   totalItems,
			HadMore: hadMore,
			Offset:  offset,
			Limit:   limit,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create client
	client := NewClient("test-token", core.WithBaseURL(server.URL+"/"))
	
	// Create iterator
	iterator := client.GetRunLogsIterator(runID, pageSize)
	
	// Test iteration
	ctx := context.Background()
	count := 0
	
	for iterator.Next(ctx) {
		entry := iterator.LogEntry()
		if entry == nil {
			t.Errorf("Expected non-nil log entry at position %d", count)
		} else {
			expectedCode := "CODE_" + strconv.Itoa(count)
			if entry.Code != expectedCode {
				t.Errorf("Expected log code %s, got %s", expectedCode, entry.Code)
			}
			
			if entry.RunID != runID {
				t.Errorf("Expected run ID %s, got %s", runID, entry.RunID)
			}
			
			index, ok := entry.Details["index"].(float64)
			if !ok || int(index) != count {
				t.Errorf("Expected index %d in details, got %v", count, entry.Details["index"])
			}
		}
		count++
	}
	
	// Check for errors
	if err := iterator.Err(); err != nil {
		t.Errorf("Iterator returned error: %v", err)
	}
	
	// Verify we got all items
	if count != totalItems {
		t.Errorf("Expected %d items, got %d", totalItems, count)
	}
}