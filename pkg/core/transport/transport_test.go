// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"
)

// mockClient implements interfaces.ClientInterface for testing
type mockClient struct {
	baseURL    string
	httpClient *http.Client
	logger     interfaces.Logger
}

func newMockClient(baseURL string) *mockClient {
	return &mockClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		logger:     &noopLogger{},
	}
}

func (m *mockClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return m.httpClient.Do(req)
}

func (m *mockClient) GetHTTPClient() *http.Client {
	return m.httpClient
}

func (m *mockClient) GetBaseURL() string {
	return m.baseURL
}

func (m *mockClient) GetUserAgent() string {
	return "test-agent"
}

func (m *mockClient) GetLogger() interfaces.Logger {
	return m.logger
}

// noopLogger is a no-op implementation of interfaces.Logger
type noopLogger struct{}

func (l *noopLogger) Debug(format string, args ...interface{}) {}
func (l *noopLogger) Info(format string, args ...interface{})  {}
func (l *noopLogger) Warn(format string, args ...interface{})  {}
func (l *noopLogger) Error(format string, args ...interface{}) {}

// Verify mockClient satisfies interfaces.ClientInterface at compile time
var _ interfaces.ClientInterface = (*mockClient)(nil)

// ----- NewTransport tests -----

func TestNewTransport_NilOptions(t *testing.T) {
	// Ensure env vars are unset
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	if tr == nil {
		t.Fatal("NewTransport returned nil")
	}
	if tr.Debug {
		t.Error("Expected Debug=false with nil options and no env vars")
	}
	if tr.Trace {
		t.Error("Expected Trace=false with nil options and no env vars")
	}
	if tr.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}
	if tr.Client == nil {
		t.Error("Expected Client to be set")
	}
}

func TestNewTransport_WithDebugOption(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, &Options{
		Debug:  true,
		Logger: logger,
	})

	if !tr.Debug {
		t.Error("Expected Debug=true")
	}
	if tr.Trace {
		t.Error("Expected Trace=false")
	}
	if tr.Logger != logger {
		t.Error("Expected custom Logger to be used")
	}
}

func TestNewTransport_WithTraceOption(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, &Options{
		Trace: true,
	})

	if !tr.Trace {
		t.Error("Expected Trace=true")
	}
	// Trace implies Debug
	if !tr.Debug {
		t.Error("Expected Debug=true when Trace=true")
	}
}

func TestNewTransport_EnvVarDebug(t *testing.T) {
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	defer os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	if !tr.Debug {
		t.Error("Expected Debug=true from env var")
	}
}

func TestNewTransport_EnvVarTrace(t *testing.T) {
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")
	defer os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	if !tr.Trace {
		t.Error("Expected Trace=true from env var")
	}
	// Trace implies Debug
	if !tr.Debug {
		t.Error("Expected Debug=true when Trace env var set")
	}
}

// ----- NewDeferredTransport tests -----

func TestNewDeferredTransport_NilOptions(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	dt := NewDeferredTransport(nil)

	if dt == nil {
		t.Fatal("NewDeferredTransport returned nil")
	}
	if dt.Debug {
		t.Error("Expected Debug=false with nil options")
	}
	if dt.Trace {
		t.Error("Expected Trace=false with nil options")
	}
	if dt.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}
}

func TestNewDeferredTransport_WithOptions(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	dt := NewDeferredTransport(&Options{
		Debug:  true,
		Logger: logger,
	})

	if !dt.Debug {
		t.Error("Expected Debug=true")
	}
	if dt.Logger != logger {
		t.Error("Expected custom Logger")
	}
}

func TestNewDeferredTransport_TraceImpliesDebug(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	dt := NewDeferredTransport(&Options{
		Trace: true,
	})

	if !dt.Trace {
		t.Error("Expected Trace=true")
	}
	if !dt.Debug {
		t.Error("Expected Debug=true when Trace=true")
	}
}

func TestNewDeferredTransport_EnvVarTrace(t *testing.T) {
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")
	defer os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")

	dt := NewDeferredTransport(nil)

	if !dt.Trace {
		t.Error("Expected Trace=true from env var")
	}
	if !dt.Debug {
		t.Error("Expected Debug=true when Trace env var set")
	}
}

// ----- DeferredTransport.AttachClient tests -----

