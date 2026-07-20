// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// This file adds transfer methods required for parity with Python globus-sdk
// 3.65.0 that were absent from the original client: endpoint update/delete,
// shared-endpoint creation and listing, server/role/ACL families, bookmarks,
// operation_stat, task sub-lists, and the endpoint_manager surface.

// commajoin joins a list the way upstream utils.commajoin encodes list query
// params (comma-separated, no spaces).
func commajoin(vs []string) string { return strings.Join(vs, ",") }

// ─── Endpoint update / delete ────────────────────────────────────────────────

// UpdateEndpoint updates fields on an endpoint (PUT endpoint/{id}).
func (c *Client) UpdateEndpoint(ctx context.Context, endpointID string, data map[string]interface{}) (*Endpoint, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var ep Endpoint
	if err := c.doRequestLowLevel(ctx, http.MethodPut, "endpoint/"+endpointID, nil, data, &ep); err != nil {
		return nil, err
	}
	return &ep, nil
}

// DeleteEndpoint deletes an endpoint (DELETE endpoint/{id}).
func (c *Client) DeleteEndpoint(ctx context.Context, endpointID string) error {
	if endpointID == "" {
		return fmt.Errorf("endpoint ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodDelete, "endpoint/"+endpointID, nil, nil, nil)
}

// CreateSharedEndpoint creates a shared (guest) endpoint (POST shared_endpoint).
func (c *Client) CreateSharedEndpoint(ctx context.Context, data map[string]interface{}) (*Endpoint, error) {
	if data == nil {
		return nil, fmt.Errorf("shared endpoint data is required")
	}
	if _, ok := data["DATA_TYPE"]; !ok {
		data["DATA_TYPE"] = "shared_endpoint"
	}
	var ep Endpoint
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "shared_endpoint", nil, data, &ep); err != nil {
		return nil, err
	}
	return &ep, nil
}

// ─── Pause / shared-endpoint listings ────────────────────────────────────────

// MyEffectivePauseRuleList lists pause rules affecting the caller on an endpoint.
func (c *Client) MyEffectivePauseRuleList(ctx context.Context, endpointID string) (*PauseRuleList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out PauseRuleList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/my_effective_pause_rule_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MySharedEndpointList lists shared endpoints the caller created on a host.
func (c *Client) MySharedEndpointList(ctx context.Context, endpointID string) (*EndpointList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out EndpointList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/my_shared_endpoint_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSharedEndpointList lists shared endpoints hosted on an endpoint
// (next_token paginated, items under "shared_endpoints").
func (c *Client) GetSharedEndpointList(ctx context.Context, endpointID string, opts *SharedEndpointListOptions) (*SharedEndpointList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	query := url.Values{}
	if opts != nil {
		if opts.MaxResults > 0 {
			query.Set("max_results", strconv.Itoa(opts.MaxResults))
		}
		if opts.NextToken != "" {
			query.Set("next_token", opts.NextToken)
		}
	}
	var out SharedEndpointList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/shared_endpoint_list", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─── Servers ─────────────────────────────────────────────────────────────────

// EndpointServerList lists the servers backing an endpoint (server_list).
func (c *Client) EndpointServerList(ctx context.Context, endpointID string) (*ServerList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out ServerList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/server_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEndpointServer retrieves a single server on an endpoint.
func (c *Client) GetEndpointServer(ctx context.Context, endpointID string, serverID int) (*Server, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out Server
	path := "endpoint/" + endpointID + "/server/" + strconv.Itoa(serverID)
	if err := c.doRequestLowLevel(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─── Roles ───────────────────────────────────────────────────────────────────

// EndpointRoleList lists role assignments on an endpoint (role_list).
func (c *Client) EndpointRoleList(ctx context.Context, endpointID string) (*RoleList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out RoleList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/role_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddEndpointRole assigns a role on an endpoint (POST endpoint/{id}/role).
func (c *Client) AddEndpointRole(ctx context.Context, endpointID string, roleData map[string]interface{}) (*OperationResult, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	if roleData == nil {
		return nil, fmt.Errorf("role data is required")
	}
	if _, ok := roleData["DATA_TYPE"]; !ok {
		roleData["DATA_TYPE"] = "role"
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint/"+endpointID+"/role", nil, roleData, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEndpointRole retrieves a single role on an endpoint.
func (c *Client) GetEndpointRole(ctx context.Context, endpointID, roleID string) (*Role, error) {
	if endpointID == "" || roleID == "" {
		return nil, fmt.Errorf("endpoint ID and role ID are required")
	}
	var out Role
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/role/"+roleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEndpointRole removes a role on an endpoint.
func (c *Client) DeleteEndpointRole(ctx context.Context, endpointID, roleID string) (*OperationResult, error) {
	if endpointID == "" || roleID == "" {
		return nil, fmt.Errorf("endpoint ID and role ID are required")
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodDelete, "endpoint/"+endpointID+"/role/"+roleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─── ACL rules ───────────────────────────────────────────────────────────────

// EndpointACLList lists access-control rules on an endpoint (access_list).
func (c *Client) EndpointACLList(ctx context.Context, endpointID string) (*ACLList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out ACLList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/access_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEndpointACLRule retrieves a single ACL rule.
func (c *Client) GetEndpointACLRule(ctx context.Context, endpointID, ruleID string) (*ACLRule, error) {
	if endpointID == "" || ruleID == "" {
		return nil, fmt.Errorf("endpoint ID and rule ID are required")
	}
	var out ACLRule
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID+"/access/"+ruleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddEndpointACLRule creates an ACL rule (POST endpoint/{id}/access).
func (c *Client) AddEndpointACLRule(ctx context.Context, endpointID string, ruleData map[string]interface{}) (*OperationResult, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	if ruleData == nil {
		return nil, fmt.Errorf("rule data is required")
	}
	if _, ok := ruleData["DATA_TYPE"]; !ok {
		ruleData["DATA_TYPE"] = "access"
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint/"+endpointID+"/access", nil, ruleData, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEndpointACLRule updates an ACL rule (PUT endpoint/{id}/access/{rule_id}).
func (c *Client) UpdateEndpointACLRule(ctx context.Context, endpointID, ruleID string, ruleData map[string]interface{}) (*OperationResult, error) {
	if endpointID == "" || ruleID == "" {
		return nil, fmt.Errorf("endpoint ID and rule ID are required")
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPut, "endpoint/"+endpointID+"/access/"+ruleID, nil, ruleData, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEndpointACLRule deletes an ACL rule.
func (c *Client) DeleteEndpointACLRule(ctx context.Context, endpointID, ruleID string) (*OperationResult, error) {
	if endpointID == "" || ruleID == "" {
		return nil, fmt.Errorf("endpoint ID and rule ID are required")
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodDelete, "endpoint/"+endpointID+"/access/"+ruleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─── Bookmarks ───────────────────────────────────────────────────────────────

// BookmarkList lists the caller's bookmarks (bookmark_list).
func (c *Client) BookmarkList(ctx context.Context) (*BookmarkList, error) {
	var out BookmarkList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "bookmark_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateBookmark creates a bookmark (POST bookmark).
func (c *Client) CreateBookmark(ctx context.Context, data map[string]interface{}) (*Bookmark, error) {
	if data == nil {
		return nil, fmt.Errorf("bookmark data is required")
	}
	if _, ok := data["DATA_TYPE"]; !ok {
		data["DATA_TYPE"] = "bookmark"
	}
	var out Bookmark
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "bookmark", nil, data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBookmark retrieves a single bookmark.
func (c *Client) GetBookmark(ctx context.Context, bookmarkID string) (*Bookmark, error) {
	if bookmarkID == "" {
		return nil, fmt.Errorf("bookmark ID is required")
	}
	var out Bookmark
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "bookmark/"+bookmarkID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBookmark updates a bookmark (PUT bookmark/{id}).
func (c *Client) UpdateBookmark(ctx context.Context, bookmarkID string, data map[string]interface{}) (*Bookmark, error) {
	if bookmarkID == "" {
		return nil, fmt.Errorf("bookmark ID is required")
	}
	var out Bookmark
	if err := c.doRequestLowLevel(ctx, http.MethodPut, "bookmark/"+bookmarkID, nil, data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBookmark deletes a bookmark.
func (c *Client) DeleteBookmark(ctx context.Context, bookmarkID string) error {
	if bookmarkID == "" {
		return fmt.Errorf("bookmark ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodDelete, "bookmark/"+bookmarkID, nil, nil, nil)
}

// ─── Operations ──────────────────────────────────────────────────────────────

// OperationStat stats a single path on an endpoint (operation_stat).
func (c *Client) OperationStat(ctx context.Context, endpointID, path string, opts *StatOptions) (*FileListItem, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	query := url.Values{}
	query.Set("path", path)
	if opts != nil && opts.LocalUser != "" {
		query.Set("local_user", opts.LocalUser)
	}
	var out FileListItem
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "operation/endpoint/"+endpointID+"/stat", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ─── Task sub-lists ──────────────────────────────────────────────────────────

// TaskEventList lists events for a task (event_list, limit/offset paged).
func (c *Client) TaskEventList(ctx context.Context, taskID string, opts *TaskEventListOptions) (*TaskEventListResponse, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	query := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("offset", strconv.Itoa(opts.Offset))
		}
	}
	var out TaskEventListResponse
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID+"/event_list", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTask updates a task's mutable fields (label, deadline).
func (c *Client) UpdateTask(ctx context.Context, taskID string, data map[string]interface{}) (*OperationResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPut, "task/"+taskID, nil, data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TaskPauseInfo returns pause information for a task.
func (c *Client) TaskPauseInfo(ctx context.Context, taskID string) (*PauseInfo, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out PauseInfo
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID+"/pause_info", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TaskSuccessfulTransfers lists successful transfers for a task (marker paged).
func (c *Client) TaskSuccessfulTransfers(ctx context.Context, taskID string, opts *MarkerPageOptions) (*SuccessfulTransfersList, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out SuccessfulTransfersList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID+"/successful_transfers", markerQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TaskSkippedErrors lists skipped errors for a task (marker paged).
func (c *Client) TaskSkippedErrors(ctx context.Context, taskID string, opts *MarkerPageOptions) (*SkippedErrorsList, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out SkippedErrorsList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID+"/skipped_errors", markerQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func markerQuery(opts *MarkerPageOptions) url.Values {
	query := url.Values{}
	if opts != nil && opts.Marker != "" {
		query.Set("marker", opts.Marker)
	}
	return query
}

// ─── Endpoint manager ────────────────────────────────────────────────────────

// EndpointManagerMonitoredEndpoints lists endpoints the caller monitors.
func (c *Client) EndpointManagerMonitoredEndpoints(ctx context.Context) (*EndpointList, error) {
	var out EndpointList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/monitored_endpoints", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerHostedEndpointList lists endpoints hosted on an endpoint.
func (c *Client) EndpointManagerHostedEndpointList(ctx context.Context, endpointID string) (*EndpointList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out EndpointList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/endpoint/"+endpointID+"/hosted_endpoint_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerGetEndpoint retrieves an endpoint via the manager view.
func (c *Client) EndpointManagerGetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out Endpoint
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/endpoint/"+endpointID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerACLList lists ACL rules via the manager view.
func (c *Client) EndpointManagerACLList(ctx context.Context, endpointID string) (*ACLList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	var out ACLList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/endpoint/"+endpointID+"/access_list", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerTaskList lists tasks via the manager view (last_key paged).
func (c *Client) EndpointManagerTaskList(ctx context.Context, opts *EndpointManagerTaskListOptions) (*TaskList, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.FilterStatus) > 0 {
			query.Set("filter_status", commajoin(opts.FilterStatus))
		}
		if len(opts.FilterTaskID) > 0 {
			query.Set("filter_task_id", commajoin(opts.FilterTaskID))
		}
		if opts.FilterOwnerID != "" {
			query.Set("filter_owner_id", opts.FilterOwnerID)
		}
		if opts.FilterEndpoint != "" {
			query.Set("filter_endpoint", opts.FilterEndpoint)
		}
		if opts.FilterEndpointUse != "" {
			query.Set("filter_endpoint_use", opts.FilterEndpointUse)
		}
		if opts.FilterIsPaused != nil {
			query.Set("filter_is_paused", strconv.FormatBool(*opts.FilterIsPaused))
		}
		if opts.FilterCompletionTime != "" {
			query.Set("filter_completion_time", opts.FilterCompletionTime)
		}
		if opts.FilterMinFaults > 0 {
			query.Set("filter_min_faults", strconv.Itoa(opts.FilterMinFaults))
		}
		if opts.FilterLocalUser != "" {
			query.Set("filter_local_user", opts.FilterLocalUser)
		}
		if opts.LastKey != "" {
			query.Set("last_key", opts.LastKey)
		}
	}
	var out TaskList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task_list", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerGetTask retrieves a task via the manager view.
func (c *Client) EndpointManagerGetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out Task
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task/"+taskID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerTaskEventList lists events for a task via the manager view.
func (c *Client) EndpointManagerTaskEventList(ctx context.Context, taskID string, opts *EMTaskEventListOptions) (*TaskEventListResponse, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	query := url.Values{}
	if opts != nil {
		if opts.Limit > 0 {
			query.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			query.Set("offset", strconv.Itoa(opts.Offset))
		}
		if opts.FilterIsError != nil {
			if *opts.FilterIsError {
				query.Set("filter_is_error", "1")
			} else {
				query.Set("filter_is_error", "0")
			}
		}
	}
	var out TaskEventListResponse
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task/"+taskID+"/event_list", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerTaskPauseInfo returns pause info for a task via the manager view.
func (c *Client) EndpointManagerTaskPauseInfo(ctx context.Context, taskID string) (*PauseInfo, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out PauseInfo
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task/"+taskID+"/pause_info", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerTaskSuccessfulTransfers lists successful transfers (manager view).
func (c *Client) EndpointManagerTaskSuccessfulTransfers(ctx context.Context, taskID string, opts *MarkerPageOptions) (*SuccessfulTransfersList, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out SuccessfulTransfersList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task/"+taskID+"/successful_transfers", markerQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerTaskSkippedErrors lists skipped errors (manager view).
func (c *Client) EndpointManagerTaskSkippedErrors(ctx context.Context, taskID string, opts *MarkerPageOptions) (*SkippedErrorsList, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	var out SkippedErrorsList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/task/"+taskID+"/skipped_errors", markerQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerCancelTasks requests admin cancellation of tasks.
func (c *Client) EndpointManagerCancelTasks(ctx context.Context, taskIDs []string, message string) (*AdminCancelStatus, error) {
	body := map[string]interface{}{
		"message":      message,
		"task_id_list": taskIDs,
	}
	var out AdminCancelStatus
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint_manager/admin_cancel", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerCancelStatus polls an admin cancellation's status.
func (c *Client) EndpointManagerCancelStatus(ctx context.Context, adminCancelID string) (*AdminCancelStatus, error) {
	if adminCancelID == "" {
		return nil, fmt.Errorf("admin cancel ID is required")
	}
	var out AdminCancelStatus
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/admin_cancel/"+adminCancelID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerPauseTasks pauses tasks as an activity manager.
func (c *Client) EndpointManagerPauseTasks(ctx context.Context, taskIDs []string, message string) (*OperationResult, error) {
	body := map[string]interface{}{
		"message":      message,
		"task_id_list": taskIDs,
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint_manager/admin_pause", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerResumeTasks resumes tasks as an activity manager (no message).
func (c *Client) EndpointManagerResumeTasks(ctx context.Context, taskIDs []string) (*OperationResult, error) {
	body := map[string]interface{}{
		"task_id_list": taskIDs,
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint_manager/admin_resume", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerPauseRuleList lists pause rules, optionally scoped to an endpoint.
func (c *Client) EndpointManagerPauseRuleList(ctx context.Context, filterEndpoint string) (*PauseRuleList, error) {
	query := url.Values{}
	if filterEndpoint != "" {
		query.Set("filter_endpoint", filterEndpoint)
	}
	var out PauseRuleList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/pause_rule_list", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerCreatePauseRule creates a pause rule.
func (c *Client) EndpointManagerCreatePauseRule(ctx context.Context, data map[string]interface{}) (*PauseRule, error) {
	if data == nil {
		return nil, fmt.Errorf("pause rule data is required")
	}
	if _, ok := data["DATA_TYPE"]; !ok {
		data["DATA_TYPE"] = "pause_rule"
	}
	var out PauseRule
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "endpoint_manager/pause_rule", nil, data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerGetPauseRule retrieves a single pause rule.
func (c *Client) EndpointManagerGetPauseRule(ctx context.Context, pauseRuleID string) (*PauseRule, error) {
	if pauseRuleID == "" {
		return nil, fmt.Errorf("pause rule ID is required")
	}
	var out PauseRule
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_manager/pause_rule/"+pauseRuleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerUpdatePauseRule updates a pause rule.
func (c *Client) EndpointManagerUpdatePauseRule(ctx context.Context, pauseRuleID string, data map[string]interface{}) (*PauseRule, error) {
	if pauseRuleID == "" {
		return nil, fmt.Errorf("pause rule ID is required")
	}
	var out PauseRule
	if err := c.doRequestLowLevel(ctx, http.MethodPut, "endpoint_manager/pause_rule/"+pauseRuleID, nil, data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointManagerDeletePauseRule deletes a pause rule.
func (c *Client) EndpointManagerDeletePauseRule(ctx context.Context, pauseRuleID string) (*OperationResult, error) {
	if pauseRuleID == "" {
		return nil, fmt.Errorf("pause rule ID is required")
	}
	var out OperationResult
	if err := c.doRequestLowLevel(ctx, http.MethodDelete, "endpoint_manager/pause_rule/"+pauseRuleID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
