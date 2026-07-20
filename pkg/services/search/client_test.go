// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupMockServer returns a test server + client pointed at it.
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client, error) {
	server := httptest.NewServer(handler)
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)
	return server, client, err
}

func TestListIndexes(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/index_list" {
			t.Errorf("%s %s, want GET /index_list", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(IndexList{Indexes: []Index{{ID: "idx-1", DisplayName: "One"}}})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	list, err := client.ListIndexes(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListIndexes() error = %v", err)
	}
	if len(list.Indexes) != 1 || list.Indexes[0].ID != "idx-1" {
		t.Errorf("indexes = %+v", list.Indexes)
	}
}

func TestCreateUpdateIndex(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/index":
			_ = json.NewEncoder(w).Encode(Index{ID: "idx-new", DisplayName: "New"})
		case r.Method == http.MethodPatch && r.URL.Path == "/index/idx-new":
			_ = json.NewEncoder(w).Encode(Index{ID: "idx-new", DisplayName: "Renamed"})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	idx, err := client.CreateIndex(context.Background(), &IndexCreateRequest{DisplayName: "New"})
	if err != nil {
		t.Fatalf("CreateIndex() error = %v", err)
	}
	if _, err := client.UpdateIndex(context.Background(), idx.ID, &IndexUpdateRequest{DisplayName: "Renamed"}); err != nil {
		t.Fatalf("UpdateIndex() error = %v", err)
	}
}

func TestSearchHitsIndexScopedPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/index/idx-1/search" {
			t.Errorf("%s %s, want POST /index/idx-1/search", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(SearchResponse{
			Count: 2, Total: 2, HasNextPage: false,
			GMeta: []SearchResult{{Subject: "s1"}, {Subject: "s2"}},
		})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	resp, err := client.Search(context.Background(), &SearchRequest{IndexID: "idx-1", Query: "hello"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Results) != 2 {
		t.Errorf("got %d results, want 2", len(resp.Results))
	}
}

func TestGetSearch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/index/idx-1/search" {
			t.Errorf("%s %s, want GET /index/idx-1/search", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("q") != "hello" {
			t.Errorf("q = %q", r.URL.Query().Get("q"))
		}
		_ = json.NewEncoder(w).Encode(SearchResponse{Total: 1, GMeta: []SearchResult{{Subject: "s1"}}})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	resp, err := client.GetSearch(context.Background(), "idx-1", "hello", 0, 0, false)
	if err != nil {
		t.Fatalf("GetSearch() error = %v", err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("got %d results", len(resp.Results))
	}
}

func TestIngestHitsIndexScopedPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/index/idx-1/ingest" {
			t.Errorf("%s %s, want POST /index/idx-1/ingest", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(IngestResponse{Task: IngestTask{TaskID: "task-1"}})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	_, err = client.IngestDocuments(context.Background(), &IngestRequest{
		IndexID:   "idx-1",
		Documents: []SearchDocument{{Subject: "s1", Content: map[string]interface{}{"a": 1}}},
	})
	if err != nil {
		t.Fatalf("IngestDocuments() error = %v", err)
	}
}

func TestDeleteDocumentsHitsBatchDeletePath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/index/idx-1/batch_delete_by_subject" {
			t.Errorf("%s %s, want POST /index/idx-1/batch_delete_by_subject", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(DeleteDocumentsResponse{TaskID: "task-1"})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	resp, err := client.DeleteDocuments(context.Background(), &DeleteDocumentsRequest{
		IndexID:  "idx-1",
		Subjects: []string{"s1", "s2"},
	})
	if err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if resp.TaskID != "task-1" {
		t.Errorf("TaskID = %s", resp.TaskID)
	}
}

func TestSubjectAndEntryUseQueryParams(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index/idx-1/subject":
			if r.URL.Query().Get("subject") != "urn:s" {
				t.Errorf("subject param = %q", r.URL.Query().Get("subject"))
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"subject": "urn:s"})
		case "/index/idx-1/entry":
			if r.URL.Query().Get("subject") != "urn:s" {
				t.Errorf("entry subject param = %q", r.URL.Query().Get("subject"))
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"subject": "urn:s"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	if _, err := client.GetSubject(context.Background(), "idx-1", "urn:s"); err != nil {
		t.Fatalf("GetSubject() error = %v", err)
	}
	if _, err := client.GetEntry(context.Background(), "idx-1", "urn:s", ""); err != nil {
		t.Fatalf("GetEntry() error = %v", err)
	}
}

func TestGetTaskStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task/task-1" {
			t.Errorf("path = %s, want /task/task-1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(TaskStatusResponse{TaskID: "task-1", IndexID: "idx-1", State: "SUCCESS"})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	task, err := client.GetTaskStatus(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("GetTaskStatus() error = %v", err)
	}
	if task.State != "SUCCESS" || task.IndexID != "idx-1" {
		t.Errorf("task = %+v", task)
	}
}

func TestGetTaskList(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task_list/idx-1" {
			t.Errorf("path = %s, want /task_list/idx-1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"tasks": []map[string]interface{}{{"task_id": "t1"}, {"task_id": "t2"}},
		})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	tasks, err := client.GetTaskList(context.Background(), "idx-1")
	if err != nil {
		t.Fatalf("GetTaskList() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("got %d tasks, want 2", len(tasks))
	}
}

func TestCreateRoleBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/index/idx-1/role" {
			t.Errorf("%s %s, want POST /index/idx-1/role", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["role_name"] != "writer" {
			t.Errorf("role_name = %v", body["role_name"])
		}
		_ = json.NewEncoder(w).Encode(SearchRole{ID: "r1", RoleName: "writer"})
	}
	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	role, err := client.CreateRole(context.Background(), "idx-1", "writer", "urn:globus:auth:identity:x")
	if err != nil {
		t.Fatalf("CreateRole() error = %v", err)
	}
	if role.RoleName != "writer" {
		t.Errorf("RoleName = %s", role.RoleName)
	}
}
