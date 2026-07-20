// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Error: GLOBUS_ACCESS_TOKEN environment variable is required")
		os.Exit(1)
	}

	// Create the authorizer
	authorizer := authorizers.StaticTokenCoreAuthorizer(accessToken)

	// Create the Groups client
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithHTTPDebugging(true),
	)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// List groups
	groupList, err := client.GetMyGroups(context.Background(), []string{"active"})
	if err != nil {
		fmt.Printf("Error listing groups: %v\n", err)
		os.Exit(1)
	}

	// Print groups
	fmt.Printf("Found %d groups:\n", len(groupList.Groups))
	for i, group := range groupList.Groups {
		fmt.Printf("%d. %s (ID: %s, DATA_TYPE: %s)\n", i+1, group.Name, group.ID, group.DATA_TYPE)
	}

	fmt.Printf("Total groups: %d\n", len(groupList.Groups))
}
