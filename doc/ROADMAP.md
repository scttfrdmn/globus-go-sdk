<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Development Roadmap

This document outlines the development plan and upcoming milestones for the Globus Go SDK project.

## Current Status

The project has established its core structure and organization:
- Repository structure following Go best practices
- Alignment with official Globus SDKs
- Core package infrastructure
- Quality assurance tools (linting, testing, CI)
- Documentation framework

## Next Steps

### Phase 1: Core Implementation (Current)

#### Auth Package
- [x] Implement OAuth authorization flows
  - [x] Authorization code flow
  - [x] Refresh token flow
  - [x] Client credentials flow
- [x] Add token management
  - [x] Token validation utilities
  - [x] Token expiry handling
  - [ ] Token storage interface
  - [ ] Memory token storage
  - [ ] File-based token storage
- [x] Enhance error handling
  - [x] Standard error types
  - [x] Error checking utilities
  - [x] Improved error parsing
- [ ] Complete authentication helpers
  - [ ] Create token refresh workflows
  - [ ] Implement persistent token storage
  - [ ] Add session management features
- [ ] Write comprehensive tests
  - [x] Unit tests with mocks
  - [ ] Integration tests with sandbox environment

#### Groups Package
- [x] Implement Groups API client
  - [x] Group listing and filtering
  - [x] Group creation and management
  - [x] Membership operations
- [x] Add pagination support
- [x] Write tests for operations
  - [ ] Add edge case tests
  - [ ] Add integration tests

#### Transfer Service (Partially Complete)
- [x] Implement basic Transfer API client
- [ ] Expand transfer capabilities
  - [ ] Add recursive directory transfer support
  - [ ] Implement resumable transfers
  - [ ] Add batch transfer capabilities
  - [ ] Create transfer monitoring tools
- [ ] Add comprehensive testing
  - [x] Unit tests for client methods
  - [ ] Integration tests with real endpoints
  - [ ] Performance tests for large transfers

#### Core Infrastructure Enhancements
- [x] Improve error handling
- [ ] Add retry mechanism for transient failures
- [ ] Implement request/response logging
- [x] Create common utilities for testing

### Phase 2: CI/CD and Quality (Parallel to Phase 1)

- [ ] Configure GitHub repository settings
  - [ ] Branch protection rules
  - [ ] Required status checks
  - [ ] Code owners file
- [ ] Set up secrets for CI environment
  - [ ] GLOBUS_CLIENT_ID
  - [ ] GLOBUS_CLIENT_SECRET
- [ ] Enable Codecov integration
- [ ] Create issue templates
  - [ ] Bug report
  - [ ] Feature request
  - [ ] Question
- [ ] Set up project board for tracking features

### Phase 3: Documentation

- [ ] Complete API documentation
  - [ ] GoDoc-compatible comments
  - [ ] Examples for each service and operation
  - [x] Auth package documentation
  - [ ] Groups package documentation
  - [ ] Transfer package documentation
- [ ] Create getting started guide
  - [ ] Basic authentication flow
  - [ ] Common operations
  - [ ] Configuration options
- [ ] Document authentication flows
  - [ ] Sequence diagrams
  - [ ] Configuration examples
- [ ] Implement CLI examples
  - [ ] Authentication flow application
  - [ ] Group management utility
  - [ ] File transfer utility with progress monitoring
- [ ] Add badges to README
  - [ ] Build status
  - [ ] Code coverage
  - [ ] Go Report Card
  - [ ] GoDoc

### Phase 4: Transfer Service

- [ ] Implement Transfer API client
  - [ ] Endpoint management
  - [ ] Transfer task submission
  - [ ] Task monitoring
- [ ] Add high-level transfer operations
- [ ] Create examples for common use cases
- [ ] Write comprehensive tests

### Phase 5: First Release

- [ ] Finalize versioning strategy
- [ ] Create CHANGELOG.md
- [ ] Complete all tests and documentation
- [ ] Ensure CI pipeline is working correctly
- [ ] Tag and release v0.1.0

## Future Enhancements

### Additional Services
- [ ] Search API
- [ ] Flows API
- [ ] Timers API
- [ ] Compute API

### Performance Optimization
- [ ] Profile and optimize network operations
- [ ] Implement efficient retry/backoff strategies
- [ ] Add connection pooling for better performance
- [ ] Optimize memory usage for large transfers

### Advanced Features
- [ ] Automatic token refresh
- [ ] Connection pooling
- [ ] Rate limiting and backoff
- [ ] Advanced logging options
- [ ] Cross-platform token storage
- [ ] Multi-factor authentication support

### Dependency Management
- [ ] Review and minimize external dependencies
- [ ] Ensure compatibility with different Go versions
- [ ] Document version requirements
- [ ] Implement vendor management

## Timeline

| Phase                       | Estimated Timeline |
|-----------------------------|-------------------|
| Phase 1: Core Implementation | Weeks 1-4         |
| Phase 2: CI/CD and Quality   | Weeks 1-2         |
| Phase 3: Documentation       | Weeks 3-5         |
| Phase 4: Transfer Service    | Weeks 5-7         |
| Phase 5: First Release       | Week 8            |

## Progress Tracking

We'll track progress using:
- GitHub Issues for individual tasks
- GitHub Project Board for overall status
- Pull Requests for code review
- Weekly sync meetings

## Success Criteria

The v0.1.0 release should include:
- Complete Auth and Groups functionality
- >80% test coverage
- Comprehensive documentation
- Working examples
- CI/CD pipeline

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for details on how to contribute to this project.

## Revision History

| Date       | Version | Notes                                  |
|------------|---------|----------------------------------------|
| 2025-04-26 | 0.1     | Initial roadmap created                |
| 2025-04-26 | 0.2     | Updated with implementation progress and detailed next steps |