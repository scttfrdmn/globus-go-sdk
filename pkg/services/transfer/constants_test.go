// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"testing"
)

// TestTransferConstants ensures that all required constants exist
// and have the correct values
func TestTransferConstants(t *testing.T) {
	// Test SyncLevel constants
	testCases := []struct {
		name     string
		constant int
		expected int
	}{
		{"SyncLevelExists", SyncLevelExists, 0},
		{"SyncLevelSize", SyncLevelSize, 1},
		{"SyncLevelModified", SyncLevelModified, 2},
		{"SyncLevelChecksum", SyncLevelChecksum, 3},
		// Backward compatibility alias
		{"SyncChecksum", SyncChecksum, SyncLevelChecksum},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Fatalf("Expected %s to be %d, got %d", tc.name, tc.expected, tc.constant)
			}
		})
	}
}

// TestMemoryOptimizedOptionsSyncLevel ensures that the DefaultMemoryOptimizedOptions
// function correctly initializes the SyncLevel field
func TestMemoryOptimizedOptionsSyncLevel(t *testing.T) {
	options := DefaultMemoryOptimizedOptions()

	// Verify SyncLevel is set to SyncLevelChecksum
	if options.SyncLevel != SyncLevelChecksum {
		t.Fatalf("Expected SyncLevel to be SyncLevelChecksum (%d), got %d",
			SyncLevelChecksum, options.SyncLevel)
	}

	// Verify backward compatibility: ensure SyncLevel equals SyncChecksum
	if options.SyncLevel != SyncChecksum {
		t.Fatalf("Expected SyncLevel (%d) to equal SyncChecksum (%d)",
			options.SyncLevel, SyncChecksum)
	}
}
