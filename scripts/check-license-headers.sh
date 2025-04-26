#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Check for license headers in Go files

set -e

LICENSE_HEADER="// SPDX-License-Identifier: Apache-2.0"
COPYRIGHT="// Copyright (c) 2025 Scott Friedman and Project Contributors"

# Find all Go files in the project
files=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*")

# Check each file for license headers
missing_license=0
missing_copyright=0

for file in $files; do
  if ! grep -q "$LICENSE_HEADER" "$file"; then
    echo "❌ Missing license header in $file"
    missing_license=1
  fi
  
  if ! grep -q "$COPYRIGHT" "$file"; then
    echo "❌ Missing copyright notice in $file"
    missing_copyright=1
  fi
done

# Report results
if [ $missing_license -eq 0 ] && [ $missing_copyright -eq 0 ]; then
  echo "✅ All Go files have correct license headers and copyright notices"
  exit 0
else
  echo "❌ Some files are missing license headers or copyright notices"
  echo "Please add the following lines at the top of the file:"
  echo "$LICENSE_HEADER"
  echo "$COPYRIGHT"
  echo ""
  exit 1
fi