// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

// resourceServerEnvVar maps a Globus resource-server name to the environment
// variable the integration tests read for a pre-supplied token. Resource
// servers without a mapping (auth, timers, compute) are omitted — those tests
// use client credentials.
var resourceServerEnvVar = map[string]string{
	"transfer.api.globus.org": "GLOBUS_TEST_TRANSFER_TOKEN",
	"groups.api.globus.org":   "GLOBUS_TEST_GROUPS_TOKEN",
	"search.api.globus.org":   "GLOBUS_TEST_SEARCH_TOKEN",
	"flows.globus.org":        "GLOBUS_TEST_FLOWS_TOKEN",
}

// newAuthClient builds the SDK auth client for the CLI. If GLOBUS_AUTH_BASE_URL
// is set, the client's base URL is overridden — used to point the login flow at
// a local mock auth server for offline testing (see cmd/mock-auth-server).
func newAuthClient(config *Config) (*auth.Client, error) {
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	authClient, err := sdkConfig.NewAuthClient()
	if err != nil {
		return nil, fmt.Errorf("error creating auth client: %w", err)
	}
	if base := os.Getenv("GLOBUS_AUTH_BASE_URL"); base != "" {
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		authClient.Client.BaseURL = base
	}
	return authClient, nil
}

// Config holds the CLI configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokensDir    string `json:"tokens_dir"`
}

// TokenInfo holds the token information
type TokenInfo struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	ExpiresIn      int       `json:"expires_in"`
	ExpiresAt      time.Time `json:"expires_at"`
	Scope          string    `json:"scope"`
	TokenType      string    `json:"token_type"`
	ResourceServer string    `json:"resource_server,omitempty"`
	ResourceID     string    `json:"resource_id,omitempty"`
}

// tokenFileForResourceServer maps a resource-server name to a token filename.
// Slashes and other separators become dashes so it is filesystem-safe.
func tokenFileForResourceServer(resourceServer string) string {
	safe := strings.NewReplacer("/", "-", ":", "-", " ", "-").Replace(resourceServer)
	return "rs-" + safe
}

const (
	// DefaultConfigFile is the default config file name
	DefaultConfigFile = "globus-cli-config"

	// DefaultTokenFile is the default token file name
	DefaultTokenFile = "globus-cli-token"

	// DefaultRedirectURI is the redirect URI for browser-based auth
	DefaultRedirectURI = "http://localhost:8080/callback"
)

// IsTokenValid checks if a token is still valid (not expired)
func IsTokenValid(token *TokenInfo) bool {
	return time.Now().Before(token.ExpiresAt)
}

// LoadOrCreateConfig loads the CLI configuration from disk
func LoadOrCreateConfig() (*Config, error) {
	// Check for the config file
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Create the config dir if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, DefaultConfigFile+".json")
	if _, err := os.Stat(configFile); err == nil {
		// Load the config file
		f, err := os.Open(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer f.Close()

		var config Config
		if err := json.NewDecoder(f).Decode(&config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		// Make sure the tokens directory exists
		if config.TokensDir == "" {
			config.TokensDir = configDir
		}

		if err := os.MkdirAll(config.TokensDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create tokens directory: %w", err)
		}

		return &config, nil
	}

	// Create a default config
	config := &Config{
		ClientID:     "91739f66-9226-4382-8295-4ab5f0a8f88e", // Default Globus CLI client ID
		ClientSecret: "",                                     // No client secret for native apps
		TokensDir:    configDir,
	}

	// Save the config
	if err := saveConfig(config, configFile); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return config, nil
}

// saveConfig saves the config to the specified file
func saveConfig(config *Config, configFile string) error {
	f, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadToken loads a token from disk
func LoadToken(config *Config, tokenName string) (*TokenInfo, error) {
	tokenFile := filepath.Join(config.TokensDir, tokenName+".json")
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file: %w", err)
	}
	defer f.Close()

	var token TokenInfo
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		// Try to refresh the token
		if token.RefreshToken != "" {
			refreshedToken, err := refreshToken(config, token.RefreshToken)
			if err != nil {
				return nil, fmt.Errorf("token expired and refresh failed: %w", err)
			}
			saveToken(config, tokenName, refreshedToken)
			return refreshedToken, nil
		}
		return nil, fmt.Errorf("token expired and no refresh token available")
	}

	return &token, nil
}

