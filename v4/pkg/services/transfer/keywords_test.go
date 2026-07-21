// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"testing"
)

// TestKeywordsUnmarshal covers both wire forms of the endpoint "keywords" field
// (array and comma-separated string). The string form was found by dogfooding
// against live Globus.
func TestKeywordsUnmarshal(t *testing.T) {
	var k Keywords
	if err := json.Unmarshal([]byte(`"Globus,demo,tutorial"`), &k); err != nil {
		t.Fatalf("string form: %v", err)
	}
	if len(k) != 3 || k[0] != "Globus" {
		t.Errorf("string form = %v", k)
	}
	if err := json.Unmarshal([]byte(`["a","b"]`), &k); err != nil {
		t.Fatalf("array form: %v", err)
	}
	if len(k) != 2 || k[1] != "b" {
		t.Errorf("array form = %v", k)
	}
}
