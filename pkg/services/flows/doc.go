// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package flows provides a client for interacting with the Globus Flows service.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees. Components listed below are
considered part of the public API and will not change incompatibly within a major version:

  - Client interface and implementation
  - Flow management methods (ListFlows, GetFlow, CreateFlow, etc.)
  - Flow execution methods (RunFlow, GetRun, ListRuns, etc.)
  - Core model types (Flow, Run, ActionProvider, etc.)
  - Pagination support and iterators
  - Client configuration options
  - Error handling patterns
  - Batch operations
  - Polling mechanisms (WaitForRun)
  - Action provider interactions

# Compatibility Guarantees

For stable packages:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v3.x to v4.x)

# Synchronized Versioning

Starting with v3.60.0-1, this package follows synchronized versioning with the Globus Python SDK.
This ensures API compatibility and feature parity across language implementations.

# Basic Usage

Create a new flows client:

	flowsClient := flows.NewClient(
		flows.WithAuthorizer(authorizer),
	)

Flow Management:

	// List flows
	flowsIterator, err := flowsClient.ListFlows(ctx, nil)
	if err != nil {
		// Handle error
	}

	for flowsIterator.HasNext() {
		flow, err := flowsIterator.Flow()
		if err != nil {
			// Handle error
		}
		fmt.Printf("Flow ID: %s, Title: %s\n", flow.ID, flow.Title)
	}

	// Get a specific flow
	flow, err := flowsClient.GetFlow(ctx, "flow_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.Definition.Description)

	// Create a flow
	newFlow := &flows.Flow{
		Title:       "My New Flow",
		Description: "A flow for my workflow",
		Definition: &flows.FlowDefinition{
			// Flow definition...
		},
	}

	created, err := flowsClient.CreateFlow(ctx, newFlow)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Created flow with ID: %s\n", created.ID)

	// Update a flow
	update := &flows.FlowUpdate{
		Title: "Updated Flow Title",
		Definition: &flows.FlowDefinition{
			// Updated flow definition...
		},
	}

	updated, err := flowsClient.UpdateFlow(ctx, "flow_id", update)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Updated flow: %s\n", updated.Title)

	// Delete a flow
	err = flowsClient.DeleteFlow(ctx, "flow_id")
	if err != nil {
		// Handle error
	}

Flow Execution:

	// Run a flow
	input := map[string]interface{}{
		"input_key": "input_value",
	}

	runID, err := flowsClient.RunFlow(ctx, "flow_id", input)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Started flow run with ID: %s\n", runID)

	// Get run status
	run, err := flowsClient.GetRun(ctx, "flow_id", runID)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Run status: %s\n", run.Status)

	// Wait for run completion
	run, err = flowsClient.WaitForRun(ctx, "flow_id", runID)
	if err != nil {
		// Handle error
	}

	if run.Status == "SUCCEEDED" {
		fmt.Println("Flow run completed successfully!")
	} else {
		fmt.Printf("Flow run failed: %s\n", run.Status)
	}

	// Cancel a run
	err = flowsClient.CancelRun(ctx, "flow_id", runID)
	if err != nil {
		// Handle error
	}

Batch Operations:

	// Run multiple flows
	batch := flows.NewRunBatch()
	batch.AddRun("flow_id_1", map[string]interface{}{"key1": "value1"})
	batch.AddRun("flow_id_2", map[string]interface{}{"key2": "value2"})

	results, err := flowsClient.BatchRunFlows(ctx, batch)
	if err != nil {
		// Handle error
	}

	for flowID, result := range results.Results {
		fmt.Printf("Flow %s run ID: %s\n", flowID, result.RunID)
	}

	// Get multiple run statuses
	runBatch := flows.NewRunIDBatch()
	runBatch.AddRun("flow_id_1", "run_id_1")
	runBatch.AddRun("flow_id_2", "run_id_2")

	runResults, err := flowsClient.BatchGetRuns(ctx, runBatch)
	if err != nil {
		// Handle error
	}

	for _, run := range runResults.Results {
		fmt.Printf("Run %s status: %s\n", run.ID, run.Status)
	}

Action Providers:

	// List action providers
	providers, err := flowsClient.ListActionProviders(ctx)
	if err != nil {
		// Handle error
	}

	for _, provider := range providers {
		fmt.Printf("Provider: %s (%s)\n", provider.Name, provider.ID)
	}

	// Get action provider details
	provider, err := flowsClient.GetActionProvider(ctx, "provider_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Provider: %s\n", provider.Name)
*/
package flows
