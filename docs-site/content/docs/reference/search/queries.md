---
title: "Search Service: Queries"
---
# Search Service: Queries

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Search service provides multiple ways to query for documents, from simple text-based queries to complex structured queries. This document covers the basic query functionality.

## Basic Text Search

The simplest way to search is with a text-based query:

```go
// Perform a simple text search
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
    Options: &search.SearchOptions{
        Limit: 10,
    },
}

response, err := client.Search(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d results\n", response.Count)
for _, result := range response.Results {
    fmt.Printf("- %s (Score: %.2f)\n", result.Subject, result.Score)
    fmt.Printf("  Title: %s\n", result.Content["title"])
}
```

## Query Syntax

The basic search query supports several advanced features:

### Field-specific Search

Search within specific fields:

```go
// Search in specific fields
query := "title:machine learning"  // Search in the title field
query := "author:\"Jane Smith\""   // Search for an exact phrase in the author field
```

### Boolean Operators

Combine search terms with boolean operators:

```go
// Boolean operators
query := "machine AND learning"     // Both terms must be present
query := "machine OR learning"      // Either term must be present
query := "machine NOT supervised"   // machine must be present, supervised must not be present
```

### Grouping

Group terms with parentheses:

```go
// Grouping
query := "(machine OR deep) AND learning"  // Either machine or deep, along with learning
```

### Wildcards

Use wildcards in search terms:

```go
// Wildcards
query := "mach*"      // Terms starting with "mach"
query := "m?chine"    // Single character wildcard (matches "machine", "mechine", etc.)
```

### Range Queries

Search for values within a range:

```go
// Range queries
query := "year:[2020 TO 2023]"           // Inclusive range
query := "created:[2023-01-01 TO *]"     // From a date to now
query := "temperature:{32 TO 100}"       // Exclusive range
```

### Fuzzy Search

Search for terms with similar spelling:

```go
// Fuzzy search
query := "machne~"   // Matches "machine" with fuzzy matching
```

### Proximity Search

Search for terms within a certain distance of each other:

```go
// Proximity search
query := "\"machine learning\"~5"   // Terms within 5 words of each other
```

### Boosting

Boost the importance of certain terms:

```go
// Boosting
query := "machine^2 learning"  // machine is twice as important as learning
```

## Search Options

The `SearchOptions` struct provides additional control over search behavior:

```go
// Create search options
options := &search.SearchOptions{
    Limit:           10,                   // Maximum results to return
    Offset:          0,                    // Starting position (for pagination)
    Sort:            []string{"_score.desc", "created.asc"}, // Sort criteria
    Fields:          []string{"title", "description", "created"}, // Fields to return
    IncludeMetadata: true,                 // Include metadata in results
    Filters:         []string{"status:published", "year:[2020 TO 2023]"}, // Post-query filters
}
```

### Pagination

Control the number of results and starting position:

```go
// First page
options := &search.SearchOptions{
    Limit:  10,   // 10 results per page
    Offset: 0,    // First page
}

// Second page
options := &search.SearchOptions{
    Limit:  10,   // 10 results per page
    Offset: 10,   // Second page
}
```

### Sorting

Control the order of results:

```go
// Sort by relevance (default)
options := &search.SearchOptions{
    Sort: []string{"_score.desc"},
}

// Sort by date
options := &search.SearchOptions{
    Sort: []string{"created.desc"},
}

// Multiple sort criteria
options := &search.SearchOptions{
    Sort: []string{"year.desc", "title.asc"},
}
```

### Field Selection

Control which fields are returned:

```go
// Return only specific fields
options := &search.SearchOptions{
    Fields: []string{"title", "description", "created"},
}
```

### Filters

Filter results after the query:

```go
// Filter results
options := &search.SearchOptions{
    Filters: []string{
        "status:published",          // Only published documents
        "year:[2020 TO 2023]",       // Date range filter
        "tags:research",             // Tag filter
    },
}
```

## Search Response

The search response contains the results and metadata:

```go
type SearchResponse struct {
    Count       int                    // Total number of matching results
    Results     []SearchResult         // Results for this page
    Facets      map[string][]FacetValue // Facet information
    HasNextPage bool                   // Whether more pages exist
    Offset      int                    // Current offset
    Total       int                    // Total results across all pages
}

type SearchResult struct {
    Subject    string                 // Unique identifier
    Score      float64                // Relevance score
    Content    map[string]interface{} // Document content
    Metadata   map[string]interface{} // Document metadata
    Highlights map[string][]string    // Highlighted matches
}
```

## Processing Search Results

Extract and use search results:

```go
// Process search results
for _, result := range response.Results {
    // Access the subject (unique identifier)
    fmt.Println("Subject:", result.Subject)
    
    // Access the relevance score
    fmt.Println("Score:", result.Score)
    
    // Access content fields
    title, ok := result.Content["title"].(string)
    if ok {
        fmt.Println("Title:", title)
    }
    
    description, ok := result.Content["description"].(string)
    if ok {
        fmt.Println("Description:", description)
    }
    
    // Access metadata
    if result.Metadata != nil {
        link, ok := result.Metadata["link"].(string)
        if ok {
            fmt.Println("Link:", link)
        }
    }
    
    // Access highlights
    if result.Highlights != nil {
        for field, highlights := range result.Highlights {
            fmt.Printf("Highlights in %s:\n", field)
            for _, highlight := range highlights {
                fmt.Printf("  - %s\n", highlight)
            }
        }
    }
}
```

