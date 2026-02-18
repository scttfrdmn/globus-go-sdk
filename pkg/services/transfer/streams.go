// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Tunnel represents a Globus Streams tunnel, which provides a persistent
// channel for streaming data between endpoints.
// Added in Python SDK v4.3.0.
type Tunnel struct {
	// ID is the unique identifier for the tunnel
	ID string `json:"id"`
	// DisplayName is the human-readable name for the tunnel
	DisplayName string `json:"display_name,omitempty"`
	// Owner is the identity of the tunnel owner
	Owner string `json:"owner,omitempty"`
	// SourceEndpointID is the source endpoint for the tunnel
	SourceEndpointID string `json:"source_endpoint_id,omitempty"`
	// SourcePath is the path on the source endpoint
	SourcePath string `json:"source_path,omitempty"`
	// Status is the current status of the tunnel (e.g., "ACTIVE", "INACTIVE")
	Status string `json:"status,omitempty"`
	// CreatedAt is when the tunnel was created
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// UpdatedAt is when the tunnel was last updated
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// ExpiresAt is when the tunnel will expire (nil means no expiration)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// Metadata holds additional tunnel metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelList represents a paginated list of tunnels
type TunnelList struct {
	Tunnels []Tunnel `json:"DATA"`
	Total   int      `json:"total,omitempty"`
	HasMore bool     `json:"has_next_page,omitempty"`
	Marker  string   `json:"next_marker,omitempty"`
}

// CreateTunnelData is the payload builder for creating a new tunnel.
// Added in Python SDK v4.3.0.
type CreateTunnelData struct {
	// DisplayName is the human-readable name for the tunnel
	DisplayName string `json:"display_name"`
	// SourceEndpointID is the source endpoint for the tunnel
	SourceEndpointID string `json:"source_endpoint_id"`
	// SourcePath is the path on the source endpoint
	SourcePath string `json:"source_path"`
	// ExpiresIn specifies when the tunnel should expire (in seconds from creation)
	ExpiresIn *int `json:"expires_in,omitempty"`
	// Metadata holds additional tunnel metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTunnelData is the payload for updating an existing tunnel.
// Added in Python SDK v4.3.0.
type UpdateTunnelData struct {
	// DisplayName is the updated human-readable name
	DisplayName string `json:"display_name,omitempty"`
	// Metadata holds additional tunnel metadata to update
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ListTunnelsOptions represents options for listing tunnels
type ListTunnelsOptions struct {
	// Limit is the maximum number of tunnels to return
	Limit int `url:"limit,omitempty"`
	// Marker is the pagination cursor for getting the next page
	Marker string `url:"marker,omitempty"`
}

// StreamAccessPoint represents a Globus Stream Access Point, providing
// access to real-time data streams.
// Added in Python SDK v4.3.0.
type StreamAccessPoint struct {
	// ID is the unique identifier for the access point
	ID string `json:"id"`
	// TunnelID is the associated tunnel ID
	TunnelID string `json:"tunnel_id,omitempty"`
	// EndpointID is the endpoint providing the stream
	EndpointID string `json:"endpoint_id,omitempty"`
	// Path is the path on the endpoint for the stream
	Path string `json:"path,omitempty"`
	// AccessURL is the URL for accessing the stream
	AccessURL string `json:"access_url,omitempty"`
	// ExpiresAt is when the access point expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// Metadata holds additional access point metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelEvent represents an event associated with a tunnel
// Added in Python SDK v4.4.0.
type TunnelEvent struct {
	// ID is the unique identifier for the event
	ID string `json:"id"`
	// TunnelID is the associated tunnel ID
	TunnelID string `json:"tunnel_id"`
	// Code is the event type code
	Code string `json:"code"`
	// Description describes the event
	Description string `json:"description,omitempty"`
	// Details holds event-specific details
	Details map[string]interface{} `json:"details,omitempty"`
	// OccurredAt is when the event occurred
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
}

// TunnelEventList represents a paginated list of tunnel events
type TunnelEventList struct {
	Events  []TunnelEvent `json:"DATA"`
	Total   int           `json:"total,omitempty"`
	HasMore bool          `json:"has_next_page,omitempty"`
	Marker  string        `json:"next_marker,omitempty"`
}

// ListTunnelEventsOptions represents options for listing tunnel events
type ListTunnelEventsOptions struct {
	// Limit is the maximum number of events to return
	Limit int `url:"limit,omitempty"`
	// Marker is the pagination cursor for getting the next page
	Marker string `url:"marker,omitempty"`
}

// CreateTunnel creates a new Globus Streams tunnel.
// Added in Python SDK v4.3.0.
func (c *Client) CreateTunnel(ctx context.Context, data *CreateTunnelData) (*Tunnel, error) {
	if data == nil {
		return nil, fmt.Errorf("tunnel data is required")
	}
	if data.DisplayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	if data.SourceEndpointID == "" {
		return nil, fmt.Errorf("source endpoint ID is required")
	}

	var tunnel Tunnel
	err := c.doRequestLowLevel(ctx, http.MethodPost, "tunnel", nil, data, &tunnel)
	if err != nil {
		return nil, err
	}
	return &tunnel, nil
}

// GetTunnel retrieves a tunnel by its ID.
// Added in Python SDK v4.3.0.
func (c *Client) GetTunnel(ctx context.Context, tunnelID string) (*Tunnel, error) {
	if tunnelID == "" {
		return nil, fmt.Errorf("tunnel ID is required")
	}

	var tunnel Tunnel
	err := c.doRequestLowLevel(ctx, http.MethodGet, "tunnel/"+tunnelID, nil, nil, &tunnel)
	if err != nil {
		return nil, err
	}
	return &tunnel, nil
}

// UpdateTunnel updates an existing tunnel.
// Added in Python SDK v4.3.0.
func (c *Client) UpdateTunnel(ctx context.Context, tunnelID string, data *UpdateTunnelData) (*Tunnel, error) {
	if tunnelID == "" {
		return nil, fmt.Errorf("tunnel ID is required")
	}
	if data == nil {
		return nil, fmt.Errorf("update data is required")
	}

	var tunnel Tunnel
	err := c.doRequestLowLevel(ctx, http.MethodPut, "tunnel/"+tunnelID, nil, data, &tunnel)
	if err != nil {
		return nil, err
	}
	return &tunnel, nil
}

// DeleteTunnel deletes a tunnel by its ID.
// Added in Python SDK v4.3.0.
func (c *Client) DeleteTunnel(ctx context.Context, tunnelID string) error {
	if tunnelID == "" {
		return fmt.Errorf("tunnel ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodDelete, "tunnel/"+tunnelID, nil, nil, nil)
}

// ListTunnels retrieves the list of tunnels owned by the current user.
// Added in Python SDK v4.3.0.
func (c *Client) ListTunnels(ctx context.Context, options *ListTunnelsOptions) (*TunnelList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}
	var list TunnelList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "tunnel_list", query, nil, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetStreamAccessPoint retrieves a Stream Access Point by its ID.
// Stream Access Points provide access to real-time data streams via Globus Streams.
// Added in Python SDK v4.3.0.
func (c *Client) GetStreamAccessPoint(ctx context.Context, accessPointID string) (*StreamAccessPoint, error) {
	if accessPointID == "" {
		return nil, fmt.Errorf("access point ID is required")
	}

	var ap StreamAccessPoint
	err := c.doRequestLowLevel(ctx, http.MethodGet, "stream_access_point/"+accessPointID, nil, nil, &ap)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

// GetTunnelEvents fetches events associated with a tunnel.
// Added in Python SDK v4.4.0.
func (c *Client) GetTunnelEvents(ctx context.Context, tunnelID string, options *ListTunnelEventsOptions) (*TunnelEventList, error) {
	if tunnelID == "" {
		return nil, fmt.Errorf("tunnel ID is required")
	}

	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}
	var list TunnelEventList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "tunnel/"+tunnelID+"/event_list", query, nil, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
