---
title: "Search Service: Client"
---
# Search Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Search client provides access to the Globus Search API, which allows you to index, search, and discover data across Globus endpoints.

## Client Structure

```go
type Client struct {
    client      *core.Client
    transport   *core.HTTPTransport
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*core.Client` | Core client for making HTTP requests |
| `transport` | `*core.HTTPTransport` | HTTP transport for request/response handling |

## Creating a Search Client

```go
// Create a search client with options
client, err := search.NewClient(
    search.WithAccessToken("access-token"),
    search.WithHTTPDebugging(),
)
if err != nil {
    // Handle error
}
```

### Options

| Option | Description |
|--------|-------------|
| `WithAccessToken(token string)` | Sets the access token for authorization |
| `WithAuthorizer(auth core.Authorizer)` | Sets a custom authorizer (alternative to access token) |
| `WithBaseURL(url string)` | Sets a custom base URL (default: "https://search.api.globus.org/v1/") |
| `WithHTTPDebugging()` | Enables HTTP debugging |
| `WithHTTPTracing()` | Enables HTTP tracing |

## Index Management

### Listing Indexes

```go
// List all accessible indexes
indexes, err := client.ListIndexes(ctx, nil)
if err != nil {
    // Handle error
}

// List indexes with options
indexes, err := client.ListIndexes(ctx, &search.ListIndexesOptions{
    Limit:  100,
    Offset: 0,
    Scope:  "all",
})
if err != nil {
    // Handle error
}

// Iterate through indexes
for _, index := range indexes.Indexes {
    fmt.Printf("Index: %s (%s)\n", index.DisplayName, index.ID)
}
```

### Getting an Index

```go
// Get a specific index by ID
index, err := client.GetIndex(ctx, "index-id")
if err != nil {
    // Handle error
}

fmt.Printf("Index Name: %s\n", index.DisplayName)
fmt.Printf("Description: %s\n", index.Description)
fmt.Printf("Owner: %s\n", index.OwnerString)
```

### Creating an Index

```go
// Create a new index
request := &search.IndexCreateRequest{
    DisplayName:  "My Research Data",
    Description:  "Index for my research project data",
    VisibleTo:    []string{"public"},
    AdminUsers:   []string{"user@example.com"},
    AdminGroups:  []string{"group-id"},
    IndexAttributes: map[string]interface{}{
        "title": "My Research Project",
        "tags":  []string{"research", "data"},
    },
}

index, err := client.CreateIndex(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Created index: %s\n", index.ID)
```

### Updating an Index

```go
// Update an existing index
request := &search.IndexUpdateRequest{
    DisplayName: "Updated Research Data",
    Description: "Updated description for my research data",
    VisibleTo:   []string{"public"},
    IndexAttributes: map[string]interface{}{
        "updated": true,
    },
}

index, err := client.UpdateIndex(ctx, "index-id", request)
if err != nil {
    // Handle error
}

fmt.Printf("Updated index: %s\n", index.DisplayName)
```

### Deleting an Index

```go
// Delete an index
err := client.DeleteIndex(ctx, "index-id")
if err != nil {
    // Handle error
}

fmt.Println("Index deleted successfully")
```

## Document Management

### Ingesting Documents

```go
// Ingest documents into an index
documents := []search.SearchDocument{
    {
        "subject":     "document-1",
        "title":       "First Document",
        "description": "This is the first document",
        "keywords":    []string{"first", "document"},
        "created":     time.Now().Format(time.RFC3339),
    },
    {
        "subject":     "document-2",
        "title":       "Second Document",
        "description": "This is the second document",
        "keywords":    []string{"second", "document"},
        "created":     time.Now().Format(time.RFC3339),
    },
}

request := &search.IngestRequest{
    Documents: documents,
}

response, err := client.IngestDocuments(ctx, "index-id", request)
if err != nil {
    // Handle error
}

fmt.Printf("Ingest task ID: %s\n", response.Task.TaskID)
```

### Checking Task Status

```go
// Check the status of an ingest task
taskStatus, err := client.GetTaskStatus(ctx, "index-id", "task-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task Status: %s\n", taskStatus.State)
if taskStatus.State == "SUCCESS" {
    fmt.Println("Task completed successfully")
} else if taskStatus.State == "FAILED" {
    fmt.Println("Task failed:", taskStatus.Message)
}
```

### Deleting Documents

```go
// Delete documents from an index
request := &search.DeleteDocumentsRequest{
    Subjects: []string{"document-1", "document-2"},
}

response, err := client.DeleteDocuments(ctx, "index-id", request)
if err != nil {
    // Handle error
}

fmt.Printf("Delete task ID: %s\n", response.Task.TaskID)
```

## Basic Search

### Simple Text Search

