// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Reporter is an interface for reporting transfer metrics
type Reporter interface {
	// ReportSummary writes a summary of the transfer metrics to the provided writer
	ReportSummary(w io.Writer, metrics *TransferMetrics) error
	
	// ReportDetailed writes a detailed report of the transfer metrics to the provided writer
	ReportDetailed(w io.Writer, metrics *TransferMetrics) error
	
	// ReportProgress writes a progress update for the transfer to the provided writer
	ReportProgress(w io.Writer, metrics *TransferMetrics) error
}

// TextReporter implements the Reporter interface for plain text output
type TextReporter struct{}

// NewTextReporter creates a new text reporter
func NewTextReporter() *TextReporter {
	return &TextReporter{}
}

// ReportSummary writes a summary of the transfer metrics to the provided writer
func (r *TextReporter) ReportSummary(w io.Writer, metrics *TransferMetrics) error {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	
	_, err := fmt.Fprintf(w, "Transfer Summary:\n")
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  ID:             %s\n", metrics.TransferID)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Task ID:        %s\n", metrics.TaskID)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Label:          %s\n", metrics.Label)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Source:         %s\n", metrics.SourceEndpoint)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Destination:    %s\n", metrics.DestEndpoint)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Status:         %s\n", metrics.Status)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Start Time:     %s\n", metrics.StartTime.Format(time.RFC3339))
	if err != nil {
		return err
	}
	
	if !metrics.EndTime.IsZero() {
		_, err = fmt.Fprintf(w, "  End Time:       %s\n", metrics.EndTime.Format(time.RFC3339))
		if err != nil {
			return err
		}
		
		duration := metrics.EndTime.Sub(metrics.StartTime)
		_, err = fmt.Fprintf(w, "  Duration:       %s\n", formatDuration(duration))
		if err != nil {
			return err
		}
	}
	
	_, err = fmt.Fprintf(w, "  Bytes:          %s / %s (%.1f%%)\n", 
		formatBytes(metrics.BytesTransferred), 
		formatBytes(metrics.TotalBytes),
		metrics.PercentComplete)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Files:          %d / %d\n", metrics.FilesTransferred, metrics.FilesTotal)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  Throughput:     %s/s (avg), %s/s (peak)\n",
		formatBytes(int64(metrics.AvgBytesPerSecond)),
		formatBytes(int64(metrics.PeakBytesPerSecond)))
	if err != nil {
		return err
	}
	
	if metrics.Status == "ACTIVE" && metrics.EstimatedTimeLeft > 0 {
		_, err = fmt.Fprintf(w, "  Est. Time Left: %s\n", formatDuration(metrics.EstimatedTimeLeft))
		if err != nil {
			return err
		}
	}
	
	if metrics.ErrorCount > 0 {
		_, err = fmt.Fprintf(w, "  Errors:         %d (%d retries)\n", metrics.ErrorCount, metrics.RetryCount)
		if err != nil {
			return err
		}
		
		if metrics.LastError != "" {
			_, err = fmt.Fprintf(w, "  Last Error:     %s\n", metrics.LastError)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// ReportDetailed writes a detailed report of the transfer metrics to the provided writer
func (r *TextReporter) ReportDetailed(w io.Writer, metrics *TransferMetrics) error {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	
	err := r.ReportSummary(w, metrics)
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "\nThroughput Samples:\n")
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  %-25s %-15s %-20s %-15s\n", "Time", "Bytes/sec", "Bytes Transferred", "Files Transferred")
	if err != nil {
		return err
	}
	
	_, err = fmt.Fprintf(w, "  %s\n", strings.Repeat("-", 80))
	if err != nil {
		return err
	}
	
	// Get a selection of samples (max 20 for readability)
	numSamples := len(metrics.ThroughputSamples)
	step := 1
	if numSamples > 20 {
		step = numSamples / 20
	}
	
	for i := 0; i < numSamples; i += step {
		sample := metrics.ThroughputSamples[i]
		_, err = fmt.Fprintf(w, "  %-25s %-15s %-20s %-15d\n",
			sample.Timestamp.Format(time.RFC3339),
			formatBytes(int64(sample.BytesPerSecond))+"/s",
			formatBytes(sample.BytesTransferred),
			sample.FilesTransferred)
		if err != nil {
			return err
		}
	}
	
	// Always include the last sample if we have one
	if numSamples > 0 && step > 1 {
		sample := metrics.ThroughputSamples[numSamples-1]
		_, err = fmt.Fprintf(w, "  %-25s %-15s %-20s %-15d\n",
			sample.Timestamp.Format(time.RFC3339),
			formatBytes(int64(sample.BytesPerSecond))+"/s",
			formatBytes(sample.BytesTransferred),
			sample.FilesTransferred)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// ReportProgress writes a progress update for the transfer to the provided writer
func (r *TextReporter) ReportProgress(w io.Writer, metrics *TransferMetrics) error {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	
	var progressBar string
	if metrics.PercentComplete > 0 {
		progressLen := 40
		completedLen := int(metrics.PercentComplete / 100 * float64(progressLen))
		
		progress := strings.Repeat("=", completedLen)
		if completedLen < progressLen {
			progress += ">"
			progress += strings.Repeat(" ", progressLen-completedLen-1)
		}
		
		progressBar = fmt.Sprintf("[%s] %.1f%%", progress, metrics.PercentComplete)
	} else {
		progressBar = "[                                        ] 0.0%"
	}
	
	// Format a compact one-line progress update
	var status string
	if metrics.Status == "ACTIVE" {
		if metrics.EstimatedTimeLeft > 0 {
			status = fmt.Sprintf("%s, %s left", progressBar, formatDuration(metrics.EstimatedTimeLeft))
		} else {
			status = fmt.Sprintf("%s", progressBar)
		}
	} else {
		status = fmt.Sprintf("%s, %s", progressBar, metrics.Status)
	}
	
	info := fmt.Sprintf("%s/s | %s / %s | %d / %d files",
		formatBytes(int64(metrics.BytesPerSecond)),
		formatBytes(metrics.BytesTransferred),
		formatBytes(metrics.TotalBytes),
		metrics.FilesTransferred,
		metrics.FilesTotal)
	
	_, err := fmt.Fprintf(w, "%s\n%s\n", status, info)
	return err
}

// JSONReporter implements the Reporter interface for JSON output
type JSONReporter struct{}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats a duration as human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	
	return fmt.Sprintf("%ds", seconds)
}