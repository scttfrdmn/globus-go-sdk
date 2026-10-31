// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Web inputs — upstream v4.9.0. Payloads mirror the upstream fixtures in
// globus_sdk/testing/data/flows/{list_web_inputs,get_web_input,respond_to_web_input}.py.

const webInputID = "e35d1f92-3e2a-4c1c-8f36-7ce4f00f9d2c"

func rawJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func newFlowsClient(t *testing.T, h http.HandlerFunc) *flows.Client {
	t.Helper()
	server := testhelpers.NewMockServer(t, h)
	c, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestListWebInputs_QueryAndDecode(t *testing.T) {
	c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/web_inputs", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "viewer,respondent", q.Get("filter_roles"))
		assert.Equal(t, "open", q.Get("filter_states"))
		assert.Equal(t, "f1,f2", q.Get("filter_flow_ids"))
		assert.Equal(t, "r1", q.Get("filter_run_ids"))
		// orderby is comma-joined for web inputs (upstream commajoin), unlike
		// registered_apis.
		assert.Equal(t, "created_timestamp DESC,title", q.Get("orderby"))
		assert.Equal(t, "10", q.Get("per_page"))
		assert.Equal(t, "m0", q.Get("marker"))
		rawJSON(w, http.StatusOK, `{"web_input_summaries": [{
			"id": "`+webInputID+`", "status": "open", "user_roles": ["viewer", "respondent"],
			"input_type": "selection", "title": "Approve deployment to production?",
			"flow": {"id": "flow-1", "title": "Multi Step Transfer"},
			"run": {"id": "run-1", "label": "Transfer all of these files!"},
			"created_timestamp": "2026-08-01T10:30:00+00:00",
			"edited_timestamp": "2026-08-01T10:30:00+00:00",
			"closed_timestamp": null}], "marker": null}`)
	})

	list, err := c.ListWebInputs(context.Background(), &flows.ListWebInputsOptions{
		FilterRoles:   []string{"viewer", "respondent"},
		FilterStates:  []string{"open"},
		FilterFlowIDs: []string{"f1", "f2"},
		FilterRunIDs:  []string{"r1"},
		OrderBy:       []string{"created_timestamp DESC", "title"},
		PerPage:       10,
		Marker:        "m0",
	})
	require.NoError(t, err)
	require.Len(t, list.WebInputSummaries, 1)
	assert.False(t, list.HasNextPage(), "null marker ends pagination")
	s := list.WebInputSummaries[0]
	assert.Equal(t, webInputID, s.ID)
	assert.Equal(t, "open", s.Status)
	assert.Equal(t, []string{"viewer", "respondent"}, s.UserRoles)
	assert.Equal(t, "flow-1", s.Flow.ID)
	assert.Equal(t, "Transfer all of these files!", s.Run.Label)
	assert.Equal(t, 2026, s.CreatedTimestamp.Year())
	assert.Nil(t, s.ClosedTimestamp)
}

func TestListWebInputs_NoOptionsSendsNoQuery(t *testing.T) {
	c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		rawJSON(w, http.StatusOK, `{"web_input_summaries": [], "marker": null}`)
	})
	list, err := c.ListWebInputs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, list.WebInputSummaries)
}

func TestNewWebInputsPager_FollowsNullableMarker(t *testing.T) {
	pages := map[string]string{
		"":              `{"web_input_summaries": [{"id": "a"}, {"id": "b"}], "marker": "fake_marker_0"}`,
		"fake_marker_0": `{"web_input_summaries": [{"id": "c"}], "marker": "fake_marker_1"}`,
		"fake_marker_1": `{"web_input_summaries": [{"id": "d"}], "marker": null}`,
	}
	calls := 0
	c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		assert.Equal(t, "open", r.URL.Query().Get("filter_states"), "filters carried to every page")
		body, ok := pages[r.URL.Query().Get("marker")]
		require.True(t, ok, "unexpected marker %q", r.URL.Query().Get("marker"))
		rawJSON(w, http.StatusOK, body)
	})

	p := c.NewWebInputsPager(&flows.ListWebInputsOptions{FilterStates: []string{"open"}})
	var ids []string
	for p.HasNext() {
		items, err := p.NextPage(context.Background())
		require.NoError(t, err)
		for _, it := range items {
			ids = append(ids, it.ID)
		}
	}
	assert.Equal(t, []string{"a", "b", "c", "d"}, ids)
	assert.Equal(t, 3, calls)
}

