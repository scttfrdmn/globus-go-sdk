#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to lint all shell scripts in the project using shellcheck

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check if shellcheck is installed
if ! command -v shellcheck &> /dev/null; then
    echo -e "${RED}Error: shellcheck is not installed.${NC}"
    echo "Please install shellcheck:"
    echo "  macOS:   brew install shellcheck"
    echo "  Ubuntu:  apt-get install shellcheck"
    echo "  Website: https://github.com/koalaman/shellcheck#installing"
    exit 1
fi

# Print shellcheck version
echo -e "${YELLOW}Using ShellCheck $(shellcheck --version | grep 'version:' | awk '{print $2}')${NC}"

# Find all shell scripts
echo -e "\n${YELLOW}Looking for shell scripts...${NC}"
SHELL_SCRIPTS=$(find . -type f -name "*.sh" -not -path "*/\.*" | sort)

if [ -z "$SHELL_SCRIPTS" ]; then
    echo -e "${RED}No shell scripts found.${NC}"
    exit 0
fi

echo -e "Found $(echo "$SHELL_SCRIPTS" | wc -l | tr -d '[:space:]') shell scripts."

# Check if specific files were provided
if [ $# -gt 0 ]; then
    SHELL_SCRIPTS="$@"
    echo -e "${YELLOW}Linting only the specified files: ${SHELL_SCRIPTS}${NC}"
fi

# Lint each script
FAILED=0
SUCCESS=0

echo -e "\n${YELLOW}Linting shell scripts...${NC}"
for script in $SHELL_SCRIPTS; do
    echo -n "Checking $script... "
    
    # Run shellcheck
    if shellcheck -x "$script"; then
        echo -e "${GREEN}PASSED${NC}"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "${RED}FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi
done

# Print summary
echo -e "\n${YELLOW}Summary:${NC}"
echo -e "${GREEN}$SUCCESS scripts passed${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}$FAILED scripts failed${NC}"
    exit 1
else
    echo -e "${GREEN}All shell scripts passed linting!${NC}"
    exit 0
fi