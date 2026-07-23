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

// Client is the v4 Transfer service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Transfer client
// In v4, config is required and must include explicit scopes
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	// Set default Transfer service URL if not provided. The base is the bare host;
	// classic routes carry a /v0.10 prefix and the Beta tunnel/stream routes carry
	// a /v2 prefix, so both surfaces are reachable through the path-joining buildURL.
	if config.BaseURL == "" {
		config.BaseURL = "https://transfer.api.globus.org"
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

// GetEndpoint retrieves information about a specific endpoint
// v4: Context is always first parameter
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}

	var endpoint Endpoint
	path := fmt.Sprintf("/v0.10/endpoint/%s", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

// CreateEndpoint creates an endpoint (POST /v0.10/endpoint). doc is a
// passthrough endpoint document. Setting "is_globus_connect": true registers a
// Globus Connect Personal (mapped) endpoint; the response then carries a
// "globus_connect_setup_key" used to configure an installed GCP agent.
func (c *Client) CreateEndpoint(ctx context.Context, doc map[string]interface{}) (GenericResponse, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "endpoint document is required"}
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v0.10/endpoint", nil, doc, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateEndpoint updates an endpoint (PUT /v0.10/endpoint/{id}). doc is a
// passthrough endpoint document.
func (c *Client) UpdateEndpoint(ctx context.Context, endpointID string, doc map[string]interface{}) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "endpoint document is required"}
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/v0.10/endpoint/%s", endpointID), nil, doc, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteEndpoint deletes an endpoint (DELETE /v0.10/endpoint/{id}).
func (c *Client) DeleteEndpoint(ctx context.Context, endpointID string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/v0.10/endpoint/%s", endpointID), nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SetSubscriptionID sets a collection's subscription
// (PUT /v0.10/endpoint/{id}/subscription).
func (c *Client) SetSubscriptionID(ctx context.Context, collectionID, subscriptionID string) (GenericResponse, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}
	body := map[string]interface{}{"subscription_id": subscriptionID}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/v0.10/endpoint/%s/subscription", collectionID), nil, body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SetSubscriptionAdminVerified sets a collection's admin-verified flag
// (PUT /v0.10/endpoint/{id}/subscription_admin_verified).
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, collectionID string, verified bool) (GenericResponse, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}
	body := map[string]interface{}{"subscription_admin_verified": verified}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/v0.10/endpoint/%s/subscription_admin_verified", collectionID), nil, body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// EndpointSearch searches endpoints (GET /v0.10/endpoint_search).
