// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package login

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

const (
	defaultAuthBaseURL  = "https://auth.globus.org"
	defaultRedirectURI  = "https://auth.globus.org/v2/web/auth-code"
)

// CommandLineLoginFlowManager drives an OAuth2 authorization code flow on the
// command line: it prints the authorization URL and reads the resulting auth
// code from stdin. It implements LoginFlowManager.
type CommandLineLoginFlowManager struct {
	clientID     string
	clientSecret string // empty for native apps
	redirectURI  string
	authBaseURL  string
	httpClient   *http.Client
}

// CLIOption is a functional option for NewCommandLineLoginFlowManager.
type CLIOption func(*CommandLineLoginFlowManager)

// WithCLIRedirectURI overrides the redirect URI used in the authorization URL.
func WithCLIRedirectURI(uri string) CLIOption {
	return func(m *CommandLineLoginFlowManager) { m.redirectURI = uri }
}

// WithCLIAuthBaseURL overrides the Globus Auth base URL (default: https://auth.globus.org).
func WithCLIAuthBaseURL(baseURL string) CLIOption {
	return func(m *CommandLineLoginFlowManager) {
		m.authBaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithCLIHTTPClient sets the HTTP client used for token exchange.
func WithCLIHTTPClient(c *http.Client) CLIOption {
	return func(m *CommandLineLoginFlowManager) { m.httpClient = c }
}

// NewCommandLineLoginFlowManager creates a CommandLineLoginFlowManager.
// clientSecret may be empty for public (native) clients.
func NewCommandLineLoginFlowManager(clientID, clientSecret string, opts ...CLIOption) *CommandLineLoginFlowManager {
	m := &CommandLineLoginFlowManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  defaultRedirectURI,
		authBaseURL:  defaultAuthBaseURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// RunLoginFlow drives the interactive authorization code flow:
//  1. Builds and prints the Globus Auth authorization URL.
//  2. Reads the auth code from stdin.
//  3. Exchanges the code for tokens via the token endpoint.
//  4. Returns LoginResult containing tokens for all resource servers.
func (m *CommandLineLoginFlowManager) RunLoginFlow(ctx context.Context, params AuthParams) (*LoginResult, error) {
	scopes := append([]string{}, params.Scopes...)
	if params.RequestRefresh {
		scopes = append(scopes, "offline_access")
	}

	redirectURI := m.redirectURI
	if params.RedirectURI != "" {
		redirectURI = params.RedirectURI
	}

	authURL := m.buildAuthURL(scopes, redirectURI, params.State)

	fmt.Println("Please open the following URL in your browser to authenticate:")
	fmt.Println()
	fmt.Println("  " + authURL)
	fmt.Println()
	fmt.Print("Enter the authorization code: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("login: read auth code: %w", err)
		}
		return nil, fmt.Errorf("login: no auth code provided")
	}
	code := strings.TrimSpace(scanner.Text())
	if code == "" {
		return nil, fmt.Errorf("login: auth code is empty")
	}

	return m.exchangeCode(ctx, code, redirectURI)
}

// buildAuthURL constructs the Globus Auth authorization URL.
func (m *CommandLineLoginFlowManager) buildAuthURL(scopes []string, redirectURI, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", m.clientID)
	q.Set("redirect_uri", redirectURI)
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}
	if state != "" {
		q.Set("state", state)
	}
	return m.authBaseURL + "/v2/oauth2/authorize?" + q.Encode()
}

// exchangeCode posts the authorization code to the token endpoint and parses
// the response into a LoginResult.
func (m *CommandLineLoginFlowManager) exchangeCode(ctx context.Context, code, redirectURI string) (*LoginResult, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", m.clientID)
	form.Set("redirect_uri", redirectURI)
	if m.clientSecret != "" {
		form.Set("client_secret", m.clientSecret)
	}

	tokenURL := m.authBaseURL + "/v2/oauth2/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("login: create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login: token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login: token endpoint returned HTTP %d", resp.StatusCode)
	}

	var result tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("login: decode token response: %w", err)
	}

	tokens := []*tokenstorage.TokenData{result.toTokenData()}
	for i := range result.OtherTokens {
		tokens = append(tokens, result.OtherTokens[i].toTokenData())
	}

	return &LoginResult{Tokens: tokens}, nil
}
