// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/metrics"
)

func main() {
	// Parse command line flags
	var (
		storageDirFlag   = flag.String("storage-dir", "", "Directory for metrics storage")
		loadExistingFlag = flag.Bool("load-existing", false, "Load existing metrics from storage")
		historyFlag      = flag.Bool("history", false, "Show historical transfers and exit")
	)
	flag.Parse()

	// Create a monitor
	monitor := metrics.NewPerformanceMonitor()

	// Set up metrics storage if specified
	var storage metrics.MetricsStorage
	if *storageDirFlag != "" {
		// Create storage directory if it doesn't exist
		storageDir := *storageDirFlag
		if storageDir == "" {
			// Default to a metrics directory in the user's home directory
			home, err := os.UserHomeDir()
			if err == nil {
				storageDir = filepath.Join(home, ".globus-sdk", "metrics")
			} else {
				// Fallback to current directory
				storageDir = "metrics"
			}
		}

		// Create the storage directory
		err := os.MkdirAll(storageDir, 0755)
		if err != nil {
			fmt.Printf("Warning: Failed to create metrics storage directory: %v\n", err)
		} else {
			// Create a file-based metrics storage
			var storageErr error
			storage, storageErr = metrics.NewFileMetricsStorage(storageDir)
			if storageErr != nil {
				fmt.Printf("Warning: Failed to initialize metrics storage: %v\n", storageErr)
				storage = nil
			} else {
				fmt.Printf("Using metrics storage: %s\n", storageDir)

				// Configure auto-save for storage
				monitor.WithStorage(&metrics.StorageConfig{
					Storage:      storage,
					SaveInterval: 5 * time.Second,
					AutoSave:     true,
					AutoCleanup:  true,
					CleanupAge:   7 * 24 * time.Hour, // 7 days
				})

				// Load existing metrics if requested
				if *loadExistingFlag {
					err := monitor.LoadAllMetrics(storage)
					if err != nil {
						fmt.Printf("Warning: Failed to load existing metrics: %v\n", err)
					} else {
						fmt.Println("Loaded existing metrics from storage")
					}
				}
			}
		}
	}

	// Create a reporter
	reporter := metrics.NewTextReporter()

	// If in history mode, show historical transfers and exit
	if *historyFlag && storage != nil {
		fmt.Println("=== Historical Transfers ===")

		// List all transfer IDs from storage
		ids, err := storage.ListTransferIDs()
		if err != nil {
			fmt.Printf("Error listing transfers: %v\n", err)
			return
		}

		if len(ids) == 0 {
			fmt.Println("No historical transfers found.")
			return
		}

		fmt.Printf("Found %d historical transfers:\n\n", len(ids))

		// Load and display each transfer
		for _, id := range ids {
			storedMetrics, err := storage.RetrieveMetrics(id)
			if err != nil {
				fmt.Printf("Error loading metrics for %s: %v\n", id, err)
				continue
			}

			fmt.Printf("==== Transfer: %s ====\n", id)
			reporter.ReportSummary(os.Stdout, storedMetrics)
			fmt.Println()
		}

		return
	}

	// Start 3 simulated transfers
	transferIDs := []string{
		"transfer-1",
		"transfer-2",
		"transfer-3",
	}

	// Total bytes for each transfer
	sizes := []int64{
		50 * 1024 * 1024,       // 50 MB
		150 * 1024 * 1024,      // 150 MB
		1 * 1024 * 1024 * 1024, // 1 GB
	}

	// Names for each transfer
	labels := []string{
		"Small Transfer",
		"Medium Transfer",
		"Large Transfer",
	}

	// Start progress bars
	progressBars := make([]*metrics.ProgressBar, len(transferIDs))

	// Start the transfers
	for i, id := range transferIDs {
		// Start monitoring this transfer
		monitor.StartMonitoring(
			id,                          // Transfer ID
			fmt.Sprintf("task-%d", i+1), // Task ID
			"source-endpoint",           // Source endpoint
			"dest-endpoint",             // Destination endpoint
			labels[i],                   // Label
		)

		// Set the total bytes
		monitor.SetTotalBytes(id, sizes[i])
		monitor.SetTotalFiles(id, int64(10*(i+1)))

		// Create a progress bar
		progressBars[i] = metrics.NewProgressBar(
			os.Stdout,
			sizes[i],
			metrics.WithWidth(50),
			metrics.WithRefreshRate(200*time.Millisecond),
			metrics.WithMessage(labels[i]),
		)
		progressBars[i].Start()

		// Start a goroutine to simulate the transfer
		go simulateTransfer(id, sizes[i], monitor, progressBars[i])
	}

	// Display summary statistics every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Done channel for coordinating completion
	done := make(chan bool)

	// Track completed transfers
	completed := 0

	// Display transfer statistics periodically
	go func() {
		for {
			select {
			case <-ticker.C:
				// Clear the screen
				fmt.Print("\033[H\033[2J")
				fmt.Println("=== Transfer Statistics ===")
				if storage != nil {
					fmt.Println("Metrics are being stored automatically")
				}
				fmt.Println()

				// Get list of active transfers
				activeTransfers := monitor.ListActiveTransfers()
				allCompleted := true

				for _, id := range transferIDs {
					fmt.Println("==== Transfer:", id, "====")
					metrics, exists := monitor.GetMetrics(id)
					if exists {
						reporter.ReportSummary(os.Stdout, metrics)
						fmt.Println()

						if metrics.Status == "ACTIVE" {
							allCompleted = false
						}
					}
				}

				// Check if all transfers are complete
				if allCompleted && completed < len(transferIDs) {
					completed = len(transferIDs)

					// Save final metrics if storage is enabled
					if storage != nil {
						fmt.Println("Saving final metrics to storage...")
						for _, id := range transferIDs {
							if err := monitor.SaveMetrics(storage, id); err != nil {
								fmt.Printf("Failed to save final metrics for %s: %v\n", id, err)
							}
						}
					}

					// Wait a moment to show final statistics before exiting
					time.AfterFunc(2*time.Second, func() {
						done <- true
					})
				}
			case <-done:
				return
			}
		}
	}()

	// Wait for all transfers to complete
	<-done
	fmt.Println("\nAll transfers completed!")

	if storage != nil {
		fmt.Println("\nRun with --history flag to view historical transfers")
	}
}

