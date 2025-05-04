# Search Service: Advanced Queries

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Advanced queries allow you to build complex, structured search expressions for precise data discovery. The Search service provides a comprehensive query building API that supports various query types.

## Query Interface

All query types implement the `Query` interface:

```go
type Query interface {
    QueryType() string
    MarshalJSON() ([]byte, error)
}
```

## Query Types

The Search service supports the following query types:

| Query Type | Description | Usage |
|------------|-------------|-------|
| `SimpleQuery` | Basic text search | Simple text searches |
| `MatchQuery` | Field matching | Match specific field values |
| `TermQuery` | Exact term matching | Match exact, not analyzed terms |
| `RangeQuery` | Range-based queries | Numeric and date ranges |
| `BoolQuery` | Boolean combinations | Combine multiple queries |
| `ExistsQuery` | Field existence checks | Check if a field exists |
| `PrefixQuery` | Prefix matching | Match field values by prefix |
| `WildcardQuery` | Pattern matching | Match using wildcard patterns |
| `GeoDistanceQuery` | Geospatial queries | Find items within a distance |

## Simple Query

The most basic query type for text search:

```go
// Create a simple query
query := search.NewSimpleQuery("machine learning")

// Use in structured search
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
}

response, err := client.StructuredSearch(ctx, request)
```

## Match Query

Match a specific field value:

```go
// Match a specific field
query := search.NewMatchQuery("title", "machine learning")

// With analyzer options
query := search.NewMatchQuery("description", "data science").
    WithAnalyzer("english").
    WithOperator("and")

// Use in structured search
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
}

response, err := client.StructuredSearch(ctx, request)
```

## Term Query

Match an exact term (not analyzed):

```go
// Match an exact term
query := search.NewTermQuery("status", "published")

// Match one of multiple terms
query := search.NewTermQuery("tags", []string{"research", "science"})
```

## Range Query

Match values within a range:

```go
// Numeric range
query := search.NewRangeQuery("year").
    WithGte(2020).
    WithLt(2024)

// Date range
query := search.NewRangeQuery("published_date").
    WithGte("2023-01-01").
    WithLte("2023-12-31")

// With boost
query := search.NewRangeQuery("priority").
    WithGt(5).
    WithLte(10).
    WithBoost(2.0)
```

### Range Operators

| Operator | Description |
|----------|-------------|
| `WithGt(value)` | Greater than |
| `WithGte(value)` | Greater than or equal to |
| `WithLt(value)` | Less than |
| `WithLte(value)` | Less than or equal to |

## Bool Query

Combine multiple queries with boolean logic:

```go
// Create a boolean query
boolQuery := search.NewBoolQuery().
    Must(search.NewMatchQuery("title", "machine learning")).
    MustNot(search.NewTermQuery("status", "draft")).
    Should(
        search.NewMatchQuery("tags", "research"),
        search.NewMatchQuery("tags", "science"),
    ).
    Filter(search.NewRangeQuery("year").WithGte(2020))

// Use in structured search
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   boolQuery,
}

response, err := client.StructuredSearch(ctx, request)
```

### Boolean Operators

| Operator | Description |
|----------|-------------|
| `Must(queries...)` | Queries that must match (AND) |
| `MustNot(queries...)` | Queries that must not match (NOT) |
| `Should(queries...)` | Queries that should match (OR) |
| `Filter(queries...)` | Queries that must match but don't contribute to score |

## Exists Query

Check if a field exists:

```go
// Find documents where a field exists
query := search.NewExistsQuery("abstract")

// Find documents with attachments
query := search.NewExistsQuery("attachments")
```

## Prefix Query

Match field values by prefix:

```go
// Match field values starting with a prefix
query := search.NewPrefixQuery("title", "intro")

// Find file paths in a specific directory
query := search.NewPrefixQuery("path", "/data/project/")
```

## Wildcard Query

Match using wildcard patterns:

```go
// Match using wildcards (* for multiple characters, ? for single character)
query := search.NewWildcardQuery("filename", "data_*.csv")

// Match email pattern
query := search.NewWildcardQuery("email", "*@example.com")
```

## Geo Distance Query

Find items within a geographical distance:

```go
// Find items within 10km of a point
query := search.NewGeoDistanceQuery("location", 47.6062, -122.3321, "10km")

// Find nearby items with higher boost
query := search.NewGeoDistanceQuery("location", 47.6062, -122.3321, "5mi").
    WithBoost(2.0)
```

## Structured Search Request

Use any query type with the structured search API:

```go
// Create a structured search request
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Limit:  10,
        Offset: 0,
        Sort:   []string{"_score.desc", "created.asc"},
        Fields: []string{"title", "description", "created"},
    },
}

// Execute the structured search
response, err := client.StructuredSearch(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d results\n", response.Count)
for _, result := range response.Results {
    fmt.Printf("- %s (Score: %.2f)\n", result.Subject, result.Score)
}
```

## Building Complex Queries

Complex queries can be constructed by combining different query types:

