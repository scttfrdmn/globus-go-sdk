// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// Package deprecation provides deprecation management for the Globus Go SDK.
//
// This package implements a deprecation system that matches the Python SDK's
// approach to handling deprecated functionality, providing clear warnings
// and migration guidance to users.
package deprecation

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// DeprecationManager manages deprecation warnings and tracking
type DeprecationManager struct {
	// warnings tracks which warnings have been issued to avoid spam
	warnings map[string]bool
	
	// mu protects the warnings map
	mu sync.RWMutex
	
	// enabled controls whether deprecation warnings are shown
	enabled bool
	
	// logger is used for outputting deprecation warnings
	logger Logger
}

// Logger interface for deprecation warnings
type Logger interface {
	Printf(format string, args ...interface{})
}

// defaultLogger is the default logger implementation
type defaultLogger struct{}

func (d *defaultLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// globalManager is the global deprecation manager instance
var globalManager *DeprecationManager
var once sync.Once

// GetManager returns the global deprecation manager
func GetManager() *DeprecationManager {
	once.Do(func() {
		globalManager = &DeprecationManager{
			warnings: make(map[string]bool),
			enabled:  os.Getenv("GLOBUS_SDK_DEPRECATION_WARNINGS") != "false",
			logger:   &defaultLogger{},
		}
	})
	return globalManager
}

// DeprecationInfo contains information about a deprecated feature
type DeprecationInfo struct {
	// Feature is the name of the deprecated feature
	Feature string
	
	// Version is the version in which the feature was deprecated
	DeprecatedIn string
	
	// RemovalVersion is the version in which the feature will be removed
	RemovalVersion string
	
	// Alternative is the recommended alternative to use
	Alternative string
	
	// Reason is the reason for deprecation
	Reason string
	
	// MoreInfo provides additional information or links
	MoreInfo string
}

// Warn issues a deprecation warning for the given feature
func (dm *DeprecationManager) Warn(info DeprecationInfo) {
	if !dm.enabled {
		return
	}
	
	key := fmt.Sprintf("%s:%s", info.Feature, info.DeprecatedIn)
	
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// Only warn once per feature/version combination
	if dm.warnings[key] {
		return
	}
	
	dm.warnings[key] = true
	
	// Format the warning message
	message := fmt.Sprintf("DEPRECATION WARNING: %s is deprecated as of v%s", 
		info.Feature, info.DeprecatedIn)
	
	if info.RemovalVersion != "" {
		message += fmt.Sprintf(" and will be removed in v%s", info.RemovalVersion)
	}
	
	if info.Alternative != "" {
		message += fmt.Sprintf(". Use %s instead", info.Alternative)
	}
	
	if info.Reason != "" {
		message += fmt.Sprintf(". Reason: %s", info.Reason)
	}
	
	if info.MoreInfo != "" {
		message += fmt.Sprintf(". More info: %s", info.MoreInfo)
	}
	
	dm.logger.Printf(message)
}

// WarnSimple issues a simple deprecation warning
func (dm *DeprecationManager) WarnSimple(feature, version, alternative string) {
	dm.Warn(DeprecationInfo{
		Feature:      feature,
		DeprecatedIn: version,
		Alternative:  alternative,
	})
}

// Enable enables deprecation warnings
func (dm *DeprecationManager) Enable() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.enabled = true
}

// Disable disables deprecation warnings
func (dm *DeprecationManager) Disable() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.enabled = false
}

// IsEnabled returns whether deprecation warnings are enabled
func (dm *DeprecationManager) IsEnabled() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.enabled
}

// SetLogger sets the logger for deprecation warnings
func (dm *DeprecationManager) SetLogger(logger Logger) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.logger = logger
}

// Reset clears all tracked warnings (useful for testing)
func (dm *DeprecationManager) Reset() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.warnings = make(map[string]bool)
}

// GetWarningCount returns the number of unique warnings that have been issued
func (dm *DeprecationManager) GetWarningCount() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return len(dm.warnings)
}

// Global convenience functions

// Warn issues a deprecation warning using the global manager
func Warn(info DeprecationInfo) {
	GetManager().Warn(info)
}

// WarnSimple issues a simple deprecation warning using the global manager
func WarnSimple(feature, version, alternative string) {
	GetManager().WarnSimple(feature, version, alternative)
}

// Enable enables deprecation warnings globally
func Enable() {
	GetManager().Enable()
}

// Disable disables deprecation warnings globally
func Disable() {
	GetManager().Disable()
}

