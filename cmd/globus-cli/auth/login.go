// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

// TokenInfo represents stored token information
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Config represents the CLI configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokensDir    string `json:"tokens_dir"`
}

// Default configuration values
const (
	// DefaultClientID is the default client ID for the CLI
	DefaultClientID = "e6c75d97-532a-4c88-b031-f5a3014430e3"

	// DefaultClientSecret is the default client secret for the CLI
	DefaultClientSecret = "YOUR_CLIENT_SECRET"

	// DefaultTokenFile is the default name for the token file
	DefaultTokenFile = "default.json"
)

// LoadOrCreateConfig loads or creates the CLI configuration
func LoadOrCreateConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error determining home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".globus-cli")
	configFile := filepath.Join(configDir, "config.json")
	tokensDir := filepath.Join(configDir, "tokens")

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	// Create the tokens directory if it doesn't exist
	if err := os.MkdirAll(tokensDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating tokens directory: %w", err)
	}

	// Try to load existing config
	config := &Config{
		ClientID:     DefaultClientID,
		ClientSecret: DefaultClientSecret,
		TokensDir:    tokensDir,
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); err == nil {
		// Read the config file
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		// Parse the config
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	} else {
		// Save the default config
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshaling config: %w", err)
		}

		if err := os.WriteFile(configFile, data, 0600); err != nil {
			return nil, fmt.Errorf("error writing config file: %w", err)
		}
	}

	return config, nil
}

// StoreToken stores a token in the tokens directory
func StoreToken(config *Config, name string, token *TokenInfo) error {
	// Create the token file path
	tokenFile := filepath.Join(config.TokensDir, name+".json")

	// Marshal the token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling token: %w", err)
	}

	// Write the token to the file
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("error writing token file: %w", err)
	}

	return nil
}

// LoadToken loads a token from the tokens directory
func LoadToken(config *Config, name string) (*TokenInfo, error) {
	// Create the token file path
	tokenFile := filepath.Join(config.TokensDir, name+".json")

	// Check if the token file exists
	if _, err := os.Stat(tokenFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("token file does not exist: %s", name)
		}
		return nil, fmt.Errorf("error checking token file: %w", err)
	}

	// Read the token file
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("error reading token file: %w", err)
	}

	// Parse the token
	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("error parsing token file: %w", err)
	}

	return &token, nil
}

// IsTokenValid checks if a token is valid
func IsTokenValid(token *TokenInfo) bool {
	// Add a buffer of 5 minutes to avoid edge cases
	return token != nil && time.Now().Add(5*time.Minute).Before(token.ExpiresAt)
}

// LoginCommand handles the login command
func LoginCommand(args []string) error {
	// Load the configuration
	config, err := LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Create a new auth client
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	authClient := sdkConfig.NewAuthClient()

	// Check if we already have a valid token
	token, err := LoadToken(config, DefaultTokenFile)
	if err == nil && IsTokenValid(token) {
		fmt.Println("Already logged in with a valid token.")
		return nil
	}

	// Set up a local server to handle the OAuth callback
	authCode := make(chan string, 1)
	authErr := make(chan error, 1)

	// Start a local server to handle the callback
	server := &http.Server{Addr: ":8888"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization code from the query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			authErr <- fmt.Errorf("no authorization code in callback")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: No authorization code received")
			return
		}

		// Send the code to the channel
		authCode <- code

		// Send a success response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Authorization successful! You can close this window and return to the CLI.")
	})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			authErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Set the redirect URL
	authClient.SetRedirectURL("http://localhost:8888/callback")

	// Get the authorization URL
	authURL := authClient.GetAuthorizationURL("state", pkg.GetScopesByService("auth", "transfer", "groups")...)

	// Print the URL for the user to open
	fmt.Println("Please open the following URL in your browser:")
	fmt.Println(authURL)
	fmt.Println("
Waiting for authorization...")

	// Wait for the authorization code or an error
	var code string
	select {
	case code = <-authCode:
		// Continue with the token exchange
	case err := <-authErr:
		return fmt.Errorf("authorization error: %w", err)
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authorization timed out after 5 minutes")
	}

	// Exchange the code for tokens
	fmt.Println("Exchanging authorization code for tokens...")
	tokenResp, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
	if err != nil {
		return fmt.Errorf("error exchanging code for tokens: %w", err)
	}

	// Convert to our token format
	token = &TokenInfo{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiryTime,
	}

	// Store the token
	if err := StoreToken(config, DefaultTokenFile, token); err != nil {
		return fmt.Errorf("error storing token: %w", err)
	}

	// Close the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	fmt.Println("Login successful!")
	printTokenInfo(token)

	return nil
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
		fmt.Println("Not logged in.")
		return nil
	}

	// Create a new auth client
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	authClient := sdkConfig.NewAuthClient()

	// Revoke the access token
	fmt.Println("Revoking access token...")
	if err := authClient.RevokeToken(context.Background(), token.AccessToken); err != nil {
		fmt.Printf("Warning: Failed to revoke access token: %v
", err)
	}

	// Revoke the refresh token
	if token.RefreshToken != "" {
		fmt.Println("Revoking refresh token...")
		if err := authClient.RevokeToken(context.Background(), token.RefreshToken); err != nil {
			fmt.Printf("Warning: Failed to revoke refresh token: %v
", err)
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

// printTokenInfo prints information about a token
func printTokenInfo(token *TokenInfo) error {
	fmt.Println("Token Information:")
	fmt.Printf("  Access Token: %s...%s
", token.AccessToken[:10], token.AccessToken[len(token.AccessToken)-10:])
	
	if token.RefreshToken != "" {
		fmt.Printf("  Refresh Token: %s...%s
", token.RefreshToken[:10], token.RefreshToken[len(token.RefreshToken)-10:])
	}
	
	fmt.Printf("  Expires At: %s
", token.ExpiresAt.Format(time.RFC3339))
	
	if IsTokenValid(token) {
		fmt.Printf("  Status: Valid (expires in %s)
", time.Until(token.ExpiresAt).Round(time.Second))
	} else {
		fmt.Println("  Status: Expired")
	}
	
	return nil
}

// revokeToken revokes a token
func revokeToken(config *Config, token *TokenInfo, args []string) error {
	// Create a new auth client
	sdkConfig := pkg.NewConfig().
		WithClientID(config.ClientID).
		WithClientSecret(config.ClientSecret)

	authClient := sdkConfig.NewAuthClient()

	// Check if we have a token type to revoke
	if len(args) > 0 {
		switch args[0] {
		case "access":
			fmt.Println("Revoking access token...")
			if err := authClient.RevokeToken(context.Background(), token.AccessToken); err != nil {
				return fmt.Errorf("error revoking access token: %w", err)
			}
			fmt.Println("Access token revoked successfully.")
			return nil
		case "refresh":
			if token.RefreshToken == "" {
				return fmt.Errorf("no refresh token available")
			}
			fmt.Println("Revoking refresh token...")
			if err := authClient.RevokeToken(context.Background(), token.RefreshToken); err != nil {
				return fmt.Errorf("error revoking refresh token: %w", err)
			}
			fmt.Println("Refresh token revoked successfully.")
			return nil
		default:
			return fmt.Errorf("unknown token type: %s", args[0])
		}
	}

	// Default to revoking both tokens
	return LogoutCommand(nil)
}
