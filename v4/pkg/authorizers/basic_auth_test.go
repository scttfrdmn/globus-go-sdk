// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package authorizers_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
)

func TestBasicAuthAuthorizer(t *testing.T) {
	a := authorizers.NewBasicAuthAuthorizer("client-id", "secret")
	got, err := a.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("GetAuthorizationHeader: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id:secret"))
	if got != want {
		t.Errorf("header = %q, want %q", got, want)
	}
	if a.HandleMissingAuthorization(context.Background()) {
		t.Error("HandleMissingAuthorization should be false for static client credentials")
	}
}

func TestBasicAuthAuthorizer_EmptyClientID(t *testing.T) {
	a := authorizers.NewBasicAuthAuthorizer("", "secret")
	if _, err := a.GetAuthorizationHeader(context.Background()); err == nil {
		t.Error("expected error for empty client ID")
	}
}
