// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package errors provides unified error handling for the Globus Go SDK.

This package implements a consistent error handling system across all Globus
services, following the patterns established by the Python SDK for
compatibility and familiarity. All service-specific errors are represented as
GlobusError values, which carry structured information including service name,
error code, HTTP status, request ID, and retryability.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - GlobusError type and all exported fields (Code, Message, Detail, RequestID,
    Service, HTTPStatus, Context, Timestamp, Retryable, Underlying)
  - GlobusError methods (Error, Unwrap, Is, String, IsRetryable,
    IsAuthenticationError, IsAuthorizationError, IsNotFoundError,
    IsRateLimitError, IsServerError, IsClientError)
  - GlobusError builder methods (WithDetail, WithRequestID, WithContext,
    WithUnderlying, WithRetryable)
  - Constructor functions (NewGlobusError, NewGlobusErrorWithStatus,
    NewGlobusErrorFromHTTPResponse)
  - Service-specific constructors (NewAuthError, NewTransferError,
    NewGroupsError, NewSearchError, NewFlowsError, NewComputeError, NewTimersError)
  - ParseGlobusErrorFromJSON helper
  - All error code constants (AuthInvalidGrant, TransferTaskNotFound, etc.)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New error code constants may be added in minor releases
  - New GlobusError fields will only be added when they do not break existing code
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create a service-specific error:

	err := errors.NewAuthError(errors.AuthInvalidGrant, "the provided refresh token is expired")

Create an error from an HTTP response:

	globusErr := errors.NewGlobusErrorFromHTTPResponse("transfer", resp)

Check error types using standard Go idioms:

	var globusErr *errors.GlobusError
	if stdErrors.As(err, &globusErr) {
		if globusErr.IsRetryable() {
			// Retry the operation
		}
		if globusErr.IsAuthenticationError() {
			// Prompt user to re-authenticate
		}
	}

Build errors with additional context:

	err := errors.NewGlobusError("transfer", errors.TransferPermissionDenied, "access denied").
		WithDetail("you do not have write permission on this endpoint").
		WithRequestID("abc-123").
		WithRetryable(false)
*/
package errors
