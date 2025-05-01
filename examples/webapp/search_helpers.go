// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SimpleSearchResponse is a simplified response from the Globus Search API
type SimpleSearchResponse struct {
	Total     int                      `json:"total"`
	Count     int                      `json:"count"`
	Offset    int                      `json:"offset"`
	HasMore   bool                     `json:"has_more"`
	GMeta     []map[string]interface{} `json:"gmeta"`
	Facets    map[string]interface{}   `json:"facets,omitempty"`
	QueryTime float64                  `json:"query_time,omitempty"`
}

// SimpleSearchQuery is a simple search query
type SimpleSearchQuery struct {
	Q     string `json:"q,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// performSimpleSearch performs a simple search against the Globus Search API
func (app *App) performSimpleSearch(ctx context.Context, searchTerm string, limit int) (*SimpleSearchResponse, error) {
	// Let's simplify this and use the access token directly
	accessToken, err := app.getAccessToken(ctx, "user", "urn:globus:auth:scope:search.api.globus.org:all")
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	bearerToken := accessToken
	
	// Create the request
	apiURL := "https://search.api.globus.org/v1/search"
	
	// Create the query
	query := SimpleSearchQuery{
		Q:     searchTerm,
		Limit: limit,
	}
	
	// Marshal the query
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}
	
	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(queryBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}
	
	// Add headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse the response
	var searchResponse SimpleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}
	
	return &searchResponse, nil
}