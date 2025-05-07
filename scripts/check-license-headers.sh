#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# This script checks that all source files have proper SPDX license headers

set -e  # Exit on error

# Print colorized output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Checking for SPDX license headers...${NC}"

MISSING_HEADERS=0

# Check Go files
echo -e "${BLUE}Checking Go files...${NC}"
for file in $(find . -name "*.go" -not -path "./vendor/*" -not -path "./pkg/gen/*"); do
  if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${RED}❌ Missing SPDX header: ${file}${NC}"
    MISSING_HEADERS=$((MISSING_HEADERS+1))
  fi
done

# Check shell scripts
echo -e "${BLUE}Checking shell scripts...${NC}"
for file in $(find . -name "*.sh"); do
  if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${RED}❌ Missing SPDX header: ${file}${NC}"
    MISSING_HEADERS=$((MISSING_HEADERS+1))
  fi
done

# Check Markdown files (optional, uncomment if needed)
echo -e "${BLUE}Checking Markdown files...${NC}"
for file in $(find . -name "*.md" -not -path "./vendor/*" -not -path "./LICENSE.md"); do
  if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${YELLOW}⚠ Missing SPDX header: ${file}${NC}"
    # Not counting this as an error, just a warning
  fi
done

# Display summary
if [ $MISSING_HEADERS -eq 0 ]; then
  echo -e "${GREEN}✓ All files have proper SPDX license headers${NC}"
  exit 0
else
  echo -e "${RED}❌ Found ${MISSING_HEADERS} files missing SPDX license headers${NC}"
  echo -e "${YELLOW}Run './scripts/standardize-spdx-headers.sh' to fix${NC}"
  exit 1
fi