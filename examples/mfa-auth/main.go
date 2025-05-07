// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

const (
	port            = 8080
	callbackPath    = "/callback"
	callbackAddress = "http://localhost:8080/callback"
)

// This example demonstrates how to handle Multi-Factor Authentication (MFA)
// when authenticating with Globus Auth.
func main() {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get client ID from environment
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	if clientID == "" {
		log.Fatal("GLOBUS_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Create SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Create an Auth client
	authClient := config.NewAuthClient()
	authClient.SetRedirectURL(callbackAddress)

	// Generate a random state value
	state := fmt.Sprintf("state-%d", time.Now().UnixNano())

	// Get the authorization URL
	authURL := authClient.GetAuthorizationURL(state,
		pkg.AuthScope,     // Basic authentication
		pkg.TransferScope, // Transfer service access
		pkg.GroupsScope,   // Groups service access
	)

	fmt.Printf("Please visit the following URL to log in:\n\n%s\n\n", authURL)
	fmt.Println("After logging in, you'll be redirected to the callback URL.")

	// Set up a channel to receive the authorization code
	codeChan := make(chan string)
	errChan := make(chan error)

	// Start an HTTP server to handle the callback
	server := startCallbackServer(codeChan, errChan, state)
	defer server.Close()

	// Wait for the code or an error
	var code string
	select {
	case code = <-codeChan:
		fmt.Println("Authorization code received!")
	case err := <-errChan:
		log.Fatalf("Error during authorization: %v", err)
	case <-ctx.Done():
		log.Fatalf("Timed out waiting for authorization")
	}

	// Exchange the code for tokens
	fmt.Println("Exchanging code for tokens (this may require MFA)...")
	tokenResp, err := authClient.ExchangeAuthorizationCodeWithMFA(ctx, code, mfaHandler)
	if err != nil {
		log.Fatalf("Failed to exchange code: %v", err)
	}

	// Print token information
	fmt.Println("\nAuthentication successful!")
	fmt.Printf("Access Token: %s...(truncated)\n", tokenResp.AccessToken[:10])
	fmt.Printf("Token Type: %s\n", tokenResp.TokenType)
	fmt.Printf("Expires In: %d seconds\n", tokenResp.ExpiresIn)
	fmt.Printf("Scopes: %s\n", tokenResp.Scope)

	// Demonstrate token refreshing with MFA
	if tokenResp.RefreshToken != "" {
		fmt.Println("\nDemonstrating token refresh (may require MFA again)...")
		refreshedResp, err := authClient.RefreshTokenWithMFA(ctx, tokenResp.RefreshToken, mfaHandler)
		if err != nil {
			fmt.Printf("Token refresh failed: %v\n", err)
		} else {
			fmt.Println("Token refresh successful!")
			fmt.Printf("New Access Token: %s...(truncated)\n", refreshedResp.AccessToken[:10])
			fmt.Printf("Expires In: %d seconds\n", refreshedResp.ExpiresIn)
		}
	}
}

// startCallbackServer starts an HTTP server to handle the OAuth2 callback
func startCallbackServer(codeChan chan string, errChan chan error, expectedState string) *http.Server {
	// Create a server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	// Set up the handler
	http.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// Get the code and state from the query parameters
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		error := r.URL.Query().Get("error")

		// Check for errors
		if error != "" {
			errDesc := r.URL.Query().Get("error_description")
			errChan <- fmt.Errorf("%s: %s", error, errDesc)
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
			return
		}

		// Validate state
		if state != expectedState {
			errChan <- fmt.Errorf("invalid state: got %s, expected %s", state, expectedState)
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		// Send the code to the channel
		codeChan <- code

		// Return a success message
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
				<body>
					<h1>Authentication Successful</h1>
					<p>You can close this window now.</p>
				</body>
			</html>
		`)
	})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			errChan <- err
		}
	}()

	fmt.Printf("Callback server started on http://localhost:%d%s\n", port, callbackPath)
	return server
}

// mfaHandler handles MFA challenges
func mfaHandler(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
	if challenge == nil {
		return nil, fmt.Errorf("received nil MFA challenge")
	}

	fmt.Printf("\nMulti-Factor Authentication Required\n")
	fmt.Printf("Challenge ID: %s\n", challenge.ChallengeID)
	fmt.Printf("Type: %s\n", challenge.Type)
	fmt.Printf("Prompt: %s\n", challenge.Prompt)

	fmt.Printf("Allowed types: %s\n\n", strings.Join(challenge.AllowedTypes, ", "))

	// Ask the user for the MFA code
	fmt.Print("Enter your MFA code: ")
	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read MFA code: %w", err)
	}

	// Clean up the code
	code = strings.TrimSpace(code)

	// Create the response
	response := &auth.MFAResponse{
		ChallengeID: challenge.ChallengeID,
		Type:        challenge.Type, // Use the same type as the challenge
		Value:       code,
	}

	return response, nil
}
