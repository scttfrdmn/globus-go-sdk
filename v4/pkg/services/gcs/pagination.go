// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import "context"

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
	client  *CollectionClient
	opts    *ListCollectionsOptions
	nextURL string // empty means first page; set from Links.Next on subsequent pages
	done    bool
}

// HasMore reports whether more pages remain. It is true before any call to
// NextPage, and true after a page whose Links.Next is non-empty.
func (p *CollectionPager) HasMore() bool {
	return !p.done
}

// NextPage fetches the next page of results. After the last page it returns
// (nil, nil) and subsequent calls return ErrNoPagesRemaining.
func (p *CollectionPager) NextPage(ctx context.Context) (*CollectionPage, error) {
	if p.done {
		return nil, ErrNoPagesRemaining
	}

	var page *CollectionPage
	var err error

	if p.nextURL == "" {
		// First page — use normal options
		page, err = p.client.ListCollections(ctx, p.opts)
	} else {
		// Subsequent page — fetch the absolute URL returned in Links.Next.
		// The GCS API embeds full next URLs in the JSON:API links object.
		page, err = p.client.listCollectionsAbsolute(ctx, p.nextURL)
	}
	if err != nil {
		return nil, err
	}

	if page.Links.Next == "" {
		p.done = true
	} else {
		p.nextURL = page.Links.Next
	}

	return page, nil
}

// ErrNoPagesRemaining is returned by NextPage when no more pages exist.
type ErrNoPagesRemainingType struct{}

func (e ErrNoPagesRemainingType) Error() string {
	return "gcs: no more pages remaining"
}

// ErrNoPagesRemaining is the sentinel error returned by CollectionPager.NextPage
// when the caller has already consumed all pages.
var ErrNoPagesRemaining error = ErrNoPagesRemainingType{}
