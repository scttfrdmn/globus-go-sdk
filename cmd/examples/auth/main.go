// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
)

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create a new Auth client
	authClient := config.NewAuthClient()
	authClient.SetRedirectURL("http://localhost:8080/callback")

	// Get authorization URL
	authURL := authClient.GetAuthorizationURL("my-state")
	fmt.Printf("Visit this URL to log in: %s
", authURL)

	// Start a local server to handle the callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		
		// Exchange code for tokens
		tokenResponse, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
		if err != nil {
			log.Fatalf("Failed to exchange code: %v", err)
		}
		
		fmt.Printf("
Access Token: %s
", tokenResponse.AccessToken)
		fmt.Printf("Refresh Token: %s
", tokenResponse.RefreshToken)
		fmt.Printf("Expires In: %d seconds
", tokenResponse.ExpiresIn)
		
		fmt.Fprintf(w, "Authentication successful! You can close this window.")
	})
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
