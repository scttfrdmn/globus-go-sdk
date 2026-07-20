// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Version is the current version of the Globus Go SDK v4.
//
// This tracks feature parity with the upstream Python globus-sdk. The v4 module
// currently implements the API surface of Python globus-sdk v4.5.0. See
// .github/upstream-versions.json for the ported-vs-seen parity tracking.
const Version = "4.5.0"

// VersionInfo provides additional info about the build
type VersionInfo struct {
	Version     string `json:"version"`     // Semver version
	GitCommit   string `json:"gitCommit"`   // Git commit hash
	BuildDate   string `json:"buildDate"`   // Build date
	GoVersion   string `json:"goVersion"`   // Go version used for building
	FullVersion string `json:"fullVersion"` // Full version with build details
}

// GetInfo returns detailed version information
func GetInfo() VersionInfo {
	info := VersionInfo{
		Version: Version,
	}

	// Try to extract build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info.GoVersion = buildInfo.GoVersion

		// Extract revision info from build settings
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.GitCommit = setting.Value
			case "vcs.time":
				info.BuildDate = setting.Value
			}
		}
	}

	// Format full version
	parts := []string{info.Version}
	if info.GitCommit != "" {
		parts = append(parts, fmt.Sprintf("commit:%s", info.GitCommit[:8]))
	}
	if info.BuildDate != "" {
		parts = append(parts, fmt.Sprintf("built:%s", info.BuildDate))
	}
	info.FullVersion = strings.Join(parts, " ")

	return info
}

// UserAgent returns the appropriate User-Agent header string for the SDK
func UserAgent() string {
	info := GetInfo()
	return fmt.Sprintf("Globus-Go-SDK/%s", info.Version)
}
