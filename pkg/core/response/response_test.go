// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package response_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/response"
)

// ---------------------------------------------------------------------------
// Helper types used across tests
// ---------------------------------------------------------------------------

type sampleData struct {
	ID   int
	Name string
}

// ---------------------------------------------------------------------------
// NewResponse[T]
// ---------------------------------------------------------------------------

func TestNewResponse_DataField(t *testing.T) {
	data := sampleData{ID: 1, Name: "alice"}
	resp := response.NewResponse(data, "test-service")

	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Data.ID != 1 || resp.Data.Name != "alice" {
		t.Errorf("unexpected Data: got %+v", resp.Data)
	}
}

func TestNewResponse_MetadataService(t *testing.T) {
	resp := response.NewResponse("hello", "auth")

	if resp.Metadata.Service != "auth" {
		t.Errorf("expected Service 'auth', got %q", resp.Metadata.Service)
	}
}

func TestNewResponse_MetadataTimestampIsRecent(t *testing.T) {
	before := time.Now()
	resp := response.NewResponse(42, "svc")
	after := time.Now()

	if resp.Metadata.Timestamp.Before(before) || resp.Metadata.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in range [%v, %v]",
			resp.Metadata.Timestamp, before, after)
	}
}

func TestNewResponse_RequestIDEmptyByDefault(t *testing.T) {
	resp := response.NewResponse(true, "svc")
	if resp.RequestID != "" {
		t.Errorf("expected empty RequestID, got %q", resp.RequestID)
	}
}

func TestNewResponse_WithStringData(t *testing.T) {
	resp := response.NewResponse("payload", "transfer")
	if resp.Data != "payload" {
		t.Errorf("unexpected Data: %v", resp.Data)
	}
}

// ---------------------------------------------------------------------------
// NewPaginatedResponse[T]
// ---------------------------------------------------------------------------

func TestNewPaginatedResponse_DataSlice(t *testing.T) {
	items := []sampleData{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
		{ID: 3, Name: "c"},
	}
	resp := response.NewPaginatedResponse(items, "groups")

	if resp == nil {
		t.Fatal("expected non-nil paginated response")
	}
	if len(resp.Data) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Data))
	}
	for idx, item := range items {
		if resp.Data[idx] != item {
			t.Errorf("item[%d]: got %+v, want %+v", idx, resp.Data[idx], item)
		}
	}
}

func TestNewPaginatedResponse_PageSizeMatchesData(t *testing.T) {
	items := []int{10, 20, 30, 40, 50}
	resp := response.NewPaginatedResponse(items, "search")

	if resp.Pagination.PageSize != len(items) {
		t.Errorf("expected PageSize %d, got %d", len(items), resp.Pagination.PageSize)
	}
}

func TestNewPaginatedResponse_EmptySlice(t *testing.T) {
	resp := response.NewPaginatedResponse([]string{}, "flows")

	if resp.Data == nil {
		t.Fatal("expected non-nil Data slice")
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Data))
	}
	if resp.Pagination.PageSize != 0 {
		t.Errorf("expected PageSize 0, got %d", resp.Pagination.PageSize)
	}
}

func TestNewPaginatedResponse_MetadataService(t *testing.T) {
	resp := response.NewPaginatedResponse([]bool{true}, "compute")

	if resp.Metadata.Service != "compute" {
		t.Errorf("expected Service 'compute', got %q", resp.Metadata.Service)
	}
}

func TestNewPaginatedResponse_MetadataTimestampIsRecent(t *testing.T) {
	before := time.Now()
	resp := response.NewPaginatedResponse([]int{1}, "timers")
	after := time.Now()

	if resp.Metadata.Timestamp.Before(before) || resp.Metadata.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in range [%v, %v]",
			resp.Metadata.Timestamp, before, after)
	}
}

// ---------------------------------------------------------------------------
// FromHTTPResponse
// ---------------------------------------------------------------------------

