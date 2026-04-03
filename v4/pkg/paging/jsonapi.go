// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package paging

import "context"

// JSONAPIFetchFn is the fetch callback for JSONAPIPaginator.
// nextURL is empty on the first call (fetch the first page normally).
// The function returns the page items and the URL of the next page
// (empty string means no more pages).
type JSONAPIFetchFn[T any] func(ctx context.Context, nextURL string) (items []T, nextPageURL string, err error)

// JSONAPIPaginator follows JSON:API Links.Next URLs for pagination.
// It is used by the GCS manager API.
type JSONAPIPaginator[T any] struct {
	fetchFn   JSONAPIFetchFn[T]
	nextURL   string
	done      bool
	firstDone bool
}

// NewJSONAPIPaginator creates a JSONAPIPaginator.
// On the first call to NextPage, fetchFn receives an empty nextURL.
// Subsequent calls pass the URL returned by the previous call.
func NewJSONAPIPaginator[T any](fetchFn JSONAPIFetchFn[T]) *JSONAPIPaginator[T] {
	return &JSONAPIPaginator[T]{fetchFn: fetchFn}
}

// HasNext returns true until the server returns an empty next URL.
func (p *JSONAPIPaginator[T]) HasNext() bool {
	if !p.firstDone {
		return true
	}
	return !p.done
}

// NextPage fetches the next page.
func (p *JSONAPIPaginator[T]) NextPage(ctx context.Context) ([]T, error) {
	items, nextURL, err := p.fetchFn(ctx, p.nextURL)
	if err != nil {
		return nil, err
	}
	p.firstDone = true
	p.nextURL = nextURL
	p.done = nextURL == ""
	return items, nil
}
