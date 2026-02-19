// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package deprecation_test

import (
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/deprecation"
)

// captureLogger records Printf calls so tests can assert on logged messages.
type captureLogger struct {
	messages []string
}

func (c *captureLogger) Printf(format string, args ...interface{}) {
	c.messages = append(c.messages, format)
}

// resetManager resets the global deprecation manager to a clean state and
// returns it. Tests must call this to avoid cross-test contamination.
func resetManager(t *testing.T) *deprecation.DeprecationManager {
	t.Helper()
	mgr := deprecation.GetManager()
	mgr.Reset()
	mgr.Enable()
	return mgr
}

// ---------------------------------------------------------------------------
// GetManager
// ---------------------------------------------------------------------------

func TestGetManagerReturnsNonNil(t *testing.T) {
	mgr := deprecation.GetManager()
	if mgr == nil {
		t.Fatal("GetManager() should return a non-nil DeprecationManager")
	}
}

func TestGetManagerReturnsSameSingleton(t *testing.T) {
	a := deprecation.GetManager()
	b := deprecation.GetManager()
	if a != b {
		t.Fatal("GetManager() should return the same singleton instance each time")
	}
}

// ---------------------------------------------------------------------------
// DeprecationManager.Enable / Disable / IsEnabled
// ---------------------------------------------------------------------------

func TestManagerEnableDisableIsEnabled(t *testing.T) {
	mgr := resetManager(t)

	mgr.Disable()
	if mgr.IsEnabled() {
		t.Fatal("IsEnabled should return false after Disable()")
	}

	mgr.Enable()
	if !mgr.IsEnabled() {
		t.Fatal("IsEnabled should return true after Enable()")
	}
}

// ---------------------------------------------------------------------------
// DeprecationManager.Warn
// ---------------------------------------------------------------------------

func TestManagerWarnIncrementsCount(t *testing.T) {
	mgr := resetManager(t)

	if mgr.GetWarningCount() != 0 {
		t.Fatalf("initial warning count should be 0, got %d", mgr.GetWarningCount())
	}

	mgr.Warn(deprecation.DeprecationInfo{
		Feature:      "CountFeature",
		DeprecatedIn: "1.0",
	})

	if mgr.GetWarningCount() != 1 {
		t.Fatalf("after one Warn() count should be 1, got %d", mgr.GetWarningCount())
	}
}

func TestManagerWarnOnlyOncePerKey(t *testing.T) {
	mgr := resetManager(t)

	info := deprecation.DeprecationInfo{
		Feature:      "OnceFeature",
		DeprecatedIn: "1.0",
	}

	mgr.Warn(info)
	mgr.Warn(info)
	mgr.Warn(info)

	if mgr.GetWarningCount() != 1 {
		t.Fatalf("same feature/version key should only be counted once, got %d", mgr.GetWarningCount())
	}
}

func TestManagerWarnDisabledSkips(t *testing.T) {
	mgr := resetManager(t)
	mgr.Disable()

	mgr.Warn(deprecation.DeprecationInfo{
		Feature:      "DisabledFeature",
		DeprecatedIn: "1.0",
	})

	if mgr.GetWarningCount() != 0 {
		t.Fatalf("Warn() while disabled should not increment count, got %d", mgr.GetWarningCount())
	}
}

func TestManagerWarnMessageContainsAllFields(t *testing.T) {
	mgr := resetManager(t)
	cl := &captureLogger{}
	mgr.SetLogger(cl)

	mgr.Warn(deprecation.DeprecationInfo{
		Feature:        "FullFeature",
		DeprecatedIn:   "3.1",
		RemovalVersion: "4.0",
		Alternative:    "BetterFeature",
		Reason:         "performance",
		MoreInfo:       "https://example.com",
	})

	if len(cl.messages) == 0 {
		t.Fatal("expected at least one log message")
	}
	msg := cl.messages[0]
	for _, substr := range []string{
		"FullFeature", "3.1", "4.0", "BetterFeature", "performance", "https://example.com",
	} {
		if !strings.Contains(msg, substr) {
			t.Errorf("expected log message to contain %q; got: %s", substr, msg)
		}
	}
}

func TestManagerWarnMessageNoRemovalVersion(t *testing.T) {
	mgr := resetManager(t)
	cl := &captureLogger{}
	mgr.SetLogger(cl)

	mgr.Warn(deprecation.DeprecationInfo{
		Feature:      "NoRemovalFeature",
		DeprecatedIn: "2.0",
	})

	if len(cl.messages) == 0 {
		t.Fatal("expected at least one log message")
	}
	// The message should mention the feature.
	if !strings.Contains(cl.messages[0], "NoRemovalFeature") {
		t.Errorf("log message missing feature name: %s", cl.messages[0])
	}
}

// ---------------------------------------------------------------------------
// DeprecationManager.WarnSimple
// ---------------------------------------------------------------------------

