// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import "fmt"

// ---- JSON:API base types ----

// JSONAPILinks holds pagination links in a JSON:API response.
type JSONAPILinks struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// JSONAPIMeta holds metadata from a JSON:API response.
type JSONAPIMeta struct {
	Total  int `json:"total,omitempty"`
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

// ---- Collection models ----

// Collection represents a Globus Connect Server collection.
// See https://docs.globus.org/globus-connect-server/v5/api/openapi_Collections/
type Collection struct {
	ID             string `json:"id"`
	DataType       string `json:"DATA_TYPE"`
	CollectionType string `json:"collection_type"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description,omitempty"`

	// OwnerString is the Globus identity URN that owns this collection.
	OwnerString string `json:"owner,omitempty"`

	// PubliclyVisible controls whether the collection appears in public listings.
	PubliclyVisible bool `json:"public,omitempty"`

	// AllowGuests controls guest collection creation from this collection.
	AllowGuests bool `json:"allow_guest_collections,omitempty"`

	// MappedCollectionID is set for guest collections to identify their parent.
	MappedCollectionID string `json:"mapped_collection_id,omitempty"`

	// StorageGatewayID identifies the storage gateway backing this collection.
	StorageGatewayID string `json:"storage_gateway_id,omitempty"`

	// TLSPort is the HTTPS port for data-plane access.
	TLSPort int `json:"tlsport,omitempty"`

	// HTTPSURL is the base URL for HTTPS data-plane transfers.
	HTTPSURL string `json:"https_url,omitempty"`
}

// CollectionPage is a JSON:API paginated list of collections.
type CollectionPage struct {
	Data  []Collection `json:"data"`
	Links JSONAPILinks `json:"links,omitempty"`
	Meta  JSONAPIMeta  `json:"meta,omitempty"`
}

// CollectionUpdate contains fields that can be changed on an existing collection.
type CollectionUpdate struct {
	DisplayName     string `json:"display_name,omitempty"`
	Description     string `json:"description,omitempty"`
	PubliclyVisible *bool  `json:"public,omitempty"`
	AllowGuests     *bool  `json:"allow_guest_collections,omitempty"`
}

// ListCollectionsOptions controls which collections are returned.
type ListCollectionsOptions struct {
	// FilterOwned restricts results to collections owned by the caller.
	FilterOwned bool
	// MappedCollectionID limits results to guest collections of this parent.
	MappedCollectionID string
	// Limit and Offset provide page control.
	Limit  int
	Offset int
}

// ---- Scope helpers ----

// CollectionScopes returns the standard Globus scope strings for a collection.
// HTTPS scope is required for data-plane file access.
// DataAccess scope is required for transfer task submission to the collection.
func CollectionScopes(collectionID string) (https, dataAccess string) {
	https = fmt.Sprintf("https://auth.globus.org/scopes/%s/https", collectionID)
	dataAccess = fmt.Sprintf("https://auth.globus.org/scopes/%s/data_access", collectionID)
	return
}