```go
// Perform a simple text search
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "research data",
    Filters: []string{},
    Options: &search.SearchOptions{
        Limit:  10,
        Offset: 0,
        Sort:   []string{"created.desc"},
    },
}

response, err := client.Search(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d results\n", response.Count)
for _, result := range response.Results {
    fmt.Printf("- %s (Score: %.2f)\n", result.Subject, result.Score)
}
```

### Paginated Search

```go
// Create a search iterator
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "research data",
    Options: &search.SearchOptions{
        Limit: 100,
    },
}

iterator, err := client.SearchIterator(ctx, request)
if err != nil {
    // Handle error
}

// Iterate through all pages of results
for iterator.HasNext() {
    response, err := iterator.Next()
    if err != nil {
        // Handle error
        break
    }
    
    fmt.Printf("Page with %d results\n", len(response.Results))
    for _, result := range response.Results {
        fmt.Printf("- %s\n", result.Subject)
    }
}
```

### Search All Results

```go
// Retrieve all search results
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "research data",
}

allResults, err := client.SearchAll(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d total results\n", len(allResults))
```

## Error Handling

The search package provides specific error types and helper functions for common error conditions:

```go
// Try an operation
_, err := client.GetIndex(ctx, "non-existent-index")
if err != nil {
    switch {
    case search.IsIndexNotFoundError(err):
        fmt.Println("Index not found")
    case search.IsPermissionDeniedError(err):
        fmt.Println("Permission denied")
    case search.IsRateLimitError(err):
        fmt.Println("Rate limit exceeded")
    case search.IsInvalidQueryError(err):
        fmt.Println("Invalid query")
    default:
        fmt.Println("Other error:", err)
    }
}
```

### Error Details

```go
// Get more details from a search error
if searchErr, ok := search.AsSearchError(err); ok {
    fmt.Println("Error code:", searchErr.Code)
    fmt.Println("Error message:", searchErr.Message)
    fmt.Println("Request ID:", searchErr.RequestID)
}
```

## Common Patterns

### Create Index and Ingest Documents

```go
// Create a new index
indexReq := &search.IndexCreateRequest{
    DisplayName: "Project Data",
    Description: "Index for project data",
    VisibleTo:   []string{"public"},
}

index, err := client.CreateIndex(ctx, indexReq)
if err != nil {
    // Handle error
}

// Prepare documents
documents := []search.SearchDocument{
    {
        "subject":     "doc1",
        "title":       "Document 1",
        "description": "First document in the index",
        "created":     time.Now().Format(time.RFC3339),
    },
}

// Ingest documents
ingestReq := &search.IngestRequest{
    Documents: documents,
}

ingestResp, err := client.IngestDocuments(ctx, index.ID, ingestReq)
if err != nil {
    // Handle error
}

// Wait for ingest to complete
for {
    taskStatus, err := client.GetTaskStatus(ctx, index.ID, ingestResp.Task.TaskID)
    if err != nil {
        // Handle error
        break
    }
    
    if taskStatus.State == "SUCCESS" {
        fmt.Println("Documents indexed successfully")
        break
    } else if taskStatus.State == "FAILED" {
        fmt.Println("Indexing failed:", taskStatus.Message)
        break
    }
    
    fmt.Println("Indexing in progress...")
    time.Sleep(1 * time.Second)
}
```

### Search with Filters

```go
// Search with filters
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "research",
    Filters: []string{
        "created:[2023-01-01 TO 2023-12-31]", // Date range filter
        "keywords:data",                       // Keyword filter
    },
    Options: &search.SearchOptions{
        Limit: 10,
        Sort:  []string{"created.desc"},      // Sort by creation date
    },
}

response, err := client.Search(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d filtered results\n", response.Count)
```

### Get All Indexes

```go
// Get all indexes with automatic pagination
var allIndexes []search.Index
offset := 0
limit := 100

for {
    indexes, err := client.ListIndexes(ctx, &search.ListIndexesOptions{
        Offset: offset,
        Limit:  limit,
    })
    if err != nil {
        // Handle error
        break
    }
    
    allIndexes = append(allIndexes, indexes.Indexes...)
    
    // Check if we've retrieved all indexes
    if len(indexes.Indexes) < limit {
        break
    }
    
    // Increment offset for next page
    offset += limit
}

fmt.Printf("Total indexes: %d\n", len(allIndexes))
```

## Best Practices

1. Use appropriate error handling for search-specific errors
2. For large result sets, use pagination or the SearchAll method
3. When ingesting documents, check task status to ensure completion
4. Use filters to narrow search results and improve performance
5. Include the "subject" field in all documents for unique identification
6. For large document sets, use batch operations (see batch.md)
7. Include meaningful metadata in documents to improve search relevance
8. Use structured queries for complex search requirements
9. Consider index visibility settings for proper access control
10. Use context timeouts for search operations to prevent long-running queries