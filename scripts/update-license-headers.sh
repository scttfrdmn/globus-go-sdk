#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# This script is a wrapper that checks and updates license headers as needed

set -e  # Exit on error

# Print colorized output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Checking license headers...${NC}"

# First run the check script to see if there are any issues
if ./scripts/check-license-headers.sh > /dev/null 2>&1; then
  echo -e "${GREEN}✓ All files already have proper SPDX license headers${NC}"
  exit 0
else
  echo -e "${YELLOW}⚠ Some files are missing SPDX license headers${NC}"
  echo -e "${BLUE}Running standardize-spdx-headers.sh to fix issues...${NC}"
  
  # Run the standardization script
  ./scripts/standardize-spdx-headers.sh
  
  # Verify that all headers are now fixed
  if ./scripts/check-license-headers.sh > /dev/null 2>&1; then
    echo -e "${GREEN}✓ All license headers have been updated successfully${NC}"
    exit 0
  else
    echo -e "${RED}❌ There are still issues with license headers${NC}"
    echo -e "${YELLOW}Please check the output of check-license-headers.sh for details${NC}"
    ./scripts/check-license-headers.sh
    exit 1
  fi
fi