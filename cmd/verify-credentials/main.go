// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

// main is a simple wrapper that calls the appropriate verification function
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

	// Check credentials
	fmt.Println("Verifying Globus credentials...")
	err := VerifyCredentialsSDK()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\n✨ Success! Your Globus credentials are valid.")
	fmt.Println("   The client credentials can be used for the Auth service.")
	fmt.Println("   Other services may require different authentication flows.")
}

// VerifyCredentialsSDK checks that the provided Globus credentials are valid
// using the SDK implementation
func VerifyCredentialsSDK() error {
	// Get required credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}

	fmt.Println("✅ Found required credentials")

	// Create SDK config and client
	config := pkg.NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Create auth client
	authClient := config.NewAuthClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify Auth service by getting a client credentials token
	fmt.Println("\nVerifying Auth service...")
	token, err := authClient.GetClientCredentialsToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client credentials token: %v", err)
	}

	fmt.Printf("✅ Successfully obtained client credentials token\n")

	// Verify token introspection
	tokenInfo, err := authClient.IntrospectToken(ctx, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to introspect token: %v", err)
	}

	if !tokenInfo.Active {
		return fmt.Errorf("token is not active")
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
		transferToken, err := authClient.GetClientCredentialsToken(ctx, pkg.TransferScope)
		if err != nil {
			fmt.Printf("❌ Failed to get transfer token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for transfer scope")
			fmt.Println("  Integration tests that use transfer will need tokens from another flow")
		} else {
			// Create transfer client
			transferClient := config.NewTransferClient(transferToken.AccessToken)

			// Check source endpoint
			sourceEndpoint, err := transferClient.GetEndpoint(ctx, sourceEndpointID)
			if err != nil {
				fmt.Printf("❌ Failed to get source endpoint: %v\n", err)
			} else {
				fmt.Printf("✅ Source endpoint accessed: %s (owner: %s)\n",
					sourceEndpoint.DisplayName, sourceEndpoint.OwnerString)

				// Check destination endpoint if specified
				if destEndpointID != "" {
					destEndpoint, err := transferClient.GetEndpoint(ctx, destEndpointID)
					if err != nil {
						fmt.Printf("❌ Failed to get destination endpoint: %v\n", err)
					} else {
						fmt.Printf("✅ Destination endpoint accessed: %s (owner: %s)\n",
							destEndpoint.DisplayName, destEndpoint.OwnerString)
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
		groupsToken, err := authClient.GetClientCredentialsToken(ctx, pkg.GroupsScope)
		if err != nil {
			fmt.Printf("❌ Failed to get groups token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for groups scope")
			fmt.Println("  Integration tests that use groups will need tokens from another flow")
		} else {
			// Create groups client
			groupsClient := config.NewGroupsClient(groupsToken.AccessToken)

			// Check group
			group, err := groupsClient.GetGroup(ctx, groupID)
			if err != nil {
				fmt.Printf("❌ Failed to get group: %v\n", err)
			} else {
				fmt.Printf("✅ Group accessed: %s (owner ID: %s)\n",
					group.Name, group.IdentityID)
			}
		}
	}

	// Check for search index if specified
	searchIndexID := os.Getenv("GLOBUS_TEST_SEARCH_INDEX_ID")
	if searchIndexID != "" {
		fmt.Println("\nVerifying Search service...")

		// Get new token with search scope
		fmt.Println("  Getting token with search scope...")
		searchToken, err := authClient.GetClientCredentialsToken(ctx, pkg.SearchScope)
		if err != nil {
			fmt.Printf("❌ Failed to get search token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for search scope")
			fmt.Println("  Integration tests that use search will need tokens from another flow")
		} else {
			// Create search client
			searchClient := config.NewSearchClient(searchToken.AccessToken)

			// Check search index
			// Note: Using a simple query to test the index access
			fmt.Printf("  Testing access to search index %s...\n", searchIndexID)
			searchRequest := &search.SearchRequest{
				IndexID: searchIndexID,
				Query:   "*", // Simple wildcard query to match all documents
				Options: &search.SearchOptions{
					Limit: 1, // Only need one result to verify access
				},
			}
			_, err := searchClient.Search(ctx, searchRequest)
			if err != nil {
				fmt.Printf("❌ Failed to query search index: %v\n", err)
			} else {
				fmt.Printf("✅ Search index %s accessed successfully\n", searchIndexID)
			}
		}
	}

	return nil
}