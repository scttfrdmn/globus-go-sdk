# Timers Client Guide

This guide provides information about using the Timers service client in the Globus Go SDK.

## Overview

The Globus Timers service allows you to schedule tasks to run at specific times or intervals. You can use it to:

- Run tasks at a specific time
- Run tasks on a recurring schedule
- Trigger Globus Flows
- Make HTTP requests to web services

## Authentication

The Timers client requires an access token with the appropriate scope. The required scope is:

```
https://auth.globus.org/scopes/b2b8f087-7a70-480c-9480-da1c3d3f8c86/timer
```

You can obtain this scope when requesting authorization from Globus Auth. Add it to your consent request:

```go
scopes := []string{
    "openid", 
    "profile", 
    "email",
    "https://auth.globus.org/scopes/b2b8f087-7a70-480c-9480-da1c3d3f8c86/timer",
}
```

## Creating a Timers Client

Create a new Timers client using either the SDK configuration or directly:

```go
// Using SDK configuration
config := pkg.NewConfigFromEnvironment()
timersClient := config.NewTimersClient(accessToken)

// Direct creation
timersClient := timers.NewClient(accessToken)
```

## Timer Schedules

The Timers service supports three types of schedules:

### One-Time Schedule

A one-time schedule runs a timer once at a specific time.

```go
schedule := timers.Schedule{
    Type:      string(timers.ScheduleTypeOnce),
    StartTime: &startTime, // time.Time
}
```

### Recurring Schedule

A recurring schedule runs a timer repeatedly at specified intervals.

```go
schedule := timers.Schedule{
    Type:      string(timers.ScheduleTypeRecurring),
    StartTime: &startTime, // time.Time
    EndTime:   &endTime,   // Optional end time
    Interval:  stringPtr("1 hour"), // Interval (e.g., "1 hour", "30 minutes", "1 day")
}
```

### Cron Schedule

A cron schedule runs a timer according to a cron expression.

```go
schedule := timers.Schedule{
    Type:           string(timers.ScheduleTypeCron),
    CronExpression: stringPtr("0 0 * * *"), // Run at midnight every day
    Timezone:       stringPtr("America/Chicago"), // Optional timezone
    EndTime:        &endTime, // Optional end time
}
```

## Timer Callbacks

The Timers service supports two types of callbacks:

### Web Callback

A web callback makes an HTTP request to a specified URL.

```go
callback := timers.Callback{
    Type:   string(timers.CallbackTypeWeb),
    URL:    stringPtr("https://example.com/webhook"),
    Method: stringPtr("POST"), // HTTP method (GET, POST, PUT, DELETE, etc.)
    Headers: map[string]string{ // Optional HTTP headers
        "Content-Type": "application/json",
        "Authorization": "Bearer token123",
    },
    Body: stringPtr(`{"message": "Hello, world!"}`), // Optional request body
}
```

### Flow Callback

A flow callback triggers a Globus Flow.

```go
callback := timers.Callback{
    Type:      string(timers.CallbackTypeFlow),
    FlowID:    stringPtr("your-flow-id"),
    FlowLabel: stringPtr("Triggered by timer"), // Optional label for the flow run
    FlowInput: map[string]interface{}{ // Optional input for the flow
        "message": "Hello from timer",
        "timestamp": time.Now().Unix(),
    },
}
```

## Creating Timers

### Basic Timer Creation

```go
request := &timers.CreateTimerRequest{
    Name:     "My Timer",
    Schedule: schedule,
    Callback: callback,
    Data: map[string]interface{}{ // Optional user-provided data
        "description": "This is my timer",
        "created_by": "Globus Go SDK Example",
    },
}

timer, err := timersClient.CreateTimer(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Created timer with ID: %s\n", timer.ID)
```

### Helper Methods for Common Timer Types

The SDK provides helper methods for creating common timer types:

#### One-Time Timer

```go
// Create a one-time timer
timer, err := timersClient.CreateOnceTimer(
    ctx,
    "One-Time Timer",
    startTime,
    callback,
    map[string]interface{}{
        "description": "Runs once at a specific time",
    },
)
```

#### Recurring Timer

```go
// Create a recurring timer
timer, err := timersClient.CreateRecurringTimer(
    ctx,
    "Recurring Timer",
    startTime,
    "1 hour", // Interval
    &endTime, // Optional end time
    callback,
    map[string]interface{}{
        "description": "Runs every hour",
    },
)
```

#### Cron Timer

```go
// Create a cron timer
timer, err := timersClient.CreateCronTimer(
    ctx,
    "Cron Timer",
    "0 0 * * *", // Cron expression (midnight every day)
    "UTC",       // Timezone
    &endTime,    // Optional end time
    callback,
    map[string]interface{}{
        "description": "Runs at midnight every day",
    },
)
```

#### Helper Methods for Callbacks

