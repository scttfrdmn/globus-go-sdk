---
title: "Timers Service Quick Start"
weight: 70
---

# Timers Service Quick Start

This guide will help you get started with the Globus Timers service using the Go SDK. The Timers service allows you to schedule one-time or recurring actions, such as triggering Globus Flows or making HTTP requests.

## Setup

First, import the required packages and create a context:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
)

func main() {
    // Create a context
    ctx := context.Background()
    
    // Continue with the examples below...
}
```

## Creating a Timers Client

There are two main ways to create a Timers client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Timers client
timersClient, err := config.NewTimersClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create timers client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Timers client with options
timersClient, err := timers.NewClient(
    timers.WithAccessToken(os.Getenv("GLOBUS_ACCESS_TOKEN")),
    timers.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create timers client: %v", err)
}
```

## Creating Timers

The Timers service supports three types of schedules: one-time, recurring, and cron-based.

### Creating a One-Time Timer

A one-time timer runs at a specific time:

```go
// Set up a webhook callback
webhookURL := "https://your-webhook-url.com/hook"
webhookMethod := "POST"
webhookBody := `{"message": "Timer triggered"}`

// Create the callback configuration
webCallback := timers.CreateWebCallback(
    webhookURL,
    webhookMethod,
    map[string]string{
        "Content-Type": "application/json",
    },
    &webhookBody,
)

// Create a one-time timer to run 30 minutes from now
startTime := time.Now().Add(30 * time.Minute)
oneTimeTimer, err := timersClient.CreateOnceTimer(
    ctx,
    "My One-Time Timer",
    startTime,
    webCallback,
    map[string]interface{}{
        "description": "This timer sends a POST request once",
        "created_by": "Globus Go SDK Example",
    },
)

if err != nil {
    log.Fatalf("Failed to create one-time timer: %v", err)
}

fmt.Printf("Created one-time timer with ID: %s\n", oneTimeTimer.ID)
fmt.Printf("Timer will run at: %s\n", oneTimeTimer.NextDue.Format(time.RFC3339))
```

### Creating a Recurring Timer

A recurring timer runs at regular intervals:

```go
// Set up the recurring timer parameters
recurringStartTime := time.Now().Add(1 * time.Hour)
endTime := time.Now().Add(7 * 24 * time.Hour) // Runs for 7 days

// Create a recurring timer that runs every 4 hours
recurringTimer, err := timersClient.CreateRecurringTimer(
    ctx,
    "My Recurring Timer",
    recurringStartTime,
    "4 hours", // Run every 4 hours
    &endTime,  // Optional end time
    webCallback,
    map[string]interface{}{
        "description": "This timer runs every 4 hours for one week",
    },
)

if err != nil {
    log.Fatalf("Failed to create recurring timer: %v", err)
}

fmt.Printf("Created recurring timer with ID: %s\n", recurringTimer.ID)
fmt.Printf("Timer will first run at: %s\n", recurringTimer.NextDue.Format(time.RFC3339))
fmt.Printf("Timer will end at: %s\n", endTime.Format(time.RFC3339))
```

### Creating a Cron-Based Timer

A cron-based timer runs according to a cron expression:

```go
// Set the cron parameters
cronExpression := "0 9 * * 1-5" // Run at 9 AM on weekdays
timezone := "America/New_York"

// Create a cron-based timer
cronTimer, err := timersClient.CreateCronTimer(
    ctx,
    "My Cron Timer",
    cronExpression,
    timezone,
    nil, // No end time
    webCallback,
    map[string]interface{}{
        "description": "This timer runs at 9 AM Eastern Time on weekdays",
    },
)

if err != nil {
    log.Fatalf("Failed to create cron timer: %v", err)
}

fmt.Printf("Created cron timer with ID: %s\n", cronTimer.ID)
fmt.Printf("Timer will first run at: %s\n", cronTimer.NextDue.Format(time.RFC3339))
```

## Timer Callbacks

Timers can trigger different types of actions. The Timers service supports two callback types: web and flow.

### Web Callbacks

A web callback makes an HTTP request when the timer triggers:

```go
// Create a web callback configuration
webCallback := timers.CreateWebCallback(
    "https://your-api.example.com/endpoint",
    "POST",
    map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer your-token",
    },
    &webhookBody,
)
```

### Flow Callbacks

A flow callback triggers a Globus Flow when the timer triggers:

