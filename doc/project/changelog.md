<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Changelog

All notable changes to the Globus Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2025-05-10

### Added
- Enhanced Compute client with workflow and task group capabilities:
  - Added workflow management (creation, execution, status tracking)
  - Implemented dependency graph support for complex compute workflows
  - Added task group functionality for parallel execution
  - Expanded container management capabilities
  - Added environment and secret management
- Improved API version compatibility checking:
  - Added version extraction from URLs
  - Implemented service-specific version compatibility rules
  - Added version checking toggles and custom version support
- Enhanced HTTP debugging with more detailed request/response logging
- Added new example for Compute workflows and task groups

### Fixed
- Improved error handling in transport layer
- Enhanced connection pool management for better stability
- Fixed integration tests for all service clients
- Standardized error reporting formats across services
- Improved thread safety in concurrent operations

## [0.8.0] - 2025-05-01

### Added
- Full Compute service client implementation with:
  - Endpoint management
  - Function registration and management
  - Execution operations with status monitoring
  - Batch task processing
  - Comprehensive models and error handling
- Version checking system with API compatibility verification
- Improved HTTP debugging capabilities
- Additional CLI examples demonstrating SDK features
- Expanded documentation on API version compatibility
  
### Fixed
- Fixed integration tests for recursive transfers by adding proper mock submission ID endpoints
- Improved test infrastructure with better error handling
- Enhanced mock server implementation for more realistic API behavior simulation
- Updated integration testing documentation with Compute service information
- Standardized SPDX license headers across all files
- Fixed various linting issues throughout the codebase

### Changed
- Updated version number to 0.8.0 to better reflect project maturity
- Reorganized project documentation for better discoverability
- Improved error handling with additional helper functions
- Enhanced README with current project status

## [0.2.0] - 2025-04-28

### Added

- Implemented enhanced logging and tracing system:
  - Structured logging with text and JSON formats
  - HTTP request/response tracing with trace IDs
  - Field-based contextual logging
  - Automatic sensitive data redaction in logs
  - Trace ID propagation for distributed tracing
  - Comprehensive logging configuration options
  - Detailed documentation and examples
- Added comprehensive integration testing:
  - Test infrastructure for resumable transfers with real endpoints
  - Environment variable loading from `.env.test`
  - Improved test setup and cleanup procedures
  - Detailed integration testing documentation
- Implemented resumable transfer functionality with checkpointing:
  - File-based checkpoint storage for persistently tracking transfer state
  - Batch processing of transfers for efficiency and error recovery
  - Progress tracking and status reporting
  - Automatic retries for failed transfers
  - Client methods for creating, resuming, and monitoring transfers
- Added connection pooling for improved performance:
  - Service-specific connection pools with optimized settings
  - Configurable pool parameters (connection limits, timeouts)
  - Automatic connection reuse for better performance
- Implemented robust rate limiting and backoff strategies:
  - Circuit breaker pattern to prevent cascading failures
  - Exponential backoff with jitter for retries
  - Rate limiting to avoid API throttling 
  - Response handler for rate limit headers
- Added new example applications:
  - Resumable transfers example with command-line interface
  - Logging and tracing example demonstrating various capabilities
  - Rate limiting and backoff example application
- Created verify-credentials utility for validating credentials:
  - Tests credentials against Auth, Transfer, Groups, and Search services
  - Multiple implementation options (SDK, API-only, standalone)
  - Comprehensive documentation and use instructions
- Created comprehensive documentation:
  - Resumable transfers guide (`doc/resumable-transfers.md`)
  - Logging and tracing guide (`doc/logging-and-tracing.md`)
  - Integration testing guide (`doc/INTEGRATION_TESTING.md`)
  - Rate limiting guide (`doc/rate-limiting.md`)
  - Performance benchmarking guide (`doc/performance-benchmarking.md`)

### Changed

- Updated authorizer interfaces to be more flexible and eliminate import cycles
- Improved error handling for better diagnostics and recovery
- Enhanced package structure for clearer API boundaries
- Updated all services to use connection pooling by default
- Standardized logging approach across all services
- Improved test coverage across all packages

### Fixed

- Resolved import cycles in multiple packages
- Fixed field naming inconsistencies in model structs
- Corrected error handling in authentication flows
- Fixed thread safety issues in connection management
- Addressed memory leaks in transfer operations

## [0.1.0] - 2025-04-26

### Added

- Implemented Search service client with comprehensive features:
  - Index management (create, read, update, delete)
  - Document operations (ingest, delete)
  - Advanced query support (match, term, range, boolean, geo queries)
  - Pagination with iterator pattern
  - Batch operations for large-scale document management
  - Task management with status tracking and waiting
  - Specialized error handling
- Implemented token storage interface with memory and file implementations
- Created token manager with automatic token refreshing
- Added recursive directory transfer functionality
- Implemented CLI example application
- Added comprehensive documentation:
  - Search client guide (`doc/search-client.md`)
  - Token storage guide (`doc/token-storage.md`)
  - Recursive transfers guide (`doc/recursive-transfers.md`)
  - User guide (`doc/user-guide.md`)
  - Data schemas reference (`doc/data-schemas.md`)
  - Error handling guide (`doc/error-handling.md`)
  - SDK extension guide (`doc/extending-the-sdk.md`)
- Enhanced error handling with typed errors and error checking utilities
- Added token validation utilities
- Implemented transfer client test additions
- Created group management example in CLI
- Enhanced authorization flows with persistent storage

### Changed

- Updated ROADMAP.md to reflect implementation progress
- Updated PROJECT_STATUS.md with completed tasks and new priorities
- Reorganized authorizer interfaces to reduce circular dependencies

### Fixed

- Resolved token refresh race conditions with mutex protection
- Fixed authorization flow to properly store refresh tokens
- Enhanced error handling throughout the codebase
- Prevented potential memory leaks in recursive transfers
- Improved thread safety in token storage implementations

## [0.0.1] - 2025-04-26

### Added

- Initial project structure
- Base client with context support
- HTTP transport with request/response handling
- Multiple authorizer types with tests
- Enhanced error types and validation helpers
- Configurable logging with levels
- Auth client implementation
- Groups client implementation
- Basic transfer client
- Environment variable support
- Development documentation
- Testing framework