<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Release Process & Checklist

This document provides a comprehensive and formal process for releasing new versions of the Globus Go SDK, ensuring consistent quality, stability, and compatibility.

## Release Classification

Each release is classified according to semantic versioning principles:

| Release Type | Version Format | Description |
|--------------|----------------|-------------|
| **Patch**    | `x.y.Z`        | Bug fixes and minor improvements with no API changes |
| **Minor**    | `x.Y.0`        | New functionality in a backward-compatible manner |
| **Major**    | `X.0.0`        | Changes that break backward compatibility |

**Note:** Until v1.0.0 is released, minor versions may contain breaking changes, but these will be clearly documented.

## Release Prerequisites

Before starting any release process, verify:

- [ ] All planned features and bug fixes for this release are merged to main
- [ ] The CHANGELOG.md is up to date with all changes included in this release
- [ ] All documentation is updated to reflect the current state of the code
- [ ] There are no open critical issues blocking this release

## Step 1: Pre-Release Verification

### Code Quality & Validation

- [ ] All linting checks pass without errors: `go vet ./...`, `golint ./...`
- [ ] Go formatting is properly applied: `go fmt ./...`
- [ ] SPDX license headers are present on all source files: `./scripts/check-license-headers.sh`
- [ ] All unit tests pass: `go test ./... -short`
- [ ] All integration tests pass: `./scripts/run_integration_tests.sh`
- [ ] All examples compile and run without errors
- [ ] Documentation builds successfully: `cd docs-site && ./build.sh`

### API Compatibility

Before proceeding with a release, API compatibility must be verified. 

#### For Patch Releases (x.y.Z):

- [ ] Run API compatibility verification tool against previous patch release:
  ```
  ./scripts/verify_api_compatibility.sh v0.9.x v0.9.z -level=patch
  ```
- [ ] Verify all API signatures match the previous patch release
- [ ] Document any deviations with justification if they exist

#### For Minor Releases (x.Y.0):

- [ ] Run API compatibility verification tool against previous minor release:
  ```
  ./scripts/verify_api_compatibility.sh v0.x.0 v0.y.0 -level=minor
  ```
- [ ] Verify all API signatures are backward compatible with previous minor release
- [ ] Document any backward-incompatible changes (acceptable pre-v1.0.0 but must be documented)

#### For Major Releases (X.0.0):

- [ ] Document all breaking changes with migration paths
- [ ] Create comprehensive migration guide for users
- [ ] Run API compatibility verification tool to generate full API difference report:
  ```
  ./scripts/verify_api_compatibility.sh vx.0.0 vy.0.0 -level=major
  ```

### Downstream Impact Assessment

- [ ] Test with downstream projects: `./scripts/test_dependent_projects.sh`
- [ ] Verify that all documented examples work with the new release
- [ ] Check compatibility with any known projects using the SDK

## Step 2: Release Preparation

### Version Update

- [ ] Update version number in `pkg/core/version.go`
- [ ] Verify version appears correctly in all relevant documentation
- [ ] Update copyright years in documentation if releasing in a new year

### Documentation Finalization

- [ ] Finalize CHANGELOG.md for the release with:
  - Release date in ISO 8601 format (YYYY-MM-DD)
  - Complete list of changes categorized as:
    - Added (new features)
    - Changed (changes in existing functionality)
    - Deprecated (soon-to-be removed features)
    - Removed (now removed features)
    - Fixed (bug fixes)
    - Security (vulnerability fixes)
  - Links to relevant issues and PRs

- [ ] Ensure all stability indicators in doc.go files are accurate

- [ ] Verify documentation site builds correctly with the new version

## Step 3: Release Creation

### Git Operations

- [ ] Create a release commit:
  ```
  git add pkg/core/version.go CHANGELOG.md
  git commit -m "Release v0.x.y"
  ```

- [ ] Tag the release:
  ```
  git tag -a v0.x.y -m "Release v0.x.y"
  ```

- [ ] Push the changes and tag to GitHub:
  ```
  git push origin main
  git push origin v0.x.y
  ```

### GitHub Release