```go
// Create a flow callback configuration
flowID := "your-flow-id"
flowCallback := timers.CreateFlowCallback(
    flowID,
    "Triggered by Globus Go SDK", // Label for the flow run
    map[string]interface{}{       // Flow input
        "message": "Hello from Timers API",
        "source": "Globus Go SDK Example",
    },
)

// Create a timer with the flow callback
flowTimer, err := timersClient.CreateOnceTimer(
    ctx,
    "My Flow Timer",
    time.Now().Add(15 * time.Minute),
    flowCallback,
    nil,
)

if err != nil {
    log.Fatalf("Failed to create flow timer: %v", err)
}

fmt.Printf("Created flow timer with ID: %s\n", flowTimer.ID)
```

## Managing Timers

### Getting Timer Details

You can retrieve details about a specific timer:

```go
// Get details about a timer
timerID := "your-timer-id"
timer, err := timersClient.GetTimer(ctx, timerID)
if err != nil {
    log.Fatalf("Failed to get timer: %v", err)
}

fmt.Printf("Timer: %s (%s)\n", timer.Name, timer.ID)
fmt.Printf("Status: %s\n", timer.Status)
if timer.NextDue != nil {
    fmt.Printf("Next due: %s\n", timer.NextDue.Format(time.RFC3339))
}
if timer.LastRun != nil {
    fmt.Printf("Last run: %s\n", timer.LastRun.Format(time.RFC3339))
    fmt.Printf("Last run status: %s\n", timer.LastRunStatus)
}
```

### Listing Timers

You can list all your timers:

```go
// Set options to limit results
limit := 10
listOptions := &timers.ListTimersOptions{
    Limit: &limit,
}

// List timers
timerList, err := timersClient.ListTimers(ctx, listOptions)
if err != nil {
    log.Fatalf("Failed to list timers: %v", err)
}

fmt.Printf("Found %d timers (total: %d)\n", len(timerList.Timers), timerList.Total)

// Print timer details
for i, timer := range timerList.Timers {
    fmt.Printf("%d. %s (ID: %s)\n", i+1, timer.Name, timer.ID)
    fmt.Printf("   Status: %s\n", timer.Status)
    if timer.NextDue != nil {
        fmt.Printf("   Next due: %s\n", timer.NextDue.Format(time.RFC3339))
    }
    fmt.Printf("   Callback type: %s\n", timer.Callback.Type)
    fmt.Printf("   Schedule type: %s\n", timer.Schedule.Type)
}
```

### Using Pagination

If you have many timers, you can use pagination to retrieve them all:

```go
// Set initial options
options := &timers.ListTimersOptions{
    Limit: &limit,
}

// Iterate through all pages
for {
    timerList, err := timersClient.ListTimers(ctx, options)
    if err != nil {
        log.Fatalf("Failed to list timers: %v", err)
    }
    
    // Process the current page of timers
    for _, timer := range timerList.Timers {
        fmt.Printf("- %s (%s)\n", timer.Name, timer.ID)
    }
    
    // Check if there are more pages
    if !timerList.HasNextPage {
        break
    }
    
    // Update the marker for the next page
    options.Marker = timerList.NextPage
}
```

### Updating a Timer

You can update various properties of an existing timer:

```go
// Create an update request
updateRequest := &timers.UpdateTimerRequest{
    Name: &updatedName,
    Data: map[string]interface{}{
        "description": "Updated timer description",
        "updated_at": time.Now().Format(time.RFC3339),
    },
}

// Update the timer
updatedTimer, err := timersClient.UpdateTimer(ctx, timerID, updateRequest)
if err != nil {
    log.Fatalf("Failed to update timer: %v", err)
}

fmt.Printf("Updated timer: %s\n", updatedTimer.Name)
```

### Pausing a Timer

You can pause a timer to temporarily stop it from running:

```go
// Pause a timer
pausedTimer, err := timersClient.PauseTimer(ctx, timerID)
if err != nil {
    log.Fatalf("Failed to pause timer: %v", err)
}

fmt.Printf("Paused timer %s\n", pausedTimer.ID)
fmt.Printf("Timer status: %s\n", pausedTimer.Status)
```

### Resuming a Timer

You can resume a paused timer:

```go
// Resume a timer
resumedTimer, err := timersClient.ResumeTimer(ctx, timerID)
if err != nil {
    log.Fatalf("Failed to resume timer: %v", err)
}

fmt.Printf("Resumed timer %s\n", resumedTimer.ID)
fmt.Printf("Timer status: %s\n", resumedTimer.Status)
```

