// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// CollectionClient is a client for the Globus Connect Server (GCS) Manager API.
// Despite its name it covers the full GCS Manager surface: endpoint, collections,
// storage gateways, roles, and user credentials.
//
// The GCS Manager exposes its REST API at <collectionAddress>/api. Every response
// is a result#1.0.0 envelope; single-object GETs return a one-element data array
// which is unpacked by DATA_TYPE, and lists are marker-paginated.
//
// STABILITY: EXPERIMENTAL — this client may change without notice.
type CollectionClient struct {
	collectionID string
	baseClient   *core.Client
	baseURL      string // e.g. "https://g-xxxxx.data.globus.org/api"
}

// NewCollectionClient creates a new GCS CollectionClient.
//
// collectionAddress is the base URL of the GCS Manager endpoint, e.g.
// "https://g-xxxxx.0ec8.aaaa.data.globus.org". The "/api" path prefix is
// appended automatically. collectionID is stored for scope derivation and pagers.
func NewCollectionClient(ctx context.Context, collectionAddress, collectionID string, config *core.Config) (*CollectionClient, error) {
	if collectionAddress == "" {
		return nil, &core.ValidationError{Field: "collectionAddress", Message: "GCS collection address is required"}
	}
	if collectionID == "" {
		return nil, &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}

	addr := strings.TrimRight(collectionAddress, "/")
	if !strings.HasPrefix(addr, "https://") && !strings.HasPrefix(addr, "http://") {
		addr = "https://" + addr
	}
	apiBase := addr + "/api"

	cfg := *config // shallow copy so we can mutate BaseURL
	cfg.BaseURL = apiBase

	baseClient, err := core.NewClient(&cfg)
	if err != nil {
		return nil, err
	}

	return &CollectionClient{
		collectionID: collectionID,
		baseClient:   baseClient,
		baseURL:      apiBase,
	}, nil
}

// CollectionID returns the collection ID this client was initialised with.
func (c *CollectionClient) CollectionID() string {
	return c.collectionID
}

// DefaultScopeRequirements returns the HTTPS and data_access scope strings for
// the collection this client was initialised with.
func (c *CollectionClient) DefaultScopeRequirements() (https, dataAccess string) {
	return CollectionScopes(c.collectionID)
}

// --- generic helpers ---

// do performs a request and decodes the result#1.0.0 envelope.
func (c *CollectionClient) do(ctx context.Context, method, path string, query url.Values, body interface{}) (*GCSResponse, error) {
	var env GCSResponse
	if err := c.baseClient.DoRequest(ctx, method, path, query, body, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// unpackOne performs a request, decodes the envelope, and unpacks the single
// object of the named DATA_TYPE into target.
func (c *CollectionClient) unpackOne(ctx context.Context, method, path string, query url.Values, body interface{}, datatype string, target interface{}) error {
	env, err := c.do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	return env.unpack(datatype, target)
}

func commaJoin(q url.Values, key string, vals []string) {
	if len(vals) > 0 {
		q.Set(key, strings.Join(vals, ","))
	}
}

// --- Info / endpoint ---

// GetGCSInfo fetches the unauthenticated GET /info document.
func (c *CollectionClient) GetGCSInfo(ctx context.Context) (*GCSInfo, error) {
	var env GCSResponse
	if err := c.baseClient.DoRequestNoAuth(ctx, http.MethodGet, "/info", nil, nil, &env); err != nil {
		return nil, err
	}
	var info GCSInfo
	if err := env.unpack("info", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetEndpoint retrieves the endpoint document (GET /endpoint).
func (c *CollectionClient) GetEndpoint(ctx context.Context) (*Endpoint, error) {
	var ep Endpoint
	if err := c.unpackOne(ctx, http.MethodGet, "/endpoint", nil, nil, "endpoint", &ep); err != nil {
		return nil, err
	}
	return &ep, nil
}

// UpdateEndpoint updates the endpoint document (PATCH /endpoint).
func (c *CollectionClient) UpdateEndpoint(ctx context.Context, doc *EndpointDocument, opts *UpdateEndpointOptions) (*Endpoint, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "endpoint document is required"}
	}
	query := url.Values{}
	if opts != nil {
		commaJoin(query, "include", opts.Include)
	}
	var ep Endpoint
	if err := c.unpackOne(ctx, http.MethodPatch, "/endpoint", query, doc, "endpoint", &ep); err != nil {
		return nil, err
	}
	return &ep, nil
}

// --- Collections ---

// GetCollection retrieves a single collection by ID (GET /collections/{id}).
func (c *CollectionClient) GetCollection(ctx context.Context, collectionID string, opts *GetCollectionOptions) (*Collection, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}
	query := url.Values{}
	if opts != nil {
		for k, v := range opts.QueryParams {
			query.Set(k, v)
		}
	}
	var col Collection
	if err := c.unpackOne(ctx, http.MethodGet, fmt.Sprintf("/collections/%s", collectionID), query, nil, "collection", &col); err != nil {
		return nil, err
	}
	return &col, nil
}

// ListCollections returns a page of collections (GET /collections). Use
// NewCollectionPager to iterate all pages.
func (c *CollectionClient) ListCollections(ctx context.Context, options *ListCollectionsOptions) (*CollectionListResponse, error) {
	query := url.Values{}
	if options != nil {
		if options.MappedCollectionID != "" {
			query.Set("mapped_collection_id", options.MappedCollectionID)
		}
		commaJoin(query, "filter", options.Filter)
		commaJoin(query, "include", options.Include)
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}

	var result CollectionListResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/collections", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCollection creates a collection (POST /collections).
func (c *CollectionClient) CreateCollection(ctx context.Context, doc *CollectionDocument) (*Collection, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "collection document is required"}
	}
	var col Collection
	if err := c.unpackOne(ctx, http.MethodPost, "/collections", nil, doc, "collection", &col); err != nil {
		return nil, err
	}
	return &col, nil
}

