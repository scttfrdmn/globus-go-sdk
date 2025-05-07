// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterContainer(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req ContainerRegistrationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-container", req.Name)
		assert.Equal(t, "python:3.9", req.Image)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := ContainerResponse{
			ID:    "container123",
			Name:  req.Name,
			Image: req.Image,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &ContainerRegistrationRequest{
		Name:  "test-container",
		Image: "python:3.9",
	}

	resp, err := client.RegisterContainer(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "container123", resp.ID)
	assert.Equal(t, "test-container", resp.Name)
	assert.Equal(t, "python:3.9", resp.Image)
}

func TestListContainers(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := ContainerList{
			Containers: []ContainerResponse{
				{
					ID:    "container123",
					Name:  "test-container-1",
					Image: "python:3.9",
				},
				{
					ID:    "container456",
					Name:  "test-container-2",
					Image: "ubuntu:20.04",
				},
			},
			Total:       2,
			HasNextPage: false,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.ListContainers(ctx, &ListContainersOptions{
		PerPage: 10,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Containers, 2)
	assert.Equal(t, "container123", resp.Containers[0].ID)
	assert.Equal(t, "container456", resp.Containers[1].ID)
}

func TestGetContainer(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers/container123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := ContainerResponse{
			ID:          "container123",
			Name:        "test-container",
			Image:       "python:3.9",
			Description: "Test container for unit tests",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.GetContainer(ctx, "container123")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "container123", resp.ID)
	assert.Equal(t, "test-container", resp.Name)
	assert.Equal(t, "python:3.9", resp.Image)
}

func TestUpdateContainer(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers/container123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		// Verify request body
		var req ContainerUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "updated-container", req.Name)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := ContainerResponse{
			ID:    "container123",
			Name:  req.Name,
			Image: "python:3.9",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &ContainerUpdateRequest{
		Name: "updated-container",
	}

	resp, err := client.UpdateContainer(ctx, "container123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "container123", resp.ID)
	assert.Equal(t, "updated-container", resp.Name)
}

func TestDeleteContainer(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/containers/container123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		// Return mock response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	err = client.DeleteContainer(ctx, "container123")
	assert.NoError(t, err)
}

func TestRunContainerFunction(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/run_container", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req ContainerTaskRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "endpoint123", req.EndpointID)
		assert.Equal(t, "container123", req.ContainerID)
		assert.Equal(t, "function123", req.FunctionID)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := TaskResponse{
			TaskID: "task123",
			Status: "ACTIVE",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &ContainerTaskRequest{
		EndpointID:  "endpoint123",
		ContainerID: "container123",
		FunctionID:  "function123",
		Args:        []interface{}{"test"},
		Environment: map[string]string{"ENV_VAR": "value"},
	}

	resp, err := client.RunContainerFunction(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "task123", resp.TaskID)
	assert.Equal(t, "ACTIVE", resp.Status)
}
