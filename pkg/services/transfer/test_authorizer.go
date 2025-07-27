// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"net/http"
)

// testAuthorizer2 implements a simple authorizer for testing
// (Using different name to avoid conflict with testAuthorizer in client_test.go)
type testAuthorizer2 struct {
	token string
}

// AddAuthToRequest implements the interfaces.Authorizer interface
func (a *testAuthorizer2) AddAuthToRequest(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.token)
	return nil
}
