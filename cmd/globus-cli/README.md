<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus CLI Example

This is an example command-line interface for the Globus Go SDK, demonstrating how to use the SDK to build a functional CLI application.

## Features

- Authentication with Globus Auth (login/logout)
- Token management and refresh
- File listing on Globus endpoints
- File transfer between endpoints
- Recursive directory transfers
- Transfer status monitoring

## Building

```bash
cd cmd/globus-cli
go build -o globus-cli
```

## Usage

```bash
# Login to Globus
./globus-cli login

# List files on an endpoint
./globus-cli ls <endpoint-id> <path>

# Transfer a file between endpoints
./globus-cli transfer <source-endpoint-id> <source-path> <dest-endpoint-id> <dest-path>

# Transfer a directory recursively
./globus-cli transfer <source-endpoint-id> <source-path> <dest-endpoint-id> <dest-path> --recursive

# Check transfer status
./globus-cli status <task-id>

# View current token information
./globus-cli token

# Log out
./globus-cli logout
```

## Configuration

The CLI stores its configuration and tokens in `~/.globus-cli/`:

- `~/.globus-cli/config.json`: General configuration
- `~/.globus-cli/tokens/`: OAuth2 tokens
- `~/.globus-cli/last-task-id`: ID of the last transfer task

You can modify the `config.json` file to use your own client ID and secret if needed.

## Implementation Details

This CLI example demonstrates several SDK features:

1. **Authentication**: Uses the SDK's auth client to perform OAuth2 flows and manage tokens.
2. **File Operations**: Shows how to list files and directories on Globus endpoints.
3. **Transfer**: Demonstrates both simple file transfers and recursive directory transfers.
4. **Task Management**: Shows how to monitor and display transfer task status.
5. **Token Storage**: Implements a simple token storage mechanism for saved sessions.

## Extension Points

This example can be extended in several ways:

- Add support for additional Globus services (Groups, Search, etc.)
- Implement more advanced transfer options
- Add support for multiple saved endpoints
- Implement automatic token refresh
- Add support for batch operations