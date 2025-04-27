# Changelog

All notable changes to the Globus Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Added new example applications:
  - Resumable transfers example with command-line interface
  - Logging and tracing example demonstrating various capabilities
- Created comprehensive documentation:
  - Resumable transfers guide (`doc/resumable-transfers.md`)
  - Logging and tracing guide (`doc/logging-and-tracing.md`)
  - Integration testing guide (`doc/INTEGRATION_TESTING.md`)

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