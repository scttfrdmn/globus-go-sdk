// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CreateEnvironment creates a new environment configuration for Compute functions
func (c *Client) CreateEnvironment(ctx context.Context, request *EnvironmentCreateRequest) (*EnvironmentResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("environment create request is required")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("environment name is required")
	}

	var environment EnvironmentResponse
	err := c.doRequest(ctx, http.MethodPost, "environments", nil, request, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// GetEnvironment retrieves an environment configuration by ID
func (c *Client) GetEnvironment(ctx context.Context, environmentID string) (*EnvironmentResponse, error) {
	if environmentID == "" {
		return nil, fmt.Errorf("environment ID is required")
	}

	var environment EnvironmentResponse
	err := c.doRequest(ctx, http.MethodGet, "environments/"+environmentID, nil, nil, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// ListEnvironments lists all environment configurations the user has access to
func (c *Client) ListEnvironments(ctx context.Context, options *ListEnvironmentsOptions) (*EnvironmentList, error) {
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

	var environmentList EnvironmentList
	err := c.doRequest(ctx, http.MethodGet, "environments", query, nil, &environmentList)
	if err != nil {
		return nil, err
	}

	return &environmentList, nil
}

// UpdateEnvironment updates an existing environment configuration
func (c *Client) UpdateEnvironment(ctx context.Context, environmentID string, request *EnvironmentUpdateRequest) (*EnvironmentResponse, error) {
	if environmentID == "" {
		return nil, fmt.Errorf("environment ID is required")
	}

	if request == nil {
		return nil, fmt.Errorf("environment update request is required")
	}

	var environment EnvironmentResponse
	err := c.doRequest(ctx, http.MethodPut, "environments/"+environmentID, nil, request, &environment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// DeleteEnvironment deletes an environment configuration
func (c *Client) DeleteEnvironment(ctx context.Context, environmentID string) error {
	if environmentID == "" {
		return fmt.Errorf("environment ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "environments/"+environmentID, nil, nil, nil)
}

// ApplyEnvironmentToTask applies an environment configuration to a task request
func (c *Client) ApplyEnvironmentToTask(ctx context.Context, taskRequest *TaskRequest, environmentID string) (*TaskRequest, error) {
	if taskRequest == nil {
		return nil, fmt.Errorf("task request is required")
	}

	if environmentID == "" {
		return nil, fmt.Errorf("environment ID is required")
	}

	// Get the environment configuration
	environment, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment configuration: %w", err)
	}

	// Create a copy of the task request
	enrichedRequest := *taskRequest

	// Apply environment variables
	if len(environment.Variables) > 0 {
		// Initialize kwargs if nil
		if enrichedRequest.Kwargs == nil {
			enrichedRequest.Kwargs = make(map[string]interface{})
		}

		// Add environment variables to kwargs
		enrichedRequest.Kwargs["environment"] = environment.Variables
	}

	// Apply resource allocation if specified
	if environment.Resources != nil {
		if enrichedRequest.Kwargs == nil {
			enrichedRequest.Kwargs = make(map[string]interface{})
		}

		// Add resource configurations to kwargs
		enrichedRequest.Kwargs["resources"] = environment.Resources
	}

	return &enrichedRequest, nil
}

// CreateSecret creates a new secret for use in compute environments
func (c *Client) CreateSecret(ctx context.Context, request *SecretCreateRequest) (*SecretResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("secret create request is required")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("secret name is required")
	}

	if request.Value == "" {
		return nil, fmt.Errorf("secret value is required")
	}

	var secret SecretResponse
	err := c.doRequest(ctx, http.MethodPost, "secrets", nil, request, &secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

// ListSecrets lists all secrets the user has access to
func (c *Client) ListSecrets(ctx context.Context) ([]SecretResponse, error) {
	var secrets []SecretResponse
	err := c.doRequest(ctx, http.MethodGet, "secrets", nil, nil, &secrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// DeleteSecret deletes a secret
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	if secretID == "" {
		return fmt.Errorf("secret ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "secrets/"+secretID, nil, nil, nil)
}