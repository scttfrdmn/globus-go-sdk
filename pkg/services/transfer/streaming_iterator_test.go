// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

func TestStreamingFileIterator(t *testing.T) {
	// Create a mock server to simulate Transfer API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Handle different paths
		switch r.URL.Path {
		case "/endpoint/mock-endpoint-id/ls":
			// Parse path parameter
			path := r.URL.Query().Get("path")
			switch path {
			case "/root":
				// Return root directory listing
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file1.txt", "type": "file", "size": 1024, "last_modified": "2021-01-01T00:00:00Z"},
						{"name": "dir1", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z"},
						{"name": "dir2", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z"}
					],
					"endpoint": "mock-endpoint-id",
					"path": "/root",
					"DATA_TYPE": "file_list"
				}`)
			case "/root/dir1":
				// Return dir1 listing
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file2.txt", "type": "file", "size": 2048, "last_modified": "2021-01-01T00:00:00Z"},
						{"name": "file3.txt", "type": "file", "size": 3072, "last_modified": "2021-01-01T00:00:00Z"}
					],
					"endpoint": "mock-endpoint-id",
					"path": "/root/dir1",
					"DATA_TYPE": "file_list"
				}`)
			case "/root/dir2":
				// Return dir2 listing
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file4.txt", "type": "file", "size": 4096, "last_modified": "2021-01-01T00:00:00Z"},
						{"name": "subdir", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z"}
					],
					"endpoint": "mock-endpoint-id",
					"path": "/root/dir2",
					"DATA_TYPE": "file_list"
				}`)
			case "/root/dir2/subdir":
				// Return subdir listing
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file5.txt", "type": "file", "size": 5120, "last_modified": "2021-01-01T00:00:00Z"}
					],
					"endpoint": "mock-endpoint-id",
					"path": "/root/dir2/subdir",
					"DATA_TYPE": "file_list"
				}`)
			default:
				// Unknown path
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `{"code": "NotFound", "message": "Path %s not found"}`, path)
			}
		default:
			// Unknown endpoint
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code": "NotFound", "message": "Unknown endpoint: %s"}`, r.URL.Path)
		}
	}))
	defer server.Close()
	
	// Create a client that uses the test server
	httpClient := &http.Client{}
	transport := core.NewHTTPTransport(server.URL, httpClient)
	client := &Client{
		Transport: transport,
	}
	
	t.Run("Iterate through all files", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create iterator
		iterator, err := NewStreamingFileIterator(ctx, client, "mock-endpoint-id", "/root", nil)
		if err != nil {
			t.Fatalf("Failed to create iterator: %v", err)
		}
		defer iterator.Close()
		
		// Collect all files
		var files []FileListItem
		for {
			file, ok := iterator.Next()
			if !ok {
				if err := iterator.Error(); err != nil {
					t.Fatalf("Iterator error: %v", err)
				}
				break
			}
			files = append(files, file)
		}
		
		// Verify we got all 7 expected items (3 in root, 2 in dir1, 2 in dir2, including subdirectory)
		expectedCount := 7
		if len(files) != expectedCount {
			t.Errorf("Expected %d files, got %d", expectedCount, len(files))
		}
		
		// Count files vs directories
		fileCount := 0
		dirCount := 0
		for _, file := range files {
			if file.Type == "file" {
				fileCount++
			} else if file.Type == "dir" {
				dirCount++
			}
		}
		
		// Verify counts
		if fileCount != 5 {
			t.Errorf("Expected 5 files, got %d", fileCount)
		}
		if dirCount != 2 {
			t.Errorf("Expected 2 directories, got %d", dirCount)
		}
	})
	
	t.Run("Limited depth", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create iterator with depth limit of 1 (should skip subdir)
		iterator, err := NewStreamingFileIterator(ctx, client, "mock-endpoint-id", "/root", &StreamingIteratorOptions{
			Recursive:   true,
			ShowHidden:  true,
			MaxDepth:    1,
			Concurrency: 2,
		})
		if err != nil {
			t.Fatalf("Failed to create iterator: %v", err)
		}
		defer iterator.Close()
		
		// Collect all files
		var files []FileListItem
		for {
			file, ok := iterator.Next()
			if !ok {
				if err := iterator.Error(); err != nil {
					t.Fatalf("Iterator error: %v", err)
				}
				break
			}
			files = append(files, file)
		}
		
		// Should see everything except the file in the subdir (6 items)
		expectedCount := 6
		if len(files) != expectedCount {
			t.Errorf("Expected %d files, got %d", expectedCount, len(files))
		}
		
		// Verify we didn't get file5.txt in the subdir
		for _, file := range files {
			if file.Name == "file5.txt" {
				t.Errorf("Found file5.txt which should have been skipped due to depth limit")
			}
		}
	})
	
	t.Run("Collect helper function", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create iterator
		iterator, err := NewStreamingFileIterator(ctx, client, "mock-endpoint-id", "/root", nil)
		if err != nil {
			t.Fatalf("Failed to create iterator: %v", err)
		}
		defer iterator.Close()
		
		// Use helper function to collect files
		files, err := CollectFiles(iterator)
		if err != nil {
			t.Fatalf("CollectFiles failed: %v", err)
		}
		
		// Verify count
		expectedCount := 7
		if len(files) != expectedCount {
			t.Errorf("Expected %d files, got %d", expectedCount, len(files))
		}
	})
	
	t.Run("Reset iterator", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create iterator
		iterator, err := NewStreamingFileIterator(ctx, client, "mock-endpoint-id", "/root", nil)
		if err != nil {
			t.Fatalf("Failed to create iterator: %v", err)
		}
		defer iterator.Close()
		
		// Collect files once
		files1, err := CollectFiles(iterator)
		if err != nil {
			t.Fatalf("First collection failed: %v", err)
		}
		
		// Reset and collect again
		if err := iterator.Reset(); err != nil {
			t.Fatalf("Failed to reset iterator: %v", err)
		}
		
		files2, err := CollectFiles(iterator)
		if err != nil {
			t.Fatalf("Second collection failed: %v", err)
		}
		
		// Both collections should have same count
		if len(files1) != len(files2) {
			t.Errorf("Expected same file count after reset, got %d and %d", len(files1), len(files2))
		}
	})
	
	t.Run("Non-recursive", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Create non-recursive iterator
		iterator, err := NewStreamingFileIterator(ctx, client, "mock-endpoint-id", "/root", &StreamingIteratorOptions{
			Recursive:   false,
			ShowHidden:  true,
			Concurrency: 2,
		})
		if err != nil {
			t.Fatalf("Failed to create iterator: %v", err)
		}
		defer iterator.Close()
		
		// Collect files
		files, err := CollectFiles(iterator)
		if err != nil {
			t.Fatalf("Collection failed: %v", err)
		}
		
		// Should only see the 3 items in root
		expectedCount := 3
		if len(files) != expectedCount {
			t.Errorf("Expected %d files for non-recursive, got %d", expectedCount, len(files))
		}
	})
}