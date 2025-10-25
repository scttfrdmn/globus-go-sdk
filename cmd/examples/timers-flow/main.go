// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package main demonstrates using FlowTimer to create timers that run Globus Flows
//
// This example showcases the FlowTimer helper added in v3.65.0, which simplifies
// creating timers that trigger flow executions.
//
// The FlowTimer feature matches the Python SDK v3.65.0 FlowTimer payload class.
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
	// Get access token from environment
	accessToken := os.Getenv("GLOBUS_SDK_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_SDK_ACCESS_TOKEN environment variable not set")
	}

	// Get flow ID and scope from environment
	flowID := os.Getenv("FLOW_ID")
	if flowID == "" {
		log.Fatal("FLOW_ID environment variable not set")
	}

	flowScope := os.Getenv("FLOW_SCOPE")
	if flowScope == "" {
		log.Fatal("FLOW_SCOPE environment variable not set")
	}

	// Create timers client
	client, err := timers.NewClient(
		timers.WithAccessToken(accessToken),
	)
	if err != nil {
		log.Fatalf("Failed to create timers client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Create a one-time FlowTimer
	fmt.Println("\n=== Example 1: One-Time FlowTimer ===")

	onceFlowTimer := &timers.FlowTimer{
		FlowID:    flowID,
		FlowScope: flowScope,
		FlowInput: map[string]interface{}{
			"source": "/path/to/source",
			"dest":   "/path/to/destination",
		},
		FlowLabel: "One-time data transfer",
	}

	startTime := time.Now().Add(1 * time.Hour)
	onceTimer, err := client.CreateFlowTimerOnce(
		ctx,
		"My One-Time Flow Timer",
		startTime,
		onceFlowTimer,
		map[string]interface{}{
			"user_data": "custom metadata",
		},
	)
	if err != nil {
		log.Fatalf("Failed to create one-time flow timer: %v", err)
	}

	fmt.Printf("✅ Created one-time flow timer: %s\n", onceTimer.ID)
	fmt.Printf("   Name: %s\n", onceTimer.Name)
	fmt.Printf("   Status: %s\n", onceTimer.Status)
	fmt.Printf("   Will run at: %s\n", startTime.Format(time.RFC3339))

	// Example 2: Create a recurring FlowTimer (daily)
	fmt.Println("\n=== Example 2: Recurring FlowTimer (Daily) ===")

	recurringFlowTimer := &timers.FlowTimer{
		FlowID:    flowID,
		FlowScope: flowScope,
		FlowInput: map[string]interface{}{
			"source": "/daily/backup/source",
			"dest":   "/daily/backup/destination",
		},
		FlowLabel: "Daily backup flow",
	}

	recurringStart := time.Now().Add(30 * time.Minute)
	endTime := recurringStart.Add(30 * 24 * time.Hour) // Run for 30 days

	recurringTimer, err := client.CreateFlowTimerRecurring(
		ctx,
		"Daily Backup Flow Timer",
		recurringStart,
		"P1D", // ISO 8601 duration: 1 day
		&endTime,
		recurringFlowTimer,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create recurring flow timer: %v", err)
	}

	fmt.Printf("✅ Created recurring flow timer: %s\n", recurringTimer.ID)
	fmt.Printf("   Name: %s\n", recurringTimer.Name)
	fmt.Printf("   Status: %s\n", recurringTimer.Status)
	fmt.Printf("   Interval: Every 24 hours\n")
	fmt.Printf("   Starts: %s\n", recurringStart.Format(time.RFC3339))
	fmt.Printf("   Ends: %s\n", endTime.Format(time.RFC3339))

	// Example 3: Create a cron-scheduled FlowTimer
	fmt.Println("\n=== Example 3: Cron-Scheduled FlowTimer ===")

	cronFlowTimer := &timers.FlowTimer{
		FlowID:    flowID,
		FlowScope: flowScope,
		FlowInput: map[string]interface{}{
			"source": "/weekly/report/source",
			"dest":   "/weekly/report/destination",
		},
		FlowLabel: "Weekly report flow",
	}

	// Run every Monday at 9:00 AM EST
	cronTimer, err := client.CreateFlowTimerCron(
		ctx,
		"Weekly Report Flow Timer",
		"0 9 * * 1", // Cron: Every Monday at 9:00 AM
		"America/New_York",
		nil, // No end time
		cronFlowTimer,
		map[string]interface{}{
			"report_type": "weekly_summary",
		},
	)
	if err != nil {
		log.Fatalf("Failed to create cron flow timer: %v", err)
	}

	fmt.Printf("✅ Created cron flow timer: %s\n", cronTimer.ID)
	fmt.Printf("   Name: %s\n", cronTimer.Name)
	fmt.Printf("   Status: %s\n", cronTimer.Status)
	fmt.Printf("   Schedule: Every Monday at 9:00 AM EST\n")

	// List all timers
	fmt.Println("\n=== Listing All Timers ===")

	listOptions := &timers.ListTimersOptions{
		CallbackType: stringPtr("flow"),
	}

	timerList, err := client.ListTimers(ctx, listOptions)
	if err != nil {
		log.Fatalf("Failed to list timers: %v", err)
	}

	fmt.Printf("Found %d flow-based timers:\n", len(timerList.Timers))
	for i, timer := range timerList.Timers {
		fmt.Printf("%d. %s (ID: %s, Status: %s)\n", i+1, timer.Name, timer.ID, timer.Status)
	}

	// Cleanup example (commented out - uncomment to actually delete)
	/*
	fmt.Println("\n=== Cleanup ===")

	for _, timer := range []*timers.Timer{onceTimer, recurringTimer, cronTimer} {
		if err := client.DeleteTimer(ctx, timer.ID); err != nil {
			log.Printf("Failed to delete timer %s: %v", timer.ID, err)
		} else {
			fmt.Printf("✅ Deleted timer: %s\n", timer.ID)
		}
	}
	*/

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("FlowTimer makes it easy to create timers that run Globus Flows!")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("  • One-time flow timer")
	fmt.Println("  • Recurring flow timer (ISO 8601 duration)")
	fmt.Println("  • Cron-scheduled flow timer")
	fmt.Println("  • Custom flow input and labels")
	fmt.Println("  • Timer listing and filtering")
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
