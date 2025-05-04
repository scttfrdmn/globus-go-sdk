# Timers Service: Timer Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Timers client provides methods for creating, retrieving, updating, and managing timer operations.

## Timer Model

The central data type for timer operations is the `Timer` struct:

```go
type Timer struct {
    // ID is the unique identifier for the timer
    ID string `json:"id"`

    // Name is the user-provided name for the timer
    Name string `json:"name"`

    // Owner is the ID of the user who created the timer
    Owner string `json:"owner"`

    // Schedule defines when the timer should run
    Schedule *Schedule `json:"schedule"`

    // Callback defines what the timer should do when triggered
    Callback *Callback `json:"callback"`

    // Status indicates the current status of the timer
    Status string `json:"status"`

    // LastUpdate is when the timer was last updated
    LastUpdate time.Time `json:"last_update"`

    // NextDue is when the timer is next scheduled to run
    NextDue *time.Time `json:"next_due,omitempty"`

    // LastRun is when the timer was last run
    LastRun *time.Time `json:"last_run,omitempty"`

    // LastRunStatus indicates the status of the last run
    LastRunStatus string `json:"last_run_status,omitempty"`

    // CreateTime is when the timer was created
    CreateTime time.Time `json:"create_time"`

    // Data contains additional user-provided data for the timer
    Data map[string]interface{} `json:"data,omitempty"`
}
```

Related structures for schedules and callbacks:

```go
// Schedule defines when a timer should run
type Schedule struct {
    // Type is the type of schedule (once, recurring, cron)
    Type string `json:"type"`

    // StartTime is when the timer should start
    StartTime *time.Time `json:"start_time,omitempty"`

    // EndTime is when the timer should stop
    EndTime *time.Time `json:"end_time,omitempty"`

    // Interval is the interval for recurring timers
    Interval *string `json:"interval,omitempty"`

    // CronExpression is the cron expression for cron-based timers
    CronExpression *string `json:"cron_expression,omitempty"`

    // Timezone is the timezone for the schedule
    Timezone *string `json:"timezone,omitempty"`
}

// Callback defines what a timer should do when triggered
type Callback struct {
    // Type is the type of callback (flow, web)
    Type string `json:"type"`

    // URL is the URL to call for web callbacks
    URL *string `json:"url,omitempty"`

    // Method is the HTTP method to use for web callbacks
    Method *string `json:"method,omitempty"`

    // FlowID is the ID of the flow to run for flow callbacks
    FlowID *string `json:"flow_id,omitempty"`

    // FlowLabel is the label to use for flow callbacks
    FlowLabel *string `json:"flow_label,omitempty"`

    // FlowInput is the input data for flow callbacks
    FlowInput map[string]interface{} `json:"flow_input,omitempty"`

    // Headers are the HTTP headers to use for web callbacks
    Headers map[string]string `json:"headers,omitempty"`

    // Body is the HTTP body to use for web callbacks
    Body *string `json:"body,omitempty"`
}
```

## Schedule Types

The Timers service supports three types of schedules:

```go
// ScheduleType represents the possible types of timer schedules
type ScheduleType string

const (
    // ScheduleTypeOnce indicates the timer should run once
    ScheduleTypeOnce ScheduleType = "once"

    // ScheduleTypeRecurring indicates the timer should run recurringly
    ScheduleTypeRecurring ScheduleType = "recurring"

    // ScheduleTypeCron indicates the timer should run on a cron schedule
    ScheduleTypeCron ScheduleType = "cron"
)
```

## Callback Types

The Timers service supports two types of callbacks:

```go
// CallbackType represents the possible types of timer callbacks
type CallbackType string

const (
    // CallbackTypeFlow indicates the timer should run a flow
    CallbackTypeFlow CallbackType = "flow"

    // CallbackTypeWeb indicates the timer should make a web request
    CallbackTypeWeb CallbackType = "web"
)
```

## Timer Status

Timers can have the following status values:

```go
// TimerStatus represents the possible statuses of a timer
type TimerStatus string

const (
    // TimerStatusActive indicates the timer is active
    TimerStatusActive TimerStatus = "active"

    // TimerStatusPaused indicates the timer is paused
    TimerStatusPaused TimerStatus = "paused"

    // TimerStatusExpired indicates the timer has expired
    TimerStatusExpired TimerStatus = "expired"

    // TimerStatusFailed indicates the timer has failed
    TimerStatusFailed TimerStatus = "failed"

    // TimerStatusComplete indicates the timer has completed
    TimerStatusComplete TimerStatus = "complete"
)
```

## Creating a Timer

The Timers client provides several helper methods for creating different types of timers:

### Creating a One-Time Timer

```go
// Create a one-time timer that runs at a specific time
startTime := time.Now().Add(1 * time.Hour)

// Create a web callback
webCallback := timers.CreateWebCallback(
    "https://example.com/webhook",
    "POST",
    map[string]string{
        "Content-Type": "application/json",
    },
    nil,
)

// Create the timer
timer, err := client.CreateOnceTimer(
    ctx,
    "My One-Time Timer",
    startTime,
    webCallback,
    map[string]interface{}{
        "description": "This is a one-time timer",
    },
)
if err != nil {
    // Handle error
}

fmt.Printf("Created timer: %s\n", timer.ID)
```

### Creating a Recurring Timer

