// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package paging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- LimitOffsetPaginator ----

func TestLimitOffsetPaginator_SinglePage(t *testing.T) {
	ctx := context.Background()
	items := []int{1, 2, 3}

	p := paging.NewLimitOffsetPaginator(func(_ context.Context, limit, offset int) ([]int, int, error) {
		return items[offset:], len(items), nil
	}, 10)

	assert.True(t, p.HasNext())
	got, err := p.NextPage(ctx)
	require.NoError(t, err)
	assert.Equal(t, items, got)
	assert.False(t, p.HasNext())
}

func TestLimitOffsetPaginator_MultiplePages(t *testing.T) {
	ctx := context.Background()
	all := []int{1, 2, 3, 4, 5}

	calls := 0
	p := paging.NewLimitOffsetPaginator(func(_ context.Context, limit, offset int) ([]int, int, error) {
		calls++
		end := offset + limit
		if end > len(all) {
			end = len(all)
		}
		return all[offset:end], len(all), nil
	}, 2)

	var collected []int
	for p.HasNext() {
		page, err := p.NextPage(ctx)
		require.NoError(t, err)
		collected = append(collected, page...)
	}

	assert.Equal(t, all, collected)
	assert.Equal(t, 3, calls) // pages: [1,2], [3,4], [5]
}

func TestLimitOffsetPaginator_Error(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("oops")

	p := paging.NewLimitOffsetPaginator(func(_ context.Context, _, _ int) ([]int, int, error) {
		return nil, 0, sentinel
	}, 10)

	_, err := p.NextPage(ctx)
	assert.ErrorIs(t, err, sentinel)
}

// ---- MarkerPaginator ----

func TestMarkerPaginator_MultiplePages(t *testing.T) {
	ctx := context.Background()
	pages := []struct {
		items   []string
		hasMore bool
		marker  string
	}{
		{[]string{"a", "b"}, true, "cursor1"},
		{[]string{"c", "d"}, true, "cursor2"},
		{[]string{"e"}, false, ""},
	}
	idx := 0

	p := paging.NewMarkerPaginator(func(_ context.Context, limit int, marker string) ([]string, bool, string, error) {
		page := pages[idx]
		idx++
		return page.items, page.hasMore, page.marker, nil
	}, 2)

	var collected []string
	for p.HasNext() {
		got, err := p.NextPage(ctx)
		require.NoError(t, err)
		collected = append(collected, got...)
	}
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, collected)
}

// ---- NextTokenPaginator ----

func TestNextTokenPaginator_MultiplePages(t *testing.T) {
	ctx := context.Background()
	pages := []struct {
		items    []string
		hasNext  bool
		token    string
	}{
		{[]string{"x", "y"}, true, "tok1"},
		{[]string{"z"}, false, ""},
	}
	idx := 0

	p := paging.NewNextTokenPaginator(func(_ context.Context, size int, token string) ([]string, bool, string, error) {
		page := pages[idx]
		idx++
		return page.items, page.hasNext, page.token, nil
	}, 2)

	var collected []string
	for p.HasNext() {
		got, err := p.NextPage(ctx)
		require.NoError(t, err)
		collected = append(collected, got...)
	}
	assert.Equal(t, []string{"x", "y", "z"}, collected)
}

// ---- JSONAPIPaginator ----

func TestJSONAPIPaginator_MultiplePages(t *testing.T) {
	ctx := context.Background()
	responses := map[string]struct {
		items   []int
		nextURL string
	}{
		"":      {[]int{1, 2}, "page2"},
		"page2": {[]int{3, 4}, "page3"},
		"page3": {[]int{5}, ""},
	}

	p := paging.NewJSONAPIPaginator(func(_ context.Context, nextURL string) ([]int, string, error) {
		r := responses[nextURL]
		return r.items, r.nextURL, nil
	})

	var collected []int
	for p.HasNext() {
		got, err := p.NextPage(ctx)
		require.NoError(t, err)
		collected = append(collected, got...)
	}
	assert.Equal(t, []int{1, 2, 3, 4, 5}, collected)
}

func TestJSONAPIPaginator_SinglePage(t *testing.T) {
	ctx := context.Background()

	p := paging.NewJSONAPIPaginator(func(_ context.Context, _ string) ([]int, string, error) {
		return []int{42}, "", nil
	})

	assert.True(t, p.HasNext())
	got, err := p.NextPage(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int{42}, got)
	assert.False(t, p.HasNext())
}