func (c *Client) EndpointSearch(ctx context.Context, options *EndpointSearchOptions) (*EndpointSearchResult, error) {
	query := url.Values{}
	if options != nil {
		if options.FilterFulltext != "" {
			query.Set("filter_fulltext", options.FilterFulltext)
		}
		if options.FilterScope != "" {
			query.Set("filter_scope", options.FilterScope)
		}
		if options.FilterOwnerID != "" {
			query.Set("filter_owner_id", options.FilterOwnerID)
		}
		if options.FilterHostEndpoint != "" {
			query.Set("filter_host_endpoint", options.FilterHostEndpoint)
		}
		if options.FilterNonFunctional != nil {
			query.Set("filter_non_functional", boolToWire(*options.FilterNonFunctional))
		}
		if options.FilterEntityType != "" {
			query.Set("filter_entity_type", options.FilterEntityType)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var result EndpointSearchResult
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v0.10/endpoint_search", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SubmitTransfer submits a transfer task
// v4: Context is always first parameter
func (c *Client) SubmitTransfer(ctx context.Context, transfer *Transfer) (*TaskSubmitResponse, error) {
	if transfer == nil {
		return nil, &core.ValidationError{
			Field:   "transfer",
			Message: "transfer data is required",
		}
	}

	// Validate required fields
	if transfer.SourceEndpoint == "" {
		return nil, &core.ValidationError{
			Field:   "SourceEndpoint",
			Message: "source endpoint is required",
		}
	}
	if transfer.DestinationEndpoint == "" {
		return nil, &core.ValidationError{
			Field:   "DestinationEndpoint",
			Message: "destination endpoint is required",
		}
	}
	if len(transfer.Items) == 0 {
		return nil, &core.ValidationError{
			Field:   "Items",
			Message: "at least one transfer item is required",
		}
	}

	// Set DATA_TYPE if not set
	if transfer.DATA_TYPE == "" {
		transfer.DATA_TYPE = "transfer"
	}
	// Auto-fetch a submission_id when the caller did not supply one, matching
	// upstream submit_transfer.
	if transfer.SubmissionID == "" {
		id, err := c.GetSubmissionID(ctx)
		if err != nil {
			return nil, err
		}
		transfer.SubmissionID = id
	}

	var response TaskSubmitResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v0.10/transfer", nil, transfer, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SubmitDelete submits a delete task
// v4: Context is always first parameter
func (c *Client) SubmitDelete(ctx context.Context, delete *Delete) (*TaskSubmitResponse, error) {
	if delete == nil {
		return nil, &core.ValidationError{
			Field:   "delete",
			Message: "delete data is required",
		}
	}

	// Validate required fields
	if delete.Endpoint == "" {
		return nil, &core.ValidationError{
			Field:   "Endpoint",
			Message: "endpoint is required",
		}
	}
	if len(delete.Items) == 0 {
		return nil, &core.ValidationError{
			Field:   "Items",
			Message: "at least one delete item is required",
		}
	}

	// Set DATA_TYPE if not set
	if delete.DATA_TYPE == "" {
		delete.DATA_TYPE = "delete"
	}
	if delete.SubmissionID == "" {
		id, err := c.GetSubmissionID(ctx)
		if err != nil {
			return nil, err
		}
		delete.SubmissionID = id
	}

	var response TaskSubmitResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v0.10/delete", nil, delete, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTask retrieves information about a specific task
// v4: Context is always first parameter
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &core.ValidationError{
			Field:   "taskID",
			Message: "task ID is required",
		}
	}

	var task Task
	path := fmt.Sprintf("/v0.10/task/%s", taskID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask updates a task's label/deadline (PUT /v0.10/task/{id}).
func (c *Client) UpdateTask(ctx context.Context, taskID string, doc map[string]interface{}) (GenericResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "update document is required"}
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/v0.10/task/%s", taskID), nil, doc, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CancelTask cancels a running task
// v4: Context is always first parameter
func (c *Client) CancelTask(ctx context.Context, taskID string) (*TaskCancelResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{
			Field:   "taskID",
			Message: "task ID is required",
		}
	}

	var response TaskCancelResponse
	path := fmt.Sprintf("/v0.10/task/%s/cancel", taskID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ListTasks lists tasks with optional filtering
// v4: Context is always first parameter
func (c *Client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
	query := url.Values{}

	if options != nil {
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if len(options.FilterStatus) > 0 {
			query.Set("filter_status", strings.Join(options.FilterStatus, ","))
		}
		if len(options.OrderBy) > 0 {
			query.Set("orderby", strings.Join(options.OrderBy, ","))
		}
	}

	var taskList TaskList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v0.10/task_list", query, nil, &taskList)
	if err != nil {
		return nil, err
	}

	return &taskList, nil
}

// TaskEventList lists a task's events (GET /v0.10/task/{id}/event_list).
func (c *Client) TaskEventList(ctx context.Context, taskID string, options *ListTaskEventsOptions) (*TaskEventList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	query := taskEventQuery(options)
	var list TaskEventList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/task/%s/event_list", taskID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// TaskPauseInfo returns why a task is paused (GET /v0.10/task/{id}/pause_info).
func (c *Client) TaskPauseInfo(ctx context.Context, taskID string) (GenericResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/task/%s/pause_info", taskID), nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// TaskSuccessfulTransfers lists a task's successful transfers
// (GET /v0.10/task/{id}/successful_transfers, nullable-marker paginated).
func (c *Client) TaskSuccessfulTransfers(ctx context.Context, taskID, marker string) (*NullableMarkerList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	query := url.Values{}
	if marker != "" {
		query.Set("marker", marker)
	}
	var list NullableMarkerList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/task/%s/successful_transfers", taskID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// TaskSkippedErrors lists a task's skipped errors
// (GET /v0.10/task/{id}/skipped_errors, nullable-marker paginated).
func (c *Client) TaskSkippedErrors(ctx context.Context, taskID, marker string) (*NullableMarkerList, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	query := url.Values{}
	if marker != "" {
		query.Set("marker", marker)
	}
	var list NullableMarkerList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/task/%s/skipped_errors", taskID), query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// ListDirectory lists directory contents on an endpoint
// v4: Context is always first parameter
func (c *Client) ListDirectory(ctx context.Context, endpointID, path string, options *ListDirectoryOptions) (*DirectoryListing, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}

	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}

	if options != nil {
		if options.ShowHidden {
			query.Set("show_hidden", "1")
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if len(options.OrderBy) > 0 {
			query.Set("orderby", strings.Join(options.OrderBy, ","))
		}
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.LocalUser != "" {
			query.Set("local_user", options.LocalUser)
		}
	}

	var listing DirectoryListing
	apiPath := fmt.Sprintf("/v0.10/operation/endpoint/%s/ls", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, apiPath, query, nil, &listing)
	if err != nil {
		return nil, err
	}

	return &listing, nil
}

// OperationStat stats a single path on an endpoint
// (GET /v0.10/operation/endpoint/{id}/stat).
func (c *Client) OperationStat(ctx context.Context, endpointID, path, localUser string) (GenericResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if path == "" {
		return nil, &core.ValidationError{Field: "path", Message: "path is required"}
	}
	query := url.Values{}
	query.Set("path", path)
	if localUser != "" {
		query.Set("local_user", localUser)
	}
	var resp GenericResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v0.10/operation/endpoint/%s/stat", endpointID), query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// MakeDirectory creates a directory on an endpoint
// (POST /v0.10/operation/endpoint/{id}/mkdir). localUser is optional (pass "").
func (c *Client) MakeDirectory(ctx context.Context, endpointID, path, localUser string) (*OperationResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if path == "" {
		return nil, &core.ValidationError{Field: "path", Message: "path is required"}
	}

	body := map[string]interface{}{
		"DATA_TYPE": "mkdir",
		"path":      path,
	}
	if localUser != "" {
		body["local_user"] = localUser
	}

	var response OperationResponse
	apiPath := fmt.Sprintf("/v0.10/operation/endpoint/%s/mkdir", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, apiPath, nil, body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Rename renames a file or directory on an endpoint
// (POST /v0.10/operation/endpoint/{id}/rename). localUser is optional (pass "").
func (c *Client) Rename(ctx context.Context, endpointID, oldPath, newPath, localUser string) (*OperationResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if oldPath == "" {
		return nil, &core.ValidationError{Field: "oldPath", Message: "old path is required"}
	}
	if newPath == "" {
		return nil, &core.ValidationError{Field: "newPath", Message: "new path is required"}
	}

	body := map[string]interface{}{
		"DATA_TYPE": "rename",
		"old_path":  oldPath,
		"new_path":  newPath,
	}
	if localUser != "" {
		body["local_user"] = localUser
	}

	var response OperationResponse
	apiPath := fmt.Sprintf("/v0.10/operation/endpoint/%s/rename", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, apiPath, nil, body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateTunnel creates a new Globus Streams tunnel (POST /v2/tunnels, JSON:API).
// BETA.
func (c *Client) CreateTunnel(ctx context.Context, data *TunnelCreate) (*Tunnel, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "tunnel data is required"}
	}
	attrs := map[string]interface{}{}
	setIf(attrs, "label", data.Label)
	setIf(attrs, "listener_ip_address", data.ListenerIPAddress)
	setIf(attrs, "submission_id", data.SubmissionID)
	if data.ListenerPort != nil {
		attrs["listener_port"] = *data.ListenerPort
	}
	if data.LifetimeMins != nil {
		attrs["lifetime_mins"] = *data.LifetimeMins
	}
	if data.Restartable != nil {
		attrs["restartable"] = *data.Restartable
	}
	doc := tunnelJSONAPI{Data: tunnelJSONAPIData{
		Type: "Tunnel",
		Relationships: map[string]jsonAPIRel{
			"listener":  {Data: jsonAPIRelData{Type: "StreamAccessPoint", ID: data.ListenerStreamAccessPoint}},
			"initiator": {Data: jsonAPIRelData{Type: "StreamAccessPoint", ID: data.InitiatorStreamAccessPoint}},
		},
		Attributes: attrs,
	}}
	return c.tunnelRequest(ctx, http.MethodPost, "/v2/tunnels", doc)
}

// GetTunnel retrieves a tunnel by ID (GET /v2/tunnels/{id}). BETA.
func (c *Client) GetTunnel(ctx context.Context, tunnelID string) (*Tunnel, error) {
	if tunnelID == "" {
		return nil, &core.ValidationError{Field: "tunnelID", Message: "tunnel ID is required"}
	}
	return c.tunnelRequest(ctx, http.MethodGet, fmt.Sprintf("/v2/tunnels/%s", tunnelID), nil)
}

// UpdateTunnel updates a tunnel (PATCH /v2/tunnels/{id}, JSON:API). BETA.
func (c *Client) UpdateTunnel(ctx context.Context, tunnelID string, data *TunnelUpdate) (*Tunnel, error) {
	if tunnelID == "" {
		return nil, &core.ValidationError{Field: "tunnelID", Message: "tunnel ID is required"}
	}
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "tunnel update data is required"}
	}
	attrs := map[string]interface{}{}
	setIf(attrs, "label", data.Label)
	setIf(attrs, "listener_ip_address", data.ListenerIPAddress)
	if data.ListenerPort != nil {
		attrs["listener_port"] = *data.ListenerPort
	}
	doc := tunnelJSONAPI{Data: tunnelJSONAPIData{Type: "Tunnel", Attributes: attrs}}
	return c.tunnelRequest(ctx, http.MethodPatch, fmt.Sprintf("/v2/tunnels/%s", tunnelID), doc)
}

// DeleteTunnel deletes a tunnel by ID (DELETE /v2/tunnels/{id}). BETA.
func (c *Client) DeleteTunnel(ctx context.Context, tunnelID string) error {
	if tunnelID == "" {
		return &core.ValidationError{Field: "tunnelID", Message: "tunnel ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/v2/tunnels/%s", tunnelID), nil, nil, nil)
}

// ListTunnels lists tunnels owned by the caller (GET /v2/tunnels, JSON:API).
// Not paginated upstream. BETA.
func (c *Client) ListTunnels(ctx context.Context, options *ListTunnelsOptions) (*TunnelList, error) {
	var doc jsonAPICollection
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/tunnels", nil, nil, &doc); err != nil {
		return nil, err
	}
	list := &TunnelList{Tunnels: make([]Tunnel, 0, len(doc.Data))}
	for _, res := range doc.Data {
		list.Tunnels = append(list.Tunnels, flattenTunnel(res))
	}
	return list, nil
}

// GetStreamAccessPoint retrieves a Stream Access Point by ID
// (GET /v2/stream_access_points/{id}, JSON:API). BETA.
func (c *Client) GetStreamAccessPoint(ctx context.Context, accessPointID string) (*StreamAccessPoint, error) {
	if accessPointID == "" {
		return nil, &core.ValidationError{Field: "accessPointID", Message: "access point ID is required"}
	}
	var doc jsonAPIResource
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v2/stream_access_points/%s", accessPointID), nil, nil, &doc); err != nil {
		return nil, err
	}
	ap := flattenSAP(doc.Data)
	return &ap, nil
}

// ListStreamAccessPoints lists stream access points
// (GET /v2/stream_access_points, JSON:API). Not paginated by marker upstream. BETA.
func (c *Client) ListStreamAccessPoints(ctx context.Context, options *ListTunnelsOptions) (*StreamAccessPointList, error) {
	var doc jsonAPICollection
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/stream_access_points", nil, nil, &doc); err != nil {
		return nil, err
	}
	list := &StreamAccessPointList{Data: make([]StreamAccessPoint, 0, len(doc.Data))}
	for _, res := range doc.Data {
		list.Data = append(list.Data, flattenSAP(res))
	}
	return list, nil
}

// GetTunnelEvents fetches a tunnel's events
// (GET /v2/tunnels/{id}/events, JSON:API). BETA.
func (c *Client) GetTunnelEvents(ctx context.Context, tunnelID string, options *ListTunnelEventsOptions) (*TunnelEventList, error) {
	if tunnelID == "" {
		return nil, &core.ValidationError{Field: "tunnelID", Message: "tunnel ID is required"}
	}
	var doc jsonAPICollection
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/v2/tunnels/%s/events", tunnelID), nil, nil, &doc); err != nil {
		return nil, err
	}
	list := &TunnelEventList{Events: make([]TunnelEvent, 0, len(doc.Data))}
	for _, res := range doc.Data {
		ev := TunnelEvent{ID: res.ID, TunnelID: tunnelID}
		if v, ok := res.Attributes["code"].(string); ok {
			ev.Code = v
		}
		if v, ok := res.Attributes["description"].(string); ok {
			ev.Description = v
		}
		ev.Details = res.Attributes
		list.Events = append(list.Events, ev)
	}
	return list, nil
}

// tunnelRequest performs a JSON:API tunnel request and flattens the single
// resource in the response into a *Tunnel.
func (c *Client) tunnelRequest(ctx context.Context, method, path string, body interface{}) (*Tunnel, error) {
	var doc jsonAPIResource
	if err := c.baseClient.DoRequest(ctx, method, path, nil, body, &doc); err != nil {
		return nil, err
	}
	t := flattenTunnel(doc.Data)
	return &t, nil
}

// GetSubmissionID retrieves a fresh submission ID from the Transfer service.
func (c *Client) GetSubmissionID(ctx context.Context) (string, error) {
	var resp struct {
		DataType string `json:"DATA_TYPE"`
		Value    string `json:"value"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v0.10/submission_id", nil, nil, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// boolToWire renders a bool as the "1"/"0" the Transfer API expects for its
// integer-boolean query params (show_hidden, filter_non_functional, ...).
func boolToWire(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// taskEventQuery builds the query for task event listing.
func taskEventQuery(o *ListTaskEventsOptions) url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}
	if o.FilterIsError != nil {
		q.Set("filter_is_error", boolToWire(*o.FilterIsError))
	}
	return q
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}
