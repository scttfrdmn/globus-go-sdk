// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ProgressBar represents a progress bar that can be updated over time
type ProgressBar struct {
	// Configuration
	writer            io.Writer
	width             int
	refreshRate       time.Duration
	showSpeed         bool
	showETA           bool
	showValues        bool
	showPercent       bool
	hideAfterComplete bool

	// State
	current     int64
	total       int64
	startTime   time.Time
	lastUpdate  time.Time
	speed       float64
	message     string
	completed   bool
	mu          sync.Mutex
	done        chan struct{}
	lastRefresh time.Time
}

// ProgressBarOption is a function that configures a ProgressBar
type ProgressBarOption func(*ProgressBar)

// WithWidth sets the width of the progress bar
func WithWidth(width int) ProgressBarOption {
	return func(p *ProgressBar) {
		p.width = width
	}
}

// WithRefreshRate sets how often the progress bar should be refreshed
func WithRefreshRate(rate time.Duration) ProgressBarOption {
	return func(p *ProgressBar) {
		p.refreshRate = rate
	}
}

// WithSpeed enables/disables showing speed information
func WithSpeed(show bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.showSpeed = show
	}
}

// WithETA enables/disables showing estimated time remaining
func WithETA(show bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.showETA = show
	}
}

// WithValues enables/disables showing current/total values
func WithValues(show bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.showValues = show
	}
}

// WithPercent enables/disables showing percentage
func WithPercent(show bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.showPercent = show
	}
}

// WithHideAfterComplete enables/disables hiding the progress bar after completion
func WithHideAfterComplete(hide bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.hideAfterComplete = hide
	}
}

// WithMessage sets a message to display alongside the progress bar
func WithMessage(message string) ProgressBarOption {
	return func(p *ProgressBar) {
		p.message = message
	}
}

// NewProgressBar creates a new progress bar
func NewProgressBar(writer io.Writer, total int64, opts ...ProgressBarOption) *ProgressBar {
	bar := &ProgressBar{
		writer:      writer,
		total:       total,
		width:       40,
		refreshRate: 200 * time.Millisecond,
		showSpeed:   true,
		showETA:     true,
		showValues:  true,
		showPercent: true,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
		lastRefresh: time.Now(),
		done:        make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(bar)
	}

	return bar
}

// Start begins displaying the progress bar
func (p *ProgressBar) Start() {
	p.mu.Lock()
	p.startTime = time.Now()
	p.lastUpdate = time.Now()
	p.lastRefresh = time.Now().Add(-p.refreshRate) // Force immediate refresh
	p.mu.Unlock()

	go p.refreshLoop()
}

// refreshLoop periodically refreshes the progress bar display
func (p *ProgressBar) refreshLoop() {
	ticker := time.NewTicker(p.refreshRate / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			if time.Since(p.lastRefresh) >= p.refreshRate {
				p.refresh()
				p.lastRefresh = time.Now()
			}
			p.mu.Unlock()
		case <-p.done:
			return
		}
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	now := time.Now()
	elapsed := now.Sub(p.lastUpdate).Seconds()

	// Calculate speed (e.g., bytes per second)
	if elapsed > 0 && p.lastUpdate != p.startTime {
		p.speed = float64(current-p.current) / elapsed
	}

	p.lastUpdate = now
}

// SetMessage sets a message to display alongside the progress bar
func (p *ProgressBar) SetMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.message = message
}

// Complete marks the progress bar as completed
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	p.completed = true
	p.current = p.total
	p.refresh()
	p.mu.Unlock()

	close(p.done)
}

// Stop stops the progress bar display
func (p *ProgressBar) Stop() {
	close(p.done)
}

// refresh redraws the progress bar
func (p *ProgressBar) refresh() {
	// Calculate percentage
	var percent float64
	if p.total > 0 {
		percent = float64(p.current) / float64(p.total) * 100
	}

	// Draw the progress bar
	width := p.width
	completed := int(float64(width) * float64(p.current) / float64(p.total))
	if completed > width {
		completed = width
	}

	// Create the bar
	bar := "["
	if completed > 0 {
		bar += strings.Repeat("=", completed-1)
		if completed < width {
			bar += ">"
		} else {
			bar += "="
		}
	}
	bar += strings.Repeat(" ", width-completed)
	bar += "]"

	// Format the output line
	line := bar

	// Add percentage if enabled
	if p.showPercent {
		line += fmt.Sprintf(" %.1f%%", percent)
	}

	// Add values if enabled
	if p.showValues {
		line += fmt.Sprintf(" %s/%s", formatBytes(p.current), formatBytes(p.total))
	}

	// Add speed if enabled
	if p.showSpeed && p.speed > 0 {
		line += fmt.Sprintf(" %s/s", formatBytes(int64(p.speed)))
	}

	// Add ETA if enabled
	if p.showETA && p.speed > 0 && p.current < p.total {
		remaining := float64(p.total-p.current) / p.speed
		if remaining > 0 {
			eta := time.Duration(remaining * float64(time.Second))
			line += fmt.Sprintf(" ETA: %s", formatDuration(eta))
		}
	}

	// Add message if present
	if p.message != "" {
		line += " " + p.message
	}

	// Print the line with a carriage return (to stay on the same line)
	fmt.Fprintf(p.writer, "\r%s", line)

	// Add a newline if completed and not hiding
	if p.completed && !p.hideAfterComplete {
		fmt.Fprintln(p.writer)
	}
}
