// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// This file implements the classic Transfer admin surface (endpoint roles, ACLs,
// servers, shared endpoints, pause rules, and the endpoint_manager family). These
// routes return open-ended service documents, so requests and responses use
// passthrough maps.

// --- endpoint helpers ---

func (c *Client) getGeneric(ctx context.Context, path string, query url.Values) (GenericResponse, error) {
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) sendGeneric(ctx context.Context, method, path string, body interface{}) (GenericResponse, error) {
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, method, path, nil, body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// MyEffectivePauseRuleList lists the caller's effective pause rules for an
// endpoint (GET /v0.10/endpoint/{id}/my_effective_pause_rule_list).
func (c *Client) MyEffectivePauseRuleList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/my_effective_pause_rule_list", endpointID), nil)
}

// --- shared endpoints ---

// MySharedEndpointList lists shared endpoints owned by the caller on an endpoint
// (GET /v0.10/endpoint/{id}/my_shared_endpoint_list).
func (c *Client) MySharedEndpointList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/my_shared_endpoint_list", endpointID), nil)
}

// GetSharedEndpointList lists shared endpoints hosted on an endpoint
// (GET /v0.10/endpoint/{id}/shared_endpoint_list, next_token paginated).
func (c *Client) GetSharedEndpointList(ctx context.Context, endpointID string, maxResults int, nextToken string) (*SharedEndpointList, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	query := url.Values{}
	if maxResults > 0 {
		query.Set("max_results", strconv.Itoa(maxResults))
	}
	if nextToken != "" {
		query.Set("next_token", nextToken)
	}
	var list SharedEndpointList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/endpoint/%s/shared_endpoint_list", endpointID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// CreateSharedEndpoint creates a shared (guest) endpoint
// (POST /v0.10/shared_endpoint). doc is a passthrough shared_endpoint document.
func (c *Client) CreateSharedEndpoint(ctx context.Context, doc map[string]interface{}) (GenericResponse, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "shared endpoint document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPost, "/v0.10/shared_endpoint", doc)
}

// --- endpoint servers ---

// EndpointServerList lists an endpoint's servers
// (GET /v0.10/endpoint/{id}/server_list).
func (c *Client) EndpointServerList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/server_list", endpointID), nil)
}

// GetEndpointServer retrieves a single endpoint server
// (GET /v0.10/endpoint/{id}/server/{server_id}).
func (c *Client) GetEndpointServer(ctx context.Context, endpointID, serverID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if serverID == "" {
		return nil, &core.ValidationError{Field: "serverID", Message: "server ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/server/%s", endpointID, serverID), nil)
}

// --- endpoint roles ---

// EndpointRoleList lists an endpoint's roles (GET /v0.10/endpoint/{id}/role_list).
func (c *Client) EndpointRoleList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/role_list", endpointID), nil)
}

// AddEndpointRole adds a role to an endpoint (POST /v0.10/endpoint/{id}/role).
func (c *Client) AddEndpointRole(ctx context.Context, endpointID string, doc map[string]interface{}) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "role document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPost, fmt.Sprintf("/v0.10/endpoint/%s/role", endpointID), doc)
}

// GetEndpointRole retrieves an endpoint role (GET /v0.10/endpoint/{id}/role/{role_id}).
func (c *Client) GetEndpointRole(ctx context.Context, endpointID, roleID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/role/%s", endpointID, roleID), nil)
}

// DeleteEndpointRole deletes an endpoint role
// (DELETE /v0.10/endpoint/{id}/role/{role_id}).
func (c *Client) DeleteEndpointRole(ctx context.Context, endpointID, roleID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	return c.sendGeneric(ctx, http.MethodDelete, fmt.Sprintf("/v0.10/endpoint/%s/role/%s", endpointID, roleID), nil)
}

// --- endpoint ACLs ---

// EndpointACLList lists an endpoint's access rules
// (GET /v0.10/endpoint/{id}/access_list).
func (c *Client) EndpointACLList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/access_list", endpointID), nil)
}

// GetEndpointACLRule retrieves an access rule
// (GET /v0.10/endpoint/{id}/access/{rule_id}).
func (c *Client) GetEndpointACLRule(ctx context.Context, endpointID, ruleID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if ruleID == "" {
		return nil, &core.ValidationError{Field: "ruleID", Message: "rule ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint/%s/access/%s", endpointID, ruleID), nil)
}

// AddEndpointACLRule adds an access rule (POST /v0.10/endpoint/{id}/access).
func (c *Client) AddEndpointACLRule(ctx context.Context, endpointID string, doc map[string]interface{}) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "access rule document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPost, fmt.Sprintf("/v0.10/endpoint/%s/access", endpointID), doc)
}

// UpdateEndpointACLRule updates an access rule
// (PUT /v0.10/endpoint/{id}/access/{rule_id}).
func (c *Client) UpdateEndpointACLRule(ctx context.Context, endpointID, ruleID string, doc map[string]interface{}) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if ruleID == "" {
		return nil, &core.ValidationError{Field: "ruleID", Message: "rule ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "access rule document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPut, fmt.Sprintf("/v0.10/endpoint/%s/access/%s", endpointID, ruleID), doc)
}

