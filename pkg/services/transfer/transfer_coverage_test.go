// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

// ──────────────────────────────────────────────────────────────────────────────
// Test infrastructure helpers
// ──────────────────────────────────────────────────────────────────────────────

// fakeAuthorizer satisfies auth.Authorizer for unit tests.
type fakeAuthorizer struct{ token string }

func (a *fakeAuthorizer) GetAuthorizationHeader(...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}
func (a *fakeAuthorizer) IsValid() bool    { return a.token != "" }
func (a *fakeAuthorizer) GetToken() string { return a.token }

// newTransferClient creates a transfer Client pointed at the given test server.
// It additionally wires a submission_id handler so that CreateTransferTask /
// CreateDeleteTask (which call GetSubmissionID internally) work without a real
// Globus endpoint.
func newTransferClient(t *testing.T, userHandler http.HandlerFunc) (*httptest.Server, *transfer.Client) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/submission_id", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"value": "sub-test-001"})
	})
	// Everything else goes to the user-supplied handler.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r)
	})

	server := httptest.NewServer(mux)

	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "test-token"}),
		transfer.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return server, client
}

// ──────────────────────────────────────────────────────────────────────────────
// Options
// ──────────────────────────────────────────────────────────────────────────────

func TestWithHTTPDebugging(t *testing.T) {
	_, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithHTTPDebugging(true),
	)
	if err != nil {
		t.Fatalf("NewClient() with WithHTTPDebugging error = %v", err)
	}
}

func TestWithHTTPTracing(t *testing.T) {
	_, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("NewClient() with WithHTTPTracing error = %v", err)
	}
}

func TestWithLogger(t *testing.T) {
	_, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithLogger(nil), // nil logger is acceptable
	)
	if err != nil {
		t.Fatalf("NewClient() with WithLogger(nil) error = %v", err)
	}
}

