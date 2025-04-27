// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Please set the GLOBUS_ACCESS_TOKEN environment variable")
		os.Exit(1)
	}

	// Get flow ID for flow callback example
	flowID := os.Getenv("GLOBUS_FLOW_ID")
	if flowID == "" {
		fmt.Println("No GLOBUS_FLOW_ID environment variable set, flow callback example will be skipped")
	}

	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment()

	// Create a new Timers client with the access token
	timersClient := config.NewTimersClient(accessToken)

	// Create context
	ctx := context.Background()

	// Get information about the current user
	user, err := timersClient.GetCurrentUser(ctx)
	if err != nil {
		fmt.Printf("Error getting user information: %v\n", err)
	} else {
		fmt.Printf("Current user: %s (ID: %s)\n", user.Username, user.ID)
	}

	// Example 1: Create a one-time timer with a web callback
	fmt.Println("\n=== Example 1: One-Time Timer with Web Callback ===")
	
	// Create the timer to run 5 minutes from now
	startTime := time.Now().Add(5 * time.Minute)
	
	// Set up a webhook URL where the timer will send a notification
	// In a real application, this would be your server's URL
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
		fmt.Printf("Error creating one-time timer: %v\n", err)
	} else {
		fmt.Printf("Created one-time timer with ID: %s\n", webTimer.ID)
		fmt.Printf("Timer will run at: %s\n", webTimer.NextDue.Format(time.RFC3339))
		
		// Clean up the timer after demo
		defer func() {
			fmt.Printf("Cleaning up timer: %s\n", webTimer.ID)
			err := timersClient.DeleteTimer(ctx, webTimer.ID)
			if err != nil {
				fmt.Printf("Error deleting timer: %v\n", err)
			}
		}()
	}
	
	// Example 2: Create a recurring timer
	fmt.Println("\n=== Example 2: Recurring Timer ===")
	
	// Create a recurring timer that runs every hour
	recurringStartTime := time.Now().Add(1 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour) // Runs for 24 hours
	
	// Create the webhook callback
	recurringCallback := timers.CreateWebCallback(
		"https://httpbin.org/post",
		"POST",
		map[string]string{
			"Content-Type": "application/json",
		},
		nil,
	)
	
	// Create the timer
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
	
	if err != nil {
		fmt.Printf("Error creating recurring timer: %v\n", err)
	} else {
		fmt.Printf("Created recurring timer with ID: %s\n", recurringTimer.ID)
		fmt.Printf("Timer will first run at: %s\n", recurringTimer.NextDue.Format(time.RFC3339))
		fmt.Printf("Timer will end at: %s\n", endTime.Format(time.RFC3339))
		
		// Clean up the timer after demo
		defer func() {
			fmt.Printf("Cleaning up timer: %s\n", recurringTimer.ID)
			err := timersClient.DeleteTimer(ctx, recurringTimer.ID)
			if err != nil {
				fmt.Printf("Error deleting timer: %v\n", err)
			}
		}()
	}
	
	// Example 3: Create a timer with flow callback (if flow ID provided)
	if flowID != "" {
		fmt.Println("\n=== Example 3: Flow Callback Timer ===")
		
		// Create a timer to run a flow
		flowStartTime := time.Now().Add(10 * time.Minute)
		
		// Create flow callback
		flowCallback := timers.CreateFlowCallback(
			flowID,
			"Triggered by Globus Go SDK", // Label for the flow run
			map[string]interface{}{ // Flow input
				"message": "Hello from Timers API",
				"source": "Globus Go SDK Example",
			},
		)
		
		// Create the timer
		flowTimer, err := timersClient.CreateOnceTimer(
			ctx,
			"Example Flow Callback",
			flowStartTime,
			flowCallback,
			nil,
		)
		
		if err != nil {
			fmt.Printf("Error creating flow timer: %v\n", err)
		} else {
			fmt.Printf("Created flow timer with ID: %s\n", flowTimer.ID)
			fmt.Printf("Flow will run at: %s\n", flowTimer.NextDue.Format(time.RFC3339))
			
			// Clean up the timer after demo
			defer func() {
				fmt.Printf("Cleaning up timer: %s\n", flowTimer.ID)
				err := timersClient.DeleteTimer(ctx, flowTimer.ID)
				if err != nil {
					fmt.Printf("Error deleting timer: %v\n", err)
				}
			}()
		}
	}
	
	// Example 4: List timers
	fmt.Println("\n=== Example 4: List Timers ===")
	
	// Set options to limit results
	limit := 10
	listOptions := &timers.ListTimersOptions{
		Limit: &limit,
	}
	
	// List timers
	timerList, err := timersClient.ListTimers(ctx, listOptions)
	if err != nil {
		fmt.Printf("Error listing timers: %v\n", err)
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
			fmt.Println()
		}
	}
	
	// Example 5: Pause and resume timer
	if webTimer != nil {
		fmt.Println("\n=== Example 5: Pause and Resume Timer ===")
		
		// Pause the timer
		pausedTimer, err := timersClient.PauseTimer(ctx, webTimer.ID)
		if err != nil {
			fmt.Printf("Error pausing timer: %v\n", err)
		} else {
			fmt.Printf("Paused timer %s\n", pausedTimer.ID)
			fmt.Printf("Timer status: %s\n", pausedTimer.Status)
		}
		
		// Get the timer to verify status
		timer, err := timersClient.GetTimer(ctx, webTimer.ID)
		if err != nil {
			fmt.Printf("Error getting timer: %v\n", err)
		} else {
			fmt.Printf("Timer status after pause: %s\n", timer.Status)
		}
		
		// Resume the timer
		resumedTimer, err := timersClient.ResumeTimer(ctx, webTimer.ID)
		if err != nil {
			fmt.Printf("Error resuming timer: %v\n", err)
		} else {
			fmt.Printf("Resumed timer %s\n", resumedTimer.ID)
			fmt.Printf("Timer status: %s\n", resumedTimer.Status)
		}
	}
	
	fmt.Println("\nCleanup will happen now...")
}