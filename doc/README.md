<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Documentation

Welcome to the Globus Go SDK documentation. This guide serves as a central index for all documentation resources.

## Getting Started

If you're new to the Globus Go SDK, these resources will help you get started:

- [Project README](../README.md) - Overview, installation, and basic examples
- [Getting Started Guide](guides/getting-started.md) - Step-by-step guide for new users
- [User Guide](guides/user-guide.md) - Comprehensive usage information

## Documentation Categories

### User Guides

Service-specific guides for using the SDK:

- [Authentication Guide](guides/authentication.md) - OAuth2 flows and token management
- [Transfer Guide](guides/transfer.md) - File transfer operations
- [Search Guide](guides/search.md) - Data search and discovery
- [Groups Guide](guides/groups.md) - Group management
- [Flows Guide](guides/flows.md) - Automation and workflow orchestration
- [Compute Guide](guides/compute.md) - Distributed computation
- [Timers Guide](guides/timers.md) - Scheduled tasks and operations

### Core Concepts & Features

Detailed documentation on key concepts:

- [Token Storage](topics/token-storage.md) - Token persistence and management
- [Error Handling](topics/error-handling.md) - Error patterns and recovery strategies
- [Rate Limiting](topics/rate-limiting.md) - Rate limitation and backoff strategies
- [Logging](topics/logging.md) - Logging and distributed tracing
- [Performance](topics/performance.md) - General performance considerations
- [Data Schemas](topics/data-schemas.md) - Data models and schema information

### Advanced Topics

Deeper dives into specialized capabilities:

- [Recursive Transfers](advanced/recursive-transfers.md) - Directory recursive transfers
- [Resumable Transfers](advanced/resumable-transfers.md) - Checkpoint and resume capabilities
- [Connection Pooling](advanced/connection-pooling.md) - HTTP connection management
- [Extending the SDK](advanced/extending.md) - Custom extensions and plugins
- [Multi-Factor Authentication](advanced/mfa.md) - MFA integration and flows

### Development & Contributing

For SDK contributors and developers:

- [Architecture](development/architecture.md) - Design patterns and principles
- [Contributing](development/contributing.md) - How to contribute to the SDK
- [Testing](development/testing.md) - Unit and integration testing
- [Security](development/security.md) - Security best practices and guidelines
- [Benchmarking](development/benchmarking.md) - Performance measurement

### Examples

Practical examples with code:

- [Authentication Examples](examples/authentication.md)
- [Transfer Examples](examples/transfer.md)
- [Search Examples](examples/search.md)
- [Groups Examples](examples/groups.md)
- [Flows Examples](examples/flows.md)
- [CLI Examples](examples/cli.md)

### Reference

Technical reference materials:

- [API Overview](reference/api-overview.md) - API structure and patterns
- [Configuration](reference/configuration.md) - Configuration options
- [Environment Variables](reference/environment.md) - Environment configuration
- [Error Codes](reference/error-codes.md) - Error reference and troubleshooting
- [Glossary](reference/glossary.md) - Terminology definitions

### Project Information

- [Roadmap](project/roadmap.md) - Future development plans
- [Status](project/status.md) - Implementation status
- [Changelog](project/changelog.md) - Version history
- [SDK Alignment](project/alignment.md) - Alignment with other Globus SDKs

## Documentation Structure

This documentation uses a hierarchical organization:

```
doc/
├── README.md              # This documentation index
├── guides/                # User-focused service guides
├── topics/                # Core concepts and features
├── advanced/              # Advanced use cases and features
├── development/           # Developer and contributor info
├── examples/              # Example code and use cases
├── reference/             # Technical reference
└── project/               # Project metadata
```

## Contributing to Documentation

We welcome contributions to improve the documentation. Please see [Contributing to Documentation](development/contributing.md#documentation) for guidelines.

## Documentation Roadmap

See our [Documentation Restructuring Plan](DOC_RESTRUCTURING_PLAN.md) for information about ongoing documentation improvements.