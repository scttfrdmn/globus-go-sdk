// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"testing"
)

// TestKeywordsUnmarshal covers both wire forms the Transfer API uses for the
// endpoint "keywords" field: a JSON array and a comma-separated string. The
// string form was found by dogfooding against live Globus (endpoint search),
// where the []string model failed to unmarshal.
func TestKeywordsUnmarshal(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"array", `["a","b","c"]`, []string{"a", "b", "c"}},
		{"comma string", `"Globus,demo,tutorial"`, []string{"Globus", "demo", "tutorial"}},
		{"string with spaces", `"a, b , c"`, []string{"a", "b", "c"}},
		{"empty string", `""`, nil},
		{"empty array", `[]`, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var k Keywords
			if err := json.Unmarshal([]byte(tc.in), &k); err != nil {
				t.Fatalf("Unmarshal(%s): %v", tc.in, err)
			}
			if len(k) != len(tc.want) {
				t.Fatalf("len = %d, want %d (%v)", len(k), len(tc.want), k)
			}
			for i := range tc.want {
				if k[i] != tc.want[i] {
					t.Errorf("[%d] = %q, want %q", i, k[i], tc.want[i])
				}
			}
		})
	}
}

// TestEndpointDecodesStringKeywords is a regression test for the live-data shape
// that broke `globus transfer endpoint search`: an endpoint whose keywords are a
// bare comma-separated string.
func TestEndpointDecodesStringKeywords(t *testing.T) {
	raw := `{"id":"ep-1","display_name":"Tutorial","keywords":"Globus,demo,tutorial,1"}`
	var ep Endpoint
	if err := json.Unmarshal([]byte(raw), &ep); err != nil {
		t.Fatalf("Unmarshal endpoint: %v", err)
	}
	if len(ep.Keywords) != 4 || ep.Keywords[0] != "Globus" {
		t.Errorf("keywords = %v, want [Globus demo tutorial 1]", ep.Keywords)
	}
}
