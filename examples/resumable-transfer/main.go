// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Parse command-line arguments
	sourceEndpointID := flag.String("source", "", "Source endpoint ID")
	sourcePath := flag.String("source-path", "", "Source path")
	destEndpointID := flag.String("dest", "", "Destination endpoint ID")
	destPath := flag.String("dest-path", "", "Destination path")
	accessToken := flag.String("token", "", "Globus access token")
	resumeID := flag.String("resume", "", "Checkpoint ID to resume")
	batchSize := flag.Int("batch-size", 100, "Batch size for transfers")
	list := flag.Bool("list", false, "List available checkpoints")
	cancelID := flag.String("cancel", "", "Cancel transfer with the given checkpoint ID")
	flag.Parse()

	// Set up logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Resumable Transfer Example")

	// Get access token from environment if not provided
	if *accessToken == "" {
		*accessToken = os.Getenv("GLOBUS_ACCESS_TOKEN")
		if *accessToken == "" {
			log.Fatal("Access token is required. Use --token or set GLOBUS_ACCESS_TOKEN environment variable.")
		}
	}

	// Create SDK configuration
	config := pkg.NewConfigFromEnvironment()

	// Create transfer client
	transferClient := config.NewTransferClient(*accessToken)

	// Create context with cancellation
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down gracefully...")
		cancelCtx()
	}()

	// List checkpoints
	if *list {
		checkpoints, err := transferClient.ListTransferCheckpoints(ctx)
		if err != nil {
			log.Fatalf("Failed to list checkpoints: %v", err)
		}

		if len(checkpoints) == 0 {
			fmt.Println("No checkpoints found.")
			return
		}

		fmt.Println("Available checkpoints:")
		for i, id := range checkpoints {
			// Get checkpoint details
			state, err := transferClient.GetResumableTransferStatus(ctx, id)
			if err != nil {
				fmt.Printf("%d. %s (Error: %v)\n", i+1, id, err)
				continue
			}

			duration := state.TaskInfo.LastUpdated.Sub(state.TaskInfo.StartTime)
			fmt.Printf("%d. ID: %s\n", i+1, id)
			fmt.Printf("   From: %s:%s\n", state.TaskInfo.SourceEndpointID, state.TaskInfo.SourceBasePath)
			fmt.Printf("   To: %s:%s\n", state.TaskInfo.DestinationEndpointID, state.TaskInfo.DestinationBasePath)
			fmt.Printf("   Started: %s (Running for %s)\n", state.TaskInfo.StartTime.Format("2006-01-02 15:04:05"), duration)
			fmt.Printf("   Status: %d/%d files completed, %d failed, %d pending\n", 
				state.Stats.CompletedItems, state.Stats.TotalItems, state.Stats.FailedItems, state.Stats.RemainingItems)
			fmt.Println()
		}

		return
	}

	// Cancel a transfer
	if *cancelID != "" {
		if err := transferClient.CancelResumableTransfer(ctx, *cancelID); err != nil {
			log.Fatalf("Failed to cancel transfer: %v", err)
		}
		fmt.Printf("Transfer with checkpoint ID %s has been cancelled.\n", *cancelID)
		return
	}

	// Resuming a transfer
	if *resumeID != "" {
		fmt.Printf("Resuming transfer with checkpoint ID: %s\n", *resumeID)

		// Set up progress callback
		options := transfer.DefaultResumableTransferOptions()
		options.BatchSize = *batchSize
		options.ProgressCallback = func(state *transfer.CheckpointState) {
			fmt.Printf("\rProgress: %d/%d files completed (%d%%), %d failed", 
				state.Stats.CompletedItems, 
				state.Stats.TotalItems,
				int(float64(state.Stats.CompletedItems)/float64(state.Stats.TotalItems)*100),
				state.Stats.FailedItems)
		}

		// Resume the transfer
		result, err := transferClient.ResumeResumableTransfer(ctx, *resumeID, options)
		if err != nil {
			log.Fatalf("Failed to resume transfer: %v", err)
		}

		// Print results
		fmt.Println("\nTransfer completed!")
		fmt.Printf("Completed Items: %d\n", result.CompletedItems)
		fmt.Printf("Failed Items: %d\n", result.FailedItems)
		fmt.Printf("Duration: %s\n", result.Duration)

		return
	}

	// Starting a new transfer
	if *sourceEndpointID == "" || *sourcePath == "" || *destEndpointID == "" || *destPath == "" {
		log.Fatal("Source endpoint, source path, destination endpoint, and destination path are required for new transfers.")
	}

	fmt.Printf("Starting new resumable transfer from %s:%s to %s:%s\n", 
		*sourceEndpointID, *sourcePath, *destEndpointID, *destPath)

	// Set up options
	options := transfer.DefaultResumableTransferOptions()
	options.BatchSize = *batchSize
	options.ProgressCallback = func(state *transfer.CheckpointState) {
		fmt.Printf("\rDiscovering files: %d found so far...", state.Stats.TotalItems)
	}

	// Create the transfer
	checkpointID, err := transferClient.SubmitResumableTransfer(
		ctx,
		*sourceEndpointID, *sourcePath,
		*destEndpointID, *destPath,
		options,
	)
	if err != nil {
		log.Fatalf("Failed to create transfer: %v", err)
	}

	fmt.Printf("\nTransfer created with checkpoint ID: %s\n", checkpointID)
	fmt.Println("You can resume this transfer later with:")
	fmt.Printf("  go run main.go --resume %s\n", checkpointID)

	// Start the transfer immediately
	options.ProgressCallback = func(state *transfer.CheckpointState) {
		fmt.Printf("\rProgress: %d/%d files completed (%d%%), %d failed", 
			state.Stats.CompletedItems, 
			state.Stats.TotalItems,
			int(float64(state.Stats.CompletedItems)/float64(state.Stats.TotalItems)*100),
			state.Stats.FailedItems)
	}

	result, err := transferClient.ResumeResumableTransfer(ctx, checkpointID, options)
	if err != nil {
		log.Fatalf("Failed to start transfer: %v", err)
	}

	// Print results
	fmt.Println("\nTransfer completed!")
	fmt.Printf("Completed Items: %d\n", result.CompletedItems)
	fmt.Printf("Failed Items: %d\n", result.FailedItems)
	fmt.Printf("Duration: %s\n", result.Duration)
}