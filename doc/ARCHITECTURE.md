<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK Architecture

This document describes the architecture of the Globus Go SDK and how it aligns with the official Globus SDKs.

## Overview

The Globus Go SDK provides a Go language interface to Globus services. It follows the patterns established by the official Globus Python and JavaScript SDKs while leveraging Go's language features.

## Directory Structure

```
github.com/scttfrdmn/globus-go-sdk/
├── .github/                     # GitHub actions and configurations
├── cmd/                         # Command-line tools and examples
│   └── examples/                # Example applications for using the SDK
│       ├── auth/                # Auth examples
│       └── groups/              # Groups examples
├── doc/                         # Documentation
│   ├── ARCHITECTURE.md          # This document
│   └── ...                      # Other documentation
├── pkg/                         # Main SDK code
│   ├── core/                    # Core SDK functionality
│   │   ├── authorizers/         # Authentication mechanisms
│   │   ├── config/              # Configuration management
│   │   ├── transport/           # HTTP transport layer
│   │   ├── client.go            # Base client
│   │   ├── errors.go            # Error types and handling
│   │   └── logger.go            # Logging utilities
│   ├── services/                # Service-specific clients
│   │   ├── auth/                # Authentication service
│   │   ├── groups/              # Groups service
│   │   └── transfer/            # Transfer service (future)
│   └── globus.go                # Main entry point and configuration
├── .gitignore
├── ALIGNMENT.md                 # Details on alignment with official SDKs
├── go.mod
├── LICENSE
└── README.md
```

## Component Responsibilities

### Core Package

The `pkg/core` package provides foundational functionality used by all service clients:

- **client.go**: Defines the base HTTP client and common request handling
- **errors.go**: Defines error types and handling mechanisms
- **logger.go**: Provides logging capabilities

#### Authorizers

The `pkg/core/authorizers` package handles authentication:

- Defines the `Authorizer` interface for all auth mechanisms
- Provides implementations for different auth methods (token, null, etc.)
- Handles token refresh and authorization header generation

#### Transport

The `pkg/core/transport` package manages the HTTP communication:

- Provides methods for making API requests (GET, POST, etc.)
- Handles request serialization and response deserialization
- Manages headers and content types

#### Config

The `pkg/core/config` package handles SDK configuration:

- Loads configuration from environment variables
- Provides defaults for SDK behavior
- Allows for configuration customization

### Services

Each service in the `pkg/services` directory represents a Globus API:

- **auth**: Handles OAuth2 authentication flows and token management
- **groups**: Provides group management functionality
- **transfer**: Manages file transfers and endpoints (future)

Each service package follows a consistent structure:

- **client.go**: Service-specific client implementation
- **models.go**: Data models for the service
- Additional files for specific service functionality

## Authentication Flow

The SDK supports multiple authentication flows:

1. **Authorization Code Flow**:
   - Get an authorization URL from the auth client
   - User visits the URL and authorizes the application
   - Application receives a callback with an authorization code
   - Exchange the code for access and refresh tokens

2. **Refresh Token Flow**:
   - Use a refresh token to obtain a new access token
   - The `TokenAuthorizer` can automatically refresh tokens

3. **Client Credentials Flow** (future):
   - Direct exchange of client credentials for an access token

## Error Handling

The SDK uses Go's error handling patterns:

- Service methods return typed errors that can be examined
- Helper functions like `IsUnauthorized()` and `IsNotFound()` help categorize errors
- Detailed error messages include status codes and error descriptions

## Alignment with Official SDKs

This SDK aligns with the official Globus SDKs through:

1. **Consistent naming**: Using the same service and method names
2. **Similar client structure**: Service-specific clients with standardized methods
3. **Parallel authentication mechanisms**: Supporting the same auth flows
4. **Equivalent data models**: Using the same model structures and fields

See [ALIGNMENT.md](../ALIGNMENT.md) for detailed comparisons.

## Design Principles

The SDK follows these principles:

1. **Idiomatic Go**: Uses Go conventions and patterns
2. **Minimal Dependencies**: Relies primarily on the standard library
3. **Context Support**: All API operations accept a context for cancellation
4. **Strong Typing**: Uses Go's type system to provide compile-time checks
5. **Extensibility**: Interfaces allow for customization and extension