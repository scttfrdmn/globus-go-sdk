// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// Web inputs: human-in-the-loop prompts raised by a running flow.
// Added in Python SDK v4.9.0 (FlowsClient.list_web_inputs, get_web_input,
// respond_to_web_input).

// WebInputFlowRef identifies the flow a web input belongs to.
type WebInputFlowRef struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

// WebInputRunRef identifies the run a web input belongs to.
type WebInputRunRef struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

// WebInputSummary is one entry of a web input listing.
type WebInputSummary struct {
	ID               string          `json:"id"`
	Status           string          `json:"status"`               // "open" or "closed"
	UserRoles        []string        `json:"user_roles,omitempty"` // "viewer", "respondent"
	InputType        string          `json:"input_type,omitempty"` // e.g. "selection"
	Title            string          `json:"title,omitempty"`
	Flow             WebInputFlowRef `json:"flow"`
	Run              WebInputRunRef  `json:"run"`
	CreatedTimestamp time.Time       `json:"created_timestamp"`
	EditedTimestamp  time.Time       `json:"edited_timestamp"`
	ClosedTimestamp  *time.Time      `json:"closed_timestamp,omitempty"`
}

// WebInputOption is one selectable option of a "selection" web input. Pass
// OptionID as the value to RespondToWebInput.
type WebInputOption struct {
	OptionID    string `json:"option_id"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

// WebInputRoles lists the principal URNs granted each role on a web input.
type WebInputRoles struct {
	ViewerURNs     []string `json:"viewer_urns,omitempty"`
	RespondentURNs []string `json:"respondent_urns,omitempty"`
}

// WebInput is the full web input document returned by GetWebInput.
type WebInput struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	UserRoles   []string               `json:"user_roles,omitempty"`
	InputType   string                 `json:"input_type,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
	// Context is the prompt shown to respondents. Its shape depends on
	// presentation_style ("text" carries "text"; "table" carries "rows").
	Context          map[string]interface{} `json:"context,omitempty"`
	Options          []WebInputOption       `json:"options,omitempty"`
	Roles            *WebInputRoles         `json:"roles,omitempty"`
	CreatorURN       string                 `json:"creator_urn,omitempty"`
	Flow             WebInputFlowRef        `json:"flow"`
	Run              WebInputRunRef         `json:"run"`
	CreatedTimestamp time.Time              `json:"created_timestamp"`
	EditedTimestamp  time.Time              `json:"edited_timestamp"`
	ClosedTimestamp  *time.Time             `json:"closed_timestamp,omitempty"`
}

// WebInputList is one page of ListWebInputs results. Marker is empty on the
// last page (the service returns "marker": null; upstream's
// NullableMarkerPaginator treats null as the end of pagination).
type WebInputList struct {
	WebInputSummaries []WebInputSummary `json:"web_input_summaries"`
	Marker            string            `json:"marker,omitempty"`
}

// HasNextPage reports whether another page is available.
func (l *WebInputList) HasNextPage() bool { return l.Marker != "" }

// ListWebInputsOptions filters a web input listing. Every slice is
// comma-joined into its query parameter, matching upstream.
type ListWebInputsOptions struct {
	FilterRoles   []string // "viewer", "respondent"
	FilterStates  []string // "open", "closed"
	FilterFlowIDs []string
	FilterRunIDs  []string
	// OrderBy criteria such as "created_timestamp DESC"; known keys are
	// created_timestamp, edited_timestamp and closed_timestamp.
	OrderBy []string
	PerPage int
	Marker  string
}

// ListWebInputs lists web inputs the caller has a role on (GET /web_inputs).
func (c *Client) ListWebInputs(ctx context.Context, options *ListWebInputsOptions) (*WebInputList, error) {
	query := url.Values{}
	if options != nil {
		setCommaJoined(query, "filter_roles", options.FilterRoles)
		setCommaJoined(query, "filter_states", options.FilterStates)
		setCommaJoined(query, "filter_flow_ids", options.FilterFlowIDs)
		setCommaJoined(query, "filter_run_ids", options.FilterRunIDs)
		setCommaJoined(query, "orderby", options.OrderBy)
		if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}

	var list WebInputList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/web_inputs", query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetWebInput fetches a web input (GET /web_inputs/{id}). If the flow's
// authentication policy is not satisfied by the caller's session, the service
// responds with an AuthenticationPolicyRequired error whose Details carry
// authorization_parameters (a GARE).
func (c *Client) GetWebInput(ctx context.Context, webInputID string) (*WebInput, error) {
	if webInputID == "" {
		return nil, &core.ValidationError{Field: "webInputID", Message: "web input ID is required"}
	}
	var wi WebInput
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/web_inputs/%s", webInputID), nil, nil, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// RespondToWebInput submits a response (POST /web_inputs/{id}/respond, body
// {"response": {"value": value}}). The caller needs the respondent role. For
// a "selection" web input, value must be one of the input's option IDs.
func (c *Client) RespondToWebInput(ctx context.Context, webInputID string, value interface{}) (map[string]interface{}, error) {
	if webInputID == "" {
		return nil, &core.ValidationError{Field: "webInputID", Message: "web input ID is required"}
	}
	body := map[string]interface{}{"response": map[string]interface{}{"value": value}}
	var result map[string]interface{}
	path := fmt.Sprintf("/web_inputs/%s/respond", webInputID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// NewWebInputsPager returns a marker Paginator over all web inputs matching
// opts. Pass nil for default options.
func (c *Client) NewWebInputsPager(opts *ListWebInputsOptions) paging.Paginator[WebInputSummary] {
	pageSize := 0
	if opts != nil && opts.PerPage > 0 {
		pageSize = opts.PerPage
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]WebInputSummary, bool, string, error) {
			o := &ListWebInputsOptions{Marker: marker, PerPage: limit}
			if opts != nil {
				o.FilterRoles = opts.FilterRoles
				o.FilterStates = opts.FilterStates
				o.FilterFlowIDs = opts.FilterFlowIDs
				o.FilterRunIDs = opts.FilterRunIDs
				o.OrderBy = opts.OrderBy
			}
			result, err := c.ListWebInputs(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.WebInputSummaries, result.HasNextPage(), result.Marker, nil
		},
		pageSize,
	)
}

func setCommaJoined(q url.Values, key string, vals []string) {
	if len(vals) > 0 {
		q.Set(key, strings.Join(vals, ","))
	}
}