// makeHTTPResponse creates a minimal *http.Response using httptest.
func makeHTTPResponse(statusCode int, headers map[string]string) *http.Response {
	rec := httptest.NewRecorder()
	rec.Code = statusCode
	for k, v := range headers {
		rec.Header().Set(k, v)
	}
	return rec.Result()
}

func TestFromHTTPResponse_HTTPStatus(t *testing.T) {
	httpResp := makeHTTPResponse(http.StatusOK, nil)
	meta := response.FromHTTPResponse(httpResp, "auth")

	if meta.HTTPStatus != http.StatusOK {
		t.Errorf("expected HTTPStatus 200, got %d", meta.HTTPStatus)
	}
}

func TestFromHTTPResponse_HTTPStatus404(t *testing.T) {
	httpResp := makeHTTPResponse(http.StatusNotFound, nil)
	meta := response.FromHTTPResponse(httpResp, "transfer")

	if meta.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTPStatus 404, got %d", meta.HTTPStatus)
	}
}

func TestFromHTTPResponse_ServiceSet(t *testing.T) {
	httpResp := makeHTTPResponse(200, nil)
	meta := response.FromHTTPResponse(httpResp, "groups")

	if meta.Service != "groups" {
		t.Errorf("expected Service 'groups', got %q", meta.Service)
	}
}

func TestFromHTTPResponse_TimestampIsRecent(t *testing.T) {
	before := time.Now()
	httpResp := makeHTTPResponse(200, nil)
	meta := response.FromHTTPResponse(httpResp, "svc")
	after := time.Now()

	if meta.Timestamp.Before(before) || meta.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in range [%v, %v]", meta.Timestamp, before, after)
	}
}

func TestFromHTTPResponse_RateLimitHeaders(t *testing.T) {
	resetUnix := time.Now().Add(60 * time.Second).Unix()
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-RateLimit-Limit":     "1000",
		"X-RateLimit-Remaining": "42",
		"X-RateLimit-Reset":     strconv.FormatInt(resetUnix, 10),
	})
	meta := response.FromHTTPResponse(httpResp, "search")

	if meta.RateLimit == nil {
		t.Fatal("expected RateLimit to be set")
	}
	if meta.RateLimit.Limit != 1000 {
		t.Errorf("expected Limit 1000, got %d", meta.RateLimit.Limit)
	}
	if meta.RateLimit.Remaining != 42 {
		t.Errorf("expected Remaining 42, got %d", meta.RateLimit.Remaining)
	}
	if meta.RateLimit.ResetTime.Unix() != resetUnix {
		t.Errorf("expected ResetTime unix %d, got %d", resetUnix, meta.RateLimit.ResetTime.Unix())
	}
}

func TestFromHTTPResponse_RateLimitWithRetryAfter(t *testing.T) {
	resetUnix := time.Now().Add(30 * time.Second).Unix()
	httpResp := makeHTTPResponse(429, map[string]string{
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": "0",
		"X-RateLimit-Reset":     strconv.FormatInt(resetUnix, 10),
		"Retry-After":           "30",
	})
	meta := response.FromHTTPResponse(httpResp, "flows")

	if meta.RateLimit == nil {
		t.Fatal("expected RateLimit to be set")
	}
	if meta.RateLimit.RetryAfter != 30 {
		t.Errorf("expected RetryAfter 30, got %d", meta.RateLimit.RetryAfter)
	}
}

func TestFromHTTPResponse_MissingRateLimitHeaders(t *testing.T) {
	// Only two of the three required headers present — RateLimit must remain nil.
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-RateLimit-Limit":     "500",
		"X-RateLimit-Remaining": "250",
		// X-RateLimit-Reset intentionally absent
	})
	meta := response.FromHTTPResponse(httpResp, "compute")

	if meta.RateLimit != nil {
		t.Errorf("expected RateLimit to be nil when reset header is missing, got %+v", meta.RateLimit)
	}
}

