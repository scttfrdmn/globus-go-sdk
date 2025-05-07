// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

// SearchQuery represents a Globus Search query
type SearchQuery struct {
	Q             string                   `json:"q,omitempty"`
	Limit         int                      `json:"limit,omitempty"`
	Offset        int                      `json:"offset,omitempty"`
	Advanced      bool                     `json:"advanced,omitempty"`
	Terms         []string                 `json:"terms,omitempty"`
	Filters       []map[string]interface{} `json:"filters,omitempty"`
	Sort          []map[string]string      `json:"sort,omitempty"`
	Facets        []string                 `json:"facets,omitempty"`
	BoostIndices  []string                 `json:"boost_indices,omitempty"`
	ExcludeFields []string                 `json:"exclude_fields,omitempty"`
	IncludeFields []string                 `json:"include_fields,omitempty"`
	IndexMapping  map[string]bool          `json:"index_mapping,omitempty"`
}

// NewBasicQuery creates a new basic search query with just a search term
func NewBasicQuery(searchTerm string, limit int) *SearchQuery {
	return &SearchQuery{
		Q:     searchTerm,
		Limit: limit,
	}
}
