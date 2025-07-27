<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Multi-Service Workflow Example

This example demonstrates how to orchestrate a complex workflow using multiple Globus services together. It showcases how different services can be combined to build sophisticated data management solutions.

## Overview

The example implements a complete scientific data workflow that:

1. Transfers raw data from a source to a processing location
2. Submits a compute job to analyze the data
3. Indexes the results in Globus Search
4. Creates a flow for future automation of the process
5. Sends notifications when key stages complete

This demonstrates how to integrate Authentication, Transfer, Compute, Search, and Flows services into a cohesive workflow.

## Architecture

The workflow uses the following services:

- **Authentication Service**: For identity and access management
- **Transfer Service**: For moving data between endpoints
- **Compute Service**: For executing analysis tasks
- **Search Service**: For indexing and discovering results
- **Flows Service**: For workflow automation
- **Timers Service**: For scheduled operations

## Prerequisites

Before running this example, you'll need:

1. Globus account with access to Transfer, Compute, Search, and Flows
2. Two Globus endpoints (source and destination)
3. Compute environment configured for your compute tasks
4. The following environment variables:
   ```
   GLOBUS_CLIENT_ID=your_client_id
   GLOBUS_CLIENT_SECRET=your_client_secret
   GLOBUS_SOURCE_ENDPOINT=source_endpoint_id
   GLOBUS_DESTINATION_ENDPOINT=dest_endpoint_id
   GLOBUS_COMPUTE_ENDPOINT=compute_endpoint_id
   GLOBUS_SEARCH_INDEX=search_index_id
   ```

## Components

The example is organized into several components:

- `main.go`: Main program orchestrating the overall workflow
- `auth.go`: Authentication handling and token management
- `transfer.go`: Data transfer operations
- `compute.go`: Compute job submission and monitoring
- `search.go`: Data indexing and search operations
- `flows.go`: Flow definition and creation
- `notification.go`: Notification handling
- `config.go`: Configuration management
- `monitoring.go`: Workflow monitoring and progress tracking

## Usage

To run the example:

```bash
cd examples/multi-service-workflow
go run .
```

With command-line options:

```bash
# Run with custom source data
go run . --source-path /path/to/data

# Run with specific compute function
go run . --compute-function "analyze_data"

# Run with search indexing only
go run . --skip-compute --search-only

# Run with monitoring dashboard
go run . --with-dashboard
```

## Workflow Steps

### 1. Authentication

The workflow begins by authenticating with the Globus Auth service to obtain the necessary tokens for each service.

### 2. Data Transfer

Raw data is transferred from the source endpoint to the destination endpoint where processing will occur.

### 3. Compute Processing

Once data is transferred, a compute job is submitted to process the data. The job's progress is monitored until completion.

### 4. Search Indexing

The results of the compute job are indexed in Globus Search with appropriate metadata for discovery.

### 5. Flow Creation

A Globus Flow is created to automate this process for future executions, with appropriate triggers and actions.

### 6. Notification

Notifications are sent at key points in the workflow (transfer complete, computation complete, indexing complete).

## Error Handling

The example demonstrates robust error handling patterns:

- Retry logic for transient failures
- Graceful degradation for non-critical components
- Comprehensive logging
- Recovery from service interruptions
- Transaction rollback when appropriate

## Extending the Example

To extend this example for your own use cases:

1. Modify `config.go` to set your specific endpoints and parameters
2. Update the compute function in `compute.go` for your analysis needs
3. Adjust the search indexing in `search.go` to capture your relevant metadata
4. Customize the flow definition in `flows.go` for your workflow requirements

## Further Reading

- [Authentication Guide](https://docs.globus.org/developer-tools/go-sdk/auth)
- [Transfer Guide](https://docs.globus.org/developer-tools/go-sdk/transfer)
- [Compute Guide](https://docs.globus.org/developer-tools/go-sdk/compute)
- [Search Guide](https://docs.globus.org/developer-tools/go-sdk/search)
- [Flows Guide](https://docs.globus.org/developer-tools/go-sdk/flows)