func TestNewClient_MissingAuthorizer(t *testing.T) {
	_, err := transfer.NewClient() // no WithAuthorizer
	if err == nil {
		t.Error("expected error when no authorizer is provided")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListDirectory / ListDirectoryV2
// ──────────────────────────────────────────────────────────────────────────────

func TestListDirectory(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operation/endpoint/ep-001/ls" {
			t.Errorf("expected ls path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("path") != "/data" {
			t.Errorf("expected path=/data, got %s", r.URL.Query().Get("path"))
		}
		resp := transfer.FileList{
			Data: []transfer.FileListItem{
				{Name: "file.txt", Type: "file", Size: 512},
				{Name: "subdir", Type: "dir"},
			},
			Path: "/data",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	opts := &transfer.ListDirectoryOptions{
		EndpointID: "ep-001",
		Path:       "/data",
		OrderBy:    "name",
		ShowHidden: true,
		Limit:      50,
	}
	list, err := client.ListDirectory(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
	if len(list.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(list.Data))
	}
	if list.Data[0].Name != "file.txt" {
		t.Errorf("expected file.txt, got %s", list.Data[0].Name)
	}
}

func TestListDirectory_NilOptions(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.ListDirectory(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil options")
	}
}

func TestListDirectory_EmptyEndpointID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.ListDirectory(context.Background(), &transfer.ListDirectoryOptions{Path: "/data"})
	if err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestListDirectory_WireParams(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// operation_ls sends orderby/filter/limit/offset; the divergent aliases
		// marker/continue_from/excluded_types are NOT sent at 3.65.0.
		if q.Get("filter") != "*.txt" {
			t.Errorf("expected filter=*.txt, got %s", q.Get("filter"))
		}
		if q.Get("marker") != "" {
			t.Errorf("marker must not be sent; got %s", q.Get("marker"))
		}
		if q.Get("continue_from") != "" || q.Get("excluded_types") != "" {
			t.Error("continue_from/excluded_types are not operation_ls wire params")
		}
		resp := transfer.FileList{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	opts := &transfer.ListDirectoryOptions{
		EndpointID:    "ep-001",
		Path:          "/",
		Marker:        "token-abc",
		ContinueFrom:  "file.txt",
		ExcludedTypes: "dir",
		Filter:        "*.txt",
	}
	_, err := client.ListDirectory(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetSubmissionID / GetSubmissionIDV2
// ──────────────────────────────────────────────────────────────────────────────

func TestGetSubmissionID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		// submission_id is handled by the mux in newTransferClient
	})
	defer server.Close()

	id, err := client.GetSubmissionID(context.Background())
	if err != nil {
		t.Fatalf("GetSubmissionID() error = %v", err)
	}
	if id == "" {
		t.Error("expected non-empty submission ID")
	}
}

func TestGetSubmissionID_BySubmissionIDField(t *testing.T) {
	// Server returns "submission_id" instead of "value"
	mux := http.NewServeMux()
	mux.HandleFunc("/submission_id", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"submission_id": "sub-field-001"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	id, err := client.GetSubmissionID(context.Background())
	if err != nil {
		t.Fatalf("GetSubmissionID() error = %v", err)
	}
	if id != "sub-field-001" {
		t.Errorf("expected sub-field-001, got %s", id)
	}
}

func TestGetSubmissionIDV2(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	resp, err := client.GetSubmissionIDV2(context.Background())
	if err != nil {
		t.Fatalf("GetSubmissionIDV2() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetSubmissionIDV2_ServerError(t *testing.T) {
	// Build server that always returns 500 for submission_id
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetSubmissionIDV2(context.Background())
	if err == nil {
		t.Error("expected error on server error")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CreateDeleteTask
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateDeleteTask(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/delete" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req transfer.DeleteTaskRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.DataType != "delete" {
			t.Errorf("expected DATA_TYPE=delete, got %s", req.DataType)
		}
		resp := transfer.TaskResponse{TaskID: "del-task-001", Code: "Accepted"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	req := &transfer.DeleteTaskRequest{
		EndpointID: "ep-001",
		Label:      "Cleanup",
		Items: []transfer.DeleteItem{
			{Path: "/data/old_file.txt"},
			{Path: "/data/another.txt"},
		},
	}
	resp, err := client.CreateDeleteTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDeleteTask() error = %v", err)
	}
	if resp.TaskID != "del-task-001" {
		t.Errorf("expected del-task-001, got %s", resp.TaskID)
	}
}

func TestCreateDeleteTask_NilRequest(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.CreateDeleteTask(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestCreateDeleteTask_EmptyEndpoint(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.CreateDeleteTask(context.Background(), &transfer.DeleteTaskRequest{
		Items: []transfer.DeleteItem{{Path: "/data/file.txt"}},
	})
	if err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestCreateDeleteTask_NoItems(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.CreateDeleteTask(context.Background(), &transfer.DeleteTaskRequest{
		EndpointID: "ep-001",
	})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestCreateDeleteTask_PresetSubmissionID(t *testing.T) {
	// Verify that when SubmissionID is already set, GetSubmissionID is NOT called.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/submission_id" {
			t.Error("GetSubmissionID should not be called when submission_id is preset")
			return
		}
		resp := transfer.TaskResponse{TaskID: "del-preset", Code: "Accepted"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&fakeAuthorizer{token: "tok"}),
		transfer.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	req := &transfer.DeleteTaskRequest{
		EndpointID:   "ep-001",
		SubmissionID: "preset-sub-id",
		Items:        []transfer.DeleteItem{{Path: "/tmp/x"}},
	}
	resp, err := client.CreateDeleteTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDeleteTask() error = %v", err)
	}
	if resp.TaskID != "del-preset" {
		t.Errorf("expected del-preset, got %s", resp.TaskID)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListTasks
// ──────────────────────────────────────────────────────────────────────────────

func TestListTasksTransfer(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task_list" {
			t.Errorf("expected /task_list, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("limit") != "25" {
			t.Errorf("expected limit=25, got %s", q.Get("limit"))
		}
		resp := transfer.TaskList{
			Data: []transfer.Task{
				{TaskID: "t-001", Type: "TRANSFER", Status: "ACTIVE"},
				{TaskID: "t-002", Type: "DELETE", Status: "SUCCEEDED"},
			},
			HasNextPage: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	opts := &transfer.ListTasksOptions{
		FilterType:   "TRANSFER",
		FilterStatus: "ACTIVE",
		Limit:        25,
		Offset:       0,
		PageSize:     25,
		PageToken:    "page-tok",
		FilterTaskID: "t-001",
	}
	list, err := client.ListTasks(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(list.Data) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(list.Data))
	}
}

func TestListTasks_NilOptions(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := transfer.TaskList{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	list, err := client.ListTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTasks() with nil options error = %v", err)
	}
	if list == nil {
		t.Error("expected non-nil list")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetTask / CancelTask (transfer)
// ──────────────────────────────────────────────────────────────────────────────

func TestGetTaskTransfer(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task/t-abc" {
			t.Errorf("expected /task/t-abc, got %s", r.URL.Path)
		}
		resp := transfer.Task{
			TaskID: "t-abc",
			Type:   "TRANSFER",
			Status: "SUCCEEDED",
			Label:  "My Transfer",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	task, err := client.GetTask(context.Background(), "t-abc")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.TaskID != "t-abc" {
		t.Errorf("expected t-abc, got %s", task.TaskID)
	}
	if task.Status != "SUCCEEDED" {
		t.Errorf("expected SUCCEEDED, got %s", task.Status)
	}
}

func TestGetTask_EmptyID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.GetTask(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty task ID")
	}
}

func TestCancelTaskTransfer(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/task/t-abc/cancel" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		resp := transfer.OperationResult{Code: "Canceled", Message: "Task has been canceled"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CancelTask(context.Background(), "t-abc")
	if err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}
	if result.Code != "Canceled" {
		t.Errorf("expected Canceled, got %s", result.Code)
	}
	if result.TaskID != "t-abc" {
		t.Errorf("expected TaskID=t-abc, got %s", result.TaskID)
	}
}

func TestCancelTask_EmptyID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.CancelTask(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty task ID")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CreateDirectory / Mkdir
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateDirectory(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/operation/endpoint/ep-001/mkdir" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		resp := transfer.OperationResult{Code: "DirectoryCreated", Message: "The directory was created"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	err := client.CreateDirectory(context.Background(), &transfer.CreateDirectoryOptions{
		EndpointID: "ep-001",
		Path:       "/new/dir",
	})
	if err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
}

func TestCreateDirectory_NilOptions(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	err := client.CreateDirectory(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil options")
	}
}

func TestCreateDirectory_EmptyEndpointID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	err := client.CreateDirectory(context.Background(), &transfer.CreateDirectoryOptions{Path: "/dir"})
	if err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestCreateDirectory_EmptyPath(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	err := client.CreateDirectory(context.Background(), &transfer.CreateDirectoryOptions{EndpointID: "ep-001"})
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestMkdir(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/operation/endpoint/ep-001/mkdir" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		resp := transfer.OperationResult{Code: "DirectoryCreated", Message: "OK"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	if err := client.Mkdir(context.Background(), "ep-001", "/new/dir", nil); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
}

func TestMkdir_EmptyEndpointID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.Mkdir(context.Background(), "", "/dir", nil); err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestMkdir_EmptyPath(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.Mkdir(context.Background(), "ep-001", "", nil); err == nil {
		t.Error("expected error for empty path")
	}
}

func TestMkdir_WrongCode(t *testing.T) {
	// Server returns a non-"DirectoryCreated" code so Mkdir should return an error.
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := transfer.OperationResult{Code: "ExternalError", Message: "Permission denied"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	if err := client.Mkdir(context.Background(), "ep-001", "/denied", nil); err == nil {
		t.Error("expected error when result code is not DirectoryCreated")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Rename
// ──────────────────────────────────────────────────────────────────────────────

func TestRename(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/operation/endpoint/ep-001/rename" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["old_path"] != "/data/a.txt" {
			t.Errorf("expected old_path=/data/a.txt, got %s", body["old_path"])
		}
		if body["new_path"] != "/data/b.txt" {
			t.Errorf("expected new_path=/data/b.txt, got %s", body["new_path"])
		}
		resp := transfer.OperationResult{Code: "FileRenamed", Message: "OK"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	if err := client.Rename(context.Background(), "ep-001", "/data/a.txt", "/data/b.txt", nil); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
}

func TestRename_EmptyEndpointID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.Rename(context.Background(), "", "/old", "/new", nil); err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestRename_EmptyPaths(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.Rename(context.Background(), "ep-001", "", "/new", nil); err == nil {
		t.Error("expected error for empty old path")
	}
	if err := client.Rename(context.Background(), "ep-001", "/old", "", nil); err == nil {
		t.Error("expected error for empty new path")
	}
}

func TestRename_WrongCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := transfer.OperationResult{Code: "ExternalError", Message: "Permission denied"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	if err := client.Rename(context.Background(), "ep-001", "/old", "/new", nil); err == nil {
		t.Error("expected error when rename code is not FileRenamed")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// SetSubscriptionID / SetSubscriptionAdminVerified
// ──────────────────────────────────────────────────────────────────────────────

func TestSetSubscriptionID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/endpoint/ep-001/subscription" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["subscription_id"] != "sub-xyz" {
			t.Errorf("expected sub-xyz, got %v", body["subscription_id"])
		}
		if _, ok := body["DATA_TYPE"]; ok {
			t.Error("subscription body must not carry DATA_TYPE at 3.65.0")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.SetSubscriptionID(context.Background(), "ep-001", "sub-xyz"); err != nil {
		t.Fatalf("SetSubscriptionID() error = %v", err)
	}
}

func TestSetSubscriptionID_EmptyCollectionID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.SetSubscriptionID(context.Background(), "", "sub-xyz"); err == nil {
		t.Error("expected error for empty collection ID")
	}
}

func TestSetSubscriptionAdminVerified(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/endpoint/ep-001/subscription_admin_verified" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["subscription_admin_verified"] != true {
			t.Errorf("expected subscription_admin_verified=true, got %v", body["subscription_admin_verified"])
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.SetSubscriptionAdminVerified(context.Background(), "ep-001", true); err != nil {
		t.Fatalf("SetSubscriptionAdminVerified() error = %v", err)
	}
}

func TestSetSubscriptionAdminVerified_EmptyCollectionID(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	if err := client.SetSubscriptionAdminVerified(context.Background(), "", true); err == nil {
		t.Error("expected error for empty collection ID")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListEndpointsV2
// ──────────────────────────────────────────────────────────────────────────────

func TestListEndpointsV2(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/endpoint_search" {
			t.Errorf("expected /endpoint_search, got %s", r.URL.Path)
		}
		resp := transfer.EndpointList{
			Data: []transfer.Endpoint{
				{ID: "ep-v2", DisplayName: "V2 Endpoint", Activated: true},
			},
			HasNextPage: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.ListEndpointsV2(context.Background(), &transfer.ListEndpointsOptions{
		FilterFullText:     "test",
		FilterOwnerID:      "user-id",
		FilterHostEndpoint: "host-ep",
		FilterScope:        "my-endpoints",
		Limit:              10,
		Offset:             0,
		PageSize:           10,
		PageToken:          "page-tok",
	})
	if err != nil {
		t.Fatalf("ListEndpointsV2() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Data.Data) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(result.Data.Data))
	}
}

func TestListEndpointsV2_Error(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.ListEndpointsV2(context.Background(), nil)
	if err == nil {
		t.Error("expected error on server failure")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// parseTransferError (additional error code coverage)
// ──────────────────────────────────────────────────────────────────────────────

// We exercise parseTransferError indirectly by issuing requests that trigger
// specific HTTP status codes / response bodies.

// TestDoRequest_ErrorStatuses verifies that various HTTP error status codes
// produce non-nil errors. The exact error type depends on the core transport
// layer, so we only assert that an error was returned.
func TestDoRequest_400EmptyBody(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from 400 response")
	}
}

func TestDoRequest_TaskCompletedCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "TaskCompleted",
			"message": "The task is already complete",
		})
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from TaskCompleted response")
	}
}

func TestDoRequest_TaskCanceledCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "TaskCanceled",
			"message": "The task was canceled",
		})
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from TaskCanceled response")
	}
}

func TestDoRequest_TaskExpiredCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "TaskExpired",
			"message": "The task has expired",
		})
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from TaskExpired response")
	}
}

func TestDoRequest_GenericCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"code":       "Conflict",
			"message":    "Resource already exists",
			"resource":   "/data",
			"request_id": "req-001",
		})
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from Conflict response")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestDoRequest_502BadGateway(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from 502 response")
	}
}

func TestDoRequest_503ServiceUnavailable(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from 503 response")
	}
}

