#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
#
# This script runs the package export verification tool and fails if any errors are found.
# Can be integrated into CI workflows to catch missing exports early.

# Allow the script to be sourced without executing the main function
[[ "${BASH_SOURCE[0]}" != "${0}" ]] && SOURCED=1 || SOURCED=0

# Main function to run the verification
function verify_exports() {
  echo "Verifying SDK package exports..."
  cd "$(dirname "$0")/.." && go run ./cmd/verify-exports || {
    echo "ERROR: Package export verification failed"
    echo "This means that some required functions or types are not properly exported"
    echo "Fix the issues before releasing a new version to avoid breaking dependent projects"
    return 1
  }

  echo "All package exports verified successfully"
  return 0
}

# Only run if not being sourced
if [[ $SOURCED -eq 0 ]]; then
  set -e
  verify_exports || exit 1
fi