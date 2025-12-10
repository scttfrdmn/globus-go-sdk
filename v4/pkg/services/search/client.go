// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
	err := c.baseClient.DoRequest(ctx, http.MethodPut, path, nil, update, &result)
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

// Search performs a search query on an index
// v4: Context is always first parameter
func (c *Client) Search(ctx context.Context, indexID string, query *SearchQuery) (*SearchResults, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if query == nil {
		return nil, &core.ValidationError{
			Field:   "query",
			Message: "query data is required",
		}
	}

	var results SearchResults
	path := fmt.Sprintf("/index/%s/search", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, query, &results)
	if err != nil {
		return nil, err
	}
	return &results, nil
}

// IngestEntry adds or updates a document in a search index
// v4: Context is always first parameter
func (c *Client) IngestEntry(ctx context.Context, indexID string, entry *IngestEntry) (*IngestResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if entry == nil {
		return nil, &core.ValidationError{
			Field:   "entry",
			Message: "entry data is required",
		}
	}
	if entry.Subject == "" {
		return nil, &core.ValidationError{
			Field:   "Subject",
			Message: "entry subject is required",
		}
	}

	var response IngestResponse
	path := fmt.Sprintf("/index/%s/ingest", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, entry, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// IngestBatch adds or updates multiple documents in a search index
// v4: Context is always first parameter
func (c *Client) IngestBatch(ctx context.Context, indexID string, batch *IngestBatch) (*IngestBatchResponse, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if batch == nil {
		return nil, &core.ValidationError{
			Field:   "batch",
			Message: "batch data is required",
		}
	}
	if len(batch.Entries) == 0 {
		return nil, &core.ValidationError{
			Field:   "Entries",
			Message: "at least one entry is required",
		}
	}

	var response IngestBatchResponse
	path := fmt.Sprintf("/index/%s/ingest", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, batch, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteEntry removes a document from a search index
// v4: Context is always first parameter
func (c *Client) DeleteEntry(ctx context.Context, indexID, subject string) error {
	if indexID == "" {
		return &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if subject == "" {
		return &core.ValidationError{
			Field:   "subject",
			Message: "subject is required",
		}
	}

	path := fmt.Sprintf("/index/%s/subject/%s", indexID, url.PathEscape(subject))
	return c.baseClient.DoRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}

// GetEntry retrieves a specific document from a search index
// v4: Context is always first parameter
func (c *Client) GetEntry(ctx context.Context, indexID, subject string) (*GMetaEntry, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if subject == "" {
		return nil, &core.ValidationError{
			Field:   "subject",
			Message: "subject is required",
		}
	}

	var entry GMetaEntry
	path := fmt.Sprintf("/index/%s/subject/%s", indexID, url.PathEscape(subject))
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetRole retrieves role information for a search index
// v4: Context is always first parameter
func (c *Client) GetRole(ctx context.Context, indexID, roleID string) (*Role, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if roleID == "" {
		return nil, &core.ValidationError{
			Field:   "roleID",
			Message: "role ID is required",
		}
	}

	var role Role
	path := fmt.Sprintf("/index/%s/role/%s", indexID, roleID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles lists roles for a search index
// v4: Context is always first parameter
func (c *Client) ListRoles(ctx context.Context, indexID string) (*RoleList, error) {
	if indexID == "" {
		return nil, &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}

	var roleList RoleList
	path := fmt.Sprintf("/index/%s/role_list", indexID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &roleList)
	if err != nil {
		return nil, err
	}
	return &roleList, nil
}

// AddRole adds a role to a search index
// v4: Context is always first parameter
func (c *Client) AddRole(ctx context.Context, indexID, principal, roleID string) error {
	if indexID == "" {
		return &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if principal == "" {
		return &core.ValidationError{
			Field:   "principal",
			Message: "principal is required",
		}
	}
	if roleID == "" {
		return &core.ValidationError{
			Field:   "roleID",
			Message: "role ID is required",
		}
	}

	body := map[string]interface{}{
		"principal":   principal,
		"role_id":     roleID,
	}

	path := fmt.Sprintf("/index/%s/role", indexID)
	return c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, body, nil)
}

// RemoveRole removes a role from a search index
// v4: Context is always first parameter
func (c *Client) RemoveRole(ctx context.Context, indexID, roleID string) error {
	if indexID == "" {
		return &core.ValidationError{
			Field:   "indexID",
			Message: "index ID is required",
		}
	}
	if roleID == "" {
		return &core.ValidationError{
			Field:   "roleID",
			Message: "role ID is required",
		}
	}

	path := fmt.Sprintf("/index/%s/role/%s", indexID, roleID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}
// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

