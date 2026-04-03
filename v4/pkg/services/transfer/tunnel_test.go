// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Tests for Streams/Tunnels API — BETA (upstream v4.3.0–v4.5.0, issue #21)

func TestCreateTunnel(t *testing.T) {
	t.Run("nil document returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateTunnel(context.Background(), nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "data", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &transfer.Tunnel{ID: "tunnel-123", Status: "ACTIVE"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/tunnel", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateTunnel(context.Background(), &transfer.TunnelCreate{
			DisplayName:      "My Tunnel",
			SourceEndpointID: "ep-src",
			SourcePath:       "/data",
		})
		require.NoError(t, err)
		assert.Equal(t, "tunnel-123", result.ID)
		assert.Equal(t, "ACTIVE", result.Status)
	})

	t.Run("API error surfaced as APIError", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusBadRequest, "invalid endpoint", "BAD_REQUEST")
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateTunnel(context.Background(), &transfer.TunnelCreate{
			SourceEndpointID: "bad-ep",
		})
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok, "expected APIError, got %T", err)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})
}

func TestUpdateTunnel(t *testing.T) {
	t.Run("empty tunnel ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.UpdateTunnel(context.Background(), "", &transfer.TunnelUpdate{DisplayName: "new"})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "tunnelID", valErr.Field)
	})

	t.Run("nil update returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.UpdateTunnel(context.Background(), "tunnel-123", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "data", valErr.Field)
	})

	t.Run("success uses PUT method", func(t *testing.T) {
		expected := &transfer.Tunnel{ID: "tunnel-123", Status: "ACTIVE"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "tunnel-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.UpdateTunnel(context.Background(), "tunnel-123",
			&transfer.TunnelUpdate{DisplayName: "updated"})
		require.NoError(t, err)
		assert.Equal(t, "tunnel-123", result.ID)
	})
}

func TestGetTunnel(t *testing.T) {
	t.Run("empty tunnel ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTunnel(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "tunnelID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &transfer.Tunnel{ID: "tunnel-abc", DisplayName: "My Tunnel", Status: "ACTIVE"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "tunnel-abc")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetTunnel(context.Background(), "tunnel-abc")
		require.NoError(t, err)
		assert.Equal(t, "tunnel-abc", result.ID)
		assert.Equal(t, "My Tunnel", result.DisplayName)
	})

	t.Run("not found returns APIError", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "tunnel not found", "NOT_FOUND")
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTunnel(context.Background(), "missing")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestDeleteTunnel(t *testing.T) {
	t.Run("empty tunnel ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteTunnel(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "tunnelID", valErr.Field)
	})

	t.Run("success uses DELETE method", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Contains(t, r.URL.Path, "tunnel-123")
			w.WriteHeader(http.StatusNoContent)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteTunnel(context.Background(), "tunnel-123")
		assert.NoError(t, err)
	})
}

func TestListTunnels(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		expected := &transfer.TunnelList{
			Tunnels: []transfer.Tunnel{{ID: "t-1"}, {ID: "t-2"}},
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/tunnel_list", r.URL.Path)
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListTunnels(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Tunnels, 2)
	})

	t.Run("limit and marker passed as query params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "marker-abc", r.URL.Query().Get("marker"))
			testhelpers.RespondJSON(w, http.StatusOK, &transfer.TunnelList{})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListTunnels(context.Background(), &transfer.ListTunnelsOptions{
			Limit: 10, Marker: "marker-abc",
		})
		assert.NoError(t, err)
	})
}

func TestGetStreamAccessPoint(t *testing.T) {
	t.Run("empty SAP ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetStreamAccessPoint(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "accessPointID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &transfer.StreamAccessPoint{ID: "sap-123", EndpointID: "ep-abc"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "sap-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetStreamAccessPoint(context.Background(), "sap-123")
		require.NoError(t, err)
		assert.Equal(t, "sap-123", result.ID)
		assert.Equal(t, "ep-abc", result.EndpointID)
	})
}

func TestGetTunnelEvents(t *testing.T) {
	t.Run("empty tunnel ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTunnelEvents(context.Background(), "", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "tunnelID", valErr.Field)
	})

	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &transfer.TunnelEventList{
			Events: []transfer.TunnelEvent{{ID: "ev-1"}},
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "tunnel-123")
			assert.Contains(t, r.URL.Path, "event_list")
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetTunnelEvents(context.Background(), "tunnel-123", nil)
		require.NoError(t, err)
		assert.Len(t, result.Events, 1)
	})

	t.Run("limit and marker passed as query params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "5", r.URL.Query().Get("limit"))
			assert.Equal(t, "marker-xyz", r.URL.Query().Get("marker"))
			testhelpers.RespondJSON(w, http.StatusOK, &transfer.TunnelEventList{})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTunnelEvents(context.Background(), "tunnel-123",
			&transfer.ListTunnelEventsOptions{Limit: 5, Marker: "marker-xyz"})
		assert.NoError(t, err)
	})
}

func TestListStreamAccessPoints(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		expected := &transfer.StreamAccessPointList{
			Data: []transfer.StreamAccessPoint{{ID: "sap-1"}, {ID: "sap-2"}},
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/stream_access_point_list", r.URL.Path)
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListStreamAccessPoints(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("limit and marker passed as query params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "3", r.URL.Query().Get("limit"))
			assert.Equal(t, "marker-m", r.URL.Query().Get("marker"))
			testhelpers.RespondJSON(w, http.StatusOK, &transfer.StreamAccessPointList{})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListStreamAccessPoints(context.Background(),
			&transfer.ListTunnelsOptions{Limit: 3, Marker: "marker-m"})
		assert.NoError(t, err)
	})
}