- [ ] Create a GitHub release at https://github.com/scttfrdmn/globus-go-sdk/releases/new
- [ ] Select the tag just created
- [ ] Title the release "Globus Go SDK v0.x.y"
- [ ] Include the relevant section from CHANGELOG.md in the description
- [ ] Highlight any significant changes, deprecations, or breaking changes
- [ ] Add migration guidance if applicable
- [ ] Publish the release

### Documentation Deployment

- [ ] Build and deploy the documentation for the new version:
  ```
  cd docs-site
  ./build.sh
  ./sync-docs.sh v0.x.y
  ```

- [ ] Verify that the documentation site correctly shows the new version

## Step 4: Post-Release Verification

### Verification

- [ ] Verify the new release can be correctly installed via go get:
  ```
  cd $(mktemp -d)
  go mod init test-globus-sdk
  go get github.com/scttfrdmn/globus-go-sdk@v0.x.y
  go build
  ```

- [ ] Check that the GitHub release page shows the correct tag and description
- [ ] Verify that the documentation site correctly shows the new version
- [ ] Test a simple program that imports and uses the SDK to confirm basic functionality
- [ ] Run the verification example: `./cmd/verify-credentials/main.go version`

### Announcements

- [ ] Create an announcement issue with release highlights
- [ ] Post the release announcement in appropriate channels
- [ ] Update the release status in RELEASE_STATUS.md

## Step 5: Preparation for Next Development Cycle

- [ ] Create a new "Unreleased" section in CHANGELOG.md
- [ ] Review the roadmap and plan for the next release
- [ ] Update project boards and milestone tracking
- [ ] Triage any pending issues and PRs

## Additional Consideration for Releases with Breaking Changes

For releases that contain breaking changes:

1. Major version changes (X.0.0) - Expected to have breaking changes:
   - Complete migration guide must be provided
   - Consider maintaining the previous major version for critical fixes
   - Provide tooling to help users migrate if possible

2. Minor versions pre-v1.0.0 (0.X.0) - May have documented breaking changes:
   - Clearly document all breaking changes in CHANGELOG.md
   - Provide detailed migration steps for each breaking change
   - Consider providing a compatibility layer if changes are significant

## Security Releases

Security releases follow an expedited process:

1. Prepare the fix in a private branch or fork
2. Follow the standard release process but prioritize speed
3. Release a patch version as soon as the fix is verified
4. Coordinate disclosure with the security issue reporter
5. Provide clear documentation on the vulnerability and fix

## Pre-Release Versions

For pre-release testing, use the following version formats:

- Alpha: `v0.x.y-alpha.n`
- Beta: `v0.x.y-beta.n`
- Release Candidate: `v0.x.y-rc.n`

Follow the standard release process, but clearly mark these as pre-releases on GitHub.

## Release Schedule Guidelines

- **Patch releases**: As needed for bug fixes, typically within 1-2 weeks of critical bugs
- **Minor releases**: Planned according to feature roadmap, typically every 1-3 months
- **Major releases**: Planned well in advance with full migration support, typically every 6-12 months

## Troubleshooting Common Release Issues

| Issue | Solution |
|-------|----------|
| Tests failing | Investigate and fix failed tests before proceeding |
| Documentation build issues | Verify all markdown links and formatting |
| API compatibility tool errors | Review changes that break compatibility and document or fix |
| Tag already exists | Verify if release was already created or delete tag if incorrect |
| Downstream project failures | Fix SDK issues or provide guidance to downstream projects |

## Checklist Completion

Release Manager: ______________________________

Release Version: v___.___.___

Date Completed: __________________

Signature: _______________________________

## Appendices

### A. Version Numbering Examples

- v0.9.15: Fifteenth patch release of the 0.9.x series
- v0.10.0: New minor release with backward-compatible new features
- v1.0.0: First stable release with API stability guarantees

### B. API Compatibility Verification Procedure

The API compatibility verification tool checks for breaking changes between releases:

1. Function signatures changes
2. Interface method additions or removals
3. Struct field changes
4. Constant value changes
5. Public API removals

See `./scripts/verify_api_compatibility.sh` for the implementation.

### C. Handling Edge Cases

1. **Emergency fixes**: For critical bugs, follow the security release process
2. **Reverting releases**: If a release must be reverted, create a new release that explicitly reverts the changes
3. **Handling conflicts**: If conflicts arise during the release process, resolve them before proceeding