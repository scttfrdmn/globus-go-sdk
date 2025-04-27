// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package config

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// WithAPIVersionCheck configures whether API version checking is enabled
func (c *Config) WithAPIVersionCheck(enabled bool) *Config {
	if c.VersionCheck == nil {
		c.VersionCheck = core.NewVersionCheck()
	}
	
	if enabled {
		c.VersionCheck.EnableVersionCheck()
	} else {
		c.VersionCheck.DisableVersionCheck()
	}
	
	return c
}

// WithCustomAPIVersion sets a custom API version for a specific service
func (c *Config) WithCustomAPIVersion(service, version string) *Config {
	if c.VersionCheck == nil {
		c.VersionCheck = core.NewVersionCheck()
	}
	
	// Ignore errors here, they will be caught when the service is used
	c.VersionCheck.SetCustomVersion(service, version)
	
	return c
}