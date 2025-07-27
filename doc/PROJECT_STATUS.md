<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Project Status

This document provides an overview of the current development status and roadmap for the Globus Go SDK.

## Current Status (v0.2.0 Development)

The Globus Go SDK is currently in active development for its v0.2.0 release. This version focuses on ensuring all core functionality is robust, well-tested, and ready for production use.

### Key Features Status

| Feature | Status | Details |
|---------|--------|---------|
| Auth Client | ✅ Complete | OAuth flows, token management, identity handling |
| Transfer Client | ✅ Complete | File/directory operations, task management, task monitoring |
| Groups Client | ✅ Complete | Group CRUD, membership management |
| Search Client | ✅ Complete | Index operations, search queries |
| Flows Client | ✅ Complete | Flow execution and monitoring |
| Compute Client | ✅ Complete | Compute job submission and monitoring |
| Rate Limiting | ✅ Complete | Adaptive token bucket algorithm, backoff strategies |
| Error Handling | ✅ Complete | Standardized error types, useful helper functions |
| Integration Testing | 🟡 In Progress | Comprehensive tests against real Globus API |

### Current Development Focus

The current development focus is on:

1. **Integration Testing**: Comprehensive testing against the real Globus API to ensure all operations function correctly with proper authorization, error handling, and rate limiting.

2. **Error Handling Improvements**: Standardizing error handling across all services with proper classification and helpful user-facing messages.

3. **API Stabilization**: Ensuring consistency in method signatures, parameter naming, and return types across the SDK.

## Release Criteria for v0.2.0

Before releasing v0.2.0, the following criteria must be met:

1. All integration tests for core services must pass against real Globus endpoints.
2. Documentation must be complete and accurate for all public APIs.
3. Error handling must be consistent with clear error messages and proper classification.
4. Rate limiting must be properly implemented with adaptive behavior.
5. Code coverage for unit tests must be at least 80%.

## Known Issues

- Some transfers may experience timeout issues with large directories
- Error messages from the underlying API are not always propagated clearly
- Some rate limit information from headers may be inconsistently available

## Next Steps After v0.2.0

Once v0.2.0 is released, development will focus on:

1. **Performance Optimizations**: Improving transfer performance for large files and directories
2. **More Example Applications**: Building example applications to demonstrate common workflows
3. **CLI Enhancements**: Expanding the CLI tool's capabilities
4. **Advanced Transfer Features**: Adding support for resumable transfers and more transfer options

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md) for details on how to contribute to the project.