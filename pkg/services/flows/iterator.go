// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
)

// FlowIterator provides an iterator for flows that handles pagination automatically.
type FlowIterator struct {
	client   *Client
	options  *ListFlowsOptions
	current  *FlowList
	position int
	err      error
}

// NewFlowIterator creates a new iterator for flows.
func NewFlowIterator(client *Client, options *ListFlowsOptions) *FlowIterator {
	if options == nil {
		options = &ListFlowsOptions{}
	}

	// Default values for pagination
	if options.Limit == 0 && options.PerPage == 0 {
		options.Limit = 100
	}

	return &FlowIterator{
		client:   client,
		options:  options,
		position: -1,
	}
}

// Next fetches the next flow in the iterator.
// Returns false when there are no more flows or an error occurred.
func (i *FlowIterator) Next(ctx context.Context) bool {
	if i.err != nil {
		return false
	}

	// If we're at the end of the current page, or haven't fetched any flows yet
	if i.current == nil || (i.position >= len(i.current.Flows)-1 && i.current.HadMore) {
		// If we have a marker from the previous page, update options
		if i.current != nil && i.current.HadMore {
			if len(i.current.Flows) > 0 {
				// Use offset-based pagination
				i.options.Offset = i.current.Offset + len(i.current.Flows)
				i.options.Marker = "" // Clear marker if it was set
			}
		}

		// Fetch the next page
		var err error
		i.current, err = i.client.ListFlows(ctx, i.options)
		if err != nil {
			i.err = err
			return false
		}

		// Reset position for the new page
		i.position = -1

		// Check if we got an empty page
		if len(i.current.Flows) == 0 {
			return false
		}
	}

	// Move to the next item
	i.position++

	// Check if we've reached the end
	return i.position < len(i.current.Flows)
}

// Flow returns the current flow in the iterator.
func (i *FlowIterator) Flow() *Flow {
	if i.current == nil || i.position < 0 || i.position >= len(i.current.Flows) {
		return nil
	}
	return &i.current.Flows[i.position]
}

// Err returns any error that occurred during iteration.
func (i *FlowIterator) Err() error {
	return i.err
}

// RunIterator provides an iterator for flow runs that handles pagination automatically.
type RunIterator struct {
	client   *Client
	options  *ListRunsOptions
	current  *RunList
	position int
	err      error
}

// NewRunIterator creates a new iterator for flow runs.
func NewRunIterator(client *Client, options *ListRunsOptions) *RunIterator {
	if options == nil {
		options = &ListRunsOptions{}
	}

	// Default values for pagination
	if options.Limit == 0 && options.PerPage == 0 {
		options.Limit = 100
	}

	return &RunIterator{
		client:   client,
		options:  options,
		position: -1,
	}
}

// Next fetches the next run in the iterator.
// Returns false when there are no more runs or an error occurred.
func (i *RunIterator) Next(ctx context.Context) bool {
	if i.err != nil {
		return false
	}

	// If we're at the end of the current page, or haven't fetched any runs yet
	if i.current == nil || (i.position >= len(i.current.Runs)-1 && i.current.HadMore) {
		// If we have a marker from the previous page, update options
		if i.current != nil && i.current.HadMore {
			if len(i.current.Runs) > 0 {
				// Use offset-based pagination
				i.options.Offset = i.current.Offset + len(i.current.Runs)
				i.options.Marker = "" // Clear marker if it was set
			}
		}

		// Fetch the next page
		var err error
		i.current, err = i.client.ListRuns(ctx, i.options)
		if err != nil {
			i.err = err
			return false
		}

		// Reset position for the new page
		i.position = -1

		// Check if we got an empty page
		if len(i.current.Runs) == 0 {
			return false
		}
	}

	// Move to the next item
	i.position++

	// Check if we've reached the end
	return i.position < len(i.current.Runs)
}

// Run returns the current run in the iterator.
func (i *RunIterator) Run() *RunResponse {
	if i.current == nil || i.position < 0 || i.position >= len(i.current.Runs) {
		return nil
	}
	return &i.current.Runs[i.position]
}

// Err returns any error that occurred during iteration.
func (i *RunIterator) Err() error {
	return i.err
}

// ActionProviderIterator provides an iterator for action providers that handles pagination automatically.
type ActionProviderIterator struct {
	client   *Client
	options  *ListActionProvidersOptions
	current  *ActionProviderList
	position int
	err      error
}

// NewActionProviderIterator creates a new iterator for action providers.
func NewActionProviderIterator(client *Client, options *ListActionProvidersOptions) *ActionProviderIterator {
	if options == nil {
		options = &ListActionProvidersOptions{}
	}

	// Default values for pagination
	if options.Limit == 0 && options.PerPage == 0 {
		options.Limit = 100
	}

	return &ActionProviderIterator{
		client:   client,
		options:  options,
		position: -1,
	}
}

