// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDefaultRecursiveTransferOptions(t *testing.T) {
	options := DefaultRecursiveTransferOptions()

	if !options.Recursive {
		t.Error("Default options should have Recursive=true")
	}
	if !options.PreserveTimestamp {
		t.Error("Default options should have PreserveTimestamp=true")
	}
	if !options.VerifyChecksum {
		t.Error("Default options should have VerifyChecksum=true")
	}
	if !options.EncryptData {
		t.Error("Default options should have EncryptData=true")
	}
	if options.DeleteDestinationExtra {
		t.Error("Default options should have DeleteDestinationExtra=false")
	}
	if options.MaxConcurrentListings != 4 {
		t.Errorf("Default MaxConcurrentListings = %d, want 4", options.MaxConcurrentListings)
	}
	if options.MaxConcurrentTransfers != 1 {
		t.Errorf("Default MaxConcurrentTransfers = %d, want 1", options.MaxConcurrentTransfers)
	}
}

func TestCountDirectories(t *testing.T) {
	files := []FileListItem{
		{Type: "dir", Name: "dir1"},
		{Type: "file", Name: "file1.txt"},
		{Type: "dir", Name: "dir2"},
		{Type: "file", Name: "file2.txt"},
		{Type: "dir", Name: "dir3"},
	}

	count := countDirectories(files)
	if count != 3 {
		t.Errorf("countDirectories() = %d, want 3", count)
	}
}

func TestCalculateTotals(t *testing.T) {
	files := []FileListItem{
		{Type: "dir", Name: "dir1", Size: 0},
		{Type: "file", Name: "file1.txt", Size: 100},
		{Type: "dir", Name: "dir2", Size: 0},
		{Type: "file", Name: "file2.txt", Size: 200},
		{Type: "file", Name: "file3.txt", Size: 300},
	}

	totalSize, totalFiles := calculateTotals(files)
	if totalSize != 600 {
		t.Errorf("calculateTotals() total size = %d, want 600", totalSize)
	}
	if totalFiles != 3 {
		t.Errorf("calculateTotals() total files = %d, want 3", totalFiles)
	}
}

func TestPrepareTransferItems(t *testing.T) {
	files := []FileListItem{
		{Type: "dir", Name: "dir1"},
		{Type: "file", Name: "file1.txt"},
		{Type: "file", Name: "file2.txt"},
	}

	sourcePath := "/source"
	destPath := "/destination"

	items := prepareTransferItems(files, sourcePath, destPath)

	if len(items) != 2 {
		t.Fatalf("prepareTransferItems() returned %d items, want 2", len(items))
	}

	// Check the first item
	if items[0].SourcePath != "/source/file1.txt" {
		t.Errorf("First item source path = %s, want /source/file1.txt", items[0].SourcePath)
	}
	if items[0].DestinationPath != "/destination/file1.txt" {
		t.Errorf("First item destination path = %s, want /destination/file1.txt", items[0].DestinationPath)
	}
	if items[0].Recursive {
		t.Error("File items should not be recursive")
	}

	// Check the second item
	if items[1].SourcePath != "/source/file2.txt" {
		t.Errorf("Second item source path = %s, want /source/file2.txt", items[1].SourcePath)
	}
	if items[1].DestinationPath != "/destination/file2.txt" {
		t.Errorf("Second item destination path = %s, want /destination/file2.txt", items[1].DestinationPath)
	}
}

func TestGetSyncLevel(t *testing.T) {
	// Test with sync off
	options := &RecursiveTransferOptions{
		Sync:           false,
		VerifyChecksum: true,
	}
	level := getSyncLevel(options)
	if level != "0" {
		t.Errorf("getSyncLevel() with sync off = %s, want 0", level)
	}

	// Test with sync on, checksum off
	options = &RecursiveTransferOptions{
		Sync:           true,
		VerifyChecksum: false,
	}
	level = getSyncLevel(options)
	if level != "1" {
		t.Errorf("getSyncLevel() with sync on, checksum off = %s, want 1", level)
	}

	// Test with sync on, checksum on
	options = &RecursiveTransferOptions{
		Sync:           true,
		VerifyChecksum: true,
	}
	level = getSyncLevel(options)
	if level != "3" {
		t.Errorf("getSyncLevel() with sync on, checksum on = %s, want 3", level)
	}
}

func TestSubmitRecursiveTransfer(t *testing.T) {
	// Setup test server to handle recursive directory listing
	dirListingHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if this is the directory listing request
		if r.URL.Path == "/operation/endpoint/source-endpoint/ls" {
			// Return a mock directory listing
			fileList := FileList{
				Data: []FileListItem{
					{Type: "dir", Name: "subdir1"},
					{Type: "file", Name: "file1.txt", Size: 100},
					{Type: "file", Name: "file2.txt", Size: 200},
				},
				Path: "/source",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(fileList)
			return
		}

		// Check if this is the subdirectory listing request
		if r.URL.Path == "/operation/endpoint/source-endpoint/ls" && r.URL.Query().Get("path") == "/source/subdir1" {
			// Return a mock subdirectory listing
			fileList := FileList{
				Data: []FileListItem{
					{Type: "file", Name: "file3.txt", Size: 300},
					{Type: "file", Name: "file4.txt", Size: 400},
				},
				Path: "/source/subdir1",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(fileList)
			return
		}

		// Check if this is the transfer submission request
		if r.URL.Path == "/transfer" {
			// Return a successful task submission response
			response := TaskResponse{
				TaskID:  "task-12345",
				Code:    "Accepted",
				Message: "Transfer task submitted successfully",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return 404 for any other requests
		w.WriteHeader(http.StatusNotFound)
	}

	server, client := setupMockServer(dirListingHandler)
	defer server.Close()

	// Create a callback to track progress
	progressUpdates := 0
	progressCallback := func(current, total int64, message string) {
		progressUpdates++
	}

	// Create transfer options
	options := DefaultRecursiveTransferOptions()
	options.ProgressCallback = progressCallback

	// Submit the recursive transfer
	result, err := client.SubmitRecursiveTransfer(
		context.Background(),
		"source-endpoint", "/source",
		"destination-endpoint", "/destination",
		options,
	)

	if err != nil {
		t.Fatalf("SubmitRecursiveTransfer() error = %v", err)
	}

	// Check the result
	if result.TaskID != "task-12345" {
		t.Errorf("SubmitRecursiveTransfer() TaskID = %s, want task-12345", result.TaskID)
	}

	if result.TotalFiles != 4 {
		t.Errorf("SubmitRecursiveTransfer() TotalFiles = %d, want 4", result.TotalFiles)
	}

	if result.TotalSize != 1000 {
		t.Errorf("SubmitRecursiveTransfer() TotalSize = %d, want 1000", result.TotalSize)
	}

	if result.Directories != 1 {
		t.Errorf("SubmitRecursiveTransfer() Directories = %d, want 1", result.Directories)
	}

	if result.Subdirectories != 1 {
		t.Errorf("SubmitRecursiveTransfer() Subdirectories = %d, want 1", result.Subdirectories)
	}

	if progressUpdates == 0 {
		t.Error("ProgressCallback was not called")
	}
}
