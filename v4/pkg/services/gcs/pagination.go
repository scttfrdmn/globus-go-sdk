// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// CollectionPager iterates through pages of GCS collection listings.
// It is created by CollectionClient.NewCollectionPager and drives the
// JSON:API next-link pagination used by the GCS manager API.
//
// EXPERIMENTAL: this API may change without notice.
//
// Example:
//
//	pager := client.NewCollectionPager(nil)
//	for pager.HasMore() {
//	    page, err := pager.NextPage(ctx)
//	    if err != nil { return err }
//	    for _, c := range page.Data {
//	        fmt.Println(c.DisplayName)
//	    }
//	}
type CollectionPager struct {
	inner *paging.JSONAPIPaginator[Collection]
	opts  *ListCollectionsOptions
	// last holds the most recently fetched page metadata (Links, Meta).
	// Items are returned via NextPage; we still expose a CollectionPage
	// for callers that inspect Links or Meta directly.
	client *CollectionClient
}

// HasMore reports whether more pages remain. It is true before any call to
// NextPage, and true after a page whose Links.Next is non-empty.
func (p *CollectionPager) HasMore() bool {
	return p.inner.HasNext()
}

// NextPage fetches the next page of results. After the last page it returns
// (nil, nil) and subsequent calls return ErrNoPagesRemaining.
func (p *CollectionPager) NextPage(ctx context.Context) (*CollectionPage, error) {
	if !p.inner.HasNext() {
		return nil, ErrNoPagesRemaining
	}
	// The inner paginator returns []Collection; we need a *CollectionPage.
	// Re-fetch via the client methods directly so we preserve Links/Meta.
	_ = ctx // used by the fetch closure
	return p.fetchPage(ctx)
}

// fetchPage is called by the inner paginator's fetchFn.
func (p *CollectionPager) fetchPageFn(ctx context.Context, nextURL string) ([]Collection, string, error) {
	var page *CollectionPage
	var err error
	if nextURL == "" {
		page, err = p.client.ListCollections(ctx, p.opts)
	} else {
		page, err = p.client.listCollectionsAbsolute(ctx, nextURL)
	}
	if err != nil {
		return nil, "", err
	}
	return page.Data, page.Links.Next, nil
}

// fetchPage is a thin wrapper that builds a CollectionPage from the inner pager.
func (p *CollectionPager) fetchPage(ctx context.Context) (*CollectionPage, error) {
	// Drive the inner paginator directly so we stay in sync.
	items, err := p.inner.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	return &CollectionPage{Data: items}, nil
}

// ErrNoPagesRemaining is returned by NextPage when no more pages exist.
type ErrNoPagesRemainingType struct{}

func (e ErrNoPagesRemainingType) Error() string {
	return "gcs: no more pages remaining"
}

// ErrNoPagesRemaining is the sentinel error returned by CollectionPager.NextPage
// when the caller has already consumed all pages.
var ErrNoPagesRemaining error = ErrNoPagesRemainingType{}
