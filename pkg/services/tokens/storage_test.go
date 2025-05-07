// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestStorageInterface tests the Storage interface implementations.
func TestStorageInterface(t *testing.T) {
	// Create test entries
	entry1 := &Entry{
		Resource:     "test-resource-1",
		AccessToken:  "test-access-token-1",
		RefreshToken: "test-refresh-token-1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "test-scope-1",
		TokenSet: &TokenSet{
			AccessToken:  "test-access-token-1",
			RefreshToken: "test-refresh-token-1",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "test-scope-1",
			ResourceID:   "test-resource-1",
		},
	}

	entry2 := &Entry{
		Resource:     "test-resource-2",
		AccessToken:  "test-access-token-2",
		RefreshToken: "test-refresh-token-2",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "test-scope-2",
		TokenSet: &TokenSet{
			AccessToken:  "test-access-token-2",
			RefreshToken: "test-refresh-token-2",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "test-scope-2",
			ResourceID:   "test-resource-2",
		},
	}

	// Test implementations
	t.Run("MemoryStorage", func(t *testing.T) {
		testStorageImplementation(t, NewMemoryStorage(), entry1, entry2)
	})

	t.Run("FileStorage", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "tokens-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		if err != nil {
			t.Fatalf("NewFileStorage() error = %v", err)
		}

		testStorageImplementation(t, storage, entry1, entry2)
	})
}

// testStorageImplementation tests a specific Storage implementation.
func testStorageImplementation(t *testing.T, storage Storage, entry1, entry2 *Entry) {
	// Test Store and Lookup
	t.Run("Store and Lookup", func(t *testing.T) {
		// Store entry1
		err := storage.Store(entry1)
		if err != nil {
			t.Fatalf("Storage.Store() error = %v", err)
		}

		// Lookup entry1
		got, err := storage.Lookup(entry1.Resource)
		if err != nil {
			t.Fatalf("Storage.Lookup() error = %v", err)
		}
		if got == nil {
			t.Fatalf("Storage.Lookup() = nil, want entry")
		}
		if got.AccessToken != entry1.AccessToken {
			t.Errorf("Storage.Lookup().AccessToken = %v, want %v", got.AccessToken, entry1.AccessToken)
		}
		if got.RefreshToken != entry1.RefreshToken {
			t.Errorf("Storage.Lookup().RefreshToken = %v, want %v", got.RefreshToken, entry1.RefreshToken)
		}
		if got.Scope != entry1.Scope {
			t.Errorf("Storage.Lookup().Scope = %v, want %v", got.Scope, entry1.Scope)
		}
		if got.TokenSet == nil {
			t.Errorf("Storage.Lookup().TokenSet = nil, want non-nil")
		} else {
			if got.TokenSet.AccessToken != entry1.TokenSet.AccessToken {
				t.Errorf("Storage.Lookup().TokenSet.AccessToken = %v, want %v", got.TokenSet.AccessToken, entry1.TokenSet.AccessToken)
			}
		}

		// Lookup non-existent entry
		got, err = storage.Lookup("non-existent")
		if err != nil {
			t.Fatalf("Storage.Lookup() error = %v", err)
		}
		if got != nil {
			t.Errorf("Storage.Lookup() = %v, want nil", got)
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		// Store entry2
		err := storage.Store(entry2)
		if err != nil {
			t.Fatalf("Storage.Store() error = %v", err)
		}

		// List entries
		resources, err := storage.List()
		if err != nil {
			t.Fatalf("Storage.List() error = %v", err)
		}
		if len(resources) != 2 {
			t.Errorf("Storage.List() = %v, want 2 items", len(resources))
		}
		foundResource1 := false
		foundResource2 := false
		for _, resource := range resources {
			if resource == entry1.Resource {
				foundResource1 = true
			}
			if resource == entry2.Resource {
				foundResource2 = true
			}
		}
		if !foundResource1 {
			t.Errorf("Storage.List() does not contain %v", entry1.Resource)
		}
		if !foundResource2 {
			t.Errorf("Storage.List() does not contain %v", entry2.Resource)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		// Delete entry1
		err := storage.Delete(entry1.Resource)
		if err != nil {
			t.Fatalf("Storage.Delete() error = %v", err)
		}

		// Verify entry1 was deleted
		got, err := storage.Lookup(entry1.Resource)
		if err != nil {
			t.Fatalf("Storage.Lookup() error = %v", err)
		}
		if got != nil {
			t.Errorf("Storage.Lookup() = %v, want nil", got)
		}

		// Verify entry2 still exists
		got, err = storage.Lookup(entry2.Resource)
		if err != nil {
			t.Fatalf("Storage.Lookup() error = %v", err)
		}
		if got == nil {
			t.Fatalf("Storage.Lookup() = nil, want entry")
		}
		if got.AccessToken != entry2.AccessToken {
			t.Errorf("Storage.Lookup().AccessToken = %v, want %v", got.AccessToken, entry2.AccessToken)
		}

		// Delete entry2
		err = storage.Delete(entry2.Resource)
		if err != nil {
			t.Fatalf("Storage.Delete() error = %v", err)
		}

		// Verify entry2 was deleted
		got, err = storage.Lookup(entry2.Resource)
		if err != nil {
			t.Fatalf("Storage.Lookup() error = %v", err)
		}
		if got != nil {
			t.Errorf("Storage.Lookup() = %v, want nil", got)
		}

		// List should be empty
		resources, err := storage.List()
		if err != nil {
			t.Fatalf("Storage.List() error = %v", err)
		}
		if len(resources) != 0 {
			t.Errorf("Storage.List() = %v, want 0 items", len(resources))
		}
	})
}

// TestStorageConcurrency tests the Storage implementations for concurrent access.
func TestStorageConcurrency(t *testing.T) {
	t.Run("MemoryStorage", func(t *testing.T) {
		testStorageConcurrency(t, NewMemoryStorage())
	})

	t.Run("FileStorage", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "tokens-test-concurrency")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		if err != nil {
			t.Fatalf("NewFileStorage() error = %v", err)
		}

		testStorageConcurrency(t, storage)
	})
}

