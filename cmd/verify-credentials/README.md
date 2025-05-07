<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Globus Credentials Verification Tool

This utility verifies Globus credentials and checks access to various Globus services. It helps you confirm that your credentials are valid and have the necessary permissions before running integration tests.

## Usage

```bash
# From the project root
go run cmd/verify-credentials/main.go

# Or after building
cd cmd/verify-credentials
go build
./verify-credentials
```

## Environment Variables

The tool looks for credentials in a `.env.test` file or in environment variables:

```
GLOBUS_TEST_CLIENT_ID=<your-client-id>
GLOBUS_TEST_CLIENT_SECRET=<your-client-secret>
GLOBUS_TEST_SOURCE_ENDPOINT_ID=<optional-source-endpoint-id>
GLOBUS_TEST_DESTINATION_ENDPOINT_ID=<optional-destination-endpoint-id>
GLOBUS_TEST_GROUP_ID=<optional-group-id>
GLOBUS_TEST_SEARCH_INDEX_ID=<optional-search-index-id>
```

The tool will automatically look for a `.env.test` file in the following locations:
1. `../../.env.test` (when run from `cmd/verify-credentials`)
2. `./.env.test` (when run from project root)
3. `.env.test` (fallback)

## What it Verifies

1. **Authentication**: Confirms that your client ID and secret can be used to obtain a token via the client credentials flow
2. **Token Introspection**: Validates that the obtained token is active
3. **Transfer Service**: If endpoint IDs are provided, checks if they're accessible
4. **Groups Service**: If a group ID is provided, attempts to access it
5. **Search Service**: If a search index ID is provided, attempts to query it

## Implementation Options

This tool provides multiple implementation options:

1. **SDK Implementation** (main.go): Uses the Globus Go SDK to verify credentials and access services. This is the recommended approach as it demonstrates how to use the SDK.

2. **Standalone API Implementation** (verify-credentials-sdk.go): Uses a standalone API client that doesn't rely on the SDK internals. This is useful for troubleshooting credentials even when there are compilation issues in the main SDK.

3. **Completely Independent Implementation** (standalone.go): A completely independent version that can be built separately if other implementations have issues.

All implementations perform similar verification steps and provide the same output. To build and run a specific implementation:

```bash
# For the standalone implementation
go build -o verify-credentials-standalone standalone.go
./verify-credentials-standalone
```

## Understanding the Results

The tool will output detailed information about each verification step, with:

- ✅ Success indicators for each passed check
- ❌ Error indicators with details for failed checks

A successful run will end with a summary message confirming your credentials are valid.

## Troubleshooting

If you encounter errors:

1. **Authentication Failures**: Check that your client ID and secret are correct
2. **Service-Specific Errors**: 
   - Transfer errors may indicate that your credentials don't have transfer scope or the endpoints don't exist
   - Groups errors may indicate that your credentials don't have groups scope or the group doesn't exist
   - Search errors may indicate that your credentials don't have search scope or the index doesn't exist
3. **Scope Issues**: The client credentials flow provides different scopes based on your client registration. If you see "permission denied" errors for specific services, you may need to use a different authentication flow for those services.

## Next Steps

After verifying your credentials, you can:

1. Use these same credentials in your `.env.test` file for running integration tests
2. Configure the Globus Go SDK to use these credentials in your application
3. Investigate any specific permission issues that were identified