// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"encoding/base64"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// BasicAuthAuthorizer authenticates as an OAuth2 confidential client using HTTP
// Basic auth (client_id:client_secret). This is the mechanism the token
// introspection (RFC 7662) and dependent-token endpoints expect for client
// authentication — they authenticate the client, not a user bearer token. It
// implements core.Authorizer.
type BasicAuthAuthorizer struct {
	clientID     string
	clientSecret string
}

// NewBasicAuthAuthorizer creates an authorizer that sends
// "Basic base64(clientID:clientSecret)".
func NewBasicAuthAuthorizer(clientID, clientSecret string) core.Authorizer {
	return &BasicAuthAuthorizer{clientID: clientID, clientSecret: clientSecret}
}

// GetAuthorizationHeader returns the Basic authorization header.
func (a *BasicAuthAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	if a.clientID == "" {
		return "", &core.ValidationError{Field: "clientID", Message: "client ID is empty"}
	}
	creds := a.clientID + ":" + a.clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds)), nil
}

// HandleMissingAuthorization always returns false — client credentials are
// static and cannot be refreshed.
func (a *BasicAuthAuthorizer) HandleMissingAuthorization(_ context.Context) bool {
	return false
}