// testStorageConcurrency tests a specific Storage implementation for concurrent access.
func testStorageConcurrency(t *testing.T, storage Storage) {
	// Number of concurrent operations
	const numConcurrent = 10
	const numOperations = 10

	// Wait group to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(numConcurrent)

	// Start goroutines
	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			defer wg.Done()

			// Perform multiple operations
			for j := 0; j < numOperations; j++ {
				// Create a unique resource ID for this goroutine and operation
				resourceID := fmt.Sprintf("resource-%d-%d", id, j)

				// Create an entry
				entry := &Entry{
					Resource:     resourceID,
					AccessToken:  fmt.Sprintf("access-token-%d-%d", id, j),
					RefreshToken: fmt.Sprintf("refresh-token-%d-%d", id, j),
					ExpiresAt:    time.Now().Add(1 * time.Hour),
					Scope:        fmt.Sprintf("scope-%d-%d", id, j),
					TokenSet: &TokenSet{
						AccessToken:  fmt.Sprintf("access-token-%d-%d", id, j),
						RefreshToken: fmt.Sprintf("refresh-token-%d-%d", id, j),
						ExpiresAt:    time.Now().Add(1 * time.Hour),
						Scope:        fmt.Sprintf("scope-%d-%d", id, j),
						ResourceID:   resourceID,
					},
				}

				// Store the entry
				err := storage.Store(entry)
				if err != nil {
					t.Errorf("Storage.Store() error = %v", err)
					continue
				}

				// Lookup the entry
				got, err := storage.Lookup(resourceID)
				if err != nil {
					t.Errorf("Storage.Lookup() error = %v", err)
					continue
				}
				if got == nil {
					t.Errorf("Storage.Lookup() = nil, want entry")
					continue
				}
				if got.AccessToken != entry.AccessToken {
					t.Errorf("Storage.Lookup().AccessToken = %v, want %v", got.AccessToken, entry.AccessToken)
				}

				// Delete the entry
				err = storage.Delete(resourceID)
				if err != nil {
					t.Errorf("Storage.Delete() error = %v", err)
					continue
				}

				// Verify the entry was deleted
				got, err = storage.Lookup(resourceID)
				if err != nil {
					t.Errorf("Storage.Lookup() error = %v", err)
					continue
				}
				if got != nil {
					t.Errorf("Storage.Lookup() = %v, want nil", got)
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify storage is empty
	resources, err := storage.List()
	if err != nil {
		t.Fatalf("Storage.List() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Storage.List() = %v, want 0 items", len(resources))
	}
}

// TestFileStorageErrors tests the FileStorage error handling.
func TestFileStorageErrors(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "tokens-test-errors")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create FileStorage
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}

	// Test invalid entry
	err = storage.Store(nil)
	if err == nil {
		t.Error("Storage.Store(nil) error = nil, want error")
	}

	err = storage.Store(&Entry{})
	if err == nil {
		t.Error("Storage.Store(empty) error = nil, want error")
	}

	// Test non-existent directory
	_, err = NewFileStorage("/non-existent-directory-that-should-not-exist")
	if err == nil {
		t.Error("NewFileStorage() error = nil, want error")
	}
}

// Helper function to check if two time values are close enough.
func isTimeClose(t1, t2 time.Time, threshold time.Duration) bool {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return diff < threshold
}
