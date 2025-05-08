// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package config

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// ConfigAccessor defines the interface for config objects that provide access to their fields
// This interface ensures that all required getter and setter methods are available
type ConfigAccessor interface {
	// GetVersionCheck returns the version check manager
	GetVersionCheck() *core.VersionCheck

	// SetVersionCheck sets the version check manager
	SetVersionCheck(vc *core.VersionCheck)

	// WithAPIVersionCheck enables or disables API version checking
	WithAPIVersionCheck(enabled bool) *Config

	// WithCustomAPIVersion sets a custom API version for a service
	WithCustomAPIVersion(service, version string) *Config

	// ApplyToClient applies the configuration to a client
	ApplyToClient(client *core.Client)
}

// Verify that Config implements ConfigAccessor at compile time
var _ ConfigAccessor = &Config{}
