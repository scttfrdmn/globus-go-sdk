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
	"strconv"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/response"
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
		// Check if it's a core.Error (from NewAPIError)
		if coreErr, ok := err.(*core.Error); ok {
			// Convert core error to groups error using the raw body
			return parseGroupsError(coreErr.StatusCode, coreErr.RawBody)
		}
		// Other errors (network, etc.)
		return err
	}
	defer resp.Body.Close()

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

// ListGroups retrieves groups the current user is a member of
func (c *Client) ListGroups(ctx context.Context, options *ListGroupsOptions) (*GroupList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.IncludeGroupMembership {
			query.Set("include_group_membership", "true")
		}
		if options.IncludeIdentitySet {
			query.Set("include_identity_set", "true")
		}
		if options.ForUserID != "" {
			query.Set("for_user_id", options.ForUserID)
		}
		if options.MyGroups {
			query.Set("my_groups", "true")
		}
		if options.PageSize > 0 {
			query.Set("per_page", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var groupList GroupList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups", query, nil, &groupList)
	if err != nil {
		return nil, err
	}

	// Ensure all returned group objects have the DATA_TYPE set
	for i := range groupList.Groups {
		if groupList.Groups[i].DATA_TYPE == "" {
			groupList.Groups[i].DATA_TYPE = "group"
		}
	}

	return &groupList, nil
}

// ListGroupsV2 retrieves groups with unified response system
func (c *Client) ListGroupsV2(ctx context.Context, options *ListGroupsOptions) (*response.GroupsResponse[GroupList], error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.IncludeGroupMembership {
			query.Set("include_group_membership", "true")
		}
		if options.IncludeIdentitySet {
			query.Set("include_identity_set", "true")
		}
		if options.ForUserID != "" {
			query.Set("for_user_id", options.ForUserID)
		}
		if options.MyGroups {
			query.Set("my_groups", "true")
		}
		if options.PageSize > 0 {
			query.Set("per_page", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var groupList GroupList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups", query, nil, &groupList)
	if err != nil {
		// Convert to GlobusError if it's not already
		if _, ok := err.(*errors.GlobusError); !ok {
			return nil, errors.NewGroupsError("GroupListError", err.Error()).WithUnderlying(err)
		}
		return nil, err
	}

	// Ensure all returned group objects have the DATA_TYPE set
	for i := range groupList.Groups {
		if groupList.Groups[i].DATA_TYPE == "" {
			groupList.Groups[i].DATA_TYPE = "group"
		}
	}

	groupsResp := response.NewGroupsResponse(groupList)
	groupsResp.WithRequestID("groups-list-" + strconv.FormatInt(time.Now().UnixNano(), 10))

	return groupsResp, nil
}

// GetGroup retrieves a specific group by ID
func (c *Client) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var group Group
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID, nil, nil, &group)
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
	err := c.doRequestLowLevel(ctx, http.MethodPatch, "groups/"+groupID, nil, update, &updatedGroup)
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

// ListMembers retrieves members of a group
func (c *Client) ListMembers(ctx context.Context, groupID string, options *ListMembersOptions) (*MemberList, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.RoleID != "" {
			query.Set("role_id", options.RoleID)
		}
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.PageSize > 0 {
			query.Set("per_page", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var memberList MemberList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/members", query, nil, &memberList)
	if err != nil {
		return nil, err
	}

	// Ensure all returned member objects have the DATA_TYPE set
	for i := range memberList.Members {
		if memberList.Members[i].DATA_TYPE == "" {
			memberList.Members[i].DATA_TYPE = "member"
		}
		if memberList.Members[i].Role.DATA_TYPE == "" {
			memberList.Members[i].Role.DATA_TYPE = "role"
		}
	}

	return &memberList, nil
}

// AddMember adds a user to a group
func (c *Client) AddMember(ctx context.Context, groupID, userID, roleID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Build the request body
	body := map[string]string{
		"identity_id": userID,
		"role_id":     roleID,
	}

	return c.doRequestLowLevel(ctx, http.MethodPost, "groups/"+groupID+"/members", nil, body, nil)
}

// RemoveMember removes a user from a group
func (c *Client) RemoveMember(ctx context.Context, groupID, userID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	return c.doRequestLowLevel(ctx, http.MethodDelete, "groups/"+groupID+"/members/"+userID, nil, nil, nil)
}

// UpdateMemberRole updates a member's role in a group
func (c *Client) UpdateMemberRole(ctx context.Context, groupID, userID, roleID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Build the request body
	body := map[string]string{
		"role_id": roleID,
	}

	return c.doRequestLowLevel(ctx, http.MethodPatch, "groups/"+groupID+"/members/"+userID, nil, body, nil)
}

// ListRoles retrieves roles defined for a group
func (c *Client) ListRoles(ctx context.Context, groupID string) (*RoleList, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var roleList RoleList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/roles", nil, nil, &roleList)
	if err != nil {
		return nil, err
	}

	// Ensure all returned role objects have the DATA_TYPE set
	for i := range roleList.Roles {
		if roleList.Roles[i].DATA_TYPE == "" {
			roleList.Roles[i].DATA_TYPE = "role"
		}
	}

	return &roleList, nil
}

// GetRole retrieves a specific role by ID
func (c *Client) GetRole(ctx context.Context, groupID, roleID string) (*Role, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	var role Role
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/roles/"+roleID, nil, nil, &role)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if role.DATA_TYPE == "" {
		role.DATA_TYPE = "role"
	}

	return &role, nil
}

// CreateRole creates a new role in a group
func (c *Client) CreateRole(ctx context.Context, groupID string, role *RoleCreate) (*Role, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	if role == nil {
		return nil, fmt.Errorf("role data is required")
	}

	if role.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}

	// Set the DATA_TYPE field if not already set
	if role.DATA_TYPE == "" {
		role.DATA_TYPE = "role_create"
	}

	var createdRole Role
	err := c.doRequestLowLevel(ctx, http.MethodPost, "groups/"+groupID+"/roles", nil, role, &createdRole)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if createdRole.DATA_TYPE == "" {
		createdRole.DATA_TYPE = "role"
	}

	return &createdRole, nil
}

// UpdateRole updates an existing role
func (c *Client) UpdateRole(ctx context.Context, groupID, roleID string, update *RoleUpdate) (*Role, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	if update == nil {
		return nil, fmt.Errorf("update data is required")
	}

	// Set the DATA_TYPE field if not already set
	if update.DATA_TYPE == "" {
		update.DATA_TYPE = "role_update"
	}

	var updatedRole Role
	err := c.doRequestLowLevel(ctx, http.MethodPatch, "groups/"+groupID+"/roles/"+roleID, nil, update, &updatedRole)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if updatedRole.DATA_TYPE == "" {
		updatedRole.DATA_TYPE = "role"
	}

	return &updatedRole, nil
}

// DeleteRole deletes a role
func (c *Client) DeleteRole(ctx context.Context, groupID, roleID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	return c.doRequestLowLevel(ctx, http.MethodDelete, "groups/"+groupID+"/roles/"+roleID, nil, nil, nil)
}

// SetSubscriptionAdminVerifiedID sets a subscription ID for a group (admin-only operation)
func (c *Client) SetSubscriptionAdminVerifiedID(ctx context.Context, groupID, subscriptionID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if subscriptionID == "" {
		return fmt.Errorf("subscription ID is required")
	}

	body := map[string]string{
		"subscription_id": subscriptionID,
		"DATA_TYPE":       "subscription_id_update",
	}

	return c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID+"/subscription_id", nil, body, nil)
}

// GetGroupSubscription retrieves the subscription information for a group
func (c *Client) GetGroupSubscription(ctx context.Context, groupID string) (*GroupSubscription, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var subscription GroupSubscription
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/subscription", nil, nil, &subscription)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if subscription.DATA_TYPE == "" {
		subscription.DATA_TYPE = "group_subscription"
	}

	return &subscription, nil
}

// GetGroupBySubscriptionID retrieves a group by its subscription ID (Python SDK parity)
func (c *Client) GetGroupBySubscriptionID(ctx context.Context, subscriptionID string) (*Group, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	// Query parameter for subscription-based lookup
	query := url.Values{}
	query.Set("subscription_id", subscriptionID)

	var group Group
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups", query, nil, &group)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if group.DATA_TYPE == "" {
		group.DATA_TYPE = "group"
	}

	return &group, nil
}

