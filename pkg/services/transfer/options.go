// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// ClientConfig holds the configuration for the transfer client
type ClientConfig struct {
	authorizer  auth.Authorizer
	debug       bool
	trace       bool
	logger      interfaces.Logger
	coreOptions []core.ClientOption
}

// Option defines a configuration option for the Transfer client
type Option func(*ClientConfig)

// WithAuthorizer sets the authorizer for the client
func WithAuthorizer(authorizer auth.Authorizer) Option {
	return func(cfg *ClientConfig) {
		cfg.authorizer = authorizer
	}
}

// WithHTTPDebugging enables HTTP request/response logging
func WithHTTPDebugging(enable bool) Option {
	return func(cfg *ClientConfig) {
		cfg.debug = enable
	}
}

// WithHTTPTracing enables detailed HTTP tracing including headers and bodies
func WithHTTPTracing(enable bool) Option {
	return func(cfg *ClientConfig) {
		cfg.trace = enable
	}
}

// WithLogger sets the logger for the client
func WithLogger(logger interfaces.Logger) Option {
	return func(cfg *ClientConfig) {
		cfg.logger = logger
	}
}

// WithCoreOption appends a core client option
func WithCoreOption(option core.ClientOption) Option {
	return func(cfg *ClientConfig) {
		if cfg.coreOptions == nil {
			cfg.coreOptions = []core.ClientOption{option}
		} else {
			cfg.coreOptions = append(cfg.coreOptions, option)
		}
	}
}
