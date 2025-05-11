// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package deprecation

import (
	"fmt"
	"os"
	"sync"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// Configuration options for deprecation warnings
var (
	// DisableWarnings disables all deprecation warnings when set to true
	DisableWarnings = false

	// WarnOnce causes each specific deprecation warning to be logged only once
	// when set to true (the default)
	WarnOnce = true
)

// FeatureInfo stores information about a deprecated feature
type FeatureInfo struct {
	Name              string // Name of the deprecated feature
	DeprecatedIn      string // Version in which the feature was deprecated
	RemovalIn         string // Version in which the feature will be removed
	MigrationGuidance string // Guidance on how to migrate away from this feature
}

var (
	// Track issued warnings to avoid duplicates when WarnOnce is true
	warnedFeatures = make(map[string]struct{})
	warnMutex      sync.Mutex
)

// LogWarning logs a deprecation warning for a feature
func LogWarning(logger interfaces.Logger, featureName, deprecatedIn, removalIn, guidance string) {
	if DisableWarnings {
		return
	}

	// If WarnOnce is true, only log the first warning for each feature
	if WarnOnce {
		warnMutex.Lock()
		if _, warned := warnedFeatures[featureName]; warned {
			warnMutex.Unlock()
			return
		}
		warnedFeatures[featureName] = struct{}{}
		warnMutex.Unlock()
	}

	// Format the warning message
	message := formatWarningMessage(featureName, deprecatedIn, removalIn, guidance)

	// Log the warning
	if logger != nil {
		logger.Warn(message)
	} else {
		// Fallback to stderr if no logger is provided
		fmt.Fprintln(os.Stderr, "[WARN] "+message)
	}
}

// formatWarningMessage creates a consistent deprecation warning message
func formatWarningMessage(featureName, deprecatedIn, removalIn, guidance string) string {
	msg := fmt.Sprintf("DEPRECATED: %s was deprecated in %s", featureName, deprecatedIn)

	if removalIn != "" {
		msg += fmt.Sprintf(" and will be removed in %s", removalIn)
	}

	msg += "."

	if guidance != "" {
		msg += fmt.Sprintf(" %s", guidance)
	}

	return msg
}

// CreateFeatureInfo creates a new FeatureInfo object
func CreateFeatureInfo(name, deprecatedIn, removalIn, guidance string) FeatureInfo {
	return FeatureInfo{
		Name:              name,
		DeprecatedIn:      deprecatedIn,
		RemovalIn:         removalIn,
		MigrationGuidance: guidance,
	}
}

// LogFeatureWarning logs a deprecation warning using a FeatureInfo object
func LogFeatureWarning(logger interfaces.Logger, info FeatureInfo) {
	LogWarning(logger, info.Name, info.DeprecatedIn, info.RemovalIn, info.MigrationGuidance)
}
