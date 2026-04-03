// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package paging

import "context"

// Paginator is the generic interface for iterating over paged API results.
// T is the item type returned by each page.
type Paginator[T any] interface {
	// HasNext reports whether more pages remain.
	// It is true before the first call to NextPage.
	HasNext() bool

	// NextPage fetches the next page and returns its items.
	// Returns an empty slice (not an error) on the last page.
	NextPage(ctx context.Context) ([]T, error)
}

// ----------------------------------------------------------------
// LimitOffsetPaginator
// ----------------------------------------------------------------

// LimitOffsetFetchFn is the fetch callback for LimitOffsetPaginator.
// It receives the current limit and offset and returns the page items,
// the total count from the server, and any error.
type LimitOffsetFetchFn[T any] func(ctx context.Context, limit, offset int) (items []T, total int, err error)

// LimitOffsetPaginator iterates using limit + offset query parameters.
// It is the most common pagination strategy across Globus services.
type LimitOffsetPaginator[T any] struct {
	fetchFn   LimitOffsetFetchFn[T]
	pageSize  int
	offset    int
	total     int
	fetched   int
	firstDone bool
}

// NewLimitOffsetPaginator creates a LimitOffsetPaginator.
// pageSize is sent as the `limit` parameter on each request.
// Pass 0 to let the server choose its default page size.
func NewLimitOffsetPaginator[T any](fetchFn LimitOffsetFetchFn[T], pageSize int) *LimitOffsetPaginator[T] {
	return &LimitOffsetPaginator[T]{
		fetchFn:  fetchFn,
		pageSize: pageSize,
	}
}

// HasNext returns true before the first fetch, and true after any fetch where
// fetched < total.
func (p *LimitOffsetPaginator[T]) HasNext() bool {
	if !p.firstDone {
		return true
	}
	return p.fetched < p.total
}

// NextPage fetches the next page. Returns an empty slice when there are no
// more items (after HasNext returns false).
func (p *LimitOffsetPaginator[T]) NextPage(ctx context.Context) ([]T, error) {
	items, total, err := p.fetchFn(ctx, p.pageSize, p.offset)
	if err != nil {
		return nil, err
	}
	p.firstDone = true
	p.total = total
	p.offset += len(items)
	p.fetched += len(items)
	return items, nil
}

// ----------------------------------------------------------------
// MarkerPaginator
// ----------------------------------------------------------------

// MarkerFetchFn is the fetch callback for MarkerPaginator.
// marker is empty on the first call. hasMore indicates whether more pages
// exist after this one. nextMarker is the cursor for the next request.
type MarkerFetchFn[T any] func(ctx context.Context, limit int, marker string) (items []T, hasMore bool, nextMarker string, err error)

// MarkerPaginator iterates using a server-side cursor/marker.
// Used by Transfer tunnel endpoints.
type MarkerPaginator[T any] struct {
	fetchFn   MarkerFetchFn[T]
	pageSize  int
	marker    string
	hasMore   bool
	firstDone bool
}

// NewMarkerPaginator creates a MarkerPaginator.
func NewMarkerPaginator[T any](fetchFn MarkerFetchFn[T], pageSize int) *MarkerPaginator[T] {
	return &MarkerPaginator[T]{
		fetchFn:  fetchFn,
		pageSize: pageSize,
		hasMore:  true,
	}
}

// HasNext returns true until the server signals no more pages.
func (p *MarkerPaginator[T]) HasNext() bool {
	if !p.firstDone {
		return true
	}
	return p.hasMore
}

// NextPage fetches the next page.
func (p *MarkerPaginator[T]) NextPage(ctx context.Context) ([]T, error) {
	items, hasMore, nextMarker, err := p.fetchFn(ctx, p.pageSize, p.marker)
	if err != nil {
		return nil, err
	}
	p.firstDone = true
	p.hasMore = hasMore
	p.marker = nextMarker
	return items, nil
}

// ----------------------------------------------------------------
// NextTokenPaginator
// ----------------------------------------------------------------

// NextTokenFetchFn is the fetch callback for NextTokenPaginator.
// pageToken is empty on the first call. hasNextPage indicates whether more
// pages exist. nextToken is the token for the next request.
type NextTokenFetchFn[T any] func(ctx context.Context, pageSize int, pageToken string) (items []T, hasNextPage bool, nextToken string, err error)

// NextTokenPaginator iterates using a server-provided page token.
// Used by Groups list endpoints.
type NextTokenPaginator[T any] struct {
	fetchFn   NextTokenFetchFn[T]
	pageSize  int
	token     string
	hasNext   bool
	firstDone bool
}

// NewNextTokenPaginator creates a NextTokenPaginator.
func NewNextTokenPaginator[T any](fetchFn NextTokenFetchFn[T], pageSize int) *NextTokenPaginator[T] {
	return &NextTokenPaginator[T]{
		fetchFn:  fetchFn,
		pageSize: pageSize,
		hasNext:  true,
	}
}

// HasNext returns true until the server signals no more pages.
func (p *NextTokenPaginator[T]) HasNext() bool {
	if !p.firstDone {
		return true
	}
	return p.hasNext
}

// NextPage fetches the next page.
func (p *NextTokenPaginator[T]) NextPage(ctx context.Context) ([]T, error) {
	items, hasNext, nextToken, err := p.fetchFn(ctx, p.pageSize, p.token)
	if err != nil {
		return nil, err
	}
	p.firstDone = true
	p.hasNext = hasNext
	p.token = nextToken
	return items, nil
}
