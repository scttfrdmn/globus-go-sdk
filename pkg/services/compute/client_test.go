// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client, error) {
	server := httptest.NewServer(handler)
	client, err := NewClient(
		WithAccessToken("test-token"),
		WithBaseURL(server.URL+"/"),
	)
	return server, client, err
}

func TestGetVersion(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/version" {
			t.Errorf("path = %s, want /v2/version", r.URL.Path)
		}
		if r.URL.Query().Get("service") != "web" {
			t.Errorf("service = %q, want web", r.URL.Query().Get("service"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"version": "1.0"})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.GetVersion(context.Background(), "web")
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}
	// With a service, /v2/version returns a JSON object.
	obj, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("GetVersion() = %T, want map[string]interface{}", res)
	}
	if obj["version"] != "1.0" {
		t.Errorf("version = %v", obj["version"])
	}
}

func TestGetVersion_ScalarResponse(t *testing.T) {
	// With no service, /v2/version returns a bare JSON string.
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("service") != "" {
			t.Errorf("service should be omitted, got %q", r.URL.Query().Get("service"))
		}
		_ = json.NewEncoder(w).Encode("2.34.0")
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.GetVersion(context.Background(), "")
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}
	if res != "2.34.0" {
		t.Errorf("GetVersion() = %v, want %q", res, "2.34.0")
	}
}

func TestGetEndpoints(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/endpoints" {
			t.Errorf("path = %s, want /v2/endpoints", r.URL.Path)
		}
		if r.URL.Query().Get("role") != "owner" {
			t.Errorf("role = %q, want owner", r.URL.Query().Get("role"))
		}
		if r.URL.Query().Get("limit") != "" {
			t.Error("limit is not an upstream param")
		}
		// /v2/endpoints returns a top-level JSON array.
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"uuid": "ep-1", "name": "one"},
			{"uuid": "ep-2", "name": "two"},
		})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	eps, err := client.GetEndpoints(context.Background(), &GetEndpointsOptions{Role: "owner"})
	if err != nil {
		t.Fatalf("GetEndpoints() error = %v", err)
	}
	if len(eps) != 2 {
		t.Fatalf("GetEndpoints() len = %d, want 2", len(eps))
	}
	if eps[0]["uuid"] != "ep-1" {
		t.Errorf("eps[0].uuid = %v, want ep-1", eps[0]["uuid"])
	}
}

func TestGetEndpoint(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/endpoints/ep-1" {
			t.Errorf("path = %s, want /v2/endpoints/ep-1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"uuid": "ep-1", "status": "online"})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.GetEndpoint(context.Background(), "ep-1")
	if err != nil {
		t.Fatalf("GetEndpoint() error = %v", err)
	}
	if res["status"] != "online" {
		t.Errorf("status = %v", res["status"])
	}

	if _, err := client.GetEndpoint(context.Background(), ""); err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestSubmitV2(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/submit" {
			t.Errorf("%s %s, want POST /v2/submit", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"request_id": "r1"})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.Submit(context.Background(), map[string]interface{}{"tasks": map[string]interface{}{}})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if res["request_id"] != "r1" {
		t.Errorf("request_id = %v", res["request_id"])
	}
}

func TestSubmitV3(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/endpoints/ep-1/submit" {
			t.Errorf("%s %s, want POST /v3/endpoints/ep-1/submit", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"request_id": "r2"})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.SubmitV3(context.Background(), "ep-1", map[string]interface{}{"tasks": []interface{}{}})
	if err != nil {
		t.Fatalf("SubmitV3() error = %v", err)
	}
	if res["request_id"] != "r2" {
		t.Errorf("request_id = %v", res["request_id"])
	}
}

func TestGetBatchStatus(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/batch_status" {
			t.Errorf("%s %s, want POST /v2/batch_status", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		ids, _ := body["task_ids"].([]interface{})
		if len(ids) != 2 {
			t.Errorf("task_ids len = %d, want 2", len(ids))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": map[string]interface{}{}})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	if _, err := client.GetBatchStatus(context.Background(), []string{"t1", "t2"}); err != nil {
		t.Fatalf("GetBatchStatus() error = %v", err)
	}
}

func TestRegisterFunction(t *testing.T) {
	server, client, err := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/functions" {
			t.Errorf("%s %s, want POST /v2/functions", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"function_id": "fn-1"})
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer server.Close()

	res, err := client.RegisterFunction(context.Background(), map[string]interface{}{"function_name": "hello"})
	if err != nil {
		t.Fatalf("RegisterFunction() error = %v", err)
	}
	if res["function_id"] != "fn-1" {
		t.Errorf("function_id = %v", res["function_id"])
	}
}
