// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package response provides unified response handling for the Globus Go SDK.
//
// This package implements consistent response structures and pagination
// patterns across all Globus services, following the patterns established
// by the Python SDK for compatibility and familiarity.
package response

import (
	"net/http"
	"strconv"
	"time"
)

// Response is a generic wrapper for all Globus service responses
type Response[T any] struct {
	// Data contains the actual response data
	Data T `json:"data"`

	// Metadata contains response metadata
	Metadata ResponseMetadata `json:"metadata"`

	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	// APIVersion is the API version used for this request
	APIVersion string `json:"api_version"`

	// Service is the name of the service that generated this response
	Service string `json:"service"`

	// Timestamp is when the response was generated
	Timestamp time.Time `json:"timestamp"`

	// HTTPStatus is the HTTP status code
	HTTPStatus int `json:"http_status"`

	// Headers contains relevant HTTP headers
	Headers map[string]string `json:"headers,omitempty"`

	// RateLimit contains rate limiting information
	RateLimit *RateLimitInfo `json:"rate_limit,omitempty"`

	// ExecutionTime is how long the request took to process
	ExecutionTime time.Duration `json:"execution_time,omitempty"`
}

// RateLimitInfo contains rate limiting information
type RateLimitInfo struct {
	// Limit is the maximum number of requests allowed
	Limit int `json:"limit"`

	// Remaining is the number of requests remaining in the current window
	Remaining int `json:"remaining"`

	// ResetTime is when the rate limit window resets
	ResetTime time.Time `json:"reset_time"`

	// RetryAfter is the number of seconds to wait before retrying
	RetryAfter int `json:"retry_after,omitempty"`
}

// PaginatedResponse is a generic wrapper for paginated responses
type PaginatedResponse[T any] struct {
	// Data contains the array of items
	Data []T `json:"data"`

	// Metadata contains response metadata
	Metadata ResponseMetadata `json:"metadata"`

	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id"`

	// Pagination contains pagination information
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	// NextToken is the token to use for the next page
	NextToken string `json:"next_token,omitempty"`

	// HasMore indicates if there are more pages available
	HasMore bool `json:"has_more"`

	// Limit is the maximum number of items per page
	Limit int `json:"limit"`

	// Total is the total number of items (if known)
	Total int `json:"total,omitempty"`

	// Offset is the starting offset for this page
	Offset int `json:"offset,omitempty"`

	// Page is the current page number (1-based)
	Page int `json:"page,omitempty"`

	// PageSize is the number of items in this page
	PageSize int `json:"page_size"`
}

// NewResponse creates a new Response with the given data
func NewResponse[T any](data T, service string) *Response[T] {
	return &Response[T]{
		Data: data,
		Metadata: ResponseMetadata{
			Service:   service,
			Timestamp: time.Now(),
		},
	}
}

// NewPaginatedResponse creates a new PaginatedResponse with the given data
func NewPaginatedResponse[T any](data []T, service string) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Data: data,
		Metadata: ResponseMetadata{
			Service:   service,
			Timestamp: time.Now(),
		},
		Pagination: PaginationInfo{
			PageSize: len(data),
		},
	}
}

// FromHTTPResponse creates response metadata from an HTTP response
func FromHTTPResponse(resp *http.Response, service string) ResponseMetadata {
	metadata := ResponseMetadata{
		Service:    service,
		HTTPStatus: resp.StatusCode,
		Timestamp:  time.Now(),
		Headers:    make(map[string]string),
	}

	// Extract important headers
	if apiVersion := resp.Header.Get("X-API-Version"); apiVersion != "" {
		metadata.APIVersion = apiVersion
	}

	// Extract rate limit information
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
			if resetTime := resp.Header.Get("X-RateLimit-Reset"); resetTime != "" {
				metadata.RateLimit = &RateLimitInfo{}

				if l, err := strconv.Atoi(limit); err == nil {
					metadata.RateLimit.Limit = l
				}

				if r, err := strconv.Atoi(remaining); err == nil {
					metadata.RateLimit.Remaining = r
				}

				if rt, err := strconv.ParseInt(resetTime, 10, 64); err == nil {
					metadata.RateLimit.ResetTime = time.Unix(rt, 0)
				}

				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if ra, err := strconv.Atoi(retryAfter); err == nil {
						metadata.RateLimit.RetryAfter = ra
					}
				}
			}
		}
	}

	// Store other relevant headers
	relevantHeaders := []string{
		"X-Request-Id",
		"X-Correlation-Id",
		"X-Response-Time",
		"Cache-Control",
		"ETag",
		"Last-Modified",
	}

	for _, header := range relevantHeaders {
		if value := resp.Header.Get(header); value != "" {
			metadata.Headers[header] = value
		}
	}

	return metadata
}

// WithRequestID adds a request ID to the response
func (r *Response[T]) WithRequestID(requestID string) *Response[T] {
	r.RequestID = requestID
	return r
}

// WithMetadata sets the response metadata
func (r *Response[T]) WithMetadata(metadata ResponseMetadata) *Response[T] {
	r.Metadata = metadata
	return r
}

// WithRequestID adds a request ID to the paginated response
func (r *PaginatedResponse[T]) WithRequestID(requestID string) *PaginatedResponse[T] {
	r.RequestID = requestID
	return r
}

// WithMetadata sets the response metadata
func (r *PaginatedResponse[T]) WithMetadata(metadata ResponseMetadata) *PaginatedResponse[T] {
	r.Metadata = metadata
	return r
}

