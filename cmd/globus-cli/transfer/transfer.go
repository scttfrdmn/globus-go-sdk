// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/cmd/globus-cli/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg"
)

// ListCommand handles the list command
func ListCommand(args []string) error {
	// Check the arguments
	if len(args) < 2 {
		return fmt.Errorf("usage: globus-cli ls <endpoint-id> <path>")
	}

	// Get the arguments
	endpointID := args[0]
	path := args[1]

	// Load the configuration
	config, err := auth.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Load the token
	token, err := auth.LoadToken(config, auth.DefaultTokenFile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if the token is valid
	if !auth.IsTokenValid(token) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Create a new SDK configuration
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	// Create a new transfer client
	transferClient := sdkConfig.NewTransferClient(token.AccessToken)

	// Set up options for listing files
	options := &pkg.ListFileOptions{
		ShowHidden: true,
	}

	// List the files
	fileList, err := transferClient.ListFiles(context.Background(), endpointID, path, options)
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	// Print the files
	fmt.Printf("Contents of %s:%s

", endpointID, fileList.Path)
	fmt.Printf("%-12s %-12s %-12s %-20s %s
", "Type", "Permissions", "Size", "Last Modified", "Name")
	fmt.Println(strings.Repeat("-", 80))

	for _, file := range fileList.Data {
		// Format size
		sizeStr := "-"
		if file.Type == "file" {
			sizeStr = formatSize(file.Size)
		}

		// Format the last modified time
		lastModified := file.LastModified
		if lastModified == "" {
			lastModified = "-"
		} else {
			// Try to parse and format the time
			if t, err := time.Parse("2006-01-02 15:04:05", lastModified); err == nil {
				lastModified = t.Format("Jan 02, 2006 15:04:05")
			}
		}

		// Format the file type
		typeStr := file.Type
		if file.Type == "dir" {
			typeStr = "directory"
		}

		// Print the file information
		fmt.Printf("%-12s %-12s %-12s %-20s %s
",
			typeStr,
			file.Permissions,
			sizeStr,
			lastModified,
			file.Name,
		)
	}

	return nil
}

// TransferCommand handles the transfer command
func TransferCommand(args []string) error {
	// Check the arguments
	if len(args) < 4 {
		return fmt.Errorf("usage: globus-cli transfer <source-endpoint-id> <source-path> <dest-endpoint-id> <dest-path> [--recursive] [--sync] [--label <label>]")
	}

	// Get the required arguments
	sourceEndpointID := args[0]
	sourcePath := args[1]
	destEndpointID := args[2]
	destPath := args[3]

	// Parse optional arguments
	var recursive bool
	var sync bool
	var label string

	for i := 4; i < len(args); i++ {
		switch args[i] {
		case "--recursive", "-r":
			recursive = true
		case "--sync", "-s":
			sync = true
		case "--label", "-l":
			if i+1 < len(args) {
				label = args[i+1]
				i++
			} else {
				return fmt.Errorf("--label requires a value")
			}
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
	}

	// Default label
	if label == "" {
		label = fmt.Sprintf("CLI Transfer %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// Load the configuration
	config, err := auth.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Load the token
	token, err := auth.LoadToken(config, auth.DefaultTokenFile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if the token is valid
	if !auth.IsTokenValid(token) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Create a new SDK configuration
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	// Create a new transfer client
	transferClient := sdkConfig.NewTransferClient(token.AccessToken)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var result *pkg.TransferResponse

	// Submit the transfer
	if recursive {
		// Set up options for recursive transfer
		options := pkg.DefaultRecursiveTransferOptions()
		options.Label = label
		options.Sync = sync
		options.ProgressCallback = func(current, total int64, message string) {
			if total > 0 {
				fmt.Printf("Progress: %d/%d files (%s)
", current, total, message)
			} else {
				fmt.Printf("Progress: %s
", message)
			}
		}

		// Submit a recursive transfer
		recursiveResult, err := transferClient.SubmitRecursiveTransfer(
			ctx,
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			options,
		)
		if err != nil {
			return fmt.Errorf("error submitting recursive transfer: %w", err)
		}

		fmt.Printf("Recursive transfer submitted with task ID: %s
", recursiveResult.TaskID)
		fmt.Printf("Found %d files (%s) in %d directories
",
			recursiveResult.TotalFiles,
			formatSize(recursiveResult.TotalSize),
			recursiveResult.Directories+recursiveResult.Subdirectories,
		)

		// Create a simplified transfer response for status tracking
		result = &pkg.TransferResponse{
			TaskID: recursiveResult.TaskID,
		}
	} else {
		// Set up options for regular transfer
		options := map[string]interface{}{
			"label":            label,
			"sync_level":       getSyncLevel(sync),
			"verify_checksum":  true,
			"preserve_mtime":   true,
			"encrypt_data":     true,
		}

		// Submit a regular transfer
		result, err = transferClient.SubmitTransfer(
			ctx,
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			options,
		)
		if err != nil {
			return fmt.Errorf("error submitting transfer: %w", err)
		}

		fmt.Printf("Transfer submitted with task ID: %s
", result.TaskID)
	}

	// Save the task ID to a file for later status checks
	homeDir, err := os.UserHomeDir()
	if err == nil {
		taskFile := filepath.Join(homeDir, ".globus-cli", "last-task-id")
		os.WriteFile(taskFile, []byte(result.TaskID), 0600)
	}

	return nil
}

// StatusCommand handles the status command
func StatusCommand(args []string) error {
	var taskID string

	// Check if a task ID was provided
	if len(args) > 0 {
		taskID = args[0]
	} else {
		// Try to load the last task ID from file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error determining home directory: %w", err)
		}

		taskFile := filepath.Join(homeDir, ".globus-cli", "last-task-id")
		data, err := os.ReadFile(taskFile)
		if err != nil {
			return fmt.Errorf("no task ID provided and couldn't load last task ID: %w", err)
		}

		taskID = string(data)
	}

	// Load the configuration
	config, err := auth.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Load the token
	token, err := auth.LoadToken(config, auth.DefaultTokenFile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if the token is valid
	if !auth.IsTokenValid(token) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Create a new SDK configuration
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	// Create a new transfer client
	transferClient := sdkConfig.NewTransferClient(token.AccessToken)

	// Get the task
	task, err := transferClient.GetTask(context.Background(), taskID)
	if err != nil {
		return fmt.Errorf("error getting task status: %w", err)
	}

	// Print the task information
	fmt.Println("Task Information:")
	fmt.Printf("  Task ID:      %s
", task.TaskID)
	fmt.Printf("  Status:       %s
", task.Status)
	fmt.Printf("  Type:         %s
", task.Type)
	fmt.Printf("  Label:        %s
", task.Label)
	fmt.Printf("  Owner:        %s
", task.Owner)
	
	// Print transfer specific information
	if task.Type == "TRANSFER" {
		fmt.Printf("  Source:       %s
", task.SourceEndpoint)
		fmt.Printf("  Destination:  %s
", task.DestEndpoint)
		fmt.Printf("  Files:        %d transferred, %d skipped, %d failed
",
			task.FilesTransferred, task.FilesSkipped, task.FilesSkippedFail)
		fmt.Printf("  Directories:  %d created
", task.DirectoriesCreated)
		fmt.Printf("  Size:         %s transferred
", formatSize(task.BytesTransferred))
		
		// Print completion information if available
		if task.Status == "SUCCEEDED" || task.Status == "FAILED" {
			fmt.Printf("  Completed:    %s
", formatDuration(task.CompletionTime, task.RequestTime))
		}
	}

	// Print additional information based on status
	switch task.Status {
	case "ACTIVE":
		fmt.Printf("  Progress:     %s of %s (%.1f%%)
",
			formatSize(task.BytesTransferred),
			formatSize(task.BytesExpected),
			percentComplete(task.BytesTransferred, task.BytesExpected))
	case "SUCCEEDED":
		fmt.Println("  Result:       Transfer completed successfully")
	case "FAILED":
		fmt.Printf("  Result:       Transfer failed: %s
", task.NiceStatusShortDescription)
	}

	return nil
}

// Helper functions

// formatSize formats a size in bytes to a human-readable string
func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
}

// getSyncLevel returns the sync level value based on options
func getSyncLevel(sync bool) string {
	if !sync {
		return "0"
	}
	return "3" // Size and timestamp with checksum
}

// percentComplete calculates the percentage completion
func percentComplete(current, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}

// formatDuration formats a duration between two timestamps
func formatDuration(endTime, startTime string) string {
	// Parse the timestamps
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return "unknown"
	}
	
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return "unknown"
	}
	
	// Calculate the duration
	duration := end.Sub(start)
	
	// Format the duration
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	
	return fmt.Sprintf("%ds", seconds)
}
