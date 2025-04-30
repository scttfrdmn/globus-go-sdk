// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RegisterContainer registers a new container with Globus Compute
func (c *Client) RegisterContainer(ctx context.Context, request *ContainerRegistrationRequest) (*ContainerResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("container registration request is required")
	}

	if request.Image == "" {
		return nil, fmt.Errorf("container image is required")
	}

	var container ContainerResponse
	err := c.doRequest(ctx, http.MethodPost, "containers", nil, request, &container)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// GetContainer retrieves a container by ID
func (c *Client) GetContainer(ctx context.Context, containerID string) (*ContainerResponse, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID is required")
	}

	var container ContainerResponse
	err := c.doRequest(ctx, http.MethodGet, "containers/"+containerID, nil, nil, &container)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// ListContainers lists all containers the user has access to
func (c *Client) ListContainers(ctx context.Context, options *ListContainersOptions) (*ContainerList, error) {
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

	var containerList ContainerList
	err := c.doRequest(ctx, http.MethodGet, "containers", query, nil, &containerList)
	if err != nil {
		return nil, err
	}

	return &containerList, nil
}

// UpdateContainer updates an existing container
func (c *Client) UpdateContainer(ctx context.Context, containerID string, request *ContainerUpdateRequest) (*ContainerResponse, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID is required")
	}

	if request == nil {
		return nil, fmt.Errorf("container update request is required")
	}

	var container ContainerResponse
	err := c.doRequest(ctx, http.MethodPut, "containers/"+containerID, nil, request, &container)
	if err != nil {
		return nil, err
	}

	return &container, nil
}

// DeleteContainer deletes a container
func (c *Client) DeleteContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "containers/"+containerID, nil, nil, nil)
}

// RunContainerFunction runs a function within a container on a specific endpoint
func (c *Client) RunContainerFunction(ctx context.Context, request *ContainerTaskRequest) (*TaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("container task request is required")
	}

	if request.FunctionID == "" && request.Code == "" {
		return nil, fmt.Errorf("either function ID or code is required")
	}

	if request.ContainerID == "" {
		return nil, fmt.Errorf("container ID is required")
	}

	if request.EndpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var response TaskResponse
	err := c.doRequest(ctx, http.MethodPost, "run_container", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}