// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"time"
)

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
	ID            jsonAPIID                 `json:"id"`
	Attributes    map[string]interface{}    `json:"attributes"`
	Relationships map[string]jsonAPIRelResp `json:"relationships"`
}

// jsonAPIRelResp is a decoded relationship. Data stays raw so null and
// to-many relationships never fail the whole response; only to-one
// {"type","id"} objects are surfaced.
type jsonAPIRelResp struct {
	Data json.RawMessage `json:"data"`
}

// JSONAPIRelationship is a to-one JSON:API relationship target.
type JSONAPIRelationship struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// toOneRelationships returns the to-one relationships of res, keyed by name.
func toOneRelationships(res jsonAPIResourceData) map[string]JSONAPIRelationship {
	out := map[string]JSONAPIRelationship{}
	for name, rel := range res.Relationships {
		var ref struct {
			Type string    `json:"type"`
			ID   jsonAPIID `json:"id"`
		}
		if len(rel.Data) == 0 || rel.Data[0] != '{' || json.Unmarshal(rel.Data, &ref) != nil || ref.ID == "" {
			continue
		}
		out[name] = JSONAPIRelationship{Type: ref.Type, ID: string(ref.ID)}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// attrInt returns attrs[key] as an int (JSON numbers decode as float64), or
// nil when absent or null.
func attrInt(attrs map[string]interface{}, key string) *int {
	if v, ok := attrs[key].(float64); ok {
		n := int(v)
		return &n
	}
	return nil
}

// attrTime parses a Transfer timestamp attribute (zone-less values are UTC),
// or returns nil when absent, null or unparseable.
func attrTime(attrs map[string]interface{}, key string) *time.Time {
	s := attrString(attrs, key)
	if s == "" {
		return nil
	}
	t, err := parseTransferTime(s)
	if err != nil {
		return nil
	}
	return &t
}

func attrString(attrs map[string]interface{}, key string) string {
	if v, ok := attrs[key].(string); ok {
		return v
	}
	return ""
}

func flattenTunnel(res jsonAPIResourceData) Tunnel {
	a := res.Attributes
	rels := toOneRelationships(res)
	label := attrString(a, "label")
	t := Tunnel{
		ID:                 string(res.ID),
		Label:              label,
		DisplayName:        label,
		State:              attrString(a, "state"),
		Status:             attrString(a, "status"),
		ListenerIPAddress:  attrString(a, "listener_ip_address"),
		ListenerPort:       attrInt(a, "listener_port"),
		InitiatorIPAddress: attrString(a, "initiator_ip_address"),
		InitiatorPort:      attrInt(a, "initiator_port"),
		LifetimeMins:       attrInt(a, "lifetime_mins"),
		SubmissionID:       attrString(a, "submission_id"),
		CreatedAt:          attrTime(a, "created_time"),
		Relationships:      rels,
		Metadata:           a,
	}
	if v, ok := a["restartable"].(bool); ok {
		t.Restartable = v
	}
	t.ListenerAccessPointID = rels["listener"].ID
	t.InitiatorAccessPointID = rels["initiator"].ID
	t.Owner = rels["owner"].ID
	return t
}

func flattenSAP(res jsonAPIResourceData) StreamAccessPoint {
	a := res.Attributes
	rels := toOneRelationships(res)
	ap := StreamAccessPoint{
		ID:              string(res.ID),
		DisplayName:     attrString(a, "display_name"),
		AdvertisedOwner: attrString(a, "advertised_owner"),
		ContactEmail:    attrString(a, "contact_email"),
		ContactInfo:     attrString(a, "contact_info"),
		Department:      attrString(a, "department"),
		Description:     attrString(a, "description"),
		InfoLink:        attrString(a, "info_link"),
		Organization:    attrString(a, "organization"),
		TLSFTPServer:    attrString(a, "tlsftp_server"),
		Relationships:   rels,
		Metadata:        a,
	}
	ap.Keywords = attrKeywords(a, "keywords")
	ap.EndpointID = rels["host_endpoint"].ID
	return ap
}

// attrKeywords accepts keywords as a JSON array or a comma-separated string
// (the same two forms Endpoint keywords arrive in); nil when absent or null.
func attrKeywords(attrs map[string]interface{}, key string) Keywords {
	raw, ok := attrs[key]
	if !ok || raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var k Keywords
	if json.Unmarshal(b, &k) != nil {
		return nil
	}
	return k
}

// setIf sets key=val in m only when val is non-empty.
func setIf(m map[string]interface{}, key, val string) {
	if val != "" {
		m[key] = val
	}
}
