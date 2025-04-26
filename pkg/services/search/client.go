// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/yourusername/globus-go-sdk/pkg/core"
	"github.com/yourusername/globus-go-sdk/pkg/core/authorizers"
)

// Constants for Globus Search
const (
	DefaultBaseURL = "https://search.api.globus.org/v1/"
	SearchScope    = "urn:globus:auth:scope:search.api.globus.org:all"
)

// Client provides methods for interacting with Globus Search
type Client struct {
	Client *core.Client
}

// NewClient creates a new Search client
func NewClient(accessToken string, options ...core.ClientOption) *Client {
	// Create the authorizer with the access token
	authorizer := authorizers.NewStaticTokenAuthorizer(accessToken)
	
	// Apply default options specific to Search
	defaultOptions := []core.ClientOption{
		core.WithBaseURL(DefaultBaseURL),
		core.WithAuthorizer(authorizer),
	}
	
	// Merge with user options
	options = append(defaultOptions, options...)
	
	// Create the base client
	baseClient := core.NewClient(options...)
	
	return &Client{
		Client: baseClient,
	}
}

// buildURL builds a URL for the search API
func (c *Client) buildURL(path string, query url.Values) string {
	baseURL := c.Client.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	
	url := baseURL + path
	if query != nil && len(query) > 0 {
		url += "?" + query.Encode()
	}
	
	return url
}

// doRequest performs an HTTP request and decodes the JSON response
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	url := c.buildURL(path, query)
	
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// For non-GET requests with no response body, just check status
	if method != http.MethodGet && response == nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Read and decode response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	if len(respBody) == 0 {
		return nil
	}
	
	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return nil
}

// ListIndexes retrieves indexes the user has access to
func (c *Client) ListIndexes(ctx context.Context, options *ListIndexesOptions) (*IndexList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		} else if options.PerPage > 0 {
			query.Set("limit", strconv.Itoa(options.PerPage))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.IsPublic {
			query.Set("is_public", "true")
		}
		if options.IsActive {
			query.Set("is_active", "true")
		}
		if options.CreatedBy != "" {
			query.Set("created_by", options.CreatedBy)
		}
		if options.ByPath != "" {
			query.Set("by_path", options.ByPath)
		}
	}
	
	var indexList IndexList
	err := c.doRequest(ctx, http.MethodGet, "index_list", query, nil, &indexList)
	if err != nil {
		return nil, err
	}
	
	return &indexList, nil
}

// GetIndex retrieves a specific index by ID
func (c *Client) GetIndex(ctx context.Context, indexID string) (*Index, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	
	var index Index
	err := c.doRequest(ctx, http.MethodGet, "index/"+indexID, nil, nil, &index)
	if err != nil {
		return nil, err
	}
	
	return &index, nil
}

// CreateIndex creates a new index
func (c *Client) CreateIndex(ctx context.Context, request *IndexCreateRequest) (*Index, error) {
	if request == nil {
		return nil, fmt.Errorf("index create request is required")
	}
	
	if request.DisplayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	
	var index Index
	err := c.doRequest(ctx, http.MethodPost, "index", nil, request, &index)
	if err != nil {
		return nil, err
	}
	
	return &index, nil
}

// UpdateIndex updates an existing index
func (c *Client) UpdateIndex(ctx context.Context, indexID string, request *IndexUpdateRequest) (*Index, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	
	if request == nil {
		return nil, fmt.Errorf("index update request is required")
	}
	
	var index Index
	err := c.doRequest(ctx, http.MethodPatch, "index/"+indexID, nil, request, &index)
	if err != nil {
		return nil, err
	}
	
	return &index, nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, indexID string) error {
	if indexID == "" {
		return fmt.Errorf("index ID is required")
	}
	
	return c.doRequest(ctx, http.MethodDelete, "index/"+indexID, nil, nil, nil)
}

// IngestDocuments ingests documents into an index
func (c *Client) IngestDocuments(ctx context.Context, request *IngestRequest) (*IngestResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("ingest request is required")
	}
	
	if request.IndexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	
	if len(request.Documents) == 0 {
		return nil, fmt.Errorf("at least one document is required")
	}
	
	var response IngestResponse
	err := c.doRequest(ctx, http.MethodPost, "ingest", nil, request, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// Search performs a search query
func (c *Client) Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("search request is required")
	}
	
	if request.IndexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	
	var response SearchResponse
	err := c.doRequest(ctx, http.MethodPost, "search", nil, request, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// DeleteDocuments deletes documents from an index
func (c *Client) DeleteDocuments(ctx context.Context, request *DeleteDocumentsRequest) (*DeleteDocumentsResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("delete request is required")
	}
	
	if request.IndexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	
	if len(request.Subjects) == 0 {
		return nil, fmt.Errorf("at least one subject is required")
	}
	
	var response DeleteDocumentsResponse
	err := c.doRequest(ctx, http.MethodPost, "delete", nil, request, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// GetTaskStatus retrieves the status of a task
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	
	var response TaskStatusResponse
	err := c.doRequest(ctx, http.MethodGet, "task/"+taskID, nil, nil, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}