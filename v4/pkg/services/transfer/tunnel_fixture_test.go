// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Wire fixtures copied from upstream globus-sdk-python 4.10.0
// src/globus_sdk/testing/data/transfer/v2/{get_tunnel,get_stream_access_point}.py
// (issue #82: the JSON:API decode dropped relationships and most attributes).

const tunnelFixture = `{
  "data": {
    "attributes": {
      "created_time": "2025-12-12T21:11:50.525278",
      "initiator_ip_address": null, "initiator_port": null,
      "label": "Buzz Tester", "lifetime_mins": 360,
      "listener_ip_address": null, "listener_port": null,
      "restartable": false, "state": "AWAITING_LISTENER",
      "status": "The tunnel is waiting for listening",
      "submission_id": "292b0054-7084-46eb-83d6-7a6821b1f77e"
    },
    "id": "tun-1",
    "relationships": {
      "initiator": {"data": {"id": "ap-init", "type": "StreamAccessPoint"}},
      "listener": {"data": {"id": "ap-listen", "type": "StreamAccessPoint"}},
      "owner": {"data": {"id": "4d443580-012d-4954-816f-e0592bd356e1", "type": "Identity"}}
    },
    "type": "Tunnel"
  },
  "meta": {"request_id": "M6kFaS949"}
}`

// An active tunnel: same shape with the listener/initiator endpoints filled in.
const activeTunnelItem = `{
  "type": "Tunnel", "id": "tun-2",
  "attributes": {"label": "live", "state": "ACTIVE", "status": "The tunnel is active",
    "listener_ip_address": "10.0.0.5", "listener_port": 50000,
    "initiator_ip_address": "10.0.0.9", "initiator_port": 50001,
    "lifetime_mins": 60, "restartable": true},
  "relationships": {
    "listener": {"data": {"id": "ap-l2", "type": "StreamAccessPoint"}},
    "initiator": {"data": {"id": "ap-i2", "type": "StreamAccessPoint"}},
    "owner": {"data": null},
    "watchers": {"data": [{"id": "x", "type": "Identity"}]}
  }
}`

const accessPointFixture = `{
  "data": {
    "attributes": {
      "advertised_owner": "john@globus.org",
      "contact_email": null, "contact_info": null, "department": null,
      "description": null, "display_name": "Buzz Dev Listener",
      "info_link": null, "keywords": null, "organization": null,
      "tlsftp_server": "tlsftp://s-463c7.e7d5e.8540.test3.zones.dnsteam.globuscs.info:443"
    },
    "id": "ap-1",
    "relationships": {
      "host_endpoint": {"data": {"id": "d6428474-c308-4a2d-8a86-d377915d978b", "type": "Endpoint"}}
    },
    "type": "StreamAccessPoint"
  },
  "meta": {"request_id": "55QRq2iBa"}
}`

func serveRaw(t *testing.T, wantPath, body string) *transfer.Client {
	t.Helper()
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, wantPath, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	})
	c, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestGetTunnel_DecodesUpstreamFixture(t *testing.T) {
	c := serveRaw(t, "/v2/tunnels/tun-1", tunnelFixture)
	tun, err := c.GetTunnel(context.Background(), "tun-1")
	require.NoError(t, err)

	assert.Equal(t, "tun-1", tun.ID)
	assert.Equal(t, "Buzz Tester", tun.Label)
	assert.Equal(t, "Buzz Tester", tun.DisplayName, "deprecated alias of Label")
	assert.Equal(t, "AWAITING_LISTENER", tun.State)
	assert.Equal(t, "The tunnel is waiting for listening", tun.Status)
	assert.Equal(t, "ap-listen", tun.ListenerAccessPointID)
	assert.Equal(t, "ap-init", tun.InitiatorAccessPointID)
	assert.Equal(t, "4d443580-012d-4954-816f-e0592bd356e1", tun.Owner)
	assert.Empty(t, tun.ListenerIPAddress, "null until a listener connects")
	assert.Nil(t, tun.ListenerPort)
	assert.Nil(t, tun.InitiatorPort)
	require.NotNil(t, tun.LifetimeMins)
	assert.Equal(t, 360, *tun.LifetimeMins)
	assert.False(t, tun.Restartable)
	assert.Equal(t, "292b0054-7084-46eb-83d6-7a6821b1f77e", tun.SubmissionID)
	require.NotNil(t, tun.CreatedAt)
	assert.True(t, tun.CreatedAt.Equal(time.Date(2025, 12, 12, 21, 11, 50, 525278000, time.UTC)))
	assert.Equal(t, transfer.JSONAPIRelationship{Type: "Identity", ID: "4d443580-012d-4954-816f-e0592bd356e1"}, tun.Relationships["owner"])
	assert.Len(t, tun.Relationships, 3)
	assert.Equal(t, "Buzz Tester", tun.Metadata["label"], "raw attributes still available")
}

