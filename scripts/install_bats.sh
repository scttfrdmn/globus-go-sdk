#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to install BATS (Bash Automated Testing System) and its dependencies

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Directories
TESTS_DIR="$(pwd)/tests"
BATS_DIR="${TESTS_DIR}/bats"

# Create directories if they don't exist
mkdir -p "$TESTS_DIR"

# Check for git
if ! command -v git &> /dev/null; then
    echo -e "${RED}Error: git is not installed. Please install git to continue.${NC}"
    exit 1
fi

# Install bats-core if not already installed
if [ ! -d "${BATS_DIR}/bats-core" ]; then
    echo -e "${YELLOW}Installing bats-core...${NC}"
    git clone https://github.com/bats-core/bats-core.git "${BATS_DIR}/bats-core"
    echo -e "${GREEN}bats-core installed successfully.${NC}"
else
    echo -e "${GREEN}bats-core is already installed.${NC}"
fi

# Install bats-support if not already installed
if [ ! -d "${BATS_DIR}/bats-support" ]; then
    echo -e "${YELLOW}Installing bats-support...${NC}"
    git clone https://github.com/bats-core/bats-support.git "${BATS_DIR}/bats-support"
    echo -e "${GREEN}bats-support installed successfully.${NC}"
else
    echo -e "${GREEN}bats-support is already installed.${NC}"
fi

# Install bats-assert if not already installed
if [ ! -d "${BATS_DIR}/bats-assert" ]; then
    echo -e "${YELLOW}Installing bats-assert...${NC}"
    git clone https://github.com/bats-core/bats-assert.git "${BATS_DIR}/bats-assert"
    echo -e "${GREEN}bats-assert installed successfully.${NC}"
else
    echo -e "${GREEN}bats-assert is already installed.${NC}"
fi

# Install bats-file if not already installed
if [ ! -d "${BATS_DIR}/bats-file" ]; then
    echo -e "${YELLOW}Installing bats-file...${NC}"
    git clone https://github.com/bats-core/bats-file.git "${BATS_DIR}/bats-file"
    echo -e "${GREEN}bats-file installed successfully.${NC}"
else
    echo -e "${GREEN}bats-file is already installed.${NC}"
fi

echo -e "\n${GREEN}BATS and all its dependencies have been installed successfully.${NC}"
echo -e "To run tests, use: ${YELLOW}${BATS_DIR}/bats-core/bin/bats tests/test_*.bats${NC}"