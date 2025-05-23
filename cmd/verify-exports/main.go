// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"os"

	"github.com/scttfrdmn/globus-go-sdk/internal/verification"
)

func main() {
	if verification.VerifyPackageExports() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