func TestFromHTTPResponse_NoRateLimitHeaders(t *testing.T) {
	httpResp := makeHTTPResponse(200, nil)
	meta := response.FromHTTPResponse(httpResp, "timers")

	if meta.RateLimit != nil {
		t.Errorf("expected RateLimit nil when no headers present, got %+v", meta.RateLimit)
	}
}

func TestFromHTTPResponse_XRequestIDStored(t *testing.T) {
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-Request-Id": "req-abc-123",
	})
	meta := response.FromHTTPResponse(httpResp, "auth")

	if meta.Headers == nil {
		t.Fatal("expected Headers map to be initialized")
	}
	if meta.Headers["X-Request-Id"] != "req-abc-123" {
		t.Errorf("expected X-Request-Id 'req-abc-123', got %q", meta.Headers["X-Request-Id"])
	}
}

func TestFromHTTPResponse_APIVersionHeader(t *testing.T) {
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-API-Version": "2.0",
	})
	meta := response.FromHTTPResponse(httpResp, "svc")

	if meta.APIVersion != "2.0" {
		t.Errorf("expected APIVersion '2.0', got %q", meta.APIVersion)
	}
}

func TestFromHTTPResponse_MissingAPIVersionHeader(t *testing.T) {
	httpResp := makeHTTPResponse(200, nil)
	meta := response.FromHTTPResponse(httpResp, "svc")

	if meta.APIVersion != "" {
		t.Errorf("expected empty APIVersion, got %q", meta.APIVersion)
	}
}

func TestFromHTTPResponse_HeadersMapAlwaysInitialised(t *testing.T) {
	httpResp := makeHTTPResponse(204, nil)
	meta := response.FromHTTPResponse(httpResp, "svc")

	if meta.Headers == nil {
		t.Error("expected Headers map to be non-nil even when no headers are present")
	}
}

func TestFromHTTPResponse_InvalidRateLimitValues(t *testing.T) {
	// Non-numeric values should not panic; RateLimit is nil because all three
	// headers must be present and parseable before the struct is created.
	resetUnix := time.Now().Add(60 * time.Second).Unix()
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-RateLimit-Limit":     "not-a-number",
		"X-RateLimit-Remaining": "42",
		"X-RateLimit-Reset":     strconv.FormatInt(resetUnix, 10),
	})
	// Should not panic.
	meta := response.FromHTTPResponse(httpResp, "svc")
	// Limit header is present (even if unparseable), Remaining and Reset are
	// present — the RateLimit struct is created but Limit defaults to 0.
	if meta.RateLimit == nil {
		t.Fatal("expected RateLimit struct to be created when all three headers are present")
	}
	if meta.RateLimit.Limit != 0 {
		t.Errorf("expected Limit 0 for unparseable value, got %d", meta.RateLimit.Limit)
	}
	if meta.RateLimit.Remaining != 42 {
		t.Errorf("expected Remaining 42, got %d", meta.RateLimit.Remaining)
	}
}

// ---------------------------------------------------------------------------
// Response builder methods: WithRequestID, WithMetadata
// ---------------------------------------------------------------------------

func TestResponse_WithRequestID(t *testing.T) {
	resp := response.NewResponse("data", "svc")
	returned := resp.WithRequestID("rid-999")

	if resp.RequestID != "rid-999" {
		t.Errorf("expected RequestID 'rid-999', got %q", resp.RequestID)
	}
	// Should return the same pointer for chaining.
	if returned != resp {
		t.Error("WithRequestID should return the same *Response pointer")
	}
}

func TestResponse_WithMetadata(t *testing.T) {
	resp := response.NewResponse(1, "svc")
	meta := response.ResponseMetadata{
		Service:    "custom-svc",
		APIVersion: "3.1",
		HTTPStatus: 201,
	}
	returned := resp.WithMetadata(meta)

	if resp.Metadata.Service != "custom-svc" {
		t.Errorf("expected Service 'custom-svc', got %q", resp.Metadata.Service)
	}
	if resp.Metadata.APIVersion != "3.1" {
		t.Errorf("expected APIVersion '3.1', got %q", resp.Metadata.APIVersion)
	}
	if resp.Metadata.HTTPStatus != 201 {
		t.Errorf("expected HTTPStatus 201, got %d", resp.Metadata.HTTPStatus)
	}
	if returned != resp {
		t.Error("WithMetadata should return the same *Response pointer")
	}
}

