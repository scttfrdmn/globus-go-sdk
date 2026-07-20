// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Tests for bookmark CRUD — upstream Python SDK v4.6.0 (JSON:API /v2/bookmarks),
// amended v4.8.0 (removed the `pinned` field). Folded into transfer.Client as a
// documented divergence from upstream's separate TransferClientV2.

// singleBookmarkResponse builds a JSON:API single-resource response body.
func singleBookmarkResponse(id, name, path, collection string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "Bookmark",
			"id":         id,
			"attributes": map[string]interface{}{"name": name, "path": path},
			"relationships": map[string]interface{}{
				"collection": map[string]interface{}{
					"data": map[string]interface{}{"type": "Collection", "id": collection},
				},
			},
		},
		"meta": map[string]interface{}{"request_id": "abc123"},
	}
}

func TestCreateBookmark(t *testing.T) {
	t.Run("nil data returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateBookmark(context.Background(), nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "data", valErr.Field)
	})

	t.Run("missing required fields", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateBookmark(context.Background(), &transfer.BookmarkCreate{Name: "n", Path: "/p"})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "Collection", valErr.Field)
	})

	t.Run("success builds JSON:API create document", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/bookmarks", r.URL.Path)

			bodyBytes, _ := io.ReadAll(r.Body)
			var body map[string]interface{}
			require.NoError(t, json.Unmarshal(bodyBytes, &body))
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Bookmark", data["type"])
			attrs := data["attributes"].(map[string]interface{})
			assert.Equal(t, "public datasets", attrs["name"])
			assert.Equal(t, "/data/public", attrs["path"])
			rel := data["relationships"].(map[string]interface{})
			coll := rel["collection"].(map[string]interface{})["data"].(map[string]interface{})
			assert.Equal(t, "Collection", coll["type"])
			assert.Equal(t, "collection-1", coll["id"])
			// The removed-in-4.8.0 pinned field must never be sent.
			assert.NotContains(t, attrs, "pinned")

			testhelpers.RespondJSON(w, http.StatusCreated,
				singleBookmarkResponse("bm-1", "public datasets", "/data/public", "collection-1"))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateBookmark(context.Background(), &transfer.BookmarkCreate{
			Collection: "collection-1",
			Name:       "public datasets",
			Path:       "/data/public",
		})
		require.NoError(t, err)
		assert.Equal(t, "bm-1", result.ID)
		assert.Equal(t, "public datasets", result.Name)
		assert.Equal(t, "/data/public", result.Path)
		assert.Equal(t, "collection-1", result.CollectionID)
	})
}

func TestGetBookmark(t *testing.T) {
	t.Run("empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetBookmark(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "bookmarkID", valErr.Field)
	})

	t.Run("success flattens JSON:API resource", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/bookmarks/bm-1", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK,
				singleBookmarkResponse("bm-1", "my bookmark", "/data", "collection-9"))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetBookmark(context.Background(), "bm-1")
		require.NoError(t, err)
		assert.Equal(t, "bm-1", result.ID)
		assert.Equal(t, "my bookmark", result.Name)
		assert.Equal(t, "/data", result.Path)
		assert.Equal(t, "collection-9", result.CollectionID)
	})
}

func TestListBookmarks(t *testing.T) {
	t.Run("success flattens JSON:API data array", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/bookmarks", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{
				"data": []interface{}{
					singleBookmarkResponse("bm-1", "public", "/pub", "c1")["data"],
					singleBookmarkResponse("bm-2", "private", "/priv", "c2")["data"],
				},
				"meta": map[string]interface{}{"request_id": "xj8AWyQr8"},
			})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListBookmarks(context.Background(), nil)
		require.NoError(t, err)
		require.Len(t, result.Bookmarks, 2)
		assert.Equal(t, "bm-1", result.Bookmarks[0].ID)
		assert.Equal(t, "public", result.Bookmarks[0].Name)
		assert.Equal(t, "c1", result.Bookmarks[0].CollectionID)
		assert.Equal(t, "bm-2", result.Bookmarks[1].ID)
		assert.Equal(t, "/priv", result.Bookmarks[1].Path)
	})

	t.Run("empty list", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListBookmarks(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, result.Bookmarks)
	})
}

func TestUpdateBookmark(t *testing.T) {
	t.Run("empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		name := "x"
		_, err = client.UpdateBookmark(context.Background(), "", &transfer.BookmarkUpdate{Name: &name})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "bookmarkID", valErr.Field)
	})

	t.Run("nil update returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.UpdateBookmark(context.Background(), "bm-1", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "update", valErr.Field)
	})

	t.Run("success sends PATCH with only provided attributes and no relationships", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/v2/bookmarks/bm-1", r.URL.Path)

			bodyBytes, _ := io.ReadAll(r.Body)
			var body map[string]interface{}
			require.NoError(t, json.Unmarshal(bodyBytes, &body))
			data := body["data"].(map[string]interface{})
			assert.Equal(t, "Bookmark", data["type"])
			attrs := data["attributes"].(map[string]interface{})
			assert.Equal(t, "renamed", attrs["name"])
			// Path was not set, so it must be absent from the PATCH document.
			assert.NotContains(t, attrs, "path")
			// Update documents carry no relationships block.
			assert.NotContains(t, data, "relationships")

			testhelpers.RespondJSON(w, http.StatusOK,
				singleBookmarkResponse("bm-1", "renamed", "/data", "c1"))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		name := "renamed"
		result, err := client.UpdateBookmark(context.Background(), "bm-1", &transfer.BookmarkUpdate{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "renamed", result.Name)
	})
}

func TestDeleteBookmark(t *testing.T) {
	t.Run("empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteBookmark(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "bookmarkID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/v2/bookmarks/bm-1", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteBookmark(context.Background(), "bm-1")
		assert.NoError(t, err)
	})
}