// DeleteEndpointACLRule deletes an access rule
// (DELETE /v0.10/endpoint/{id}/access/{rule_id}).
func (c *Client) DeleteEndpointACLRule(ctx context.Context, endpointID, ruleID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if ruleID == "" {
		return nil, &core.ValidationError{Field: "ruleID", Message: "rule ID is required"}
	}
	return c.sendGeneric(ctx, http.MethodDelete, fmt.Sprintf("/v0.10/endpoint/%s/access/%s", endpointID, ruleID), nil)
}

// --- endpoint_manager: monitored endpoints & tasks ---

// EndpointManagerMonitoredEndpoints lists endpoints the caller monitors as an
// activity manager (GET /v0.10/endpoint_manager/monitored_endpoints).
func (c *Client) EndpointManagerMonitoredEndpoints(ctx context.Context) (GenericResponse, error) {
	return c.getGeneric(ctx, "/v0.10/endpoint_manager/monitored_endpoints", nil)
}

// EndpointManagerHostedEndpointList lists guest collections hosted on an endpoint
// (GET /v0.10/endpoint_manager/endpoint/{id}/hosted_endpoint_list).
func (c *Client) EndpointManagerHostedEndpointList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/endpoint/%s/hosted_endpoint_list", endpointID), nil)
}

// EndpointManagerGetEndpoint retrieves an endpoint as an activity manager
// (GET /v0.10/endpoint_manager/endpoint/{id}).
func (c *Client) EndpointManagerGetEndpoint(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/endpoint/%s", endpointID), nil)
}

// EndpointManagerACLList lists an endpoint's ACL rules as an activity manager
// (GET /v0.10/endpoint_manager/endpoint/{id}/access_list).
func (c *Client) EndpointManagerACLList(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/endpoint/%s/access_list", endpointID), nil)
}