```go
// Create a recurring timer that runs every day for one week
startTime := time.Now().Add(12 * time.Hour)
endTime := time.Now().Add(7 * 24 * time.Hour)

// Create a flow callback
flowCallback := timers.CreateFlowCallback(
    "12345678-1234-1234-1234-123456789012", // Flow ID
    "Daily Report", // Flow label
    map[string]interface{}{ // Flow input
        "report_type": "daily",
        "include_details": true,
    },
)

// Create the timer
timer, err := client.CreateRecurringTimer(
    ctx,
    "Daily Report Timer",
    startTime,
    "1 day", // Run every day
    &endTime,
    flowCallback,
    nil,
)
if err != nil {
    // Handle error
}

fmt.Printf("Created recurring timer: %s\n", timer.ID)
```

### Creating a Cron Timer

```go
// Create a cron timer that runs at 8:00 AM Monday-Friday
cronExpression := "0 8 * * 1-5"
timezone := "America/New_York"

// Create the timer
timer, err := client.CreateCronTimer(
    ctx,
    "Weekday Morning Timer",
    cronExpression,
    timezone,
    nil, // No end time
    callback,
    nil,
)
if err != nil {
    // Handle error
}

fmt.Printf("Created cron timer: %s\n", timer.ID)
```

### Creating a Timer Manually

For more control, you can create a timer manually:

```go
// Create a timer request
request := &timers.CreateTimerRequest{
    Name: "Custom Timer",
    Schedule: timers.Schedule{
        Type: string(timers.ScheduleTypeOnce),
        StartTime: &startTime,
    },
    Callback: timers.Callback{
        Type: string(timers.CallbackTypeWeb),
        URL: &url,
        Method: &method,
        Headers: headers,
        Body: &body,
    },
    Data: map[string]interface{}{
        "custom_field": "custom_value",
    },
}

// Create the timer
timer, err := client.CreateTimer(ctx, request)
if err != nil {
    // Handle error
}
```

## Retrieving a Timer

```go
// Get a timer by ID
timerID := "12345678-1234-1234-1234-123456789012"
timer, err := client.GetTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Timer: %s\n", timer.Name)
fmt.Printf("Status: %s\n", timer.Status)
fmt.Printf("Next run: %s\n", timer.NextDue.Format(time.RFC3339))
```

## Listing Timers

```go
// Create options for listing timers
limit := 10
options := &timers.ListTimersOptions{
    Limit: &limit,
    // Optional filters
    Status: &status,
    ScheduleType: &scheduleType,
    CallbackType: &callbackType,
}

// List timers
timerList, err := client.ListTimers(ctx, options)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d timers (total: %d)\n", len(timerList.Timers), timerList.Total)

// Process results
for _, timer := range timerList.Timers {
    fmt.Printf("Timer: %s (%s)\n", timer.Name, timer.ID)
    fmt.Printf("Status: %s\n", timer.Status)
    if timer.NextDue != nil {
        fmt.Printf("Next due: %s\n", timer.NextDue.Format(time.RFC3339))
    }
}

// Check if there are more timers
if timerList.HasNextPage {
    // Use the NextPage marker to get the next page
    nextMarker := timerList.NextPage
    nextOptions := &timers.ListTimersOptions{
        Limit: &limit,
        Marker: nextMarker,
    }
    // Get next page...
}
```

## Updating a Timer

```go
// Create an update request
newName := "Updated Timer Name"
request := &timers.UpdateTimerRequest{
    Name: &newName,
    // Update other fields as needed
    Data: map[string]interface{}{
        "updated_at": time.Now(),
        "modified_by": "user123",
    },
}

// Update the timer
updatedTimer, err := client.UpdateTimer(ctx, timerID, request)
if err != nil {
    // Handle error
}

fmt.Printf("Updated timer: %s\n", updatedTimer.Name)
```

## Pausing and Resuming a Timer

```go
// Pause a timer
pausedTimer, err := client.PauseTimer(ctx, timerID)
if err != nil {
    // Handle error
}
fmt.Printf("Timer paused: %s\n", pausedTimer.Status)

// Resume a timer
resumedTimer, err := client.ResumeTimer(ctx, timerID)
if err != nil {
    // Handle error
}
fmt.Printf("Timer resumed: %s\n", resumedTimer.Status)
```

## Manually Running a Timer

```go
// Manually trigger a timer run
run, err := client.RunTimer(ctx, timerID)
if err != nil {
    // Handle error
}
fmt.Printf("Manual run initiated: %s\n", run.ID)
fmt.Printf("Status: %s\n", run.Status)
```

## Deleting a Timer

```go
// Delete a timer
err := client.DeleteTimer(ctx, timerID)
if err != nil {
    // Handle error
}
fmt.Println("Timer deleted successfully")
```

## Error Handling

Timer operations can return the following types of errors:

- Validation errors (invalid timer ID, missing required fields)
- Authentication errors (insufficient permissions)
- Resource not found errors (timer doesn't exist)
- API communication errors

Example error handling:

```go
timer, err := client.GetTimer(ctx, timerID)
if err != nil {
    if strings.Contains(err.Error(), "404") {
        // Timer not found
        fmt.Printf("Timer %s does not exist\n", timerID)
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied
        fmt.Println("You don't have permission to access this timer")
    } else {
        // Other error
        fmt.Printf("Error retrieving timer: %v\n", err)
    }
}
```