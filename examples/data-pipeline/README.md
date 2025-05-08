# Data Pipeline Example

This example demonstrates how to build a comprehensive data pipeline using the Globus Go SDK, showcasing:

1. **Transfer Service** - For moving data between endpoints with resumable capabilities
2. **Search Service** - For indexing and cataloging the transferred data
3. **Compute Service** - For processing the data after transfer
4. **Flows Service** - For orchestrating the entire pipeline as a repeatable workflow

## Features

- Resumable transfers with checkpointing
- Progress monitoring and metrics collection
- Automatic token refreshing
- Robust error handling with retries
- Environment-based configuration

## Usage

```bash
# Set required environment variables
export GLOBUS_CLIENT_ID=your_client_id
export GLOBUS_CLIENT_SECRET=your_client_secret
export SOURCE_ENDPOINT_ID=source_endpoint_id
export SOURCE_PATH=/path/on/source
export DESTINATION_ENDPOINT_ID=destination_endpoint_id
export DESTINATION_PATH=/path/on/destination
export SEARCH_INDEX_ID=search_index_id
export COMPUTE_ENDPOINT_ID=compute_endpoint_id
export CONTAINER_IMAGE=python:3.9-slim

# Optional
export FLOW_ID=existing_flow_id  # Only if using an existing flow

# Run the example
go run main.go
```

## Architecture

The pipeline implements a common data processing workflow:

1. Transfer data from source to destination endpoint
2. Index the transferred data in Search for discovery
3. Process the data using Compute functions
4. Create a Flow to automate this pipeline for future runs

## Error Handling

The example demonstrates best practices for error handling:
- Exponential backoff for retries
- Checkpointing for resumable operations
- Graceful shutdown with signal handling
- Detailed logging and metrics

## Next Steps

This example can be extended to:
- Add email notifications using Timers service
- Implement more complex data transformation in Compute
- Add a web UI for monitoring using the metrics collected
- Integrate with other systems using webhooks