// ---------------------------------------------------------------------------
// PaginatedResponse builder methods: WithRequestID, WithMetadata, WithPagination
// ---------------------------------------------------------------------------

func TestPaginatedResponse_WithRequestID(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1, 2}, "svc")
	returned := resp.WithRequestID("prid-42")

	if resp.RequestID != "prid-42" {
		t.Errorf("expected RequestID 'prid-42', got %q", resp.RequestID)
	}
	if returned != resp {
		t.Error("WithRequestID should return the same *PaginatedResponse pointer")
	}
}

func TestPaginatedResponse_WithMetadata(t *testing.T) {
	resp := response.NewPaginatedResponse([]string{"a"}, "svc")
	meta := response.ResponseMetadata{
		Service:    "override-svc",
		HTTPStatus: 200,
	}
	returned := resp.WithMetadata(meta)

	if resp.Metadata.Service != "override-svc" {
		t.Errorf("expected Service 'override-svc', got %q", resp.Metadata.Service)
	}
	if returned != resp {
		t.Error("WithMetadata should return the same *PaginatedResponse pointer")
	}
}

func TestPaginatedResponse_WithPagination(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1}, "svc")
	pg := response.PaginationInfo{
		NextToken: "tok-xyz",
		HasMore:   true,
		Limit:     25,
		Total:     100,
		Offset:    25,
		Page:      2,
		PageSize:  25,
	}
	returned := resp.WithPagination(pg)

	if resp.Pagination.NextToken != "tok-xyz" {
		t.Errorf("expected NextToken 'tok-xyz', got %q", resp.Pagination.NextToken)
	}
	if !resp.Pagination.HasMore {
		t.Error("expected HasMore true")
	}
	if resp.Pagination.Limit != 25 {
		t.Errorf("expected Limit 25, got %d", resp.Pagination.Limit)
	}
	if resp.Pagination.Total != 100 {
		t.Errorf("expected Total 100, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.Offset != 25 {
		t.Errorf("expected Offset 25, got %d", resp.Pagination.Offset)
	}
	if resp.Pagination.Page != 2 {
		t.Errorf("expected Page 2, got %d", resp.Pagination.Page)
	}
	if returned != resp {
		t.Error("WithPagination should return the same *PaginatedResponse pointer")
	}
}

// ---------------------------------------------------------------------------
// Pagination helpers: IsLastPage, GetNextToken, GetPageSize, GetTotalCount
// ---------------------------------------------------------------------------

func TestIsLastPage_TrueWhenNoMore(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1, 2}, "svc")
	resp.WithPagination(response.PaginationInfo{HasMore: false})

	if !resp.IsLastPage() {
		t.Error("expected IsLastPage to return true")
	}
}

func TestIsLastPage_FalseWhenHasMore(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1}, "svc")
	resp.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "next"})

	if resp.IsLastPage() {
		t.Error("expected IsLastPage to return false")
	}
}

func TestGetNextToken(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1}, "svc")
	resp.WithPagination(response.PaginationInfo{NextToken: "page2-token", HasMore: true})

	if got := resp.GetNextToken(); got != "page2-token" {
		t.Errorf("expected 'page2-token', got %q", got)
	}
}

func TestGetNextToken_EmptyWhenLastPage(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1}, "svc")
	// Default HasMore is false, NextToken is empty.
	if got := resp.GetNextToken(); got != "" {
		t.Errorf("expected empty token, got %q", got)
	}
}

