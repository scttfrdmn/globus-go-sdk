<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Import Cycle Resolution

This document details the approach taken to resolve import cycles in the Globus Go SDK, a common issue in Go projects with complex package relationships.

## Problem Description

The SDK was experiencing import cycle issues between several critical packages:

1. **Primary Circular Dependency**: 
   - `pkg/core/transport` imported `pkg/core`
   - `pkg/core` needed to use `pkg/core/transport` functionality

2. **Additional Circular Dependencies**:
   - Circular dependencies involving `pkg/core/client.go`, `pkg/core/http`, and `pkg/core/pool`
   - Duplicated pool implementation code between `pkg/core/transport/connection_pool.go` and `pkg/core/http/pool.go`

These import cycles prevented the project from compiling correctly and created obstacles for integration testing and overall code maintenance.

## Solution Approach

The solution follows best practices for resolving import cycles in Go, primarily using the **interface extraction pattern**:

1. **Interface Extraction**: 
   - Created a dedicated `pkg/core/interfaces` package to define key interfaces without importing implementations
   - Moved interface definitions to this package to break circular dependencies
   
2. **Implementation of Interfaces**: 
   - Updated existing structs to implement the new interfaces
   - Added adapter files to confirm implementation of interfaces using compile-time checks
   
3. **Dependency Inversion**: 
   - Modified code to depend on interfaces rather than concrete implementations
   - Used dependency injection to provide concrete implementations at runtime

## Files Created

1. **Interface Definitions**:
   - `/pkg/core/interfaces/transport.go` - Transport interfaces for HTTP communication
   - `/pkg/core/interfaces/client.go` - Client interfaces for base functionality
   - `/pkg/core/interfaces/pool.go` - Connection pool interfaces
   - `/pkg/core/interfaces/auth.go` - Authorization interfaces

2. **Interface Implementation Verification**:
   - `/pkg/core/client_interface.go` - Ensures Client implements ClientInterface
   - `/pkg/core/logger_interface.go` - Ensures DefaultLogger implements Logger interface
   - `/pkg/core/auth/auth_interface.go` - Ensures Authorizer implements interfaces.Authorizer
   - `/pkg/core/transport/transport_interface.go` - Ensures Transport implements interfaces.Transport
   - `/pkg/core/transport/pool_interface.go` - Ensures ConnectionPool implements interfaces.ConnectionPool
   - `/pkg/core/http/pool_interface.go` - Ensures HTTP implementations implement interfaces

## Files Modified

1. `/pkg/core/transport/transport.go`:
   - Changed import from `pkg/core` to `pkg/core/interfaces`
   - Modified Transport struct to use ClientInterface instead of Client
   - Updated method parameters and function calls

2. `/pkg/core/client.go`:
   - Added import for `pkg/core/interfaces`
   - Updated Logger field to use interfaces.Logger type

3. `/pkg/core/pool/client.go`:
   - Added import for interfaces package

## Benefits of This Approach

1. **Elimination of Import Cycles**: The primary goal was achieved by breaking circular dependencies
2. **Better Architecture**: Improved code organization with clear separation between interfaces and implementations
3. **Enhanced Testability**: Interfaces make mocking and testing significantly easier
4. **Clearer Dependencies**: Package dependencies are now more explicit and easier to understand
5. **Future Extensibility**: New implementations can be added without modifying existing code

## Preventing Future Import Cycles

To prevent import cycles from recurring as the SDK evolves:

1. **Keep Interface Definitions Separate**: Continue to define interfaces in dedicated packages
2. **Follow Dependency Inversion**: Depend on abstractions (interfaces), not concrete implementations
3. **Use Compile-Time Checks**: Maintain type assertion variables to ensure implementations satisfy interfaces
4. **Pay Attention to Package Structure**: Be mindful of package dependencies when adding new functionality
5. **Consider Dependency Graphs**: Periodically review the project's dependency graph to identify potential issues

## Conclusion

By extracting interfaces to a dedicated package and implementing the adapter pattern, we've successfully resolved the import cycles in the Globus Go SDK. This architectural improvement not only fixes the immediate compilation issues but also enhances the overall maintainability and extensibility of the codebase.

This approach aligns with Go best practices and principles of clean architecture, ensuring a more robust foundation for the SDK's continued development.