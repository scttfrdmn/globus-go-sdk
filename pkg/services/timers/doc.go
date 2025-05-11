// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package timers provides a client for interacting with the Globus Timers service.

# STABILITY: BETA

This package is approaching stability but may still undergo minor changes.
Components listed below are considered relatively stable, but may have
minor signature changes before the package is marked as stable:

  - Client interface and implementation
  - Timer management operations (CreateTimer, GetTimer, UpdateTimer, DeleteTimer, ListTimers)
  - Timer control operations (PauseTimer, ResumeTimer, RunTimer)
  - Run management operations (ListRuns, GetRun)
  - Core model types (Timer, Schedule, Callback, TimerRun)
  - Status and type constants (TimerStatus, ScheduleType, CallbackType, RunStatus)
  - Helper methods for creating common timer types
  - Client configuration options

The following components are less stable and more likely to evolve:

  - Error handling patterns
  - Advanced scheduling options
  - Callback implementation details

# Compatibility Notes

For beta packages:
  - Minor backward-incompatible changes may still occur in minor releases
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices
  - Migration paths will be provided for any breaking changes

This package is expected to reach stable status in version v1.0.0.
Until then, users should review the CHANGELOG when upgrading.

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
*/
package timers
