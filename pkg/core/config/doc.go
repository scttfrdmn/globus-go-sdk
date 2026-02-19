// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package config provides SDK-wide configuration management for the Globus Go SDK.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Config struct and all exported fields (HTTPClient, BaseURL, UserAgent, LogLevel,
    Timeout, RetryMax, RetryWaitMin, RetryWaitMax, VersionCheck, Debug, Trace, CustomTransport)
  - DefaultConfig constructor function
  - FromEnvironment constructor function (reads GLOBUS_SDK_BASE_URL, GLOBUS_SDK_USER_AGENT)
  - Config.ApplyToClient method
  - Config.GetVersionCheck and Config.SetVersionCheck methods
  - Environment variable support (GLOBUS_DISABLE_CONNECTION_POOL)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - New fields added to Config will have sensible zero values or defaults
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create a default configuration:

	cfg := config.DefaultConfig()

Load configuration from environment variables:

	cfg := config.FromEnvironment()

Customize the configuration:

	cfg := config.DefaultConfig()
	cfg.BaseURL = "https://auth.globus.org/"
	cfg.Timeout = 60 * time.Second
	cfg.Debug = true

Apply configuration to a client:

	client := core.NewClient("https://auth.globus.org/")
	cfg.ApplyToClient(client)
*/
package config
