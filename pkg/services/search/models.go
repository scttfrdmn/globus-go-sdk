// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"time"
)

// Index represents a Globus Search index
type Index struct {
	ID                  string                 `json:"id,omitempty"`
	DisplayName         string                 `json:"display_name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	IsActive            bool                   `json:"is_active,omitempty"`
	IsPublic            bool                   `json:"is_public,omitempty"`
	IsMonitored         bool                   `json:"is_monitored,omitempty"`
	MonitoringFrequency int                    `json:"monitoring_frequency,omitempty"`
	CreatedBy           string                 `json:"created_by,omitempty"`
	CreatedAt           time.Time              `json:"created_at,omitempty"`
	UpdatedAt           time.Time              `json:"updated_at,omitempty"`
	MaxSize             int64                  `json:"max_size_in_megabytes,omitempty"`
	DefinitionDocument  map[string]interface{} `json:"definition_document,omitempty"`
}

// IndexList is a list of indexes
type IndexList struct {
	Indexes []Index `json:"index_list,omitempty"`
}

// IndexCreateRequest is the request to create a new index
type IndexCreateRequest struct {
	DisplayName         string                 `json:"display_name"`
	Description         string                 `json:"description,omitempty"`
	IsMonitored         bool                   `json:"is_monitored,omitempty"`
	MonitoringFrequency int                    `json:"monitoring_frequency,omitempty"`
	DefinitionDocument  map[string]interface{} `json:"definition_document,omitempty"`
}

// IndexUpdateRequest is the request to update an index
type IndexUpdateRequest struct {
	DisplayName         string                 `json:"display_name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	IsMonitored         bool                   `json:"is_monitored,omitempty"`
	MonitoringFrequency int                    `json:"monitoring_frequency,omitempty"`
	IsActive            bool                   `json:"is_active,omitempty"`
	DefinitionDocument  map[string]interface{} `json:"definition_document,omitempty"`
}

// ListIndexesOptions are the options for listing indexes
type ListIndexesOptions struct {
	Limit     int    `url:"limit,omitempty"`
	Offset    int    `url:"offset,omitempty"`
	Marker    string `url:"marker,omitempty"`
	PerPage   int    `url:"per_page,omitempty"` // Alias for Limit
	IsPublic  bool   `url:"is_public,omitempty"`
	IsActive  bool   `url:"is_active,omitempty"`
	CreatedBy string `url:"created_by,omitempty"`
	ByPath    string `url:"by_path,omitempty"`
}

// SearchDocument represents a document in a Globus Search index
type SearchDocument struct {
	Subject   string                 `json:"subject"`
	Content   map[string]interface{} `json:"content"`
	VisibleTo []string               `json:"visible_to,omitempty"`
	IndexedAt time.Time              `json:"indexed_at,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	DeletedAt time.Time              `json:"deleted_at,omitempty"`
	Version   string                 `json:"version,omitempty"`
	GMETA     map[string]interface{} `json:"gmeta,omitempty"`
}

// IngestRequest represents a request to ingest documents
type IngestRequest struct {
	IndexID   string           `json:"index_id"`
	Documents []SearchDocument `json:"documents"`
	Task      *IngestTask      `json:"task,omitempty"`
}

// IngestTask represents the task configuration for an ingest operation
type IngestTask struct {
	TaskID          string `json:"task_id,omitempty"`
	ProcessingState string `json:"processing_state,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	DetailLocation  string `json:"detail_location,omitempty"`
}

// IngestResponse is the response from an ingest operation
type IngestResponse struct {
	Task      IngestTask `json:"task"`
	Succeeded int        `json:"succeeded"`
	Failed    int        `json:"failed"`
	Total     int        `json:"total"`
}

// SearchOptions are the options for search queries
type SearchOptions struct {
	Limit             int      `url:"limit,omitempty"`
	Offset            int      `url:"offset,omitempty"`
	PageSize          int      `url:"page_size,omitempty"`  // Alias for Limit
	PageToken         string   `url:"page_token,omitempty"` // Alias for Marker
	Sort              []string `url:"sort,omitempty"`
	Filter            string   `url:"filter,omitempty"`
	Facets            []string `url:"facets,omitempty"`
	FacetSize         int      `url:"facet_size,omitempty"`
	IncludeAllContent bool     `url:"include_all_content,omitempty"`
	ByPath            string   `url:"by_path,omitempty"`
}

// SearchRequest represents a search query request
type SearchRequest struct {
	IndexID string         `json:"index_id"`
	Query   string         `json:"q"`
	Options *SearchOptions `json:"options,omitempty"`
}

// FacetValue represents a value in a facet
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// Facet represents a facet in search results
type Facet struct {
	Name   string       `json:"name"`
	Type   string       `json:"type"`
	Values []FacetValue `json:"values"`
}

// SearchResult represents a document in search results
type SearchResult struct {
	Subject   string                 `json:"subject"`
	Content   map[string]interface{} `json:"content"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
	Score     float64                `json:"score"`
}

// SearchResponse is the response from a search operation. Results are decoded
// from the upstream "gmeta" key; Marker carries the scroll cursor.
type SearchResponse struct {
	Count       int            `json:"count"`
	Total       int            `json:"total"`
	Offset      int            `json:"offset"`
	GMeta       []SearchResult `json:"gmeta"`
	Facets      []Facet        `json:"facet_results,omitempty"`
	HasNextPage bool           `json:"has_next_page"`
	Marker      string         `json:"marker,omitempty"`

	// Results is an alias for GMeta retained for the iterator helpers.
	Results   []SearchResult `json:"-"`
	HasMore   bool           `json:"-"`
	PageToken string         `json:"-"`
}

// DeleteDocumentsRequest represents a request to delete documents by subject.
type DeleteDocumentsRequest struct {
	IndexID  string   `json:"-"`
	Subjects []string `json:"subjects"`
}

// DeleteDocumentsResponse is the response from a delete operation. task_id is at
// the top level.
type DeleteDocumentsResponse struct {
	TaskID string `json:"task_id"`
}

// TaskStatusResponse represents the status of a task.
type TaskStatusResponse struct {
	TaskID         string `json:"task_id"`
	IndexID        string `json:"index_id"`
	State          string `json:"state"`
	CreatedAt      string `json:"created_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
	DetailLocation string `json:"detail_location,omitempty"`
	Message        string `json:"message,omitempty"`
}
