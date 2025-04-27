<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Resumable Transfers Example

This example demonstrates how to use the resumable transfer functionality in the Globus Go SDK to perform large transfers that can be paused and resumed.

## Features

- **Checkpointing**: Tracks transfer progress with persistent checkpoints
- **Resumable Transfers**: Ability to stop and resume transfers
- **Batch Processing**: Transfers files in configurable batch sizes
- **Progress Reporting**: Real-time progress updates
- **Error Handling**: Tracks and reports failed transfers
- **Graceful Shutdown**: Handles interruptions gracefully

## Usage

### Prerequisites

1. A valid Globus access token with transfer scope
2. Source and destination endpoint IDs

### Running the Example

Start a new transfer:

```bash
go run main.go \
  --source "source-endpoint-id" \
  --source-path "/path/on/source" \
  --dest "destination-endpoint-id" \
  --dest-path "/path/on/destination" \
  --batch-size 50
```

Resume an existing transfer:

```bash
go run main.go --resume "checkpoint-id"
```

List available checkpoints:

```bash
go run main.go --list
```

Cancel a transfer:

```bash
go run main.go --cancel "checkpoint-id"
```

### Command Line Options

| Option | Description |
|--------|-------------|
| `--source` | Source endpoint ID |
| `--source-path` | Path on the source endpoint |
| `--dest` | Destination endpoint ID |
| `--dest-path` | Path on the destination endpoint |
| `--token` | Globus access token (optional, can use GLOBUS_ACCESS_TOKEN env var) |
| `--resume` | Checkpoint ID to resume |
| `--batch-size` | Number of files to include in each transfer batch (default: 100) |
| `--list` | List available checkpoints |
| `--cancel` | Cancel a transfer with the given checkpoint ID |

## How It Works

1. **File Discovery**: When starting a new transfer, the application scans the source directory to identify all files that need to be transferred.

2. **Checkpointing**: A checkpoint file is created with details about the transfer, including pending items, completed items, and failed items.

3. **Batch Processing**: Files are transferred in batches, with each batch containing a configurable number of files. This improves performance and allows for better error recovery.

4. **Progress Tracking**: As files are transferred, the application updates the checkpoint file with progress information.

5. **Resume Capability**: If the transfer is interrupted, it can be resumed from the last saved checkpoint.

## Error Handling

- Failed transfers are tracked and reported
- The application can be interrupted with Ctrl+C and will gracefully save its state
- Transfers can be resumed multiple times until complete

## Implementation Details

The resumable transfer functionality is implemented in the following files:

- `pkg/services/transfer/checkpoint.go`: Checkpoint data structures and storage
- `pkg/services/transfer/resumable.go`: Resumable transfer implementation
- `pkg/services/transfer/client.go`: Client methods for resumable transfers