// Next fetches the next action provider in the iterator.
// Returns false when there are no more providers or an error occurred.
func (i *ActionProviderIterator) Next(ctx context.Context) bool {
	if i.err != nil {
		return false
	}

	// If we're at the end of the current page, or haven't fetched any providers yet
	if i.current == nil || (i.position >= len(i.current.ActionProviders)-1 && i.current.HadMore) {
		// If we have a marker from the previous page, update options
		if i.current != nil && i.current.HadMore {
			if len(i.current.ActionProviders) > 0 {
				// Use offset-based pagination
				i.options.Offset = i.current.Offset + len(i.current.ActionProviders)
				i.options.Marker = "" // Clear marker if it was set
			}
		}

		// Fetch the next page
		var err error
		i.current, err = i.client.ListActionProviders(ctx, i.options)
		if err != nil {
			i.err = err
			return false
		}

		// Reset position for the new page
		i.position = -1

		// Check if we got an empty page
		if len(i.current.ActionProviders) == 0 {
			return false
		}
	}

	// Move to the next item
	i.position++

	// Check if we've reached the end
	return i.position < len(i.current.ActionProviders)
}

// ActionProvider returns the current action provider in the iterator.
func (i *ActionProviderIterator) ActionProvider() *ActionProvider {
	if i.current == nil || i.position < 0 || i.position >= len(i.current.ActionProviders) {
		return nil
	}
	return &i.current.ActionProviders[i.position]
}

// Err returns any error that occurred during iteration.
func (i *ActionProviderIterator) Err() error {
	return i.err
}

// RunLogIterator provides an iterator for run logs that handles pagination automatically.
type RunLogIterator struct {
	client   *Client
	runID    string
	limit    int
	offset   int
	current  *RunLogList
	position int
	err      error
}

// NewRunLogIterator creates a new iterator for run logs.
func NewRunLogIterator(client *Client, runID string, limit int) *RunLogIterator {
	if limit <= 0 {
		limit = 100
	}

	return &RunLogIterator{
		client:   client,
		runID:    runID,
		limit:    limit,
		offset:   0,
		position: -1,
	}
}

// Next fetches the next log entry in the iterator.
// Returns false when there are no more entries or an error occurred.
func (i *RunLogIterator) Next(ctx context.Context) bool {
	if i.err != nil {
		return false
	}

	// If we're at the end of the current page, or haven't fetched any logs yet
	if i.current == nil || (i.position >= len(i.current.Entries)-1 && i.current.HadMore) {
		// Fetch the next page
		var err error
		i.current, err = i.client.GetRunLogs(ctx, i.runID, i.limit, i.offset)
		if err != nil {
			i.err = err
			return false
		}

		// Update offset for next page
		i.offset += len(i.current.Entries)

		// Reset position for the new page
		i.position = -1

		// Check if we got an empty page
		if len(i.current.Entries) == 0 {
			return false
		}
	}

	// Move to the next item
	i.position++

	// Check if we've reached the end
	return i.position < len(i.current.Entries)
}

// LogEntry returns the current log entry in the iterator.
func (i *RunLogIterator) LogEntry() *RunLogEntry {
	if i.current == nil || i.position < 0 || i.position >= len(i.current.Entries) {
		return nil
	}
	return &i.current.Entries[i.position]
}

// Err returns any error that occurred during iteration.
func (i *RunLogIterator) Err() error {
	return i.err
}

// ActionRoleIterator provides an iterator for action roles that handles pagination automatically.
type ActionRoleIterator struct {
	client     *Client
	providerID string
	limit      int
	offset     int
	current    *ActionRoleList
	position   int
	err        error
}

// NewActionRoleIterator creates a new iterator for action roles.
func NewActionRoleIterator(client *Client, providerID string, limit int) *ActionRoleIterator {
	if limit <= 0 {
		limit = 100
	}

	return &ActionRoleIterator{
		client:     client,
		providerID: providerID,
		limit:      limit,
		offset:     0,
		position:   -1,
	}
}

// Next fetches the next action role in the iterator.
// Returns false when there are no more roles or an error occurred.
func (i *ActionRoleIterator) Next(ctx context.Context) bool {
	if i.err != nil {
		return false
	}

	// If we're at the end of the current page, or haven't fetched any roles yet
	if i.current == nil || (i.position >= len(i.current.ActionRoles)-1 && i.current.HadMore) {
		// Fetch the next page
		var err error
		i.current, err = i.client.ListActionRoles(ctx, i.providerID, i.limit, i.offset)
		if err != nil {
			i.err = err
			return false
		}

		// Update offset for next page
		i.offset += len(i.current.ActionRoles)

		// Reset position for the new page
		i.position = -1

		// Check if we got an empty page
		if len(i.current.ActionRoles) == 0 {
			return false
		}
	}

	// Move to the next item
	i.position++

	// Check if we've reached the end
	return i.position < len(i.current.ActionRoles)
}

// ActionRole returns the current action role in the iterator.
func (i *ActionRoleIterator) ActionRole() *ActionRole {
	if i.current == nil || i.position < 0 || i.position >= len(i.current.ActionRoles) {
		return nil
	}
	return &i.current.ActionRoles[i.position]
}

// Err returns any error that occurred during iteration.
func (i *ActionRoleIterator) Err() error {
	return i.err
}
