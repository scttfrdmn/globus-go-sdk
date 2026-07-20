// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

// JSON:API decode helpers for the Beta tunnel/stream_access_point surface. The
// wire shape is {"data": {"type","id","attributes":{...},"relationships":{...}}}
// for single resources and {"data": [ ... ]} for collections. Callers work with
// the flat Tunnel / StreamAccessPoint types; these helpers flatten the envelope.

type jsonAPIResource struct {
	Data jsonAPIResourceData `json:"data"`
}

type jsonAPICollection struct {
	Data []jsonAPIResourceData `json:"data"`
}

type jsonAPIResourceData struct {
	Type          string                    `json:"type"`
	ID            string                    `json:"id"`
	Attributes    map[string]interface{}    `json:"attributes"`
	Relationships map[string]jsonAPIRelData `json:"-"`
}

func attrString(attrs map[string]interface{}, key string) string {
	if v, ok := attrs[key].(string); ok {
		return v
	}
	return ""
}

func flattenTunnel(res jsonAPIResourceData) Tunnel {
	return Tunnel{
		ID:          res.ID,
		DisplayName: attrString(res.Attributes, "label"),
		Status:      attrString(res.Attributes, "status"),
		Metadata:    res.Attributes,
	}
}

func flattenSAP(res jsonAPIResourceData) StreamAccessPoint {
	return StreamAccessPoint{
		ID:       res.ID,
		Metadata: res.Attributes,
	}
}

// setIf sets key=val in m only when val is non-empty.
func setIf(m map[string]interface{}, key, val string) {
	if val != "" {
		m[key] = val
	}
}
