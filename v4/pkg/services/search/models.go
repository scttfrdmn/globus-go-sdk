// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import "time"

// Index represents a Globus Search index
type Index struct {
	ID                  string                 `json:"id"`
	DisplayName         string                 `json:"display_name"`
	Description         string                 `json:"description,omitempty"`
	Status              string                 `json:"status"`
	MaxSizeInMB         int                    `json:"max_size_in_mb,omitempty"`
	SizeInMB            float64                `json:"size_in_mb,omitempty"`
	NumEntries          int                    `json:"num_entries,omitempty"`
	NumSubjects         int                    `json:"num_subjects,omitempty"`
	SubscriptionID      string                 `json:"subscription_id,omitempty"`
	Created             time.Time              `json:"created,omitempty"`
	LastModified        time.Time              `json:"last_modified,omitempty"`
	Settings            map[string]interface{} `json:"settings,omitempty"`
}

// IndexCreate represents the data needed to create a new search index
type IndexCreate struct {
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description,omitempty"`
	SubscriptionID string                 `json:"subscription_id,omitempty"`
	MaxSizeInMB    int                    `json:"max_size_in_mb,omitempty"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
}

// IndexUpdate represents the data to update in a search index
type IndexUpdate struct {
	DisplayName string                 `json:"display_name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	Q                string                   `json:"q"`                           // Query string
	Filters          []map[string]interface{} `json:"filters,omitempty"`           // Filters
	Facets           []string                 `json:"facets,omitempty"`            // Facets to return
	Sort             []map[string]interface{} `json:"sort,omitempty"`              // Sort criteria
	Offset           int                      `json:"offset,omitempty"`            // Pagination offset
	Limit            int                      `json:"limit,omitempty"`             // Results per page
	AdvancedQuery    bool                     `json:"advanced,omitempty"`          // Use advanced query syntax
	BypassVisible    bool                     `json:"bypass_visible_to,omitempty"` // Bypass visibility checks
}

// SearchResults represents search query results
type SearchResults struct {
	Count       int                      `json:"count"`
	Offset      int                      `json:"offset"`
	HasNextPage bool                     `json:"has_next_page"`
	Total       int                      `json:"total"`
	GMeta       []GMetaResult            `json:"gmeta"`
	Facets      map[string]interface{}   `json:"facets,omitempty"`
}

// GMetaResult represents a single search result
type GMetaResult struct {
	Subject  string                   `json:"subject"`
	Entries  []GMetaEntry             `json:"entries"`
	Content  []map[string]interface{} `json:"content,omitempty"`
}

// GMetaEntry represents a single entry in search results
type GMetaEntry struct {
	EntryID      string                 `json:"entry_id"`
	Subject      string                 `json:"subject"`
	Content      map[string]interface{} `json:"content"`
	VisibleTo    []string               `json:"visible_to,omitempty"`
	Created      time.Time              `json:"created,omitempty"`
	LastModified time.Time              `json:"last_modified,omitempty"`
}

// IngestEntry represents a document to ingest
type IngestEntry struct {
	Subject   string                 `json:"subject"`
	VisibleTo []string               `json:"visible_to,omitempty"`
	Content   map[string]interface{} `json:"content"`
	ID        string                 `json:"id,omitempty"`
}

// IngestBatch represents multiple documents to ingest
type IngestBatch struct {
	Ingest  []IngestEntry `json:"ingest"`
	Entries []IngestEntry `json:"entries"` // Alternative field name
}

// IngestResponse represents the response from ingesting a document
type IngestResponse struct {
	TaskID   string `json:"task_id"`
	Accepted int    `json:"accepted"`
	Message  string `json:"message,omitempty"`
}

// IngestBatchResponse represents the response from batch ingest
type IngestBatchResponse struct {
	TaskID   string `json:"task_id"`
	Accepted int    `json:"accepted"`
	Rejected int    `json:"rejected,omitempty"`
	Message  string `json:"message,omitempty"`
}

// Role represents a role assignment in a search index
type Role struct {
	ID        string    `json:"id"`
	IndexID   string    `json:"index_id"`
	Principal string    `json:"principal"`
	RoleID    string    `json:"role_id"`
	RoleName  string    `json:"role_name,omitempty"`
	Created   time.Time `json:"created,omitempty"`
}

// RoleList represents a list of roles
type RoleList struct {
	Roles []Role `json:"role_list"`
}