func TestDeferredTransport_AttachClient(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	dt := NewDeferredTransport(&Options{
		Debug:  true,
		Trace:  true,
		Logger: logger,
	})

	client := newMockClient(server.URL + "/")
	tr := dt.AttachClient(client)

	if tr == nil {
		t.Fatal("AttachClient returned nil")
	}
	if tr.Client == nil {
		t.Error("Expected Client to be the attached client")
	}
	if !tr.Debug {
		t.Error("Expected Debug=true from DeferredTransport")
	}
	if !tr.Trace {
		t.Error("Expected Trace=true from DeferredTransport")
	}
	if tr.Logger != logger {
		t.Error("Expected Logger to be preserved from DeferredTransport")
	}
}

func TestDeferredTransport_AttachClient_NoDebug(t *testing.T) {
	os.Unsetenv("GLOBUS_SDK_HTTP_DEBUG")
	os.Unsetenv("GLOBUS_SDK_HTTP_TRACE")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDeferredTransport(nil)
	client := newMockClient(server.URL + "/")
	tr := dt.AttachClient(client)

	if tr.Debug {
		t.Error("Expected Debug=false")
	}
	if tr.Trace {
		t.Error("Expected Trace=false")
	}
}

// ----- HTTP method tests -----

func TestTransport_Get(t *testing.T) {
	var receivedMethod string
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	resp, err := tr.Get(context.Background(), "/test-path", nil, nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodGet {
		t.Errorf("Expected GET method, got %s", receivedMethod)
	}
	if receivedPath != "/test-path" {
		t.Errorf("Expected path /test-path, got %s", receivedPath)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestTransport_Post(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	body := map[string]string{"key": "value"}
	resp, err := tr.Post(context.Background(), "/create", body, nil, nil)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodPost {
		t.Errorf("Expected POST method, got %s", receivedMethod)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestTransport_Put(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	body := map[string]string{"key": "updated"}
	resp, err := tr.Put(context.Background(), "/update", body, nil, nil)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodPut {
		t.Errorf("Expected PUT method, got %s", receivedMethod)
	}
}

func TestTransport_Delete(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	resp, err := tr.Delete(context.Background(), "/delete", nil, nil)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodDelete {
		t.Errorf("Expected DELETE method, got %s", receivedMethod)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestTransport_Patch(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	body := map[string]string{"patch": "field"}
	resp, err := tr.Patch(context.Background(), "/patch", body, nil, nil)
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodPatch {
		t.Errorf("Expected PATCH method, got %s", receivedMethod)
	}
}

// ----- Transport.Request detailed tests -----

func TestTransport_Request_JSONBody(t *testing.T) {
	var receivedBody []byte
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	type payload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	body := payload{Name: "test", Value: 42}

	resp, err := tr.Request(context.Background(), http.MethodPost, "/data", body, nil, nil)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedContentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", receivedContentType)
	}

	var decoded payload
	if err := json.Unmarshal(receivedBody, &decoded); err != nil {
		t.Fatalf("Failed to decode received body: %v", err)
	}
	if decoded.Name != "test" || decoded.Value != 42 {
		t.Errorf("Received body mismatch: got %+v", decoded)
	}
}

func TestTransport_Request_QueryParams(t *testing.T) {
	var receivedQuery url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	query := url.Values{}
	query.Set("filter", "active")
	query.Set("limit", "10")

	resp, err := tr.Request(context.Background(), http.MethodGet, "/search", nil, query, nil)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedQuery.Get("filter") != "active" {
		t.Errorf("Expected filter=active, got %s", receivedQuery.Get("filter"))
	}
	if receivedQuery.Get("limit") != "10" {
		t.Errorf("Expected limit=10, got %s", receivedQuery.Get("limit"))
	}
}

func TestTransport_Request_CustomHeaders(t *testing.T) {
	var receivedHeader string
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		receivedAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	headers := http.Header{}
	headers.Set("X-Custom-Header", "custom-value")

	resp, err := tr.Request(context.Background(), http.MethodGet, "/endpoint", nil, nil, headers)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedHeader != "custom-value" {
		t.Errorf("Expected X-Custom-Header=custom-value, got %s", receivedHeader)
	}
	// Accept header should be set by default
	if receivedAccept != "application/json" {
		t.Errorf("Expected Accept=application/json, got %s", receivedAccept)
	}
}

func TestTransport_Request_AcceptHeaderNotOverridden(t *testing.T) {
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	headers := http.Header{}
	headers.Set("Accept", "text/plain")

	resp, err := tr.Request(context.Background(), http.MethodGet, "/endpoint", nil, nil, headers)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedAccept != "text/plain" {
		t.Errorf("Expected Accept=text/plain (custom), got %s", receivedAccept)
	}
}

func TestTransport_Request_InvalidJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	// Channels cannot be marshaled to JSON
	invalidBody := make(chan int)

	_, err := tr.Request(context.Background(), http.MethodPost, "/data", invalidBody, nil, nil)
	if err == nil {
		t.Error("Expected error for non-marshallable body, got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal request body") {
		t.Errorf("Expected marshal error, got: %v", err)
	}
}

func TestTransport_Request_BaseURLTrailingSlash(t *testing.T) {
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Base URL without trailing slash
	client := newMockClient(server.URL)
	tr := NewTransport(client, nil)

	resp, err := tr.Get(context.Background(), "/api/endpoint", nil, nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedPath != "/api/endpoint" {
		t.Errorf("Expected path /api/endpoint, got %s", receivedPath)
	}
}

func TestTransport_Request_WithDebugLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, &Options{
		Debug:  true,
		Logger: logger,
	})

	resp, err := tr.Get(context.Background(), "/debug-test", nil, nil)
	if err != nil {
		t.Fatalf("Get() with debug error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "HTTP Request") {
		t.Errorf("Expected 'HTTP Request' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "HTTP Response") {
		t.Errorf("Expected 'HTTP Response' in log output, got: %s", logOutput)
	}
}

func TestTransport_Request_WithTraceLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, &Options{
		Trace:  true,
		Logger: logger,
	})

	body := map[string]string{"trace": "body"}
	resp, err := tr.Post(context.Background(), "/trace-test", body, nil, nil)
	if err != nil {
		t.Fatalf("Post() with trace error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "HTTP Request") {
		t.Errorf("Expected 'HTTP Request' in log output, got: %s", logOutput)
	}
}

// ----- Transport.RoundTrip tests -----

func TestTransport_RoundTrip(t *testing.T) {
	var receivedMethod string
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("round trip response"))
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/round-trip", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	if receivedMethod != http.MethodGet {
		t.Errorf("Expected GET, got %s", receivedMethod)
	}
	if receivedPath != "/round-trip" {
		t.Errorf("Expected /round-trip, got %s", receivedPath)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestTransport_RoundTrip_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, &Options{
		Debug:  true,
		Logger: logger,
	})

	// RoundTrip with a body
	bodyContent := []byte(`{"key": "val"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/rt-debug", bytes.NewReader(bodyContent))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "HTTP Request") {
		t.Errorf("Expected 'HTTP Request' in log output, got: %s", logOutput)
	}
}

func TestTransport_RoundTrip_ImplementsHTTPRoundTripper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newMockClient(server.URL + "/")
	tr := NewTransport(client, nil)

	// Use the transport as an http.RoundTripper in an http.Client
	httpClient := &http.Client{
		Transport: tr,
	}

	resp, err := httpClient.Get(server.URL + "/rt-interface")
	if err != nil {
		t.Fatalf("http.Client with Transport as RoundTripper failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ----- DecodeResponse tests -----

func TestDecodeResponse_Success(t *testing.T) {
	type responseData struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	jsonBody := `{"name":"test","count":5}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}

	var data responseData
	if err := DecodeResponse(resp, &data); err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}
	if data.Name != "test" {
		t.Errorf("Expected Name=test, got %s", data.Name)
	}
	if data.Count != 5 {
		t.Errorf("Expected Count=5, got %d", data.Count)
	}
}

func TestDecodeResponse_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	var data map[string]interface{}
	if err := DecodeResponse(resp, &data); err != nil {
		t.Fatalf("DecodeResponse() with empty body error = %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil data for empty body, got %v", data)
	}
}

func TestDecodeResponse_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("not valid json {")),
	}

	var data map[string]interface{}
	err := DecodeResponse(resp, &data)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response body") {
		t.Errorf("Expected unmarshal error, got: %v", err)
	}
}

func TestDecodeResponse_BodyClosed(t *testing.T) {
	// Verify that DecodeResponse closes the body
	closed := false
	body := &trackingReadCloser{
		Reader:    strings.NewReader(`{"key":"val"}`),
		onClose:   func() { closed = true },
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       body,
	}

	var data map[string]interface{}
	_ = DecodeResponse(resp, &data)

	if !closed {
		t.Error("Expected response body to be closed after DecodeResponse")
	}
}

// trackingReadCloser tracks whether Close() was called
type trackingReadCloser struct {
	io.Reader
	onClose func()
}

func (t *trackingReadCloser) Close() error {
	if t.onClose != nil {
		t.onClose()
	}
	return nil
}
