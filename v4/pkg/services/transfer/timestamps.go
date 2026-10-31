// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// transferTimeLayouts are the timestamp shapes Globus Transfer is known to
// emit. Task fields use RFC 3339, but operation/ls and operation/stat return
// last_modified with a space separator ("2023-12-18 16:52:50+00:00"), and v2
// tunnel events return zone-less microsecond times
// ("2026-02-12T21:59:01.857473"). Zone-less values are interpreted as UTC.
var transferTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05.999999999",
}

// parseTransferTime parses a Transfer API timestamp in any known layout.
// An empty string yields the zero time.
func parseTransferTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range transferTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("transfer: unrecognized timestamp %q", s)
}

// UnmarshalJSON decodes a directory entry, accepting the space-separated
// last_modified form Transfer returns (not valid RFC 3339, so the default
// time.Time decoder rejects the whole listing). A null or missing
// last_modified leaves LastModified as the zero time.
func (e *DirectoryEntry) UnmarshalJSON(data []byte) error {
	type plain DirectoryEntry // drops this method to avoid recursion
	aux := struct {
		*plain
		LastModified *string `json:"last_modified"`
	}{plain: (*plain)(e)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.LastModified = time.Time{}
	if aux.LastModified != nil {
		t, err := parseTransferTime(*aux.LastModified)
		if err != nil {
			return err
		}
		e.LastModified = t
	}
	return nil
}

// jsonAPIID is a JSON:API resource id. The spec says ids are strings, but
// Transfer's /v2/tunnels/{id}/events returns integer ids (e.g. 2768), so
// accept either and normalise to the decimal string form.
type jsonAPIID string

func (id *jsonAPIID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = ""
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*id = jsonAPIID(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("transfer: JSON:API id must be a string or number: %w", err)
	}
	*id = jsonAPIID(n.String())
	return nil
}
