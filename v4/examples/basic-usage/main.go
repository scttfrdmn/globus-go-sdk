// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package main demonstrates basic usage of the Globus Go SDK v4
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
)

func main() {
	// Get access token from environment
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_ACCESS_TOKEN environment variable is required")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example 1: Using the Auth client
	fmt.Println("=== Auth Client Example ===")
	authExample(ctx, accessToken)

	// Example 2: Using the Groups client
	fmt.Println("\n=== Groups Client Example ===")
	groupsExample(ctx, accessToken)
}

func authExample(ctx context.Context, accessToken string) {
	// In v4, you must explicitly specify scopes for security
	config := &core.Config{
		AccessToken: accessToken,
		Scopes: []string{
			core.Scopes.AuthOpenID,
			core.Scopes.AuthEmail,
			core.Scopes.AuthProfile,
		},
	}

	// Create auth client
	client, err := auth.NewClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create auth client: %v", err)
		return
	}

	// Get user info - context is always the first parameter in v4
	userInfo, err := client.GetUserInfo(ctx)
	if err != nil {
		// v4 has enhanced error types
		if apiErr, ok := err.(*core.APIError); ok {
			log.Printf("API error (HTTP %d): %s", apiErr.StatusCode, apiErr.Message)
			if apiErr.RequestID != "" {
				log.Printf("Request ID: %s", apiErr.RequestID)
			}
		} else {
			log.Printf("Error: %v", err)
		}
		return
	}

	fmt.Printf("User: %s (%s)\n", userInfo.Name, userInfo.Email)
	fmt.Printf("Username: %s\n", userInfo.PreferredUsername)
}

func groupsExample(ctx context.Context, accessToken string) {
	// Configure groups client with explicit scopes
	config := &core.Config{
		AccessToken: accessToken,
		Scopes: []string{
			core.Scopes.GroupsView,
		},
	}

	// Create groups client
	client, err := groups.NewClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create groups client: %v", err)
		return
	}

	// List groups - context is always the first parameter in v4
	options := &groups.ListGroupsOptions{
		MyGroups: true,
		PageSize: 10,
	}

	groupList, err := client.ListGroups(ctx, options)
	if err != nil {
		if apiErr, ok := err.(*core.APIError); ok {
			log.Printf("API error (HTTP %d): %s", apiErr.StatusCode, apiErr.Message)
			if apiErr.IsAuthError() {
				log.Printf("Authentication failed - check your token and scopes")
			}
		} else {
			log.Printf("Error: %v", err)
		}
		return
	}

	fmt.Printf("Found %d groups:\n", len(groupList.Groups))
	for i, group := range groupList.Groups {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, group.Name, group.ID)
		fmt.Printf("   Members: %d, Admin: %v\n", group.MemberCount, group.IsGroupAdmin)
	}

	if groupList.HasNextPage {
		fmt.Printf("   (more groups available with next_page_token: %s)\n", groupList.NextPageToken)
	}
}
