// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	// Get credentials from environment
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpoint := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpoint := os.Getenv("GLOBUS_TEST_DESTINATION_ENDPOINT_ID")

	if clientID == "" || clientSecret == "" {
		log.Fatal("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	fmt.Println("🔑 Testing Globus credentials...")
	fmt.Printf("Client ID: %s\n", clientID)
	fmt.Println("Client Secret: [REDACTED]")

	// Get token with client credentials
	token, err := getClientCredentialsToken(client, clientID, clientSecret)
	if err != nil {
		log.Fatalf("❌ Failed to get token: %v", err)
	}
	fmt.Println("✅ Successfully obtained access token")

	// Test Auth API
	introspectResult, err := introspectToken(client, token, clientID, clientSecret)
	if err != nil {
		log.Fatalf("❌ Failed to introspect token: %v", err)
	}
	fmt.Printf("✅ Token introspection successful: active=%v, expires in %v\n",
		introspectResult.Active,
		time.Until(time.Unix(introspectResult.Exp, 0)))

	// List all endpoints
	fmt.Println("\n🔄 Testing Transfer service...")
	endpoints, err := listEndpoints(client, token)
	if err != nil {
		log.Fatalf("❌ Failed to list endpoints: %v", err)
	}
	fmt.Printf("✅ Found %d endpoints\n", len(endpoints.DATA))

	// Display endpoints if available
	if len(endpoints.DATA) > 0 {
		fmt.Println("\nYour endpoints:")
		for i, ep := range endpoints.DATA {
			if i >= 5 {
				fmt.Println("... (more endpoints available)")
				break
			}
			fmt.Printf("  - %s (%s)\n", ep.DisplayName, ep.ID)
		}
	}

	// Test endpoint info if specified
	if sourceEndpoint != "" {
		fmt.Printf("\nSource endpoint: %s\n", sourceEndpoint)
		fmt.Println("⚠️ Note: Your client may not have direct access to this endpoint.")
		fmt.Println("   This is normal for service accounts or endpoints you don't own.")
	}

	if destEndpoint != "" {
		fmt.Printf("\nDestination endpoint: %s\n", destEndpoint)
		fmt.Println("⚠️ Note: Your client may not have direct access to this endpoint.")
		fmt.Println("   This is normal for service accounts or endpoints you don't own.")
	}

	// Test other services
	fmt.Println("\n🔍 Testing other Globus services...")

	// Test Groups service
	groupScopes := "urn:globus:auth:scope:groups.api.globus.org:all"
	_, err = getClientCredentialsToken(client, clientID, clientSecret, groupScopes)
	if err != nil {
		fmt.Printf("⚠️ Groups API access not available: %v\n", err)
	} else {
		fmt.Println("✅ Successfully obtained Groups service access token")
	}

	// Test Search service
	searchScopes := "urn:globus:auth:scope:search.api.globus.org:all"
	_, err = getClientCredentialsToken(client, clientID, clientSecret, searchScopes)
	if err != nil {
		fmt.Printf("⚠️ Search API access not available: %v\n", err)
	} else {
		fmt.Println("✅ Successfully obtained Search service access token")
	}

	fmt.Println("\n✨ All tests completed successfully!")
	fmt.Println("\n🔐 Your Globus credentials are working correctly for client credential flows.")
	fmt.Println("   This confirms you can use the credential for automated testing.")
}

func getClientCredentialsToken(client *http.Client, clientID, clientSecret string, extraScopes ...string) (string, error) {
	// Prepare request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Default to transfer scope, but allow additional scopes
	scopes := []string{"urn:globus:auth:scope:transfer.api.globus.org:all"}
	scopes = append(scopes, extraScopes...)

	data.Set("scope", strings.Join(scopes, " "))

	// Create request
	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

type TokenInfo struct {
	Active   bool   `json:"active"`
	Scope    string `json:"scope"`
	ClientID string `json:"client_id"`
	UserName string `json:"username"`
	Exp      int64  `json:"exp"`
	Iss      string `json:"iss"`
	Sub      string `json:"sub"`
}

func introspectToken(client *http.Client, token, clientID, clientSecret string) (*TokenInfo, error) {
	// Prepare request body
	data := url.Values{}
	data.Set("token", token)

	// Create request
	req, err := http.NewRequest("POST", "https://auth.globus.org/v2/oauth2/token/introspect",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspect request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var result TokenInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type EndpointInfo struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	Owner           string `json:"owner_string"`
	IsGlobusConnect bool   `json:"is_globus_connect"`
}

type EndpointList struct {
	DATA []EndpointInfo `json:"DATA"`
}

func listEndpoints(client *http.Client, token string) (*EndpointList, error) {
	// Create request
	req, err := http.NewRequest("GET",
		"https://transfer.api.globus.org/v0.10/endpoint_search?filter_scope=my-endpoints", nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list endpoints request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var result EndpointList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
