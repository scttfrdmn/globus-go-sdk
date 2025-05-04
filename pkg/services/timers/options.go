// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timers

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

// ClientOption configures a Timers client
type ClientOption func(*clientOptions)

// clientOptions represents options for configuring a Timers client
type clientOptions struct {
	accessToken  string
	baseURL      string
	coreOptions  []core.ClientOption
}

// WithAccessToken sets the access token for authorization
func WithAccessToken(accessToken string) ClientOption {
	return func(o *clientOptions) {
		o.accessToken = accessToken
	}
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(o *clientOptions) {
		o.baseURL = baseURL
		o.coreOptions = append(o.coreOptions, core.WithBaseURL(baseURL))
	}
}

// WithAuthorizer sets the authorizer for the client
func WithAuthorizer(authorizer auth.Authorizer) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, core.WithAuthorizer(authorizer))
	}
}

// WithCoreOption adds a core option
func WithCoreOption(option core.ClientOption) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, option)
	}
}

// WithHTTPDebugging enables HTTP debugging
func WithHTTPDebugging(enable bool) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, core.WithHTTPDebugging(enable))
	}
}

// WithHTTPTracing enables HTTP tracing
func WithHTTPTracing(enable bool) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, core.WithHTTPTracing(enable))
	}
}

// defaultOptions returns the default client options
func defaultOptions() *clientOptions {
	return &clientOptions{
		baseURL: DefaultBaseURL,
		coreOptions: []core.ClientOption{
			core.WithBaseURL(DefaultBaseURL),
		},
	}
}