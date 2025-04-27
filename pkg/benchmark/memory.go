// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package benchmark

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryStats contains memory usage statistics
type MemoryStats struct {
	Alloc         uint64    `json:"alloc"`           // bytes allocated and not yet freed
	TotalAlloc    uint64    `json:"total_alloc"`     // bytes allocated (even if freed)
	Sys           uint64    `json:"sys"`             // bytes obtained from system
	NumGC         uint32    `json:"num_gc"`          // number of completed GC cycles
	GCCPUFraction float64   `json:"gc_cpu_fraction"` // fraction of CPU time used by GC
	HeapAlloc     uint64    `json:"heap_alloc"`      // bytes allocated and not yet freed (same as Alloc)
	HeapSys       uint64    `json:"heap_sys"`        // bytes obtained from system
	Time          time.Time `json:"time"`            // time when stats were collected
}

// MemorySampler continuously samples memory usage
type MemorySampler struct {
	interval time.Duration
	samples  []MemoryStats
	done     chan struct{}
	maxAlloc uint64
	maxSys   uint64
}

// NewMemorySampler creates a new memory sampler with the specified sampling interval
func NewMemorySampler(interval time.Duration) *MemorySampler {
	if interval < time.Millisecond {
		interval = time.Millisecond
	}

	return &MemorySampler{
		interval: interval,
		samples:  make([]MemoryStats, 0),
		done:     make(chan struct{}),
	}
}

// Start begins sampling memory usage
func (s *MemorySampler) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := s.collectStats()
				s.samples = append(s.samples, stats)

				// Update max values
				if stats.Alloc > s.maxAlloc {
					s.maxAlloc = stats.Alloc
				}
				if stats.Sys > s.maxSys {
					s.maxSys = stats.Sys
				}

			case <-s.done:
				return
			}
		}
	}()
}

// Stop ends sampling memory usage
func (s *MemorySampler) Stop() {
	close(s.done)
}

// GetSamples returns all collected memory samples
func (s *MemorySampler) GetSamples() []MemoryStats {
	return s.samples
}

// GetPeakMemory returns the peak memory usage in megabytes
func (s *MemorySampler) GetPeakMemory() float64 {
	return float64(s.maxAlloc) / (1024 * 1024)
}

// GetSummary returns a summary of memory usage statistics
func (s *MemorySampler) GetSummary() map[string]interface{} {
	if len(s.samples) == 0 {
		return map[string]interface{}{
			"error": "no samples collected",
		}
	}

	first := s.samples[0]
	last := s.samples[len(s.samples)-1]

	return map[string]interface{}{
		"samples_count":     len(s.samples),
		"sampling_duration": last.Time.Sub(first.Time).String(),
		"sampling_interval": s.interval.String(),
		"peak_alloc_mb":     float64(s.maxAlloc) / (1024 * 1024),
		"peak_sys_mb":       float64(s.maxSys) / (1024 * 1024),
		"final_alloc_mb":    float64(last.Alloc) / (1024 * 1024),
		"total_alloc_mb":    float64(last.TotalAlloc) / (1024 * 1024),
		"gc_cycles":         last.NumGC - first.NumGC,
		"gc_cpu_fraction":   last.GCCPUFraction,
	}
}

// PrintSummary prints a summary of memory usage statistics to the provided writer
func (s *MemorySampler) PrintSummary() {
	summary := s.GetSummary()

	fmt.Println("\nMemory Usage Summary:")
	fmt.Printf("  Samples Count:       %d\n", summary["samples_count"])
	fmt.Printf("  Sampling Duration:   %s\n", summary["sampling_duration"])
	fmt.Printf("  Sampling Interval:   %s\n", summary["sampling_interval"])
	fmt.Printf("  Peak Allocated:      %.2f MB\n", summary["peak_alloc_mb"])
	fmt.Printf("  Peak System Memory:  %.2f MB\n", summary["peak_sys_mb"])
	fmt.Printf("  Final Allocated:     %.2f MB\n", summary["final_alloc_mb"])
	fmt.Printf("  Total Allocated:     %.2f MB\n", summary["total_alloc_mb"])
	fmt.Printf("  GC Cycles:           %d\n", summary["gc_cycles"])
	fmt.Printf("  GC CPU Fraction:     %.2f%%\n", summary["gc_cpu_fraction"].(float64)*100)
}

// collectStats gathers current memory statistics
func (s *MemorySampler) collectStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		Alloc:         memStats.Alloc,
		TotalAlloc:    memStats.TotalAlloc,
		Sys:           memStats.Sys,
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction,
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		Time:          time.Now(),
	}
}
