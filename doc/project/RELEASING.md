<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Releasing the Globus Go SDK

This document outlines the steps to create a new release of the Globus Go SDK.

## Prerequisites

Before starting the release process, ensure you have:

1. Proper access to the GitHub repository
2. The necessary tools installed:
   - Git
   - Go (matching the minimum version in go.mod)
   - golangci-lint (for code quality checks)
3. All tests are passing locally
4. All PRs that should be included in the release are merged

## Release Preparation

1. **Verify the codebase is ready**:
   - Run `make test` to ensure all tests pass
   - Run `make lint` to ensure code quality
   - Run `make verify-credentials` to test against Globus APIs (requires proper credentials)

2. **Update documentation**:
   - Update the version number in `pkg/globus.go`
   - Update `CHANGELOG.md` with all changes since the last release
   - Ensure README.md is up to date with the latest features

3. **Run integration tests**:
   - Set up the `.env.test` file with your Globus credentials
   - Run `make test-integration` to ensure API integration is working

## Creating the Release

1. **Create and push a release branch**:
   ```bash
   git checkout -b release-vX.Y.Z
   git push origin release-vX.Y.Z
   ```

2. **Create a PR for the release**:
   - Title: "Release vX.Y.Z"
   - Description: Summary of the main changes in this release
   - Apply the "release" label

3. **After PR approval and merge**:
   - Pull the latest main branch
   - Create and push the version tag:
     ```bash
     git tag -a vX.Y.Z -m "Release vX.Y.Z"
     git push origin vX.Y.Z
     ```

4. **Create the GitHub release**:
   - Navigate to the "Releases" section in the GitHub repository
   - Click "Draft a new release"
   - Choose the tag you just created
   - Title: "Globus Go SDK vX.Y.Z"
   - Description: Copy the entry from CHANGELOG.md for this release
   - If it's a pre-release, mark the "This is a pre-release" checkbox
   - Click "Publish release"

5. **Update the Go module proxy**:
   ```bash
   GOPROXY=proxy.golang.org go list -m github.com/scttfrdmn/globus-go-sdk@vX.Y.Z
   ```

## Post-Release Actions

1. **Announcements**:
   - Post an announcement to relevant channels
   - Update documentation site if applicable

2. **Version bump for continued development**:
   - Create a PR to update version number in `pkg/globus.go` to next development version (e.g., vX.Y.Z+1-dev)
   - Add a new "Unreleased" section to the CHANGELOG.md

3. **Review and plan**:
   - Review any issues that arose during the release process
   - Plan the next release cycle
   - Update the roadmap if needed

## Release Types

### Patch Releases (vX.Y.Z → vX.Y.Z+1)

For bug fixes and minor improvements that do not change the API.

### Minor Releases (vX.Y.Z → vX.Y+1.0)

For new features, improvements, and non-breaking changes.

### Major Releases (vX.Y.Z → vX+1.0.0)

For major features and breaking changes to the API.

## Versioning Guidelines

We follow [Semantic Versioning](https://semver.org/) (SemVer) for our releases:

- **MAJOR version** when you make incompatible API changes
- **MINOR version** when you add functionality in a backward compatible manner
- **PATCH version** when you make backward compatible bug fixes

For pre-release versions, use a suffix like `-alpha.1`, `-beta.1`, or `-rc.1`.