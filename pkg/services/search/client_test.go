// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/core"
)

// Test mock server
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	
	// Create a client that uses the test server
	client := NewClient("test-token", 
		core.WithBaseURL(server.URL+"/"),
	)
	
	return server, client
}

func TestListIndexes(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/index_list" {
			t.Errorf("Expected path /index_list, got %s", r.URL.Path)
		}
		
		// Check query parameters
		queryParams := r.URL.Query()
		if limit := queryParams.Get("limit"); limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
		}
		
		// Return mock response
		indexTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := IndexList{
			Indexes: []Index{
				{
					ID:          "test-index-id",
					DisplayName: "Test Index",
					Description: "A test index",
					IsActive:    true,
					IsPublic:    false,
					CreatedBy:   "test-user",
					CreatedAt:   indexTime,
					UpdatedAt:   indexTime,
				},
			},
			Total:     1,
			HadErrors: false,
			HasMore:   false,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test list indexes
	options := &ListIndexesOptions{
		Limit: 10,
	}
	
	indexList, err := client.ListIndexes(context.Background(), options)
	if err != nil {
		t.Fatalf("ListIndexes() error = %v", err)
	}
	
	// Check response
	if len(indexList.Indexes) != 1 {
		t.Errorf("Expected 1 index, got %d", len(indexList.Indexes))
	}
	
	index := indexList.Indexes[0]
	if index.ID != "test-index-id" {
		t.Errorf("Expected index ID = test-index-id, got %s", index.ID)
	}
	if index.DisplayName != "Test Index" {
		t.Errorf("Expected display name = Test Index, got %s", index.DisplayName)
	}
}

func TestGetIndex(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/index/test-index-id" {
			t.Errorf("Expected path /index/test-index-id, got %s", r.URL.Path)
		}
		
		// Return mock response
		indexTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Index{
			ID:          "test-index-id",
			DisplayName: "Test Index",
			Description: "A test index",
			IsActive:    true,
			IsPublic:    false,
			CreatedBy:   "test-user",
			CreatedAt:   indexTime,
			UpdatedAt:   indexTime,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test get index
	index, err := client.GetIndex(context.Background(), "test-index-id")
	if err != nil {
		t.Fatalf("GetIndex() error = %v", err)
	}
	
	// Check response
	if index.ID != "test-index-id" {
		t.Errorf("Expected index ID = test-index-id, got %s", index.ID)
	}
	if index.DisplayName != "Test Index" {
		t.Errorf("Expected display name = Test Index, got %s", index.DisplayName)
	}
	
	// Test empty index ID
	_, err = client.GetIndex(context.Background(), "")
	if err == nil {
		t.Error("GetIndex() with empty ID should return error")
	}
}

func TestCreateIndex(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/index" {
			t.Errorf("Expected path /index, got %s", r.URL.Path)
		}
		
		// Decode request body
		var request IndexCreateRequest
		json.NewDecoder(r.Body).Decode(&request)
		
		// Check request body
		if request.DisplayName != "New Test Index" {
			t.Errorf("Expected display name = New Test Index, got %s", request.DisplayName)
		}
		
		// Return mock response
		indexTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Index{
			ID:          "new-test-index-id",
			DisplayName: request.DisplayName,
			Description: request.Description,
			IsActive:    true,
			IsPublic:    false,
			CreatedBy:   "test-user",
			CreatedAt:   indexTime,
			UpdatedAt:   indexTime,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test create index
	createRequest := &IndexCreateRequest{
		DisplayName: "New Test Index",
		Description: "A new test index",
	}
	
	index, err := client.CreateIndex(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("CreateIndex() error = %v", err)
	}
	
	// Check response
	if index.ID != "new-test-index-id" {
		t.Errorf("Expected index ID = new-test-index-id, got %s", index.ID)
	}
	if index.DisplayName != "New Test Index" {
		t.Errorf("Expected display name = New Test Index, got %s", index.DisplayName)
	}
	
	// Test nil request
	_, err = client.CreateIndex(context.Background(), nil)
	if err == nil {
		t.Error("CreateIndex() with nil request should return error")
	}
	
	// Test empty display name
	_, err = client.CreateIndex(context.Background(), &IndexCreateRequest{})
	if err == nil {
		t.Error("CreateIndex() with empty display name should return error")
	}
}