### Manually Running a Timer

You can trigger a timer to run immediately:

```go
// Manually run a timer
run, err := timersClient.RunTimer(ctx, timerID)
if err != nil {
    log.Fatalf("Failed to run timer: %v", err)
}

fmt.Printf("Triggered timer run: %s\n", run.ID)
fmt.Printf("Run status: %s\n", run.Status)
```

### Deleting a Timer

You can delete a timer when it's no longer needed:

```go
// Delete a timer
err = timersClient.DeleteTimer(ctx, timerID)
if err != nil {
    log.Fatalf("Failed to delete timer: %v", err)
}

fmt.Println("Timer deleted successfully")
```

## Working with Timer Runs

### Getting Run Details

You can retrieve details about a specific timer run:

```go
// Get details about a timer run
runID := "your-run-id"
run, err := timersClient.GetRun(ctx, timerID, runID)
if err != nil {
    log.Fatalf("Failed to get run: %v", err)
}

fmt.Printf("Run: %s\n", run.ID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Start time: %s\n", run.StartTime.Format(time.RFC3339))
if run.EndTime != nil {
    fmt.Printf("End time: %s\n", run.EndTime.Format(time.RFC3339))
}
if run.Result != nil {
    fmt.Printf("Result status: %s\n", run.Result.Status)
    if run.Result.StatusCode != nil {
        fmt.Printf("Status code: %d\n", *run.Result.StatusCode)
    }
    if run.Result.RunID != nil {
        fmt.Printf("Flow run ID: %s\n", *run.Result.RunID)
    }
}
```

### Listing Runs

You can list the run history for a timer:

```go
// Set options to limit results
limit := 10
listRunsOptions := &timers.ListRunsOptions{
    Limit: &limit,
}

// List runs for a timer
runList, err := timersClient.ListRuns(ctx, timerID, listRunsOptions)
if err != nil {
    log.Fatalf("Failed to list runs: %v", err)
}

fmt.Printf("Found %d runs (total: %d)\n", len(runList.Runs), runList.Total)

// Print run details
for i, run := range runList.Runs {
    fmt.Printf("%d. Run ID: %s\n", i+1, run.ID)
    fmt.Printf("   Status: %s\n", run.Status)
    fmt.Printf("   Start time: %s\n", run.StartTime.Format(time.RFC3339))
    if run.EndTime != nil {
        fmt.Printf("   End time: %s\n", run.EndTime.Format(time.RFC3339))
    }
    if run.Result != nil {
        fmt.Printf("   Result status: %s\n", run.Result.Status)
    }
}
```

### Filtering Runs by Date

You can filter runs by their start time:

```go
// Filter runs by date range
startAfter := time.Now().Add(-7 * 24 * time.Hour) // 1 week ago
startBefore := time.Now()

filterOptions := &timers.ListRunsOptions{
    StartAfter:  &startAfter,
    StartBefore: &startBefore,
}

// List runs in date range
filteredRuns, err := timersClient.ListRuns(ctx, timerID, filterOptions)
if err != nil {
    log.Fatalf("Failed to list filtered runs: %v", err)
}

fmt.Printf("Found %d runs in the last week\n", len(filteredRuns.Runs))
```

### Filtering Runs by Status

You can filter runs by their status:

```go
// Filter runs by status
status := "success"
statusOptions := &timers.ListRunsOptions{
    Status: &status,
}

// List successful runs
successfulRuns, err := timersClient.ListRuns(ctx, timerID, statusOptions)
if err != nil {
    log.Fatalf("Failed to list successful runs: %v", err)
}

fmt.Printf("Found %d successful runs\n", len(successfulRuns.Runs))
```

## Complete Example

Here's a complete example that creates a one-time timer with a web callback, pauses and resumes it, and then lists all timers:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
)