func TestListTunnels_ActiveTunnelAndOddRelationships(t *testing.T) {
	c := serveRaw(t, "/v2/tunnels", `{"data": [`+activeTunnelItem+`]}`)
	list, err := c.ListTunnels(context.Background(), nil)
	require.NoError(t, err, "null and to-many relationships must not fail the decode")
	require.Len(t, list.Tunnels, 1)
	tun := list.Tunnels[0]

	assert.Equal(t, "ACTIVE", tun.State)
	assert.Equal(t, "10.0.0.5", tun.ListenerIPAddress)
	require.NotNil(t, tun.ListenerPort)
	assert.Equal(t, 50000, *tun.ListenerPort)
	assert.Equal(t, "10.0.0.9", tun.InitiatorIPAddress)
	require.NotNil(t, tun.InitiatorPort)
	assert.Equal(t, 50001, *tun.InitiatorPort)
	assert.True(t, tun.Restartable)
	assert.Equal(t, "ap-l2", tun.ListenerAccessPointID)
	assert.Equal(t, "ap-i2", tun.InitiatorAccessPointID)
	assert.Empty(t, tun.Owner, "null owner relationship")
	assert.NotContains(t, tun.Relationships, "watchers", "to-many relationships are not surfaced as to-one")
	assert.NotContains(t, tun.Relationships, "owner")
}

func TestGetStreamAccessPoint_DecodesUpstreamFixture(t *testing.T) {
	c := serveRaw(t, "/v2/stream_access_points/ap-1", accessPointFixture)
	ap, err := c.GetStreamAccessPoint(context.Background(), "ap-1")
	require.NoError(t, err)

	assert.Equal(t, "ap-1", ap.ID)
	assert.Equal(t, "Buzz Dev Listener", ap.DisplayName)
	assert.Equal(t, "john@globus.org", ap.AdvertisedOwner)
	assert.Equal(t, "tlsftp://s-463c7.e7d5e.8540.test3.zones.dnsteam.globuscs.info:443", ap.TLSFTPServer)
	assert.Equal(t, "d6428474-c308-4a2d-8a86-d377915d978b", ap.EndpointID)
	assert.Equal(t, "Endpoint", ap.Relationships["host_endpoint"].Type)
	for name, v := range map[string]string{
		"ContactEmail": ap.ContactEmail, "ContactInfo": ap.ContactInfo, "Department": ap.Department,
		"Description": ap.Description, "InfoLink": ap.InfoLink, "Organization": ap.Organization,
	} {
		assert.Empty(t, v, "%s is null in the fixture", name)
	}
	assert.Nil(t, ap.Keywords)
}

func TestListStreamAccessPoints_PopulatesContactFields(t *testing.T) {
	c := serveRaw(t, "/v2/stream_access_points", `{"data": [{
		"type": "StreamAccessPoint", "id": "ap-2",
		"attributes": {"display_name": "Lab AP", "contact_email": "ops@example.org",
		  "contact_info": "+1 555 0100", "department": "Physics", "description": "Detector feed",
		  "info_link": "https://example.org/ap", "organization": "Example U",
		  "keywords": "stream,detector"},
		"relationships": {"host_endpoint": {"data": {"id": "ep-9", "type": "Endpoint"}}}}]}`)
	list, err := c.ListStreamAccessPoints(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, list.Data, 1)
	ap := list.Data[0]
	assert.Equal(t, "Lab AP", ap.DisplayName)
	assert.Equal(t, "ops@example.org", ap.ContactEmail)
	assert.Equal(t, "+1 555 0100", ap.ContactInfo)
	assert.Equal(t, "Physics", ap.Department)
	assert.Equal(t, "Detector feed", ap.Description)
	assert.Equal(t, "https://example.org/ap", ap.InfoLink)
	assert.Equal(t, "Example U", ap.Organization)
	assert.Equal(t, transfer.Keywords{"stream", "detector"}, ap.Keywords)
	assert.Equal(t, "ep-9", ap.EndpointID)
}
