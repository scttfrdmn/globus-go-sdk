# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Globus Go SDK Alignment with Official Globus SDKs

This document describes how the Globus Go SDK aligns with Globus' official Python and JavaScript SDKs in terms of structure, naming conventions, and design patterns.

## Foundational Goal

A foundational goal of this project is to follow, as much as possible, the structure and patterns established by the official Globus SDKs (Python and JavaScript). By maintaining this alignment:

1. Developers familiar with other Globus SDKs can easily adopt the Go SDK
2. The SDK maintains consistency with Globus' design philosophies
3. Future updates to the Globus APIs can be implemented following established patterns

## Alignment with Python SDK

The Python SDK (`globus-sdk-python`) uses the following structure:

- **Service-specific clients**: Each Globus service (Auth, Transfer, Groups, etc.) has its own client implementation
- **Authorization system**: Flexible authorizers that handle different authentication workflows
- **Base client functionality**: Common HTTP handling, error management, and configuration
- **Transport layer**: Manages the details of API communication

Our Go SDK follows this pattern by implementing:

- Service-specific packages (`auth`, `groups`, etc.)
- A common package for shared functionality
- Authorizer interfaces for flexible authentication
- Base client implementations that service-specific clients extend

## Alignment with JavaScript SDK

The JavaScript SDK (`globus-sdk-javascript`) uses these structural elements:

- **Modular services**: Each service is in its own directory with clear entry points
- **Core utilities**: Authentication, error handling, and logging are centralized 
- **TypeScript types**: Strong typing across the codebase
- **Flexible authentication**: Support for different token management approaches

Our Go SDK aligns by:

- Using strong typing inherent to Go
- Implementing modular service packages
- Centralizing core functions in the common package
- Providing flexible authentication options

## Key Differences

While we aim to follow the patterns of the official SDKs, some differences exist due to Go's language characteristics:

1. **Error handling**: Using Go's idiomatic error handling rather than exceptions
2. **Context support**: Extensive use of Go's `context.Context` for cancellation and timeouts
3. **Interfaces**: Leveraging Go interfaces for extensibility
4. **Concurrency**: Utilizing Go's goroutines and channels where appropriate

## Implementation Alignment

| Feature | Python SDK | JavaScript SDK | Go SDK |
|---------|------------|----------------|--------|
| Authentication | Authorizers | TokenManager | Authorizer interfaces |
| Service Clients | Class-based | Module-based | Interface-based |
| Configuration | Environment + code | Config objects | Option functions |
| Error Handling | Exceptions | Error objects | Error interfaces |
| Pagination | Iterator-based | Async/await | Context-based |

## Continuous Alignment

As the Globus APIs evolve, we commit to:

1. Reviewing changes in the official SDKs
2. Adapting our Go SDK to maintain alignment
3. Documenting any divergences and their rationales

By maintaining this alignment, the Globus Go SDK aims to provide a familiar and idiomatic experience for Go developers while remaining faithful to the design patterns established by Globus.