// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"fmt"
)

// CreateFlowTimer is a convenience helper that builds a flow timer create
// document and submits it via CreateTimer (POST /v2/timer). body is the flow run
// input document.
func (c *Client) CreateFlowTimer(ctx context.Context, name, flowID string, schedule *Schedule, body map[string]interface{}) (*Timer, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if flowID == "" {
		return nil, fmt.Errorf("flowID is required")
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule is required")
	}
	return c.CreateTimer(ctx, NewFlowTimer(name, flowID, schedule, body))
}

// CreateTransferTimer is a convenience helper that builds a transfer timer create
// document and submits it via CreateTimer (POST /v2/timer). body is a
// TransferData document.
func (c *Client) CreateTransferTimer(ctx context.Context, name string, schedule *Schedule, body map[string]interface{}) (*Timer, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if schedule == nil {
		return nil, fmt.Errorf("schedule is required")
	}
	if body == nil {
		return nil, fmt.Errorf("body is required")
	}
	return c.CreateTimer(ctx, NewTransferTimer(name, schedule, body))
}