func TestDoRequest_UnknownStatusCode(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 - unexpected status
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from unknown status code")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// TransferError.Error()
// ──────────────────────────────────────────────────────────────────────────────

func TestTransferError_Error(t *testing.T) {
	te := &transfer.TransferError{
		Code:    "ResourceNotFound",
		Message: "not found",
	}
	if te.Error() == "" {
		t.Error("expected non-empty error string")
	}

	teWithID := &transfer.TransferError{
		Code:      "ResourceNotFound",
		Message:   "not found",
		RequestID: "req-abc",
	}
	if teWithID.Error() == "" {
		t.Error("expected non-empty error string with request ID")
	}
	// Should contain the request_id
	if teWithID.RequestID != "req-abc" {
		t.Errorf("unexpected RequestID %s", teWithID.RequestID)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// IsAuthenticationRequired / IsEndpointNotActivated / IsTaskCompleted
// ──────────────────────────────────────────────────────────────────────────────

func TestIsAuthenticationRequired(t *testing.T) {
	cases := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{errors.New("generic"), false},
		{transfer.ErrAuthenticationRequired, true},
		{&transfer.TransferError{Code: transfer.ErrCodeAuthenticationRequired, StatusCode: http.StatusUnauthorized}, true},
		{&transfer.TransferError{Code: "other", StatusCode: http.StatusUnauthorized}, true},
	}
	for _, tc := range cases {
		if got := transfer.IsAuthenticationRequired(tc.err); got != tc.expected {
			t.Errorf("IsAuthenticationRequired(%v) = %v, want %v", tc.err, got, tc.expected)
		}
	}
}

func TestIsEndpointNotActivated(t *testing.T) {
	cases := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{errors.New("generic"), false},
		{transfer.ErrEndpointNotActivated, true},
		{&transfer.TransferError{Code: transfer.ErrCodeEndpointNotActivated}, true},
	}
	for _, tc := range cases {
		if got := transfer.IsEndpointNotActivated(tc.err); got != tc.expected {
			t.Errorf("IsEndpointNotActivated(%v) = %v, want %v", tc.err, got, tc.expected)
		}
	}
}

func TestIsTaskCompleted(t *testing.T) {
	cases := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{errors.New("generic"), false},
		{transfer.ErrTaskCompleted, true},
		{&transfer.TransferError{Code: transfer.ErrCodeTaskCompleted}, true},
	}
	for _, tc := range cases {
		if got := transfer.IsTaskCompleted(tc.err); got != tc.expected {
			t.Errorf("IsTaskCompleted(%v) = %v, want %v", tc.err, got, tc.expected)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// SyncLevel constants interface (constants_interface.go)
// ──────────────────────────────────────────────────────────────────────────────

func TestSyncLevelConstants(t *testing.T) {
	// Exercise the package-level constants.
	if transfer.SyncLevelExists != 0 {
		t.Errorf("SyncLevelExists should be 0, got %d", transfer.SyncLevelExists)
	}
	if transfer.SyncLevelSize != 1 {
		t.Errorf("SyncLevelSize should be 1, got %d", transfer.SyncLevelSize)
	}
	if transfer.SyncLevelModified != 2 {
		t.Errorf("SyncLevelModified should be 2, got %d", transfer.SyncLevelModified)
	}
	if transfer.SyncLevelChecksum != 3 {
		t.Errorf("SyncLevelChecksum should be 3, got %d", transfer.SyncLevelChecksum)
	}
	if transfer.SyncChecksum != transfer.SyncLevelChecksum {
		t.Errorf("SyncChecksum should equal SyncLevelChecksum")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// IsRateLimitExceeded with 429 StatusCode via core.Error
// ──────────────────────────────────────────────────────────────────────────────

func TestIsRateLimitExceeded_ViaHTTP(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer server.Close()

	_, err := client.GetEndpoint(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from 429")
	}
	// The error may be a core.Error (429 from core transport) or a transfer
	// sentinel error depending on what layer intercepts it first.
	// Just verify that an error was returned.
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// 204 No Content path
// ──────────────────────────────────────────────────────────────────────────────

func TestDoRequest_204NoContent(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	// SetSubscriptionAdminVerified should succeed on 204
	err := client.SetSubscriptionAdminVerified(context.Background(), "ep-001", true)
	if err != nil {
		t.Fatalf("expected success on 204, got error = %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Rate-limit headers parsing (parseIntHeader indirectly)
// ──────────────────────────────────────────────────────────────────────────────

func TestRateLimitHeaders(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "50")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transfer.EndpointList{})
	})
	defer server.Close()

	_, err := client.ListEndpoints(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}
	// No assertions needed - just verifying no panic from header parsing.
}

// ──────────────────────────────────────────────────────────────────────────────
// Resumable transfer wrapper methods
// ──────────────────────────────────────────────────────────────────────────────

// ──────────────────────────────────────────────────────────────────────────────
// Resumable transfer wrapper methods (error paths via invalid checkpoint ID)
// ──────────────────────────────────────────────────────────────────────────────

func TestGetResumableTransferStatus_NotFound(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	// A non-existent checkpoint ID should cause a "not found" error from the
	// file-system checkpoint storage.
	_, err := client.GetResumableTransferStatus(context.Background(), "nonexistent-checkpoint-id-xyz")
	if err == nil {
		t.Error("expected error for nonexistent checkpoint ID")
	}
}

func TestCancelResumableTransfer_NotFound(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	err := client.CancelResumableTransfer(context.Background(), "nonexistent-checkpoint-id-xyz")
	if err == nil {
		t.Error("expected error for nonexistent checkpoint ID")
	}
}

func TestResumeResumableTransfer_NotFound(t *testing.T) {
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	_, err := client.ResumeResumableTransfer(context.Background(), "nonexistent-checkpoint-id-xyz", nil)
	if err == nil {
		t.Error("expected error for nonexistent checkpoint ID")
	}
}

func TestSubmitResumableTransfer(t *testing.T) {
	// SubmitResumableTransfer delegates to CreateResumableTransfer which uses
	// ListFiles and CreateTransferTask internally. We provide a minimal mock.
	server, client := newTransferClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/operation/endpoint/src-ep/ls":
			items := []transfer.FileListItem{{Name: "file.dat", Type: "file", Size: 100}}
			resp := transfer.FileList{Data: items, Path: "/src"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case "/transfer":
			resp := transfer.TaskResponse{TaskID: "resumable-task-001", Code: "Accepted"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	checkpointID, err := client.SubmitResumableTransfer(
		context.Background(),
		"src-ep", "/src",
		"dst-ep", "/dst",
		nil,
	)
	if err != nil {
		t.Fatalf("SubmitResumableTransfer() error = %v", err)
	}
	if checkpointID == "" {
		t.Error("expected non-empty checkpoint ID")
	}
}
