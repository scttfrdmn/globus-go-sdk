# Globus Go SDK API Requirements

## Minimum API Versions

The Globus Go SDK requires specific minimum versions of the Globus APIs:

| Service     | Minimum API Version | Notes                                        |
|-------------|---------------------|----------------------------------------------|
| Transfer    | v0.10               | Explicit endpoint activation not supported   |
| Auth        | v2                  | Modern OAuth 2.0 flows                       |
| Search      | v1.0                | Current index document format                |
| Flows       | v1.0                | Current flow definitions                     |
| Groups      | v1                  | Current group membership model               |
| Compute     | v2                  | Latest compute job submission format         |

## Modern Authentication Requirements

This SDK is designed to work with modern Globus authentication practices:

1. **Properly Scoped Tokens**: All API calls require tokens with appropriate scopes for the requested operations.

2. **Automatic Endpoint Activation**: The Transfer client assumes endpoints support automatic activation with properly scoped tokens, which is standard in Globus endpoints supporting v0.10 and later.

3. **OAuth 2.0 Flows**: The Auth client supports standard OAuth 2.0 flows (authorization code, client credentials) as implemented by Globus Auth v2.

## Important Changes

### Transfer Client

- **Removed Explicit Activation**: `ActivateEndpoint()` and `GetActivationRequirements()` methods have been removed.
- **Path Handling**: Paths without leading slashes are recommended for Guest Collections.
- **Error Handling**: Detailed error types are provided for common error conditions.

### Auth Client

- **Token Management**: Comprehensive token management with refresh, revoke, and introspect support.
- **Scope Support**: Detailed scope handling for each Globus service.

## Compatibility Notes

This SDK is not compatible with legacy Globus endpoints or services that:

1. Require manual activation steps
2. Use pre-OAuth authentication methods
3. Do not support the minimum API versions listed above

For older endpoints, use the REST API directly or the Globus CLI tools.