```go
// Create a web callback
webCallback := timers.CreateWebCallback(
    "https://example.com/webhook",
    "POST",
    map[string]string{
        "Content-Type": "application/json",
    },
    stringPtr(`{"message": "Hello, world!"}`),
)

// Create a flow callback
flowCallback := timers.CreateFlowCallback(
    "your-flow-id",
    "Triggered by timer",
    map[string]interface{}{
        "message": "Hello from timer",
    },
)
```

## Managing Timers

### Listing Timers

```go
// List all timers
timers, err := timersClient.ListTimers(ctx, nil)

// List with options
limit := 10
status := string(timers.TimerStatusActive)
options := &timers.ListTimersOptions{
    Limit:  &limit,
    Status: &status,
}
timers, err := timersClient.ListTimers(ctx, options)

// Iterate through timers
for _, timer := range timers.Timers {
    fmt.Printf("Timer: %s (ID: %s)\n", timer.Name, timer.ID)
}

// Handle pagination
if timers.HasNextPage {
    nextOptions := &timers.ListTimersOptions{
        Marker: timers.NextPage,
    }
    nextPage, err := timersClient.ListTimers(ctx, nextOptions)
    // ...
}
```

### Getting Timer Details

```go
timer, err := timersClient.GetTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Timer: %s\n", timer.Name)
fmt.Printf("Status: %s\n", timer.Status)
if timer.NextDue != nil {
    fmt.Printf("Next due: %s\n", timer.NextDue.Format(time.RFC3339))
}
```

### Updating Timers

```go
newName := "Updated Timer Name"
request := &timers.UpdateTimerRequest{
    Name: &newName,
    // Other fields to update...
}

updatedTimer, err := timersClient.UpdateTimer(ctx, timerID, request)
if err != nil {
    // Handle error
}
```

### Pausing and Resuming Timers

```go
// Pause a timer
pausedTimer, err := timersClient.PauseTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Timer status after pause: %s\n", pausedTimer.Status)

// Resume a timer
resumedTimer, err := timersClient.ResumeTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Timer status after resume: %s\n", resumedTimer.Status)
```

### Manually Triggering Timers

```go
run, err := timersClient.RunTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Timer run started: %s\n", run.ID)
fmt.Printf("Run status: %s\n", run.Status)
```

### Deleting Timers

```go
err := timersClient.DeleteTimer(ctx, timerID)
if err != nil {
    // Handle error
}
```

## Timer Runs

### Listing Timer Runs

```go
// List all runs for a timer
runs, err := timersClient.ListRuns(ctx, timerID, nil)

// List with options
limit := 10
status := string(timers.RunStatusSuccess)
options := &timers.ListRunsOptions{
    Limit:  &limit,
    Status: &status,
}
runs, err := timersClient.ListRuns(ctx, timerID, options)

// Iterate through runs
for _, run := range runs.Runs {
    fmt.Printf("Run: %s\n", run.ID)
    fmt.Printf("Status: %s\n", run.Status)
    fmt.Printf("Start time: %s\n", run.StartTime.Format(time.RFC3339))
}
```

### Getting Run Details

```go
run, err := timersClient.GetRun(ctx, timerID, runID)
if err != nil {
    // Handle error
}

fmt.Printf("Run: %s\n", run.ID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Start time: %s\n", run.StartTime.Format(time.RFC3339))

if run.Result != nil {
    fmt.Printf("Result status: %s\n", run.Result.Status)
    if run.Result.StatusCode != nil {
        fmt.Printf("HTTP status code: %d\n", *run.Result.StatusCode)
    }
    if run.Result.RunID != nil {
        fmt.Printf("Flow run ID: %s\n", *run.Result.RunID)
    }
}
```

## Best Practices

### Error Handling

Always check for errors when making API calls:

```go
timer, err := timersClient.CreateTimer(ctx, request)
if err != nil {
    // Handle error appropriately
    fmt.Printf("Failed to create timer: %v\n", err)
    return
}
```

### Resource Cleanup

Delete timers when they are no longer needed:

```go
// Clean up the timer when done
defer func() {
    err := timersClient.DeleteTimer(ctx, timer.ID)
    if err != nil {
        fmt.Printf("Warning: Failed to delete timer: %v\n", err)
    }
}()
```

### Schedule Planning

- For one-time timers, specify a time in the future
- For recurring timers, consider the interval and end time carefully
- For cron timers, test the cron expression to ensure it matches your expectations

### Callback Reliability

- Web callbacks should be to reliable, publicly accessible endpoints
- Flow callbacks should reference flows that are owned by or shared with your Globus account
- Include appropriate retry logic in your callback handlers

## Example Application

For a complete example of using the Timers service, see the [Timers Example](../examples/timers/README.md).

## API Reference

For detailed information about the Timers API, see the [Globus Timers API Documentation](https://docs.globus.org/api/timers/).