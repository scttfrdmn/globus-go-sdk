// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build standalone

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	
	"github.com/joho/godotenv"
)

// This file contains a standalone utility for verifying Globus credentials
// It doesn't rely on the SDK to avoid build issues while fixing import cycles

const (
	authBaseURL = "https://auth.globus.org/v2/"
	transferBaseURL = "https://transfer.api.globus.org/v0.10/"
	groupsBaseURL = "https://groups.api.globus.org/v2/"
	searchBaseURL = "https://search.api.globus.org/v1/"
)

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// TokenInfo represents token introspection response
type TokenInfo struct {
	Active     bool     `json:"active"`
	Scope      string   `json:"scope,omitempty"`
	ClientID   string   `json:"client_id,omitempty"`
	Username   string   `json:"username,omitempty"`
	Exp        int64    `json:"exp,omitempty"`
	Sub        string   `json:"sub,omitempty"`
	Aud        []string `json:"aud,omitempty"`
	Iss        string   `json:"iss,omitempty"`
	Identity   string   `json:"identity,omitempty"`
	Name       string   `json:"name,omitempty"`
	Email      string   `json:"email,omitempty"`
	Nbf        int64    `json:"nbf,omitempty"`
	Iat        int64    `json:"iat,omitempty"`
	Jti        string   `json:"jti,omitempty"`
	TokenType  string   `json:"token_type,omitempty"`
	TokenClass string   `json:"token_class,omitempty"`
}

// EndpointInfo represents Globus endpoint information
type EndpointInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Owner       string `json:"owner_string"`
	Activated   bool   `json:"activated"`
}

// GroupInfo represents Globus group information
type GroupInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

