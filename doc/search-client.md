<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Search Client

The Globus Search client in the Go SDK provides a comprehensive interface to the Globus Search service, allowing you to index, search, and manage search data.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Index Management](#index-management)
- [Document Management](#document-management)
- [Basic Search](#basic-search)
- [Advanced Queries](#advanced-queries)
- [Pagination](#pagination)
- [Batch Operations](#batch-operations)
- [Error Handling](#error-handling)
- [Task Management](#task-management)
- [Complete Example](#complete-example)

## Basic Usage

### Creating a Search Client

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Create a search client with an access token
client := search.NewClient("your-access-token")

// With custom options
client := search.NewClient("your-access-token",
    core.WithBaseURL("https://custom-search-url.globus.org/v1/"),
    core.WithTimeout(30 * time.Second),
)
```

## Index Management

### Listing Indexes

```go
// List all indexes with default options
indexes, err := client.ListIndexes(context.Background(), nil)

// List with custom options
indexes, err := client.ListIndexes(context.Background(), &search.ListIndexesOptions{
    Limit:     10,
    IsPublic:  true,
    CreatedBy: "user@example.com",
})

// Process results
for _, index := range indexes.Indexes {
    fmt.Printf("Index: %s (%s)\n", index.DisplayName, index.ID)
}
```

### Getting an Index

```go
// Get a specific index by ID
index, err := client.GetIndex(context.Background(), "your-index-id")
if err != nil {
    // Handle error
}

fmt.Printf("Index: %s\nDescription: %s\n", index.DisplayName, index.Description)
```

### Creating an Index

```go
// Create a new index
createReq := &search.IndexCreateRequest{
    DisplayName: "My Research Data",
    Description: "Index for my research project data",
    IsMonitored: true,
    DefinitionDocument: map[string]interface{}{
        "mappings": map[string]interface{}{
            "properties": map[string]interface{}{
                "title": map[string]interface{}{
                    "type": "text",
                },
                "keywords": map[string]interface{}{
                    "type": "keyword",
                },
                "date": map[string]interface{}{
                    "type": "date",
                },
            },
        },
    },
}

index, err := client.CreateIndex(context.Background(), createReq)
if err != nil {
    // Handle error
}

// Use the new index ID
fmt.Printf("Created index with ID: %s\n", index.ID)
```

### Updating an Index

```go
// Update an existing index
updateReq := &search.IndexUpdateRequest{
    DisplayName: "Updated Research Data",
    Description: "Updated description for my research project",
    IsActive:    true,
}

index, err := client.UpdateIndex(context.Background(), "your-index-id", updateReq)
if err != nil {
    // Handle error
}
```

### Deleting an Index

```go
// Delete an index
err := client.DeleteIndex(context.Background(), "your-index-id")
if err != nil {
    // Handle error
}
```

## Document Management

### Ingesting Documents

```go
// Create documents
docs := []search.SearchDocument{
    {
        Subject: "doc1",
        Content: map[string]interface{}{
            "title":    "First Document",
            "keywords": []string{"research", "data"},
            "date":     "2023-01-01",
            "content":  "This is the content of the first document.",
        },
        // Optional: control visibility
        VisibleTo: []string{"public"},
    },
    {
        Subject: "doc2",
        Content: map[string]interface{}{
            "title":    "Second Document",
            "keywords": []string{"experiment", "results"},
            "date":     "2023-02-15",
            "content":  "This is the content of the second document.",
        },
    },
}

// Ingest documents
ingestReq := &search.IngestRequest{
    IndexID:   "your-index-id",
    Documents: docs,
}

resp, err := client.IngestDocuments(context.Background(), ingestReq)
if err != nil {
    // Handle error
}

fmt.Printf("Successfully ingested %d/%d documents\n", resp.Succeeded, resp.Total)
fmt.Printf("Task ID: %s\n", resp.Task.TaskID)
```

### Deleting Documents

```go
// Delete documents by their subjects
deleteReq := &search.DeleteDocumentsRequest{
    IndexID:  "your-index-id",
    Subjects: []string{"doc1", "doc2"},
}

resp, err := client.DeleteDocuments(context.Background(), deleteReq)
if err != nil {
    // Handle error
}

fmt.Printf("Successfully deleted %d/%d documents\n", resp.Succeeded, resp.Total)
```

## Basic Search

```go
// Perform a simple search
searchReq := &search.SearchRequest{
    IndexID: "your-index-id",
    Query:   "research data",
    Options: &search.SearchOptions{
        Limit:             20,
        Sort:              []string{"date:desc"},
        IncludeAllContent: true,
    },
}

results, err := client.Search(context.Background(), searchReq)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d results out of %d total\n", results.Count, results.Total)

// Process results
for _, result := range results.Results {
    fmt.Printf("Subject: %s (Score: %.2f)\n", result.Subject, result.Score)
    fmt.Printf("Title: %s\n", result.Content["title"])
    
    // Access other fields
    if keywords, ok := result.Content["keywords"].([]interface{}); ok {
        fmt.Printf("Keywords: %v\n", keywords)
    }
}
```

## Advanced Queries

The SDK provides a powerful query builder for constructing complex search queries:

### Match Query

```go
// Search for documents with a specific title
query := search.NewMatchQuery("title", "research data")
```

### Term Query

```go
// Exact match on a keyword field
query := search.NewTermQuery("keywords", "experiment")
```

### Range Query

```go
// Date range query
query := search.NewRangeQuery("date").
    WithGTE("2023-01-01").
    WithLT("2024-01-01").
    WithFormat("yyyy-MM-dd")
```

### Boolean Query

```go
// Combined query with boolean logic
query := search.NewBoolQuery().
    AddMust(search.NewMatchQuery("title", "research")).
    AddMustNot(search.NewTermQuery("status", "archived")).
    AddShould(search.NewTermQuery("keywords", "important")).
    AddShould(search.NewTermQuery("keywords", "urgent")).
    SetMinimumShouldMatch(1)
```

### Other Query Types

```go
// Check if a field exists
query := search.NewExistsQuery("attachment")

// Prefix search
query := search.NewPrefixQuery("title", "res")

// Wildcard search
query := search.NewWildcardQuery("title", "res*ch")

// Geo distance search
query := search.NewGeoDistanceQuery("location", "10km", 37.7749, -122.4194)
```

### Using Advanced Queries

```go
// Create a structured search request
searchReq := &search.StructuredSearchRequest{
    IndexID: "your-index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Limit: 20,
        Sort:  []string{"date:desc"},
        // Add facets
        Facets:    []string{"keywords", "status"},
        FacetSize: 10,
    },
    // Add additional parameters
    Extra: map[string]interface{}{
        "highlight": map[string]interface{}{
            "fields": map[string]interface{}{
                "content": map[string]interface{}{},
            },
        },
    },
}

// Perform the search
results, err := client.StructuredSearch(context.Background(), searchReq)
if err != nil {
    // Handle error
}

// Process results
for _, result := range results.Results {
    fmt.Printf("Subject: %s (Score: %.2f)\n", result.Subject, result.Score)
    
    // Access highlights if available
    if highlights, ok := result.Highlight["content"]; ok {
        for _, highlight := range highlights {
            fmt.Printf("Highlight: %s\n", highlight)
        }
    }
}

// Process facets
for _, facet := range results.Facets {
    fmt.Printf("Facet: %s\n", facet.Name)
    for _, value := range facet.Values {
        fmt.Printf("  %s (%d)\n", value.Value, value.Count)
    }
}
```

## Pagination

The SDK provides several ways to handle pagination:

### Manual Pagination

```go
// Initial search with a limit
searchReq := &search.SearchRequest{
    IndexID: "your-index-id",
    Query:   "research data",
    Options: &search.SearchOptions{
        Limit: 10,
    },
}

results, err := client.Search(context.Background(), searchReq)
if err != nil {
    // Handle error
}

// Process first page
processResults(results.Results)

// If there are more results, fetch the next page
if results.HasMore {
    searchReq.Options.PageToken = results.PageToken
    nextPage, err := client.Search(context.Background(), searchReq)
    if err != nil {
        // Handle error
    }
    
    processResults(nextPage.Results)
}
```

### Using Iterators

```go
// Create a search iterator
searchReq := &search.SearchRequest{
    IndexID: "your-index-id",
    Query:   "research data",
}

// Create iterator with page size
it := client.NewSearchIterator(context.Background(), searchReq, 10)

// Iterate through all pages
for it.Next() {
    resp := it.Response()
    processResults(resp.Results)
    
    fmt.Printf("Processing page with %d results\n", resp.Count)
}

// Check for errors
if err := it.Error(); err != nil {
    // Handle error
}
```

### Using SearchAll to Collect All Results

```go
// Search request
searchReq := &search.SearchRequest{
    IndexID: "your-index-id",
    Query:   "research data",
}

// Get all results at once
allResults, err := client.SearchAll(context.Background(), searchReq, 100)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d total results\n", len(allResults))
processResults(allResults)
```

## Batch Operations

For large-scale operations, the SDK provides batch functionality:

### Batch Ingest

```go
// Create a large number of documents
var docs []search.SearchDocument
for i := 0; i < 5000; i++ {
    docs = append(docs, search.SearchDocument{
        Subject: fmt.Sprintf("doc%d", i),
        Content: map[string]interface{}{
            "title": fmt.Sprintf("Document %d", i),
            "index": i,
        },
    })
}

// Set up batch options
options := &search.BatchIngestOptions{
    BatchSize:     1000,       // Documents per batch
    MaxConcurrent: 5,          // Max concurrent requests
    TaskIDPrefix:  "batch-ingest", // Prefix for task IDs
    ProgressCallback: func(processed, total int) {
        fmt.Printf("Progress: %d/%d documents (%.2f%%)\n", 
            processed, total, float64(processed)/float64(total)*100)
    },
}

// Perform batch ingest
result, err := client.BatchIngestDocuments(context.Background(), "your-index-id", docs, options)
if err != nil {
    // Handle error
}

fmt.Printf("Ingested %d/%d documents successfully\n", 
    result.SuccessDocuments, result.TotalDocuments)
fmt.Printf("Created %d tasks\n", len(result.TaskIDs))

// Check for any errors
if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        fmt.Printf("Batch error: %v\n", err)
    }
}
```

### Batch Delete

```go
// Create a list of subjects to delete
var subjects []string
for i := 0; i < 5000; i++ {
    subjects = append(subjects, fmt.Sprintf("doc%d", i))
}

// Set up batch options
options := &search.BatchDeleteOptions{
    BatchSize:     1000,
    MaxConcurrent: 5,
    ProgressCallback: func(processed, total int) {
        fmt.Printf("Progress: %d/%d subjects (%.2f%%)\n", 
            processed, total, float64(processed)/float64(total)*100)
    },
}

// Perform batch delete
result, err := client.BatchDeleteDocuments(context.Background(), "your-index-id", subjects, options)
if err != nil {
    // Handle error
}

fmt.Printf("Deleted %d/%d subjects successfully\n", 
    result.SuccessSubjects, result.TotalSubjects)
```

## Error Handling

The SDK provides specialized error handling for search operations:

```go
_, err := client.GetIndex(context.Background(), "non-existent-index")
if err != nil {
    if search.IsIndexNotFoundError(err) {
        fmt.Println("Index does not exist, creating it...")
        // Create index
    } else if search.IsPermissionDeniedError(err) {
        fmt.Println("Permission denied to access this index")
    } else if search.IsRateLimitError(err) {
        fmt.Println("Rate limit exceeded, try again later")
    } else {
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

### Using Search Errors

```go
_, err := client.SearchAll(context.Background(), searchReq, 100)
if err != nil {
    // Extract search error details
    if searchErr, ok := search.AsSearchError(err); ok {
        fmt.Printf("Search error: %s (Code: %s, Status: %d)\n", 
            searchErr.Message, searchErr.Code, searchErr.Status)
        
        if searchErr.RequestID != "" {
            fmt.Printf("Request ID: %s (for support reference)\n", searchErr.RequestID)
        }
    } else {
        fmt.Printf("Non-search error: %v\n", err)
    }
}
```

## Task Management

Many operations in Globus Search are asynchronous and return a task ID. The SDK provides functions to monitor and wait for these tasks:

### Checking Task Status

```go
// Get the status of a task
taskStatus, err := client.GetTaskStatus(context.Background(), "task-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task status: %s\n", taskStatus.State)
fmt.Printf("Documents: %d/%d processed\n", 
    taskStatus.SuccessDocuments, taskStatus.TotalDocuments)

if taskStatus.State == "FAILED" {
    fmt.Printf("Failed documents: %d\n", taskStatus.FailedDocuments)
    for _, subject := range taskStatus.FailedSubjects {
        fmt.Printf("Failed subject: %s\n", subject)
    }
}
```

### Waiting for Tasks to Complete

```go
// Wait for multiple tasks to complete
taskIDs := []string{"task1", "task2", "task3"}
results, err := client.WaitForTasks(ctx, taskIDs, 2*time.Second)
if err != nil {
    // Handle error (like context timeout)
}

// Process results
for i, result := range results {
    fmt.Printf("Task %s: %s\n", taskIDs[i], result.State)
    if result.State == "SUCCESS" {
        fmt.Printf("  Completed at: %s\n", result.CompletedAt)
        fmt.Printf("  Success: %d/%d\n", result.SuccessDocuments, result.TotalDocuments)
    } else if result.State == "FAILED" {
        fmt.Printf("  Failed: %d/%d\n", result.FailedDocuments, result.TotalDocuments)
    }
}
```

## Complete Example

Here's a complete example showing many features of the Search client:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Get access token from environment
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if accessToken == "" {
        fmt.Println("GLOBUS_ACCESS_TOKEN environment variable not set")
        os.Exit(1)
    }

    // Create search client
    client := search.NewClient(accessToken)

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // 1. Create a new index
    createReq := &search.IndexCreateRequest{
        DisplayName: "Research Data " + time.Now().Format("2006-01-02"),
        Description: "Index for research project data",
        DefinitionDocument: map[string]interface{}{
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "title": map[string]interface{}{
                        "type": "text",
                    },
                    "project": map[string]interface{}{
                        "type": "keyword",
                    },
                    "date": map[string]interface{}{
                        "type": "date",
                        "format": "yyyy-MM-dd",
                    },
                    "location": map[string]interface{}{
                        "type": "geo_point",
                    },
                },
            },
        },
    }

    index, err := client.CreateIndex(ctx, createReq)
    if err != nil {
        fmt.Printf("Error creating index: %v\n", err)
        os.Exit(1)
    }

    indexID := index.ID
    fmt.Printf("Created index with ID: %s\n", indexID)

    // 2. Ingest documents
    fmt.Println("Generating documents...")
    var docs []search.SearchDocument
    projects := []string{"Alpha", "Beta", "Gamma"}
    
    for i := 0; i < 500; i++ {
        project := projects[i%3]
        date := time.Now().AddDate(0, 0, -i%90).Format("2006-01-02")
        
        docs = append(docs, search.SearchDocument{
            Subject: fmt.Sprintf("doc-%s-%03d", project, i),
            Content: map[string]interface{}{
                "title":    fmt.Sprintf("%s Project Report %d", project, i),
                "project":  project,
                "date":     date,
                "content":  fmt.Sprintf("This is report %d for the %s project.", i, project),
                "location": map[string]interface{}{
                    "lat": 37.0 + float64(i%10)/10.0,
                    "lon": -122.0 + float64(i%20)/10.0,
                },
            },
            VisibleTo: []string{"public"},
        })
    }

    // Perform batch ingest
    fmt.Println("Batch ingesting documents...")
    batchOpts := &search.BatchIngestOptions{
        BatchSize:     100,
        MaxConcurrent: 5,
        ProgressCallback: func(processed, total int) {
            fmt.Printf("Ingesting: %d/%d documents (%.1f%%)\n", 
                processed, total, float64(processed)/float64(total)*100)
        },
    }

    batchResult, err := client.BatchIngestDocuments(ctx, indexID, docs, batchOpts)
    if err != nil {
        fmt.Printf("Error in batch ingest: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Ingested %d/%d documents successfully\n", 
        batchResult.SuccessDocuments, batchResult.TotalDocuments)

    // Wait for indexing to complete
    fmt.Println("Waiting for indexing to complete...")
    taskResults, err := client.WaitForTasks(ctx, batchResult.TaskIDs, 2*time.Second)
    if err != nil {
        fmt.Printf("Error waiting for tasks: %v\n", err)
        // Continue anyway
    }

    successCount := 0
    for _, result := range taskResults {
        if result.State == "SUCCESS" {
            successCount++
        }
    }
    fmt.Printf("%d/%d tasks completed successfully\n", successCount, len(taskResults))

    // Wait a bit for indexing to be fully available
    time.Sleep(5 * time.Second)

    // 3. Perform a simple search
    fmt.Println("\nPerforming simple search...")
    simpleSearchReq := &search.SearchRequest{
        IndexID: indexID,
        Query:   "project report",
        Options: &search.SearchOptions{
            Limit: 5,
            Sort:  []string{"date:desc"},
        },
    }

    simpleResults, err := client.Search(ctx, simpleSearchReq)
    if err != nil {
        fmt.Printf("Error in simple search: %v\n", err)
    } else {
        fmt.Printf("Simple search found %d results (showing first 5 of %d total)\n", 
            simpleResults.Count, simpleResults.Total)
        
        for i, result := range simpleResults.Results {
            title := result.Content["title"].(string)
            date := result.Content["date"].(string)
            fmt.Printf("%d. %s (%s) - Score: %.2f\n", i+1, title, date, result.Score)
        }
    }

    // 4. Perform an advanced search
    fmt.Println("\nPerforming advanced search...")
    
    // Build complex query
    query := search.NewBoolQuery().
        AddMust(search.NewMatchQuery("title", "Report")).
        AddMust(search.NewRangeQuery("date").
            WithGTE(time.Now().AddDate(0, -1, 0).Format("2006-01-02")).
            WithFormat("yyyy-MM-dd")).
        AddShould(search.NewTermQuery("project", "Alpha")).
        AddShould(search.NewTermQuery("project", "Beta")).
        SetMinimumShouldMatch(1)

    advancedReq := &search.StructuredSearchRequest{
        IndexID: indexID,
        Query:   query,
        Options: &search.SearchOptions{
            Limit:  5,
            Sort:   []string{"date:desc"},
            Facets: []string{"project"},
        },
        Extra: map[string]interface{}{
            "highlight": map[string]interface{}{
                "fields": map[string]interface{}{
                    "title": map[string]interface{}{},
                    "content": map[string]interface{}{},
                },
            },
        },
    }

    advancedResults, err := client.StructuredSearch(ctx, advancedReq)
    if err != nil {
        fmt.Printf("Error in advanced search: %v\n", err)
    } else {
        fmt.Printf("Advanced search found %d results (showing first 5 of %d total)\n", 
            advancedResults.Count, advancedResults.Total)
        
        for i, result := range advancedResults.Results {
            title := result.Content["title"].(string)
            project := result.Content["project"].(string)
            date := result.Content["date"].(string)
            
            fmt.Printf("%d. %s - %s (%s) - Score: %.2f\n", 
                i+1, title, project, date, result.Score)
                
            // Show highlights if available
            if highlights, ok := result.Highlight["title"]; ok && len(highlights) > 0 {
                fmt.Printf("   Highlight: %s\n", highlights[0])
            }
        }
        
        // Print facets
        if len(advancedResults.Facets) > 0 {
            fmt.Println("\nFacets:")
            for _, facet := range advancedResults.Facets {
                fmt.Printf("  %s:\n", facet.Name)
                for _, value := range facet.Values {
                    fmt.Printf("    %s: %d\n", value.Value, value.Count)
                }
            }
        }
    }

    // 5. Find documents near a location
    fmt.Println("\nPerforming geo search...")
    geoQuery := search.NewGeoDistanceQuery("location", "100km", 37.5, -122.0)
    
    geoReq := &search.StructuredSearchRequest{
        IndexID: indexID,
        Query:   geoQuery,
        Options: &search.SearchOptions{
            Limit: 5,
        },
    }
    
    geoResults, err := client.StructuredSearch(ctx, geoReq)
    if err != nil {
        fmt.Printf("Error in geo search: %v\n", err)
    } else {
        fmt.Printf("Geo search found %d results within 100km of (37.5, -122.0)\n", 
            geoResults.Total)
        
        for i, result := range geoResults.Results[:5] {
            title := result.Content["title"].(string)
            location := result.Content["location"].(map[string]interface{})
            lat := location["lat"].(float64)
            lon := location["lon"].(float64)
            
            fmt.Printf("%d. %s at (%.1f, %.1f) - Score: %.2f\n", 
                i+1, title, lat, lon, result.Score)
        }
    }

    // 6. Clean up (delete index if it was just a test)
    if os.Getenv("KEEP_INDEX") != "true" {
        fmt.Println("\nCleaning up (deleting index)...")
        err := client.DeleteIndex(ctx, indexID)
        if err != nil {
            fmt.Printf("Error deleting index: %v\n", err)
        } else {
            fmt.Println("Index deleted successfully")
        }
    } else {
        fmt.Printf("\nIndex %s has been preserved. You can access it via the Globus Search API.\n", indexID)
    }
}
```

This document covered the major features of the Globus Search client in the Go SDK. For specific details about parameters and return types, refer to the GoDoc documentation or explore the source code.