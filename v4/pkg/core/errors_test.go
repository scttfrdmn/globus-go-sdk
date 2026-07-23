// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// mkResp builds a minimal *http.Response with a drained-equivalent body, to
// mirror the real call path where the caller has already read resp.Body before
// calling NewAPIError.
func mkResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		// Body is intentionally an empty reader: NewAPIError must parse the
		// message argument, not re-read this (issue #63).
		Body: io.NopCloser(strings.NewReader("")),
	}
}

// TestNewAPIErrorPopulatesDetails is the regression test for issue #63:
// NewAPIError must populate Details from the body passed as message, rather than
// re-reading the already-drained resp.Body (which yielded io.EOF and left
// Details nil — breaking the session_required_policies retry in globus-go-cli).
func TestNewAPIErrorPopulatesDetails(t *testing.T) {
	body := `{
		"code": "FORBIDDEN",
		"message": "high assurance required",
		"authorization_parameters": {
			"session_required_policies": ["34285468-cba9-4615-a719-eff66d292409"]
		}
	}`
	resp := mkResp(http.StatusForbidden, body)

	apiErr := NewAPIError(resp, body)

	if apiErr.Details == nil {
		t.Fatal("Details is nil — body was not parsed (issue #63 regression)")
	}
	if apiErr.Code != "FORBIDDEN" {
		t.Errorf("Code = %q, want FORBIDDEN", apiErr.Code)
	}
	if apiErr.Message != "high assurance required" {
		t.Errorf("Message = %q, want the parsed message", apiErr.Message)
	}

	ap, ok := apiErr.Details["authorization_parameters"].(map[string]interface{})
	if !ok {
		t.Fatalf("authorization_parameters missing from Details: %#v", apiErr.Details)
	}
	pols, ok := ap["session_required_policies"].([]interface{})
	if !ok || len(pols) != 1 || pols[0] != "34285468-cba9-4615-a719-eff66d292409" {
		t.Errorf("session_required_policies not extracted: %#v", ap["session_required_policies"])
	}
}

// TestNewAPIErrorNestedErrors covers the JSON:API-style shape Globus Auth uses,
// where code/message live under errors[0] rather than at the top level (issue
// #63 fix #3).
func TestNewAPIErrorNestedErrors(t *testing.T) {
	body := `{"errors":[{"code":"FORBIDDEN","message":"nested detail"}]}`
	resp := mkResp(http.StatusForbidden, body)

	apiErr := NewAPIError(resp, body)

	if apiErr.Code != "FORBIDDEN" {
		t.Errorf("Code = %q, want FORBIDDEN from errors[0]", apiErr.Code)
	}
	if apiErr.Message != "nested detail" {
		t.Errorf("Message = %q, want the nested message", apiErr.Message)
	}
	if apiErr.Details == nil {
		t.Error("Details should be populated for nested-error bodies")
	}
}

// TestNewAPIErrorNonJSONBody verifies a non-JSON body is preserved verbatim as
// Message and leaves Details nil (no panic, no garbage).
func TestNewAPIErrorNonJSONBody(t *testing.T) {
	body := "502 Bad Gateway"
	resp := mkResp(http.StatusBadGateway, body)

	apiErr := NewAPIError(resp, body)

	if apiErr.Message != body {
		t.Errorf("Message = %q, want the raw body preserved", apiErr.Message)
	}
	if apiErr.Details != nil {
		t.Errorf("Details should be nil for a non-JSON body, got %#v", apiErr.Details)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want 502", apiErr.StatusCode)
	}
}
