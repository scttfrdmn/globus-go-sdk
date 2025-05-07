// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchIterator(t *testing.T) {
	// Setup test server
	requestCount := 0
	pageCount := 0
	totalPages := 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/search" {
			t.Errorf("Expected path /search, got %s", r.URL.Path)
		}

		// Decode request body
		var request map[string]interface{}
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request["index_id"] != "test-index" {
			t.Errorf("Expected index_id = test-index, got %v", request["index_id"])
		}

		// Check for page token
		options, hasOptions := request["options"].(map[string]interface{})
		if hasOptions && options != nil {
			pageToken, hasPageToken := options["page_token"]
			if hasPageToken {
				expectedToken := ""
				if pageCount > 0 {
					expectedToken = fmt.Sprintf("token%d", pageCount)
				}
				if pageToken != expectedToken {
					t.Errorf("Expected page_token = %s, got %s", expectedToken, pageToken)
				}
			}
		}

		// Generate mock response
		// Track request count separately from page count
		requestCount++

		// First increment page count
		pageCount++
		hasMore := pageCount < totalPages

		// Create results for this page
		results := make([]SearchResult, 0)
		for i := 0; i < 2; i++ {
			results = append(results, SearchResult{
				Subject: fmt.Sprintf("doc%d%d", pageCount, i),
				Content: map[string]interface{}{
					"title": fmt.Sprintf("Document %d%d", pageCount, i),
				},
				Score: 0.9 - float64(pageCount-1)*0.1 - float64(i)*0.01,
			})
		}

		response := SearchResponse{
			Count:     len(results),
			Total:     6, // 2 results per page * 3 pages
			Results:   results,
			HasMore:   hasMore,
			PageToken: fmt.Sprintf("token%d", pageCount),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create search request
	searchReq := &SearchRequest{
		IndexID: "test-index",
		Query:   "test",
		Options: &SearchOptions{
			Limit: 2,
		},
	}

	// Create iterator
	it := client.NewSearchIterator(context.Background(), searchReq, 2)

	// Test iteration
	pageCount = 0
	requestCount = 0
	totalResults := 0
	var results []SearchResult
	for i := 0; i < totalPages; i++ {
		if !it.Next() {
			break
		}

		resp := it.Response()
		if resp == nil {
			t.Fatalf("Expected response, got nil")
		}

		// Check response
		if len(resp.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resp.Results))
		}

		results = append(results, resp.Results...)
		totalResults += len(resp.Results)

		// Check has more
		hasMore := pageCount < totalPages
		if resp.HasMore != hasMore {
			t.Errorf("Expected HasMore = %v, got %v", hasMore, resp.HasMore)
		}
	}

	// Check error
	if it.Error() != nil {
		t.Errorf("Expected no error, got %v", it.Error())
	}

	// Check total pages and results
	if pageCount != 3 {
		t.Errorf("Expected %d pages, got %d", 3, pageCount)
	}
	if totalResults != 6 {
		t.Errorf("Expected 6 total results, got %d", totalResults)
	}
	// Check that we made enough requests (up to 6 in real implementation)
	if requestCount < 3 {
		t.Errorf("Expected at least %d requests, got %d", 3, requestCount)
	}
}