// EndpointManagerTaskList lists tasks visible to the activity manager
// (GET /v0.10/endpoint_manager/task_list, last_key paginated).
func (c *Client) EndpointManagerTaskList(ctx context.Context, options *EndpointManagerTaskListOptions) (*EndpointManagerTaskList, error) {
	query := url.Values{}
	if options != nil {
		if len(options.FilterStatus) > 0 {
			query.Set("filter_status", strings.Join(options.FilterStatus, ","))
		}
		if len(options.FilterTaskID) > 0 {
			query.Set("filter_task_id", strings.Join(options.FilterTaskID, ","))
		}
		if options.FilterOwnerID != "" {
			query.Set("filter_owner_id", options.FilterOwnerID)
		}
		if options.FilterEndpoint != "" {
			query.Set("filter_endpoint", options.FilterEndpoint)
		}
		if options.FilterEndpointUse != "" {
			query.Set("filter_endpoint_use", options.FilterEndpointUse)
		}
		if options.LastKey != "" {
			query.Set("last_key", options.LastKey)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
	}
	var list EndpointManagerTaskList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v0.10/endpoint_manager/task_list", query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// EndpointManagerGetTask retrieves a task as an activity manager
// (GET /v0.10/endpoint_manager/task/{id}).
func (c *Client) EndpointManagerGetTask(ctx context.Context, taskID string) (GenericResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/task/%s", taskID), nil)
}

// EndpointManagerTaskEventList lists a task's events as an activity manager
// (GET /v0.10/endpoint_manager/task/{id}/event_list).
func (c *Client) EndpointManagerTaskEventList(ctx context.Context, taskID string, options *ListTaskEventsOptions) (*TaskEventList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	var list TaskEventList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/endpoint_manager/task/%s/event_list", taskID), taskEventQuery(options), nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// EndpointManagerTaskPauseInfo returns why a task is paused, as an activity
// manager (GET /v0.10/endpoint_manager/task/{id}/pause_info).
func (c *Client) EndpointManagerTaskPauseInfo(ctx context.Context, taskID string) (GenericResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/task/%s/pause_info", taskID), nil)
}

// EndpointManagerTaskSuccessfulTransfers lists a task's successful transfers as
// an activity manager (GET /v0.10/endpoint_manager/task/{id}/successful_transfers).
func (c *Client) EndpointManagerTaskSuccessfulTransfers(ctx context.Context, taskID, marker string) (*NullableMarkerList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	query := url.Values{}
	if marker != "" {
		query.Set("marker", marker)
	}
	var list NullableMarkerList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/endpoint_manager/task/%s/successful_transfers", taskID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// EndpointManagerTaskSkippedErrors lists a task's skipped errors as an activity
// manager (GET /v0.10/endpoint_manager/task/{id}/skipped_errors).
func (c *Client) EndpointManagerTaskSkippedErrors(ctx context.Context, taskID, marker string) (*NullableMarkerList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	query := url.Values{}
	if marker != "" {
		query.Set("marker", marker)
	}
	var list NullableMarkerList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/endpoint_manager/task/%s/skipped_errors", taskID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// --- endpoint_manager: task control ---

// EndpointManagerCancelTasks requests cancellation of tasks as an admin
// (POST /v0.10/endpoint_manager/admin_cancel).
func (c *Client) EndpointManagerCancelTasks(ctx context.Context, taskIDs []string, message string) (GenericResponse, error) {
	if len(taskIDs) == 0 {
		return nil, &core.ValidationError{Field: "taskIDs", Message: "at least one task ID is required"}
	}
	body := map[string]interface{}{"task_id_list": taskIDs}
	if message != "" {
		body["message"] = message
	}
	return c.sendGeneric(ctx, http.MethodPost, "/v0.10/endpoint_manager/admin_cancel", body)
}

// EndpointManagerCancelStatus checks the status of an admin cancel
// (GET /v0.10/endpoint_manager/admin_cancel/{id}).
func (c *Client) EndpointManagerCancelStatus(ctx context.Context, adminCancelID string) (GenericResponse, error) {
	if adminCancelID == "" {
		return nil, &core.ValidationError{Field: "adminCancelID", Message: "admin cancel ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/admin_cancel/%s", adminCancelID), nil)
}

// EndpointManagerPauseTasks pauses tasks as an admin
// (POST /v0.10/endpoint_manager/admin_pause).
func (c *Client) EndpointManagerPauseTasks(ctx context.Context, taskIDs []string, message string) (GenericResponse, error) {
	if len(taskIDs) == 0 {
		return nil, &core.ValidationError{Field: "taskIDs", Message: "at least one task ID is required"}
	}
	body := map[string]interface{}{"task_id_list": taskIDs}
	if message != "" {
		body["message"] = message
	}
	return c.sendGeneric(ctx, http.MethodPost, "/v0.10/endpoint_manager/admin_pause", body)
}

// EndpointManagerResumeTasks resumes tasks as an admin
// (POST /v0.10/endpoint_manager/admin_resume).
func (c *Client) EndpointManagerResumeTasks(ctx context.Context, taskIDs []string) (GenericResponse, error) {
	if len(taskIDs) == 0 {
		return nil, &core.ValidationError{Field: "taskIDs", Message: "at least one task ID is required"}
	}
	body := map[string]interface{}{"task_id_list": taskIDs}
	return c.sendGeneric(ctx, http.MethodPost, "/v0.10/endpoint_manager/admin_resume", body)
}

// --- endpoint_manager: pause rules ---

// EndpointManagerPauseRuleList lists pause rules
// (GET /v0.10/endpoint_manager/pause_rule_list).
func (c *Client) EndpointManagerPauseRuleList(ctx context.Context, filterEndpoint string) (GenericResponse, error) {
	query := url.Values{}
	if filterEndpoint != "" {
		query.Set("filter_endpoint", filterEndpoint)
	}
	return c.getGeneric(ctx, "/v0.10/endpoint_manager/pause_rule_list", query)
}

// EndpointManagerCreatePauseRule creates a pause rule
// (POST /v0.10/endpoint_manager/pause_rule).
func (c *Client) EndpointManagerCreatePauseRule(ctx context.Context, doc map[string]interface{}) (GenericResponse, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "pause rule document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPost, "/v0.10/endpoint_manager/pause_rule", doc)
}

// EndpointManagerGetPauseRule retrieves a pause rule
// (GET /v0.10/endpoint_manager/pause_rule/{id}).
func (c *Client) EndpointManagerGetPauseRule(ctx context.Context, pauseRuleID string) (GenericResponse, error) {
	if pauseRuleID == "" {
		return nil, &core.ValidationError{Field: "pauseRuleID", Message: "pause rule ID is required"}
	}
	return c.getGeneric(ctx, fmt.Sprintf("/v0.10/endpoint_manager/pause_rule/%s", pauseRuleID), nil)
}

// EndpointManagerUpdatePauseRule updates a pause rule
// (PUT /v0.10/endpoint_manager/pause_rule/{id}).
func (c *Client) EndpointManagerUpdatePauseRule(ctx context.Context, pauseRuleID string, doc map[string]interface{}) (GenericResponse, error) {
	if pauseRuleID == "" {
		return nil, &core.ValidationError{Field: "pauseRuleID", Message: "pause rule ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "pause rule document is required"}
	}
	return c.sendGeneric(ctx, http.MethodPut, fmt.Sprintf("/v0.10/endpoint_manager/pause_rule/%s", pauseRuleID), doc)
}

// EndpointManagerDeletePauseRule deletes a pause rule
// (DELETE /v0.10/endpoint_manager/pause_rule/{id}).
func (c *Client) EndpointManagerDeletePauseRule(ctx context.Context, pauseRuleID string) (GenericResponse, error) {
	if pauseRuleID == "" {
		return nil, &core.ValidationError{Field: "pauseRuleID", Message: "pause rule ID is required"}
	}
	return c.sendGeneric(ctx, http.MethodDelete, fmt.Sprintf("/v0.10/endpoint_manager/pause_rule/%s", pauseRuleID), nil)
}
