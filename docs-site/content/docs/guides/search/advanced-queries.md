---
title: "SPDX-License-Identifier: Apache-2.0"
---
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Advanced Search Queries

_Last Updated: May 3, 2025_
_Compatible with SDK versions: v0.1.0 and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

This guide covers the advanced query capabilities of the Search service in the Globus Go SDK, which enables sophisticated search operations against Globus Search indexes.

## Table of Contents

- [Overview](#overview)
- [Query Types](#query-types)
  - [Simple Query](#simple-query)
  - [Match Query](#match-query)
  - [Term Query](#term-query)
  - [Range Query](#range-query)
  - [Boolean Query](#boolean-query)
  - [Exists Query](#exists-query)
  - [Prefix Query](#prefix-query)
  - [Wildcard Query](#wildcard-query)
  - [Geo Distance Query](#geo-distance-query)
- [Building Complex Queries](#building-complex-queries)
  - [Builder Pattern](#builder-pattern)
  - [Combining Queries](#combining-queries)
  - [Nested Queries](#nested-queries)
- [Executing Queries](#executing-queries)
  - [Basic Query Execution](#basic-query-execution)
  - [Using Search Options](#using-search-options)
  - [Handling Results](#handling-results)
  - [Pagination](#pagination)
- [Query Parsing](#query-parsing)
- [Best Practices](#best-practices)
  - [Performance Optimization](#performance-optimization)
  - [Query Structure](#query-structure)
  - [Resource Usage](#resource-usage)
- [Examples](#examples)
  - [Document Search](#document-search)
  - [Faceted Search](#faceted-search)
  - [Advanced Filtering](#advanced-filtering)
- [Troubleshooting](#troubleshooting)
- [Related Topics](#related-topics)

## Overview

The Globus Go SDK provides a robust query system for the Search service, enabling you to build structured queries for efficiently searching Globus Search indexes. The SDK implements a builder pattern that makes it easy to construct complex queries while maintaining type safety and readability.

## Query Types

The SDK supports multiple query types, each designed for different search scenarios.

### Simple Query

A simple query is the most basic form of search, using a query string to match across multiple fields.

```go
// Create a simple query
simpleQuery := search.NewSimpleQuery("machine learning")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   simpleQuery,
}
```

Simple queries are useful for general text searches across all indexed fields.

### Match Query

A match query searches for a specific term in a specified field, with text analysis applied.

```go
// Match a specific term in a field
matchQuery := search.NewMatchQuery("title", "machine learning")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   matchQuery,
}
```

Match queries are ideal for text fields where you want to match words after analysis (tokenization, stemming, etc.).

### Term Query

A term query looks for an exact match of a term in a field, without text analysis.

```go
// Exact match for a term
termQuery := search.NewTermQuery("status", "active")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   termQuery,
}
```

Term queries are best for structured fields like status codes, ids, or keywords.

### Range Query

A range query finds documents with field values within a specified range.

```go
// Range query for dates
rangeQuery := search.NewRangeQuery("created_at").
    WithGTE("2023-01-01").
    WithLT("2024-01-01").
    WithFormat("yyyy-MM-dd")

// Range query for numeric values
sizeRangeQuery := search.NewRangeQuery("file_size").
    WithGT(1024).
    WithLTE(1048576)

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   rangeQuery,
}
```

Range queries support multiple comparison operators:
- `WithGT`: Greater than
- `WithGTE`: Greater than or equal to
- `WithLT`: Less than
- `WithLTE`: Less than or equal to

For date fields, you can specify the format and time zone:
- `WithFormat`: Date format (e.g., "yyyy-MM-dd")
- `WithTimeZone`: Time zone for date comparison

### Boolean Query

A boolean query combines multiple queries with boolean logic.

```go
// Create a boolean query
boolQuery := search.NewBoolQuery().
    AddMust(search.NewMatchQuery("title", "research")).
    AddMustNot(search.NewTermQuery("status", "archived")).
    AddShould(search.NewTermQuery("tags", "important")).
    AddShould(search.NewTermQuery("tags", "urgent")).
    SetMinimumShouldMatch(1)

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   boolQuery,
}
```

Boolean queries support three types of clauses:
- `Must`: Documents MUST match these conditions (AND logic)
- `MustNot`: Documents MUST NOT match these conditions (NOT logic)
- `Should`: Documents SHOULD match these conditions (OR logic)

The `SetMinimumShouldMatch` method allows you to specify how many of the "should" clauses must match for a document to be included.

### Exists Query

An exists query finds documents where a specified field exists and is not null.

```go
// Find documents with a specific field
existsQuery := search.NewExistsQuery("attachment")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   existsQuery,
}
```

This is useful for finding documents that have a certain property or attribute.

### Prefix Query

A prefix query finds terms that begin with a specified prefix.

```go
// Find terms starting with a prefix
prefixQuery := search.NewPrefixQuery("title", "neuro")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   prefixQuery,
}
```

Useful for implementing autocomplete or partial matching functionality.

### Wildcard Query

A wildcard query supports pattern matching with `*` (multiple characters) and `?` (single character) wildcards.

```go
// Find terms matching a pattern
wildcardQuery := search.NewWildcardQuery("filename", "data_*.csv")

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   wildcardQuery,
}
```

Wildcard queries are powerful but can be expensive, so use them judiciously.

### Geo Distance Query

A geo distance query finds documents with geo points within a specified distance from a location.

```go
// Find locations within 10km of a point
geoQuery := search.NewGeoDistanceQuery(
    "location",
    "10km",
    37.7749, // latitude
    -122.4194, // longitude
)

// Execute the query
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   geoQuery,
}
```

Useful for geo-based searches in applications that store location data.

## Building Complex Queries

The SDK's query system is designed to be compositional, allowing you to build complex queries by combining simpler ones.

### Builder Pattern

Each query type implements a builder pattern, enabling fluent method chaining:

```go
// Range query with chained methods
rangeQuery := search.NewRangeQuery("created_at").
    WithGTE("2023-01-01").
    WithLT("2024-01-01").
    WithFormat("yyyy-MM-dd").
    WithTimeZone("UTC")
```

This pattern makes queries more readable and maintainable.

### Combining Queries

Boolean queries allow you to combine different query types:

```go
// Create a complex query
complexQuery := search.NewBoolQuery().
    // Find recent documents
    AddMust(
        search.NewRangeQuery("created_at").
            WithGTE("2023-01-01").
            WithFormat("yyyy-MM-dd")
    ).
    // With a specific field
    AddMust(
        search.NewExistsQuery("metadata.author")
    ).
    // Not in a specific status
    AddMustNot(
        search.NewTermQuery("status", "deleted")
    ).
    // Preferring certain keywords
    AddShould(
        search.NewMatchQuery("content", "important")
    ).
    SetMinimumShouldMatch(0) // Should clauses are optional
```

This approach lets you build arbitrarily complex queries with clear, readable code.

### Nested Queries

You can nest boolean queries for even more complex conditions:

```go
// Create a nested query
outerQuery := search.NewBoolQuery().
    // Must match this condition
    AddMust(
        search.NewBoolQuery().
            // Either match title
            AddShould(search.NewMatchQuery("title", "research")).
            // Or match description
            AddShould(search.NewMatchQuery("description", "research")).
            SetMinimumShouldMatch(1)
    ).
    // Additional filters
    AddMust(
        search.NewRangeQuery("publication_date").
            WithGTE("2020-01-01").
            WithFormat("yyyy-MM-dd")
    )
```

Nested queries allow for sophisticated search logic that can model complex real-world requirements.

## Executing Queries

The SDK provides multiple ways to execute queries against the Globus Search service.

### Basic Query Execution

To execute a query, create a structured search request and submit it:

```go
// Create a client
client, err := search.NewClient(search.WithAccessToken("your-access-token"))
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

// Create a search request
request := &search.StructuredSearchRequest{
    IndexID: "your-index-id",
    Query:   query, // Your query object
}

// Execute the query
ctx := context.Background()
response, err := client.StructuredSearch(ctx, request)
if err != nil {
    log.Fatalf("Search failed: %v", err)
}

// Process results
fmt.Printf("Found %d results out of %d total\n", response.Count, response.Total)
for _, result := range response.Results {
    fmt.Printf("Subject: %s, Score: %f\n", result.Subject, result.Score)
    // Access content fields
    if title, ok := result.Content["title"].(string); ok {
        fmt.Printf("Title: %s\n", title)
    }
}
```

### Using Search Options

You can customize your search with various options:

```go
// Create a search request with options
request := &search.StructuredSearchRequest{
    IndexID: "your-index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Limit:             100,           // Number of results to return
        Sort:              []string{"created_at:desc"}, // Sort criteria
        Filter:            "access_level:public",       // Additional filter
        Facets:            []string{"tags", "author"},  // Facet fields
        FacetSize:         10,                          // Max facet values
        IncludeAllContent: true,                        // Include full content
    },
}
```

Options allow you to control pagination, sorting, filtering, and more.

### Handling Results

The search response contains several useful fields:

```go
// Process search results
fmt.Printf("Found %d results out of %d total\n", response.Count, response.Total)
fmt.Printf("Has more results: %v\n", response.HasMore)

// Access result details
for _, result := range response.Results {
    fmt.Printf("Subject: %s, Score: %f\n", result.Subject, result.Score)
    
    // Access content fields - type assertion is needed
    if title, ok := result.Content["title"].(string); ok {
        fmt.Printf("Title: %s\n", title)
    }
    
    // Access highlights (if requested)
    for field, highlights := range result.Highlight {
        fmt.Printf("Highlighted %s: %v\n", field, highlights)
    }
}

// Process facets (if requested)
for _, facet := range response.Facets {
    fmt.Printf("Facet: %s (%s)\n", facet.Name, facet.Type)
    for _, value := range facet.Values {
        fmt.Printf("  %s: %d\n", value.Value, value.Count)
    }
}
```

### Pagination

For large result sets, you can use pagination:

```go
// Using the iterator pattern
it := client.NewStructuredSearchIterator(ctx, request, 100)
for it.Next() {
    resp := it.Response()
    for _, result := range resp.Results {
        // Process each result
        fmt.Printf("Subject: %s\n", result.Subject)
    }
}
if err := it.Error(); err != nil {
    log.Fatalf("Iterator error: %v", err)
}

// Or fetching all results at once
allResults, err := client.StructuredSearchAll(ctx, request, 100)
if err != nil {
    log.Fatalf("Search failed: %v", err)
}
fmt.Printf("Retrieved %d total results\n", len(allResults))
```

The iterator pattern is particularly useful for efficiently processing large result sets.

## Query Parsing

The SDK includes a QueryParser for parsing string-based queries:

```go
// Parse a query string
parser := &search.QueryParser{}
query, err := parser.ParseQuery("created_at:[2023-01-01 TO 2024-01-01] AND status:active")
if err != nil {
    log.Fatalf("Failed to parse query: %v", err)
}

// Execute the parsed query
request := &search.StructuredSearchRequest{
    IndexID: "your-index-id",
    Query:   query,
}
```

This allows you to accept user input in a query string format and convert it to structured queries.

## Best Practices

### Performance Optimization

1. **Be specific**: Use the most specific query type for your needs rather than relying on simple queries.

   ```go
   // Good: Specific field matching
   termQuery := search.NewTermQuery("status", "active")
   
   // Less Efficient: Simple query that must search all fields
   simpleQuery := search.NewSimpleQuery("active")
   ```

2. **Limit results**: Always set reasonable limits on result counts.

   ```go
   request := &search.StructuredSearchRequest{
       IndexID: "your-index-id",
       Query:   query,
       Options: &search.SearchOptions{
           Limit: 100, // Always set a reasonable limit
       },
   }
   ```

3. **Use pagination**: For large result sets, use the iterator pattern rather than fetching all results at once.

4. **Optimize boolean queries**: Put the most selective conditions first in a boolean query.

   ```go
   // More efficient: specific ID check first
   boolQuery := search.NewBoolQuery().
       AddMust(search.NewTermQuery("project_id", specificID)).
       AddMust(search.NewRangeQuery("created_at").WithGTE("2023-01-01"))
   ```

### Query Structure

1. **Use the right query for the job**:
   - Use `TermQuery` for exact matches on keyword fields
   - Use `MatchQuery` for text fields that need analysis
   - Use `RangeQuery` for dates and numeric values

2. **Combine queries effectively**:
   - Use `Must` for criteria that must be matched (AND)
   - Use `Should` for criteria that should be matched (OR)
   - Use `MustNot` for exclusions (NOT)

3. **Avoid expensive queries** when possible:
   - Wildcard queries with leading wildcards (`*text`) are expensive
   - Very broad terms in simple queries scan many documents

### Resource Usage

1. **Batch processing**: When searching for many items, batch your requests.

   ```go
   // Process in manageable chunks
   const batchSize = 100
   it := client.NewStructuredSearchIterator(ctx, request, batchSize)
   for it.Next() {
       // Process batch
       processBatch(it.Response().Results)
   }
   ```

2. **Limit fields**: Specify only the fields you need to reduce response size.

   ```go
   // Add query parameter to limit returned fields
   request.Extra = map[string]interface{}{
       "_source": []string{"title", "author", "created_at"},
   }
   ```

3. **Use facets carefully**: Request only the facets you need with appropriate sizes.

## Examples

### Document Search

This example searches for research papers in a specific date range:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Create a client
    client, err := search.NewClient(search.WithAccessToken("your-access-token"))
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Create a complex query for research papers
    query := search.NewBoolQuery().
        // Must be a research paper
        AddMust(search.NewTermQuery("document_type", "research_paper")).
        // Published in 2023
        AddMust(
            search.NewRangeQuery("publication_date").
                WithGTE("2023-01-01").
                WithLT("2024-01-01").
                WithFormat("yyyy-MM-dd")
        ).
        // Should be about machine learning or AI
        AddShould(search.NewMatchQuery("title", "machine learning")).
        AddShould(search.NewMatchQuery("title", "artificial intelligence")).
        AddShould(search.NewMatchQuery("keywords", "machine learning")).
        AddShould(search.NewMatchQuery("keywords", "artificial intelligence")).
        // At least one should clause must match
        SetMinimumShouldMatch(1)
    
    // Create the search request
    request := &search.StructuredSearchRequest{
        IndexID: "research-papers-index",
        Query:   query,
        Options: &search.SearchOptions{
            Limit:     20,
            Sort:      []string{"citation_count:desc"},
            Facets:    []string{"author", "journal", "keywords"},
            FacetSize: 5,
        },
    }
    
    // Execute the search
    ctx := context.Background()
    response, err := client.StructuredSearch(ctx, request)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    
    // Process results
    fmt.Printf("Found %d research papers out of %d total\n", response.Count, response.Total)
    for i, result := range response.Results {
        fmt.Printf("Result %d: %s (Score: %.2f)\n", i+1, result.Content["title"], result.Score)
        fmt.Printf("  Author: %s\n", result.Content["author"])
        fmt.Printf("  Journal: %s\n", result.Content["journal"])
        fmt.Printf("  Published: %s\n", result.Content["publication_date"])
        fmt.Printf("  Citations: %v\n", result.Content["citation_count"])
        fmt.Println()
    }
    
    // Process facets
    fmt.Println("Top Authors:")
    for _, value := range response.Facets[0].Values {
        fmt.Printf("  %s (%d papers)\n", value.Value, value.Count)
    }
    
    fmt.Println("Top Journals:")
    for _, value := range response.Facets[1].Values {
        fmt.Printf("  %s (%d papers)\n", value.Value, value.Count)
    }
    
    fmt.Println("Top Keywords:")
    for _, value := range response.Facets[2].Values {
        fmt.Printf("  %s (%d papers)\n", value.Value, value.Count)
    }
}
```

### Faceted Search

This example demonstrates how to use facets for filtering and analysis:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Create a client
    client, err := search.NewClient(search.WithAccessToken("your-access-token"))
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Simple search with facets
    query := search.NewSimpleQuery("climate change")
    
    // Create the search request with facets
    request := &search.StructuredSearchRequest{
        IndexID: "datasets-index",
        Query:   query,
        Options: &search.SearchOptions{
            Limit:     0,  // We only want facets, not results
            Facets:    []string{"data_type", "organization", "year", "license"},
            FacetSize: 10,
        },
    }
    
    // Execute the search
    ctx := context.Background()
    response, err := client.StructuredSearch(ctx, request)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    
    // Process facets
    fmt.Printf("Analysis of %d climate change datasets:\n\n", response.Total)
    
    for _, facet := range response.Facets {
        fmt.Printf("%s breakdown:\n", facet.Name)
        for _, value := range facet.Values {
            fmt.Printf("  %s: %d datasets\n", value.Value, value.Count)
        }
        fmt.Println()
    }
    
    // Now use facet for filtering
    dataTypeFilter := "csv"
    yearFilter := "2022"
    
    // Build a new query with the filter
    filteredQuery := search.NewBoolQuery().
        AddMust(search.NewSimpleQuery("climate change")).
        AddMust(search.NewTermQuery("data_type", dataTypeFilter)).
        AddMust(search.NewTermQuery("year", yearFilter))
    
    // Create a new request with filters
    filteredRequest := &search.StructuredSearchRequest{
        IndexID: "datasets-index",
        Query:   filteredQuery,
        Options: &search.SearchOptions{
            Limit: 10,
            Sort:  []string{"downloads:desc"},
        },
    }
    
    // Execute the filtered search
    filteredResponse, err := client.StructuredSearch(ctx, filteredRequest)
    if err != nil {
        log.Fatalf("Filtered search failed: %v", err)
    }
    
    // Process filtered results
    fmt.Printf("Top %d climate change datasets in CSV format from 2022:\n\n", 
        filteredResponse.Count)
    
    for i, result := range filteredResponse.Results {
        fmt.Printf("%d. %s\n", i+1, result.Content["title"])
        fmt.Printf("   Organization: %s\n", result.Content["organization"])
        fmt.Printf("   Downloads: %v\n", result.Content["downloads"])
        fmt.Printf("   License: %s\n", result.Content["license"])
        fmt.Println()
    }
}
```

### Advanced Filtering

This example shows how to combine multiple query types for sophisticated filtering:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    // Create a client
    client, err := search.NewClient(search.WithAccessToken("your-access-token"))
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Calculate date ranges
    now := time.Now()
    sixMonthsAgo := now.AddDate(0, -6, 0)
    lastYear := now.AddDate(-1, 0, 0)
    
    // Format dates
    sixMonthsAgoStr := sixMonthsAgo.Format("2006-01-02")
    lastYearStr := lastYear.Format("2006-01-02")
    
    // Build a complex query for active projects with recent updates
    query := search.NewBoolQuery().
        // Must be an active project
        AddMust(search.NewTermQuery("status", "active")).
        // Must be publicly accessible
        AddMust(search.NewTermQuery("access_level", "public")).
        // Must have been created at least a year ago
        AddMust(
            search.NewRangeQuery("created_at").
                WithLTE(lastYearStr).
                WithFormat("yyyy-MM-dd")
        ).
        // Must have been updated in the last 6 months
        AddMust(
            search.NewRangeQuery("updated_at").
                WithGTE(sixMonthsAgoStr).
                WithFormat("yyyy-MM-dd")
        ).
        // Must have a description field
        AddMust(search.NewExistsQuery("description")).
        // Must not be a test project
        AddMustNot(search.NewPrefixQuery("name", "test_")).
        // Should match any of these topics
        AddShould(search.NewTermQuery("topics", "genomics")).
        AddShould(search.NewTermQuery("topics", "proteomics")).
        AddShould(search.NewTermQuery("topics", "bioinformatics")).
        // At least one topic should match
        SetMinimumShouldMatch(1)
    
    // Create the search request with specific sorting
    request := &search.StructuredSearchRequest{
        IndexID: "projects-index",
        Query:   query,
        Options: &search.SearchOptions{
            Limit: 20,
            Sort:  []string{"activity_score:desc", "updated_at:desc"},
        },
        // Add extra highlighting parameters
        Extra: map[string]interface{}{
            "highlight": map[string]interface{}{
                "fields": map[string]interface{}{
                    "description": map[string]interface{}{
                        "fragment_size": 150,
                        "number_of_fragments": 1,
                    },
                },
            },
        },
    }
    
    // Execute the search
    ctx := context.Background()
    response, err := client.StructuredSearch(ctx, request)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    
    // Process results
    fmt.Printf("Found %d active bioinformatics projects with recent updates\n\n", 
        response.Count)
    
    for i, result := range response.Results {
        fmt.Printf("%d. %s\n", i+1, result.Content["name"])
        
        // Show topics if available
        if topics, ok := result.Content["topics"].([]interface{}); ok {
            fmt.Printf("   Topics: ")
            for i, topic := range topics {
                if i > 0 {
                    fmt.Print(", ")
                }
                fmt.Print(topic)
            }
            fmt.Println()
        }
        
        // Show project dates
        fmt.Printf("   Created: %s, Last Updated: %s\n", 
            result.Content["created_at"], 
            result.Content["updated_at"])
        
        // Show highlight if available
        if highlights, ok := result.Highlight["description"]; ok && len(highlights) > 0 {
            fmt.Printf("   Highlight: %s\n", highlights[0])
        }
        
        fmt.Println()
    }
}
```

## Troubleshooting

### Common Issues

1. **Query Returning Too Many Results**

   **Solution**: Make your query more specific by adding additional constraints:

   ```go
   // Add constraints to narrow results
   query = search.NewBoolQuery().
       AddMust(existingQuery).
       AddMust(search.NewRangeQuery("created_at").WithGT("2023-01-01"))
   ```

2. **Query Returning No Results**

   **Solution**: Verify field names and values, and consider relaxing constraints:

   ```go
   // More permissive query
   query = search.NewBoolQuery().
       AddShould(query1).
       AddShould(query2).
       SetMinimumShouldMatch(1)
   ```

3. **Pagination Issues**

   **Solution**: Use the iterator pattern for reliable pagination:

   ```go
   it := client.NewStructuredSearchIterator(ctx, request, 100)
   for it.Next() {
       // Process current page
   }
   ```

4. **Type Assertion Errors**

   **Solution**: Always use type assertions carefully when accessing content fields:

   ```go
   // Safe way to access content fields
   if value, ok := result.Content["field"].(string); ok {
       // Use the string value
   } else {
       // Handle the case where it's not a string
   }
   ```

## Related Topics

- [Search Service Overview](../search.md)
- [Document Indexing Guide](./indexing.md)
- [Performance Optimization](../../topics/performance.md)
- [Authentication Guide](../authentication.md)
- [Error Handling](../../topics/error-handling.md)