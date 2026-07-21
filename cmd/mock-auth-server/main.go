// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors

// Command mock-auth-server is a local stand-in for Globus Auth's OAuth2
// endpoints, for exercising the globus-cli login flow offline. It is a test/dev
// helper, NOT a production artifact.
//
// It implements:
//
//	GET  /v2/oauth2/authorize  — an HTML page that immediately redirects to the
//	                             caller's redirect_uri with a dummy code+state.
//	POST /v2/oauth2/token      — returns a primary token plus one per resource
//	                             server in other_tokens (mirroring a real Globus
//	                             multi-resource-server grant).
//
// Usage:
//
//	go run ./cmd/mock-auth-server &            # serves on :8099 by default
//	GLOBUS_AUTH_BASE_URL=http://localhost:8099/v2/ \
//	  go run ./cmd/globus-cli login
//	go run ./cmd/globus-cli token export-env
//
// The tokens are obviously fake ("mock-<rs>-token") and only prove the login →
// other_tokens → per-resource-server-file → export-env plumbing; they will not
// authenticate against the real Globus API.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

// resourceServers are the resource-server names the mock issues tokens for.
// These match the names globus-cli maps to GLOBUS_TEST_<SVC>_TOKEN.
var resourceServers = []string{
	"transfer.api.globus.org",
	"groups.api.globus.org",
	"search.api.globus.org",
	"flows.globus.org",
}

func main() {
	addr := flag.String("addr", ":8099", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/oauth2/authorize", handleAuthorize)
	mux.HandleFunc("/v2/oauth2/token", handleToken)

	log.Printf("mock-auth-server listening on %s", *addr)
	log.Printf("point the CLI at it with: GLOBUS_AUTH_BASE_URL=http://localhost%s/v2/", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("mock-auth-server: %v", err)
	}
}

// handleAuthorize renders a page that immediately redirects back to the CLI's
// local callback with a dummy authorization code, echoing the state.
func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	if redirectURI == "" {
		http.Error(w, "missing redirect_uri", http.StatusBadRequest)
		return
	}

	target := fmt.Sprintf("%s?code=mock-auth-code&state=%s", redirectURI, state)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html><head><meta http-equiv="refresh" content="0; url=%s">
<title>Mock Globus Auth</title></head>
<body>
<h1>Mock Globus Auth</h1>
<p>Approving the login and redirecting back to the CLI…</p>
<p>If you are not redirected, <a href="%s">click here</a>.</p>
</body></html>`, target, target)
}

// handleToken returns a mock token response: a primary token for the first
// resource server and the remainder under other_tokens.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	mkToken := func(rs string) map[string]interface{} {
		return map[string]interface{}{
			"access_token":    "mock-" + rs + "-token",
			"refresh_token":   "mock-" + rs + "-refresh",
			"expires_in":      3600,
			"resource_server": rs,
			"token_type":      "Bearer",
			"scope":           "urn:mock:" + rs,
		}
	}

	primary := mkToken(resourceServers[0])
	others := make([]map[string]interface{}, 0, len(resourceServers)-1)
	for _, rs := range resourceServers[1:] {
		others = append(others, mkToken(rs))
	}
	primary["other_tokens"] = others

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(primary); err != nil {
		log.Printf("mock-auth-server: encode error: %v", err)
	}
}