// Python SDK Parity Methods - Additional functionality to match upstream

// GetGroupPolicies retrieves policy configuration for a group (Python SDK parity)
func (c *Client) GetGroupPolicies(ctx context.Context, groupID string) (*GroupPolicies, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var policies GroupPolicies
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/policies", nil, nil, &policies)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if policies.DATA_TYPE == "" {
		policies.DATA_TYPE = "group_policies"
	}

	return &policies, nil
}

// SetGroupPolicies sets policy configuration for a group (Python SDK parity)
func (c *Client) SetGroupPolicies(ctx context.Context, groupID string, policies *GroupPolicies) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	if policies == nil {
		return fmt.Errorf("policies are required")
	}

	// Set the DATA_TYPE field if not already set
	if policies.DATA_TYPE == "" {
		policies.DATA_TYPE = "group_policies_update"
	}

	return c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID+"/policies", nil, policies, nil)
}

// GetIdentityPreferences retrieves identity preferences for a group (Python SDK parity)
func (c *Client) GetIdentityPreferences(ctx context.Context, groupID, identityID string) (*IdentityPreferences, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}
	if identityID == "" {
		return nil, fmt.Errorf("identity ID is required")
	}

	var preferences IdentityPreferences
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/identity_preferences/"+identityID, nil, nil, &preferences)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if preferences.DATA_TYPE == "" {
		preferences.DATA_TYPE = "identity_preferences"
	}

	return &preferences, nil
}

