// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// APIVersion represents a Globus API version with optional components
type APIVersion struct {
	Service string // Service name (e.g., "auth", "transfer")
	Major   int    // Major version number
	Minor   int    // Minor version number
	Patch   int    // Patch version (if specified)
	Beta    bool   // Whether this is a beta version
}

// String returns a string representation of the API version
func (v APIVersion) String() string {
	if v.Beta {
		return fmt.Sprintf("%s/Beta", v.Service)
	}

	if v.Patch > 0 {
		return fmt.Sprintf("%s/v%d.%d.%d", v.Service, v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%s/v%d.%d", v.Service, v.Major, v.Minor)
}

// Endpoint returns the base URL endpoint for this API version
func (v APIVersion) Endpoint() string {
	switch v.Service {
	case "auth":
		return fmt.Sprintf("https://auth.globus.org/v%d/", v.Major)
	case "transfer":
		return fmt.Sprintf("https://transfer.api.globus.org/v%d.%d/", v.Major, v.Minor)
	case "search":
		return fmt.Sprintf("https://search.api.globus.org/v%d/", v.Major)
	case "groups":
		return fmt.Sprintf("https://groups.api.globus.org/v%d/", v.Major)
	case "flows":
		return "https://flows.globus.org/api/"
	default:
		return ""
	}
}

// SupportedAPIVersions defines the API versions supported by the SDK
var SupportedAPIVersions = map[string]APIVersion{
	"auth": {
		Service: "auth",
		Major:   2,
		Minor:   0,
	},
	"transfer": {
		Service: "transfer",
		Major:   0,
		Minor:   10,
	},
	"search": {
		Service: "search",
		Major:   1,
		Minor:   0,
	},
	"groups": {
		Service: "groups",
		Major:   2,
		Minor:   0,
	},
	"flows": {
		Service: "flows",
		Major:   0,
		Minor:   0,
		Beta:    true,
	},
}

// ParseAPIVersion parses a version string like "v2" or "v0.10" into an APIVersion
func ParseAPIVersion(service, version string) (APIVersion, error) {
	result := APIVersion{
		Service: service,
	}

	// Beta version
	if strings.ToLower(version) == "beta" {
		result.Beta = true
		return result, nil
	}

	// Regular version (v2, v0.10, etc.)
	re := regexp.MustCompile(`^v(\d+)(?:\.(\d+))?(?:\.(\d+))?$`)
	matches := re.FindStringSubmatch(version)
	if matches == nil {
		return result, fmt.Errorf("invalid version format: %s", version)
	}

	// Parse major version (required)
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return result, fmt.Errorf("invalid major version: %s", matches[1])
	}
	result.Major = major

	// Parse minor version (optional)
	if len(matches) > 2 && matches[2] != "" {
		minor, err := strconv.Atoi(matches[2])
		if err != nil {
			return result, fmt.Errorf("invalid minor version: %s", matches[2])
		}
		result.Minor = minor
	}

	// Parse patch version (optional)
	if len(matches) > 3 && matches[3] != "" {
		patch, err := strconv.Atoi(matches[3])
		if err != nil {
			return result, fmt.Errorf("invalid patch version: %s", matches[3])
		}
		result.Patch = patch
	}

	return result, nil
}

// IsCompatible checks if two API versions are compatible
// Compatibility rules:
// - Major versions must match exactly
// - The client minor version must be <= server minor version
func (v APIVersion) IsCompatible(other APIVersion) bool {
	// Different services are never compatible
	if v.Service != other.Service {
		return false
	}

	// Beta versions are only compatible with themselves
	if v.Beta || other.Beta {
		return v.Beta == other.Beta
	}

	// Major versions must match exactly
	if v.Major != other.Major {
		return false
	}

	// Minor version compatibility (client minor <= server minor)
	return v.Minor <= other.Minor
}

// VersionCheck performs compatibility checking against supported API versions
type VersionCheck struct {
	Enabled           bool                   // Whether version checking is enabled
	SupportedVersions map[string]APIVersion // Map of supported API versions
	CustomVersions    map[string]APIVersion // Map of custom API versions
}

// NewVersionCheck creates a new version checker with default settings
func NewVersionCheck() *VersionCheck {
	return &VersionCheck{
		Enabled:           true,
		SupportedVersions: SupportedAPIVersions,
		CustomVersions:    make(map[string]APIVersion),
	}
}

// CheckServiceVersion checks if a service version is compatible with the SDK
func (vc *VersionCheck) CheckServiceVersion(service string, version string) error {
	if !vc.Enabled {
		return nil
	}

	// Parse the provided version
	serverVersion, err := ParseAPIVersion(service, version)
	if err != nil {
		return err
	}

	// Check custom versions first
	if customVersion, ok := vc.CustomVersions[service]; ok {
		if !customVersion.IsCompatible(serverVersion) {
			return fmt.Errorf("incompatible %s API version: SDK supports %s, server is %s",
				service, customVersion, serverVersion)
		}
		return nil
	}

	// Check against supported versions
	if supportedVersion, ok := vc.SupportedVersions[service]; ok {
		if !supportedVersion.IsCompatible(serverVersion) {
			return fmt.Errorf("incompatible %s API version: SDK supports %s, server is %s",
				service, supportedVersion, serverVersion)
		}
		return nil
	}

	return fmt.Errorf("unsupported service: %s", service)
}

// SetCustomVersion sets a custom API version for a specific service
func (vc *VersionCheck) SetCustomVersion(service string, version string) error {
	customVersion, err := ParseAPIVersion(service, version)
	if err != nil {
		return err
	}

	vc.CustomVersions[service] = customVersion
	return nil
}

// DisableVersionCheck disables API version checking
func (vc *VersionCheck) DisableVersionCheck() {
	vc.Enabled = false
}

// EnableVersionCheck enables API version checking
func (vc *VersionCheck) EnableVersionCheck() {
	vc.Enabled = true
}