func TestGetPageSize(t *testing.T) {
	items := []string{"x", "y", "z"}
	resp := response.NewPaginatedResponse(items, "svc")

	if got := resp.GetPageSize(); got != 3 {
		t.Errorf("expected GetPageSize 3, got %d", got)
	}
}

func TestGetPageSize_Empty(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{}, "svc")
	if got := resp.GetPageSize(); got != 0 {
		t.Errorf("expected GetPageSize 0, got %d", got)
	}
}

func TestGetTotalCount(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{1, 2}, "svc")
	resp.WithPagination(response.PaginationInfo{Total: 500})

	if got := resp.GetTotalCount(); got != 500 {
		t.Errorf("expected GetTotalCount 500, got %d", got)
	}
}

func TestGetTotalCount_Zero(t *testing.T) {
	resp := response.NewPaginatedResponse([]int{}, "svc")
	if got := resp.GetTotalCount(); got != 0 {
		t.Errorf("expected GetTotalCount 0, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Iterator[T]
// ---------------------------------------------------------------------------

func TestIterator_NewIterator(t *testing.T) {
	page := response.NewPaginatedResponse([]int{1, 2, 3}, "svc")
	iter := response.NewIterator(page, nil)
	if iter == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestIterator_Next_SinglePage(t *testing.T) {
	page := response.NewPaginatedResponse([]int{10, 20, 30}, "svc")
	iter := response.NewIterator(page, nil)

	var got []int
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}

	expected := []int{10, 20, 30}
	if len(got) != len(expected) {
		t.Fatalf("expected %d items, got %d: %v", len(expected), len(got), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("item[%d]: got %d, want %d", i, got[i], expected[i])
		}
	}
}

func TestIterator_Next_EmptyPage(t *testing.T) {
	page := response.NewPaginatedResponse([]int{}, "svc")
	iter := response.NewIterator(page, nil)

	_, ok := iter.Next()
	if ok {
		t.Error("expected Next to return false for empty page")
	}
}

func TestIterator_Next_MultiplePages(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1, 2}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "p2"})

	page2 := response.NewPaginatedResponse([]int{3, 4}, "svc")
	page2.WithPagination(response.PaginationInfo{HasMore: false})

	fetchNext := func(token string) (*response.PaginatedResponse[int], error) {
		if token == "p2" {
			return page2, nil
		}
		return nil, fmt.Errorf("unexpected token: %q", token)
	}

	iter := response.NewIterator(page1, fetchNext)
	var got []int
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}

	if iter.Error() != nil {
		t.Fatalf("unexpected error: %v", iter.Error())
	}

	expected := []int{1, 2, 3, 4}
	if len(got) != len(expected) {
		t.Fatalf("expected %d items, got %d: %v", len(expected), len(got), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("item[%d]: got %d, want %d", i, got[i], expected[i])
		}
	}
}

func TestIterator_Next_FetchError(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "bad"})

	fetchErr := errors.New("network error")
	fetchNext := func(_ string) (*response.PaginatedResponse[int], error) {
		return nil, fetchErr
	}

	iter := response.NewIterator(page1, fetchNext)

	// Consume page 1.
	iter.Next()

	// Next call should trigger fetch and fail.
	_, ok := iter.Next()
	if ok {
		t.Error("expected Next to return false after fetch error")
	}
	if iter.Error() == nil {
		t.Fatal("expected iterator to store error")
	}
	if !errors.Is(iter.Error(), fetchErr) {
		t.Errorf("expected fetchErr, got %v", iter.Error())
	}
}

func TestIterator_Error_NilByDefault(t *testing.T) {
	page := response.NewPaginatedResponse([]int{1}, "svc")
	iter := response.NewIterator(page, nil)
	if iter.Error() != nil {
		t.Errorf("expected nil error, got %v", iter.Error())
	}
}

