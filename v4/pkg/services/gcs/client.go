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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// CollectionClient is an EXPERIMENTAL client for the Globus Connect Server (GCS)
// manager Collections API.
//
// The GCS manager exposes its REST API at <collectionAddress>/api.
// All methods require appropriate scopes — use CollectionScopes to obtain the
// HTTPS and data_access scope strings for a given collection ID.
//
// STABILITY: EXPERIMENTAL — this client may change without notice.
type CollectionClient struct {
	collectionID string
	baseClient   *core.Client
	baseURL      string // e.g. "https://g-xxxxx.data.globus.org/api"
}

// NewCollectionClient creates a new GCS CollectionClient.
//
// collectionAddress is the base URL of the GCS manager endpoint, e.g.
// "https://g-xxxxx.0ec8.aaaa.data.globus.org". The "/api" path prefix is
// appended automatically.
//
// collectionID is the UUID of the default collection. It is used by
// DefaultScopeRequirements and stored for NewCollectionPager.
func NewCollectionClient(ctx context.Context, collectionAddress, collectionID string, config *core.Config) (*CollectionClient, error) {
	if collectionAddress == "" {
		return nil, &core.ValidationError{
			Field:   "collectionAddress",
			Message: "GCS collection address is required",
		}
	}
	if collectionID == "" {
		return nil, &core.ValidationError{
			Field:   "collectionID",
			Message: "collection ID is required",
		}
	}

	// Normalise: strip trailing slash, ensure HTTPS
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

// GetCollection retrieves a single collection by ID.
func (c *CollectionClient) GetCollection(ctx context.Context, collectionID string) (*Collection, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{
			Field:   "collectionID",
			Message: "collection ID is required",
		}
	}

	var collection Collection
	path := fmt.Sprintf("/collections/%s", collectionID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

// ListCollections returns the first page of collections visible to the caller.
// Use NewCollectionPager to iterate through all pages.
func (c *CollectionClient) ListCollections(ctx context.Context, options *ListCollectionsOptions) (*CollectionPage, error) {
	query := url.Values{}

	if options != nil {
		if options.FilterOwned {
			query.Set("filter_owned", "true")
		}
		if options.MappedCollectionID != "" {
			query.Set("mapped_collection_id", options.MappedCollectionID)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var page CollectionPage
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/collections", query, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// listCollectionsAbsolute fetches a page using an absolute URL returned in
// Links.Next. This is used internally by CollectionPager.
func (c *CollectionClient) listCollectionsAbsolute(ctx context.Context, absoluteURL string) (*CollectionPage, error) {
	// Strip the base URL prefix to get just the path+query
	path := strings.TrimPrefix(absoluteURL, c.baseURL)
	if path == absoluteURL {
		// URL doesn't start with our base — use it as-is (relative)
		path = absoluteURL
	}

	// Split path and query
	var query url.Values
	if idx := strings.Index(path, "?"); idx >= 0 {
		var err error
		query, err = url.ParseQuery(path[idx+1:])
		if err != nil {
			return nil, err
		}
		path = path[:idx]
	}

	var page CollectionPage
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdateCollection applies a partial update to a collection.
func (c *CollectionClient) UpdateCollection(ctx context.Context, collectionID string, update *CollectionUpdate) (*Collection, error) {
	if collectionID == "" {
		return nil, &core.ValidationError{
			Field:   "collectionID",
			Message: "collection ID is required",
		}
	}
	if update == nil {
		return nil, &core.ValidationError{
			Field:   "update",
			Message: "update data is required",
		}
	}

	var result Collection
	path := fmt.Sprintf("/collections/%s", collectionID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPatch, path, nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCollection removes a collection.
func (c *CollectionClient) DeleteCollection(ctx context.Context, collectionID string) error {
	if collectionID == "" {
		return &core.ValidationError{
			Field:   "collectionID",
			Message: "collection ID is required",
		}
	}

	path := fmt.Sprintf("/collections/%s", collectionID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, path, nil, nil, nil)
}

// NewCollectionPager returns a pager for iterating through all collection pages.
func (c *CollectionClient) NewCollectionPager(options *ListCollectionsOptions) *CollectionPager {
	p := &CollectionPager{
		client: c,
		opts:   options,
	}
	p.inner = paging.NewJSONAPIPaginator(p.fetchPageFn)
	return p
}

// Close releases resources held by the client.
func (c *CollectionClient) Close() error {
	return c.baseClient.Close()
}
