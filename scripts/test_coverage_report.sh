#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to generate a detailed test coverage report for the Globus Go SDK
# This script runs tests, generates coverage data, and produces a markdown
# report showing coverage by package.

set -e  # Exit on error

# Print colorized output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create temp directory for coverage files
TMP_DIR=$(mktemp -d)
echo -e "${BLUE}Created temporary directory: ${TMP_DIR}${NC}"

# Function to clean up
cleanup() {
  echo -e "${BLUE}Cleaning up temporary files...${NC}"
  rm -rf "$TMP_DIR"
}

# Register cleanup function to run on exit
trap cleanup EXIT

echo -e "${BLUE}=========================================================${NC}"
echo -e "${BLUE}        Globus Go SDK Test Coverage Reporter${NC}"
echo -e "${BLUE}=========================================================${NC}"

echo -e "${BLUE}Running tests and collecting coverage data...${NC}"

# Run tests with coverage
go test -covermode=atomic -coverprofile="$TMP_DIR/coverage.out" ./...

# Create combined coverage profile
COVERAGE_FILE="$TMP_DIR/coverage.out"

# If tests fail, still try to generate report from whatever coverage data we have
echo -e "${BLUE}Generating coverage report...${NC}"

# Get total coverage
TOTAL_COV=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
echo -e "${BLUE}Total code coverage: ${GREEN}${TOTAL_COV}${NC}"

# Generate HTML report
REPORT_HTML="coverage.html"
go tool cover -html="$COVERAGE_FILE" -o "$REPORT_HTML"
echo -e "${BLUE}HTML coverage report generated: ${GREEN}${REPORT_HTML}${NC}"

# Function to calculate color based on coverage percentage
get_color() {
  local cov=$1
  cov=${cov%\%}  # Remove % sign
  
  if (( $(echo "$cov >= 80" | bc -l) )); then
    echo "${GREEN}"
  elif (( $(echo "$cov >= 50" | bc -l) )); then
    echo "${YELLOW}"
  else
    echo "${RED}"
  fi
}

# Generate detailed package coverage report in markdown format
REPORT_MD="coverage_report.md"

echo "# Globus Go SDK Test Coverage Report" > "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "Generated on: $(date)" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "## Overall Coverage: $TOTAL_COV" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "## Package Coverage" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "| Package | Coverage | Status |" >> "$REPORT_MD"
echo "|---------|----------|--------|" >> "$REPORT_MD"

echo -e "${BLUE}Package coverage details:${NC}"
echo -e "${BLUE}-----------------------${NC}"

# Generate package coverage details
go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | sort | while read -r line; do
  PKG=$(echo "$line" | awk '{print $1}' | sed 's/github.com\/scttfrdmn\/globus-go-sdk\///')
  FUNC=$(echo "$line" | awk '{print $2}')
  COV=$(echo "$line" | awk '{print $3}')
  
  # Skip individual functions, only process package totals
  if [[ "$FUNC" != "total:" ]]; then
    continue
  fi
  
  # Get appropriate color
  COLOR=$(get_color "$COV")
  
  # Print package coverage to terminal
  echo -e "${COLOR}${PKG}: ${COV}${NC}"
  
  # Determine status based on coverage
  STATUS=""
  if (( $(echo "${COV%\%} >= 80" | bc -l) )); then
    STATUS="✅ Complete"
  elif (( $(echo "${COV%\%} >= 50" | bc -l) )); then
    STATUS="⚠️ Partial"
  else
    STATUS="❌ Insufficient"
  fi
  
  # Add to markdown report
  echo "| $PKG | $COV | $STATUS |" >> "$REPORT_MD"
done

echo "" >> "$REPORT_MD"
echo "## Coverage Threshold Goals" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "- ✅ **Complete**: >= 80% coverage" >> "$REPORT_MD"
echo "- ⚠️ **Partial**: >= 50% coverage" >> "$REPORT_MD"
echo "- ❌ **Insufficient**: < 50% coverage" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"
echo "## Recommendations" >> "$REPORT_MD"
echo "" >> "$REPORT_MD"

if (( $(echo "${TOTAL_COV%\%} >= 80" | bc -l) )); then
  echo "- Coverage goal of 80% has been achieved! 🎉" >> "$REPORT_MD"
  echo "- Continue to maintain this level of coverage for new code" >> "$REPORT_MD"
else
  echo "- Focus on packages with insufficient coverage first" >> "$REPORT_MD"
  echo "- Add tests for critical components even if their package coverage is already good" >> "$REPORT_MD"
  echo "- Prioritize adding tests for error handling and edge cases" >> "$REPORT_MD"
fi

echo -e "${BLUE}Markdown coverage report generated: ${GREEN}${REPORT_MD}${NC}"
echo -e "${BLUE}Done!${NC}"

# Return to the caller if we're being sourced, exit with success otherwise
if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
  return 0
else
  exit 0
fi