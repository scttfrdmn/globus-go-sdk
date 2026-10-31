// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"testing"
	"time"
)

// TestDirectoryEntryLastModified covers issue #65: operation/ls returns
// last_modified with a space separator, which time.Time's RFC 3339 decoder
// rejects, failing the whole listing. Payload is the one from the report.
func TestDirectoryEntryLastModified(t *testing.T) {
	const entry = `{
		"DATA_TYPE": "file", "group": "XXXX\\YYYYY",
		"last_modified": "2026-08-12 19:08:37+00:00",
		"link_group": null, "link_last_modified": null, "link_size": null,
		"link_target": null, "link_user": null,
		"name": "a.bin", "permissions": "0644", "size": 50,
		"type": "file", "user": "XXXXXX"
	}`
	var listing DirectoryListing
	if err := json.Unmarshal([]byte(`{"DATA_TYPE":"file_list","DATA":[`+entry+`]}`), &listing); err != nil {
		t.Fatalf("decode listing: %v", err)
	}
	if len(listing.Data) != 1 {
		t.Fatalf("got %d entries", len(listing.Data))
	}
	e := listing.Data[0]
	want := time.Date(2026, 8, 12, 19, 8, 37, 0, time.UTC)
	if !e.LastModified.Equal(want) {
		t.Errorf("LastModified = %v, want %v", e.LastModified, want)
	}
	if e.Name != "a.bin" || e.Size != 50 || e.Permissions != "0644" || e.Group != `XXXX\YYYYY` || e.User != "XXXXXX" {
		t.Errorf("other fields not decoded: %+v", e)
	}
}

func TestDirectoryEntryLastModifiedVariants(t *testing.T) {
	cases := map[string]time.Time{
		`"2026-08-12T19:08:37+00:00"`:        time.Date(2026, 8, 12, 19, 8, 37, 0, time.UTC),
		`"2026-08-12 19:08:37-05:00"`:        time.Date(2026, 8, 13, 0, 8, 37, 0, time.UTC),
		`"2026-08-12 19:08:37.250000+00:00"`: time.Date(2026, 8, 12, 19, 8, 37, 250e6, time.UTC),
		`null`:                               {},
		`""`:                                 {},
	}
	for raw, want := range cases {
		var e DirectoryEntry
		if err := json.Unmarshal([]byte(`{"name":"x","last_modified":`+raw+`}`), &e); err != nil {
			t.Errorf("%s: %v", raw, err)
			continue
		}
		if !e.LastModified.Equal(want) {
			t.Errorf("%s: got %v, want %v", raw, e.LastModified, want)
		}
	}
	var e DirectoryEntry
	if err := json.Unmarshal([]byte(`{"last_modified":"yesterday"}`), &e); err == nil {
		t.Error("expected error for unparseable timestamp")
	}
}

func TestParseTransferTimeZoneless(t *testing.T) {
	got, err := parseTransferTime("2026-02-12T21:59:01.857473")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 2, 12, 21, 59, 1, 857473000, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestJSONAPIIDAcceptsNumbersAndStrings(t *testing.T) {
	var doc jsonAPICollection
	if err := json.Unmarshal([]byte(`{"data":[{"type":"TunnelEvent","id":2768},{"type":"Tunnel","id":"abc"},{"type":"X","id":null}]}`), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Data[0].ID != "2768" || doc.Data[1].ID != "abc" || doc.Data[2].ID != "" {
		t.Errorf("ids = %q %q %q", doc.Data[0].ID, doc.Data[1].ID, doc.Data[2].ID)
	}
}