func TestUpdateIndex(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/index/test-index-id" {
			t.Errorf("Expected path /index/test-index-id, got %s", r.URL.Path)
		}
		
		// Decode request body
		var request IndexUpdateRequest
		json.NewDecoder(r.Body).Decode(&request)
		
		// Check request body
		if request.DisplayName != "Updated Test Index" {
			t.Errorf("Expected display name = Updated Test Index, got %s", request.DisplayName)
		}
		
		// Return mock response
		indexTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Index{
			ID:          "test-index-id",
			DisplayName: request.DisplayName,
			Description: "Updated description",
			IsActive:    true,
			IsPublic:    false,
			CreatedBy:   "test-user",
			CreatedAt:   indexTime,
			UpdatedAt:   time.Now(),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test update index
	updateRequest := &IndexUpdateRequest{
		DisplayName: "Updated Test Index",
		Description: "Updated description",
	}
	
	index, err := client.UpdateIndex(context.Background(), "test-index-id", updateRequest)
	if err != nil {
		t.Fatalf("UpdateIndex() error = %v", err)
	}
	
	// Check response
	if index.ID != "test-index-id" {
		t.Errorf("Expected index ID = test-index-id, got %s", index.ID)
	}
	if index.DisplayName != "Updated Test Index" {
		t.Errorf("Expected display name = Updated Test Index, got %s", index.DisplayName)
	}
	
	// Test empty index ID
	_, err = client.UpdateIndex(context.Background(), "", updateRequest)
	if err == nil {
		t.Error("UpdateIndex() with empty ID should return error")
	}
	
	// Test nil request
	_, err = client.UpdateIndex(context.Background(), "test-index-id", nil)
	if err == nil {
		t.Error("UpdateIndex() with nil request should return error")
	}
}

func TestDeleteIndex(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/index/test-index-id" {
			t.Errorf("Expected path /index/test-index-id, got %s", r.URL.Path)
		}
		
		// Return success response
		w.WriteHeader(http.StatusNoContent)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test delete index
	err := client.DeleteIndex(context.Background(), "test-index-id")
	if err != nil {
		t.Fatalf("DeleteIndex() error = %v", err)
	}
	
	// Test empty index ID
	err = client.DeleteIndex(context.Background(), "")
	if err == nil {
		t.Error("DeleteIndex() with empty ID should return error")
	}
}

func TestSearch(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/search" {
			t.Errorf("Expected path /search, got %s", r.URL.Path)
		}
		
		// Decode request body
		var request SearchRequest
		json.NewDecoder(r.Body).Decode(&request)
		
		// Check request body
		if request.IndexID != "test-index-id" {
			t.Errorf("Expected index ID = test-index-id, got %s", request.IndexID)
		}
		if request.Query != "test query" {
			t.Errorf("Expected query = test query, got %s", request.Query)
		}
		
		// Return mock response
		response := SearchResponse{
			Count:     2,
			Total:     2,
			Subjects:  []string{"subject1", "subject2"},
			Results: []SearchResult{
				{
					Subject: "subject1",
					Content: map[string]interface{}{
						"title": "Result 1",
						"data":  "Content 1",
					},
					Score: 0.95,
				},
				{
					Subject: "subject2",
					Content: map[string]interface{}{
						"title": "Result 2",
						"data":  "Content 2",
					},
					Score: 0.85,
				},
			},
			HadErrors: false,
			HasMore:   false,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test search
	searchRequest := &SearchRequest{
		IndexID: "test-index-id",
		Query:   "test query",
		Options: &SearchOptions{
			Limit: 10,
		},
	}
	
	searchResponse, err := client.Search(context.Background(), searchRequest)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	
	// Check response
	if searchResponse.Count != 2 {
		t.Errorf("Expected count = 2, got %d", searchResponse.Count)
	}
	if len(searchResponse.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(searchResponse.Results))
	}
	
	// Test nil request
	_, err = client.Search(context.Background(), nil)
	if err == nil {
		t.Error("Search() with nil request should return error")
	}
	
	// Test empty index ID
	_, err = client.Search(context.Background(), &SearchRequest{Query: "test"})
	if err == nil {
		t.Error("Search() with empty index ID should return error")
	}
}

func TestIngestDocuments(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/ingest" {
			t.Errorf("Expected path /ingest, got %s", r.URL.Path)
		}
		
		// Decode request body
		var request IngestRequest
		json.NewDecoder(r.Body).Decode(&request)
		
		// Check request body
		if request.IndexID != "test-index-id" {
			t.Errorf("Expected index ID = test-index-id, got %s", request.IndexID)
		}
		if len(request.Documents) != 2 {
			t.Errorf("Expected 2 documents, got %d", len(request.Documents))
		}
		
		// Return mock response
		response := IngestResponse{
			Task: IngestTask{
				TaskID:          "test-task-id",
				ProcessingState: "SUCCESS",
				CreatedAt:       time.Now().Format(time.RFC3339),
				CompletedAt:     time.Now().Format(time.RFC3339),
			},
			Succeeded: 2,
			Failed:    0,
			Total:     2,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test ingest documents
	ingestRequest := &IngestRequest{
		IndexID: "test-index-id",
		Documents: []SearchDocument{
			{
				Subject: "subject1",
				Content: map[string]interface{}{
					"title": "Document 1",
					"data":  "Content 1",
				},
			},
			{
				Subject: "subject2",
				Content: map[string]interface{}{
					"title": "Document 2",
					"data":  "Content 2",
				},
			},
		},
	}
	
	ingestResponse, err := client.IngestDocuments(context.Background(), ingestRequest)
	if err != nil {
		t.Fatalf("IngestDocuments() error = %v", err)
	}
	
	// Check response
	if ingestResponse.Task.TaskID != "test-task-id" {
		t.Errorf("Expected task ID = test-task-id, got %s", ingestResponse.Task.TaskID)
	}
	if ingestResponse.Succeeded != 2 {
		t.Errorf("Expected succeeded = 2, got %d", ingestResponse.Succeeded)
	}
	
	// Test nil request
	_, err = client.IngestDocuments(context.Background(), nil)
	if err == nil {
		t.Error("IngestDocuments() with nil request should return error")
	}
	
	// Test empty index ID
	_, err = client.IngestDocuments(context.Background(), &IngestRequest{
		Documents: []SearchDocument{{Subject: "test", Content: map[string]interface{}{"test": "test"}}},
	})
	if err == nil {
		t.Error("IngestDocuments() with empty index ID should return error")
	}
	
	// Test empty documents
	_, err = client.IngestDocuments(context.Background(), &IngestRequest{IndexID: "test-index-id"})
	if err == nil {
		t.Error("IngestDocuments() with empty documents should return error")
	}
}

