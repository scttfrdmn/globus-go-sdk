// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package timers provides a client for interacting with the Globus Timers service.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees. Components listed below are
considered part of the public API and will not change incompatibly within a major version:

  - Client interface and implementation
  - Timer management operations (CreateTimer, GetTimer, UpdateTimer, DeleteTimer, ListTimers)
  - Timer control operations (PauseTimer, ResumeTimer, RunTimer)
  - Run management operations (ListRuns, GetRun)
  - Core model types (Timer, Schedule, Callback, TimerRun)
  - Status and type constants (TimerStatus, ScheduleType, CallbackType, RunStatus)
  - Helper methods for creating common timer types
  - Client configuration options
  - Error handling patterns
  - Advanced scheduling options
  - Callback implementation details

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

Create a new timers client:

	timersClient := timers.NewClient(
		timers.WithAuthorizer(authorizer),
	)

Timer Management:

	// Create a timer
	timer := &timers.Timer{
		Name:        "My Timer",
		Description: "A timer for demonstration",
		Schedule: &timers.Schedule{
			Type:     timers.ScheduleTypeRecurring,
			Interval: "PT1H", // ISO 8601 duration - every hour
		},
		Callback: &timers.Callback{
			Type: timers.CallbackTypeFlow,
			URL:  "https://flows.globus.org/flows/12345",
			Body: map[string]interface{}{
				"input_param": "value",
			},
		},
	}

	created, err := timersClient.CreateTimer(ctx, timer)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Created timer with ID: %s\n", created.ID)

	// List timers
	timers, err := timersClient.ListTimers(ctx, nil)
	if err != nil {
		// Handle error
	}

	for _, t := range timers.Timers {
		fmt.Printf("Timer: %s (%s)\n", t.Name, t.ID)
	}

	// Get a specific timer
	timer, err := timersClient.GetTimer(ctx, "timer_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Timer: %s, Status: %s\n", timer.Name, timer.Status)

	// Update a timer
	update := &timers.Timer{
		Name:        "Updated Timer Name",
		Description: "Updated description",
	}

	updated, err := timersClient.UpdateTimer(ctx, "timer_id", update)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Updated timer: %s\n", updated.Name)

	// Delete a timer
	err = timersClient.DeleteTimer(ctx, "timer_id")
	if err != nil {
		// Handle error
	}

Timer Control:

	// Pause a timer
	err = timersClient.PauseTimer(ctx, "timer_id")
	if err != nil {
		// Handle error
	}

	// Resume a timer
	err = timersClient.ResumeTimer(ctx, "timer_id")
	if err != nil {
		// Handle error
	}

	// Run a timer manually
	runID, err := timersClient.RunTimer(ctx, "timer_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Manual run started with ID: %s\n", runID)

Run Management:

	// List runs for a timer
	runs, err := timersClient.ListRuns(ctx, "timer_id", nil)
	if err != nil {
		// Handle error
	}

	for _, run := range runs.Runs {
		fmt.Printf("Run: %s, Status: %s, Start time: %s\n", run.ID, run.Status, run.StartTime)
	}

	// Get a specific run
	run, err := timersClient.GetRun(ctx, "timer_id", "run_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Run details: Status: %s, Completion time: %s\n", run.Status, run.CompletionTime)

Helper Methods:

	// Create a one-time timer (runs once at a specific time)
	oneTimeTimer, err := timersClient.CreateOnceTimer(
		ctx,
		"One-time Timer",
		"Runs once at the specified time",
		time.Now().Add(24*time.Hour), // Run tomorrow
		timersClient.CreateFlowCallback(
			"https://flows.globus.org/flows/12345",
			map[string]interface{}{"param": "value"},
		),
	)
	if err != nil {
		// Handle error
	}

	// Create a recurring timer (runs at regular intervals)
	recurringTimer, err := timersClient.CreateRecurringTimer(
		ctx,
		"Recurring Timer",
		"Runs every hour",
		"PT1H", // ISO 8601 duration - every hour
		timersClient.CreateWebCallback(
			"https://example.com/webhook",
			map[string]interface{}{"event": "timer_triggered"},
		),
	)
	if err != nil {
		// Handle error
	}

	// Create a cron timer (runs on a cron schedule)
	cronTimer, err := timersClient.CreateCronTimer(
		ctx,
		"Cron Timer",
		"Runs at 10:00 AM every day",
		"0 10 * * *", // Cron expression
		timersClient.CreateFlowCallback(
			"https://flows.globus.org/flows/67890",
			map[string]interface{}{"action": "daily_process"},
		),
	)
	if err != nil {
		// Handle error
	}

FlowTimer Helper (Added in v3.65.0):

The FlowTimer helper simplifies creating timers that run Globus Flows.
This matches the Python SDK v3.65.0 FlowTimer payload class.

	// Define a flow to run
	flowTimer := &timers.FlowTimer{
		FlowID:    "my-flow-id",
		FlowScope: "https://auth.globus.org/scopes/my-flow-id/flow_run",
		FlowInput: map[string]interface{}{
			"source": "/path/to/source",
			"dest":   "/path/to/dest",
		},
		FlowLabel: "My Flow Run",
	}

	// Create a one-time flow timer
	timer, err := timersClient.CreateFlowTimerOnce(
		ctx,
		"Daily Backup",
		time.Now().Add(1*time.Hour),
		flowTimer,
		nil, // Optional additional data
	)

	// Create a recurring flow timer (every 24 hours)
	timer, err := timersClient.CreateFlowTimerRecurring(
		ctx,
		"Daily Backup",
		time.Now(),
		"P1D", // ISO 8601: 1 day
		nil,   // No end time
		flowTimer,
		nil,
	)

	// Create a cron-scheduled flow timer (every Monday at 9 AM)
	timer, err := timersClient.CreateFlowTimerCron(
		ctx,
		"Weekly Report",
		"0 9 * * 1",        // Every Monday at 9:00 AM
		"America/New_York", // Timezone
		nil,                // No end time
		flowTimer,
		nil,
	)
*/
package timers