func TestIterator_Reset(t *testing.T) {
	page := response.NewPaginatedResponse([]int{1, 2, 3}, "svc")
	iter := response.NewIterator(page, nil)

	// Consume one item.
	iter.Next()

	// Reset — should restart from index 0 within the current page.
	iter.Reset()

	// After reset, Next should yield the first item again.
	v, ok := iter.Next()
	if !ok {
		t.Fatal("expected item after Reset")
	}
	if v != 1 {
		t.Errorf("expected first item 1 after reset, got %d", v)
	}
}

func TestIterator_Reset_ClearsError(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "bad"})

	fetchNext := func(_ string) (*response.PaginatedResponse[int], error) {
		return nil, errors.New("oops")
	}

	iter := response.NewIterator(page1, fetchNext)
	iter.Next() // consume page1 item
	iter.Next() // triggers fetch error

	if iter.Error() == nil {
		t.Fatal("expected error to be set")
	}

	iter.Reset()

	if iter.Error() != nil {
		t.Errorf("expected error to be cleared after Reset, got %v", iter.Error())
	}
}

func TestIterator_ToSlice_SinglePage(t *testing.T) {
	page := response.NewPaginatedResponse([]string{"a", "b", "c"}, "svc")
	iter := response.NewIterator(page, nil)

	result, err := iter.ToSlice()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
}

func TestIterator_ToSlice_MultiplePages(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1, 2}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "next"})

	page2 := response.NewPaginatedResponse([]int{3, 4, 5}, "svc")

	fetchNext := func(_ string) (*response.PaginatedResponse[int], error) {
		return page2, nil
	}

	iter := response.NewIterator(page1, fetchNext)
	result, err := iter.ToSlice()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 items, got %d: %v", len(result), result)
	}
}

func TestIterator_ToSlice_FetchError(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "bad"})

	fetchErr := errors.New("fetch failed")
	fetchNext := func(_ string) (*response.PaginatedResponse[int], error) {
		return nil, fetchErr
	}

	iter := response.NewIterator(page1, fetchNext)
	_, err := iter.ToSlice()
	if err == nil {
		t.Fatal("expected error from ToSlice")
	}
	if !errors.Is(err, fetchErr) {
		t.Errorf("expected fetchErr, got %v", err)
	}
}

