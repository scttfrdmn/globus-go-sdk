// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/gcs"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

const testCollectionID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func newTestClient(t *testing.T, handler http.HandlerFunc) *gcs.CollectionClient {
	t.Helper()
	server := testhelpers.NewMockServer(t, handler)
	client, err := gcs.NewCollectionClient(
		context.Background(),
		server.URL,
		testCollectionID,
		testhelpers.NewTestConfig(server.URL),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestNewCollectionClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(),
			server.URL,
			testCollectionID,
			testhelpers.NewTestConfig(server.URL),
		)
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, testCollectionID, client.CollectionID())
		_ = client.Close()
	})

	t.Run("empty collectionAddress returns error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := gcs.NewCollectionClient(
			context.Background(),
			"",
			testCollectionID,
			testhelpers.NewTestConfig(server.URL),
		)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "collectionAddress", valErr.Field)
	})

	t.Run("empty collectionID returns error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := gcs.NewCollectionClient(
			context.Background(),
			server.URL,
			"",
			testhelpers.NewTestConfig(server.URL),
		)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "collectionID", valErr.Field)
	})

	t.Run("missing access token returns error", func(t *testing.T) {
		_, err := gcs.NewCollectionClient(
			context.Background(),
			"https://g-xxxxx.data.globus.org",
			testCollectionID,
			&core.Config{Scopes: []string{"test-scope"}},
		)
		assert.Error(t, err)
	})
}

func TestDefaultScopeRequirements(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := gcs.NewCollectionClient(
		context.Background(), server.URL, testCollectionID, testhelpers.NewTestConfig(server.URL),
	)
	require.NoError(t, err)
	defer client.Close()

	https, dataAccess := client.DefaultScopeRequirements()
	assert.Contains(t, https, testCollectionID)
	assert.Contains(t, https, "https")
	assert.Contains(t, dataAccess, testCollectionID)
	assert.Contains(t, dataAccess, "data_access")
}

func TestCollectionScopes(t *testing.T) {
	https, dataAccess := gcs.CollectionScopes("my-collection-uuid")
	assert.Equal(t, "https://auth.globus.org/scopes/my-collection-uuid/https", https)
	assert.Equal(t, "https://auth.globus.org/scopes/my-collection-uuid/data_access", dataAccess)
}

func TestGetCollection(t *testing.T) {
	t.Run("empty collection ID returns validation error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := client.GetCollection(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "collectionID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &gcs.Collection{
			ID:          "col-123",
			DisplayName: "My Collection",
			DataType:    "collection#1.9.0",
		}
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "col-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})

		result, err := client.GetCollection(context.Background(), "col-123")
		require.NoError(t, err)
		assert.Equal(t, "col-123", result.ID)
		assert.Equal(t, "My Collection", result.DisplayName)
	})

	t.Run("not found returns APIError", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "collection not found", "NOT_FOUND")
		})

		_, err := client.GetCollection(context.Background(), "missing")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestListCollections(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		expected := &gcs.CollectionPage{
			Data: []gcs.Collection{{ID: "col-1"}, {ID: "col-2"}},
		}
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "/collections")
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})

		result, err := client.ListCollections(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("limit and offset passed as query params", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "20", r.URL.Query().Get("offset"))
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionPage{})
		})

		_, err := client.ListCollections(context.Background(), &gcs.ListCollectionsOptions{
			Limit: 10, Offset: 20,
		})
		assert.NoError(t, err)
	})

	t.Run("filter_owned query param", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "true", r.URL.Query().Get("filter_owned"))
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionPage{})
		})

		_, err := client.ListCollections(context.Background(), &gcs.ListCollectionsOptions{
			FilterOwned: true,
		})
		assert.NoError(t, err)
	})
}

func TestUpdateCollection(t *testing.T) {
	t.Run("empty collection ID returns validation error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := client.UpdateCollection(context.Background(), "", &gcs.CollectionUpdate{})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "collectionID", valErr.Field)
	})

	t.Run("nil update returns validation error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := client.UpdateCollection(context.Background(), "col-123", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "update", valErr.Field)
	})

	t.Run("success uses PATCH", func(t *testing.T) {
		expected := &gcs.Collection{ID: "col-123", DisplayName: "Renamed"}
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Contains(t, r.URL.Path, "col-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})

		result, err := client.UpdateCollection(context.Background(), "col-123",
			&gcs.CollectionUpdate{DisplayName: "Renamed"})
		require.NoError(t, err)
		assert.Equal(t, "Renamed", result.DisplayName)
	})
}

func TestDeleteCollection(t *testing.T) {
	t.Run("empty collection ID returns validation error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
		err := client.DeleteCollection(context.Background(), "")
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})

	t.Run("success uses DELETE", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Contains(t, r.URL.Path, "col-123")
			w.WriteHeader(http.StatusNoContent)
		})

		err := client.DeleteCollection(context.Background(), "col-123")
		assert.NoError(t, err)
	})
}

func TestCollectionPager(t *testing.T) {
	t.Run("single page — HasMore false after first NextPage", func(t *testing.T) {
		calls := 0
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			// No Links.Next → last page
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionPage{
				Data: []gcs.Collection{{ID: "col-1"}},
			})
		})

		pager := client.NewCollectionPager(nil)
		assert.True(t, pager.HasMore())

		page, err := pager.NextPage(context.Background())
		require.NoError(t, err)
		assert.Len(t, page.Data, 1)
		assert.False(t, pager.HasMore())
		assert.Equal(t, 1, calls)
	})

	t.Run("ErrNoPagesRemaining after last page", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionPage{})
		})

		pager := client.NewCollectionPager(nil)
		_, err := pager.NextPage(context.Background())
		require.NoError(t, err)

		_, err = pager.NextPage(context.Background())
		require.Error(t, err)
		assert.Equal(t, gcs.ErrNoPagesRemaining, err)
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := gcs.NewCollectionClient(
		context.Background(), server.URL, testCollectionID, testhelpers.NewTestConfig(server.URL),
	)
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}