// simulateTransfer simulates a transfer with progress updates
func simulateTransfer(id string, totalBytes int64, monitor *metrics.DefaultPerformanceMonitor, progressBar *metrics.ProgressBar) {
	// Set up a random source with a seed based on the transfer ID
	// to get different behavior for each transfer
	source := rand.NewSource(int64(id[len(id)-1]))
	rnd := rand.New(source)

	// Simulate a transfer with variable speed
	var currentBytes int64
	var currentFiles int64

	// Record start time
	startTime := time.Now()

	// Simulate a delay before starting
	time.Sleep(time.Duration(rnd.Intn(2000)) * time.Millisecond)

	// Simulate the transfer
	for currentBytes < totalBytes {
		// Sleep to simulate processing time
		time.Sleep(time.Duration(25+rnd.Intn(75)) * time.Millisecond)

		// Calculate the next chunk size (variable rate)
		// Larger transfers go faster, with some randomness
		baseChunkSize := totalBytes / 100
		if baseChunkSize < 1024 {
			baseChunkSize = 1024
		}

		// Add some variability to the transfer rate
		variabilityFactor := 0.5 + rnd.Float64()
		chunkSize := int64(float64(baseChunkSize) * variabilityFactor)

		// Ensure we don't exceed the total
		if currentBytes+chunkSize > totalBytes {
			chunkSize = totalBytes - currentBytes
		}

		// Update the current bytes
		currentBytes += chunkSize

		// Occasionally update files transferred
		if rnd.Intn(10) == 0 && currentFiles < 10 {
			currentFiles++
		}

		// Occasionally simulate an error
		if rnd.Intn(50) == 0 {
			monitor.RecordError(id, fmt.Errorf("simulated temporary error"))
			monitor.RecordRetry(id)
			// Simulate a retry delay
			time.Sleep(time.Duration(500+rnd.Intn(500)) * time.Millisecond)
		}

		// Update the metrics
		monitor.UpdateMetrics(id, currentBytes, currentFiles)
		progressBar.Update(currentBytes)
	}

	// Ensure files count reaches 10 at the end
	monitor.UpdateMetrics(id, totalBytes, 10)
	progressBar.Update(totalBytes)

	// Mark as completed
	completionTime := time.Now()
	elapsedTime := completionTime.Sub(startTime)

	fmt.Printf("\n%s completed in %s\n", id, elapsedTime)

	// Complete the progress bar
	progressBar.Complete()

	// Set final status in the monitor
	monitor.SetStatus(id, "SUCCEEDED")
	monitor.StopMonitoring(id)
}