// UpdateCollection updates a collection (PATCH /collections/{id}).
func (c *CollectionClient) UpdateCollection(ctx context.Context, collectionID string, doc *CollectionDocument) (*Collection, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "collection document is required"}
	}
	var col Collection
	if err := c.unpackOne(ctx, http.MethodPatch, fmt.Sprintf("/collections/%s", collectionID), nil, doc, "collection", &col); err != nil {
		return nil, err
	}
	return &col, nil
}

// DeleteCollection removes a collection (DELETE /collections/{id}).
func (c *CollectionClient) DeleteCollection(ctx context.Context, collectionID string) error {
	if collectionID == "" {
		return &core.ValidationError{Field: "collectionID", Message: "collection ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/collections/%s", collectionID), nil, nil, nil)
}

// --- Storage gateways ---

// GetStorageGatewayList returns a page of storage gateways (GET /storage_gateways).
func (c *CollectionClient) GetStorageGatewayList(ctx context.Context, options *StorageGatewayListOptions) (*StorageGatewayListResponse, error) {
	query := url.Values{}
	if options != nil {
		commaJoin(query, "include", options.Include)
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}
	var result StorageGatewayListResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/storage_gateways", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateStorageGateway creates a storage gateway (POST /storage_gateways).
func (c *CollectionClient) CreateStorageGateway(ctx context.Context, doc *StorageGatewayDocument) (*StorageGateway, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "storage gateway document is required"}
	}
	var sg StorageGateway
	if err := c.unpackOne(ctx, http.MethodPost, "/storage_gateways", nil, doc, "storage_gateway", &sg); err != nil {
		return nil, err
	}
	return &sg, nil
}

// GetStorageGateway retrieves a storage gateway (GET /storage_gateways/{id}).
func (c *CollectionClient) GetStorageGateway(ctx context.Context, storageGatewayID string, opts *GetStorageGatewayOptions) (*StorageGateway, error) {
	if storageGatewayID == "" {
		return nil, &core.ValidationError{Field: "storageGatewayID", Message: "storage gateway ID is required"}
	}
	query := url.Values{}
	if opts != nil {
		commaJoin(query, "include", opts.Include)
	}
	var sg StorageGateway
	if err := c.unpackOne(ctx, http.MethodGet, fmt.Sprintf("/storage_gateways/%s", storageGatewayID), query, nil, "storage_gateway", &sg); err != nil {
		return nil, err
	}
	return &sg, nil
}

