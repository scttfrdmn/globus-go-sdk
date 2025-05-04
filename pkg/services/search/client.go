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

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
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
func NewClient(opts ...ClientOption) (*Client, error) {
	// Apply default options
	options := defaultOptions()
	
	// Apply user options
	for _, opt := range opts {
		opt(options)
	}
	
	// If an access token was provided, create a static token authorizer
	if options.accessToken != "" {
		authorizer := authorizers.StaticTokenCoreAuthorizer(options.accessToken)
		options.coreOptions = append(options.coreOptions, core.WithAuthorizer(authorizer))
	}
	
	// Create the base client
	baseClient := core.NewClient(options.coreOptions...)
	
	return &Client{
		Client: baseClient,
	}, nil
}

// buildURLLowLevel builds a URL for the search API
// This is an internal method used by the client.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
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

// doRequestLowLevel performs an HTTP request and decodes the JSON response
// This is an internal method used by higher-level API methods.
func (c *Client) doRequestLowLevel(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	url := c.buildURLLowLevel(path, query)

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

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		}

		// Try to parse as JSON error
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Message != "" {
				return &SearchError{
					Code:      errorResp.Code,
					Message:   errorResp.Message,
					Status:    resp.StatusCode,
					RequestID: errorResp.RequestID,
				}
			}
		}

		// Fallback to generic error message
		return &SearchError{
			Code:    fmt.Sprintf("HTTP%d", resp.StatusCode),
			Message: fmt.Sprintf("request failed with status %d: %s", resp.StatusCode, string(respBody)),
			Status:  resp.StatusCode,
		}
	}

	// For empty responses, return early
	if len(respBody) == 0 || response == nil {
		return nil
	}

	// Parse the response body
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "index_list", query, nil, &indexList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "index/"+indexID, nil, nil, &index)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "index", nil, request, &index)
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
	err := c.doRequestLowLevel(ctx, http.MethodPatch, "index/"+indexID, nil, request, &index)
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

	return c.doRequestLowLevel(ctx, http.MethodDelete, "index/"+indexID, nil, nil, nil)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "ingest", nil, request, &response)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "search", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// StructuredSearch performs a search query with a structured query object
func (c *Client) StructuredSearch(ctx context.Context, request *StructuredSearchRequest) (*SearchResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("search request is required")
	}

	if request.IndexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}

	if request.Query == nil {
		return nil, fmt.Errorf("query is required")
	}

	var response SearchResponse
	err := c.doRequestLowLevel(ctx, http.MethodPost, "search", nil, request, &response)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "delete", nil, request, &response)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID, nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SearchIterator provides a way to iterate through search results
type SearchIterator struct {
	client        *Client
	ctx           context.Context
	request       *SearchRequest
	structRequest *StructuredSearchRequest
	pageSize      int
	currentResp   *SearchResponse
	hasMore       bool
	err           error
}

// NewSearchIterator creates a new search iterator
func (c *Client) NewSearchIterator(ctx context.Context, request *SearchRequest, pageSize int) *SearchIterator {
	if pageSize <= 0 {
		pageSize = 100
	}

	// Copy the request and set the page size
	reqCopy := *request
	if reqCopy.Options == nil {
		reqCopy.Options = &SearchOptions{}
	}
	reqCopy.Options.Limit = pageSize

	return &SearchIterator{
		client:   c,
		ctx:      ctx,
		request:  &reqCopy,
		pageSize: pageSize,
		hasMore:  true,
	}
}

// NewStructuredSearchIterator creates a new structured search iterator
func (c *Client) NewStructuredSearchIterator(ctx context.Context, request *StructuredSearchRequest, pageSize int) *SearchIterator {
	if pageSize <= 0 {
		pageSize = 100
	}

	// Copy the request and set the page size
	reqCopy := *request
	if reqCopy.Options == nil {
		reqCopy.Options = &SearchOptions{}
	}
	reqCopy.Options.Limit = pageSize

	return &SearchIterator{
		client:        c,
		ctx:           ctx,
		structRequest: &reqCopy,
		pageSize:      pageSize,
		hasMore:       true,
	}
}

// Next fetches the next page of results
func (it *SearchIterator) Next() bool {
	// Don't proceed if we're already at the end or have an error
	if !it.hasMore || it.err != nil {
		return false
	}

	var resp *SearchResponse
	var err error

	if it.structRequest != nil {
		// For structured search
		resp, err = it.client.StructuredSearch(it.ctx, it.structRequest)
	} else {
		// For regular search
		resp, err = it.client.Search(it.ctx, it.request)
	}

	if err != nil {
		it.err = err
		return false
	}

	it.currentResp = resp
	it.hasMore = resp.HasMore

	// Update the page token for the next request only if there are more pages
	if it.hasMore {
		if it.structRequest != nil {
			if it.structRequest.Options == nil {
				it.structRequest.Options = &SearchOptions{}
			}
			it.structRequest.Options.PageToken = resp.PageToken
		} else {
			if it.request.Options == nil {
				it.request.Options = &SearchOptions{}
			}
			it.request.Options.PageToken = resp.PageToken
		}
	}

	return true
}

// Response returns the current page of results
func (it *SearchIterator) Response() *SearchResponse {
	return it.currentResp
}

// Error returns any error that occurred during iteration
func (it *SearchIterator) Error() error {
	return it.err
}

// SearchAll retrieves all search results across multiple pages
func (c *Client) SearchAll(ctx context.Context, request *SearchRequest, pageSize int) ([]SearchResult, error) {
	if pageSize <= 0 {
		pageSize = 100
	}

	// Create iterator
	it := c.NewSearchIterator(ctx, request, pageSize)

	// Collect all results
	var allResults []SearchResult

	for it.Next() {
		resp := it.Response()
		allResults = append(allResults, resp.Results...)
	}

	if it.Error() != nil {
		return nil, it.Error()
	}

	return allResults, nil
}

// StructuredSearchAll retrieves all search results across multiple pages for a structured search
func (c *Client) StructuredSearchAll(ctx context.Context, request *StructuredSearchRequest, pageSize int) ([]SearchResult, error) {
	if pageSize <= 0 {
		pageSize = 100
	}

	// Create iterator
	it := c.NewStructuredSearchIterator(ctx, request, pageSize)

	// Collect all results
	var allResults []SearchResult

	for it.Next() {
		resp := it.Response()
		allResults = append(allResults, resp.Results...)
	}

	if it.Error() != nil {
		return nil, it.Error()
	}

	return allResults, nil
}
