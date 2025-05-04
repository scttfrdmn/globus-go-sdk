---
title: "Search Service Quick Start"
weight: 30
---

# Search Service Quick Start

This guide will help you get started with the Globus Search service using the Go SDK. The Search service enables indexing, searching, and discovering data across the Globus ecosystem.

## Setup

First, import the required packages and create a context:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // Continue with the examples below...
}
```

## Creating a Search Client

There are two main ways to create a Search client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Search client
searchClient, err := config.NewSearchClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create search client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Search client with options
searchClient, err := search.NewClient(
    search.WithAccessToken(os.Getenv("GLOBUS_ACCESS_TOKEN")),
    search.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create search client: %v", err)
}
```

## Working with Indexes

Indexes in Globus Search are where your data is stored and made searchable.

### Listing Your Indexes

```go
// List all indexes you have access to
indexes, err := searchClient.ListIndexes(ctx, nil)
if err != nil {
    log.Fatalf("Failed to list indexes: %v", err)
}

fmt.Printf("Found %d indexes:\n", len(indexes.Indexes))
for i, index := range indexes.Indexes {
    fmt.Printf("%d. %s (%s)\n", i+1, index.DisplayName, index.ID)
}
```

### Listing Indexes with Options

```go
// List indexes with pagination and scope
indexes, err := searchClient.ListIndexes(ctx, &search.ListIndexesOptions{
    Limit:  25,           // Limit results per page
    Offset: 0,            // Starting offset
    Scope:  "all",        // Options: "all", "private", "public"
})
if err != nil {
    log.Fatalf("Failed to list indexes: %v", err)
}
```

### Getting Index Details

```go
// Get details about a specific index
indexID := "your-index-id"
index, err := searchClient.GetIndex(ctx, indexID)
if err != nil {
    log.Fatalf("Failed to get index: %v", err)
}

fmt.Printf("Index: %s\n", index.DisplayName)
fmt.Printf("Description: %s\n", index.Description)
fmt.Printf("Owner: %s\n", index.OwnerString)
fmt.Printf("Created: %s\n", index.CreatedAt)
```

### Creating a New Index

```go
// Create a new search index
request := &search.IndexCreateRequest{
    DisplayName: "Research Dataset Index",
    Description: "Index for my research project datasets",
    VisibleTo:   []string{"public"},           // Make it publicly visible
    // Optional: specify admin users and groups
    AdminUsers:  []string{"your-user-id"},     // Users who can admin the index
    AdminGroups: []string{"your-group-id"},    // Groups who can admin the index
    // Optional: add custom attributes
    IndexAttributes: map[string]interface{}{
        "project": "Climate Research",
        "tags":    []string{"climate", "datasets", "research"},
    },
}

newIndex, err := searchClient.CreateIndex(ctx, request)
if err != nil {
    log.Fatalf("Failed to create index: %v", err)
}

fmt.Printf("Created new index with ID: %s\n", newIndex.ID)
```

### Updating an Index

```go
// Update an existing index
request := &search.IndexUpdateRequest{
    DisplayName: "Updated Research Dataset Index",
    Description: "Updated description for my research project datasets",
    // Add or update index attributes
    IndexAttributes: map[string]interface{}{
        "updated_at": time.Now().Format(time.RFC3339),
        "version":    "2.0",
    },
}

updatedIndex, err := searchClient.UpdateIndex(ctx, indexID, request)
if err != nil {
    log.Fatalf("Failed to update index: %v", err)
}

fmt.Printf("Updated index: %s\n", updatedIndex.DisplayName)
```

## Working with Documents

Documents are the searchable items within an index.

### Ingesting Documents

```go
// Prepare documents to ingest
documents := []search.SearchDocument{
    {
        "subject":     "document-1",             // Required: unique identifier
        "title":       "Climate Dataset 2023",
        "description": "Global temperature measurements for 2023",
        "keywords":    []string{"climate", "temperature", "dataset"},
        "published":   time.Now().Format(time.RFC3339),
        "format":      "CSV",
        "size":        1548576,
        "author":      "Research Team Alpha",
        "location":    map[string]interface{}{
            "lat": 37.7749,
            "lon": -122.4194,
        },
    },
    {
        "subject":     "document-2",
        "title":       "Rainfall Analysis",
        "description": "Precipitation patterns in North America",
        "keywords":    []string{"rainfall", "precipitation", "climate"},
        "published":   time.Now().Format(time.RFC3339),
    },
}

// Create ingest request
ingestRequest := &search.IngestRequest{
    Documents: documents,
}

// Ingest documents into the index
response, err := searchClient.IngestDocuments(ctx, indexID, ingestRequest)
if err != nil {
    log.Fatalf("Failed to ingest documents: %v", err)
}

fmt.Printf("Ingest started with task ID: %s\n", response.Task.TaskID)
```