// UpdateStorageGateway updates a storage gateway (PATCH /storage_gateways/{id}).
// Upstream returns the raw envelope here (no data unpacking).
func (c *CollectionClient) UpdateStorageGateway(ctx context.Context, storageGatewayID string, doc *StorageGatewayDocument) (*GCSResponse, error) {
	if storageGatewayID == "" {
		return nil, &core.ValidationError{Field: "storageGatewayID", Message: "storage gateway ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "storage gateway document is required"}
	}
	return c.do(ctx, http.MethodPatch, fmt.Sprintf("/storage_gateways/%s", storageGatewayID), nil, doc)
}

// DeleteStorageGateway removes a storage gateway (DELETE /storage_gateways/{id}).
func (c *CollectionClient) DeleteStorageGateway(ctx context.Context, storageGatewayID string) error {
	if storageGatewayID == "" {
		return &core.ValidationError{Field: "storageGatewayID", Message: "storage gateway ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/storage_gateways/%s", storageGatewayID), nil, nil, nil)
}

// --- Roles ---

// GetRoleList returns a page of roles (GET /roles).
func (c *CollectionClient) GetRoleList(ctx context.Context, options *RoleListOptions) (*RoleListResponse, error) {
	query := url.Values{}
	if options != nil {
		if options.CollectionID != "" {
			query.Set("collection_id", options.CollectionID)
		}
		if options.Include != "" {
			query.Set("include", options.Include) // raw, not comma-joined
		}
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}
	var result RoleListResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/roles", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateRole creates a role (POST /roles).
func (c *CollectionClient) CreateRole(ctx context.Context, doc *GCSRoleDocument) (*GCSRole, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "role document is required"}
	}
	var role GCSRole
	if err := c.unpackOne(ctx, http.MethodPost, "/roles", nil, doc, "role", &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRole retrieves a role (GET /roles/{id}).
func (c *CollectionClient) GetRole(ctx context.Context, roleID string) (*GCSRole, error) {
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	var role GCSRole
	if err := c.unpackOne(ctx, http.MethodGet, fmt.Sprintf("/roles/%s", roleID), nil, nil, "role", &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// DeleteRole removes a role (DELETE /roles/{id}).
func (c *CollectionClient) DeleteRole(ctx context.Context, roleID string) error {
	if roleID == "" {
		return &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/roles/%s", roleID), nil, nil, nil)
}

// --- User credentials ---

// GetUserCredentialList returns a page of user credentials (GET /user_credentials).
func (c *CollectionClient) GetUserCredentialList(ctx context.Context, options *UserCredentialListOptions) (*UserCredentialListResponse, error) {
	query := url.Values{}
	if options != nil {
		if options.StorageGateway != "" {
			query.Set("storage_gateway", options.StorageGateway)
		}
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}
	var result UserCredentialListResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/user_credentials", query, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateUserCredential creates a user credential (POST /user_credentials).
func (c *CollectionClient) CreateUserCredential(ctx context.Context, doc *UserCredentialDocument) (*UserCredential, error) {
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "user credential document is required"}
	}
	var uc UserCredential
	if err := c.unpackOne(ctx, http.MethodPost, "/user_credentials", nil, doc, "user_credential", &uc); err != nil {
		return nil, err
	}
	return &uc, nil
}

// GetUserCredential retrieves a user credential (GET /user_credentials/{id}).
func (c *CollectionClient) GetUserCredential(ctx context.Context, userCredentialID string) (*UserCredential, error) {
	if userCredentialID == "" {
		return nil, &core.ValidationError{Field: "userCredentialID", Message: "user credential ID is required"}
	}
	var uc UserCredential
	if err := c.unpackOne(ctx, http.MethodGet, fmt.Sprintf("/user_credentials/%s", userCredentialID), nil, nil, "user_credential", &uc); err != nil {
		return nil, err
	}
	return &uc, nil
}

// UpdateUserCredential updates a user credential (PATCH /user_credentials/{id}).
func (c *CollectionClient) UpdateUserCredential(ctx context.Context, userCredentialID string, doc *UserCredentialDocument) (*UserCredential, error) {
	if userCredentialID == "" {
		return nil, &core.ValidationError{Field: "userCredentialID", Message: "user credential ID is required"}
	}
	if doc == nil {
		return nil, &core.ValidationError{Field: "doc", Message: "user credential document is required"}
	}
	var uc UserCredential
	if err := c.unpackOne(ctx, http.MethodPatch, fmt.Sprintf("/user_credentials/%s", userCredentialID), nil, doc, "user_credential", &uc); err != nil {
		return nil, err
	}
	return &uc, nil
}

// DeleteUserCredential removes a user credential (DELETE /user_credentials/{id}).
func (c *CollectionClient) DeleteUserCredential(ctx context.Context, userCredentialID string) error {
	if userCredentialID == "" {
		return &core.ValidationError{Field: "userCredentialID", Message: "user credential ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/user_credentials/%s", userCredentialID), nil, nil, nil)
}

// Close releases resources held by the client.
func (c *CollectionClient) Close() error {
	return c.baseClient.Close()
}
