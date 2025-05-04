// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create a new Auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Get token using client credentials for simplicity
	// In a real application, you would likely use the authorization code flow
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.SearchScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)
	accessToken := tokenResp.AccessToken

	// Create Search client
	searchClient := config.NewSearchClient(accessToken)

	// Check if index ID is provided
	indexID := os.Getenv("GLOBUS_SEARCH_INDEX_ID")
	if indexID == "" {
		// List available indexes if no index ID is provided
		fmt.Println("\n=== Available Indexes ===")
		
		indexes, err := searchClient.ListIndexes(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to list indexes: %v", err)
		}
		
		if len(indexes.Indexes) == 0 {
			fmt.Println("No indexes found. Create an index first.")
			
			// Create a new index
			fmt.Println("\n=== Creating New Index ===")
			
			timestamp := time.Now().Format("20060102_150405")
			createRequest := &search.IndexCreateRequest{
				DisplayName: fmt.Sprintf("SDK Example Index %s", timestamp),
				Description: "An example index created by the Globus Go SDK",
			}
			
			newIndex, err := searchClient.CreateIndex(ctx, createRequest)
			if err != nil {
				log.Fatalf("Failed to create index: %v", err)
			}
			
			fmt.Printf("Created new index: %s (%s)\n", newIndex.DisplayName, newIndex.ID)
			indexID = newIndex.ID
		} else {
			// Use the first available index
			fmt.Printf("Found %d indexes:\n", len(indexes.Indexes))
			for i, index := range indexes.Indexes {
				fmt.Printf("%d. %s (%s)\n", i+1, index.DisplayName, index.ID)
			}
			
			indexID = indexes.Indexes[0].ID
			fmt.Printf("\nUsing first index: %s\n", indexID)
		}
	} else {
		// If index ID is provided, show details
		index, err := searchClient.GetIndex(ctx, indexID)
		if err != nil {
			log.Fatalf("Failed to get index: %v", err)
		}
		
		fmt.Printf("\n=== Index Details ===\n")
		fmt.Printf("ID: %s\n", index.ID)
		fmt.Printf("Name: %s\n", index.DisplayName)
		fmt.Printf("Description: %s\n", index.Description)
		fmt.Printf("Is Active: %t\n", index.IsActive)
		fmt.Printf("Is Public: %t\n", index.IsPublic)
	}

	// Ingest some sample documents
	fmt.Println("\n=== Ingesting Documents ===")
	
	timestamp := time.Now().Format("20060102_150405")
	documents := []search.SearchDocument{
		{
			Subject: fmt.Sprintf("example-doc-1-%s", timestamp),
			Content: map[string]interface{}{
				"title":       "Example Document 1",
				"description": "This is an example document created by the Globus Go SDK",
				"tags":        []string{"example", "sdk", "go"},
				"number":      42,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			VisibleTo: []string{"public"},
		},
		{
			Subject: fmt.Sprintf("example-doc-2-%s", timestamp),
			Content: map[string]interface{}{
				"title":       "Example Document 2",
				"description": "Another example document demonstrating the Globus Go SDK",
				"tags":        []string{"example", "sdk", "globus"},
				"number":      123,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			VisibleTo: []string{"public"},
		},
	}
	
	ingestRequest := &search.IngestRequest{
		IndexID:   indexID,
		Documents: documents,
	}
	
	ingestResponse, err := searchClient.IngestDocuments(ctx, ingestRequest)
	if err != nil {
		log.Fatalf("Failed to ingest documents: %v", err)
	}
	
	fmt.Printf("Ingest task ID: %s\n", ingestResponse.Task.TaskID)
	fmt.Printf("Documents: %d succeeded, %d failed, %d total\n",
		ingestResponse.Succeeded, ingestResponse.Failed, ingestResponse.Total)
	
	// Wait for indexing to complete
	fmt.Println("\nWaiting for indexing to complete...")
	time.Sleep(3 * time.Second)
	
	// Get task status
	taskStatus, err := searchClient.GetTaskStatus(ctx, ingestResponse.Task.TaskID)
	if err != nil {
		log.Printf("Failed to get task status: %v", err)
	} else {
		fmt.Printf("Task status: %s\n", taskStatus.State)
	}

	// Search for documents
	fmt.Println("\n=== Searching Documents ===")
	
	// First search using a general term
	searchRequest := &search.SearchRequest{
		IndexID: indexID,
		Query:   "example",
		Options: &search.SearchOptions{
			Limit: 10,
		},
	}
	
	searchResponse, err := searchClient.Search(ctx, searchRequest)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}
	
	fmt.Printf("Found %d documents for query 'example'\n", searchResponse.Count)
	
	if len(searchResponse.Results) > 0 {
		fmt.Println("\nResults:")
		for i, result := range searchResponse.Results {
			title := result.Content["title"]
			fmt.Printf("%d. %s (Subject: %s, Score: %.2f)\n", 
				i+1, title, result.Subject, result.Score)
		}
		
		// Print the first result as JSON for demonstration
		firstResult := searchResponse.Results[0]
		resultJSON, _ := json.MarshalIndent(firstResult, "", "  ")
		fmt.Printf("\nFirst result details:\n%s\n", resultJSON)
	} else {
		fmt.Println("No results found. The documents may still be indexing.")
	}

	// Search with a more specific query
	fmt.Println("\n=== Advanced Search ===")
	
	advancedRequest := &search.SearchRequest{
		IndexID: indexID,
		Query:   "tags:go",
		Options: &search.SearchOptions{
			Limit: 10,
		},
	}
	
	advancedResponse, err := searchClient.Search(ctx, advancedRequest)
	if err != nil {
		log.Printf("Failed to perform advanced search: %v", err)
	} else {
		fmt.Printf("Found %d documents for query 'tags:go'\n", advancedResponse.Count)
		
		if len(advancedResponse.Results) > 0 {
			fmt.Println("\nResults:")
			for i, result := range advancedResponse.Results {
				title := result.Content["title"]
				fmt.Printf("%d. %s (Subject: %s)\n", i+1, title, result.Subject)
			}
		}
	}

	fmt.Println("\nSearch example complete!")
}