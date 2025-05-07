// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test mock server
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client, error) {
	server := httptest.NewServer(handler)

	// Create a client that uses the test server
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)

	return server, client, err
}

func TestListEndpoints(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoints" {
			t.Errorf("Expected path /endpoints, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if perPage := queryParams.Get("per_page"); perPage != "10" {
			t.Errorf("Expected per_page=10, got %s", perPage)
		}

		// Return mock response
		endpointTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		// Create a ComputeEndpointList object
		response := ComputeEndpointList{
			Endpoints: []ComputeEndpoint{
				{
					ID:           "test-endpoint-id",
					UUID:         "test-uuid",
					Status:       "online",
					Name:         "Test Endpoint",
					Description:  "A test endpoint",
					Owner:        "test-user",
					CreatedAt:    endpointTime,
					LastModified: endpointTime,
					Connected:    true,
					Type:         "container",
					Public:       false,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client, err := setupMockServer(handler)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer server.Close()

	// Test list endpoints
	options := &ListEndpointsOptions{
		PerPage: 10,
	}

	endpointList, err := client.ListEndpoints(context.Background(), options)
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}

	// Check response
	if len(endpointList.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpointList.Endpoints))
	}

	endpoint := endpointList.Endpoints[0]
	if endpoint.ID != "test-endpoint-id" {
		t.Errorf("Expected endpoint ID = test-endpoint-id, got %s", endpoint.ID)
	}
	if endpoint.Name != "Test Endpoint" {
		t.Errorf("Expected endpoint name = Test Endpoint, got %s", endpoint.Name)
	}
}

// Rest of the file remains unchanged...
