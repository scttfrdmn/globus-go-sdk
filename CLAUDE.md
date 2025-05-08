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

1. **API Stability Implementation**:
   - Complete stability indicators for all packages
   - Implement API compatibility verification tools
   - Add formal deprecation mechanisms
   - Continue enhancing CHANGELOG structure

2. **Testing Enhancements**:
   - Add integration tests for all services
   - Implement contract tests for interfaces
   - Improve code coverage
   - Create compatibility test suite

3. **Documentation**:
   - Complete GoDoc documentation
   - Add more examples and usage guidance
   - Create API reference documentation
   - Add migration guides for any breaking changes

4. **Performance Optimization**:
   - Review and optimize HTTP client settings
   - Implement caching where appropriate

5. **Security Audit**:
   - Review token handling
   - Ensure secure defaults

## Overview

Claude can help implement the Globus Go SDK by:
- Generating code scaffolding for packages
- Helping with API implementations
- Writing tests and documentation
- Suggesting optimizations
- Troubleshooting issues

## API Stability Guidelines

The Globus Go SDK uses a package-level stability system to clearly communicate API guarantees to users:

### Stability Levels

- **STABLE**: APIs will not change incompatibly within a major version
- **BETA**: APIs may have minor changes in minor releases (with migration guidance)
- **ALPHA**: APIs may change significantly in any release
- **EXPERIMENTAL**: APIs may be completely removed or rewritten without warning

Each package has a `doc.go` file with a clear stability indicator and list of stable components.

### Working with API Stability

When modifying the codebase, follow these guidelines:

1. **Respect stability indicators**: Don't make breaking changes to stable components
2. **Manage deprecations**: Use the deprecation system for any API changes
3. **Document changes**: Update the CHANGELOG with all API modifications
4. **Test compatibility**: Verify API compatibility between versions

## Working with Claude

### Effective Prompting

When asking Claude to help with implementation:

1. **Be specific about the task**: "Help me implement the `RefreshToken` method for the Auth client" is better than "Help with the Auth package."

2. **Provide context**: Share the relevant parts of existing code when asking for new implementations.

3. **Specify requirements**: When asking for implementations, mention specific requirements like error handling, context support, etc.

4. **Request incremental changes**: Break down complex implementations into smaller pieces.

5. **Ask for tests**: Request tests when getting implementations.

6. **Consider API stability**: Specify the stability level of new components and mention any compatibility requirements.

### Example Prompts

Here are some example prompts for working with Claude:

#### Implementing a Specific Method

```
Implement the ExchangeAuthorizationCode method for the Auth client that:
1. Takes a context and authorization code
2. Makes a POST request to the token endpoint
3. Handles errors appropriately
4. Returns a TokenResponse
5. Maintains compatibility with the current stable API
```

#### Updating Existing Functionality

```
Update the connection pool implementation to support connection timeouts while:
1. Maintaining backward compatibility with existing code
2. Following the package's beta stability level guidelines
3. Adding appropriate documentation
4. Including tests for the new functionality
```

#### Writing Tests

```
Write unit tests for the ListGroups method in the Groups client. Include tests for:
1. Successful response with multiple groups
2. Empty response
3. Error handling
4. Context cancellation
5. API contract conformance tests
```

#### Documentation with Stability Indicators

```
Generate GoDoc style documentation for the Auth package, including:
1. Clear STABILITY indicator section
2. List of stable API components
3. Examples of creating a client and using key API methods
4. Notes about any beta or experimental features
```

#### Adding New Functionality

```
Implement resumable transfers support in the Transfer client that:
1. Matches the experimental stability level of this feature
2. Documents the experimental status clearly
3. Provides basic usage examples
4. Adds comprehensive test coverage
```

## Guidelines for Code Review

Ask Claude to review generated code with stability considerations:

```
Please review this implementation of the Groups client and suggest improvements:
1. Check for API stability concerns
2. Verify compatibility with existing code
3. Ensure proper documentation of stability level
4. Identify potential breaking changes

[paste code here]
```

## Best Practices

When working with Claude on this project:

- **Verify all implementations**: Review code for correctness and security
- **Test thoroughly**: Don't rely solely on Claude-generated tests
- **Keep security in mind**: Review authentication implementations carefully
- **Maintain consistency**: Ensure Claude follows the established patterns
- **Respect stability levels**: Don't break compatibility for stable components
- **Document stability**: Include stability indicators in all new packages
- **Think about versioning**: Consider how changes affect semantic versioning
- **Iterate**: Refine implementations over multiple prompts

## Resources

Refer to the following resources when implementing:

- [Globus API Documentation](https://docs.globus.org/api/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Semantic Versioning](https://semver.org/)
- [API Stability Implementation Plan](API_STABILITY_IMPLEMENTATION_PLAN.md)
