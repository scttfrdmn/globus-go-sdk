#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Script to summarize test results for the Globus Go SDK

set -e

echo "# Globus Go SDK Test Summary"
echo "Generated: $(date)"
echo

# Run unit tests for each package
echo "## Unit Tests"
echo

packages=(
  "./pkg/core"
  "./pkg/core/auth"
  "./pkg/core/authorizers" 
  "./pkg/core/config"
  "./pkg/core/http"
  "./pkg/core/logging"
  "./pkg/core/ratelimit"
  "./pkg/core/transport"
  "./pkg/services/auth"
  "./pkg/services/compute"
  "./pkg/services/flows"
  "./pkg/services/groups"
  "./pkg/services/search"
  "./pkg/services/timers"
  "./pkg/services/transfer"
)

for pkg in "${packages[@]}"; do
  pkg_name=$(echo "$pkg" | sed 's/\.\///g')
  echo "### $pkg_name"
  
  if ! go test "$pkg" -v 2>&1 | grep -E "PASS|FAIL|SKIP"; then
    echo "Error running tests"
  fi
  echo
done

# Run integration tests for service packages
echo "## Integration Tests"
echo

service_packages=(
  "./pkg/services/auth"
  "./pkg/services/compute"
  "./pkg/services/flows"
  "./pkg/services/groups"
  "./pkg/services/search"
  "./pkg/services/timers"
  "./pkg/services/transfer"
)

for pkg in "${service_packages[@]}"; do
  pkg_name=$(echo "$pkg" | sed 's/\.\///g')
  echo "### $pkg_name"
  
  # Run with integration tag
  echo "```"
  result=$(go test -tags=integration "$pkg" -v 2>&1)
  echo "$result" | grep -E "PASS:|FAIL:|SKIP:" || true
  
  # Count the number of passed, failed, and skipped tests
  passed_count=$(echo "$result" | grep -c "PASS:" || true)
  failed_count=$(echo "$result" | grep -c "FAIL:" || true)
  skipped_count=$(echo "$result" | grep -c "SKIP:" || true)
  
  # Extract skipped test names
  if [ "$skipped_count" -gt 0 ]; then
    echo
    echo "Skipped tests:"
    echo "$result" | grep "SKIP:" | sed 's/--- SKIP: /- /g' | cut -d' ' -f1-2
  fi
  
  if [ "$failed_count" -gt 0 ]; then
    echo
    echo "Failed tests:"
    echo "$result" | grep "FAIL:" | sed 's/--- FAIL: /- /g' | cut -d' ' -f1-2
  fi
  echo "```"
  
  echo "**Summary:** $passed_count passed, $failed_count failed, $skipped_count skipped"
  echo
done

# Generate overall status table
echo "## Overall Status"
echo
echo "| Package | Unit Tests | Integration Tests |"
echo "|---------|------------|-------------------|"

for pkg in "${service_packages[@]}"; do
  pkg_name=$(echo "$pkg" | sed 's/\.\///g' | sed 's/pkg\/services\///g')
  
  # Check unit tests
  unit_result=$(go test "$pkg" 2>&1)
  if echo "$unit_result" | grep -q "PASS"; then
    unit_status="✅ Pass"
  else
    unit_status="❌ Fail"
  fi
  
  # Check integration tests
  int_result=$(go test -tags=integration "$pkg" 2>&1)
  if echo "$int_result" | grep -q "PASS"; then
    int_status="✅ Pass"
  elif echo "$int_result" | grep -q "SKIP" && ! echo "$int_result" | grep -q "FAIL:"; then
    int_status="⚠️ Partial"
  else
    int_status="❌ Fail"
  fi
  
  echo "| $pkg_name | $unit_status | $int_status |"
done