// saveToken saves a token to disk
func saveToken(config *Config, tokenName string, token *TokenInfo) error {
	tokenFile := filepath.Join(config.TokensDir, tokenName+".json")
	f, err := os.OpenFile(tokenFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open token file for writing: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// getConfigDir returns the config directory
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".globus-cli")
	return configDir, nil
}

// LoginCommand handles the login command
func LoginCommand(args []string) error {
	// Load the configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Start the local server
	server, err := startLocalServer()
	if err != nil {
		return fmt.Errorf("error starting local server: %w", err)
	}

	// Generate a random state value
	state, err := generateRandomState()
	if err != nil {
		return fmt.Errorf("error generating state: %w", err)
	}

	// Generate the authorization URL
	authClient, err := newAuthClient(config)
	if err != nil {
		return err
	}

	// Use all scopes to make the login useful for all commands
	scopes := []string{
		pkg.AuthScope,
		pkg.TransferScope,
		pkg.GroupsScope,
		pkg.SearchScope,
		pkg.FlowsScope,
		pkg.ComputeScope,
		pkg.TimersScope,
	}

	// Set the redirect URL
	authClient.SetRedirectURL(DefaultRedirectURI)

	// Get the URL for the login
	authURL := authClient.GetAuthorizationURL(state, scopes...)

	// Open the browser
	fmt.Printf("Opening browser to login at: %s\n", authURL)
	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically. Please open this URL in your browser:\n%s\n", authURL)
	}

	// Wait for the callback
	select {
	case result := <-server.ResultChan:
		if result.Error != nil {
			return fmt.Errorf("error during login: %w", result.Error)
		}

		// Check state value
		if result.State != state {
			return fmt.Errorf("state mismatch, possible CSRF attack")
		}

		// Exchange code for token(s). Globus returns the primary token plus one
		// per additional resource server in other_tokens.
		token, otherTokens, err := exchangeCodeForToken(config, result.Code)
		if err != nil {
			return fmt.Errorf("error exchanging code for token: %w", err)
		}

		// Save the primary token under the default name.
		if err := saveToken(config, DefaultTokenFile, token); err != nil {
			return fmt.Errorf("error saving token: %w", err)
		}

		// Also save every per-resource-server token under its resource-server
		// name so `globus-cli token export-env` can surface them.
		saved := 0
		for _, ot := range append([]*TokenInfo{token}, otherTokens...) {
			if ot.ResourceServer == "" {
				continue
			}
			if err := saveToken(config, tokenFileForResourceServer(ot.ResourceServer), ot); err != nil {
				return fmt.Errorf("error saving token for %s: %w", ot.ResourceServer, err)
			}
			saved++
		}

		fmt.Printf("Login successful! Stored tokens for %d resource server(s).\n", saved)
		return nil
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("login timed out after 5 minutes")
	}
}

// CallbackResult holds the result of the OAuth callback
type CallbackResult struct {
	Code  string
	State string
	Error error
}

// CallbackServer is a simple HTTP server for handling OAuth callbacks
type CallbackServer struct {
	Server     *http.Server
	ResultChan chan CallbackResult
}

