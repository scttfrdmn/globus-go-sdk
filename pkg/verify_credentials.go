// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pkg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

// VerifyCredentials checks that the provided Globus credentials are valid
// and have the necessary permissions to access the specified services.
// It looks for credentials in environment variables:
// - GLOBUS_TEST_CLIENT_ID: The client ID to verify
// - GLOBUS_TEST_CLIENT_SECRET: The client secret to verify
// - GLOBUS_TEST_SOURCE_ENDPOINT_ID: (Optional) A source endpoint ID to verify transfer permissions
// - GLOBUS_TEST_DESTINATION_ENDPOINT_ID: (Optional) A destination endpoint ID to verify transfer permissions
// - GLOBUS_TEST_GROUP_ID: (Optional) A group ID to verify groups permissions
// - GLOBUS_TEST_SEARCH_INDEX_ID: (Optional) A search index ID to verify search permissions
func VerifyCredentials() error {
	// Get required credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}

	fmt.Println("✅ Found required credentials")

	// Create SDK config and client
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Create auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

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
		transferToken, err := authClient.GetClientCredentialsToken(ctx, TransferScope)
		if err != nil {
			fmt.Printf("❌ Failed to get transfer token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for transfer scope")
			fmt.Println("  Integration tests that use transfer will need tokens from another flow")
		} else {
			// Create transfer client
			transferClient, err := config.NewTransferClient(transferToken.AccessToken)
			if err != nil {
				fmt.Printf("❌ Failed to create transfer client: %v\n", err)
				return fmt.Errorf("failed to create transfer client: %w", err)
			}

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
		groupsToken, err := authClient.GetClientCredentialsToken(ctx, GroupsScope)
		if err != nil {
			fmt.Printf("❌ Failed to get groups token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for groups scope")
			fmt.Println("  Integration tests that use groups will need tokens from another flow")
		} else {
			// Create groups client
			groupsClient, err := config.NewGroupsClient(groupsToken.AccessToken)
			if err != nil {
				fmt.Printf("❌ Failed to create groups client: %v\n", err)
				return fmt.Errorf("failed to create groups client: %w", err)
			}

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
		searchToken, err := authClient.GetClientCredentialsToken(ctx, SearchScope)
		if err != nil {
			fmt.Printf("❌ Failed to get search token: %v\n", err)
			fmt.Println("  This is ok - client credentials flow may not be enabled for search scope")
			fmt.Println("  Integration tests that use search will need tokens from another flow")
		} else {
			// Create search client
			searchClient, err := config.NewSearchClient(searchToken.AccessToken)
			if err != nil {
				fmt.Printf("❌ Failed to create search client: %v\n", err)
				return fmt.Errorf("failed to create search client: %w", err)
			}

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
			_, searchErr := searchClient.Search(ctx, searchRequest)
			if searchErr != nil {
				fmt.Printf("❌ Failed to query search index: %v\n", searchErr)
			} else {
				fmt.Printf("✅ Search index %s accessed successfully\n", searchIndexID)
			}
		}
	}

	return nil
}