### Checking Ingest Task Status

Documents are indexed asynchronously, so you should check the task status:

```go
// Check the status of an ingest task
taskID := response.Task.TaskID

// Poll for task completion
for {
    taskStatus, err := searchClient.GetTaskStatus(ctx, indexID, taskID)
    if err != nil {
        log.Fatalf("Failed to get task status: %v", err)
    }
    
    fmt.Printf("Task status: %s\n", taskStatus.State)
    
    if taskStatus.State == "SUCCESS" {
        fmt.Println("Documents successfully indexed!")
        break
    } else if taskStatus.State == "FAILED" {
        fmt.Printf("Indexing failed: %s\n", taskStatus.Message)
        break
    }
    
    // Wait before checking again
    time.Sleep(1 * time.Second)
}
```

### Deleting Documents

```go
// Delete documents from an index
deleteRequest := &search.DeleteDocumentsRequest{
    Subjects: []string{"document-1", "document-2"},
}

response, err := searchClient.DeleteDocuments(ctx, indexID, deleteRequest)
if err != nil {
    log.Fatalf("Failed to delete documents: %v", err)
}

fmt.Printf("Delete task started with ID: %s\n", response.Task.TaskID)

// You can check the delete task status using GetTaskStatus, similar to ingest tasks
```

## Basic Searching

### Simple Text Search

```go
// Perform a simple text search
request := &search.SearchRequest{
    IndexID: indexID,
    Query:   "climate temperature",   // Search for documents about climate and temperature
    Options: &search.SearchOptions{
        Limit:  10,                   // Limit to 10 results
        Sort:   []string{"published.desc"},  // Sort by published date, newest first
    },
}

// Execute the search
results, err := searchClient.Search(ctx, request)
if err != nil {
    log.Fatalf("Failed to search: %v", err)
}

fmt.Printf("Found %d results\n", results.Count)
for i, result := range results.Results {
    fmt.Printf("%d. %s (Score: %.2f)\n", i+1, result.Content["title"], result.Score)
    
    // Access other fields from the content
    if desc, ok := result.Content["description"].(string); ok {
        fmt.Printf("   Description: %s\n", desc)
    }
}
```

### Searching with Filters

```go
// Search with filters
request := &search.SearchRequest{
    IndexID: indexID,
    Query:   "climate",
    Filters: []string{
        "keywords:temperature",            // Filter by keyword
        "published:[2023-01-01 TO *]",    // Filter by date range (from Jan 1, 2023 to present)
    },
    Options: &search.SearchOptions{
        Limit: 10,
    },
}

results, err := searchClient.Search(ctx, request)
if err != nil {
    log.Fatalf("Failed to search with filters: %v", err)
}

fmt.Printf("Found %d filtered results\n", results.Count)
```

### Paginating Search Results

```go
// Search with pagination
limit := 5
offset := 0

// First page
request := &search.SearchRequest{
    IndexID: indexID,
    Query:   "climate",
    Options: &search.SearchOptions{
        Limit:  limit,
        Offset: offset,
    },
}

firstPage, err := searchClient.Search(ctx, request)
if err != nil {
    log.Fatalf("Failed to get first page: %v", err)
}

fmt.Printf("Page 1: %d results\n", len(firstPage.Results))

// Next page
if firstPage.Count > limit {
    offset += limit
    request.Options.Offset = offset
    
    secondPage, err := searchClient.Search(ctx, request)
    if err != nil {
        log.Fatalf("Failed to get second page: %v", err)
    }
    
    fmt.Printf("Page 2: %d results\n", len(secondPage.Results))
}
```

### Using a Search Iterator

For large result sets, an iterator is more convenient:

```go
// Create a search iterator
request := &search.SearchRequest{
    IndexID: indexID,
    Query:   "climate",
    Options: &search.SearchOptions{
        Limit: 100,  // Results per page
    },
}

iterator, err := searchClient.SearchIterator(ctx, request)
if err != nil {
    log.Fatalf("Failed to create iterator: %v", err)
}

// Iterate through all results
totalResults := 0
for iterator.HasNext() {
    page, err := iterator.Next()
    if err != nil {
        log.Fatalf("Error getting next page: %v", err)
    }
    
    totalResults += len(page.Results)
    fmt.Printf("Processing page with %d results...\n", len(page.Results))
    
    // Process results...
}

fmt.Printf("Processed %d total results\n", totalResults)
```

### Getting All Search Results

If you need all results at once:

```go
// Get all search results
request := &search.SearchRequest{
    IndexID: indexID,
    Query:   "climate",
}

allResults, err := searchClient.SearchAll(ctx, request)
if err != nil {
    log.Fatalf("Failed to get all results: %v", err)
}

fmt.Printf("Retrieved all %d results\n", len(allResults))
```

## Advanced Searching

### Using Structured Queries

For more complex search requirements, you can use structured queries:

```go
// Create a structured query that looks for climate data from 2023
boolQuery := search.NewBoolQuery().
    Must(
        // Must contain "climate" in title or description
        search.NewBoolQuery().
            Should(
                search.NewMatchQuery("title", "climate"),
                search.NewMatchQuery("description", "climate"),
            ),
    ).
    Filter(
        // Only from 2023
        search.NewRangeQuery("published").
            WithGte("2023-01-01").
            WithLt("2024-01-01"),
        // Only datasets (not papers or other types)
        search.NewTermQuery("format", "CSV"),
    )

// Create the structured search request
request := &search.StructuredSearchRequest{
    IndexID: indexID,
    Query:   boolQuery,
    Options: &search.SearchOptions{
        Limit: 10,
        Sort:  []string{"_score.desc", "published.desc"},  // Sort by relevance, then date
    },
}

// Execute the structured search
results, err := searchClient.StructuredSearch(ctx, request)
if err != nil {
    log.Fatalf("Failed to execute structured search: %v", err)
}

fmt.Printf("Found %d structured search results\n", results.Count)
```

### Field-specific Matching

```go
// Match documents with specific field values
matchQuery := search.NewMatchQuery("author", "Research Team Alpha")

request := &search.StructuredSearchRequest{
    IndexID: indexID,
    Query:   matchQuery,
}

results, err := searchClient.StructuredSearch(ctx, request)
if err != nil {
    log.Fatalf("Failed to match field: %v", err)
}
```

### Range Queries

```go
// Find documents within a size range
rangeQuery := search.NewRangeQuery("size").
    WithGte(1000000).    // At least 1MB
    WithLt(10000000)     // Less than 10MB

request := &search.StructuredSearchRequest{
    IndexID: indexID,
    Query:   rangeQuery,
}

results, err := searchClient.StructuredSearch(ctx, request)
if err != nil {
    log.Fatalf("Failed to execute range query: %v", err)
}
```

### Geo-Distance Search

```go
// Find documents near San Francisco (within 100km)
geoQuery := search.NewGeoDistanceQuery("location", 37.7749, -122.4194, "100km")

request := &search.StructuredSearchRequest{
    IndexID: indexID,
    Query:   geoQuery,
}

results, err := searchClient.StructuredSearch(ctx, request)
if err != nil {
    log.Fatalf("Failed to execute geo query: %v", err)
}
```

## Error Handling

The Search service provides specific error types for better error handling:

```go
// Try to get a non-existent index
_, err := searchClient.GetIndex(ctx, "non-existent-index")
if err != nil {
    switch {
    case search.IsIndexNotFoundError(err):
        fmt.Println("Index not found - check the index ID")
    case search.IsPermissionDeniedError(err):
        fmt.Println("Permission denied - check your access token and permissions")
    case search.IsRateLimitError(err):
        fmt.Println("Rate limit exceeded - slow down your requests")
    case search.IsInvalidQueryError(err):
        fmt.Println("Invalid query syntax - check your query format")
    default:
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Complete Example

Here's a complete example that creates an index, ingests documents, and searches for them:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Get access token from environment
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if accessToken == "" {
        log.Fatalf("GLOBUS_ACCESS_TOKEN environment variable is required")
    }
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create search client
    searchClient, err := config.NewSearchClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create search client: %v", err)
    }
    
    // Step 1: Create a new index
    indexReq := &search.IndexCreateRequest{
        DisplayName: "Quick Start Example Index",
        Description: "Index created for the quick start guide",
        VisibleTo:   []string{"public"},
    }
    
    index, err := searchClient.CreateIndex(ctx, indexReq)
    if err != nil {
        log.Fatalf("Failed to create index: %v", err)
    }
    fmt.Printf("Created index: %s (ID: %s)\n", index.DisplayName, index.ID)
    
    // Step 2: Ingest some documents
    documents := []search.SearchDocument{
        {
            "subject":     "doc1",
            "title":       "Introduction to Globus Search",
            "description": "A beginner's guide to using Globus Search",
            "tags":        []string{"tutorial", "globus", "search"},
            "created":     time.Now().Format(time.RFC3339),
        },
        {
            "subject":     "doc2",
            "title":       "Advanced Search Techniques",
            "description": "Learn advanced techniques for Globus Search",
            "tags":        []string{"advanced", "globus", "search", "techniques"},
            "created":     time.Now().Format(time.RFC3339),
        },
    }
    
    ingestReq := &search.IngestRequest{
        Documents: documents,
    }
    
    ingestResp, err := searchClient.IngestDocuments(ctx, index.ID, ingestReq)
    if err != nil {
        log.Fatalf("Failed to ingest documents: %v", err)
    }
    fmt.Printf("Started ingesting documents, task ID: %s\n", ingestResp.Task.TaskID)
    
    // Step 3: Wait for ingestion to complete
    fmt.Println("Waiting for documents to be indexed...")
    var taskComplete bool
    for i := 0; i < 30 && !taskComplete; i++ {
        taskStatus, err := searchClient.GetTaskStatus(ctx, index.ID, ingestResp.Task.TaskID)
        if err != nil {
            log.Fatalf("Failed to get task status: %v", err)
        }
        
        if taskStatus.State == "SUCCESS" {
            fmt.Println("Documents successfully indexed!")
            taskComplete = true
        } else if taskStatus.State == "FAILED" {
            log.Fatalf("Indexing failed: %s", taskStatus.Message)
        } else {
            fmt.Printf("Indexing in progress, status: %s\n", taskStatus.State)
            time.Sleep(1 * time.Second)
        }
    }
    
    if !taskComplete {
        log.Fatalf("Indexing timed out after 30 seconds")
    }
    
    // Step 4: Search for documents
    fmt.Println("\nSearching for documents about 'globus'...")
    searchReq := &search.SearchRequest{
        IndexID: index.ID,
        Query:   "globus",
        Options: &search.SearchOptions{
            Sort: []string{"created.desc"},
        },
    }
    
    results, err := searchClient.Search(ctx, searchReq)
    if err != nil {
        log.Fatalf("Failed to search: %v", err)
    }
    
    fmt.Printf("Found %d results:\n", results.Count)
    for i, result := range results.Results {
        title := result.Content["title"].(string)
        desc := result.Content["description"].(string)
        fmt.Printf("%d. %s\n   %s\n", i+1, title, desc)
    }
    
    // Step 5: Try a more specific search with a filter
    fmt.Println("\nSearching for documents about 'advanced' techniques...")
    advancedReq := &search.SearchRequest{
        IndexID: index.ID,
        Query:   "advanced",
        Filters: []string{"tags:techniques"},
    }
    
    advResults, err := searchClient.Search(ctx, advancedReq)
    if err != nil {
        log.Fatalf("Failed to search with filter: %v", err)
    }
    
    fmt.Printf("Found %d filtered results:\n", advResults.Count)
    for i, result := range advResults.Results {
        title := result.Content["title"].(string)
        fmt.Printf("%d. %s (Score: %.2f)\n", i+1, title, result.Score)
    }
    
    // Optional cleanup: delete the index when done with this example
    fmt.Println("\nCleaning up - deleting the example index...")
    err = searchClient.DeleteIndex(ctx, index.ID)
    if err != nil {
        log.Fatalf("Failed to delete index: %v", err)
    }
    fmt.Println("Index deleted successfully")
}
```

## Next Steps

Now that you understand the basics of the Search service, you can:

1. **Create advanced queries**: Explore the [Advanced Queries](/docs/reference/search/advanced) documentation for more complex search capabilities
2. **Implement batch operations**: Use the [Batch Operations](/docs/reference/search/batch) for efficient bulk operations
3. **Explore faceted search**: Implement facets to help users refine search results
4. **Build search interfaces**: Create user interfaces that leverage the powerful search capabilities

For more details, check out the [Search Service API Reference](/docs/reference/search/) documentation.