func TestDeleteDocuments(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/delete" {
			t.Errorf("Expected path /delete, got %s", r.URL.Path)
		}
		
		// Decode request body
		var request DeleteDocumentsRequest
		json.NewDecoder(r.Body).Decode(&request)
		
		// Check request body
		if request.IndexID != "test-index-id" {
			t.Errorf("Expected index ID = test-index-id, got %s", request.IndexID)
		}
		if len(request.Subjects) != 2 {
			t.Errorf("Expected 2 subjects, got %d", len(request.Subjects))
		}
		
		// Return mock response
		response := DeleteDocumentsResponse{
			Task: IngestTask{
				TaskID:          "test-task-id",
				ProcessingState: "SUCCESS",
				CreatedAt:       time.Now().Format(time.RFC3339),
				CompletedAt:     time.Now().Format(time.RFC3339),
			},
			Succeeded: 2,
			Failed:    0,
			Total:     2,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test delete documents
	deleteRequest := &DeleteDocumentsRequest{
		IndexID:  "test-index-id",
		Subjects: []string{"subject1", "subject2"},
	}
	
	deleteResponse, err := client.DeleteDocuments(context.Background(), deleteRequest)
	if err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	
	// Check response
	if deleteResponse.Task.TaskID != "test-task-id" {
		t.Errorf("Expected task ID = test-task-id, got %s", deleteResponse.Task.TaskID)
	}
	if deleteResponse.Succeeded != 2 {
		t.Errorf("Expected succeeded = 2, got %d", deleteResponse.Succeeded)
	}
	
	// Test nil request
	_, err = client.DeleteDocuments(context.Background(), nil)
	if err == nil {
		t.Error("DeleteDocuments() with nil request should return error")
	}
	
	// Test empty index ID
	_, err = client.DeleteDocuments(context.Background(), &DeleteDocumentsRequest{
		Subjects: []string{"subject1"},
	})
	if err == nil {
		t.Error("DeleteDocuments() with empty index ID should return error")
	}
	
	// Test empty subjects
	_, err = client.DeleteDocuments(context.Background(), &DeleteDocumentsRequest{
		IndexID: "test-index-id",
	})
	if err == nil {
		t.Error("DeleteDocuments() with empty subjects should return error")
	}
}

func TestGetTaskStatus(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		// Check path
		if r.URL.Path != "/task/test-task-id" {
			t.Errorf("Expected path /task/test-task-id, got %s", r.URL.Path)
		}
		
		// Return mock response
		response := TaskStatusResponse{
			TaskID:           "test-task-id",
			State:            "SUCCESS",
			CreatedAt:        time.Now().Format(time.RFC3339),
			CompletedAt:      time.Now().Format(time.RFC3339),
			DetailLocation:   "detail/location",
			TotalDocuments:   10,
			FailedDocuments:  1,
			SuccessDocuments: 9,
			ErrorCount:       1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
	
	server, client := setupMockServer(handler)
	defer server.Close()
	
	// Test get task status
	taskStatus, err := client.GetTaskStatus(context.Background(), "test-task-id")
	if err != nil {
		t.Fatalf("GetTaskStatus() error = %v", err)
	}
	
	// Check response
	if taskStatus.TaskID != "test-task-id" {
		t.Errorf("Expected task ID = test-task-id, got %s", taskStatus.TaskID)
	}
	if taskStatus.State != "SUCCESS" {
		t.Errorf("Expected state = SUCCESS, got %s", taskStatus.State)
	}
	
	// Test empty task ID
	_, err = client.GetTaskStatus(context.Background(), "")
	if err == nil {
		t.Error("GetTaskStatus() with empty task ID should return error")
	}
}