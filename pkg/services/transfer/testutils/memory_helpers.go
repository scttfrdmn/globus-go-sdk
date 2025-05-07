// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package testutils

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

// MemoryTracker tracks memory usage during tests
type MemoryTracker struct {
	MaxUsage uint64
	mutex    sync.Mutex
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		MaxUsage: 0,
	}
}

// TrackMemoryUsage monitors memory usage during a test
func (m *MemoryTracker) TrackMemoryUsage(ctx context.Context, t *testing.T, interval int) {
	t.Helper()

	done := make(chan struct{})
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				memStats := &runtime.MemStats{}
				runtime.ReadMemStats(memStats)

				m.mutex.Lock()
				if memStats.Alloc > m.MaxUsage {
					m.MaxUsage = memStats.Alloc
				}
				m.mutex.Unlock()

			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	<-done
}
