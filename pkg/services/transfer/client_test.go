// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Test helper to set up a mock server and client
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)

	// Create a client that uses the test server
	client := NewClient("test-access-token",
		core.WithBaseURL(server.URL+"/"),
	)

	return server, client
}

func TestBuildURL(t *testing.T) {
	client := NewClient("test-access-token",
		core.WithBaseURL("https://example.com"),
	)

	// Test with no query parameters
	url := client.buildURL("test/path", nil)
	if url != "https://example.com/test/path" {
		t.Errorf("buildURL() = %v, want %v", url, "https://example.com/test/path")
	}

	// Test with query parameters
	query := map[string][]string{
		"param1": {"value1"},
		"param2": {"value2"},
	}
	url = client.buildURL("test/path", query)
	if url != "https://example.com/test/path?param1=value1&param2=value2" {
		t.Errorf("buildURL() with query = %v, want %v", url, "https://example.com/test/path?param1=value1&param2=value2")
	}

	// Test with trailing slash in base URL
	client = NewClient("test-access-token",
		core.WithBaseURL("https://example.com/"),
	)
	url = client.buildURL("test/path", nil)
	if url != "https://example.com/test/path" {
		t.Errorf("buildURL() with trailing slash = %v, want %v", url, "https://example.com/test/path")
	}
}

