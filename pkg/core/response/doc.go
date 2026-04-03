// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package response provides unified response handling for the Globus Go SDK.

This package implements consistent response structures and pagination patterns
across all Globus services, following the patterns established by the Python
SDK for compatibility and familiarity. It provides generic wrappers for both
single-item and paginated responses, along with an iterator abstraction for
consuming paginated result sets without managing pagination tokens directly.

# STABILITY: BETA

This package is in beta. The generic response types require Go 1.18+ and the
API design may be refined in minor releases as usage patterns become clearer.
Changes will be documented in the CHANGELOG with migration guidance.

The following components are considered beta-stable:

  - Response[T] generic type and fields (Data, Metadata, RequestID)
  - Response[T] methods (WithRequestID, WithMetadata)
  - NewResponse[T] constructor
  - PaginatedResponse[T] generic type and fields (Data, Metadata, RequestID, Pagination)
  - PaginatedResponse[T] methods (WithRequestID, WithMetadata, WithPagination,
    IsLastPage, GetNextToken, GetPageSize, GetTotalCount)
  - NewPaginatedResponse[T] constructor
  - ResponseMetadata struct and all fields (APIVersion, Service, Timestamp,
    HTTPStatus, Headers, RateLimit, ExecutionTime)
  - RateLimitInfo struct and all fields (Limit, Remaining, ResetTime, RetryAfter)
  - PaginationInfo struct and all fields (NextToken, HasMore, Limit, Total,
    Offset, Page, PageSize)
  - FromHTTPResponse function
  - Iterator[T] generic type and constructor (NewIterator)
  - Iterator[T] methods (Next, Error, Reset, ToSlice, Count)
  - Service-specific response wrapper types (AuthResponse, TransferResponse,
    GroupsResponse, SearchResponse, FlowsResponse, ComputeResponse, TimersResponse)
  - Service-specific response constructors (NewAuthResponse, NewTransferResponse, etc.)

# Compatibility Guarantees

For beta components:
  - Minor backward-incompatible changes may still occur in minor releases
  - The generic type parameter constraints may be refined
  - New fields may be added to ResponseMetadata and PaginationInfo
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Wrap a service response for uniform handling:

	data := MyServiceResult{}
	resp := response.NewResponse(data, "my-service")

Wrap a paginated response:

	items := []MyItem{}
	paged := response.NewPaginatedResponse(items, "my-service").
		WithPagination(response.PaginationInfo{
			NextToken: "next-page-token",
			HasMore:   true,
			Limit:     25,
			Total:     100,
		})

Iterate over all pages of results:

	fetchPage := func(nextToken string) (*response.PaginatedResponse[MyItem], error) {
		return client.ListItems(ctx, nextToken)
	}

	firstPage, _ := client.ListItems(ctx, "")
	iter := response.NewIterator(firstPage, fetchPage)

	for {
		item, ok := iter.Next()
		if !ok {
			break
		}
		// Process item
	}
	if err := iter.Error(); err != nil {
		// Handle error
	}

Populate metadata from an HTTP response:

	metadata := response.FromHTTPResponse(httpResp, "transfer")
*/
package response
