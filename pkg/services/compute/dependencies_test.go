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

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

func TestRegisterDependency(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependencies", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req DependencyRegistrationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "data-science-deps", req.Name)
		assert.Equal(t, "numpy==1.22.0\npandas>=1.3.0\nscipy", req.PythonRequirements)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyResponse{
			ID:                 "dep123",
			Name:               req.Name,
			PythonRequirements: req.PythonRequirements,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	req := &DependencyRegistrationRequest{
		Name:               "data-science-deps",
		Description:        "Common data science libraries",
		PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy",
	}

	resp, err := client.RegisterDependency(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "dep123", resp.ID)
	assert.Equal(t, "data-science-deps", resp.Name)
	assert.Equal(t, "numpy==1.22.0\npandas>=1.3.0\nscipy", resp.PythonRequirements)
}

func TestListDependencies(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependencies", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyList{
			Dependencies: []DependencyResponse{
				{
					ID:                 "dep123",
					Name:               "data-science-deps",
					PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy",
				},
				{
					ID:   "dep456",
					Name: "web-deps",
					PythonPackages: []PythonPackage{
						{Name: "flask", Version: "2.0.0"},
						{Name: "requests", Version: "2.27.1"},
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
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.ListDependencies(ctx, &ListDependenciesOptions{
		PerPage: 10,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Dependencies, 2)
	assert.Equal(t, "dep123", resp.Dependencies[0].ID)
	assert.Equal(t, "dep456", resp.Dependencies[1].ID)
}

func TestGetDependency(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependencies/dep123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyResponse{
			ID:                 "dep123",
			Name:               "data-science-deps",
			Description:        "Common data science libraries",
			PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.GetDependency(ctx, "dep123")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "dep123", resp.ID)
	assert.Equal(t, "data-science-deps", resp.Name)
	assert.Equal(t, "numpy==1.22.0\npandas>=1.3.0\nscipy", resp.PythonRequirements)
}

func TestUpdateDependency(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependencies/dep123", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		// Verify request body
		var req DependencyUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "updated-deps", req.Name)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyResponse{
			ID:                 "dep123",
			Name:               req.Name,
			PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy==1.8.0",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	req := &DependencyUpdateRequest{
		Name:               "updated-deps",
		PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy==1.8.0",
	}

	resp, err := client.UpdateDependency(ctx, "dep123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "dep123", resp.ID)
	assert.Equal(t, "updated-deps", resp.Name)
}

func TestDeleteDependency(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependencies/dep123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		// Return mock response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	err := client.DeleteDependency(ctx, "dep123")
	assert.NoError(t, err)
}

func TestAttachDependencyToFunction(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/functions/func123/dependencies", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "dep123", req["dependency_id"])

		// Return mock response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	err := client.AttachDependencyToFunction(ctx, "func123", "dep123")
	assert.NoError(t, err)
}

func TestDetachDependencyFromFunction(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/functions/func123/dependencies/dep123", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		// Return mock response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	err := client.DetachDependencyFromFunction(ctx, "func123", "dep123")
	assert.NoError(t, err)
}

func TestListFunctionDependencies(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/functions/func123/dependencies", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := []DependencyResponse{
			{
				ID:                 "dep123",
				Name:               "data-science-deps",
				PythonRequirements: "numpy==1.22.0\npandas>=1.3.0\nscipy",
			},
			{
				ID:   "dep456",
				Name: "web-deps",
				PythonPackages: []PythonPackage{
					{Name: "flask", Version: "2.0.0"},
					{Name: "requests", Version: "2.27.1"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		Client: core.NewClient(
			core.WithBaseURL(server.URL+"/"),
		),
	}

	// Test the method
	ctx := context.Background()
	deps, err := client.ListFunctionDependencies(ctx, "func123")

	assert.NoError(t, err)
	assert.Len(t, deps, 2)
	assert.Equal(t, "dep123", deps[0].ID)
	assert.Equal(t, "dep456", deps[1].ID)
}