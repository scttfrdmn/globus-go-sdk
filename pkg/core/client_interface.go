// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"

// Ensure Client implements the interfaces.ClientInterface interface
var _ interfaces.ClientInterface = (*Client)(nil)
