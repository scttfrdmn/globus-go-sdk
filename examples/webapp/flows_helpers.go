// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SimpleFlowList is a simplified response structure from the Globus Flows API
type SimpleFlowList struct {
	Flows     []SimpleFlow `json:"flows"`
	Total     int          `json:"total,omitempty"`
	HasMore   bool         `json:"has_more,omitempty"`
	NextPage  string       `json:"next_page,omitempty"`
}

// SimpleFlow is a simplified flow structure
type SimpleFlow struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Definition   string    `json:"definition,omitempty"`
	FlowOwner    string    `json:"flow_owner,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Public       bool      `json:"public,omitempty"`
	Subscription bool      `json:"subscription,omitempty"`
}

// ListFlowsOptions specifies options for listing flows
type SimpleListFlowsOptions struct {
	Limit      int    `json:"limit,omitempty"`
	Marker     string `json:"marker,omitempty"`
	FilterRole string `json:"filter_role,omitempty"`
}

// performSimpleFlowsList implements a simple Flows API client for listing flows
func (app *App) performSimpleFlowsList(ctx context.Context, limit int) (*SimpleFlowList, error) {
	// Get an access token
	accessToken, err := app.getAccessToken(ctx, "user", "https://auth.globus.org/scopes/eec9b274-0c81-4334-bdc2-54e90e689b9a/manage_flows")
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	
	// Create the request URL with query parameters
	apiURL := "https://flows.globus.org/v1/flows"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create flows list request: %w", err)
	}
	
	// Add query parameters
	q := req.URL.Query()
	q.Add("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()
	
	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flows list request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flows list request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse the response
	var flowList SimpleFlowList
	if err := json.NewDecoder(resp.Body).Decode(&flowList); err != nil {
		return nil, fmt.Errorf("failed to decode flows list response: %w", err)
	}
	
	return &flowList, nil
}