## Paginating Through Results

For manual pagination:

```go
// Manual pagination through results
var allResults []search.SearchResult
offset := 0
limit := 100

for {
    response, err := client.Search(ctx, &search.SearchRequest{
        IndexID: "index-id",
        Query:   "machine learning",
        Options: &search.SearchOptions{
            Limit:  limit,
            Offset: offset,
        },
    })
    if err != nil {
        // Handle error
        break
    }
    
    // Add results to our collection
    allResults = append(allResults, response.Results...)
    
    // Check if we've retrieved all results
    if !response.HasNextPage {
        break
    }
    
    // Update offset for next page
    offset += limit
}

fmt.Printf("Retrieved %d total results\n", len(allResults))
```

## Using Search Iterator

For easier pagination, use the iterator:

```go
// Create a search iterator
iterator, err := client.SearchIterator(ctx, &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
    Options: &search.SearchOptions{
        Limit: 100, // Results per page
    },
})
if err != nil {
    // Handle error
}

// Iterate through all pages
var allResults []search.SearchResult
for iterator.HasNext() {
    response, err := iterator.Next()
    if err != nil {
        // Handle error
        break
    }
    
    // Process this page of results
    fmt.Printf("Processing page with %d results\n", len(response.Results))
    
    // Add to our collection
    allResults = append(allResults, response.Results...)
}

fmt.Printf("Retrieved %d total results\n", len(allResults))
```

## Search All Method

For retrieving all results in one call:

```go
// Retrieve all matching results with one call
allResults, err := client.SearchAll(ctx, &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
})
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d total results\n", len(allResults))
```

## Faceted Search

Retrieve facet information alongside search results:

```go
// Request facets in search options
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
    Options: &search.SearchOptions{
        Facets: []search.Facet{
            {
                Name:        "file_type",
                Size:        10,            // Number of facet values to return
                Aggregation: "terms",       // Type of aggregation
            },
            {
                Name:        "year",
                Size:        5,
                Aggregation: "terms",
            },
        },
    },
}

response, err := client.Search(ctx, request)
if err != nil {
    // Handle error
}

// Process facets
if response.Facets != nil {
    for name, values := range response.Facets {
        fmt.Printf("Facet: %s\n", name)
        for _, value := range values {
            fmt.Printf("  - %s: %d\n", value.Value, value.Count)
        }
    }
}
```

## Highlighting

Request highlighted matches in search results:

```go
// Request highlighting in search options
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
    Options: &search.SearchOptions{
        Highlight: true,               // Enable highlighting
        HighlightFields: []string{     // Fields to highlight
            "title", "description", "content",
        },
    },
}

response, err := client.Search(ctx, request)
if err != nil {
    // Handle error
}

// Process highlights
for _, result := range response.Results {
    if result.Highlights != nil {
        for field, highlights := range result.Highlights {
            fmt.Printf("Highlights in %s:\n", field)
            for _, highlight := range highlights {
                fmt.Printf("  - %s\n", highlight)
            }
        }
    }
}
```

## Common Query Patterns

### Search with Filters

Combine search terms with filters:

```go
// Search with filters
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "machine learning",
    Options: &search.SearchOptions{
        Filters: []string{
            "created:[2020-01-01 TO 2023-12-31]",
            "status:published",
            "tags:research",
        },
    },
}
```

### Field-specific Search

Search within specific fields:

```go
// Search in specific fields
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "title:machine learning AND author:\"Jane Smith\"",
}
```

### Complex Text Query

Build complex queries with text syntax:

```go
// Complex text query
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "(machine OR deep) AND learning NOT supervised",
}
```

### Fuzzy Search with Proximity

Use fuzzy matching and proximity:

```go
// Fuzzy search with proximity
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "\"machine lerning\"~5 OR machne~",
}
```

### Date Range Query

Search within date ranges:

```go
// Date range query
request := &search.SearchRequest{
    IndexID: "index-id",
    Query:   "created:[2020-01-01 TO 2023-12-31]",
}
```

## Best Practices

1. **Use Specific Fields**: When possible, search within specific fields rather than across all fields
2. **Optimize Pagination**: Use appropriate page sizes (50-100 items per page)
3. **Use the Iterator**: For programmatic access to large result sets
4. **Implement Filtering**: Use filters to narrow results and improve performance
5. **Apply Proper Sorting**: Choose sort criteria based on your use case
6. **Limit Field Selection**: Only request fields you need
7. **Use Facets**: Implement facets for exploratory search interfaces
8. **Enable Highlighting**: Use highlighting to show matches in context
9. **Use Text Analysis**: Understand how fields are analyzed for better queries
10. **Handle Errors**: Implement proper error handling for search failures