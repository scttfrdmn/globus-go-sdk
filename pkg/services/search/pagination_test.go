// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// paginatedSearchServer returns a mock that serves totalPages pages of 2 gmeta
// results each on the index-scoped POST search path. It counts requests to
// decide when to signal the last page, mirroring offset-based iteration.
func paginatedSearchServer(t *testing.T, totalPages int) *httptest.Server {
	t.Helper()
	var calls int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/index/test-index/search" {
			t.Errorf("Expected /index/test-index/search, got %s", r.URL.Path)
		}
		page := int(atomic.AddInt32(&calls, 1)) - 1
		results := []SearchResult{
			{Subject: fmt.Sprintf("doc%d0", page)},
			{Subject: fmt.Sprintf("doc%d1", page)},
		}
		resp := SearchResponse{
			Count:       len(results),
			Total:       totalPages * 2,
			Offset:      page * 2,
			GMeta:       results,
			HasNextPage: page+1 < totalPages,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestSearchIterator(t *testing.T) {
	totalPages := 3
	server := paginatedSearchServer(t, totalPages)
	defer server.Close()

	client, err := NewClient(WithAccessToken("test-token"), WithBaseURL(server.URL+"/"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	searchReq := &SearchRequest{IndexID: "test-index", Query: "test", Options: &SearchOptions{Limit: 2}}
	it := client.NewSearchIterator(context.Background(), searchReq, 2)

	pages, totalResults := 0, 0
	for it.Next() {
		resp := it.Response()
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		if len(resp.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resp.Results))
		}
		totalResults += len(resp.Results)
		pages++
	}
	if it.Error() != nil {
		t.Errorf("Expected no error, got %v", it.Error())
	}
	if pages != totalPages {
		t.Errorf("Expected %d pages, got %d", totalPages, pages)
	}
	if totalResults != totalPages*2 {
		t.Errorf("Expected %d total results, got %d", totalPages*2, totalResults)
	}
}

func TestSearchAll(t *testing.T) {
	server := paginatedSearchServer(t, 3)
	defer server.Close()

	client, err := NewClient(WithAccessToken("test-token"), WithBaseURL(server.URL+"/"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	results, err := client.SearchAll(context.Background(),
		&SearchRequest{IndexID: "test-index", Query: "test"}, 2)
	if err != nil {
		t.Fatalf("SearchAll() error = %v", err)
	}
	if len(results) != 6 {
		t.Errorf("Expected 6 results, got %d", len(results))
	}
}
