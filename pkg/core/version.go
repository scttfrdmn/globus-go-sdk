// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package core

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

// Version is the current version of the Globus Go SDK
const Version = "0.9.7"

// APIVersion represents a Globus API version
type APIVersion struct {
	// Service is the name of the service (e.g., "transfer", "auth")
	Service string

	// Major is the major version number
	Major int

	// Minor is the minor version number
	Minor int

	// Patch is the patch version number (optional)
	Patch int

	// Beta indicates whether this is a beta version
	Beta bool
}

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

// IsDevelopment returns true if the version is a development version
func IsDevelopment() bool {
	return strings.Contains(Version, "-rc") || strings.Contains(Version, "-beta")
}

// UserAgent returns the appropriate User-Agent header string for the SDK
func UserAgent() string {
	info := GetInfo()
	return fmt.Sprintf("Globus-Go-SDK/%s", info.Version)
}

// ParseVersion parses a version string into an APIVersion
// Supported formats:
// - v1
// - v1.2
// - v1.2.3
// - v1.2-beta
func ParseVersion(service, version string) (*APIVersion, error) {
	// Handle empty version
	if version == "" {
		return nil, fmt.Errorf("empty version string")
	}

	// Strip leading 'v' if present
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	// Check for beta flag
	beta := false
	if strings.Contains(version, "-beta") {
		beta = true
		version = strings.Replace(version, "-beta", "", 1)
	}

	// Split version into components
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	// Parse major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse minor version if present
	minor := 0
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	// Parse patch version if present
	patch := 0
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return &APIVersion{
		Service: service,
		Major:   major,
		Minor:   minor,
		Patch:   patch,
		Beta:    beta,
	}, nil
}

// ParseAPIVersion is an alias for ParseVersion for backward compatibility
func ParseAPIVersion(service, version string) (*APIVersion, error) {
	return ParseVersion(service, version)
}

// String returns the string representation of the version
func (v *APIVersion) String() string {
	if v.Patch > 0 {
		if v.Beta {
			return fmt.Sprintf("v%d.%d.%d-beta", v.Major, v.Minor, v.Patch)
		}
		return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	if v.Minor > 0 {
		if v.Beta {
			return fmt.Sprintf("v%d.%d-beta", v.Major, v.Minor)
		}
		return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	}

	if v.Beta {
		return fmt.Sprintf("v%d-beta", v.Major)
	}
	return fmt.Sprintf("v%d", v.Major)
}

// Compare compares this version to another version
// Returns:
//
//	-1 if this version is less than the other version
//	 0 if this version is equal to the other version
//	 1 if this version is greater than the other version
func (v *APIVersion) Compare(other *APIVersion) int {
	// Compare major version
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare beta flag
	if v.Beta != other.Beta {
		if v.Beta {
			return -1 // Beta is "less than" stable
		}
		return 1
	}

	// Versions are equal
	return 0
}

// GetEndpoint returns the API endpoint for this service and version
func (v *APIVersion) GetEndpoint() string {
	switch v.Service {
	case "transfer":
		return fmt.Sprintf("https://transfer.api.globus.org/%s/", v.String())
	case "auth":
		return fmt.Sprintf("https://auth.globus.org/%s/", v.String())
	case "search":
		return fmt.Sprintf("https://search.api.globus.org/%s/", v.String())
	case "flows":
		return "https://flows.globus.org/api/"
	default:
		return ""
	}
}

// Endpoint is an alias for GetEndpoint for backward compatibility
func (v *APIVersion) Endpoint() string {
	return v.GetEndpoint()
}

// IsCompatible checks if this version is compatible with another version
// We consider versions compatible if they have the same major version
// For services that use semver, we also check the minor version
func (v *APIVersion) IsCompatible(other interface{}) bool {
	// Handle both pointer and value parameter
	var otherVersion *APIVersion
	switch o := other.(type) {
	case *APIVersion:
		otherVersion = o
	case APIVersion:
		otherVersion = &o
	default:
		return false
	}
	// Different major versions are never compatible
	if v.Major != otherVersion.Major {
		return false
	}

	// For services using semver (most services):
	// - v1.0 is compatible with v1.1 (minor version increments are backward compatible)
	// - v1.2 is compatible with v1.2.3 (patch version increments are backward compatible)
	// - v2.0 is not compatible with v1.0 (major version increments are not backward compatible)

	// Some services (like Auth) might have different conventions, but this is the default

	// Everything else is compatible
	return true
}

// ExtractVersionFromURL extracts a version string from a URL
// For example, "https://transfer.api.globus.org/v0.10/endpoint" -> "v0.10"
func ExtractVersionFromURL(url string) string {
	re := regexp.MustCompile(`/v([0-9]+(\.[0-9]+)*(-beta)?)(/|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return "v" + matches[1]
	}
	return ""
}

// VersionCheck manages API version checking
type VersionCheck struct {
	// enabled determines whether version checking is enabled
	enabled bool

	// customVersions maps service names to custom API versions
	customVersions map[string]string

	// checkedServices tracks services that have already been checked
	checkedServices map[string]bool

	// mu guards the maps
	mu sync.Mutex
}

// NewVersionCheck creates a new VersionCheck
func NewVersionCheck() *VersionCheck {
	return &VersionCheck{
		enabled:         true,
		customVersions:  make(map[string]string),
		checkedServices: make(map[string]bool),
	}
}

// EnableVersionCheck enables version checking
func (vc *VersionCheck) EnableVersionCheck() {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.enabled = true
}

// DisableVersionCheck disables version checking
func (vc *VersionCheck) DisableVersionCheck() {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.enabled = false
}

// IsEnabled returns whether version checking is enabled
func (vc *VersionCheck) IsEnabled() bool {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	return vc.enabled
}

// SetCustomVersion sets a custom API version for a service
func (vc *VersionCheck) SetCustomVersion(service, version string) error {
	// Validate the version string
	_, err := ParseVersion(service, version)
	if err != nil {
		return err
	}

	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.customVersions[service] = version
	return nil
}

// GetCustomVersion gets a custom API version for a service
func (vc *VersionCheck) GetCustomVersion(service string) (string, bool) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	version, ok := vc.customVersions[service]
	return version, ok
}

// MarkServiceChecked marks a service as checked
func (vc *VersionCheck) MarkServiceChecked(service string) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.checkedServices[service] = true
}

// IsServiceChecked checks if a service has been checked
func (vc *VersionCheck) IsServiceChecked(service string) bool {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	return vc.checkedServices[service]
}

// Enabled is a getter for vc.enabled
func (vc *VersionCheck) Enabled() bool {
	return vc.IsEnabled()
}

// CheckServiceVersion checks the version of a service
// This is a compatibility method that simply marks the service as checked
func (vc *VersionCheck) CheckServiceVersion(service string, version string) error {
	// Validate the version
	_, err := ParseVersion(service, version)
	if err != nil {
		return err
	}

	// Mark as checked
	vc.MarkServiceChecked(service)
	return nil
}
