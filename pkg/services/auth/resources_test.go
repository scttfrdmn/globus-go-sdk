// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
)

func newResourceClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := NewClient(WithCoreOption(core.WithBaseURL(server.URL + "/")))
	if err != nil {
		server.Close()
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, server
}

func TestGetIdentities(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/identities" {
			t.Errorf("path = %s, want /api/identities", r.URL.Path)
		}
		if got := r.URL.Query().Get("usernames"); got != "a@x.org,b@x.org" {
			t.Errorf("usernames = %q, want comma-joined", got)
		}
		if got := r.URL.Query().Get("provision"); got != "true" {
			t.Errorf("provision = %q, want true", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"identities": []map[string]interface{}{{"id": "id-1", "username": "a@x.org"}},
		})
	})
	defer server.Close()

	ids, err := client.GetIdentities(context.Background(), &GetIdentitiesOptions{
		Usernames: []string{"a@x.org", "b@x.org"}, Provision: true,
	})
	if err != nil {
		t.Fatalf("GetIdentities() error = %v", err)
	}
	if len(ids) != 1 || ids[0].ID != "id-1" {
		t.Errorf("identities = %+v", ids)
	}
}

func TestGetIdentitiesMutuallyExclusive(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()
	_, err := client.GetIdentities(context.Background(), &GetIdentitiesOptions{
		Usernames: []string{"a"}, IDs: []string{"b"},
	})
	if err == nil {
		t.Error("expected error for mutually exclusive usernames/ids")
	}
}

func TestCreateProjectEnvelope(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/projects" {
			t.Errorf("%s %s, want POST /api/projects", r.Method, r.URL.Path)
		}
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["project"]; !ok {
			t.Error("create body must be wrapped in a \"project\" key")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"project": map[string]interface{}{"id": "proj-1", "display_name": "P"},
		})
	})
	defer server.Close()

	proj, err := client.CreateProject(context.Background(), &ProjectCreate{DisplayName: "P", AdminIDs: []string{"id-1"}})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if proj.ID != "proj-1" {
		t.Errorf("ID = %s", proj.ID)
	}
}

func TestGetProjectsEnvelope(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"projects": []map[string]interface{}{{"id": "p1"}, {"id": "p2"}},
		})
	})
	defer server.Close()

	projects, err := client.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("GetProjects() error = %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("got %d projects, want 2", len(projects))
	}
}

func TestGetOpenIDConfigurationHitsRoot(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			t.Errorf("path = %s, want host-root .well-known", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"issuer": "https://auth.globus.org"})
	})
	defer server.Close()

	doc, err := client.GetOpenIDConfiguration(context.Background())
	if err != nil {
		t.Fatalf("GetOpenIDConfiguration() error = %v", err)
	}
	if doc["issuer"] != "https://auth.globus.org" {
		t.Errorf("issuer = %v", doc["issuer"])
	}
}

func TestOAuth2GetDependentTokens(t *testing.T) {
	client, server := newResourceClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "urn:globus:auth:grant_type:dependent_token" {
			t.Errorf("grant_type = %q", r.PostForm.Get("grant_type"))
		}
		if r.PostForm.Get("access_type") != "offline" {
			t.Errorf("access_type = %q, want offline", r.PostForm.Get("access_type"))
		}
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"access_token": "at1", "resource_server": "rs1", "expires_in": 3600},
		})
	})
	defer server.Close()

	toks, err := client.OAuth2GetDependentTokens(context.Background(), "primary", true, nil)
	if err != nil {
		t.Fatalf("OAuth2GetDependentTokens() error = %v", err)
	}
	if len(toks) != 1 || toks[0].ResourceServer != "rs1" {
		t.Errorf("tokens = %+v", toks)
	}
}
