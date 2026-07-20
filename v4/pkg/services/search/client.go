// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Search service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Search client
// In v4, config is required and must include explicit scopes
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	// Set default Search service URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://search.api.globus.org/v1"
	}

	// Create base client
	baseClient, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    config.BaseURL,
	}, nil
}

// GetIndex retrieves information about a search index
// v4: Context is always first parameter
func (c *Client) GetIndex(ctx context.Context, indexID string) (*Index, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}

	var index Index
	path := fmt.Sprintf("/index/%s", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &index)
	if err != nil {
		return nil, err
	}
	return &index, nil
}

// CreateIndex creates a new search index
// v4: Context is always first parameter
func (c *Client) CreateIndex(ctx context.Context, index *IndexCreate) (*Index, error) {
	if index == nil {
		return nil, &core.ValidationError{
			Field:   "index",
			Message: "index data is required",
		}
	}
	if index.DisplayName == "" {
		return nil, &core.ValidationError{
			Field:   "DisplayName",
			Message: "index display name is required",
		}
	}

	var result Index
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/index", nil, index, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateIndex updates an existing search index
// v4: Context is always first parameter
func (c *Client) UpdateIndex(ctx context.Context, indexID string, update *IndexUpdate) (*Index, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if update == nil {
		return nil, &core.ValidationError{
			Field:   "update",
			Message: "update data is required",
		}
	}

	var result Index
	path := fmt.Sprintf("/index/%s", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodPatch, path, nil, update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteIndex deletes a search index
// v4: Context is always first parameter
func (c *Client) DeleteIndex(ctx context.Context, indexID string) error {
	if indexID == "" {
		return &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}

	path := fmt.Sprintf("/index/%s", indexID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}

// Search performs a search query on an index via POST (upstream post_search).
// v4: Context is always first parameter
func (c *Client) Search(ctx context.Context, indexID string, query *SearchQuery) (*SearchResults, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if query == nil {
		return nil, &core.ValidationError{Field: "query", Message: "query data is required"}
	}
	if query.Version == "" {
		query.Version = SearchQueryVersion
	}

	var results SearchResults
	path := fmt.Sprintf("/index/%s/search", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, query, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

// SearchGet performs a search query on an index via GET (upstream search).
func (c *Client) SearchGet(ctx context.Context, indexID string, opts *SearchGetOptions) (*SearchResults, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if opts == nil || opts.Q == "" {
		return nil, &core.ValidationError{Field: "q", Message: "query string is required"}
	}

	query := url.Values{}
	query.Set("q", opts.Q)
	if opts.Offset > 0 {
		query.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Advanced {
		query.Set("advanced", "true")
	}

	var results SearchResults
	path := fmt.Sprintf("/index/%s/search", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

// Scroll performs a scroll (marker-paginated) query (POST /index/{id}/scroll).
func (c *Client) Scroll(ctx context.Context, indexID string, query *ScrollQuery) (*SearchResults, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if query == nil {
		return nil, &core.ValidationError{Field: "query", Message: "query data is required"}
	}

	var results SearchResults
	path := fmt.Sprintf("/index/%s/scroll", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, query, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

// Ingest submits an ingest document (POST /index/{id}/ingest). Pass a document
// built with NewGMetaEntryIngest or NewGMetaListIngest (or any value that
// marshals to a valid {ingest_type, ingest_data} document).
func (c *Client) Ingest(ctx context.Context, indexID string, data interface{}) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "ingest document is required"}
	}

	var response IngestResponse
	path := fmt.Sprintf("/index/%s/ingest", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteByQuery deletes all documents matching a query document
// (POST /index/{id}/delete_by_query).
func (c *Client) DeleteByQuery(ctx context.Context, indexID string, data interface{}) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "query document is required"}
	}

	var response IngestResponse
	path := fmt.Sprintf("/index/%s/delete_by_query", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// BatchDeleteBySubject deletes documents by subject
// (POST /index/{id}/batch_delete_by_subject). additionalParams are merged into
// the request body.
func (c *Client) BatchDeleteBySubject(ctx context.Context, indexID string, subjects []string, additionalParams map[string]interface{}) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if len(subjects) == 0 {
		return nil, &core.ValidationError{Field: "subjects", Message: "at least one subject is required"}
	}

	body := map[string]interface{}{"subjects": subjects}
	for k, v := range additionalParams {
		if k != "subjects" {
			body[k] = v
		}
	}

	var response IngestResponse
	path := fmt.Sprintf("/index/%s/batch_delete_by_subject", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetSubject retrieves all entries for a subject
// (GET /index/{id}/subject?subject=...). subject is a query param.
func (c *Client) GetSubject(ctx context.Context, indexID, subject string) (*GMetaResult, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if subject == "" {
		return nil, &core.ValidationError{Field: "subject", Message: "subject is required"}
	}

	query := url.Values{}
	query.Set("subject", subject)
	var result GMetaResult
	path := fmt.Sprintf("/index/%s/subject", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSubject deletes all entries for a subject
// (DELETE /index/{id}/subject?subject=...).
func (c *Client) DeleteSubject(ctx context.Context, indexID, subject string) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if subject == "" {
		return nil, &core.ValidationError{Field: "subject", Message: "subject is required"}
	}

	query := url.Values{}
	query.Set("subject", subject)
	var response IngestResponse
	path := fmt.Sprintf("/index/%s/subject", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodDelete, path, query, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetEntry retrieves a specific entry (GET /index/{id}/entry?subject=&entry_id=).
// entryID is optional; pass "" to omit it.
func (c *Client) GetEntry(ctx context.Context, indexID, subject, entryID string) (*GMetaResult, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if subject == "" {
		return nil, &core.ValidationError{Field: "subject", Message: "subject is required"}
	}

	query := url.Values{}
	query.Set("subject", subject)
	if entryID != "" {
		query.Set("entry_id", entryID)
	}
	var result GMetaResult
	path := fmt.Sprintf("/index/%s/entry", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEntry deletes a specific entry
// (DELETE /index/{id}/entry?subject=&entry_id=). entryID is optional.
func (c *Client) DeleteEntry(ctx context.Context, indexID, subject, entryID string) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if subject == "" {
		return nil, &core.ValidationError{Field: "subject", Message: "subject is required"}
	}

	query := url.Values{}
	query.Set("subject", subject)
	if entryID != "" {
		query.Set("entry_id", entryID)
	}
	var response IngestResponse
	path := fmt.Sprintf("/index/%s/entry", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodDelete, path, query, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ListRoles lists roles for a search index (GET /index/{id}/role_list).
// v4: Context is always first parameter
func (c *Client) ListRoles(ctx context.Context, indexID string) (*RoleList, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}

	var roleList RoleList
	path := fmt.Sprintf("/index/%s/role_list", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &roleList); err != nil {
		return nil, err
	}
	return &roleList, nil
}

// AddRole adds a role to a search index (POST /index/{id}/role).
// v4: Context is always first parameter
func (c *Client) AddRole(ctx context.Context, indexID string, role *RoleCreate) (*Role, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if role == nil || role.RoleName == "" || role.Principal == "" {
		return nil, &core.ValidationError{Field: "role", Message: "role_name and principal are required"}
	}

	var result Role
	path := fmt.Sprintf("/index/%s/role", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, role, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveRole removes a role from a search index (DELETE /index/{id}/role/{role_id}).
// v4: Context is always first parameter
func (c *Client) RemoveRole(ctx context.Context, indexID, roleID string) error {
	if indexID == "" {
		return &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}
	if roleID == "" {
		return &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}

	path := fmt.Sprintf("/index/%s/role/%s", indexID, roleID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}

// ReopenIndex reopens a previously deleted index.
// Added in Python SDK v4.0.0b1.
func (c *Client) ReopenIndex(ctx context.Context, indexID string) (*Index, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}

	var index Index
	path := fmt.Sprintf("/index/%s/reopen", indexID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, nil, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// IndexList lists search indexes with optional role filtering (GET /index_list).
// FilterRoles is sent as a single comma-joined query param. Not paginated upstream.
func (c *Client) IndexList(ctx context.Context, options *ListIndexesOptions) (*IndexList, error) {
	query := url.Values{}
	if options != nil && len(options.FilterRoles) > 0 {
		query.Set("filter_roles", strings.Join(options.FilterRoles, ","))
	}

	var indexList IndexList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/index_list", query, nil, &indexList)
	if err != nil {
		return nil, err
	}
	return &indexList, nil
}

// GetTask retrieves a task by ID (GET /task/{task_id}). It is index-independent.
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}

	var task Task
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/task/%s", taskID), nil, nil, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskList lists the tasks for an index (GET /task_list/{index_id}).
func (c *Client) GetTaskList(ctx context.Context, indexID string) (*TaskList, error) {
	if indexID == "" {
		return nil, &core.ValidationError{Field: "indexID", Message: "index ID is required"}
	}

	var taskList TaskList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/task_list/%s", indexID), nil, nil, &taskList); err != nil {
		return nil, err
	}
	return &taskList, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}
