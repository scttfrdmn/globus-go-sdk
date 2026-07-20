// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Bookmarks — upstream Python SDK v4.6.0 (amended v4.8.0).
//
// DIVERGENCE: upstream exposes bookmark CRUD on a separate experimental
// TransferClientV2 that speaks JSON:API under /v2/bookmarks. The Go v4 module
// has no TransferClientV2, so these methods are folded into transfer.Client —
// the same approach already used for the Streams/Tunnels API. See
// docs/divergence.md.
//
// The v4.8.0 amendment removed the `pinned` bookmark field; it is intentionally
// absent from the models below.

// Bookmark represents a Globus Transfer bookmark (a named collection + path).
type Bookmark struct {
	ID           string
	Name         string
	Path         string
	CollectionID string
}

// BookmarkCreate is the payload for creating a bookmark.
// Collection, Name, and Path are required.
type BookmarkCreate struct {
	// Collection is the collection (endpoint) UUID the bookmark points at.
	Collection string
	// Name is the human-readable bookmark name.
	Name string
	// Path is the path on the collection the bookmark points at.
	Path string
	// AdditionalFields, if set, are merged into the JSON:API attributes object.
	AdditionalFields map[string]interface{}
}

// BookmarkUpdate is the payload for updating a bookmark.
// Fields left nil are omitted from the PATCH document.
type BookmarkUpdate struct {
	// Name, if set, updates the bookmark name.
	Name *string
	// Path, if set, updates the bookmark path.
	Path *string
	// AdditionalFields, if set, are merged into the JSON:API attributes object.
	AdditionalFields map[string]interface{}
}

// ListBookmarksOptions controls which bookmarks are returned.
//
// Upstream list_bookmarks (as of v4.8.1) is not paginated — the service
// returns the full set in a single JSON:API response — so there is no marker
// or per_page option here.
type ListBookmarksOptions struct {
	// QueryParams are passed through as-is on the request.
	QueryParams url.Values
}

// jsonAPIResourceRef is a JSON:API resource identifier ({type, id}).
type jsonAPIResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// bookmarkAttributes holds the JSON:API attributes for a bookmark.
type bookmarkAttributes struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// bookmarkResource is a single JSON:API bookmark resource object.
type bookmarkResource struct {
	Type          string             `json:"type"`
	ID            string             `json:"id"`
	Attributes    bookmarkAttributes `json:"attributes"`
	Relationships struct {
		Collection struct {
			Data jsonAPIResourceRef `json:"data"`
		} `json:"collection"`
	} `json:"relationships"`
}

// toBookmark flattens a JSON:API resource object into the public Bookmark type.
func (r *bookmarkResource) toBookmark() *Bookmark {
	return &Bookmark{
		ID:           r.ID,
		Name:         r.Attributes.Name,
		Path:         r.Attributes.Path,
		CollectionID: r.Relationships.Collection.Data.ID,
	}
}

// bookmarkSingleResponse wraps a single-resource JSON:API response.
type bookmarkSingleResponse struct {
	Data bookmarkResource `json:"data"`
}

// bookmarkListResponse wraps a multi-resource JSON:API response.
type bookmarkListResponse struct {
	Data []bookmarkResource `json:"data"`
}

// BookmarkList is a list of bookmarks.
type BookmarkList struct {
	Bookmarks []Bookmark
}

// buildAttributes merges the fixed attributes with any additional fields.
func buildAttributes(base map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	for k, v := range additional {
		base[k] = v
	}
	return base
}

// CreateBookmark creates a new bookmark.
// Upstream: POST /v2/bookmarks (Python SDK v4.6.0).
func (c *Client) CreateBookmark(ctx context.Context, data *BookmarkCreate) (*Bookmark, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "bookmark data is required"}
	}
	if data.Collection == "" {
		return nil, &core.ValidationError{Field: "Collection", Message: "collection is required"}
	}
	if data.Name == "" {
		return nil, &core.ValidationError{Field: "Name", Message: "name is required"}
	}
	if data.Path == "" {
		return nil, &core.ValidationError{Field: "Path", Message: "path is required"}
	}

	attrs := buildAttributes(map[string]interface{}{
		"name": data.Name,
		"path": data.Path,
	}, data.AdditionalFields)

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "Bookmark",
			"attributes": attrs,
			"relationships": map[string]interface{}{
				"collection": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "Collection",
						"id":   data.Collection,
					},
				},
			},
		},
	}

	var resp bookmarkSingleResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v2/bookmarks", nil, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data.toBookmark(), nil
}

// GetBookmark retrieves a bookmark by ID.
// Upstream: GET /v2/bookmarks/{id} (Python SDK v4.6.0).
func (c *Client) GetBookmark(ctx context.Context, bookmarkID string) (*Bookmark, error) {
	if bookmarkID == "" {
		return nil, &core.ValidationError{Field: "bookmarkID", Message: "bookmark ID is required"}
	}

	var resp bookmarkSingleResponse
	path := fmt.Sprintf("/v2/bookmarks/%s", bookmarkID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data.toBookmark(), nil
}

// ListBookmarks lists the caller's bookmarks.
// Upstream: GET /v2/bookmarks (Python SDK v4.6.0). Not paginated as of v4.8.1.
func (c *Client) ListBookmarks(ctx context.Context, options *ListBookmarksOptions) (*BookmarkList, error) {
	var query url.Values
	if options != nil {
		query = options.QueryParams
	}

	var resp bookmarkListResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/bookmarks", query, nil, &resp); err != nil {
		return nil, err
	}

	list := &BookmarkList{Bookmarks: make([]Bookmark, 0, len(resp.Data))}
	for i := range resp.Data {
		list.Bookmarks = append(list.Bookmarks, *resp.Data[i].toBookmark())
	}
	return list, nil
}

// UpdateBookmark updates a bookmark by ID.
// Upstream: PATCH /v2/bookmarks/{id} (Python SDK v4.6.0).
func (c *Client) UpdateBookmark(ctx context.Context, bookmarkID string, update *BookmarkUpdate) (*Bookmark, error) {
	if bookmarkID == "" {
		return nil, &core.ValidationError{Field: "bookmarkID", Message: "bookmark ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	attrs := map[string]interface{}{}
	if update.Name != nil {
		attrs["name"] = *update.Name
	}
	if update.Path != nil {
		attrs["path"] = *update.Path
	}
	attrs = buildAttributes(attrs, update.AdditionalFields)

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "Bookmark",
			"attributes": attrs,
		},
	}

	var resp bookmarkSingleResponse
	path := fmt.Sprintf("/v2/bookmarks/%s", bookmarkID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPatch, path, nil, body, &resp); err != nil {
		return nil, err
	}
	return resp.Data.toBookmark(), nil
}

// DeleteBookmark deletes a bookmark by ID.
// Upstream: DELETE /v2/bookmarks/{id} (Python SDK v4.6.0).
func (c *Client) DeleteBookmark(ctx context.Context, bookmarkID string) error {
	if bookmarkID == "" {
		return &core.ValidationError{Field: "bookmarkID", Message: "bookmark ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/v2/bookmarks/%s", bookmarkID), nil, nil, nil)
}
