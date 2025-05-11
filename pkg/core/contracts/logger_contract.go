// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyLoggerContract verifies that a Logger implementation
// satisfies the behavioral contract of the interface.
func VerifyLoggerContract(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	t.Run("Debug method", func(t *testing.T) {
		verifyDebugMethod(t, logger)
	})

	t.Run("Info method", func(t *testing.T) {
		verifyInfoMethod(t, logger)
	})

	t.Run("Warn method", func(t *testing.T) {
		verifyWarnMethod(t, logger)
	})

	t.Run("Error method", func(t *testing.T) {
		verifyErrorMethod(t, logger)
	})

	t.Run("Format string handling", func(t *testing.T) {
		verifyFormatStringHandling(t, logger)
	})
}

// verifyDebugMethod tests the behavior of the Debug method
func verifyDebugMethod(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	// Debug should not panic
	logger.Debug("Debug test message")
	logger.Debug("Debug message with one arg: %s", "arg1")
	logger.Debug("Debug message with multiple args: %s, %d, %v", "arg1", 42, true)
}

// verifyInfoMethod tests the behavior of the Info method
func verifyInfoMethod(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	// Info should not panic
	logger.Info("Info test message")
	logger.Info("Info message with one arg: %s", "arg1")
	logger.Info("Info message with multiple args: %s, %d, %v", "arg1", 42, true)
}

// verifyWarnMethod tests the behavior of the Warn method
func verifyWarnMethod(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	// Warn should not panic
	logger.Warn("Warning test message")
	logger.Warn("Warning message with one arg: %s", "arg1")
	logger.Warn("Warning message with multiple args: %s, %d, %v", "arg1", 42, true)
}

// verifyErrorMethod tests the behavior of the Error method
func verifyErrorMethod(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	// Error should not panic
	logger.Error("Error test message")
	logger.Error("Error message with one arg: %s", "arg1")
	logger.Error("Error message with multiple args: %s, %d, %v", "arg1", 42, true)
}

// verifyFormatStringHandling tests handling of format strings and arguments
func verifyFormatStringHandling(t *testing.T, logger interfaces.Logger) {
	t.Helper()

	// Mismatched format specifiers should not panic
	logger.Debug("Debug message with mismatched specifiers: %s, %d", "arg1")
	logger.Info("Info message with mismatched specifiers: %s, %d", "arg1")
	logger.Warn("Warning message with mismatched specifiers: %s, %d", "arg1")
	logger.Error("Error message with mismatched specifiers: %s, %d", "arg1")

	// Extra arguments should not panic
	logger.Debug("Debug message with extra args", "arg1", "arg2")
	logger.Info("Info message with extra args", "arg1", "arg2")
	logger.Warn("Warning message with extra args", "arg1", "arg2")
	logger.Error("Error message with extra args", "arg1", "arg2")

	// Special format characters should not panic
	logger.Debug("Debug message with %%, %[1]s, %*s", "arg1", 5, "arg2")
	logger.Info("Info message with %%, %[1]s, %*s", "arg1", 5, "arg2")
	logger.Warn("Warning message with %%, %[1]s, %*s", "arg1", 5, "arg2")
	logger.Error("Error message with %%, %[1]s, %*s", "arg1", 5, "arg2")

	// Nil arguments should not panic
	var nilPtr *string
	logger.Debug("Debug message with nil arg: %v", nilPtr)
	logger.Info("Info message with nil arg: %v", nilPtr)
	logger.Warn("Warning message with nil arg: %v", nilPtr)
	logger.Error("Error message with nil arg: %v", nilPtr)
}
