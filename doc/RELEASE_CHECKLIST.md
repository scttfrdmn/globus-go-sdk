<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Release Checklist for Globus Go SDK

This document outlines the steps required to prepare and publish a release of the Globus Go SDK.

## Pre-Release Checklist

### Documentation
- [ ] Update documentation to reflect any changes
- [ ] Ensure API documentation (godoc) is complete and accurate
- [ ] Verify README contains up-to-date installation and usage instructions
- [ ] Update CHANGELOG.md with all significant changes
- [ ] Check examples for correctness and relevance

### Testing
- [ ] All unit tests passing
- [ ] Integration tests running with real credentials
- [ ] Code coverage is acceptable (target: >80%)
- [ ] Tests exist for all new functionality

### Code Quality
- [ ] No import cycles present
- [ ] No unused code or dead imports
- [ ] Code follows Go best practices
- [ ] Code has been reviewed for security concerns
- [ ] Error handling is consistent and comprehensive

### Performance
- [ ] Performance benchmarks run and results documented
- [ ] No performance regressions from previous release

### Compatibility
- [ ] All supported Globus API versions are tested
- [ ] Backward compatibility maintained (or breaking changes documented)
- [ ] Go module compatibility verified (go.mod and go.sum up to date)

## Release Process

### 1. Version Update
- [ ] Update version number in `pkg/globus.go`
- [ ] Update version number in any other relevant files
- [ ] Ensure version follows semantic versioning (X.Y.Z)

### 2. Final Testing
- [ ] Run the full test suite
- [ ] Run integration tests against all supported services
- [ ] Verify examples with the new version

### 3. Documentation Finalization
- [ ] Update CHANGELOG.md with final release date
- [ ] Review documentation for version references
- [ ] Update any version-specific documentation

### 4. Create Release
- [ ] Create a new git tag: `git tag v0.1.0`
- [ ] Push tag to repository: `git push origin v0.1.0`
- [ ] Create GitHub release with release notes
- [ ] Upload any release artifacts

### 5. Publishing
- [ ] Publish to Go package repository
- [ ] Update pkg.go.dev documentation
- [ ] Announce release on relevant channels

## Post-Release
- [ ] Verify the released package can be installed: `go get github.com/scttfrdmn/globus-go-sdk@v0.1.0`
- [ ] Verify the documentation appears correctly on pkg.go.dev
- [ ] Check that examples work with the released version
- [ ] Create new development branch for next version

## Current Status for v0.1.0

### Completed
- [x] Core implementation of all major Globus services
- [x] Connection pooling with interface extraction to resolve import cycles
- [x] Rate limiting and backoff strategies
- [x] Documentation structure and content
- [x] Authentication and token management
- [x] Error handling and standardization
- [x] Transfer functionality for files and directories
- [x] Groups service with membership management
- [x] Search functionality with advanced query support
- [x] Example applications for major services

### In Progress
- [ ] Finalize integration tests with credentials
- [ ] Complete performance benchmarking
- [ ] Final security audit

### Ready for Release
- v0.1.0 is nearly ready for release
- Key remaining tasks are to complete final integration testing and performance validation
- All major functionality is implemented and working correctly

### Release Timeline
- Target release date: End of Q2 2025