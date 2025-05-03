<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# GitHub Actions Status and Plan

This document outlines the current status of GitHub Actions workflows in the Globus Go SDK project and the plan for improvements.

## Current Status

The GitHub Actions workflows have been updated with the following changes:

1. **Workflow Directory Structure**:
   - Moved workflows from root directory to proper `.github/workflows/` location
   - Created standardized workflow files for CI, Documentation, and Release

2. **Workflow Configuration Updates**:
   - Updated to latest action versions (checkout@v4, setup-go@v5, etc.)
   - Added Go 1.22 to the matrix alongside 1.21
   - Improved dependency caching
   - Enhanced security scanning
   - Added proper API documentation generation

3. **Current Issues**:
   - Import cycle in the codebase prevents successful builds
   - Build, test, and lint jobs are failing due to this cycle
   - Documentation workflow is the only one currently working

## Temporary Fixes

To maintain a working CI pipeline while resolving the underlying issues, the following temporary changes have been made:

1. **Manual Trigger Only**:
   - CI, Go, and CodeQL workflows are set to manual trigger only (workflow_dispatch)
   - This prevents failed CI jobs from blocking PRs
   - The Documentation workflow still runs automatically on push/PR

2. **Documentation Updates**:
   - Updated documentation generation to use godoc for API documentation
   - Documentation workflow runs successfully
   - GitHub Actions badges in README updated to reflect actual workflow status

## Action Plan

### Short-term (Next PR)

1. **Fix Import Cycle**:
   - Follow the plan in [IMPORT_CYCLE_RESOLUTION.md](IMPORT_CYCLE_RESOLUTION.md)
   - Structure interfaces correctly to avoid circular dependencies
   - Ensure proper implementation of interfaces

2. **Update Build Script**:
   - Modify build scripts to handle examples separately from core library
   - Add error handling for examples that may have import issues

### Medium-term

1. **Re-enable Workflows**:
   - Once the core package builds successfully, re-enable all workflows
   - Update triggers to run on push to main and pull requests
   - Validate full CI pipeline functionality

2. **Comprehensive Testing**:
   - Add integration test improvements
   - Add test summary reporting
   - Improve code coverage reporting

### Long-term

1. **Enhanced Quality Checks**:
   - Add more sophisticated linting rules
   - Implement semantic version checking for releases
   - Add performance benchmarking for critical operations

2. **Documentation Improvements**:
   - Automated API documentation generation
   - Release notes generation
   - Examples validation

## Testing Locally

To test GitHub Actions workflows locally before pushing to the repository:

1. Install [act](https://github.com/nektos/act) for local GitHub Actions testing:
   ```bash
   brew install act
   ```

2. Run workflows locally:
   ```bash
   act -j build   # Run the build job
   act -j lint    # Run the lint job
   act -j test    # Run the test job
   ```

3. Verify workflow changes:
   ```bash
   act -n         # Dry run to check workflow configuration
   ```

## References

- [GitHub Actions documentation](https://docs.github.com/en/actions)
- [Go GitHub Actions](https://github.com/marketplace?type=actions&query=go)
- [Nektos Act](https://github.com/nektos/act) - Run GitHub Actions locally