// IsEnabled returns whether deprecation warnings are enabled globally
func IsEnabled() bool {
	return GetManager().IsEnabled()
}

// SetLogger sets the logger for deprecation warnings globally
func SetLogger(logger Logger) {
	GetManager().SetLogger(logger)
}

// Common deprecation scenarios for v3.60.0

// WarnLegacyClientInit warns about legacy client initialization
func WarnLegacyClientInit(service string) {
	WarnSimple(
		fmt.Sprintf("Legacy %s client initialization", service),
		"3.60.0",
		fmt.Sprintf("Use %s.NewClient() with unified configuration", service),
	)
}

// WarnLegacyErrorHandling warns about legacy error handling
func WarnLegacyErrorHandling() {
	WarnSimple(
		"Legacy error handling patterns",
		"3.60.0",
		"Use the new GlobusError type for consistent error handling",
	)
}

// WarnLegacyResponseStructure warns about legacy response structures
func WarnLegacyResponseStructure(service string) {
	WarnSimple(
		fmt.Sprintf("Legacy %s response structure", service),
		"3.60.0",
		fmt.Sprintf("Use the new Response[T] wrapper for %s responses", service),
	)
}

// DeprecationSchedule represents the deprecation schedule for features
type DeprecationSchedule struct {
	// Feature is the name of the feature
	Feature string
	
	// DeprecatedIn is the version where the feature was deprecated
	DeprecatedIn string
	
	// RemovalVersion is the version where the feature will be removed
	RemovalVersion string
	
	// Status is the current status of the deprecation
	Status DeprecationStatus
	
	// LastWarning is the timestamp of the last warning
	LastWarning time.Time
}

// DeprecationStatus represents the status of a deprecated feature
type DeprecationStatus string

const (
	// StatusActive indicates the feature is active but deprecated
	StatusActive DeprecationStatus = "active"
	
	// StatusPendingRemoval indicates the feature is scheduled for removal
	StatusPendingRemoval DeprecationStatus = "pending_removal"
	
	// StatusRemoved indicates the feature has been removed
	StatusRemoved DeprecationStatus = "removed"
)

// ScheduleManager manages the deprecation schedule
type ScheduleManager struct {
	schedule map[string]DeprecationSchedule
	mu       sync.RWMutex
}

// NewScheduleManager creates a new deprecation schedule manager
func NewScheduleManager() *ScheduleManager {
	return &ScheduleManager{
		schedule: make(map[string]DeprecationSchedule),
	}
}

// AddDeprecation adds a feature to the deprecation schedule
func (sm *ScheduleManager) AddDeprecation(feature, deprecatedIn, removalVersion string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.schedule[feature] = DeprecationSchedule{
		Feature:        feature,
		DeprecatedIn:   deprecatedIn,
		RemovalVersion: removalVersion,
		Status:         StatusActive,
	}
}

// GetDeprecation gets deprecation information for a feature
func (sm *ScheduleManager) GetDeprecation(feature string) (DeprecationSchedule, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	schedule, exists := sm.schedule[feature]
	return schedule, exists
}

// IsDeprecated checks if a feature is deprecated
func (sm *ScheduleManager) IsDeprecated(feature string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	schedule, exists := sm.schedule[feature]
	return exists && schedule.Status == StatusActive
}

// MarkRemoved marks a feature as removed
func (sm *ScheduleManager) MarkRemoved(feature string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if schedule, exists := sm.schedule[feature]; exists {
		schedule.Status = StatusRemoved
		sm.schedule[feature] = schedule
	}
}

// GetActiveDeprecations returns all active deprecations
func (sm *ScheduleManager) GetActiveDeprecations() []DeprecationSchedule {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var active []DeprecationSchedule
	for _, schedule := range sm.schedule {
		if schedule.Status == StatusActive {
			active = append(active, schedule)
		}
	}
	
	return active
}

// globalSchedule is the global deprecation schedule manager
var globalSchedule *ScheduleManager
var scheduleOnce sync.Once

// GetScheduleManager returns the global deprecation schedule manager
func GetScheduleManager() *ScheduleManager {
	scheduleOnce.Do(func() {
		globalSchedule = NewScheduleManager()
		
		// Initialize with known deprecations for v3.60.0
		globalSchedule.AddDeprecation("Legacy client initialization", "3.60.0", "4.0.0")
		globalSchedule.AddDeprecation("Legacy error handling", "3.60.0", "4.0.0")
		globalSchedule.AddDeprecation("Legacy response structures", "3.60.0", "4.0.0")
	})
	return globalSchedule
}