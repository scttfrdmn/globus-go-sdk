// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package contracts provides contract tests for the core interfaces of the Globus Go SDK.
//
// Contract tests verify that implementations of interfaces adhere to the
// behavioral expectations (the "contract") of those interfaces. This goes beyond
// simple type checking to ensure that implementations behave correctly.
//
// Using this package helps ensure that alternative implementations of core interfaces
// maintain compatibility with the rest of the SDK. It also serves as a form of
// executable documentation for the expected behavior of each interface.
//
// # Usage
//
// Each interface has a corresponding VerifyXXXContract function that takes an
// implementation of that interface and runs a series of tests to verify that it
// behaves correctly. These functions can be called from your own tests.
//
// Example:
//
//	func TestMyClientImplementation(t *testing.T) {
//	    client := NewMyClient()
//	    contracts.VerifyClientContract(t, client)
//	}
//
// # Stability
//
// This package is marked as BETA. The API may change in minor versions but not in
// patch versions.
//
// Contract definitions will be expanded and refined over time, but existing behavioral
// expectations will not be changed in backward-incompatible ways without a major
// version increment.
package contracts
