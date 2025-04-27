<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Timers API Example

This example demonstrates how to use the Timers service in the Globus Go SDK to schedule and manage tasks.

## Overview

Globus Timers is a scheduling service that allows you to trigger actions at specific times or intervals. It supports one-time and recurring schedules, and can trigger either web callbacks (HTTP requests) or Globus Flow executions.

This example demonstrates:

1. Creating one-time timers with web callbacks
2. Creating recurring timers
3. Creating timers that trigger Globus Flows
4. Listing active timers
5. Pausing and resuming timers

## Prerequisites

To run this example, you need:

1. A Globus account
2. An access token with the Timers scope
3. (Optional) A Globus Flow ID if you want to try the flow callback example

## Getting Started

1. Set the required environment variables:

```bash
export GLOBUS_ACCESS_TOKEN=your-access-token
export GLOBUS_FLOW_ID=your-flow-id  # Optional
```

2. Run the example:

```bash
go run main.go
```

## Example Details

### One-Time Timer with Web Callback

Creates a timer that will trigger once at a specified time. When triggered, the timer will make an HTTP POST request to the specified URL with a JSON payload.

```go
webCallback := timers.CreateWebCallback(
    "https://httpbin.org/post", 
    "POST", 
    map[string]string{
        "Content-Type": "application/json",
    },
    &webhookBody,
)

webTimer, err := timersClient.CreateOnceTimer(
    ctx,
    "Example One-Time Web Callback",
    startTime,
    webCallback,
    map[string]interface{}{
        "description": "This timer sends a POST request to httpbin.org",
    },
)
```

### Recurring Timer

Creates a timer that will trigger repeatedly at the specified interval. In this example, the timer will run every hour for a day.

```go
recurringTimer, err := timersClient.CreateRecurringTimer(
    ctx,
    "Example Recurring Timer",
    recurringStartTime,
    "1 hour", // Run every hour
    &endTime,
    recurringCallback,
    map[string]interface{}{
        "description": "This timer runs every hour for one day",
    },
)
```

### Flow Callback Timer

Creates a timer that will trigger a Globus Flow execution. When triggered, the timer will start a flow run with the specified flow ID, label, and input data.

```go
flowCallback := timers.CreateFlowCallback(
    flowID,
    "Triggered by Globus Go SDK", // Label for the flow run
    map[string]interface{}{ // Flow input
        "message": "Hello from Timers API",
        "source": "Globus Go SDK Example",
    },
)

flowTimer, err := timersClient.CreateOnceTimer(
    ctx,
    "Example Flow Callback",
    flowStartTime,
    flowCallback,
    nil,
)
```

### Timer Management

The example also demonstrates how to:

- List active timers with optional filtering
- Pause a timer to temporarily prevent it from running
- Resume a paused timer
- Delete timers when they are no longer needed

## Notes

- The example automatically deletes the timers it creates at the end of execution
- HTTP webhook callbacks need a publicly accessible endpoint to receive the requests
- Flow callbacks require a flow that is owned by or shared with your Globus account

## Additional Resources

- [Globus Timers API Documentation](https://docs.globus.org/api/timers/)
- [Globus Go SDK Documentation](../README.md)