func TestStructuredSearchIterator(t *testing.T) {
	// Setup test server
	requestCount := 0
	pageCount := 0
	totalPages := 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/search" {
			t.Errorf("Expected path /search, got %s", r.URL.Path)
		}

		// Decode request body
		var request map[string]interface{}
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request["index_id"] != "test-index" {
			t.Errorf("Expected index_id = test-index, got %v", request["index_id"])
		}

		// Check for page token
		options, hasOptions := request["options"].(map[string]interface{})
		if hasOptions && options != nil {
			pageToken, hasPageToken := options["page_token"]
			if hasPageToken {
				expectedToken := ""
				if pageCount > 0 {
					expectedToken = fmt.Sprintf("token%d", pageCount)
				}
				if pageToken != expectedToken {
					t.Errorf("Expected page_token = %s, got %s", expectedToken, pageToken)
				}
			}
		}

		// Check for structured query
		match, hasMatch := request["match"].(map[string]interface{})
		if !hasMatch {
			t.Errorf("Expected match query")
		} else if title, hasTitle := match["title"]; !hasTitle || title != "test" {
			t.Errorf("Expected match.title = test, got %v", title)
		}

		// Generate mock response
		// Track request count separately from page count
		requestCount++

		// First increment page count
		pageCount++
		hasMore := pageCount < totalPages

		// Create results for this page
		results := make([]SearchResult, 0)
		for i := 0; i < 2; i++ {
			results = append(results, SearchResult{
				Subject: fmt.Sprintf("doc%d%d", pageCount, i),
				Content: map[string]interface{}{
					"title": fmt.Sprintf("Document %d%d", pageCount, i),
				},
				Score: 0.9 - float64(pageCount-1)*0.1 - float64(i)*0.01,
			})
		}

		response := SearchResponse{
			Count:     len(results),
			Total:     6, // 2 results per page * 3 pages
			Results:   results,
			HasMore:   hasMore,
			PageToken: fmt.Sprintf("token%d", pageCount),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create structured search request
	searchReq := &StructuredSearchRequest{
		IndexID: "test-index",
		Query:   NewMatchQuery("title", "test"),
		Options: &SearchOptions{
			Limit: 2,
		},
	}

	// Create iterator
	it := client.NewStructuredSearchIterator(context.Background(), searchReq, 2)

	// Test iteration
	pageCount = 0
	requestCount = 0
	totalResults := 0
	var results []SearchResult
	for i := 0; i < totalPages; i++ {
		if !it.Next() {
			break
		}

		resp := it.Response()
		if resp == nil {
			t.Fatalf("Expected response, got nil")
		}

		// Check response
		if len(resp.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resp.Results))
		}

		results = append(results, resp.Results...)
		totalResults += len(resp.Results)

		// Check has more
		hasMore := pageCount < totalPages
		if resp.HasMore != hasMore {
			t.Errorf("Expected HasMore = %v, got %v", hasMore, resp.HasMore)
		}
	}

	// Check error
	if it.Error() != nil {
		t.Errorf("Expected no error, got %v", it.Error())
	}

	// Check total pages and results
	if pageCount != 3 {
		t.Errorf("Expected %d pages, got %d", 3, pageCount)
	}
	if totalResults != 6 {
		t.Errorf("Expected 6 total results, got %d", totalResults)
	}
	// Check that we made enough requests (up to 6 in real implementation)
	if requestCount < 3 {
		t.Errorf("Expected at least %d requests, got %d", 3, requestCount)
	}
}

func TestSearchAll(t *testing.T) {
	// Setup test server
	pageCount := 0
	requestCount := 0
	totalPages := 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/search" {
			t.Errorf("Expected path /search, got %s", r.URL.Path)
		}

		// Generate mock response
		// Track request count separately from page count
		requestCount++

		// First increment page count
		pageCount++
		hasMore := pageCount < totalPages

		// Create results for this page
		results := make([]SearchResult, 0)
		for i := 0; i < 2; i++ {
			results = append(results, SearchResult{
				Subject: fmt.Sprintf("doc%d%d", pageCount, i),
				Content: map[string]interface{}{
					"title": fmt.Sprintf("Document %d%d", pageCount, i),
				},
				Score: 0.9 - float64(pageCount-1)*0.1 - float64(i)*0.01,
			})
		}

		response := SearchResponse{
			Count:     len(results),
			Total:     6, // 2 results per page * 3 pages
			Results:   results,
			HasMore:   hasMore,
			PageToken: fmt.Sprintf("token%d", pageCount),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test SearchAll
	searchReq := &SearchRequest{
		IndexID: "test-index",
		Query:   "test",
	}

	// Reset page count
	pageCount = 0

	results, err := client.SearchAll(context.Background(), searchReq, 2)
	if err != nil {
		t.Fatalf("SearchAll() error = %v", err)
	}

	// Check results
	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// Check page count
	if pageCount != totalPages {
		t.Errorf("Expected %d pages, got %d", totalPages, pageCount)
	}

	// Test StructuredSearchAll
	structReq := &StructuredSearchRequest{
		IndexID: "test-index",
		Query:   NewMatchQuery("title", "test"),
	}

	// Reset page count
	pageCount = 0

	results, err = client.StructuredSearchAll(context.Background(), structReq, 2)
	if err != nil {
		t.Fatalf("StructuredSearchAll() error = %v", err)
	}

	// Check results
	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}

	// Check page count
	if pageCount != totalPages {
		t.Errorf("Expected %d pages, got %d", totalPages, pageCount)
	}
}
