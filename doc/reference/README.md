# Globus Go SDK API Reference

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This directory contains comprehensive API reference documentation for all services in the Globus Go SDK.

## Table of Contents

- [Core](#core)
  - [Client Initialization](core/client.md)
  - [Configuration Options](core/config.md)
  - [Error Handling](core/errors.md)
  - [Logging](core/logging.md)
  - [Authorizers](core/authorizers.md)
  - [HTTP Transport](core/transport.md)
  - [Rate Limiting](core/ratelimit.md)

- [Auth Service](#auth-service)
  - [Client](auth/client.md)
  - [OAuth2 Flows](auth/oauth2.md)
  - [Token Validation](auth/token.md)
  - [MFA Support](auth/mfa.md)

- [Tokens Package](#tokens-package)
  - [Manager](tokens/manager.md)
  - [Storage](tokens/storage.md)
  - [Refresh](tokens/refresh.md)
  - [Background Refresh](tokens/background.md)

- [Transfer Service](#transfer-service)
  - [Client](transfer/client.md)
  - [Endpoint Operations](transfer/endpoints.md)
  - [Transfer Operations](transfer/transfers.md)
  - [Recursive Transfers](transfer/recursive.md)
  - [Resumable Transfers](transfer/resumable.md)

- [Search Service](#search-service)
  - [Client](search/client.md)
  - [Indexing](search/indexing.md)
  - [Queries](search/queries.md)
  - [Advanced Queries](search/advanced.md)
  - [Batch Operations](search/batch.md)

- [Flows Service](#flows-service)
  - [Client](flows/client.md)
  - [Flow Operations](flows/flows.md)
  - [Run Operations](flows/runs.md)
  - [Action Providers](flows/actions.md)
  - [Batch Operations](flows/batch.md)

- [Compute Service](#compute-service)
  - [Client](compute/client.md)
  - [Endpoints](compute/endpoints.md)
  - [Functions](compute/functions.md)
  - [Containers](compute/containers.md)
  - [Batch Operations](compute/batch.md)
  - [Workflows](compute/workflows.md)

- [Groups Service](#groups-service)
  - [Client](groups/client.md)
  - [Group Operations](groups/groups.md)
  - [Membership Operations](groups/members.md)
  - [Role Operations](groups/roles.md)

- [Timers Service](#timers-service)
  - [Client](timers/client.md)
  - [Timer Operations](timers/timers.md)
  - [Job Operations](timers/jobs.md)

## Core

The core package provides the foundation for all service clients, including client initialization, configuration, error handling, logging, and HTTP transport.

## Auth Service

The auth service provides functionality for authenticating with Globus Auth, managing OAuth2 flows, and validating tokens.

## Tokens Package

The tokens package provides functionality for storing, retrieving, and refreshing OAuth2 tokens, including automatic background refresh.

## Transfer Service

The transfer service provides functionality for transferring files between Globus endpoints, including recursive and resumable transfers.

## Search Service

The search service provides functionality for indexing and searching data using Globus Search.

## Flows Service

The flows service provides functionality for creating and executing automated workflows.

## Compute Service

The compute service provides functionality for executing functions on remote compute endpoints.

## Groups Service

The groups service provides functionality for managing Globus groups and their memberships.

## Timers Service

The timers service provides functionality for scheduling tasks to run at specific times.