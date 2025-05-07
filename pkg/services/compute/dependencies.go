// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RegisterDependency registers a new dependency definition with Globus Compute
func (c *Client) RegisterDependency(ctx context.Context, request *DependencyRegistrationRequest) (*DependencyResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("dependency registration request is required")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("dependency name is required")
	}

	// At least one type of dependency specification is required
	if request.PythonRequirements == "" && len(request.PythonPackages) == 0 &&
		len(request.CustomDependencies) == 0 && request.GitRepo == "" {
		return nil, fmt.Errorf("at least one dependency specification is required")
	}

	var dependency DependencyResponse
	err := c.doRequest(ctx, http.MethodPost, "dependencies", nil, request, &dependency)
	if err != nil {
		return nil, err
	}

	return &dependency, nil
}

// GetDependency retrieves a dependency by ID
func (c *Client) GetDependency(ctx context.Context, dependencyID string) (*DependencyResponse, error) {
	if dependencyID == "" {
		return nil, fmt.Errorf("dependency ID is required")
	}

	var dependency DependencyResponse
	err := c.doRequest(ctx, http.MethodGet, "dependencies/"+dependencyID, nil, nil, &dependency)
	if err != nil {
		return nil, err
	}

	return &dependency, nil
}

// ListDependencies lists all dependencies the user has access to
func (c *Client) ListDependencies(ctx context.Context, options *ListDependenciesOptions) (*DependencyList, error) {
	query := url.Values{}
	if options != nil {
		if options.PerPage > 0 {
			query.Set("per_page", fmt.Sprintf("%d", options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.Search != "" {
			query.Set("search", options.Search)
		}
	}

	var dependencyList DependencyList
	err := c.doRequest(ctx, http.MethodGet, "dependencies", query, nil, &dependencyList)
	if err != nil {
		return nil, err
	}

	return &dependencyList, nil
}

// UpdateDependency updates an existing dependency
func (c *Client) UpdateDependency(ctx context.Context, dependencyID string, request *DependencyUpdateRequest) (*DependencyResponse, error) {
	if dependencyID == "" {
		return nil, fmt.Errorf("dependency ID is required")
	}

	if request == nil {
		return nil, fmt.Errorf("dependency update request is required")
	}

	var dependency DependencyResponse
	err := c.doRequest(ctx, http.MethodPut, "dependencies/"+dependencyID, nil, request, &dependency)
	if err != nil {
		return nil, err
	}

	return &dependency, nil
}

// DeleteDependency deletes a dependency
func (c *Client) DeleteDependency(ctx context.Context, dependencyID string) error {
	if dependencyID == "" {
		return fmt.Errorf("dependency ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "dependencies/"+dependencyID, nil, nil, nil)
}

// AttachDependencyToFunction attaches a dependency to a function
func (c *Client) AttachDependencyToFunction(ctx context.Context, functionID string, dependencyID string) error {
	if functionID == "" {
		return fmt.Errorf("function ID is required")
	}

	if dependencyID == "" {
		return fmt.Errorf("dependency ID is required")
	}

	request := map[string]string{
		"dependency_id": dependencyID,
	}

	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("functions/%s/dependencies", functionID), nil, request, nil)
}

// DetachDependencyFromFunction detaches a dependency from a function
func (c *Client) DetachDependencyFromFunction(ctx context.Context, functionID string, dependencyID string) error {
	if functionID == "" {
		return fmt.Errorf("function ID is required")
	}

	if dependencyID == "" {
		return fmt.Errorf("dependency ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("functions/%s/dependencies/%s", functionID, dependencyID), nil, nil, nil)
}

// ListFunctionDependencies lists all dependencies attached to a function
func (c *Client) ListFunctionDependencies(ctx context.Context, functionID string) ([]DependencyResponse, error) {
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}

	var dependencies []DependencyResponse
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("functions/%s/dependencies", functionID), nil, nil, &dependencies)
	if err != nil {
		return nil, err
	}

	return dependencies, nil
}