// startLocalServer starts a local HTTP server to receive the OAuth callback
func startLocalServer() (*CallbackServer, error) {
	resultChan := make(chan CallbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			resultChan <- CallbackResult{Error: fmt.Errorf("%s: %s", errParam, errDesc)}
			fmt.Fprintf(w, "<html><body><h1>Login Failed</h1><p>%s: %s</p></body></html>", errParam, errDesc)
			return
		}

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			resultChan <- CallbackResult{Error: fmt.Errorf("no code in callback")}
			fmt.Fprintf(w, "<html><body><h1>Login Failed</h1><p>No authorization code received.</p></body></html>")
			return
		}

		resultChan <- CallbackResult{Code: code, State: state}
		fmt.Fprintf(w, "<html><body><h1>Login Successful</h1><p>You can close this window and return to the CLI.</p></body></html>")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			resultChan <- CallbackResult{Error: err}
		}
	}()

	return &CallbackServer{
		Server:     server,
		ResultChan: resultChan,
	}, nil
}

// generateRandomState generates a random state value
func generateRandomState() (string, error) {
	buffer := make([]byte, 32)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buffer), nil
}

// exchangeCodeForToken exchanges an authorization code for the primary token
// and any additional per-resource-server tokens (other_tokens).
func exchangeCodeForToken(config *Config, code string) (*TokenInfo, []*TokenInfo, error) {
	authClient, err := newAuthClient(config)
	if err != nil {
		return nil, nil, err
	}

	// Set redirect URL for authorization code exchange
	authClient.SetRedirectURL(DefaultRedirectURI)

	// Exchange code for token
	tokenResp, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
	if err != nil {
		return nil, nil, fmt.Errorf("error exchanging authorization code: %w", err)
	}

	token := tokenInfoFromResponse(tokenResp.AccessToken, tokenResp.RefreshToken,
		tokenResp.ExpiresIn, tokenResp.Scope, tokenResp.TokenType, tokenResp.ResourceServer)

	// Globus returns one token per additional resource server in other_tokens.
	others := []*TokenInfo{}
	otherResps, err := tokenResp.GetOtherTokens()
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing other_tokens: %w", err)
	}
	for _, ot := range otherResps {
		others = append(others, tokenInfoFromResponse(ot.AccessToken, ot.RefreshToken,
			ot.ExpiresIn, ot.Scope, ot.TokenType, ot.ResourceServer))
	}

	return token, others, nil
}

// tokenInfoFromResponse builds a TokenInfo, computing the absolute expiry.
func tokenInfoFromResponse(access, refresh string, expiresIn int, scope, tokenType, resourceServer string) *TokenInfo {
	return &TokenInfo{
		AccessToken:    access,
		RefreshToken:   refresh,
		ExpiresIn:      expiresIn,
		ExpiresAt:      time.Now().Add(time.Duration(expiresIn) * time.Second),
		Scope:          scope,
		TokenType:      tokenType,
		ResourceServer: resourceServer,
	}
}

// refreshToken refreshes a token
func refreshToken(config *Config, refreshToken string) (*TokenInfo, error) {
	// Create SDK configuration
	authClient, err := newAuthClient(config)
	if err != nil {
		return nil, err
	}

	// Refresh the token
	tokenResp, err := authClient.RefreshToken(context.Background(), refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}

	// Convert to token info
	token := &TokenInfo{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
		TokenType:    tokenResp.TokenType,
	}

	return token, nil
}

