// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package deprecation

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// MockLogger implements the interfaces.Logger interface for testing
type MockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		debugMsgs: make([]string, 0),
		infoMsgs:  make([]string, 0),
		warnMsgs:  make([]string, 0),
		errorMsgs: make([]string, 0),
	}
}

func (l *MockLogger) Debug(format string, args ...interface{}) {
	l.debugMsgs = append(l.debugMsgs, format)
}

func (l *MockLogger) Info(format string, args ...interface{}) {
	l.infoMsgs = append(l.infoMsgs, format)
}

func (l *MockLogger) Warn(format string, args ...interface{}) {
	l.warnMsgs = append(l.warnMsgs, format)
}

func (l *MockLogger) Error(format string, args ...interface{}) {
	l.errorMsgs = append(l.errorMsgs, format)
}

func TestLogWarning(t *testing.T) {
	// Reset state
	DisableWarnings = false
	WarnOnce = true
	warnedFeatures = make(map[string]struct{})

	// Create a mock logger
	logger := NewMockLogger()

	// Test basic warning
	LogWarning(logger, "TestFeature", "v1.0.0", "v2.0.0", "Use NewFeature instead.")

	if len(logger.warnMsgs) != 1 {
		t.Errorf("Expected 1 warning message, got %d", len(logger.warnMsgs))
	}

	expectedMsg := "DEPRECATED: TestFeature was deprecated in v1.0.0 and will be removed in v2.0.0. Use NewFeature instead."
	if logger.warnMsgs[0] != expectedMsg {
		t.Errorf("Expected warning message:\n%s\nGot:\n%s", expectedMsg, logger.warnMsgs[0])
	}

	// Test WarnOnce = true (second warning for same feature should be ignored)
	LogWarning(logger, "TestFeature", "v1.0.0", "v2.0.0", "Use NewFeature instead.")

	if len(logger.warnMsgs) != 1 {
		t.Errorf("With WarnOnce=true, expected still 1 warning message, got %d", len(logger.warnMsgs))
	}

	// Test WarnOnce = false
	WarnOnce = false
	LogWarning(logger, "TestFeature", "v1.0.0", "v2.0.0", "Use NewFeature instead.")

	if len(logger.warnMsgs) != 2 {
		t.Errorf("With WarnOnce=false, expected 2 warning messages, got %d", len(logger.warnMsgs))
	}

	// Test DisableWarnings = true
	DisableWarnings = true
	LogWarning(logger, "AnotherFeature", "v1.0.0", "v2.0.0", "Use NewFeature instead.")

	if len(logger.warnMsgs) != 2 {
		t.Errorf("With DisableWarnings=true, expected no new warnings, got %d", len(logger.warnMsgs))
	}

	// Test nil logger (should not panic)
	// Create a temporary file to capture os.Stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	DisableWarnings = false
	LogWarning(nil, "NilLoggerFeature", "v1.0.0", "v2.0.0", "Use NewFeature instead.")

	// Restore stderr and close the pipe
	w.Close()
	os.Stderr = oldStderr

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "NilLoggerFeature") {
		t.Error("Expected warning to be written to stderr with nil logger")
	}
}

func TestLogFeatureWarning(t *testing.T) {
	// Reset state
	DisableWarnings = false
	WarnOnce = true
	warnedFeatures = make(map[string]struct{})

	// Create a mock logger
	logger := NewMockLogger()

	// Create feature info
	info := CreateFeatureInfo(
		"FeatureInfoTest",
		"v1.0.0",
		"v2.0.0",
		"Use NewFeature instead.",
	)

	// Test LogFeatureWarning
	LogFeatureWarning(logger, info)

	if len(logger.warnMsgs) != 1 {
		t.Errorf("Expected 1 warning message, got %d", len(logger.warnMsgs))
	}

	expectedMsg := "DEPRECATED: FeatureInfoTest was deprecated in v1.0.0 and will be removed in v2.0.0. Use NewFeature instead."
	if logger.warnMsgs[0] != expectedMsg {
		t.Errorf("Expected warning message:\n%s\nGot:\n%s", expectedMsg, logger.warnMsgs[0])
	}
}

func TestFormatWarningMessage(t *testing.T) {
	tests := []struct {
		name         string
		featureName  string
		deprecatedIn string
		removalIn    string
		guidance     string
		expected     string
	}{
		{
			name:         "Complete message",
			featureName:  "TestFeature",
			deprecatedIn: "v1.0.0",
			removalIn:    "v2.0.0",
			guidance:     "Use NewFeature instead.",
			expected:     "DEPRECATED: TestFeature was deprecated in v1.0.0 and will be removed in v2.0.0. Use NewFeature instead.",
		},
		{
			name:         "No removal version",
			featureName:  "TestFeature",
			deprecatedIn: "v1.0.0",
			removalIn:    "",
			guidance:     "Use NewFeature instead.",
			expected:     "DEPRECATED: TestFeature was deprecated in v1.0.0. Use NewFeature instead.",
		},
		{
			name:         "No guidance",
			featureName:  "TestFeature",
			deprecatedIn: "v1.0.0",
			removalIn:    "v2.0.0",
			guidance:     "",
			expected:     "DEPRECATED: TestFeature was deprecated in v1.0.0 and will be removed in v2.0.0.",
		},
		{
			name:         "Minimal",
			featureName:  "TestFeature",
			deprecatedIn: "v1.0.0",
			removalIn:    "",
			guidance:     "",
			expected:     "DEPRECATED: TestFeature was deprecated in v1.0.0.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatWarningMessage(
				test.featureName,
				test.deprecatedIn,
				test.removalIn,
				test.guidance,
			)

			if result != test.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", test.expected, result)
			}
		})
	}
}
