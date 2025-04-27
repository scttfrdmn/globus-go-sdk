<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Working with Claude to Develop the Globus Go SDK

This document provides guidance on using Claude (via Claude Code or other interfaces) to assist with the development of the Globus Go SDK.

## Project Status Summary

The Globus Go SDK has made significant progress with the following components implemented:

### Core Infrastructure (Complete)
- Core client with context support, error handling, and logging
- Configuration management with environment variable support
- HTTP transport layer
- Authorization mechanisms (static token, refreshable token, client credentials)

### Auth Package (Complete)
- OAuth2 flow implementations (authorization code, client credentials)
- Token management (refresh, introspect, revoke)
- Token models with utility methods

### Groups Package (Complete)
- Group management (create, list, update, delete)
- Membership management (add, remove, update roles)
- Role management operations

### SDK Configuration (Complete)
- Main SDK entry point for client creation
- Factory methods for service clients
- Environment-based configuration
- Client options pattern

### Examples (Complete)
- Auth flow example
- Groups management example
- Examples for other services

### Progress on Other Services
- Transfer client (implemented)
- Search client (implemented)
- Flows client (implemented)
- Compute client (implemented)

## Current File Structure
```
github.com/scttfrdmn/globus-go-sdk/
├── .github/                # GitHub workflows, issue templates
├── cmd/                    # CLI examples
│   └── examples/           # Example applications
├── pkg/                    # Main package
│   ├── core/               # Core functionality
│   │   ├── authorizers/    # Authentication mechanisms
│   │   ├── config/         # Configuration
│   │   ├── transport/      # HTTP transport
│   │   ├── client.go       # Base client
│   │   ├── errors.go       # Error handling
│   │   └── logger.go       # Logging
│   ├── services/           # Service-specific packages 
│   │   ├── auth/           # Auth service
│   │   ├── groups/         # Groups service
│   │   ├── transfer/       # Transfer service
│   │   ├── search/         # Search service
│   │   ├── flows/          # Flows service
│   │   └── compute/        # Compute service
│   └── globus.go           # Main SDK entry point
```

## Next Steps

1. **Testing Enhancements**:
   - Add integration tests for all services
   - Improve code coverage

2. **Documentation**:
   - Complete GoDoc documentation
   - Add more examples and usage guidance
   - Create API reference documentation

3. **Performance Optimization**:
   - Review and optimize HTTP client settings
   - Implement caching where appropriate

4. **Security Audit**:
   - Review token handling
   - Ensure secure defaults

## Overview

Claude can help implement the Globus Go SDK by:
- Generating code scaffolding for packages
- Helping with API implementations
- Writing tests and documentation
- Suggesting optimizations
- Troubleshooting issues

## Working with Claude

### Effective Prompting

When asking Claude to help with implementation:

1. **Be specific about the task**: "Help me implement the `RefreshToken` method for the Auth client" is better than "Help with the Auth package."

2. **Provide context**: Share the relevant parts of existing code when asking for new implementations.

3. **Specify requirements**: When asking for implementations, mention specific requirements like error handling, context support, etc.

4. **Request incremental changes**: Break down complex implementations into smaller pieces.

5. **Ask for tests**: Request tests when getting implementations.

### Example Prompts

Here are some example prompts for working with Claude:

#### Implementing a Specific Method

```
Implement the ExchangeAuthorizationCode method for the Auth client that:
1. Takes a context and authorization code
2. Makes a POST request to the token endpoint
3. Handles errors appropriately
4. Returns a TokenResponse
```

#### Writing Tests

```
Write unit tests for the ListGroups method in the Groups client. Include tests for:
1. Successful response with multiple groups
2. Empty response
3. Error handling
4. Context cancellation
```

#### Documentation

```
Generate GoDoc style documentation for the Auth package, including examples of:
1. Creating a client
2. Getting an authorization URL
3. Exchanging a code for tokens
4. Refreshing tokens
```

## Guidelines for Code Review

Ask Claude to review generated code:

```
Please review this implementation of the Groups client and suggest improvements:

[paste code here]
```

## Best Practices

When working with Claude on this project:

- **Verify all implementations**: Review code for correctness and security
- **Test thoroughly**: Don't rely solely on Claude-generated tests
- **Keep security in mind**: Review authentication implementations carefully
- **Maintain consistency**: Ensure Claude follows the established patterns
- **Iterate**: Refine implementations over multiple prompts

## Resources

Refer to the following resources when implementing:

- [Globus API Documentation](https://docs.globus.org/api/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
