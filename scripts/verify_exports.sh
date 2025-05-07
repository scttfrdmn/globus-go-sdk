#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
#
# This script runs the package export verification tool and fails if any errors are found.
# Can be integrated into CI workflows to catch missing exports early.

set -e

echo "Verifying SDK package exports..."
go run scripts/verify_package_exports.go || {
  echo "ERROR: Package export verification failed"
  echo "This means that some required functions or types are not properly exported"
  echo "Fix the issues before releasing a new version to avoid breaking dependent projects"
  exit 1
}

echo "All package exports verified successfully"