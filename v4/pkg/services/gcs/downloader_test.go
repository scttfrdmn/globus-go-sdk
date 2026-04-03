// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/gcs"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewDownloader(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := gcs.NewCollectionClient(
		context.Background(), server.URL, testCollectionID, testhelpers.NewTestConfig(server.URL),
	)
	require.NoError(t, err)
	defer client.Close()

	dl := gcs.NewDownloader(client)
	require.NotNil(t, dl)
	assert.NoError(t, dl.Close())
}

func TestReadFile(t *testing.T) {
	t.Run("empty URI returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(), server.URL, testCollectionID, testhelpers.NewTestConfig(server.URL),
		)
		require.NoError(t, err)
		defer client.Close()

		dl := gcs.NewDownloader(client)
		defer dl.Close()

		_, err = dl.ReadFile(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "fileURI", valErr.Field)
	})

	t.Run("success returns file content", func(t *testing.T) {
		content := []byte("hello, world\n")
		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/share/file.txt", r.URL.Path)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
		}))
		defer fileServer.Close()

		managerServer := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(), managerServer.URL, testCollectionID, testhelpers.NewTestConfig(managerServer.URL),
		)
		require.NoError(t, err)
		defer client.Close()

		dl := gcs.NewDownloaderWithToken(client, "test-access-token")
		defer dl.Close()

		data, err := dl.ReadFile(context.Background(), fileServer.URL+"/share/file.txt")
		require.NoError(t, err)
		assert.Equal(t, content, data)
	})

	t.Run("ReadFileAsText returns string", func(t *testing.T) {
		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("text content"))
		}))
		defer fileServer.Close()

		managerServer := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(), managerServer.URL, testCollectionID, testhelpers.NewTestConfig(managerServer.URL),
		)
		require.NoError(t, err)
		defer client.Close()

		dl := gcs.NewDownloaderWithToken(client, "test-access-token")
		defer dl.Close()

		text, err := dl.ReadFileAsText(context.Background(), fileServer.URL+"/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "text content", text)
	})

	t.Run("server error returns APIError", func(t *testing.T) {
		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("permission denied"))
		}))
		defer fileServer.Close()

		managerServer := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(), managerServer.URL, testCollectionID, testhelpers.NewTestConfig(managerServer.URL),
		)
		require.NoError(t, err)
		defer client.Close()

		dl := gcs.NewDownloader(client)
		defer dl.Close()

		_, err = dl.ReadFile(context.Background(), fileServer.URL+"/restricted.txt")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok, "expected APIError, got %T", err)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	})

	t.Run("Authorization header sent when token provided", func(t *testing.T) {
		var capturedAuth string
		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer fileServer.Close()

		managerServer := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := gcs.NewCollectionClient(
			context.Background(), managerServer.URL, testCollectionID, testhelpers.NewTestConfig(managerServer.URL),
		)
		require.NoError(t, err)
		defer client.Close()

		dl := gcs.NewDownloaderWithToken(client, "my-https-token")
		defer dl.Close()

		_, err = dl.ReadFile(context.Background(), fileServer.URL+"/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Bearer my-https-token", capturedAuth)
	})
}

func TestDownloaderClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := gcs.NewCollectionClient(
		context.Background(), server.URL, testCollectionID, testhelpers.NewTestConfig(server.URL),
	)
	require.NoError(t, err)
	defer client.Close()

	dl := gcs.NewDownloader(client)
	assert.NoError(t, dl.Close())
	assert.NoError(t, dl.Close()) // idempotent
}