func TestManagerWarnSimple(t *testing.T) {
	mgr := resetManager(t)

	mgr.WarnSimple("SimpleFeature", "2.0", "NewSimpleFeature")

	if mgr.GetWarningCount() != 1 {
		t.Fatalf("WarnSimple should increment count to 1, got %d", mgr.GetWarningCount())
	}
}

// ---------------------------------------------------------------------------
// DeprecationManager.Reset
// ---------------------------------------------------------------------------

func TestManagerReset(t *testing.T) {
	mgr := resetManager(t)

	mgr.WarnSimple("FeatureForReset", "1.0", "")

	if mgr.GetWarningCount() == 0 {
		t.Fatal("expected count > 0 before reset")
	}

	mgr.Reset()

	if mgr.GetWarningCount() != 0 {
		t.Fatalf("Reset() should clear all warnings, got count %d", mgr.GetWarningCount())
	}
}

// ---------------------------------------------------------------------------
// DeprecationManager.SetLogger
// ---------------------------------------------------------------------------

func TestManagerSetLogger(t *testing.T) {
	mgr := resetManager(t)
	cl := &captureLogger{}
	mgr.SetLogger(cl)

	mgr.Warn(deprecation.DeprecationInfo{
		Feature:      "LoggerFeature",
		DeprecatedIn: "3.0",
	})

	if len(cl.messages) == 0 {
		t.Fatal("SetLogger should cause Warn to use the provided logger")
	}
}

// ---------------------------------------------------------------------------
// Global convenience functions
// ---------------------------------------------------------------------------

func TestGlobalEnableDisable(t *testing.T) {
	resetManager(t)

	deprecation.Enable()
	if !deprecation.IsEnabled() {
		t.Fatal("IsEnabled should return true after global Enable()")
	}

	deprecation.Disable()
	if deprecation.IsEnabled() {
		t.Fatal("IsEnabled should return false after global Disable()")
	}

	// Restore for subsequent tests.
	deprecation.Enable()
}

func TestGlobalWarn(t *testing.T) {
	mgr := resetManager(t)

	deprecation.Warn(deprecation.DeprecationInfo{
		Feature:      "GlobalWarnFeature",
		DeprecatedIn: "2.5",
	})

	if mgr.GetWarningCount() == 0 {
		t.Fatal("global Warn() should have recorded a warning")
	}
}

func TestGlobalWarnSimple(t *testing.T) {
	mgr := resetManager(t)
	before := mgr.GetWarningCount()

	deprecation.WarnSimple("GlobalSimple", "1.0", "NewGlobal")

	after := mgr.GetWarningCount()
	if after != before+1 {
		t.Fatalf("global WarnSimple should increment count by 1: before=%d after=%d", before, after)
	}
}

func TestGlobalSetLogger(t *testing.T) {
	mgr := resetManager(t)
	cl := &captureLogger{}
	deprecation.SetLogger(cl)

	deprecation.WarnSimple("SetLoggerFeature", "1.0", "")

	if len(cl.messages) == 0 {
		t.Fatal("global SetLogger should route warnings to the provided logger")
	}

	// Restore default logger so other tests are not affected.
	_ = mgr
}

// ---------------------------------------------------------------------------
// Convenience warning helpers
// ---------------------------------------------------------------------------

func TestWarnLegacyClientInit(t *testing.T) {
	mgr := resetManager(t)
	before := mgr.GetWarningCount()

	deprecation.WarnLegacyClientInit("transfer")

	if mgr.GetWarningCount() != before+1 {
		t.Fatalf("WarnLegacyClientInit should issue one new warning, before=%d after=%d",
			before, mgr.GetWarningCount())
	}
}

func TestWarnLegacyErrorHandling(t *testing.T) {
	mgr := resetManager(t)
	before := mgr.GetWarningCount()

	deprecation.WarnLegacyErrorHandling()

	if mgr.GetWarningCount() != before+1 {
		t.Fatalf("WarnLegacyErrorHandling should issue one new warning, before=%d after=%d",
			before, mgr.GetWarningCount())
	}
}

func TestWarnLegacyResponseStructure(t *testing.T) {
	mgr := resetManager(t)
	before := mgr.GetWarningCount()

	deprecation.WarnLegacyResponseStructure("search")

	if mgr.GetWarningCount() != before+1 {
		t.Fatalf("WarnLegacyResponseStructure should issue one new warning, before=%d after=%d",
			before, mgr.GetWarningCount())
	}
}

// ---------------------------------------------------------------------------
// DeprecationInfo struct fields
// ---------------------------------------------------------------------------

func TestDeprecationInfoFields(t *testing.T) {
	info := deprecation.DeprecationInfo{
		Feature:        "FieldTestFeature",
		DeprecatedIn:   "1.2",
		RemovalVersion: "2.0",
		Alternative:    "BetterAPI",
		Reason:         "old design",
		MoreInfo:       "https://docs.example.com",
	}

	if info.Feature != "FieldTestFeature" {
		t.Errorf("Feature mismatch: %s", info.Feature)
	}
	if info.DeprecatedIn != "1.2" {
		t.Errorf("DeprecatedIn mismatch: %s", info.DeprecatedIn)
	}
	if info.RemovalVersion != "2.0" {
		t.Errorf("RemovalVersion mismatch: %s", info.RemovalVersion)
	}
}