// WithPagination sets the pagination information
func (r *PaginatedResponse[T]) WithPagination(pagination PaginationInfo) *PaginatedResponse[T] {
	r.Pagination = pagination
	return r
}

// IsLastPage returns true if this is the last page of results
func (r *PaginatedResponse[T]) IsLastPage() bool {
	return !r.Pagination.HasMore
}

// GetNextToken returns the token for the next page
func (r *PaginatedResponse[T]) GetNextToken() string {
	return r.Pagination.NextToken
}

// GetPageSize returns the number of items in this page
func (r *PaginatedResponse[T]) GetPageSize() int {
	return len(r.Data)
}

// GetTotalCount returns the total number of items (if known)
func (r *PaginatedResponse[T]) GetTotalCount() int {
	return r.Pagination.Total
}

// Iterator provides a way to iterate through paginated results
type Iterator[T any] struct {
	// current holds the current page of results
	current *PaginatedResponse[T]

	// index is the current index within the current page
	index int

	// fetchNext is a function to fetch the next page
	fetchNext func(nextToken string) (*PaginatedResponse[T], error)

	// err holds any error that occurred during iteration
	err error
}

// NewIterator creates a new iterator for paginated results
func NewIterator[T any](
	initial *PaginatedResponse[T],
	fetchNext func(nextToken string) (*PaginatedResponse[T], error),
) *Iterator[T] {
	return &Iterator[T]{
		current:   initial,
		index:     0,
		fetchNext: fetchNext,
	}
}

// Next returns the next item in the iteration
func (i *Iterator[T]) Next() (T, bool) {
	var zero T

	// Check if we have an error
	if i.err != nil {
		return zero, false
	}

	// Check if we need to fetch the next page
	if i.current == nil || i.index >= len(i.current.Data) {
		if i.current == nil || !i.current.Pagination.HasMore {
			return zero, false
		}

		// Fetch the next page
		nextPage, err := i.fetchNext(i.current.Pagination.NextToken)
		if err != nil {
			i.err = err
			return zero, false
		}

		i.current = nextPage
		i.index = 0
	}

	// Return the current item
	if i.index < len(i.current.Data) {
		item := i.current.Data[i.index]
		i.index++
		return item, true
	}

	return zero, false
}

// Error returns any error that occurred during iteration
func (i *Iterator[T]) Error() error {
	return i.err
}

// Reset resets the iterator to the beginning
func (i *Iterator[T]) Reset() {
	i.index = 0
	i.err = nil
}

// ToSlice converts the iterator to a slice by consuming all remaining items
func (i *Iterator[T]) ToSlice() ([]T, error) {
	var results []T

	for {
		item, ok := i.Next()
		if !ok {
			break
		}
		results = append(results, item)
	}

	return results, i.err
}

// Count returns the total number of items by consuming the iterator
func (i *Iterator[T]) Count() (int, error) {
	count := 0

	for {
		_, ok := i.Next()
		if !ok {
			break
		}
		count++
	}

	return count, i.err
}

// Service-specific response types (examples)

// AuthResponse is a response wrapper for Auth service
type AuthResponse[T any] struct {
	*Response[T]
}

// TransferResponse is a response wrapper for Transfer service
type TransferResponse[T any] struct {
	*Response[T]
}

// GroupsResponse is a response wrapper for Groups service
type GroupsResponse[T any] struct {
	*Response[T]
}

// SearchResponse is a response wrapper for Search service
type SearchResponse[T any] struct {
	*Response[T]
}

// FlowsResponse is a response wrapper for Flows service
type FlowsResponse[T any] struct {
	*Response[T]
}

// ComputeResponse is a response wrapper for Compute service
type ComputeResponse[T any] struct {
	*Response[T]
}

// TimersResponse is a response wrapper for Timers service
type TimersResponse[T any] struct {
	*Response[T]
}

// Helper functions for creating service-specific responses

// NewAuthResponse creates a new Auth service response
func NewAuthResponse[T any](data T) *AuthResponse[T] {
	return &AuthResponse[T]{
		Response: NewResponse(data, "auth"),
	}
}

// NewTransferResponse creates a new Transfer service response
func NewTransferResponse[T any](data T) *TransferResponse[T] {
	return &TransferResponse[T]{
		Response: NewResponse(data, "transfer"),
	}
}

// NewGroupsResponse creates a new Groups service response
func NewGroupsResponse[T any](data T) *GroupsResponse[T] {
	return &GroupsResponse[T]{
		Response: NewResponse(data, "groups"),
	}
}

// NewSearchResponse creates a new Search service response
func NewSearchResponse[T any](data T) *SearchResponse[T] {
	return &SearchResponse[T]{
		Response: NewResponse(data, "search"),
	}
}

// NewFlowsResponse creates a new Flows service response
func NewFlowsResponse[T any](data T) *FlowsResponse[T] {
	return &FlowsResponse[T]{
		Response: NewResponse(data, "flows"),
	}
}

// NewComputeResponse creates a new Compute service response
func NewComputeResponse[T any](data T) *ComputeResponse[T] {
	return &ComputeResponse[T]{
		Response: NewResponse(data, "compute"),
	}
}

// NewTimersResponse creates a new Timers service response
func NewTimersResponse[T any](data T) *TimersResponse[T] {
	return &TimersResponse[T]{
		Response: NewResponse(data, "timers"),
	}
}
