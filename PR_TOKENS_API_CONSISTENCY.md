# API Consistency Implementation for Tokens Package

## Summary

This PR implements the functional options pattern for the tokens package to ensure consistent API design across all service clients in the SDK. These changes maintain backward compatibility while providing a more flexible, consistent interface for configuring token managers.

## Changes

1. **New Options Implementation**:
   - Created `options.go` in the tokens package
   - Implemented `ClientOption` type for configuration
   - Added multiple option functions like `WithStorage`, `WithRefreshHandler`, `WithAuthClient`
   - Implemented default options for ease of use

2. **Manager Updates**:
   - Modified the `NewManager` constructor to use the functional options pattern
   - Added proper error handling and validation
   - Ensured backward compatibility with sensible defaults

3. **SDK Integration**:
   - Updated `pkg/globus.go` with new methods:
     - `NewTokenManager`: Creates a token manager with custom options 
     - `NewTokenManagerWithAuth`: Creates a token manager with an auth client for refreshing
   - Added TokensScope constant to maintain consistency with other services
   - Improved error handling throughout

4. **Example Updates**:
   - Updated `examples/token-management/main.go` to use the new API
   - Updated `examples/token-management/mock.go` to use the functional options
   - Improved error handling in examples

5. **Documentation**:
   - Updated `TOKENS_PACKAGE_SUMMARY.md` with API consistency improvements
   - Added detailed documentation to all new option functions

## Testing

All existing functionality has been preserved with these changes, and the API is now consistent with other service clients. The examples have been updated to demonstrate proper usage of the new API.

## Benefits

1. **Consistent API** across all service clients for better developer experience
2. **Improved Error Handling** with proper error propagation from constructors
3. **More Flexible Configuration** with typed option functions
4. **Better SDK Integration** through standardized patterns
5. **Simplified Examples** demonstrating best practices

## Breaking Changes

None. The API changes are fully backward compatible through helper methods and sensible defaults.

## Future Work

Consider adding more option functions for additional configuration flexibility, such as customizing HTTP transport settings and timeout configurations.