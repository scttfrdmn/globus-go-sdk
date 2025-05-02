// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// ClientOption configures an Auth client
type ClientOption func(*clientOptions)

// clientOptions represents options for configuring an Auth client
type clientOptions struct {
	clientID      string
	clientSecret  string
	redirectURL   string
	coreOptions   []core.ClientOption
}

// WithClientID sets the client ID
func WithClientID(clientID string) ClientOption {
	return func(o *clientOptions) {
		o.clientID = clientID
	}
}

// WithClientSecret sets the client secret
func WithClientSecret(clientSecret string) ClientOption {
	return func(o *clientOptions) {
		o.clientSecret = clientSecret
	}
}

// WithRedirectURL sets the redirect URL
func WithRedirectURL(redirectURL string) ClientOption {
	return func(o *clientOptions) {
		o.redirectURL = redirectURL
	}
}

// WithAuthorizer sets the authorizer for the client
func WithAuthorizer(authorizer interfaces.Authorizer) ClientOption {
	return func(o *clientOptions) {
		// Create an adapter to bridge between the two authorizer interfaces
		adapter := NewAuthorizerAdapter(authorizer)
		o.coreOptions = append(o.coreOptions, core.WithAuthorizer(adapter))
	}
}

// WithCoreOption adds a core option
func WithCoreOption(option core.ClientOption) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, option)
	}
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(o *clientOptions) {
		o.coreOptions = append(o.coreOptions, core.WithBaseURL(baseURL))
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
		coreOptions: []core.ClientOption{
			core.WithBaseURL(DefaultBaseURL),
		},
	}
}