#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# 
# Script to verify API compatibility between versions

set -e

# Default values
PREV_VERSION=""
CURRENT_VERSION=""
COMPATIBILITY_LEVEL="minor"
OUTPUT_DIR="./api-compatibility-results"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --prev-version)
      PREV_VERSION="$2"
      shift
      shift
      ;;
    --current-version)
      CURRENT_VERSION="$2"
      shift
      shift
      ;;
    --level)
      COMPATIBILITY_LEVEL="$2"
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

# Validate inputs
if [[ -z "$PREV_VERSION" ]]; then
  # Get the latest release tag if not specified
  PREV_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  echo "Using latest release tag: $PREV_VERSION"
fi

if [[ -z "$CURRENT_VERSION" ]]; then
  # Use the current commit if not specified
  CURRENT_VERSION=$(git rev-parse --short HEAD)
  echo "Using current commit: $CURRENT_VERSION"
fi

# Validate compatibility level
if [[ "$COMPATIBILITY_LEVEL" != "patch" && "$COMPATIBILITY_LEVEL" != "minor" && "$COMPATIBILITY_LEVEL" != "major" ]]; then
  echo "Invalid compatibility level: $COMPATIBILITY_LEVEL. Must be one of: patch, minor, major"
  exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Step 1: Generate API signatures for current version
echo "Generating API signatures for current version..."
go run ./cmd/apigen/main.go -dir ./pkg -version "$CURRENT_VERSION" -output "$OUTPUT_DIR/api-current.json"

# Step 2: Generate API signatures for previous version
echo "Generating API signatures for previous version..."

# Save current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Checkout previous version
git checkout "$PREV_VERSION"

# Try building apigen in the previous version
if [ -d "./cmd/apigen" ]; then
  go build -o "$OUTPUT_DIR/apigen-prev" ./cmd/apigen/main.go
  
  # Use the built tool
  "$OUTPUT_DIR/apigen-prev" -dir ./pkg -version "$PREV_VERSION" -output "$OUTPUT_DIR/api-prev.json"
else
  # Tool doesn't exist in the previous version, go back and use current version
  git checkout "$CURRENT_BRANCH"
  go run ./cmd/apigen/main.go -dir ./pkg -version "$PREV_VERSION" -output "$OUTPUT_DIR/api-prev.json"
  
  # Checkout previous version again for next steps
  git checkout "$PREV_VERSION"
fi

# Step 3: Generate deprecation report for previous version
echo "Generating deprecation report for previous version..."
if [ -d "./cmd/depreport" ]; then
  go build -o "$OUTPUT_DIR/depreport-prev" ./cmd/depreport/main.go
  "$OUTPUT_DIR/depreport-prev" -dir ./pkg -o "$OUTPUT_DIR/deprecated-prev.md" || echo "No deprecated features in previous version"
else
  echo "Deprecation reporting tool not found in previous version"
fi

# Go back to current branch
git checkout "$CURRENT_BRANCH"

# Step 4: Generate deprecation report for current version
echo "Generating deprecation report for current version..."
go run ./cmd/depreport/main.go -dir ./pkg -o "$OUTPUT_DIR/deprecated-current.md"

# Step 5: Compare API signatures
echo "Comparing API signatures with compatibility level: $COMPATIBILITY_LEVEL..."
go run ./cmd/apicompare/main.go \
  -old "$OUTPUT_DIR/api-prev.json" \
  -new "$OUTPUT_DIR/api-current.json" \
  -level "$COMPATIBILITY_LEVEL" \
  -output "$OUTPUT_DIR/api-comparison.json"

# Step 6: Generate comprehensive report
echo "Generating comprehensive report..."
cat > "$OUTPUT_DIR/compatibility-report.md" << EOF
# API Compatibility Report

## Overview

Comparing versions:
- Previous: $PREV_VERSION
- Current: $CURRENT_VERSION
- Compatibility level: $COMPATIBILITY_LEVEL

## API Changes

EOF

# Extract API changes from comparison
REMOVALS=$(jq '.removals | length' "$OUTPUT_DIR/api-comparison.json")
ADDITIONS=$(jq '.additions | length' "$OUTPUT_DIR/api-comparison.json")
CHANGES=$(jq '.changes | length' "$OUTPUT_DIR/api-comparison.json")
BREAKING_CHANGES=$(jq '.breaking_changes | length' "$OUTPUT_DIR/api-comparison.json")

# Add summary to report
cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF
### Summary of Changes

- New APIs added: $ADDITIONS
- APIs removed: $REMOVALS
- APIs changed: $CHANGES
- Breaking changes: $BREAKING_CHANGES

EOF

# Check if API is compatible
if [[ "$COMPATIBILITY_LEVEL" != "major" && "$BREAKING_CHANGES" -gt 0 ]]; then
  echo "COMPATIBILITY WARNING: Found $BREAKING_CHANGES breaking changes for $COMPATIBILITY_LEVEL compatibility level"
  
  cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF
⚠️ **COMPATIBILITY ALERT**: Found $BREAKING_CHANGES breaking changes that violate $COMPATIBILITY_LEVEL compatibility level!

### Breaking Changes

EOF

  # List breaking changes
  jq -r '.breaking_changes[] | "- " + .type + " " + .package + "." + .name + " (" + .change_type + ")"' "$OUTPUT_DIR/api-comparison.json" >> "$OUTPUT_DIR/compatibility-report.md"
  
  # Set exit code
  EXIT_CODE=1
else
  echo "API is compatible at $COMPATIBILITY_LEVEL level"
  cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF
✅ **COMPATIBILITY PASS**: API changes are compatible with $COMPATIBILITY_LEVEL level requirements.
EOF
  
  # Set exit code
  EXIT_CODE=0
fi

# Add details about API changes
cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF

## Detailed Changes

### New APIs

EOF

# List additions
jq -r '.additions[] | "- " + .type + " " + .package + "." + .name' "$OUTPUT_DIR/api-comparison.json" >> "$OUTPUT_DIR/compatibility-report.md" || echo "No additions" >> "$OUTPUT_DIR/compatibility-report.md"

cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF

### Removed APIs

EOF

# List removals
jq -r '.removals[] | "- " + .type + " " + .package + "." + .name' "$OUTPUT_DIR/api-comparison.json" >> "$OUTPUT_DIR/compatibility-report.md" || echo "No removals" >> "$OUTPUT_DIR/compatibility-report.md"

cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF

### Changed APIs

EOF

# List changes
jq -r '.changes[] | "- " + .type + " " + .package + "." + .name' "$OUTPUT_DIR/api-comparison.json" >> "$OUTPUT_DIR/compatibility-report.md" || echo "No changes" >> "$OUTPUT_DIR/compatibility-report.md"

cat >> "$OUTPUT_DIR/compatibility-report.md" << EOF

## Deprecation Analysis

EOF

# Add deprecation information to report
if [ -f "$OUTPUT_DIR/deprecated-current.md" ]; then
  cat "$OUTPUT_DIR/deprecated-current.md" >> "$OUTPUT_DIR/compatibility-report.md"
else
  echo "No deprecation information available." >> "$OUTPUT_DIR/compatibility-report.md"
fi

# Print report location
echo "Compatibility report generated: $OUTPUT_DIR/compatibility-report.md"

# Return appropriate exit code
exit $EXIT_CODE