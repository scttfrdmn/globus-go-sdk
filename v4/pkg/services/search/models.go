// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import "time"

// Index represents a Globus Search index
type Index struct {
	ID             string                 `json:"id"`
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description,omitempty"`
	Status         string                 `json:"status"`
	MaxSizeInMB    int                    `json:"max_size_in_mb,omitempty"`
	SizeInMB       float64                `json:"size_in_mb,omitempty"`
	NumEntries     int                    `json:"num_entries,omitempty"`
	NumSubjects    int                    `json:"num_subjects,omitempty"`
	SubscriptionID string                 `json:"subscription_id,omitempty"`
	Created        time.Time              `json:"created,omitempty"`
	LastModified   time.Time              `json:"last_modified,omitempty"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
}

// IndexCreate represents the data needed to create a new search index. Upstream
// create_index sends only display_name and description.
type IndexCreate struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

// IndexUpdate represents the data to update in a search index. Upstream
// update_index sends only display_name and description.
type IndexUpdate struct {
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SearchQueryVersion is the envelope version SearchQueryV1 always carries.
const SearchQueryVersion = "query#1.0.0"

// SearchQuery represents a GSearchRequest (SearchQueryV1) posted to the search
// endpoint. Version defaults to SearchQueryVersion when empty.
type SearchQuery struct {
	Version          string                   `json:"@version"`
	Q                string                   `json:"q"`
	Filters          []map[string]interface{} `json:"filters,omitempty"`
	Facets           []map[string]interface{} `json:"facets,omitempty"`
	PostFacetFilters []map[string]interface{} `json:"post_facet_filters,omitempty"`
	Boosts           []map[string]interface{} `json:"boosts,omitempty"`
	Sort             []map[string]interface{} `json:"sort,omitempty"`
	Offset           int                      `json:"offset,omitempty"`
	Limit            int                      `json:"limit,omitempty"`
	AdvancedQuery    bool                     `json:"advanced,omitempty"`
}

// ScrollQuery is the body for a scroll query (POST /index/{id}/scroll). Marker
// carries the cursor from the previous page's response.
type ScrollQuery struct {
	Q             string `json:"q"`
	Limit         int    `json:"limit,omitempty"`
	AdvancedQuery bool   `json:"advanced,omitempty"`
	Marker        string `json:"marker,omitempty"`
}

// SearchGetOptions holds query params for the GET search variant. Offset and
// Limit are omitted from the request when zero.
type SearchGetOptions struct {
	Q        string
	Offset   int
	Limit    int
	Advanced bool
}

// SearchResults represents search query results. Marker is populated by scroll
// queries and carries the cursor for the next page.
type SearchResults struct {
	Count        int                      `json:"count"`
	Offset       int                      `json:"offset"`
	HasNextPage  bool                     `json:"has_next_page"`
	Total        int                      `json:"total"`
	GMeta        []GMetaResult            `json:"gmeta"`
	FacetResults []map[string]interface{} `json:"facet_results,omitempty"`
	Marker       string                   `json:"marker,omitempty"`
}

// GMetaResult represents a single search result
type GMetaResult struct {
	Subject string                   `json:"subject"`
	Entries []GMetaEntry             `json:"entries"`
	Content []map[string]interface{} `json:"content,omitempty"`
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

// GMetaEntryDocument builds a single-entry ingest document (ingest_type
// "GMetaEntry"). Use NewGMetaEntryIngest to construct one.
type GMetaEntryDocument struct {
	Subject   string                 `json:"subject"`
	VisibleTo []string               `json:"visible_to,omitempty"`
	Content   map[string]interface{} `json:"content"`
	ID        string                 `json:"id,omitempty"`
}

// NewGMetaEntryIngest builds a single-entry ingest document with the required
// {ingest_type, ingest_data} envelope, suitable for passing to Ingest.
func NewGMetaEntryIngest(entry GMetaEntryDocument) map[string]interface{} {
	return map[string]interface{}{
		"ingest_type": "GMetaEntry",
		"ingest_data": entry,
	}
}

// NewGMetaListIngest builds a bulk ingest document with the required
// {ingest_type, ingest_data:{gmeta:[...]}} envelope, suitable for Ingest.
func NewGMetaListIngest(entries []GMetaEntryDocument) map[string]interface{} {
	return map[string]interface{}{
		"ingest_type": "GMetaList",
		"ingest_data": map[string]interface{}{"gmeta": entries},
	}
}

// IngestResponse represents the response from an ingest/delete task submission.
type IngestResponse struct {
	TaskID       string `json:"task_id"`
	Acknowledged bool   `json:"acknowledged,omitempty"`
	Success      bool   `json:"success,omitempty"`
	Message      string `json:"message,omitempty"`
}

// Role represents a role assignment in a search index.
type Role struct {
	ID        string    `json:"id"`
	IndexID   string    `json:"index_id"`
	Principal string    `json:"principal"`
	RoleName  string    `json:"role_name,omitempty"`
	Created   time.Time `json:"created,omitempty"`
}

// RoleCreate is the create-role body: role_name (owner|admin|writer) + principal.
type RoleCreate struct {
	RoleName  string `json:"role_name"`
	Principal string `json:"principal"`
}

// RoleList represents a list of roles.
type RoleList struct {
	Roles []Role `json:"role_list"`
}

// IndexList is a list of search indexes. index_list is not paginated upstream.
type IndexList struct {
	Indexes []Index `json:"index_list"`
}

// ListIndexesOptions controls which indexes are returned. FilterRoles is
// comma-joined into a single filter_roles query param.
type ListIndexesOptions struct {
	FilterRoles []string
}

// Task represents a Search task (get_task / task_list entry). The status field
// is "state" upstream.
type Task struct {
	TaskID    string    `json:"task_id"`
	IndexID   string    `json:"index_id,omitempty"`
	State     string    `json:"state,omitempty"`
	Created   time.Time `json:"creation_date,omitempty"`
	Completed time.Time `json:"completion_date,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// TaskList is the response envelope for GET /task_list/{index_id}.
type TaskList struct {
	Tasks []Task `json:"tasks"`
}
