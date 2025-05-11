// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package deprecation provides functionality for managing API deprecations.
//
// This package is designed to help track and communicate deprecated features
// within the Globus Go SDK. It includes tools for logging deprecation warnings,
// tracking when deprecated features are used, and providing migration guidance
// to users.
//
// # Stability
//
// This package is marked as BETA. The API may change in minor versions but not in
// patch versions.
//
// # Usage
//
// The primary function of this package is to provide a way to mark functions, methods,
// types, or fields as deprecated and to log appropriate warnings when they are used.
//
// Example:
//
//	func SomeDeprecatedFunction() {
//	    deprecation.LogWarning(
//	        "SomeDeprecatedFunction",
//	        "v1.0.0",
//	        "v2.0.0",
//	        "Use NewFunction instead.")
//	    // ... function implementation
//	}
//
// This will log a warning message when SomeDeprecatedFunction is called, indicating
// when it was deprecated, when it will be removed, and what to use instead.
package deprecation
