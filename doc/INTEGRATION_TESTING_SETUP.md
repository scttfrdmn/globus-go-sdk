<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Integration Testing Setup Guide

This document provides step-by-step instructions for setting up Globus API credentials and endpoints to run integration tests for the Globus Go SDK.

## Getting Started

### 1. Create a Globus Account

If you don't already have one, create a Globus account:

1. Go to [https://app.globus.org/](https://app.globus.org/)
2. Click "Log In" and follow the registration process
3. Verify your email and complete the setup

### 2. Create a Globus App Registration

To access Globus APIs, you need to create an app registration:

1. Go to [https://developers.globus.org/](https://developers.globus.org/)
2. Log in with your Globus account
3. Click "Register your app with Globus"
4. Fill in the required fields:
   - App Name: "Globus Go SDK Testing"
   - Contact Email: Your email
   - Redirect URLs: `https://localhost:8000/callback` (for testing)
   - Scopes: Select all applicable scopes (at minimum: openid, profile, email, urn:globus:auth:scope:transfer.api.globus.org:all, urn:globus:auth:scope:groups.api.globus.org:all)
5. Click "Create App"
6. Note your Client ID and Client Secret (will be needed for tests)

### 3. Set Up Endpoints for Transfer Testing

For transfer tests, you need access to at least two endpoints:

#### Option A: Use Personal Endpoints (Recommended for Initial Testing)

1. **Create a Personal Endpoint**:
   - Install Globus Connect Personal from [https://www.globus.org/globus-connect-personal](https://www.globus.org/globus-connect-personal)
   - Follow the setup instructions to create a personal endpoint
   - Note your endpoint ID

2. **Set up Test Directories**:
   - Create a directory named `/globus-test` on your personal endpoint
   - Create a few test files in this directory

3. **Use the Same Endpoint for Both Source and Destination**:
   - For initial testing, you can use the same personal endpoint as both source and destination
   - Just use different subdirectories for source and destination paths

#### Option B: Use Existing Endpoints

If you have access to existing Globus endpoints:

1. **Identify Two Endpoints**:
   - Choose one endpoint as source and one as destination
   - Ensure you have write access to both endpoints
   - Note both endpoint IDs

2. **Create Test Directories**:
   - Create `/globus-test` directories on both endpoints
   - Add some test files to the source endpoint directory

### 4. Configure Test Environment

Create a `.env.test` file in the root of the project:

```
# Required credentials
GLOBUS_TEST_CLIENT_ID=your-client-id
GLOBUS_TEST_CLIENT_SECRET=your-client-secret

# Endpoints for transfer tests
GLOBUS_TEST_SOURCE_ENDPOINT_ID=your-source-endpoint-id
GLOBUS_TEST_DEST_ENDPOINT_ID=your-destination-endpoint-id
GLOBUS_TEST_SOURCE_PATH=/globus-test
GLOBUS_TEST_DEST_PATH=/globus-test

# Your user ID for testing
GLOBUS_TEST_USER_ID=your-user-id

# Optional: A group ID if you have one
GLOBUS_TEST_GROUP_ID=your-group-id
```

### 5. Finding Your User ID

To get your Globus user ID:

1. Go to [https://app.globus.org/account](https://app.globus.org/account)
2. Your user ID is listed as "Username/ID"
3. It will look like `your-name@globusid.org` or a UUID

### 6. Creating a Test Group (Optional)

For testing Groups API functionality:

1. Go to [https://app.globus.org/groups](https://app.globus.org/groups)
2. Click "Create New Group"
3. Fill in the details:
   - Group Name: "Go SDK Test Group"
   - Description: "Test group for Globus Go SDK integration testing"
   - Visibility: Private
4. Click "Create Group"
5. Note the Group ID from the URL: `https://app.globus.org/groups/GROUP_ID/`

## Verifying Your Setup

Run the verification script to check your configuration:

```bash
./scripts/run_integration_tests.sh pkg/integration_test TestIntegration_VerifySetup
```

This will:
1. Verify your credentials are valid
2. Check endpoint accessibility
3. Verify you have the necessary permissions

## Common Issues and Solutions

### Invalid Credentials

**Symptoms**: Authentication errors, "invalid_client" error messages

**Solutions**:
- Double-check client ID and secret for typos
- Ensure the app is still active on the Developers Dashboard
- Create a new client secret if needed

### Endpoint Access Issues

**Symptoms**: "Permission denied" errors, endpoint not found

**Solutions**:
- Verify endpoint IDs are correct
- Ensure endpoints are activated
- Check path permissions on the endpoints
- Make sure test directories exist

### Rate Limiting

**Symptoms**: HTTP 429 errors, "too many requests" messages

**Solutions**:
- Add delays between tests
- Reduce concurrent test execution
- Implement exponential backoff in tests

## Advanced Configuration

### Using Temporary Test Files

To avoid conflicts between test runs:

```go
// Create a unique test directory using a timestamp
testDir := fmt.Sprintf("/globus-test/%s", time.Now().Format("20060102-150405"))

// Make sure to clean up afterward
defer deleteTestDir(client, endpointID, testDir)
```

### Parallel Test Execution

When running tests in parallel:

```go
func TestIntegration_ParallelTests(t *testing.T) {
    // Create subtests
    t.Run("TestA", func(t *testing.T) {
        // This prevents these subtests from running in parallel
        // which could cause conflicts with shared resources
        // t.Parallel() // Uncomment only if tests are isolated
        
        // Test code...
    })
    
    t.Run("TestB", func(t *testing.T) {
        // Test code...
    })
}
```

## Continuous Integration

For CI/CD environments:

1. Store credentials as encrypted secrets
2. Add the secrets as environment variables in your CI config
3. Ensure tests are idempotent and clean up after themselves
4. Consider using a dedicated test account with limited permissions

Example GitHub Actions configuration:

```yaml
name: Integration Tests

on:
  workflow_dispatch:  # Manual trigger only
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sundays

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    
    env:
      GLOBUS_TEST_CLIENT_ID: ${{ secrets.GLOBUS_TEST_CLIENT_ID }}
      GLOBUS_TEST_CLIENT_SECRET: ${{ secrets.GLOBUS_TEST_CLIENT_SECRET }}
      GLOBUS_TEST_SOURCE_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_SOURCE_ENDPOINT_ID }}
      GLOBUS_TEST_DEST_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_DEST_ENDPOINT_ID }}
      GLOBUS_TEST_USER_ID: ${{ secrets.GLOBUS_TEST_USER_ID }}
      GLOBUS_TEST_GROUP_ID: ${{ secrets.GLOBUS_TEST_GROUP_ID }}
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'
          
      - name: Run integration tests
        run: ./scripts/run_integration_tests.sh
```

## Next Steps

After setting up your test environment:

1. Run the basic authentication tests first:
   ```bash
   ./scripts/run_integration_tests.sh pkg/services/auth
   ```

2. Then run transfer tests:
   ```bash
   ./scripts/run_integration_tests.sh pkg/services/transfer
   ```

3. Finally, run the complete test suite:
   ```bash
   ./scripts/run_integration_tests.sh
   ```

4. If any tests fail, check the error messages and refer to the troubleshooting section.

5. Consider creating additional test cases for your specific use cases.