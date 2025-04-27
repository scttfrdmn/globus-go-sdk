#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to run integration tests for the Globus Go SDK
# This script checks for required environment variables and runs the tests

set -e  # Exit on error

# Print colorized output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if required environment variables are set
check_env_vars() {
  local missing=0
  
  echo -e "${BLUE}Checking for required environment variables...${NC}"
  
  if [ -z "$GLOBUS_TEST_CLIENT_ID" ]; then
    echo -e "${RED}❌ GLOBUS_TEST_CLIENT_ID is not set${NC}"
    missing=1
  else
    echo -e "${GREEN}✅ GLOBUS_TEST_CLIENT_ID is set${NC}"
  fi
  
  if [ -z "$GLOBUS_TEST_CLIENT_SECRET" ]; then
    echo -e "${RED}❌ GLOBUS_TEST_CLIENT_SECRET is not set${NC}"
    missing=1
  else
    echo -e "${GREEN}✅ GLOBUS_TEST_CLIENT_SECRET is set${NC}"
  fi
  
  # Optional but recommended for transfer tests
  if [ -z "$GLOBUS_TEST_SOURCE_ENDPOINT_ID" ]; then
    echo -e "${YELLOW}⚠️  GLOBUS_TEST_SOURCE_ENDPOINT_ID is not set (transfer tests may be limited)${NC}"
  else
    echo -e "${GREEN}✅ GLOBUS_TEST_SOURCE_ENDPOINT_ID is set${NC}"
  fi
  
  if [ -z "$GLOBUS_TEST_DEST_ENDPOINT_ID" ]; then
    echo -e "${YELLOW}⚠️  GLOBUS_TEST_DEST_ENDPOINT_ID is not set (transfer tests may be limited)${NC}"
  else
    echo -e "${GREEN}✅ GLOBUS_TEST_DEST_ENDPOINT_ID is set${NC}"
  fi
  
  # Optional for group tests
  if [ -z "$GLOBUS_TEST_GROUP_ID" ]; then
    echo -e "${YELLOW}⚠️  GLOBUS_TEST_GROUP_ID is not set (some group tests may be skipped)${NC}"
  else
    echo -e "${GREEN}✅ GLOBUS_TEST_GROUP_ID is set${NC}"
  fi
  
  if [ $missing -eq 1 ]; then
    echo -e "${RED}❌ Some required environment variables are missing.${NC}"
    echo -e "${BLUE}Please set them before running integration tests.${NC}"
    echo -e "${YELLOW}Example:${NC}"
    echo -e "${GREEN}export GLOBUS_TEST_CLIENT_ID=\"your-client-id\"${NC}"
    echo -e "${GREEN}export GLOBUS_TEST_CLIENT_SECRET=\"your-client-secret\"${NC}"
    echo -e "${BLUE}Or create a .env.test file in the project root.${NC}"
    exit 1
  fi
}

# Function to load environment variables from .env.test file if it exists
load_env_file() {
  if [ -f ".env.test" ]; then
    echo -e "${BLUE}Loading environment variables from .env.test file...${NC}"
    
    # Read each line from .env.test
    while IFS= read -r line || [[ -n "$line" ]]; do
      # Skip empty lines and comments
      if [[ -z "$line" || "$line" =~ ^# ]]; then
        continue
      fi
      
      # Split the line into key and value
      key=$(echo "$line" | cut -d= -f1)
      value=$(echo "$line" | cut -d= -f2-)
      
      # Export the variable
      export "$key"="$value"
      echo -e "${GREEN}Loaded $key${NC}"
    done < ".env.test"
    
    echo -e "${GREEN}Environment variables loaded successfully.${NC}"
  else
    echo -e "${YELLOW}No .env.test file found, using existing environment variables.${NC}"
    echo -e "${BLUE}You can create a .env.test file for easier credential management.${NC}"
    echo -e "${BLUE}See doc/INTEGRATION_TESTING_SETUP.md for details.${NC}"
  fi
}

# Function to run tests with a specific pattern
run_tests() {
  local package=$1
  local pattern=$2
  
  if [ -z "$pattern" ]; then
    echo -e "${BLUE}Running all tests for ${GREEN}$package${NC}"
    go test -v -tags=integration $package
  else
    echo -e "${BLUE}Running tests matching ${GREEN}$pattern${BLUE} for ${GREEN}$package${NC}"
    go test -v -tags=integration $package -run $pattern
  fi
}

# Main function
main() {
  # Show header
  echo -e "${BLUE}=========================================================${NC}"
  echo -e "${BLUE}        Globus Go SDK Integration Test Runner${NC}"
  echo -e "${BLUE}=========================================================${NC}"

  # Load environment variables from .env.test file if it exists
  load_env_file
  
  # Check for required environment variables
  check_env_vars
  
  echo ""
  echo -e "${BLUE}Running integration tests...${NC}"

  # Check if verification flag is present
  if [ "$1" = "--verify" ] || [ "$1" = "-v" ]; then
    echo -e "${YELLOW}Running verification test to check environment setup...${NC}"
    go test -v -tags=integration ./pkg -run TestIntegration_VerifySetup
    echo -e "${GREEN}Verification completed! If successful, you can now run the full test suite.${NC}"
    exit 0
  fi
  
  if [ $# -eq 0 ]; then
    # Run all integration tests
    echo -e "${BLUE}Running integration tests for ${GREEN}all packages${NC}"
    go test -v -tags=integration ./...
  elif [ $# -eq 1 ]; then
    # Run tests for a specific package
    run_tests "./$1/..." ""
  elif [ $# -eq 2 ]; then
    # Run tests for a specific package with a pattern
    run_tests "./$1/..." "$2"
  else
    echo -e "${YELLOW}Usage: $0 [package] [pattern]${NC}"
    echo -e "${BLUE}Examples:${NC}"
    echo -e "  ${GREEN}$0 --verify                       ${BLUE}# Run verification test only${NC}"
    echo -e "  ${GREEN}$0                               ${BLUE}# Run all integration tests${NC}"
    echo -e "  ${GREEN}$0 pkg/services/auth             ${BLUE}# Run auth integration tests${NC}"
    echo -e "  ${GREEN}$0 pkg/services/transfer Transfer ${BLUE}# Run transfer tests with 'Transfer' in the name${NC}"
    exit 1
  fi
}

# Run the script
main "$@"