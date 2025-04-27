// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

func getTestCredentials(t *testing.T) (string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	indexID := os.Getenv("GLOBUS_TEST_SEARCH_INDEX_ID")

	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}

	return clientID, clientSecret, indexID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	authClient := auth.NewClient(clientID, clientSecret)

	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), "urn:globus:auth:scope:search.api.globus.org:all")
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	return tokenResp.AccessToken
}

func TestIntegration_ListIndexes(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Search client
	client := NewClient(accessToken)
	ctx := context.Background()

	// List indexes
	indexes, err := client.ListIndexes(ctx, &ListIndexesOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListIndexes failed: %v", err)
	}

	// Verify we got some data
	t.Logf("Found %d indexes", len(indexes.Indexes))

	// The user might not have any indexes, so this isn't necessarily an error
	if len(indexes.Indexes) > 0 {
		// Check that the first index has expected fields
		firstIndex := indexes.Indexes[0]
		if firstIndex.ID == "" {
			t.Error("First index is missing ID")
		}
		if firstIndex.DisplayName == "" {
			t.Error("First index is missing display name")
		}
	}
}

func TestIntegration_IndexLifecycle(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Search client
	client := NewClient(accessToken)
	ctx := context.Background()

	// 1. Create a new index
	timestamp := time.Now().Format("20060102_150405")
	indexName := fmt.Sprintf("Test Index %s", timestamp)
	indexDescription := "A test index created by integration tests"

	createRequest := &IndexCreateRequest{
		DisplayName: indexName,
		Description: indexDescription,
	}

	createdIndex, err := client.CreateIndex(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Make sure the index gets deleted after the test
	defer func() {
		err := client.DeleteIndex(ctx, createdIndex.ID)
		if err != nil {
			t.Logf("Warning: Failed to delete test index (%s): %v", createdIndex.ID, err)
		} else {
			t.Logf("Successfully deleted test index (%s)", createdIndex.ID)
		}
	}()

	t.Logf("Created index: %s (%s)", createdIndex.DisplayName, createdIndex.ID)

	// 2. Verify the index was created correctly
	if createdIndex.DisplayName != indexName {
		t.Errorf("Created index name = %s, want %s", createdIndex.DisplayName, indexName)
	}
	if createdIndex.Description != indexDescription {
		t.Errorf("Created index description = %s, want %s", createdIndex.Description, indexDescription)
	}

	// 3. Get the index
	fetchedIndex, err := client.GetIndex(ctx, createdIndex.ID)
	if err != nil {
		t.Fatalf("Failed to get index: %v", err)
	}

	if fetchedIndex.ID != createdIndex.ID {
		t.Errorf("Fetched index ID = %s, want %s", fetchedIndex.ID, createdIndex.ID)
	}

	// 4. Update the index
	updatedDescription := "Updated description for integration test"
	updateRequest := &IndexUpdateRequest{
		Description: updatedDescription,
	}

	updatedIndex, err := client.UpdateIndex(ctx, createdIndex.ID, updateRequest)
	if err != nil {
		t.Fatalf("Failed to update index: %v", err)
	}

	if updatedIndex.Description != updatedDescription {
		t.Errorf("Updated index description = %s, want %s", updatedIndex.Description, updatedDescription)
	}

	// 5. Ingest and search documents
	documents := []SearchDocument{
		{
			Subject: fmt.Sprintf("test-doc-1-%s", timestamp),
			Content: map[string]interface{}{
				"title":       "Test Document 1",
				"description": "This is a test document for integration testing",
				"tags":        []string{"test", "integration", "document1"},
				"count":       1,
			},
			VisibleTo: []string{"public"},
		},
		{
			Subject: fmt.Sprintf("test-doc-2-%s", timestamp),
			Content: map[string]interface{}{
				"title":       "Test Document 2",
				"description": "Another test document for integration testing",
				"tags":        []string{"test", "integration", "document2"},
				"count":       2,
			},
			VisibleTo: []string{"public"},
		},
	}

	ingestRequest := &IngestRequest{
		IndexID:   createdIndex.ID,
		Documents: documents,
	}

	ingestResponse, err := client.IngestDocuments(ctx, ingestRequest)
	if err != nil {
		t.Fatalf("Failed to ingest documents: %v", err)
	}

	t.Logf("Ingested documents: %d succeeded, %d failed, %d total",
		ingestResponse.Succeeded, ingestResponse.Failed, ingestResponse.Total)

	// Wait for indexing to complete
	time.Sleep(3 * time.Second)

	// 6. Get task status
	taskStatus, err := client.GetTaskStatus(ctx, ingestResponse.Task.TaskID)
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}

	t.Logf("Task status: %s", taskStatus.State)

	// 7. Search for documents
	searchRequest := &SearchRequest{
		IndexID: createdIndex.ID,
		Query:   "test",
		Options: &SearchOptions{
			Limit: 10,
		},
	}

	searchResponse, err := client.Search(ctx, searchRequest)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	t.Logf("Search found %d documents", searchResponse.Count)

	// 8. Delete documents
	deleteRequest := &DeleteDocumentsRequest{
		IndexID:  createdIndex.ID,
		Subjects: []string{documents[0].Subject},
	}

	deleteResponse, err := client.DeleteDocuments(ctx, deleteRequest)
	if err != nil {
		t.Fatalf("Failed to delete documents: %v", err)
	}

	t.Logf("Deleted documents: %d succeeded, %d failed, %d total",
		deleteResponse.Succeeded, deleteResponse.Failed, deleteResponse.Total)
}

func TestIntegration_ExistingIndex(t *testing.T) {
	clientID, clientSecret, indexID := getTestCredentials(t)

	// Skip if no existing index ID is provided
	if indexID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SEARCH_INDEX_ID for existing index operations")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Search client
	client := NewClient(accessToken)
	ctx := context.Background()

	// Verify we can get the index
	index, err := client.GetIndex(ctx, indexID)
	if err != nil {
		t.Fatalf("Failed to get index: %v", err)
	}

	t.Logf("Found index: %s (%s)", index.DisplayName, index.ID)

	// Search the existing index
	searchRequest := &SearchRequest{
		IndexID: indexID,
		Query:   "*",
		Options: &SearchOptions{
			Limit: 5,
		},
	}

	searchResponse, err := client.Search(ctx, searchRequest)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	t.Logf("Search found %d documents", searchResponse.Count)

	if len(searchResponse.Results) > 0 {
		// Check that the first result has expected fields
		firstResult := searchResponse.Results[0]
		if firstResult.Subject == "" {
			t.Error("First result is missing subject")
		}
	}
}
