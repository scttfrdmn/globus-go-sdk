// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
)

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment()

	// Create a new Groups client with an access token
	groupsClient, err := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create groups client: %v", err)
	}

	// List groups the user is a member of
	groupList, err := groupsClient.ListGroups(context.Background(), &groups.ListGroupsOptions{
		MyGroups: true,
		PageSize: 100,
	})
	if err != nil {
		log.Fatalf("Failed to list groups: %v", err)
	}

	fmt.Printf("You are a member of %d groups:\n", len(groupList.Groups))
	for _, group := range groupList.Groups {
		fmt.Printf("- %s (%s)\n", group.Name, group.ID)
	}

	// Create a new group
	newGroup := &groups.GroupCreate{
		Name:        "Test Group",
		Description: "A test group created using the Globus Go SDK",
		PublicGroup: true,
	}

	createdGroup, err := groupsClient.CreateGroup(context.Background(), newGroup)
	if err != nil {
		log.Fatalf("Failed to create group: %v", err)
	}

	fmt.Printf("\nCreated group: %s (%s)\n", createdGroup.Name, createdGroup.ID)

	// Add a member to the group
	err = groupsClient.AddMember(context.Background(), createdGroup.ID, "user@example.com", "member")
	if err != nil {
		log.Fatalf("Failed to add member: %v", err)
	}

	fmt.Println("Added member to group.")
}
