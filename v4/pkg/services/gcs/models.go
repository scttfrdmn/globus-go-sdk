// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GCSResponse is the generic Globus Connect Server "result#1.0.0" envelope. Every
// GCS Manager response is wrapped in this shape; single-object GETs return a
// one-element Data array, lists return the whole array.
type GCSResponse struct {
	DataType         string            `json:"DATA_TYPE"`
	Code             string            `json:"code"`
	Detail           json.RawMessage   `json:"detail,omitempty"`
	Message          string            `json:"message,omitempty"`
	HTTPResponseCode int               `json:"http_response_code,omitempty"`
	Data             []json.RawMessage `json:"data"`
	HasNextPage      bool              `json:"has_next_page"`
	Marker           string            `json:"marker,omitempty"`
}

// unpack finds the first Data element whose DATA_TYPE name (the part before '#')
// matches want, and unmarshals it into target.
func (r *GCSResponse) unpack(want string, target interface{}) error {
	for _, raw := range r.Data {
		var probe struct {
			DataType string `json:"DATA_TYPE"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil {
			continue
		}
		if datatypeName(probe.DataType) == want {
			return json.Unmarshal(raw, target)
		}
	}
	return fmt.Errorf("gcs: no %q object found in response data", want)
}

// datatypeName returns the name portion of a DATA_TYPE string ("collection#1.0.0"
// -> "collection").
func datatypeName(dt string) string {
	if i := strings.IndexByte(dt, '#'); i >= 0 {
		return dt[:i]
	}
	return dt
}

// ---- Collection ----

// Collection represents a Globus Connect Server collection.
type Collection struct {
	ID                        string          `json:"id"`
	DataType                  string          `json:"DATA_TYPE"`
	CollectionType            string          `json:"collection_type"`
	DisplayName               string          `json:"display_name"`
	Description               string          `json:"description,omitempty"`
	IdentityID                string          `json:"identity_id,omitempty"`
	PubliclyVisible           bool            `json:"public,omitempty"`
	AllowGuests               bool            `json:"allow_guest_collections,omitempty"`
	MappedCollectionID        string          `json:"mapped_collection_id,omitempty"`
	StorageGatewayID          string          `json:"storage_gateway_id,omitempty"`
	UserCredentialID          string          `json:"user_credential_id,omitempty"`
	ConnectorID               string          `json:"connector_id,omitempty"`
	CollectionBasePath        string          `json:"collection_base_path,omitempty"`
	RootPath                  string          `json:"root_path,omitempty"`
	DefaultDirectory          string          `json:"default_directory,omitempty"`
	DomainName                string          `json:"domain_name,omitempty"`
	ManagerURL                string          `json:"manager_url,omitempty"`
	HTTPSURL                  string          `json:"https_url,omitempty"`
	TLSFTPURL                 string          `json:"tlsftp_url,omitempty"`
	HighAssurance             bool            `json:"high_assurance,omitempty"`
	ForceEncryption           bool            `json:"force_encryption,omitempty"`
	DisableVerify             bool            `json:"disable_verify,omitempty"`
	Organization              string          `json:"organization,omitempty"`
	Department                string          `json:"department,omitempty"`
	Keywords                  []string        `json:"keywords,omitempty"`
	ContactEmail              string          `json:"contact_email,omitempty"`
	ContactInfo               string          `json:"contact_info,omitempty"`
	InfoLink                  string          `json:"info_link,omitempty"`
	AuthenticationTimeoutMins int             `json:"authentication_timeout_mins,omitempty"`
	Policies                  json.RawMessage `json:"policies,omitempty"`
	SharingRestrictPaths      json.RawMessage `json:"sharing_restrict_paths,omitempty"`
}

// CollectionListResponse is the marker-paginated list of collections.
type CollectionListResponse struct {
	Data        []Collection `json:"data"`
	HasNextPage bool         `json:"has_next_page"`
	Marker      string       `json:"marker,omitempty"`
	DataType    string       `json:"DATA_TYPE"`
	Code        string       `json:"code"`
}

// ListCollectionsOptions controls which collections are returned. Filter and
// Include are comma-joined into single query params.
type ListCollectionsOptions struct {
	MappedCollectionID string
	Filter             []string
	Include            []string
	PageSize           int
	Marker             string
}

// GetCollectionOptions carries optional passthrough query params for GetCollection.
type GetCollectionOptions struct {
	QueryParams map[string]string
}

// ---- Endpoint ----

// GCSInfo is the GET /info document (unauthenticated).
type GCSInfo struct {
	DataType   string `json:"DATA_TYPE"`
	ClientID   string `json:"client_id"`
	EndpointID string `json:"endpoint_id,omitempty"`
}

// Endpoint is the GET/PATCH /endpoint document.
type Endpoint struct {
	DataType       string   `json:"DATA_TYPE"`
	ID             string   `json:"id"`
	DisplayName    string   `json:"display_name,omitempty"`
	GCSManagerURL  string   `json:"gcs_manager_url,omitempty"`
	Organization   string   `json:"organization,omitempty"`
	Description    string   `json:"description,omitempty"`
	ContactEmail   string   `json:"contact_email,omitempty"`
	NetworkUse     string   `json:"network_use,omitempty"`
	Public         bool     `json:"public,omitempty"`
	Keywords       []string `json:"keywords,omitempty"`
	SubscriptionID string   `json:"subscription_id,omitempty"`
}

// UpdateEndpointOptions carries the comma-joined include param for UpdateEndpoint.
type UpdateEndpointOptions struct {
	Include []string
}

// ---- Storage gateway ----

// StorageGateway is a storage gateway document.
type StorageGateway struct {
	DataType                  string          `json:"DATA_TYPE"`
	ID                        string          `json:"id"`
	DisplayName               string          `json:"display_name"`
	ConnectorID               string          `json:"connector_id"`
	HighAssurance             bool            `json:"high_assurance,omitempty"`
	RequireMFA                bool            `json:"require_mfa,omitempty"`
	AllowedDomains            []string        `json:"allowed_domains,omitempty"`
	AuthenticationTimeoutMins int             `json:"authentication_timeout_mins,omitempty"`
	Policies                  json.RawMessage `json:"policies,omitempty"`
}

// StorageGatewayListResponse is the marker-paginated list of storage gateways.
type StorageGatewayListResponse struct {
	Data        []StorageGateway `json:"data"`
	HasNextPage bool             `json:"has_next_page"`
	Marker      string           `json:"marker,omitempty"`
	DataType    string           `json:"DATA_TYPE"`
	Code        string           `json:"code"`
}

// StorageGatewayListOptions controls GetStorageGatewayList. Include is comma-joined.
type StorageGatewayListOptions struct {
	Include  []string
	PageSize int
	Marker   string
}

// GetStorageGatewayOptions carries the comma-joined include param.
type GetStorageGatewayOptions struct {
	Include []string
}

// ---- Role ----

// GCSRole is a role document.
type GCSRole struct {
	DataType   string  `json:"DATA_TYPE"`
	ID         string  `json:"id"`
	Collection *string `json:"collection"`
	Principal  string  `json:"principal"`
	Role       string  `json:"role"`
}

// RoleListResponse is the marker-paginated list of roles.
type RoleListResponse struct {
	Data        []GCSRole `json:"data"`
	HasNextPage bool      `json:"has_next_page"`
	Marker      string    `json:"marker,omitempty"`
	DataType    string    `json:"DATA_TYPE"`
	Code        string    `json:"code"`
}

// RoleListOptions controls GetRoleList. Include is passed raw (not comma-joined).
type RoleListOptions struct {
	CollectionID string
	Include      string
	PageSize     int
	Marker       string
}

// ---- User credential ----

// UserCredential is a user credential document.
type UserCredential struct {
	DataType         string          `json:"DATA_TYPE"`
	ID               string          `json:"id"`
	ConnectorID      string          `json:"connector_id"`
	IdentityID       string          `json:"identity_id"`
	StorageGatewayID string          `json:"storage_gateway_id"`
	Username         string          `json:"username"`
	DisplayName      string          `json:"display_name,omitempty"`
	Invalid          bool            `json:"invalid,omitempty"`
	Provisioned      bool            `json:"provisioned,omitempty"`
	Policies         json.RawMessage `json:"policies,omitempty"`
}

// UserCredentialListResponse is the marker-paginated list of user credentials.
type UserCredentialListResponse struct {
	Data        []UserCredential `json:"data"`
	HasNextPage bool             `json:"has_next_page"`
	Marker      string           `json:"marker,omitempty"`
	DataType    string           `json:"DATA_TYPE"`
	Code        string           `json:"code"`
}

// UserCredentialListOptions controls GetUserCredentialList.
type UserCredentialListOptions struct {
	StorageGateway string
	PageSize       int
	Marker         string
}

// ---- Scope helpers ----

// CollectionScopes returns the standard Globus scope strings for a collection.
// HTTPS scope is required for data-plane file access.
// DataAccess scope is required for transfer task submission to the collection.
//
// These are "collection" scopes in URL format, keyed by the collection ID (the
// data-access resource server). See globus-sdk-python GCSCollectionScopes.
func CollectionScopes(collectionID string) (https, dataAccess string) {
	https = fmt.Sprintf("https://auth.globus.org/scopes/%s/https", collectionID)
	dataAccess = fmt.Sprintf("https://auth.globus.org/scopes/%s/data_access", collectionID)
	return
}

// EndpointManageCollectionsScope returns the GCS endpoint's manage_collections
// scope, required for management operations against the GCS Manager API
// (listing/creating/deleting collections, storage gateways, roles, and user
// credentials). Unlike the collection data-plane scopes, this is an endpoint
// scope in Globus Auth URN format, keyed by the endpoint ID. See
// globus-sdk-python GCSEndpointScopes.
func EndpointManageCollectionsScope(endpointID string) string {
	return fmt.Sprintf("urn:globus:auth:scope:%s:manage_collections", endpointID)
}
