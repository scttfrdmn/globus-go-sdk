// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Tests for the Beta Streams/Tunnels + StreamAccessPoints surface, which speaks
// JSON:API under /v2/ on the experimental TransferClientV2 upstream.

// jsonapiResource builds a JSON:API single-resource envelope.
func jsonapiResource(typ, id string, attrs map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{"type": typ, "id": id, "attributes": attrs},
	}
}

// jsonapiCollection builds a JSON:API collection envelope.
func jsonapiCollection(items ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"data": items}
}

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

	t.Run("success posts JSON:API doc to /v2/tunnels", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/tunnels", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusCreated,
				jsonapiResource("Tunnel", "tunnel-123", map[string]interface{}{"label": "My Tunnel", "status": "ACTIVE"}))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateTunnel(context.Background(), &transfer.TunnelCreate{
			ListenerStreamAccessPoint:  "sap-l",
			InitiatorStreamAccessPoint: "sap-i",
			Label:                      "My Tunnel",
		})
		require.NoError(t, err)
		assert.Equal(t, "tunnel-123", result.ID)
		assert.Equal(t, "ACTIVE", result.Status)
	})
}

func TestUpdateTunnel(t *testing.T) {
	t.Run("empty tunnel ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.UpdateTunnel(context.Background(), "", &transfer.TunnelUpdate{Label: "new"})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "tunnelID", valErr.Field)
	})

	t.Run("success uses PATCH on /v2/tunnels/{id}", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/v2/tunnels/tunnel-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK,
				jsonapiResource("Tunnel", "tunnel-123", map[string]interface{}{"label": "updated"}))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.UpdateTunnel(context.Background(), "tunnel-123", &transfer.TunnelUpdate{Label: "updated"})
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

	t.Run("success flattens JSON:API resource", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/tunnels/tunnel-abc", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK,
				jsonapiResource("Tunnel", "tunnel-abc", map[string]interface{}{"label": "My Tunnel"}))
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetTunnel(context.Background(), "tunnel-abc")
		require.NoError(t, err)
		assert.Equal(t, "tunnel-abc", result.ID)
		assert.Equal(t, "My Tunnel", result.DisplayName)
	})
}

func TestDeleteTunnel(t *testing.T) {
	t.Run("success uses DELETE on /v2/tunnels/{id}", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/v2/tunnels/tunnel-123", r.URL.Path)
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
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/tunnels", r.URL.Path)
		testhelpers.RespondJSON(w, http.StatusOK, jsonapiCollection(
			map[string]interface{}{"type": "Tunnel", "id": "t-1", "attributes": map[string]interface{}{}},
			map[string]interface{}{"type": "Tunnel", "id": "t-2", "attributes": map[string]interface{}{}},
		))
	})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	result, err := client.ListTunnels(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, result.Tunnels, 2)
	assert.Equal(t, "t-1", result.Tunnels[0].ID)
}

func TestGetStreamAccessPoint(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/stream_access_points/sap-123", r.URL.Path)
		testhelpers.RespondJSON(w, http.StatusOK,
			jsonapiResource("StreamAccessPoint", "sap-123", map[string]interface{}{"endpoint_id": "ep-abc"}))
	})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	result, err := client.GetStreamAccessPoint(context.Background(), "sap-123")
	require.NoError(t, err)
	assert.Equal(t, "sap-123", result.ID)
}

func TestListStreamAccessPoints(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/stream_access_points", r.URL.Path)
		testhelpers.RespondJSON(w, http.StatusOK, jsonapiCollection(
			map[string]interface{}{"type": "StreamAccessPoint", "id": "sap-1", "attributes": map[string]interface{}{}},
			map[string]interface{}{"type": "StreamAccessPoint", "id": "sap-2", "attributes": map[string]interface{}{}},
		))
	})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	result, err := client.ListStreamAccessPoints(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
}

func TestGetTunnelEvents(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/tunnels/tunnel-123/events", r.URL.Path)
		// Wire shape from upstream fixture testing/data/transfer/v2/get_tunnel_events.py
		// (4.10.0): integer ids, string details, zone-less "time" (issue #66).
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": [
			{"type": "TunnelEvent", "id": 2768, "attributes": {"code": "STARTED", "is_error": false,
			 "description": "started", "details": "Attempting tunnel establishment",
			 "time": "2026-02-12T21:59:01.857473"}},
			{"type": "TunnelEvent", "id": 2770, "attributes": {"code": "TUNNEL_FAILED", "is_error": true,
			 "description": "failed", "details": "Peer unreachable", "time": "2026-02-12T22:00:00.000001"}}
		]}`))
	})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	result, err := client.GetTunnelEvents(context.Background(), "tunnel-123", nil)
	require.NoError(t, err)
	require.Len(t, result.Events, 2)
	ev := result.Events[0]
	assert.Equal(t, "2768", ev.ID)
	assert.Equal(t, "tunnel-123", ev.TunnelID)
	assert.Equal(t, "STARTED", ev.Code)
	assert.Equal(t, "started", ev.Description)
	assert.Equal(t, "Attempting tunnel establishment", ev.DetailText)
	assert.False(t, ev.IsError)
	require.NotNil(t, ev.OccurredAt)
	assert.True(t, ev.OccurredAt.Equal(time.Date(2026, 2, 12, 21, 59, 1, 857473000, time.UTC)))
	assert.Equal(t, "Attempting tunnel establishment", ev.Details["details"])
	assert.True(t, result.Events[1].IsError)
}
