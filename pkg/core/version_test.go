// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"testing"
)

func TestAPIVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  APIVersion
		expected string
	}{
		{
			name: "major and minor version",
			version: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			expected: "auth/v2.0",
		},
		{
			name: "with patch version",
			version: APIVersion{
				Service: "transfer",
				Major:   0,
				Minor:   10,
				Patch:   5,
			},
			expected: "transfer/v0.10.5",
		},
		{
			name: "beta version",
			version: APIVersion{
				Service: "flows",
				Beta:    true,
			},
			expected: "flows/Beta",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.version.String()
			if result != tc.expected {
				t.Errorf("String() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestAPIVersionEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		version  APIVersion
		expected string
	}{
		{
			name: "auth endpoint",
			version: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			expected: "https://auth.globus.org/v2/",
		},
		{
			name: "transfer endpoint",
			version: APIVersion{
				Service: "transfer",
				Major:   0,
				Minor:   10,
			},
			expected: "https://transfer.api.globus.org/v0.10/",
		},
		{
			name: "search endpoint",
			version: APIVersion{
				Service: "search",
				Major:   1,
				Minor:   0,
			},
			expected: "https://search.api.globus.org/v1/",
		},
		{
			name: "groups endpoint",
			version: APIVersion{
				Service: "groups",
				Major:   2,
				Minor:   0,
			},
			expected: "https://groups.api.globus.org/v2/",
		},
		{
			name: "flows endpoint",
			version: APIVersion{
				Service: "flows",
				Beta:    true,
			},
			expected: "https://flows.globus.org/api/",
		},
		{
			name: "unknown service",
			version: APIVersion{
				Service: "unknown",
				Major:   1,
				Minor:   0,
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.version.Endpoint()
			if result != tc.expected {
				t.Errorf("Endpoint() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestParseAPIVersion(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		version     string
		expected    APIVersion
		expectError bool
	}{
		{
			name:    "major only",
			service: "auth",
			version: "v2",
			expected: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
				Patch:   0,
			},
		},
		{
			name:    "major and minor",
			service: "transfer",
			version: "v0.10",
			expected: APIVersion{
				Service: "transfer",
				Major:   0,
				Minor:   10,
				Patch:   0,
			},
		},
		{
			name:    "major, minor, and patch",
			service: "search",
			version: "v1.0.5",
			expected: APIVersion{
				Service: "search",
				Major:   1,
				Minor:   0,
				Patch:   5,
			},
		},
		{
			name:    "beta version",
			service: "flows",
			version: "beta",
			expected: APIVersion{
				Service: "flows",
				Beta:    true,
			},
		},
		{
			name:        "invalid format",
			service:     "auth",
			version:     "version2",
			expectError: true,
		},
		{
			name:        "invalid major",
			service:     "auth",
			version:     "vX",
			expectError: true,
		},
		{
			name:        "invalid minor",
			service:     "transfer",
			version:     "v0.X",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseAPIVersion(tc.service, tc.version)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result.Service != tc.expected.Service {
				t.Errorf("Service = %q, want %q", result.Service, tc.expected.Service)
			}
			
			if result.Major != tc.expected.Major {
				t.Errorf("Major = %d, want %d", result.Major, tc.expected.Major)
			}
			
			if result.Minor != tc.expected.Minor {
				t.Errorf("Minor = %d, want %d", result.Minor, tc.expected.Minor)
			}
			
			if result.Patch != tc.expected.Patch {
				t.Errorf("Patch = %d, want %d", result.Patch, tc.expected.Patch)
			}
			
			if result.Beta != tc.expected.Beta {
				t.Errorf("Beta = %v, want %v", result.Beta, tc.expected.Beta)
			}
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name      string
		version1  APIVersion
		version2  APIVersion
		compatible bool
	}{
		{
			name: "exact match",
			version1: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			version2: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			compatible: true,
		},
		{
			name: "server has higher minor",
			version1: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			version2: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   1,
			},
			compatible: true,
		},
		{
			name: "client has higher minor",
			version1: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   1,
			},
			version2: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			compatible: false,
		},
		{
			name: "different major versions",
			version1: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			version2: APIVersion{
				Service: "auth",
				Major:   3,
				Minor:   0,
			},
			compatible: false,
		},
		{
			name: "different services",
			version1: APIVersion{
				Service: "auth",
				Major:   2,
				Minor:   0,
			},
			version2: APIVersion{
				Service: "transfer",
				Major:   2,
				Minor:   0,
			},
			compatible: false,
		},
		{
			name: "beta versions match",
			version1: APIVersion{
				Service: "flows",
				Beta:    true,
			},
			version2: APIVersion{
				Service: "flows",
				Beta:    true,
			},
			compatible: true,
		},
		{
			name: "beta vs non-beta",
			version1: APIVersion{
				Service: "flows",
				Beta:    true,
			},
			version2: APIVersion{
				Service: "flows",
				Major:   1,
				Minor:   0,
			},
			compatible: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.version1.IsCompatible(tc.version2)
			if result != tc.compatible {
				t.Errorf("IsCompatible() = %v, want %v", result, tc.compatible)
			}
		})
	}
}

func TestVersionCheck(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		version     string
		customVersions map[string]string
		enabled     bool
		expectError bool
	}{
		{
			name:    "compatible version",
			service: "auth",
			version: "v2",
			enabled: true,
			expectError: false,
		},
		{
			name:    "incompatible major version",
			service: "auth",
			version: "v3",
			enabled: true,
			expectError: true,
		},
		{
			name:    "compatible minor version",
			service: "transfer",
			version: "v0.11",
			enabled: true,
			expectError: false,
		},
		{
			name:    "incompatible minor version",
			service: "transfer",
			version: "v0.9",
			enabled: true,
			expectError: true,
		},
		{
			name:    "unsupported service",
			service: "unknown",
			version: "v1",
			enabled: true,
			expectError: true,
		},
		{
			name:    "version checking disabled",
			service: "auth",
			version: "v3",
			enabled: false,
			expectError: false,
		},
		{
			name:    "custom version compatible",
			service: "auth",
			version: "v3",
			customVersions: map[string]string{
				"auth": "v3",
			},
			enabled: true,
			expectError: false,
		},
		{
			name:    "custom version incompatible",
			service: "auth",
			version: "v3",
			customVersions: map[string]string{
				"auth": "v2",
			},
			enabled: true,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vc := NewVersionCheck()
			vc.Enabled = tc.enabled
			
			// Set any custom versions
			if tc.customVersions != nil {
				for service, version := range tc.customVersions {
					err := vc.SetCustomVersion(service, version)
					if err != nil {
						t.Fatalf("Failed to set custom version: %v", err)
					}
				}
			}
			
			err := vc.CheckServiceVersion(tc.service, tc.version)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}