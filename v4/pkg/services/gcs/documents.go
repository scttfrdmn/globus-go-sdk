// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import "encoding/json"

// This file holds the request-document builders for the GCS Manager API. Each is
// posted/patched as the request body directly (no envelope). DataType is
// omitempty; when unset the service assumes the base "#1.0.0" version. Nullable
// fields use pointers so an explicit null can be distinguished from omitted.

// CollectionDocument is the create/update body for a mapped or guest collection.
type CollectionDocument struct {
	DataType       string `json:"DATA_TYPE,omitempty"`
	CollectionType string `json:"collection_type,omitempty"` // "mapped" | "guest"

	CollectionBasePath              string          `json:"collection_base_path,omitempty"`
	ContactEmail                    *string         `json:"contact_email,omitempty"`
	ContactInfo                     *string         `json:"contact_info,omitempty"`
	DefaultDirectory                string          `json:"default_directory,omitempty"`
	Department                      *string         `json:"department,omitempty"`
	Description                     *string         `json:"description,omitempty"`
	DisplayName                     string          `json:"display_name,omitempty"`
	IdentityID                      string          `json:"identity_id,omitempty"`
	InfoLink                        *string         `json:"info_link,omitempty"`
	Organization                    string          `json:"organization,omitempty"`
	RestrictTransfersToHighAssurance string         `json:"restrict_transfers_to_high_assurance,omitempty"`
	UserMessage                     *string         `json:"user_message,omitempty"`
	UserMessageLink                 *string         `json:"user_message_link,omitempty"`
	Keywords                        []string        `json:"keywords,omitempty"`
	DisableVerify                   *bool           `json:"disable_verify,omitempty"`
	EnableHTTPS                     *bool           `json:"enable_https,omitempty"`
	ForceEncryption                 *bool           `json:"force_encryption,omitempty"`
	ForceVerify                     *bool           `json:"force_verify,omitempty"`
	Public                          *bool           `json:"public,omitempty"`
	ACLExpirationMins               *int            `json:"acl_expiration_mins,omitempty"`
	AssociatedFlowPolicy            json.RawMessage `json:"associated_flow_policy,omitempty"`

	// Mapped-only.
	DomainName             string          `json:"domain_name,omitempty"`
	GuestAuthPolicyID      *string         `json:"guest_auth_policy_id,omitempty"`
	StorageGatewayID       string          `json:"storage_gateway_id,omitempty"`
	SharingUsersAllow      []string        `json:"sharing_users_allow,omitempty"`
	SharingUsersDeny       []string        `json:"sharing_users_deny,omitempty"`
	SharingRestrictPaths   json.RawMessage `json:"sharing_restrict_paths,omitempty"`
	DeleteProtected        *bool           `json:"delete_protected,omitempty"`
	AllowGuestCollections  *bool           `json:"allow_guest_collections,omitempty"`
	DisableAnonymousWrites *bool           `json:"disable_anonymous_writes,omitempty"`
	AutoDeleteTimeout      *int            `json:"auto_delete_timeout,omitempty"`
	Policies               json.RawMessage `json:"policies,omitempty"`

	// Guest-only.
	MappedCollectionID         string          `json:"mapped_collection_id,omitempty"`
	UserCredentialID           string          `json:"user_credential_id,omitempty"`
	SkipAutoDelete             *bool           `json:"skip_auto_delete,omitempty"`
	ActivityNotificationPolicy json.RawMessage `json:"activity_notification_policy,omitempty"`
}

// EndpointDocument is the update body for the endpoint.
type EndpointDocument struct {
	DataType                  string   `json:"DATA_TYPE,omitempty"`
	ContactEmail              string   `json:"contact_email,omitempty"`
	ContactInfo               string   `json:"contact_info,omitempty"`
	Department                string   `json:"department,omitempty"`
	Description               string   `json:"description,omitempty"`
	DisplayName               string   `json:"display_name,omitempty"`
	InfoLink                  string   `json:"info_link,omitempty"`
	NetworkUse                string   `json:"network_use,omitempty"` // normal|minimal|aggressive|custom
	Organization              string   `json:"organization,omitempty"`
	SubscriptionID            *string  `json:"subscription_id,omitempty"`
	Keywords                  []string `json:"keywords,omitempty"`
	AllowUDT                  *bool    `json:"allow_udt,omitempty"`
	Public                    *bool    `json:"public,omitempty"`
	MaxConcurrency            *int     `json:"max_concurrency,omitempty"`
	MaxParallelism            *int     `json:"max_parallelism,omitempty"`
	PreferredConcurrency      *int     `json:"preferred_concurrency,omitempty"`
	PreferredParallelism      *int     `json:"preferred_parallelism,omitempty"`
	GridFTPControlChannelPort *int     `json:"gridftp_control_channel_port,omitempty"`
}

// StorageGatewayDocument is the create/update body for a storage gateway.
type StorageGatewayDocument struct {
	DataType                  string            `json:"DATA_TYPE,omitempty"`
	DisplayName               string            `json:"display_name,omitempty"`
	ConnectorID               string            `json:"connector_id,omitempty"`
	Root                      string            `json:"root,omitempty"`
	IdentityMappings          []json.RawMessage `json:"identity_mappings,omitempty"`
	Policies                  json.RawMessage   `json:"policies,omitempty"`
	AllowedDomains            []string          `json:"allowed_domains,omitempty"`
	HighAssurance             *bool             `json:"high_assurance,omitempty"`
	RequireMFA                *bool             `json:"require_mfa,omitempty"`
	AuthenticationTimeoutMins *int              `json:"authentication_timeout_mins,omitempty"`
	UsersAllow                []string          `json:"users_allow,omitempty"`
	UsersDeny                 []string          `json:"users_deny,omitempty"`
}

// GCSRoleDocument is the create body for a role. Collection is omitted for
// endpoint-level roles. Role is one of owner, administrator, access_manager,
// activity_manager, activity_monitor.
type GCSRoleDocument struct {
	DataType   string `json:"DATA_TYPE,omitempty"`
	Collection string `json:"collection,omitempty"`
	Principal  string `json:"principal"`
	Role       string `json:"role"`
}

// UserCredentialDocument is the create/update body for a user credential.
type UserCredentialDocument struct {
	DataType         string          `json:"DATA_TYPE,omitempty"`
	IdentityID       string          `json:"identity_id,omitempty"`
	ConnectorID      string          `json:"connector_id,omitempty"`
	Username         string          `json:"username,omitempty"`
	DisplayName      string          `json:"display_name,omitempty"`
	StorageGatewayID string          `json:"storage_gateway_id,omitempty"`
	Policies         json.RawMessage `json:"policies,omitempty"`
}