func TestIterator_Count(t *testing.T) {
	page := response.NewPaginatedResponse([]int{1, 2, 3, 4, 5}, "svc")
	iter := response.NewIterator(page, nil)

	count, err := iter.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestIterator_Count_MultiplePages(t *testing.T) {
	page1 := response.NewPaginatedResponse([]int{1, 2, 3}, "svc")
	page1.WithPagination(response.PaginationInfo{HasMore: true, NextToken: "p2"})

	page2 := response.NewPaginatedResponse([]int{4, 5}, "svc")

	fetchNext := func(_ string) (*response.PaginatedResponse[int], error) {
		return page2, nil
	}

	iter := response.NewIterator(page1, fetchNext)
	count, err := iter.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestIterator_Count_Empty(t *testing.T) {
	page := response.NewPaginatedResponse([]string{}, "svc")
	iter := response.NewIterator(page, nil)

	count, err := iter.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Service-specific response constructors
// ---------------------------------------------------------------------------

func TestNewAuthResponse(t *testing.T) {
	resp := response.NewAuthResponse("token-data")
	if resp == nil {
		t.Fatal("expected non-nil AuthResponse")
	}
	if resp.Response == nil {
		t.Fatal("embedded Response must not be nil")
	}
	if resp.Metadata.Service != "auth" {
		t.Errorf("expected Service 'auth', got %q", resp.Metadata.Service)
	}
	if resp.Data != "token-data" {
		t.Errorf("expected Data 'token-data', got %v", resp.Data)
	}
}

func TestNewTransferResponse(t *testing.T) {
	resp := response.NewTransferResponse(sampleData{ID: 7, Name: "task"})
	if resp == nil {
		t.Fatal("expected non-nil TransferResponse")
	}
	if resp.Metadata.Service != "transfer" {
		t.Errorf("expected Service 'transfer', got %q", resp.Metadata.Service)
	}
	if resp.Data.ID != 7 {
		t.Errorf("expected Data.ID 7, got %d", resp.Data.ID)
	}
}

func TestNewGroupsResponse(t *testing.T) {
	resp := response.NewGroupsResponse(42)
	if resp == nil {
		t.Fatal("expected non-nil GroupsResponse")
	}
	if resp.Metadata.Service != "groups" {
		t.Errorf("expected Service 'groups', got %q", resp.Metadata.Service)
	}
	if resp.Data != 42 {
		t.Errorf("expected Data 42, got %v", resp.Data)
	}
}

func TestNewSearchResponse(t *testing.T) {
	resp := response.NewSearchResponse(true)
	if resp == nil {
		t.Fatal("expected non-nil SearchResponse")
	}
	if resp.Metadata.Service != "search" {
		t.Errorf("expected Service 'search', got %q", resp.Metadata.Service)
	}
}

func TestNewFlowsResponse(t *testing.T) {
	resp := response.NewFlowsResponse("run-id")
	if resp == nil {
		t.Fatal("expected non-nil FlowsResponse")
	}
	if resp.Metadata.Service != "flows" {
		t.Errorf("expected Service 'flows', got %q", resp.Metadata.Service)
	}
	if resp.Data != "run-id" {
		t.Errorf("expected Data 'run-id', got %v", resp.Data)
	}
}

func TestNewComputeResponse(t *testing.T) {
	resp := response.NewComputeResponse(3.14)
	if resp == nil {
		t.Fatal("expected non-nil ComputeResponse")
	}
	if resp.Metadata.Service != "compute" {
		t.Errorf("expected Service 'compute', got %q", resp.Metadata.Service)
	}
}

func TestNewTimersResponse(t *testing.T) {
	resp := response.NewTimersResponse([]string{"timer-1", "timer-2"})
	if resp == nil {
		t.Fatal("expected non-nil TimersResponse")
	}
	if resp.Metadata.Service != "timers" {
		t.Errorf("expected Service 'timers', got %q", resp.Metadata.Service)
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 items in Data, got %d", len(resp.Data))
	}
}

// ---------------------------------------------------------------------------
// RateLimitInfo struct fields
// ---------------------------------------------------------------------------

func TestRateLimitInfo_AllFields(t *testing.T) {
	resetTime := time.Unix(1700000000, 0)
	ri := response.RateLimitInfo{
		Limit:      500,
		Remaining:  123,
		ResetTime:  resetTime,
		RetryAfter: 60,
	}

	if ri.Limit != 500 {
		t.Errorf("expected Limit 500, got %d", ri.Limit)
	}
	if ri.Remaining != 123 {
		t.Errorf("expected Remaining 123, got %d", ri.Remaining)
	}
	if !ri.ResetTime.Equal(resetTime) {
		t.Errorf("expected ResetTime %v, got %v", resetTime, ri.ResetTime)
	}
	if ri.RetryAfter != 60 {
		t.Errorf("expected RetryAfter 60, got %d", ri.RetryAfter)
	}
}

func TestRateLimitInfo_FromHTTPResponse_ValuesMatchHeaders(t *testing.T) {
	resetUnix := int64(1800000000)
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-RateLimit-Limit":     "200",
		"X-RateLimit-Remaining": "199",
		"X-RateLimit-Reset":     strconv.FormatInt(resetUnix, 10),
		"Retry-After":           "5",
	})

	meta := response.FromHTTPResponse(httpResp, "svc")

	if meta.RateLimit == nil {
		t.Fatal("expected RateLimit to be populated")
	}
	if meta.RateLimit.Limit != 200 {
		t.Errorf("Limit: want 200, got %d", meta.RateLimit.Limit)
	}
	if meta.RateLimit.Remaining != 199 {
		t.Errorf("Remaining: want 199, got %d", meta.RateLimit.Remaining)
	}
	if meta.RateLimit.ResetTime.Unix() != resetUnix {
		t.Errorf("ResetTime: want unix %d, got %d", resetUnix, meta.RateLimit.ResetTime.Unix())
	}
	if meta.RateLimit.RetryAfter != 5 {
		t.Errorf("RetryAfter: want 5, got %d", meta.RateLimit.RetryAfter)
	}
}

// ---------------------------------------------------------------------------
// ResponseMetadata fields
// ---------------------------------------------------------------------------

func TestResponseMetadata_TimestampSet(t *testing.T) {
	resp := response.NewResponse("x", "svc")
	if resp.Metadata.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set (non-zero)")
	}
}

