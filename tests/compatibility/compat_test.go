// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package compatibility

import (
	"context"
	"testing"
	"time"
)

func TestCompatibility(t *testing.T) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run all registered compatibility tests
	GlobalRegistry.RunAllTests(ctx, t)
}
