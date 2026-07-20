// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"
)

// Constants for Globus Groups
const (
	DefaultBaseURL = "https://groups.api.globus.org/v2/"
	GroupsScope    = "urn:globus:auth:scope:groups.api.globus.org:all"
)

// Client provides methods for interacting with Globus Groups
type Client struct {
	Client *core.Client
}

// Option is a function that configures a clientConfig
type Option func(*clientConfig)

// clientConfig holds the configuration for the client
type clientConfig struct {
	authorizer  auth.Authorizer
	debug       bool
	trace       bool
	logger      core.Logger
	coreOptions []core.ClientOption
}

// WithAuthorizer sets the authorizer for the client
func WithAuthorizer(authorizer auth.Authorizer) Option {
	return func(c *clientConfig) {
		c.authorizer = authorizer
	}
}

// WithHTTPDebugging enables HTTP debugging
func WithHTTPDebugging(enable bool) Option {
	return func(c *clientConfig) {
		c.debug = enable
	}
}

// WithHTTPTracing enables HTTP tracing
func WithHTTPTracing(enable bool) Option {
	return func(c *clientConfig) {
		c.trace = enable
	}
}

// WithLogger sets the logger for the client
func WithLogger(logger core.Logger) Option {
	return func(c *clientConfig) {
		c.logger = logger
	}
}

// WithCoreOptions adds core client options
func WithCoreOptions(options ...core.ClientOption) Option {
	return func(c *clientConfig) {
		c.coreOptions = append(c.coreOptions, options...)
	}
}

// NewClient creates a new Groups client
func NewClient(options ...Option) (*Client, error) {
	// Apply the options to create the client configuration
	cfg := &clientConfig{}
	for _, option := range options {
		option(cfg)
	}

	// Validate configuration
	if cfg.authorizer == nil {
		return nil, fmt.Errorf("authorizer is required")
	}

	// Apply default options specific to Groups
	defaultOptions := []core.ClientOption{
		core.WithBaseURL(DefaultBaseURL),
		core.WithAuthorizer(cfg.authorizer),
	}

	// Apply debug options if enabled
	if cfg.debug {
		defaultOptions = append(defaultOptions, core.WithHTTPDebugging(true))
	}
	if cfg.trace {
		defaultOptions = append(defaultOptions, core.WithHTTPTracing(true))
	}
	if cfg.logger != nil {
		defaultOptions = append(defaultOptions, core.WithLogger(cfg.logger))
	}

	// Apply any additional core options
	if cfg.coreOptions != nil {
		defaultOptions = append(defaultOptions, cfg.coreOptions...)
	}

	// Create the base client
	baseClient := core.NewClient(defaultOptions...)

	return &Client{
		Client: baseClient,
	}, nil
}

// buildURLLowLevel builds a URL for the groups API
// This is an internal method used by the client.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
	baseURL := c.Client.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	url := baseURL + path
	if len(query) > 0 {
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
		// Check if it's a core.Error (from NewAPIError)
		if coreErr, ok := err.(*core.Error); ok {
			// Convert core error to groups error using the raw body
			return parseGroupsError(coreErr.StatusCode, coreErr.RawBody)
		}
		// Other errors (network, etc.)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body for successful responses
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// For non-GET requests with no response body expected, return success
	if method != http.MethodGet && response == nil {
		return nil
	}

	// If no response body expected or empty, return success
	if len(respBody) == 0 {
		return nil
	}

	// Decode successful response
	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// GetMyGroups retrieves the groups the current user belongs to
// (GET /groups/my_groups). statuses is comma-joined into a single query param
// (valid: active, invited, pending, rejected, removed, left, declined). The
// response is a top-level JSON array.
// v3.65.0.
func (c *Client) GetMyGroups(ctx context.Context, statuses []string) (*GroupList, error) {
	query := url.Values{}
	if len(statuses) > 0 {
		query.Set("statuses", strings.Join(statuses, ","))
	}

	var groups []Group
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/my_groups", query, nil, &groups); err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].DATA_TYPE == "" {
			groups[i].DATA_TYPE = "group"
		}
	}
	return &GroupList{Groups: groups}, nil
}