// LogoutCommand handles the logout command
func LogoutCommand(args []string) error {
	// Load the configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Check if we have a token
	token, err := LoadToken(config, DefaultTokenFile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Create SDK configuration
	authClient, err := newAuthClient(config)
	if err != nil {
		return err
	}

	// Revoke the access token
	fmt.Println("Revoking access token...")
	if err := authClient.RevokeToken(context.Background(), token.AccessToken); err != nil {
		fmt.Printf("Warning: Failed to revoke access token: %v\n", err)
	}

	// Revoke the refresh token
	if token.RefreshToken != "" {
		fmt.Println("Revoking refresh token...")
		if err := authClient.RevokeToken(context.Background(), token.RefreshToken); err != nil {
			fmt.Printf("Warning: Failed to revoke refresh token: %v\n", err)
		}
	}

	// Delete the token file
	tokenFile := filepath.Join(config.TokensDir, DefaultTokenFile+".json")
	if err := os.Remove(tokenFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error deleting token file: %w", err)
	}

	fmt.Println("Logout successful!")
	return nil
}

// TokenCommand handles the token command
func TokenCommand(args []string) error {
	// Load the configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// export-env works off the per-resource-server token files and does not
	// require the default combined token, so handle it before loading that.
	if len(args) > 0 && args[0] == "export-env" {
		return exportEnvTokens(config)
	}

	// Check if we have a token
	token, err := LoadToken(config, DefaultTokenFile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if we have a subcommand
	if len(args) > 0 {
		switch args[0] {
		case "info":
			return printTokenInfo(token)
		case "revoke":
			return revokeToken(config, token, args[1:])
		default:
			return fmt.Errorf("unknown subcommand: %s", args[0])
		}
	}

	// Default to showing token info
	return printTokenInfo(token)
}

// exportEnvTokens prints `export GLOBUS_TEST_<SVC>_TOKEN=...` lines for each
// stored per-resource-server token that maps to an integration-test variable.
// Intended to be eval'd:  eval "$(globus-cli token export-env)".
func exportEnvTokens(config *Config) error {
	printed := 0
	for resourceServer, envVar := range resourceServerEnvVar {
		token, err := LoadToken(config, tokenFileForResourceServer(resourceServer))
		if err != nil {
			// No token stored for this resource server (not requested at login,
			// or expired without a refresh token) — skip it.
			continue
		}
		fmt.Printf("export %s=%s\n", envVar, token.AccessToken)
		printed++
	}
	if printed == 0 {
		fmt.Fprintln(os.Stderr, "No per-resource-server tokens found. Run `globus-cli login` first.")
	}
	return nil
}

// printTokenInfo prints information about a token
func printTokenInfo(token *TokenInfo) error {
	fmt.Println("Token Information:")
	fmt.Printf("  Access Token: %s...%s\n", token.AccessToken[:10], token.AccessToken[len(token.AccessToken)-10:])

	if token.RefreshToken != "" {
		fmt.Printf("  Refresh Token: %s...%s\n", token.RefreshToken[:10], token.RefreshToken[len(token.RefreshToken)-10:])
	}

	fmt.Printf("  Expires At: %s\n", token.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("  Scopes: %s\n", token.Scope)

	return nil
}

// revokeToken revokes a token
func revokeToken(config *Config, token *TokenInfo, args []string) error {
	// Create SDK configuration
	authClient, err := newAuthClient(config)
	if err != nil {
		return err
	}

	// Check which token to revoke
	if len(args) > 0 && args[0] == "refresh" {
		if token.RefreshToken == "" {
			return fmt.Errorf("no refresh token to revoke")
		}

		fmt.Println("Revoking refresh token...")
		if err := authClient.RevokeToken(context.Background(), token.RefreshToken); err != nil {
			return fmt.Errorf("failed to revoke refresh token: %w", err)
		}

		// Update the token
		token.RefreshToken = ""
		if err := saveToken(config, DefaultTokenFile, token); err != nil {
			return fmt.Errorf("error saving updated token: %w", err)
		}

		fmt.Println("Refresh token revoked!")
		return nil
	}

	// Default to revoking the access token
	fmt.Println("Revoking access token...")
	if err := authClient.RevokeToken(context.Background(), token.AccessToken); err != nil {
		return fmt.Errorf("failed to revoke access token: %w", err)
	}

	// Delete the token file - this effectively logs the user out
	tokenFile := filepath.Join(config.TokensDir, DefaultTokenFile+".json")
	if err := os.Remove(tokenFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error deleting token file: %w", err)
	}

	fmt.Println("Access token revoked! You are now logged out.")
	return nil
}
