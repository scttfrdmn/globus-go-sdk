---
title: "API Reference"
weight: 10
bookCollapseSection: true
---

# Globus Go SDK API Reference

This section provides comprehensive API reference documentation for all services in the Globus Go SDK.

## Services

- [Auth Service]({{< ref "/docs/reference/auth" >}})
  - Authentication, OAuth2 flows, token management
  - Multi-factor authentication support

- [Transfer Service]({{< ref "/docs/reference/transfer" >}})
  - File transfers between Globus endpoints
  - Endpoint management
  - Recursive and resumable transfers

- [Search Service]({{< ref "/docs/reference/search" >}})
  - Index and search data
  - Advanced queries
  - Batch operations

- [Flows Service]({{< ref "/docs/reference/flows" >}})
  - Create and manage automated workflows
  - Execute runs
  - Batch operations

- [Compute Service]({{< ref "/docs/reference/compute" >}})
  - Remote function execution
  - Container management
  - Environment configuration
  - Batch operations

- [Groups Service]({{< ref "/docs/reference/groups" >}})
  - Group management
  - Membership operations
  - Role management

- [Timers Service]({{< ref "/docs/reference/timers" >}})
  - Schedule tasks
  - Timer management
  - Run operations

## Packages

- [Tokens Package]({{< ref "/docs/reference/tokens" >}})
  - Token storage
  - Refresh mechanisms
  - Token management

## Core Components

- Client Initialization
- Configuration Options
- Error Handling
- Logging
- HTTP Transport
- Rate Limiting