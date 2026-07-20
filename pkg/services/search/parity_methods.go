// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// normalize backfills the iterator-facing alias fields from the wire fields.
func (r *SearchResponse) normalize() {
	r.Results = r.GMeta
	r.HasMore = r.HasNextPage
}

// GetSearch performs a GET search query (upstream "search").
func (c *Client) GetSearch(ctx context.Context, indexID, q string, offset, limit int, advanced bool) (*SearchResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	if q == "" {
		return nil, fmt.Errorf("query string is required")
	}
	query := url.Values{}
	query.Set("q", q)
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if advanced {
		query.Set("advanced", "true")
	}
	var resp SearchResponse
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "index/"+indexID+"/search", query, nil, &resp); err != nil {
		return nil, err
	}
	resp.normalize()
	return &resp, nil
}

// Scroll performs a scroll (marker-paginated) query (POST /index/{id}/scroll).
func (c *Client) Scroll(ctx context.Context, indexID string, query map[string]interface{}) (*SearchResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	var resp SearchResponse
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "index/"+indexID+"/scroll", nil, query, &resp); err != nil {
		return nil, err
	}
	resp.normalize()
	return &resp, nil
}

// DeleteByQuery deletes documents matching a query (POST /index/{id}/delete_by_query).
func (c *Client) DeleteByQuery(ctx context.Context, indexID string, query map[string]interface{}) (*DeleteDocumentsResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	var resp DeleteDocumentsResponse
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "index/"+indexID+"/delete_by_query", nil, query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSubject retrieves all entries for a subject
// (GET /index/{id}/subject?subject=...).
func (c *Client) GetSubject(ctx context.Context, indexID, subject string) (map[string]interface{}, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	query := url.Values{}
	query.Set("subject", subject)
	var resp map[string]interface{}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "index/"+indexID+"/subject", query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteSubject deletes all entries for a subject
// (DELETE /index/{id}/subject?subject=...).
func (c *Client) DeleteSubject(ctx context.Context, indexID, subject string) (*DeleteDocumentsResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	query := url.Values{}
	query.Set("subject", subject)
	var resp DeleteDocumentsResponse
	if err := c.doRequestLowLevel(ctx, http.MethodDelete, "index/"+indexID+"/subject", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEntry retrieves a specific entry
// (GET /index/{id}/entry?subject=&entry_id=). entryID is optional.
func (c *Client) GetEntry(ctx context.Context, indexID, subject, entryID string) (map[string]interface{}, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	query := url.Values{}
	query.Set("subject", subject)
	if entryID != "" {
		query.Set("entry_id", entryID)
	}
	var resp map[string]interface{}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "index/"+indexID+"/entry", query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteEntry deletes a specific entry
// (DELETE /index/{id}/entry?subject=&entry_id=). entryID is optional.
func (c *Client) DeleteEntry(ctx context.Context, indexID, subject, entryID string) (*DeleteDocumentsResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	query := url.Values{}
	query.Set("subject", subject)
	if entryID != "" {
		query.Set("entry_id", entryID)
	}
	var resp DeleteDocumentsResponse
	if err := c.doRequestLowLevel(ctx, http.MethodDelete, "index/"+indexID+"/entry", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTaskList lists the tasks for an index (GET /task_list/{index_id}).
func (c *Client) GetTaskList(ctx context.Context, indexID string) ([]TaskStatusResponse, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	var envelope struct {
		Tasks []TaskStatusResponse `json:"tasks"`
	}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "task_list/"+indexID, nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Tasks, nil
}

// SearchRole represents a role assignment on a search index.
type SearchRole struct {
	ID        string `json:"id"`
	IndexID   string `json:"index_id,omitempty"`
	RoleName  string `json:"role_name"`
	Principal string `json:"principal"`
}

// CreateRole creates a role on an index (POST /index/{id}/role).
func (c *Client) CreateRole(ctx context.Context, indexID, roleName, principal string) (*SearchRole, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	body := map[string]string{"role_name": roleName, "principal": principal}
	var role SearchRole
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "index/"+indexID+"/role", nil, body, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleList lists roles on an index (GET /index/{id}/role_list).
func (c *Client) GetRoleList(ctx context.Context, indexID string) ([]SearchRole, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}
	var envelope struct {
		RoleList []SearchRole `json:"role_list"`
	}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "index/"+indexID+"/role_list", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.RoleList, nil
}

// DeleteRole deletes a role from an index (DELETE /index/{id}/role/{role_id}).
func (c *Client) DeleteRole(ctx context.Context, indexID, roleID string) error {
	if indexID == "" {
		return fmt.Errorf("index ID is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodDelete, "index/"+indexID+"/role/"+roleID, nil, nil, nil)
}
