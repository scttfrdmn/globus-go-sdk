#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to run security scans on the Globus Go SDK codebase
# This script runs various security scanning tools and reports the results

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Function to install security tools
install_tools() {
    echo -e "${YELLOW}Checking for required security tools...${NC}"
    
    # Check for gosec
    if ! command_exists gosec; then
        echo -e "${YELLOW}Installing gosec...${NC}"
        go install github.com/securego/gosec/v2/cmd/gosec@latest
        echo -e "${GREEN}Installed gosec successfully${NC}"
    else
        echo -e "${GREEN}gosec is already installed${NC}"
    fi
    
    # Check for nancy
    if ! command_exists nancy; then
        echo -e "${YELLOW}Installing nancy...${NC}"
        curl -sSfL https://raw.githubusercontent.com/sonatype-nexus-community/nancy/master/scripts/install-nancy.sh | sh -s -- -b "$(go env GOPATH)/bin"
        echo -e "${GREEN}Installed nancy successfully${NC}"
    else
        echo -e "${GREEN}nancy is already installed${NC}"
    fi
    
    # Check for gitleaks
    if ! command_exists gitleaks; then
        echo -e "${YELLOW}Installing gitleaks...${NC}"
        GITLEAKS_VERSION="8.8.8"
        
        TEMP_DIR=$(mktemp -d)
        pushd "$TEMP_DIR" > /dev/null
        
        case "$(uname -s)" in
            Linux*)
                wget "https://github.com/zricethezav/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_linux_x64.tar.gz"
                tar -xzf "gitleaks_${GITLEAKS_VERSION}_linux_x64.tar.gz"
                ;;
            Darwin*)
                wget "https://github.com/zricethezav/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_darwin_x64.tar.gz"
                tar -xzf "gitleaks_${GITLEAKS_VERSION}_darwin_x64.tar.gz"
                ;;
            *)
                echo -e "${RED}Unsupported operating system. Please install gitleaks manually.${NC}"
                popd > /dev/null
                rm -rf "$TEMP_DIR"
                exit 1
                ;;
        esac
        
        chmod +x gitleaks
        mv gitleaks "$(go env GOPATH)/bin/"
        
        popd > /dev/null
        rm -rf "$TEMP_DIR"
        
        echo -e "${GREEN}Installed gitleaks successfully${NC}"
    else
        echo -e "${GREEN}gitleaks is already installed${NC}"
    fi
}

# Function to run gosec scan
run_gosec() {
    echo -e "\n${YELLOW}Running gosec security scan...${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p reports
    
    # Run gosec with various output formats
    gosec -fmt=json -out=reports/gosec-results.json ./...
    gosec -fmt=text -out=reports/gosec-results.txt ./...
    
    # Check if any issues were found
    ISSUES=$(jq '.Issues | length' reports/gosec-results.json)
    if [ "$ISSUES" -gt 0 ]; then
        echo -e "${RED}gosec found ${ISSUES} potential security issues${NC}"
        echo -e "${YELLOW}See reports/gosec-results.txt for details${NC}"
        ISSUES_FOUND=true
    else
        echo -e "${GREEN}No security issues found by gosec${NC}"
    fi
}

# Function to run nancy vulnerability scan
run_nancy() {
    echo -e "\n${YELLOW}Running nancy dependency vulnerability scan...${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p reports
    
    # Run nancy
    go list -json -m all > reports/go-modules.json
    if ! nancy sleuth < reports/go-modules.json > reports/nancy-results.txt; then
        echo -e "${RED}nancy found vulnerable dependencies${NC}"
        echo -e "${YELLOW}See reports/nancy-results.txt for details${NC}"
        ISSUES_FOUND=true
    else
        echo -e "${GREEN}No vulnerable dependencies found by nancy${NC}"
    fi
}

# Function to run gitleaks
run_gitleaks() {
    echo -e "\n${YELLOW}Running gitleaks secrets scan...${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p reports
    
    # Run gitleaks
    if ! gitleaks detect --report-format json --report-path reports/gitleaks-report.json; then
        echo -e "${RED}gitleaks found potential secrets in the codebase${NC}"
        echo -e "${YELLOW}See reports/gitleaks-report.json for details${NC}"
        ISSUES_FOUND=true
    else
        echo -e "${GREEN}No secrets found by gitleaks${NC}"
    fi
}

# Function to run go vet
run_govet() {
    echo -e "\n${YELLOW}Running go vet code analysis...${NC}"
    
    # Create output directory if it doesn't exist
    mkdir -p reports
    
    # Run go vet and capture output
    if ! go vet ./... 2> reports/govet-results.txt; then
        echo -e "${RED}go vet found issues in the code${NC}"
        echo -e "${YELLOW}See reports/govet-results.txt for details${NC}"
        ISSUES_FOUND=true
    else
        echo -e "${GREEN}No issues found by go vet${NC}"
    fi
}

# Main function
main() {
    # Print header
    echo -e "${YELLOW}Starting security scan for Globus Go SDK${NC}"
    echo -e "${YELLOW}=======================================${NC}"
    
    # Install security tools if needed
    install_tools
    
    # Initialize issues flag
    ISSUES_FOUND=false
    
    # Run security scans
    run_gosec
    run_nancy
    run_gitleaks
    run_govet
    
    # Print summary
    echo -e "\n${YELLOW}Security Scan Summary${NC}"
    echo -e "${YELLOW}====================${NC}"
    
    if [ "$ISSUES_FOUND" = true ]; then
        echo -e "${RED}Security scan found potential issues.${NC}"
        echo -e "${YELLOW}Please review the reports in the 'reports' directory.${NC}"
        exit 1
    else
        echo -e "${GREEN}All security scans passed successfully!${NC}"
        exit 0
    fi
}

# Run main function
main "$@"