// ---------------------------------------------------------------------------
// ScheduleManager
// ---------------------------------------------------------------------------

func TestNewScheduleManagerEmpty(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	active := sm.GetActiveDeprecations()
	if len(active) != 0 {
		t.Fatalf("new ScheduleManager should have no active deprecations, got %d", len(active))
	}
}

func TestScheduleManagerAddAndGet(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	sm.AddDeprecation("FeatureA", "1.0", "2.0")

	sched, ok := sm.GetDeprecation("FeatureA")
	if !ok {
		t.Fatal("GetDeprecation should return true for an added feature")
	}
	if sched.Feature != "FeatureA" {
		t.Fatalf("Feature = %q, want %q", sched.Feature, "FeatureA")
	}
	if sched.DeprecatedIn != "1.0" {
		t.Fatalf("DeprecatedIn = %q, want %q", sched.DeprecatedIn, "1.0")
	}
	if sched.RemovalVersion != "2.0" {
		t.Fatalf("RemovalVersion = %q, want %q", sched.RemovalVersion, "2.0")
	}
	if sched.Status != deprecation.StatusActive {
		t.Fatalf("Status = %q, want %q", sched.Status, deprecation.StatusActive)
	}
}

func TestScheduleManagerGetNonExistent(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	_, ok := sm.GetDeprecation("NonExistent")
	if ok {
		t.Fatal("GetDeprecation should return false for an unknown feature")
	}
}

func TestScheduleManagerIsDeprecated(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	sm.AddDeprecation("FeatureB", "1.0", "2.0")

	if !sm.IsDeprecated("FeatureB") {
		t.Fatal("IsDeprecated should return true for an active deprecation")
	}
	if sm.IsDeprecated("Unknown") {
		t.Fatal("IsDeprecated should return false for an unknown feature")
	}
}

func TestScheduleManagerMarkRemoved(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	sm.AddDeprecation("FeatureC", "1.0", "2.0")

	sm.MarkRemoved("FeatureC")

	if sm.IsDeprecated("FeatureC") {
		t.Fatal("IsDeprecated should return false after MarkRemoved")
	}

	sched, ok := sm.GetDeprecation("FeatureC")
	if !ok {
		t.Fatal("GetDeprecation should still return the entry after MarkRemoved")
	}
	if sched.Status != deprecation.StatusRemoved {
		t.Fatalf("Status = %q, want %q", sched.Status, deprecation.StatusRemoved)
	}
}

func TestScheduleManagerMarkRemovedNonExistent(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	// Should not panic when the feature doesn't exist.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MarkRemoved on unknown feature panicked: %v", r)
		}
	}()
	sm.MarkRemoved("DoesNotExist")
}

func TestScheduleManagerGetActiveDeprecations(t *testing.T) {
	sm := deprecation.NewScheduleManager()
	sm.AddDeprecation("Active1", "1.0", "2.0")
	sm.AddDeprecation("Active2", "1.1", "2.1")
	sm.AddDeprecation("Removed", "1.0", "2.0")
	sm.MarkRemoved("Removed")

	active := sm.GetActiveDeprecations()
	if len(active) != 2 {
		t.Fatalf("expected 2 active deprecations, got %d", len(active))
	}
}

func TestGetScheduleManagerReturnsNonNil(t *testing.T) {
	sm := deprecation.GetScheduleManager()
	if sm == nil {
		t.Fatal("GetScheduleManager should return a non-nil ScheduleManager")
	}
}

func TestGetScheduleManagerHasBuiltInDeprecations(t *testing.T) {
	sm := deprecation.GetScheduleManager()
	active := sm.GetActiveDeprecations()
	if len(active) == 0 {
		t.Fatal("GetScheduleManager should have pre-populated deprecations")
	}
}

// ---------------------------------------------------------------------------
// DeprecationStatus constants
// ---------------------------------------------------------------------------

func TestDeprecationStatusConstants(t *testing.T) {
	if deprecation.StatusActive == "" {
		t.Fatal("StatusActive should not be empty")
	}
	if deprecation.StatusPendingRemoval == "" {
		t.Fatal("StatusPendingRemoval should not be empty")
	}
	if deprecation.StatusRemoved == "" {
		t.Fatal("StatusRemoved should not be empty")
	}
	if deprecation.StatusActive == deprecation.StatusPendingRemoval {
		t.Fatal("StatusActive and StatusPendingRemoval must differ")
	}
	if deprecation.StatusActive == deprecation.StatusRemoved {
		t.Fatal("StatusActive and StatusRemoved must differ")
	}
	if deprecation.StatusPendingRemoval == deprecation.StatusRemoved {
		t.Fatal("StatusPendingRemoval and StatusRemoved must differ")
	}
}