func TestListEndpoints(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoint_search" {
			t.Errorf("Expected path /endpoint_search, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("filter_fulltext") != "test" {
			t.Errorf("Expected filter_fulltext=test, got %s", query.Get("filter_fulltext"))
		}
		if query.Get("filter_scope") != "my-endpoints" {
			t.Errorf("Expected filter_scope=my-endpoints, got %s", query.Get("filter_scope"))
		}
		if query.Get("limit") != "100" {
			t.Errorf("Expected limit=100, got %s", query.Get("limit"))
		}

		// Return mock response
		endpoints := []Endpoint{
			{
				ID:          "endpoint1",
				DisplayName: "Endpoint 1",
				OwnerString: "user1",
				OwnerID:     "user-id-1",
				Activated:   true,
				Public:      true,
			},
			{
				ID:          "endpoint2",
				DisplayName: "Endpoint 2",
				OwnerString: "user2",
				OwnerID:     "user-id-2",
				Activated:   false,
				Public:      false,
			},
		}

		response := EndpointList{
			Data:          endpoints,
			HasNextPage:   true,
			NextPageToken: "next-page-token",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list endpoints
	options := &ListEndpointsOptions{
		FilterFullText: "test",
		FilterScope:    "my-endpoints",
		Limit:          100,
	}

	endpointList, err := client.ListEndpoints(context.Background(), options)
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}

	// Check response
	if len(endpointList.Data) != 2 {
		t.Fatalf("ListEndpoints() returned %d endpoints, want 2", len(endpointList.Data))
	}
	if endpointList.Data[0].ID != "endpoint1" {
		t.Errorf("ListEndpoints() endpoint[0].ID = %v, want %v", endpointList.Data[0].ID, "endpoint1")
	}
	if endpointList.Data[1].ID != "endpoint2" {
		t.Errorf("ListEndpoints() endpoint[1].ID = %v, want %v", endpointList.Data[1].ID, "endpoint2")
	}
	if !endpointList.HasNextPage {
		t.Errorf("ListEndpoints() HasNextPage = %v, want %v", endpointList.HasNextPage, true)
	}
	if endpointList.NextPageToken != "next-page-token" {
		t.Errorf("ListEndpoints() NextPageToken = %v, want %v", endpointList.NextPageToken, "next-page-token")
	}

	// Test with nil options
	handler = func(w http.ResponseWriter, r *http.Request) {
		response := EndpointList{
			Data:        []Endpoint{},
			HasNextPage: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	endpointList, err = client.ListEndpoints(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListEndpoints() with nil options error = %v", err)
	}

	if len(endpointList.Data) != 0 {
		t.Fatalf("ListEndpoints() with nil options returned %d endpoints, want 0", len(endpointList.Data))
	}
}

func TestGetEndpoint(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoint/endpoint1" {
			t.Errorf("Expected path /endpoint/endpoint1, got %s", r.URL.Path)
		}

		// Return mock response
		endpoint := Endpoint{
			ID:          "endpoint1",
			DisplayName: "Test Endpoint",
			OwnerString: "testuser",
			OwnerID:     "user-id",
			Activated:   true,
			Public:      true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(endpoint)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get endpoint
	endpoint, err := client.GetEndpoint(context.Background(), "endpoint1")
	if err != nil {
		t.Fatalf("GetEndpoint() error = %v", err)
	}

	// Check response
	if endpoint.ID != "endpoint1" {
		t.Errorf("GetEndpoint() ID = %v, want %v", endpoint.ID, "endpoint1")
	}
	if endpoint.DisplayName != "Test Endpoint" {
		t.Errorf("GetEndpoint() DisplayName = %v, want %v", endpoint.DisplayName, "Test Endpoint")
	}
	if endpoint.OwnerString != "testuser" {
		t.Errorf("GetEndpoint() OwnerString = %v, want %v", endpoint.OwnerString, "testuser")
	}
	if !endpoint.Activated {
		t.Errorf("GetEndpoint() Activated = %v, want %v", endpoint.Activated, true)
	}

	// Test with empty ID
	_, err = client.GetEndpoint(context.Background(), "")
	if err == nil {
		t.Error("GetEndpoint() with empty ID should return error")
	}
}

func TestActivateEndpoint(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoint/endpoint1/autoactivate" {
			t.Errorf("Expected path /endpoint/endpoint1/autoactivate, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody map[string]bool
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if autoActivate, ok := requestBody["auto_activate"]; !ok || !autoActivate {
			t.Errorf("Expected auto_activate=true, got %v", autoActivate)
		}

		// Return mock response
		result := OperationResult{
			Code:    "AutoActivated",
			Message: "Endpoint activated successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test activate endpoint
	err := client.ActivateEndpoint(context.Background(), "endpoint1")
	if err != nil {
		t.Fatalf("ActivateEndpoint() error = %v", err)
	}

	// Test with failure response
	handler = func(w http.ResponseWriter, r *http.Request) {
		result := OperationResult{
			Code:    "ActivationFailed",
			Message: "Failed to activate endpoint",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	err = client.ActivateEndpoint(context.Background(), "endpoint1")
	if err == nil {
		t.Error("ActivateEndpoint() with failure response should return error")
	}

	// Test with empty ID
	_, client = setupMockServer(handler)
	err = client.ActivateEndpoint(context.Background(), "")
	if err == nil {
		t.Error("ActivateEndpoint() with empty ID should return error")
	}
}

func TestListFiles(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/operation/endpoint/endpoint1/ls" {
			t.Errorf("Expected path /operation/endpoint/endpoint1/ls, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("path") != "/path/to/dir" {
			t.Errorf("Expected path=/path/to/dir, got %s", query.Get("path"))
		}
		if query.Get("orderby") != "name" {
			t.Errorf("Expected orderby=name, got %s", query.Get("orderby"))
		}
		if query.Get("show_hidden") != "1" {
			t.Errorf("Expected show_hidden=1, got %s", query.Get("show_hidden"))
		}

		// Return mock response
		files := []FileListItem{
			{
				DataType:     "file",
				Name:         "file1.txt",
				Type:         "file",
				Size:         1024,
				LastModified: "2023-01-01 12:00:00",
				Permissions:  "rw-r--r--",
				User:         "user1",
				Group:        "group1",
			},
			{
				DataType:     "dir",
				Name:         "dir1",
				Type:         "dir",
				LastModified: "2023-01-01 12:00:00",
				Permissions:  "rwxr-xr-x",
				User:         "user1",
				Group:        "group1",
			},
		}

		response := FileList{
			Data:         files,
			EndpointID:   "endpoint1",
			Path:         "/path/to/dir",
			HasNextPage:  true,
			Marker:       "marker-value",
			AbsolutePath: "/absolute/path/to/dir",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list files
	options := &ListFileOptions{
		OrderBy:    "name",
		ShowHidden: true,
	}

	fileList, err := client.ListFiles(context.Background(), "endpoint1", "/path/to/dir", options)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	// Check response
	if len(fileList.Data) != 2 {
		t.Fatalf("ListFiles() returned %d files, want 2", len(fileList.Data))
	}
	if fileList.Data[0].Name != "file1.txt" {
		t.Errorf("ListFiles() file[0].Name = %v, want %v", fileList.Data[0].Name, "file1.txt")
	}
	if fileList.Data[0].Type != "file" {
		t.Errorf("ListFiles() file[0].Type = %v, want %v", fileList.Data[0].Type, "file")
	}
	if fileList.Data[1].Name != "dir1" {
		t.Errorf("ListFiles() file[1].Name = %v, want %v", fileList.Data[1].Name, "dir1")
	}
	if fileList.Data[1].Type != "dir" {
		t.Errorf("ListFiles() file[1].Type = %v, want %v", fileList.Data[1].Type, "dir")
	}
	if fileList.EndpointID != "endpoint1" {
		t.Errorf("ListFiles() EndpointID = %v, want %v", fileList.EndpointID, "endpoint1")
	}
	if fileList.Path != "/path/to/dir" {
		t.Errorf("ListFiles() Path = %v, want %v", fileList.Path, "/path/to/dir")
	}
	if !fileList.HasNextPage {
		t.Errorf("ListFiles() HasNextPage = %v, want %v", fileList.HasNextPage, true)
	}

	// Test with empty endpoint ID
	_, err = client.ListFiles(context.Background(), "", "/path/to/dir", options)
	if err == nil {
		t.Error("ListFiles() with empty endpoint ID should return error")
	}
}

func TestCreateTransferTask(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/transfer" {
			t.Errorf("Expected path /transfer, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody TransferTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody.DataType != "transfer" {
			t.Errorf("Expected DataType=transfer, got %s", requestBody.DataType)
		}
		if requestBody.Label != "Test Transfer" {
			t.Errorf("Expected Label=Test Transfer, got %s", requestBody.Label)
		}
		if requestBody.SourceEndpointID != "source-endpoint" {
			t.Errorf("Expected SourceEndpointID=source-endpoint, got %s", requestBody.SourceEndpointID)
		}
		if requestBody.DestinationEndpointID != "destination-endpoint" {
			t.Errorf("Expected DestinationEndpointID=destination-endpoint, got %s", requestBody.DestinationEndpointID)
		}
		if len(requestBody.Items) != 1 {
			t.Fatalf("Expected 1 transfer item, got %d", len(requestBody.Items))
		}
		if requestBody.Items[0].SourcePath != "/source/file.txt" {
			t.Errorf("Expected Items[0].SourcePath=/source/file.txt, got %s", requestBody.Items[0].SourcePath)
		}
		if requestBody.Items[0].DestinationPath != "/destination/file.txt" {
			t.Errorf("Expected Items[0].DestinationPath=/destination/file.txt, got %s", requestBody.Items[0].DestinationPath)
		}

		// Return mock response
		response := TaskResponse{
			TaskID:  "task-id",
			Code:    "Accepted",
			Message: "Task created successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test create transfer task
	request := &TransferTaskRequest{
		DataType:              "transfer",
		Label:                 "Test Transfer",
		SourceEndpointID:      "source-endpoint",
		DestinationEndpointID: "destination-endpoint",
		Items: []TransferItem{
			{
				SourcePath:      "/source/file.txt",
				DestinationPath: "/destination/file.txt",
			},
		},
	}

	response, err := client.CreateTransferTask(context.Background(), request)
	if err != nil {
		t.Fatalf("CreateTransferTask() error = %v", err)
	}

	// Check response
	if response.TaskID != "task-id" {
		t.Errorf("CreateTransferTask() TaskID = %v, want %v", response.TaskID, "task-id")
	}
	if response.Code != "Accepted" {
		t.Errorf("CreateTransferTask() Code = %v, want %v", response.Code, "Accepted")
	}

	// Test with nil request
	_, err = client.CreateTransferTask(context.Background(), nil)
	if err == nil {
		t.Error("CreateTransferTask() with nil request should return error")
	}

	// Test with missing source endpoint
	_, err = client.CreateTransferTask(context.Background(), &TransferTaskRequest{
		DataType:              "transfer",
		DestinationEndpointID: "destination-endpoint",
		Items: []TransferItem{
			{
				SourcePath:      "/source/file.txt",
				DestinationPath: "/destination/file.txt",
			},
		},
	})
	if err == nil {
		t.Error("CreateTransferTask() with missing source endpoint should return error")
	}

	// Test with missing destination endpoint
	_, err = client.CreateTransferTask(context.Background(), &TransferTaskRequest{
		DataType:         "transfer",
		SourceEndpointID: "source-endpoint",
		Items: []TransferItem{
			{
				SourcePath:      "/source/file.txt",
				DestinationPath: "/destination/file.txt",
			},
		},
	})
	if err == nil {
		t.Error("CreateTransferTask() with missing destination endpoint should return error")
	}

	// Test with no items
	_, err = client.CreateTransferTask(context.Background(), &TransferTaskRequest{
		DataType:              "transfer",
		SourceEndpointID:      "source-endpoint",
		DestinationEndpointID: "destination-endpoint",
		Items:                 []TransferItem{},
	})
	if err == nil {
		t.Error("CreateTransferTask() with no items should return error")
	}
}

func TestSubmitTransfer(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/transfer" {
			t.Errorf("Expected path /transfer, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody TransferTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody.DataType != "transfer" {
			t.Errorf("Expected DataType=transfer, got %s", requestBody.DataType)
		}
		if requestBody.Label != "Test Label" {
			t.Errorf("Expected Label=Test Label, got %s", requestBody.Label)
		}
		if requestBody.SourceEndpointID != "source-endpoint" {
			t.Errorf("Expected SourceEndpointID=source-endpoint, got %s", requestBody.SourceEndpointID)
		}
		if requestBody.DestinationEndpointID != "destination-endpoint" {
			t.Errorf("Expected DestinationEndpointID=destination-endpoint, got %s", requestBody.DestinationEndpointID)
		}
		if !requestBody.VerifyChecksum {
			t.Errorf("Expected VerifyChecksum=true, got %v", requestBody.VerifyChecksum)
		}
		if !requestBody.Items[0].Recursive {
			t.Errorf("Expected Items[0].Recursive=true, got %v", requestBody.Items[0].Recursive)
		}

		// Return mock response
		response := TaskResponse{
			TaskID: "task-id",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test submit transfer
	options := map[string]interface{}{
		"recursive":       true,
		"verify_checksum": true,
	}

	response, err := client.SubmitTransfer(
		context.Background(),
		"source-endpoint", "/source/path",
		"destination-endpoint", "/destination/path",
		"Test Label",
		options,
	)
	if err != nil {
		t.Fatalf("SubmitTransfer() error = %v", err)
	}

	// Check response
	if response.TaskID != "task-id" {
		t.Errorf("SubmitTransfer() TaskID = %v, want %v", response.TaskID, "task-id")
	}
}
