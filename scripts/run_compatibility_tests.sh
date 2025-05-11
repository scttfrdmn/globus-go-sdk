#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
#
# This script runs comprehensive compatibility tests against previous versions

set -e

# Default values
COMPARE_VERSION=""
OUTPUT_DIR="./compatibility-test-results"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      COMPARE_VERSION="$2"
      shift
      shift
      ;;
    --output-dir)
      OUTPUT_DIR="$2"
      shift
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Get the latest release tag if not specified
if [[ -z "$COMPARE_VERSION" ]]; then
  COMPARE_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  echo "Using latest release tag: $COMPARE_VERSION"
fi

echo "Running compatibility tests comparing with version: $COMPARE_VERSION"

# Save current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Run compatibility tests against current version
echo "Running compatibility tests against current version..."
go test -v ./tests/compatibility/... | tee "$OUTPUT_DIR/current-compatibility.log"

# Checkout the specified version
echo "Checking out version $COMPARE_VERSION..."
git checkout "$COMPARE_VERSION"

# Run compatibility tests against the specified version
echo "Running compatibility tests against version $COMPARE_VERSION..."
VERSION="$COMPARE_VERSION" go test -v ./tests/compatibility/... | tee "$OUTPUT_DIR/$COMPARE_VERSION-compatibility.log" || echo "Tests failed for version $COMPARE_VERSION"

# Return to the original branch
git checkout "$CURRENT_BRANCH"

# Create a comparison report
echo "Generating compatibility report..."
cat > "$OUTPUT_DIR/compatibility-report.md" << EOF
# Compatibility Test Report

## Overview

This report compares the compatibility of the Globus Go SDK between:
- Current version ($(git rev-parse --short HEAD))
- Version $COMPARE_VERSION

## Test Results

### Current Version

\`\`\`
$(grep -E '^(--- |PASS|FAIL|ok |SKIP)' "$OUTPUT_DIR/current-compatibility.log")
\`\`\`

### Version $COMPARE_VERSION

\`\`\`
$(grep -E '^(--- |PASS|FAIL|ok |SKIP)' "$OUTPUT_DIR/$COMPARE_VERSION-compatibility.log")
\`\`\`

## Summary

$(grep -c 'PASS' "$OUTPUT_DIR/current-compatibility.log") tests passed in current version
$(grep -c 'PASS' "$OUTPUT_DIR/$COMPARE_VERSION-compatibility.log") tests passed in version $COMPARE_VERSION

$(grep -c 'FAIL' "$OUTPUT_DIR/current-compatibility.log") tests failed in current version
$(grep -c 'FAIL' "$OUTPUT_DIR/$COMPARE_VERSION-compatibility.log") tests failed in version $COMPARE_VERSION
EOF

echo "Compatibility report generated: $OUTPUT_DIR/compatibility-report.md"