```go
// Build a complex query for scientific datasets
query := search.NewBoolQuery().
    Must(
        // Match either "dataset" or "data collection" in title
        search.NewBoolQuery().
            Should(
                search.NewMatchQuery("title", "dataset"),
                search.NewMatchQuery("title", "data collection"),
            ),
    ).
    Filter(
        // Filter by date range
        search.NewRangeQuery("published_date").
            WithGte("2020-01-01").
            WithLte("2023-12-31"),
        // Filter by file type
        search.NewTermQuery("file_type", []string{"csv", "parquet", "jsonl"}),
    ).
    MustNot(
        // Exclude draft status
        search.NewTermQuery("status", "draft"),
        // Exclude private visibility
        search.NewTermQuery("visibility", "private"),
    ).
    Should(
        // Boost if has citations
        search.NewExistsQuery("citations").WithBoost(1.5),
        // Boost if has DOI
        search.NewExistsQuery("doi").WithBoost(2.0),
    )
```

## Query Boosting

Most query types support boosting to influence relevance scoring:

```go
// Boost a query to increase its importance
query := search.NewMatchQuery("title", "important topic").WithBoost(3.0)

// Combine boosted queries
boolQuery := search.NewBoolQuery().
    Should(
        search.NewMatchQuery("title", "primary").WithBoost(3.0),
        search.NewMatchQuery("description", "primary").WithBoost(1.5),
        search.NewMatchQuery("content", "primary").WithBoost(1.0),
    )
```

## Using Search Results

The structured search returns the same result format as the basic search:

```go
// Process search results
for _, result := range response.Results {
    // Access the subject (unique identifier)
    fmt.Println("Subject:", result.Subject)
    
    // Access the relevance score
    fmt.Println("Score:", result.Score)
    
    // Access content fields as a map
    content := result.Content
    title, ok := content["title"].(string)
    if ok {
        fmt.Println("Title:", title)
    }
    
    description, ok := content["description"].(string)
    if ok {
        fmt.Println("Description:", description)
    }
    
    // Access metadata fields
    metadata := result.Metadata
    link, ok := metadata["link"].(string)
    if ok {
        fmt.Println("Link:", link)
    }
}
```

## Search All Results

For retrieving all matching results across multiple pages:

```go
// Retrieve all results from a structured query
allResults, err := client.StructuredSearchAll(ctx, &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
})
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d total results\n", len(allResults))
```

## Search Iterator

For efficient pagination through large result sets:

```go
// Create an iterator for structured search
iterator, err := client.StructuredSearchIterator(ctx, &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Limit: 100, // Results per page
    },
})
if err != nil {
    // Handle error
}

// Iterate through pages of results
for iterator.HasNext() {
    page, err := iterator.Next()
    if err != nil {
        // Handle error
        break
    }
    
    fmt.Printf("Processing page with %d results\n", len(page.Results))
    // Process results...
}
```

## Faceted Search

Retrieve facet information along with search results:

```go
// Request facets in search options
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Facets: []search.Facet{
            {
                Name:   "file_type",
                Size:   10,
                Aggregation: "terms",
            },
            {
                Name:   "year",
                Size:   5,
                Aggregation: "terms",
            },
        },
    },
}

response, err := client.StructuredSearch(ctx, request)
if err != nil {
    // Handle error
}

// Process facets
for name, values := range response.Facets {
    fmt.Printf("Facet: %s\n", name)
    for _, value := range values {
        fmt.Printf("  - %s: %d\n", value.Value, value.Count)
    }
}
```

## Common Patterns

### Filtering by Date Range and Keywords

```go
// Create a query for documents within a date range containing specific keywords
query := search.NewBoolQuery().
    Must(
        // Match any of these keywords
        search.NewBoolQuery().
            Should(
                search.NewMatchQuery("content", "climate"),
                search.NewMatchQuery("content", "environment"),
                search.NewMatchQuery("content", "temperature"),
            ),
    ).
    Filter(
        // Only from the last 5 years
        search.NewRangeQuery("published_date").
            WithGte("2018-01-01"),
        // Only published documents
        search.NewTermQuery("status", "published"),
    )
```

### Search with Geo-filtering

```go
// Search for documents about "restaurants" within 5km of a location
query := search.NewBoolQuery().
    Must(
        search.NewMatchQuery("content", "restaurant"),
    ).
    Filter(
        search.NewGeoDistanceQuery("location", 37.7749, -122.4194, "5km"),
    )
```

### Field-specific Search with Boosting

```go
// Search across multiple fields with different weights
query := search.NewBoolQuery().
    Should(
        search.NewMatchQuery("title", "neural networks").WithBoost(3.0),
        search.NewMatchQuery("abstract", "neural networks").WithBoost(2.0),
        search.NewMatchQuery("content", "neural networks").WithBoost(1.0),
    )
```

### Existence and Missing Fields

```go
// Find documents with missing required fields
query := search.NewBoolQuery().
    MustNot(
        search.NewExistsQuery("abstract"),
        search.NewExistsQuery("keywords"),
    )
```

## Best Practices

1. **Use the Right Query Type**: Choose the appropriate query type for your use case
2. **Combine with Bool Queries**: Use bool queries to build complex search expressions
3. **Apply Appropriate Boosting**: Use boosting to control result relevance
4. **Filter Non-scoring Criteria**: Use filters for criteria that shouldn't affect scoring
5. **Pagination for Large Results**: Use iterators or the SearchAll method for large result sets
6. **Field Selection**: Use the Fields option to limit returned fields for better performance
7. **Sort Appropriately**: Use appropriate sort criteria for your use case
8. **Error Handling**: Handle query errors, especially InvalidQueryError
9. **Use Facets**: Implement facets for exploratory search interfaces
10. **Test Complex Queries**: Validate complex queries with representative data