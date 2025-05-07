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

func TestCreateEnvironment(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req EnvironmentCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-environment", req.Name)
		assert.Equal(t, "API_KEY", req.Variables["AUTH_TOKEN_NAME"])

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := EnvironmentResponse{
			ID:   "env123",
			Name: req.Name,
			Variables: map[string]string{
				"AUTH_TOKEN_NAME": "API_KEY",
				"DEBUG":           "true",
			},
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
	req := &EnvironmentCreateRequest{
		Name:        "test-environment",
		Description: "Test environment for unit tests",
		Variables: map[string]string{
			"AUTH_TOKEN_NAME": "API_KEY",
			"DEBUG":           "true",
		},
		Resources: map[string]interface{}{
			"cpus":   2,
			"memory": "4GB",
		},
	}

	resp, err := client.CreateEnvironment(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "env123", resp.ID)
	assert.Equal(t, "test-environment", resp.Name)
	assert.Equal(t, "API_KEY", resp.Variables["AUTH_TOKEN_NAME"])
}

func TestListEnvironments(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := EnvironmentList{
			Environments: []EnvironmentResponse{
				{
					ID:   "env123",
					Name: "test-environment-1",
					Variables: map[string]string{
						"DEBUG": "true",
					},
				},
				{
					ID:   "env456",
					Name: "test-environment-2",
					Variables: map[string]string{
						"LOG_LEVEL": "DEBUG",
					},
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
	resp, err := client.ListEnvironments(ctx, &ListEnvironmentsOptions{
		PerPage: 10,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Environments, 2)
	assert.Equal(t, "env123", resp.Environments[0].ID)
	assert.Equal(t, "env456", resp.Environments[1].ID)
}

func TestGetEnvironment(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments/env123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := EnvironmentResponse{
			ID:   "env123",
			Name: "test-environment",
			Variables: map[string]string{
				"DEBUG":     "true",
				"LOG_LEVEL": "INFO",
			},
			Resources: map[string]interface{}{
				"cpus":   2,
				"memory": "4GB",
			},
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
	resp, err := client.GetEnvironment(ctx, "env123")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "env123", resp.ID)
	assert.Equal(t, "test-environment", resp.Name)
	assert.Equal(t, "true", resp.Variables["DEBUG"])
	assert.Equal(t, float64(2), resp.Resources["cpus"])
}

func TestUpdateEnvironment(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments/env123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		// Verify request body
		var req EnvironmentUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "updated-environment", req.Name)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := EnvironmentResponse{
			ID:   "env123",
			Name: req.Name,
			Variables: map[string]string{
				"DEBUG":     "true",
				"LOG_LEVEL": "DEBUG",
			},
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
	req := &EnvironmentUpdateRequest{
		Name: "updated-environment",
		Variables: map[string]string{
			"DEBUG":     "true",
			"LOG_LEVEL": "DEBUG",
		},
	}

	resp, err := client.UpdateEnvironment(ctx, "env123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "env123", resp.ID)
	assert.Equal(t, "updated-environment", resp.Name)
	assert.Equal(t, "DEBUG", resp.Variables["LOG_LEVEL"])
}

func TestDeleteEnvironment(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments/env123", r.URL.Path)
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
	err = client.DeleteEnvironment(ctx, "env123")
	assert.NoError(t, err)
}

func TestApplyEnvironmentToTask(t *testing.T) {
	// Setup test server for GetEnvironment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/environments/env123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := EnvironmentResponse{
			ID:   "env123",
			Name: "test-environment",
			Variables: map[string]string{
				"API_KEY":   "secret-key",
				"LOG_LEVEL": "DEBUG",
			},
			Resources: map[string]interface{}{
				"cpus":   4,
				"memory": "8GB",
				"gpu":    true,
			},
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

	// Create task request
	taskRequest := &TaskRequest{
		FunctionID: "func123",
		EndpointID: "endpoint123",
		Args:       []interface{}{"arg1", "arg2"},
		Kwargs: map[string]interface{}{
			"param1": "value1",
		},
	}

	// Test the method
	ctx := context.Background()
	enrichedRequest, err := client.ApplyEnvironmentToTask(ctx, taskRequest, "env123")

	assert.NoError(t, err)
	assert.NotNil(t, enrichedRequest)
	assert.Equal(t, "func123", enrichedRequest.FunctionID)
	assert.Equal(t, "endpoint123", enrichedRequest.EndpointID)
	assert.Equal(t, []interface{}{"arg1", "arg2"}, enrichedRequest.Args)

	// Check that original kwargs are preserved
	assert.Equal(t, "value1", enrichedRequest.Kwargs["param1"])

	// Check that environment variables were added
	variables, ok := enrichedRequest.Kwargs["environment"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "secret-key", variables["API_KEY"])
	assert.Equal(t, "DEBUG", variables["LOG_LEVEL"])

	// Check that resources were added
	resources, ok := enrichedRequest.Kwargs["resources"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(4), resources["cpus"])
	assert.Equal(t, "8GB", resources["memory"])
	assert.Equal(t, true, resources["gpu"])
}

func TestCreateSecret(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req SecretCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "API_KEY", req.Name)
		assert.Equal(t, "secret-value", req.Value)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := SecretResponse{
			ID:          "secret123",
			Name:        req.Name,
			Description: req.Description,
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
	req := &SecretCreateRequest{
		Name:        "API_KEY",
		Description: "API key for external service",
		Value:       "secret-value",
	}

	resp, err := client.CreateSecret(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "secret123", resp.ID)
	assert.Equal(t, "API_KEY", resp.Name)
	assert.Equal(t, "API key for external service", resp.Description)
}

func TestListSecrets(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := []SecretResponse{
			{
				ID:          "secret123",
				Name:        "API_KEY",
				Description: "API key for external service",
			},
			{
				ID:          "secret456",
				Name:        "DATABASE_PASSWORD",
				Description: "Database password",
			},
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
	secrets, err := client.ListSecrets(ctx)

	assert.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, "secret123", secrets[0].ID)
	assert.Equal(t, "secret456", secrets[1].ID)
}

func TestDeleteSecret(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets/secret123", r.URL.Path)
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
	err = client.DeleteSecret(ctx, "secret123")
	assert.NoError(t, err)
}
