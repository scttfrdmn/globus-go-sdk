// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors

// Package main demonstrates creating a flow timer with the Timers service.
//
// At 3.65.0 a flow timer is created with POST /v2/timer using a document of
// {timer_type:"flow", flow_id, name, schedule, body}. The Go helper
// CreateFlowTimer builds and submits that document.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
)

func main() {
	accessToken := os.Getenv("GLOBUS_SDK_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_SDK_ACCESS_TOKEN environment variable not set")
	}
	flowID := os.Getenv("FLOW_ID")
	if flowID == "" {
		log.Fatal("FLOW_ID environment variable not set")
	}

	client, err := timers.NewClient(timers.WithAccessToken(accessToken))
	if err != nil {
		log.Fatalf("Failed to create timers client: %v", err)
	}
	ctx := context.Background()

	// One-time flow timer.
	onceTimer, err := client.CreateFlowTimer(
		ctx,
		"My One-Time Flow Timer",
		flowID,
		timers.NewOnceSchedule(time.Now().Add(time.Hour).Format(time.RFC3339)),
		map[string]interface{}{"source": "/path/to/source", "dest": "/path/to/destination"},
	)
	if err != nil {
		log.Fatalf("Failed to create one-time flow timer: %v", err)
	}
	fmt.Printf("Created one-time flow timer: %s\n", onceTimer.JobID)

	// Recurring flow timer (daily).
	recurringTimer, err := client.CreateFlowTimer(
		ctx,
		"Daily Flow Timer",
		flowID,
		timers.NewRecurringSchedule(86400, time.Now().Format(time.RFC3339), nil),
		map[string]interface{}{"source": "/daily/source", "dest": "/daily/dest"},
	)
	if err != nil {
		log.Fatalf("Failed to create recurring flow timer: %v", err)
	}
	fmt.Printf("Created recurring flow timer: %s\n", recurringTimer.JobID)

	// List timers.
	timerList, err := client.ListTimers(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list timers: %v", err)
	}
	fmt.Printf("Found %d timers\n", len(timerList.Timers))

	fmt.Println("Done.")
}
