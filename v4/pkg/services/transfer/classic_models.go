// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

// This file holds option and response types for the classic Transfer routes
// added in the Phase 2 parity audit. Many admin documents are open-ended service
// documents represented as map[string]interface{} passthrough.

// EndpointSearchOptions controls EndpointSearch (GET /endpoint_search).
type EndpointSearchOptions struct {
	FilterFulltext      string
	FilterScope         string
	FilterOwnerID       string
	FilterHostEndpoint  string
	FilterNonFunctional *bool // wire is 1/0
	FilterEntityType    string
	Limit               int
	Offset              int
}

// EndpointSearchResult is a page of endpoint_search results.
type EndpointSearchResult struct {
	DataType    string     `json:"DATA_TYPE"`
	Data        []Endpoint `json:"DATA"`
	Offset      int        `json:"offset"`
	Limit       int        `json:"limit"`
	HasNextPage bool       `json:"has_next_page"`
}

// GenericResponse is a passthrough Transfer response envelope for endpoints
// whose body shape is not otherwise modeled.
type GenericResponse = map[string]interface{}

// TaskEventList is a page of task events (task_event_list). Marker/limit-offset
// depending on the route; both keys are captured.
type TaskEventList struct {
	DataType string                   `json:"DATA_TYPE"`
	Data     []map[string]interface{} `json:"DATA"`
	Offset   int                      `json:"offset"`
	Limit    int                      `json:"limit"`
	Total    int                      `json:"total"`
}

// ListTaskEventsOptions controls TaskEventList / EndpointManagerTaskEventList.
type ListTaskEventsOptions struct {
	Limit         int
	Offset        int
	FilterIsError *bool // wire is 1/0
}

// NullableMarkerList is a page of results paginated by a nullable next_marker
// (task_successful_transfers, task_skipped_errors). Iteration stops when
// NextMarker is null.
type NullableMarkerList struct {
	DataType   string                   `json:"DATA_TYPE"`
	Data       []map[string]interface{} `json:"DATA"`
	NextMarker *string                  `json:"next_marker"`
}

// EndpointManagerTaskList is a page of admin tasks paginated by last_key.
type EndpointManagerTaskList struct {
	DataType string                   `json:"DATA_TYPE"`
	Data     []map[string]interface{} `json:"DATA"`
	LastKey  *string                  `json:"last_key"`
}

// EndpointManagerTaskListOptions controls EndpointManagerTaskList.
type EndpointManagerTaskListOptions struct {
	FilterStatus      []string // comma-joined
	FilterTaskID      []string // comma-joined
	FilterOwnerID     string
	FilterEndpoint    string
	FilterEndpointUse string
	LastKey           string
	Limit             int
}

// SharedEndpointList is a page of shared endpoints (next_token paginated).
type SharedEndpointList struct {
	DataType        string                   `json:"DATA_TYPE"`
	SharedEndpoints []map[string]interface{} `json:"shared_endpoints"`
	NextToken       *string                  `json:"next_token"`
}