func TestResponseMetadata_ServiceField(t *testing.T) {
	services := []string{"auth", "transfer", "groups", "search", "flows", "compute", "timers"}
	for _, svc := range services {
		t.Run(svc, func(t *testing.T) {
			resp := response.NewResponse(0, svc)
			if resp.Metadata.Service != svc {
				t.Errorf("expected Service %q, got %q", svc, resp.Metadata.Service)
			}
		})
	}
}

func TestResponseMetadata_APIVersionField(t *testing.T) {
	resp := response.NewResponse[any](nil, "svc")
	resp.WithMetadata(response.ResponseMetadata{
		Service:    "svc",
		APIVersion: "v1.2.3",
	})

	if resp.Metadata.APIVersion != "v1.2.3" {
		t.Errorf("expected APIVersion 'v1.2.3', got %q", resp.Metadata.APIVersion)
	}
}

func TestResponseMetadata_APIVersionFromHTTPHeader(t *testing.T) {
	httpResp := makeHTTPResponse(200, map[string]string{
		"X-API-Version": "1.5",
	})
	meta := response.FromHTTPResponse(httpResp, "auth")

	if meta.APIVersion != "1.5" {
		t.Errorf("expected APIVersion '1.5', got %q", meta.APIVersion)
	}
}

func TestResponseMetadata_HTTPStatusFromHTTPResponse(t *testing.T) {
	codes := []int{200, 201, 400, 401, 403, 404, 429, 500}
	for _, code := range codes {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			httpResp := makeHTTPResponse(code, nil)
			meta := response.FromHTTPResponse(httpResp, "svc")
			if meta.HTTPStatus != code {
				t.Errorf("expected HTTPStatus %d, got %d", code, meta.HTTPStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Chaining builder methods
// ---------------------------------------------------------------------------

func TestResponse_BuilderChaining(t *testing.T) {
	meta := response.ResponseMetadata{
		Service:    "auth",
		APIVersion: "2",
		HTTPStatus: 200,
	}
	resp := response.NewResponse("data", "svc").
		WithRequestID("chain-rid").
		WithMetadata(meta)

	if resp.RequestID != "chain-rid" {
		t.Errorf("expected RequestID 'chain-rid', got %q", resp.RequestID)
	}
	if resp.Metadata.Service != "auth" {
		t.Errorf("expected Service 'auth', got %q", resp.Metadata.Service)
	}
}

func TestPaginatedResponse_BuilderChaining(t *testing.T) {
	pg := response.PaginationInfo{
		NextToken: "chained-token",
		HasMore:   true,
		Total:     99,
	}
	meta := response.ResponseMetadata{Service: "transfer"}
	resp := response.NewPaginatedResponse([]int{1, 2, 3}, "svc").
		WithRequestID("pagid-7").
		WithMetadata(meta).
		WithPagination(pg)

	if resp.RequestID != "pagid-7" {
		t.Errorf("expected RequestID 'pagid-7', got %q", resp.RequestID)
	}
	if resp.Metadata.Service != "transfer" {
		t.Errorf("expected Service 'transfer', got %q", resp.Metadata.Service)
	}
	if resp.Pagination.NextToken != "chained-token" {
		t.Errorf("expected NextToken 'chained-token', got %q", resp.Pagination.NextToken)
	}
}
