// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"

// Ensure Transport implements the interfaces.Transport interface
var _ interfaces.Transport = (*Transport)(nil)
