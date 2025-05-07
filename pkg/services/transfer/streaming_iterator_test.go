// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"fmt"
	"testing"
)

type mockIterator struct {
	files     []FileListItem
	index     int
	hasError  bool
	mockError error
}

func (m *mockIterator) Next() (FileListItem, bool) {
	// If we're past the end or have an error, return false
	if m.hasError || m.index >= len(m.files) {
		return FileListItem{}, false
	}

	// Get the current file and advance the index
	file := m.files[m.index]
	m.index++
	return file, true
}

func (m *mockIterator) Error() error {
	if m.hasError {
		return m.mockError
	}
	return nil
}

func (m *mockIterator) Reset() error {
	m.index = 0
	return nil
}

func (m *mockIterator) Close() error {
	return nil
}

// TestFileIteratorInterface tests the file iteration functionality using a mock iterator
func TestFileIteratorInterface(t *testing.T) {
	// Create test file list
	files := []FileListItem{
		{DataType: "file", Name: "file1.txt", Type: "file", Size: 1024, LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "dir", Name: "dir1", Type: "dir", LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "dir", Name: "dir2", Type: "dir", LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "file", Name: "file2.txt", Type: "file", Size: 2048, LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "file", Name: "file3.txt", Type: "file", Size: 3072, LastModified: "2021-01-01T00:00:00Z"},
	}

	t.Run("Basic iteration", func(t *testing.T) {
		// Create mock iterator
		iterator := &mockIterator{
			files: files,
			index: 0,
		}

		// Collect files from the iterator
		var collected []FileListItem
		for {
			file, ok := iterator.Next()
			if !ok {
				if err := iterator.Error(); err != nil {
					t.Fatalf("Iterator error: %v", err)
				}
				break
			}
			collected = append(collected, file)
		}

		// Verify count
		if len(collected) != len(files) {
			t.Errorf("Expected %d files, got %d", len(files), len(collected))
		}

		// Verify content
		for i, file := range collected {
			if file.Name != files[i].Name {
				t.Errorf("Expected file %s at index %d, got %s", files[i].Name, i, file.Name)
			}
		}
	})

	t.Run("Reset functionality", func(t *testing.T) {
		// Create mock iterator
		iterator := &mockIterator{
			files: files,
			index: 0,
		}

		// Read half the files
		var firstHalf []FileListItem
		halfCount := len(files) / 2
		for i := 0; i < halfCount; i++ {
			file, ok := iterator.Next()
			if !ok {
				t.Fatalf("Unexpected end of iteration")
			}
			firstHalf = append(firstHalf, file)
		}

		// Reset the iterator
		err := iterator.Reset()
		if err != nil {
			t.Fatalf("Failed to reset iterator: %v", err)
		}

		// Read all files
		var allFiles []FileListItem
		for {
			file, ok := iterator.Next()
			if !ok {
				break
			}
			allFiles = append(allFiles, file)
		}

		// Verify we got all files after reset
		if len(allFiles) != len(files) {
			t.Errorf("Expected %d files after reset, got %d", len(files), len(allFiles))
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		// Create mock iterator with error
		mockError := fmt.Errorf("test error")
		iterator := &mockIterator{
			files:     files,
			index:     0,
			hasError:  true,
			mockError: mockError,
		}

		// Try to get a file
		_, ok := iterator.Next()
		if ok {
			t.Errorf("Expected Next() to return false when iterator has error")
		}

		// Check error
		err := iterator.Error()
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if err.Error() != mockError.Error() {
			t.Errorf("Expected error %q, got %q", mockError.Error(), err.Error())
		}
	})
}

// TestCollectFiles tests the CollectFiles utility function
func TestCollectFiles(t *testing.T) {
	// Create test file list
	files := []FileListItem{
		{DataType: "file", Name: "file1.txt", Type: "file", Size: 1024, LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "dir", Name: "dir1", Type: "dir", LastModified: "2021-01-01T00:00:00Z"},
		{DataType: "file", Name: "file2.txt", Type: "file", Size: 2048, LastModified: "2021-01-01T00:00:00Z"},
	}

	// Create mock iterator
	iterator := &mockIterator{
		files: files,
		index: 0,
	}

	// Use CollectFiles to collect the files
	collected, err := CollectFiles(iterator)
	if err != nil {
		t.Fatalf("CollectFiles failed: %v", err)
	}

	// Verify count
	if len(collected) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(collected))
	}

	// Verify content
	for i, file := range collected {
		if file.Name != files[i].Name {
			t.Errorf("Expected file %s at index %d, got %s", files[i].Name, i, file.Name)
		}
	}
}