// SetIdentityPreferences sets identity preferences for a group (Python SDK parity)
func (c *Client) SetIdentityPreferences(ctx context.Context, groupID, identityID string, preferences *IdentityPreferences) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if identityID == "" {
		return fmt.Errorf("identity ID is required")
	}
	if preferences == nil {
		return fmt.Errorf("preferences are required")
	}

	// Set the DATA_TYPE field if not already set
	if preferences.DATA_TYPE == "" {
		preferences.DATA_TYPE = "identity_preferences_update"
	}

	return c.doRequestLowLevel(ctx, http.MethodPut, "groups/"+groupID+"/identity_preferences/"+identityID, nil, preferences, nil)
}

// GetMembershipFields retrieves custom membership fields for a group (Python SDK parity)
func (c *Client) GetMembershipFields(ctx context.Context, groupID string) (*MembershipFields, error) {
	if groupID == "" {
		return nil, fmt.Errorf("group ID is required")
	}

	var fields MembershipFields
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/membership_fields", nil, nil, &fields)
	if err != nil {
		return nil, err
	}

	// Ensure the returned object has the DATA_TYPE set
	if fields.DATA_TYPE == "" {
		fields.DATA_TYPE = "membership_fields"
	}

	return &fields, nil
}

// SetMembershipFields sets custom membership fields for a group (Python SDK parity)
func (c *Client) SetMembershipFields(ctx context.Context, groupID string, fields *MembershipFields) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if fields == nil {
		return fmt.Errorf("membership fields are required")
	}

	// Set the DATA_TYPE field if not already set
	if fields.DATA_TYPE == "" {
		fields.DATA_TYPE = "membership_fields_update"
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
