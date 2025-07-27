// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package search provides a client for interacting with the Globus Search service.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees. Components listed below are
considered part of the public API and will not change incompatibly within a major version:

  - Client interface and implementation
  - Index operations (ListIndexes, GetIndex, CreateIndex, UpdateIndex, DeleteIndex)
  - Basic document operations (IngestDocuments, DeleteDocuments)
  - Simple search operations (Search)
  - Task monitoring (GetTaskStatus)
  - Core model types (Index, SearchDocument, SearchResult)
  - Client configuration options
  - Advanced query framework (Query interfaces and implementations)
  - Batch operations for large datasets
  - Query parsing functionality
  - Error handling patterns
  - Search iterator and pagination

# Compatibility Guarantees

For stable packages:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v3.x to v4.x)

# Synchronized Versioning

Starting with v3.60.0-1, this package follows synchronized versioning with the Globus Python SDK.
This ensures API compatibility and feature parity across language implementations.

# Basic Usage

Create a new search client:

	searchClient := search.NewClient(
		search.WithAuthorizer(authorizer),
	)

Index Management:

	// List indexes
	indexes, err := searchClient.ListIndexes(ctx)
	if err != nil {
		// Handle error
	}

	for _, index := range indexes.Indexes {
		fmt.Printf("Index ID: %s, Display Name: %s\n", index.ID, index.DisplayName)
	}

	// Get a specific index
	index, err := searchClient.GetIndex(ctx, "index_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Index: %s (%s)\n", index.DisplayName, index.Description)

	// Create an index
	newIndex := &search.IndexCreateRequest{
		DisplayName: "My New Index",
		Description: "An index for my data",
	}

	created, err := searchClient.CreateIndex(ctx, newIndex)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Created index with ID: %s\n", created.ID)

	// Update an index
	update := &search.IndexUpdateRequest{
		DisplayName: "Updated Index Name",
		Description: "Updated description",
	}

	updated, err := searchClient.UpdateIndex(ctx, "index_id", update)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Updated index: %s\n", updated.DisplayName)

	// Delete an index
	err = searchClient.DeleteIndex(ctx, "index_id")
	if err != nil {
		// Handle error
	}

Document Operations:

	// Ingest documents
	documents := []search.SearchDocument{
		{
			"subject":   "document_1",
			"title":     "First Document",
			"content":   "This is the content of the first document",
			"timestamp": time.Now().Format(time.RFC3339),
		},
		{
			"subject":   "document_2",
			"title":     "Second Document",
			"content":   "This is the content of the second document",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	ingestResponse, err := searchClient.IngestDocuments(ctx, "index_id", documents)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Ingest task ID: %s\n", ingestResponse.TaskID)

	// Wait for task completion
	taskStatus, err := searchClient.WaitForTask(ctx, "index_id", ingestResponse.TaskID)
	if err != nil {
		// Handle error
	}

	if taskStatus.State == "success" {
		fmt.Println("Documents ingested successfully!")
	} else {
		fmt.Printf("Ingest failed: %s\n", taskStatus.Message)
	}

	// Delete documents
	subjects := []string{"document_1", "document_2"}
	deleteResponse, err := searchClient.DeleteDocuments(ctx, "index_id", subjects)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Delete task ID: %s\n", deleteResponse.TaskID)

Search Operations:

	// Simple search
	searchRequest := &search.SearchRequest{
		Q:          "example query",
		Limit:      10,
		Offset:     0,
		SortFields: []string{"timestamp:desc"},
	}

	searchResponse, err := searchClient.Search(ctx, "index_id", searchRequest)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Found %d results\n", searchResponse.Count)
	for _, result := range searchResponse.Results {
		fmt.Printf("Subject: %s, Title: %s\n", result.Subject, result.Content["title"])
	}

	// Search with iterator (for pagination)
	iterator, err := searchClient.SearchIterator(ctx, "index_id", searchRequest)
	if err != nil {
		// Handle error
	}

	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			// Handle error
		}
		fmt.Printf("Subject: %s, Title: %s\n", result.Subject, result.Content["title"])
	}

	// Get all results at once
	allResults, err := searchClient.SearchAll(ctx, "index_id", searchRequest)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Retrieved %d total results\n", len(allResults))

Advanced Query:

	// Structured search using query builders
	query := search.NewBoolQuery().
		Must(search.NewTermQuery("type", "document")).
		Should(
			search.NewMatchQuery("content", "example").Boost(2.0),
			search.NewRangeQuery("timestamp").WithGT("2023-01-01"),
		)

	structuredRequest := &search.SearchRequest{
		Query: query,
		Limit: 10,
	}

	results, err := searchClient.StructuredSearch(ctx, "index_id", structuredRequest)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Found %d results with structured query\n", results.Count)

Batch Operations:

	// Batch ingest for large datasets
	batchSize := 1000
	allDocuments := make([]search.SearchDocument, 0, 10000)
	// Populate allDocuments...

	batchResponse, err := searchClient.BatchIngestDocuments(ctx, "index_id", allDocuments, batchSize)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Successfully ingested %d documents in %d batches\n",
		batchResponse.TotalDocuments,
		len(batchResponse.TaskIDs))
*/
package search