// main is the entry point for the standalone credential verification tool
func main() {
	// Try multiple locations for .env.test file
	err1 := godotenv.Load("../../.env.test") // When run from cmd/verify-credentials
	err2 := godotenv.Load("./.env.test")     // When run from project root
	err3 := godotenv.Load(".env.test")       // Fallback
	
	if err1 != nil && err2 != nil && err3 != nil {
		fmt.Println("Warning: No .env.test file found, using environment variables")
		fmt.Println("Create a .env.test file with GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	} else {
		fmt.Println("Loaded environment variables from .env.test file")
	}
	
	// Get required credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	
	if clientID == "" || clientSecret == "" {
		log.Fatal("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}
	
	fmt.Println("✅ Found required credentials")
	
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Verify Auth service
	fmt.Println("\nVerifying Auth service...")
	token, err := getClientCredentialsToken(client, clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to get client credentials token: %v", err)
	}
	
	fmt.Printf("✅ Successfully obtained client credentials token\n")
	
	// Verify token introspection
	tokenInfo, err := introspectToken(client, clientID, clientSecret, token.AccessToken)
	if err != nil {
		log.Fatalf("Failed to introspect token: %v", err)
	}
	
	if !tokenInfo.Active {
		log.Fatal("Token is not active")
	}
	
	fmt.Println("✅ Token introspection successful")
	fmt.Printf("   Token scopes: %s\n", tokenInfo.Scope)
	
	// Check for transfer endpoints if specified
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DESTINATION_ENDPOINT_ID")
	
	if sourceEndpointID != "" {
		fmt.Println("\nVerifying Transfer service...")
		
		// Get new token with transfer scope
		fmt.Println("  Getting token with transfer scope...")
		transferToken, err := getClientCredentialsToken(client, clientID, clientSecret, "urn:globus:auth:scope:transfer.api.globus.org:all")
		if err != nil {
			fmt.Printf("❌ Failed to get transfer token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for transfer scope")
			fmt.Println("  Integration tests that use transfer will need tokens from another flow")
		} else {
			// Check source endpoint
			sourceEndpoint, err := getEndpoint(client, transferToken.AccessToken, sourceEndpointID)
			if err != nil {
				fmt.Printf("❌ Failed to get source endpoint: %v\n", err)
			} else {
				fmt.Printf("✅ Source endpoint accessed: %s (owner: %s)\n", 
					sourceEndpoint.DisplayName, sourceEndpoint.Owner)
				
				// Check destination endpoint if specified
				if destEndpointID != "" {
					destEndpoint, err := getEndpoint(client, transferToken.AccessToken, destEndpointID)
					if err != nil {
						fmt.Printf("❌ Failed to get destination endpoint: %v\n", err)
					} else {
						fmt.Printf("✅ Destination endpoint accessed: %s (owner: %s)\n", 
							destEndpoint.DisplayName, destEndpoint.Owner)
					}
				}
			}
		}
	}
	
	// Check group if specified
	groupID := os.Getenv("GLOBUS_TEST_GROUP_ID")
	if groupID != "" {
		fmt.Println("\nVerifying Groups service...")
		
		// Get new token with groups scope
		fmt.Println("  Getting token with groups scope...")
		groupsToken, err := getClientCredentialsToken(client, clientID, clientSecret, "urn:globus:auth:scope:groups.api.globus.org:all")
		if err != nil {
			fmt.Printf("❌ Failed to get groups token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for groups scope")
			fmt.Println("  Integration tests that use groups will need tokens from another flow")
		} else {
			// Check group
			group, err := getGroup(client, groupsToken.AccessToken, groupID)
			if err != nil {
				fmt.Printf("❌ Failed to get group: %v\n", err)
			} else {
				fmt.Printf("✅ Group accessed: %s (owner: %s)\n", 
					group.Name, group.Owner)
			}
		}
	}
	
	// Check for search index if specified
	searchIndexID := os.Getenv("GLOBUS_TEST_SEARCH_INDEX_ID")
	if searchIndexID != "" {
		fmt.Println("\nVerifying Search service...")
		
		// Get new token with search scope
		fmt.Println("  Getting token with search scope...")
		searchToken, err := getClientCredentialsToken(client, clientID, clientSecret, "urn:globus:auth:scope:search.api.globus.org:all")
		if err != nil {
			fmt.Printf("❌ Failed to get search token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for search scope")
			fmt.Println("  Integration tests that use search will need tokens from another flow")
		} else {
			// Check search index with a basic query
			fmt.Printf("  Testing access to search index %s...\n", searchIndexID)
			ok, err := searchIndex(client, searchToken.AccessToken, searchIndexID)
			if err != nil {
				fmt.Printf("❌ Failed to query search index: %v\n", err)
			} else if !ok {
				fmt.Printf("❌ Search index query returned no results\n")
			} else {
				fmt.Printf("✅ Search index %s accessed successfully\n", searchIndexID)
			}
		}
	}
	
	fmt.Println("\n✨ Success! Your Globus credentials are valid.")
	fmt.Println("   The client credentials can be used for the Auth service.")
	fmt.Println("   Other services may require different authentication flows.")
}

// getClientCredentialsToken gets an OAuth token using client credentials
func getClientCredentialsToken(client *http.Client, clientID, clientSecret string, scopes ...string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}
	
	req, err := http.NewRequest("POST", authBaseURL+"oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s (%s)", resp.Status, string(body))
	}
	
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	
	return &tokenResp, nil
}

// introspectToken performs token introspection
func introspectToken(client *http.Client, clientID, clientSecret, token string) (*TokenInfo, error) {
	data := url.Values{}
	data.Set("token", token)
	
	req, err := http.NewRequest("POST", authBaseURL+"oauth2/token/introspect", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s (%s)", resp.Status, string(body))
	}
	
	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}
	
	return &tokenInfo, nil
}

// getEndpoint gets information about a Globus endpoint
func getEndpoint(client *http.Client, token, endpointID string) (*EndpointInfo, error) {
	req, err := http.NewRequest("GET", transferBaseURL+"endpoint/"+endpointID, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s (%s)", resp.Status, string(body))
	}
	
	var endpoint EndpointInfo
	if err := json.NewDecoder(resp.Body).Decode(&endpoint); err != nil {
		return nil, err
	}
	
	return &endpoint, nil
}

// getGroup gets information about a Globus group
func getGroup(client *http.Client, token, groupID string) (*GroupInfo, error) {
	req, err := http.NewRequest("GET", groupsBaseURL+"groups/"+groupID, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s (%s)", resp.Status, string(body))
	}
	
	var group GroupInfo
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}
	
	return &group, nil
}

// SearchResponse represents a simplified response from the Search API
type SearchResponse struct {
	Count   int                      `json:"count"`
	Total   int                      `json:"total"`
	Subjects []map[string]interface{} `json:"subjects"`
}

// searchIndex searches a Globus search index with a basic query
func searchIndex(client *http.Client, token, indexID string) (bool, error) {
	// Construct a simple query to test access
	query := map[string]interface{}{
		"q": "*",      // Simple wildcard query to match all documents
		"limit": 1,    // Only need one result to verify access
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", searchBaseURL+"index/"+indexID+"/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("bad status: %s (%s)", resp.Status, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return false, err
	}

	// Consider successful if we got any results or at least count is 0 with no error
	// (which means the index exists but might be empty)
	return searchResp.Count > 0 || (searchResp.Count == 0 && searchResp.Total >= 0), nil
}