func main() {
    // Get access token from environment variable
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if accessToken == "" {
        log.Fatalf("Please set the GLOBUS_ACCESS_TOKEN environment variable")
    }

    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()

    // Create a new Timers client with the access token
    timersClient, err := config.NewTimersClient(accessToken)
    if err != nil {
        log.Fatalf("Error creating timers client: %v", err)
    }

    // Create context
    ctx := context.Background()

    // Get information about the current user
    user, err := timersClient.GetCurrentUser(ctx)
    if err != nil {
        log.Printf("Error getting user information: %v", err)
    } else {
        fmt.Printf("Current user: %s (ID: %s)\n", user.Username, user.ID)
    }

    // Create a one-time timer with a web callback
    fmt.Println("\n=== Creating a One-Time Timer with Web Callback ===")
    
    // Create the timer to run 5 minutes from now
    startTime := time.Now().Add(5 * time.Minute)
    
    // Set up a webhook URL
    webhookURL := "https://httpbin.org/post"
    webhookMethod := "POST"
    webhookBody := `{"message": "Timer triggered"}`
    
    // Create callback configuration
    webCallback := timers.CreateWebCallback(
        webhookURL, 
        webhookMethod, 
        map[string]string{
            "Content-Type": "application/json",
        },
        &webhookBody,
    )
    
    // Create the timer
    webTimer, err := timersClient.CreateOnceTimer(
        ctx,
        "Example One-Time Web Callback",
        startTime,
        webCallback,
        map[string]interface{}{
            "description": "This timer sends a POST request to httpbin.org",
            "created_by": "Globus Go SDK Example",
        },
    )
    
    if err != nil {
        log.Fatalf("Error creating one-time timer: %v", err)
    }
    
    fmt.Printf("Created one-time timer with ID: %s\n", webTimer.ID)
    fmt.Printf("Timer will run at: %s\n", webTimer.NextDue.Format(time.RFC3339))
    
    // Pause the timer
    fmt.Println("\n=== Pausing the Timer ===")
    pausedTimer, err := timersClient.PauseTimer(ctx, webTimer.ID)
    if err != nil {
        log.Printf("Error pausing timer: %v", err)
    } else {
        fmt.Printf("Paused timer %s\n", pausedTimer.ID)
        fmt.Printf("Timer status: %s\n", pausedTimer.Status)
    }
    
    // Resume the timer
    fmt.Println("\n=== Resuming the Timer ===")
    resumedTimer, err := timersClient.ResumeTimer(ctx, webTimer.ID)
    if err != nil {
        log.Printf("Error resuming timer: %v", err)
    } else {
        fmt.Printf("Resumed timer %s\n", resumedTimer.ID)
        fmt.Printf("Timer status: %s\n", resumedTimer.Status)
    }
    
    // List all timers
    fmt.Println("\n=== Listing All Timers ===")
    
    // Set options to limit results
    limit := 10
    listOptions := &timers.ListTimersOptions{
        Limit: &limit,
    }
    
    // List timers
    timerList, err := timersClient.ListTimers(ctx, listOptions)
    if err != nil {
        log.Printf("Error listing timers: %v", err)
    } else {
        fmt.Printf("Found %d timers (total: %d)\n", len(timerList.Timers), timerList.Total)
        
        // Print timer details
        for i, timer := range timerList.Timers {
            fmt.Printf("%d. %s (ID: %s)\n", i+1, timer.Name, timer.ID)
            fmt.Printf("   Status: %s\n", timer.Status)
            if timer.NextDue != nil {
                fmt.Printf("   Next due: %s\n", timer.NextDue.Format(time.RFC3339))
            }
            fmt.Printf("   Callback type: %s\n", timer.Callback.Type)
            fmt.Printf("   Schedule type: %s\n", timer.Schedule.Type)
        }
    }
    
    // Clean up the timer
    fmt.Println("\n=== Cleaning Up ===")
    err = timersClient.DeleteTimer(ctx, webTimer.ID)
    if err != nil {
        log.Printf("Error deleting timer: %v", err)
    } else {
        fmt.Println("Timer deleted successfully")
    }
}
```

## Error Handling

The Timers service methods return errors when operations fail. You can handle these errors by checking the error message:

```go
// Try to get a non-existent timer
_, err = timersClient.GetTimer(ctx, "non-existent-timer-id")
if err != nil {
    if err.Error() == "request failed with status 404: " {
        fmt.Println("Timer not found - check the timer ID")
    } else if err.Error() == "request failed with status 403: " {
        fmt.Println("Permission denied - check your access token and permissions")
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Next Steps

Now that you understand the basics of the Timers service, you can:

1. **Schedule Recurring Tasks**: Create timers that execute tasks on a regular schedule
2. **Set Up Automated Workflows**: Use flow callbacks to trigger complex workflows at specific times
3. **Build Integration Systems**: Use web callbacks to notify external systems at scheduled times
4. **Create Time-Based Controls**: Implement time-based access controls or data processing

For more details, check out the [Timers Service API Reference](/docs/reference/timers/) documentation.