// GetGroup retrieves a specific group by ID (GET /groups/{id}). Pass opts.Include
// (e.g. "memberships", "policies") to expand the document; opts may be nil.
func (c *Client) GetGroup(ctx context.Context, groupID string, opts *GetGroupOptions) (*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	query := url.Values{}
	if opts != nil && len(opts.Include) > 0 {
		query.Set("include", strings.Join(opts.Include, ","))
	}

	var group Group
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID, query, nil, &group)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if group.DATA_TYPE == "" {
		group.DATA_TYPE = "group"
	}

	return &group, nil
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, group *GroupCreate) (*Group, error) {
	if group == nil {
		return nil, fmt.Errorf("group data is required")
	}

	if group.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	// Set the DATA_TYPE field if not already set
	if group.DATA_TYPE == "" {
		group.DATA_TYPE = "group_create"
	}

	var createdGroup Group
	err := c.doRequestLowLevel(ctx, http.MethodPost, "groups", nil, group, &createdGroup)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if createdGroup.DATA_TYPE == "" {
		createdGroup.DATA_TYPE = "group"
	}

	return &createdGroup, nil
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(ctx context.Context, groupID string, update *GroupUpdate) (*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	// Set the DATA_TYPE field if not already set
	if update.DATA_TYPE == "" {
		update.DATA_TYPE = "group_update"
	}

	var updatedGroup Group
	err := c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID, nil, update, &updatedGroup)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if updatedGroup.DATA_TYPE == "" {
		updatedGroup.DATA_TYPE = "group"
	}

	return &updatedGroup, nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	return c.doRequestLowLevel(ctx, http.MethodDelete, "groups/"+groupID, nil, nil, nil)
}

// BatchMembershipAction applies one or more membership actions in a single call
// (POST /groups/{id}). This is the sole membership-mutation endpoint: add,
// invite, accept, approve, change_role, decline, join, leave, reject, remove, and
// request_join are all expressed here.
func (c *Client) BatchMembershipAction(ctx context.Context, groupID string, actions *BatchMembershipActions) (*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}
	if actions == nil {
		return nil, fmt.Errorf("actions are required")
	}

	var group Group
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "groups/"+groupID, nil, actions, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupBySubscriptionID retrieves the group associated with a subscription
// (GET /subscription_info/{subscription_id}).
func (c *Client) GetGroupBySubscriptionID(ctx context.Context, subscriptionID string) (*Group, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	var group Group
	err := c.doRequestLowLevel(ctx, http.MethodGet, "subscription_info/"+subscriptionID, nil, nil, &group)
	if err != nil {
		return nil, err
	}
	if group.DATA_TYPE == "" {
		group.DATA_TYPE = "group"
	}
	return &group, nil
}

// GetGroupPolicies retrieves the policy settings for a group
// (GET /groups/{id}/policies).
func (c *Client) GetGroupPolicies(ctx context.Context, groupID string) (*GroupPolicies, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var policies GroupPolicies
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/policies", nil, nil, &policies)
	if err != nil {
		return nil, err
	}
	return &policies, nil
}

// SetGroupPolicies replaces the policy settings for a group
// (PUT /groups/{id}/policies).
func (c *Client) SetGroupPolicies(ctx context.Context, groupID string, policies *GroupPolicies) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if policies == nil {
		return fmt.Errorf("policies are required")
	}
	return c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID+"/policies", nil, policies, nil)
}

// GetIdentityPreferences retrieves the caller's Groups preferences
// (GET /preferences). Preferences are account-level, not group-scoped.
func (c *Client) GetIdentityPreferences(ctx context.Context) (map[string]interface{}, error) {
	var prefs map[string]interface{}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "preferences", nil, nil, &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// SetIdentityPreferences updates the caller's Groups preferences
// (PUT /preferences), e.g. {"allow_add": false}.
func (c *Client) SetIdentityPreferences(ctx context.Context, prefs map[string]interface{}) error {
	if prefs == nil {
		return fmt.Errorf("preferences are required")
	}
	return c.doRequestLowLevel(ctx, http.MethodPut, "preferences", nil, prefs, nil)
}

// GetMembershipFields retrieves the caller's membership field values for a group
// (GET /groups/{id}/membership_fields).
func (c *Client) GetMembershipFields(ctx context.Context, groupID string) (map[string]interface{}, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var fields map[string]interface{}
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/membership_fields", nil, nil, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// SetMembershipFields sets the caller's membership field values for a group
// (PUT /groups/{id}/membership_fields).
func (c *Client) SetMembershipFields(ctx context.Context, groupID string, fields map[string]interface{}) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if fields == nil {
		return fmt.Errorf("membership fields are required")
	}
	return c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID+"/membership_fields", nil, fields, nil)
}

// parseGroupsError parses JSON error responses from the Groups API
func parseGroupsError(statusCode int, responseBody []byte) error {
	// Try to parse JSON error response with common formats
	var errorResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}

	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &errorResp); err == nil {
			if errorResp.Error != "" && errorResp.Code != "" {
				// Create a proper GlobusError with full error details
				return errors.NewGlobusErrorWithStatus("groups", errorResp.Code, errorResp.Error, statusCode)
			} else if errorResp.Error != "" {
				// Error message without code
				return errors.NewGlobusErrorWithStatus("groups", "GROUPS_ERROR", errorResp.Error, statusCode)
			}
		}
	}

	// Fallback to generic error message with status
	return errors.NewGlobusErrorWithStatus("groups", fmt.Sprintf("HTTP_%d", statusCode),
		fmt.Sprintf("Request failed with status code %d", statusCode), statusCode)
}