func TestGetWebInput(t *testing.T) {
	t.Run("empty ID", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := c.GetWebInput(context.Background(), "")
		var ve *core.ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "webInputID", ve.Field)
	})

	t.Run("success", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/web_inputs/"+webInputID, r.URL.Path)
			rawJSON(w, http.StatusOK, `{
				"id": "`+webInputID+`", "status": "closed", "user_roles": ["viewer", "respondent"],
				"input_type": "selection",
				"input_schema": {"type": "string", "oneOf": [{"const": "opt-approve", "title": "Approve"}]},
				"context": {"title": "Approve deployment to production?", "presentation_style": "table",
				            "rows": [{"field": "Environment", "value": "production"}]},
				"options": [{"option_id": "opt-approve", "label": "Approve", "description": "Proceed."},
				            {"option_id": "opt-reject", "label": "Reject"}],
				"roles": {"viewer_urns": ["urn:globus:groups:id:g"], "respondent_urns": ["urn:globus:auth:identity:u2"]},
				"creator_urn": "urn:globus:auth:identity:u1",
				"flow": {"id": "flow-1", "title": "Multi Step Transfer"},
				"run": {"id": "run-1", "label": "Transfer all of these files!"},
				"created_timestamp": "2026-08-01T10:30:00+00:00",
				"edited_timestamp": "2026-08-01T10:30:00+00:00",
				"closed_timestamp": "2026-08-02T09:12:00+00:00"}`)
		})
		wi, err := c.GetWebInput(context.Background(), webInputID)
		require.NoError(t, err)
		assert.Equal(t, "closed", wi.Status)
		require.Len(t, wi.Options, 2)
		assert.Equal(t, "opt-approve", wi.Options[0].OptionID)
		require.NotNil(t, wi.Roles)
		assert.Equal(t, []string{"urn:globus:auth:identity:u2"}, wi.Roles.RespondentURNs)
		assert.Equal(t, "urn:globus:auth:identity:u1", wi.CreatorURN)
		assert.Equal(t, "table", wi.Context["presentation_style"])
		assert.Equal(t, "string", wi.InputSchema["type"])
		require.NotNil(t, wi.ClosedTimestamp)
		assert.Equal(t, 2, wi.ClosedTimestamp.Day())
	})

	t.Run("authentication policy required surfaces GARE details", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
			rawJSON(w, http.StatusForbidden, `{
				"error": {"code": "AUTHENTICATION_POLICY_REQUIRED", "detail": "reauthenticate"},
				"code": "AuthenticationPolicyRequired",
				"authorization_parameters": {"session_required_policies": ["pol-1"],
				  "session_message": "Globus Flows detected an unsatisfied session policy for this web input."}}`)
		})
		_, err := c.GetWebInput(context.Background(), webInputID)
		var apiErr *core.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
		ap, ok := apiErr.Details["authorization_parameters"].(map[string]interface{})
		require.True(t, ok, "authorization_parameters in Details: %#v", apiErr.Details)
		assert.Equal(t, []interface{}{"pol-1"}, ap["session_required_policies"])
	})
}

func TestRespondToWebInput(t *testing.T) {
	t.Run("empty ID", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {})
		_, err := c.RespondToWebInput(context.Background(), "", "x")
		var ve *core.ValidationError
		require.ErrorAs(t, err, &ve)
	})

	t.Run("posts response envelope", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/web_inputs/"+webInputID+"/respond", r.URL.Path)
			var body map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, map[string]interface{}{"response": map[string]interface{}{"value": "opt-approve"}}, body)
			rawJSON(w, http.StatusOK, `{"status": "ok"}`)
		})
		res, err := c.RespondToWebInput(context.Background(), webInputID, "opt-approve")
		require.NoError(t, err)
		assert.Equal(t, "ok", res["status"])
	})

	t.Run("closed web input returns API error", func(t *testing.T) {
		c := newFlowsClient(t, func(w http.ResponseWriter, r *http.Request) {
			rawJSON(w, http.StatusConflict, `{"error": {"code": "STATE_CONFLICT", "detail": "Web Input is already closed."}}`)
		})
		_, err := c.RespondToWebInput(context.Background(), webInputID, "opt-approve")
		var apiErr *core.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}
