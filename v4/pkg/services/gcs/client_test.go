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

// gcsEnvelope wraps one or more data objects in the GCS result#1.0.0 envelope.
func gcsEnvelope(objs ...map[string]interface{}) map[string]interface{} {
	data := make([]map[string]interface{}, len(objs))
	copy(data, objs)
	return map[string]interface{}{
		"DATA_TYPE":          "result#1.0.0",
		"code":               "success",
		"http_response_code": 200,
		"data":               data,
	}
}

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
		_, err := client.GetCollection(context.Background(), "", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "collectionID", valErr.Field)
	})

	t.Run("success unpacks result envelope", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/collections/col-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, gcsEnvelope(map[string]interface{}{
				"DATA_TYPE": "collection#1.9.0", "id": "col-123", "display_name": "My Collection",
			}))
		})

		result, err := client.GetCollection(context.Background(), "col-123", nil)
		require.NoError(t, err)
		assert.Equal(t, "col-123", result.ID)
		assert.Equal(t, "My Collection", result.DisplayName)
	})

	t.Run("not found returns APIError", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "collection not found", "NOT_FOUND")
		})

		_, err := client.GetCollection(context.Background(), "missing", nil)
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestListCollections(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/collections", r.URL.Path)
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionListResponse{
				Data: []gcs.Collection{{ID: "col-1"}, {ID: "col-2"}},
			})
		})

		result, err := client.ListCollections(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("filter and include comma-joined; page_size/marker", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "mapped_collections,guest_collections", q.Get("filter"))
			assert.Equal(t, "private_policies", q.Get("include"))
			assert.Equal(t, "50", q.Get("page_size"))
			assert.Equal(t, "m1", q.Get("marker"))
			assert.Empty(t, q.Get("filter_owned"))
			assert.Empty(t, q.Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionListResponse{})
		})

		_, err := client.ListCollections(context.Background(), &gcs.ListCollectionsOptions{
			Filter:   []string{"mapped_collections", "guest_collections"},
			Include:  []string{"private_policies"},
			PageSize: 50,
			Marker:   "m1",
		})
		assert.NoError(t, err)
	})
}

func TestUpdateCollection(t *testing.T) {
	t.Run("empty collection ID returns validation error", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := client.UpdateCollection(context.Background(), "", &gcs.CollectionDocument{})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "collectionID", valErr.Field)
	})

	t.Run("success uses PATCH and unpacks envelope", func(t *testing.T) {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/api/collections/col-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, gcsEnvelope(map[string]interface{}{
				"DATA_TYPE": "collection#1.9.0", "id": "col-123", "display_name": "Renamed",
			}))
		})

		result, err := client.UpdateCollection(context.Background(), "col-123",
			&gcs.CollectionDocument{DisplayName: "Renamed"})
		require.NoError(t, err)
		assert.Equal(t, "Renamed", result.DisplayName)
	})
}

func TestGetGCSInfoUnauthenticated(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/info", r.URL.Path)
		assert.Empty(t, r.Header.Get("Authorization"), "GET /info must be unauthenticated")
		testhelpers.RespondJSON(w, http.StatusOK, gcsEnvelope(map[string]interface{}{
			"DATA_TYPE": "info#1.0.0", "client_id": "client-xyz",
		}))
	})

	info, err := client.GetGCSInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "client-xyz", info.ClientID)
}

func TestCreateRoleBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/roles", r.URL.Path)
		testhelpers.RespondJSON(w, http.StatusOK, gcsEnvelope(map[string]interface{}{
			"DATA_TYPE": "role#1.0.0", "id": "role-1", "role": "administrator",
		}))
	})

	role, err := client.CreateRole(context.Background(), &gcs.GCSRoleDocument{
		Principal: "urn:globus:auth:identity:x", Role: "administrator",
	})
	require.NoError(t, err)
	assert.Equal(t, "role-1", role.ID)
	assert.Equal(t, "administrator", role.Role)
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
	t.Run("marker pagination across two pages", func(t *testing.T) {
		calls := 0
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			marker := r.URL.Query().Get("marker")
			calls++
			if marker == "" {
				testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionListResponse{
					Data: []gcs.Collection{{ID: "col-1"}}, HasNextPage: true, Marker: "m1",
				})
			} else {
				assert.Equal(t, "m1", marker)
				testhelpers.RespondJSON(w, http.StatusOK, &gcs.CollectionListResponse{
					Data: []gcs.Collection{{ID: "col-2"}}, HasNextPage: false,
				})
			}
		})

		pager := client.NewCollectionPager(nil)
		var got []string
		for pager.HasNext() {
			page, err := pager.NextPage(context.Background())
			require.NoError(t, err)
			for _, c := range page {
				got = append(got, c.ID)
			}
		}
		assert.Equal(t, []string{"col-1", "col-2"}, got)
		assert.Equal(t, 2, calls)
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

// TestScopeHelpers verifies the collection (URL-format) and endpoint
// (URN-format) scope strings, matching globus-sdk-python's GCSCollectionScopes
// and GCSEndpointScopes.
func TestScopeHelpers(t *testing.T) {
	https, dataAccess := gcs.CollectionScopes("col-123")
	assert.Equal(t, "https://auth.globus.org/scopes/col-123/https", https)
	assert.Equal(t, "https://auth.globus.org/scopes/col-123/data_access", dataAccess)

	manage := gcs.EndpointManageCollectionsScope("ep-456")
	assert.Equal(t, "urn:globus:auth:scope:ep-456:manage_collections", manage)
}
