// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
)

func main() {
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Please set the GLOBUS_ACCESS_TOKEN environment variable")
		os.Exit(1)
	}

	config := pkg.NewConfigFromEnvironment()
	timersClient, err := config.NewTimersClient(accessToken)
	if err != nil {
		fmt.Printf("Error creating timers client: %v\n", err)
		os.Exit(1)
	}
	ctx := context.Background()

	// List the caller's timers (GET /jobs/).
	list, err := timersClient.ListTimers(ctx, nil)
	if err != nil {
		fmt.Printf("Error listing timers: %v\n", err)
	} else {
		fmt.Printf("You have %d timers\n", len(list.Timers))
	}

	// Create a recurring transfer timer (POST /v2/timer). The body is a
	// TransferData document; a minimal placeholder is shown here.
	schedule := timers.NewRecurringSchedule(3600, time.Now().Format(time.RFC3339), nil)
	transferBody := map[string]interface{}{
		"DATA_TYPE":            "transfer",
		"source_endpoint":      os.Getenv("GLOBUS_SOURCE_ENDPOINT_ID"),
		"destination_endpoint": os.Getenv("GLOBUS_DEST_ENDPOINT_ID"),
	}
	timer, err := timersClient.CreateTransferTimer(ctx, "Example Transfer Timer", schedule, transferBody)
	if err != nil {
		fmt.Printf("Error creating transfer timer: %v\n", err)
		return
	}
	fmt.Printf("Created timer: %s\n", timer.JobID)

	// Pause and resume it.
	if err := timersClient.PauseTimer(ctx, timer.JobID); err != nil {
		fmt.Printf("Error pausing timer: %v\n", err)
	}
	if err := timersClient.ResumeTimer(ctx, timer.JobID, nil); err != nil {
		fmt.Printf("Error resuming timer: %v\n", err)
	}

	// Clean up.
	if err := timersClient.DeleteTimer(ctx, timer.JobID); err != nil {
		fmt.Printf("Error deleting timer: %v\n", err)
	}
	fmt.Println("Done.")
}
