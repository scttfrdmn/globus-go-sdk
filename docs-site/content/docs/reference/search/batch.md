---
title: "Search Service: Batch Operations"
---
# Search Service: Batch Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Batch operations in the Search service allow you to efficiently process large numbers of documents for ingestion or deletion. These operations use parallel processing and automatic batching for optimal performance.

## Batch Ingestion

Batch ingestion allows you to add many documents to an index efficiently:

```go
// Prepare documents
var documents []search.SearchDocument
for i := 0; i < 10000; i++ {
    documents = append(documents, search.SearchDocument{
        "subject":     fmt.Sprintf("doc-%d", i),
        "title":       fmt.Sprintf("Document %d", i),
        "description": fmt.Sprintf("Description for document %d", i),
        "created":     time.Now().Format(time.RFC3339),
    })
}

// Batch ingest with default options
result, err := client.BatchIngestDocuments(ctx, "index-id", documents, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Batch ingest complete:\n")
fmt.Printf("- Documents submitted: %d\n", result.TotalSubmitted)
fmt.Printf("- Documents succeeded: %d\n", result.TotalSucceeded)
fmt.Printf("- Documents failed: %d\n", result.TotalFailed)
```

### Batch Ingest Options

```go
// Batch ingest with custom options
options := &search.BatchIngestOptions{
    BatchSize:         500,        // Documents per batch (default: 100)
    Concurrency:       8,          // Concurrent batches (default: 4)
    WaitForCompletion: true,       // Wait for tasks to complete (default: true)
    PollInterval:      time.Second, // How often to check task status (default: 1s)
    MaxWaitTime:       10 * time.Minute, // Maximum wait time (default: 1h)
    ProgressCallback:  func(current, total int, done bool) {
        percent := float64(current) / float64(total) * 100
        fmt.Printf("\rProgress: %.1f%% (%d/%d)", percent, current, total)
        if done {
            fmt.Println("\nBatch ingestion complete!")
        }
    },
}

result, err := client.BatchIngestDocuments(ctx, "index-id", documents, options)
```

## Batch Deletion

Batch deletion allows you to remove many documents from an index efficiently:

```go
// Prepare document subjects to delete
var subjects []string
for i := 0; i < 5000; i++ {
    subjects = append(subjects, fmt.Sprintf("doc-%d", i))
}

// Batch delete with default options
result, err := client.BatchDeleteDocuments(ctx, "index-id", subjects, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Batch deletion complete:\n")
fmt.Printf("- Documents submitted for deletion: %d\n", result.TotalSubmitted)
fmt.Printf("- Documents successfully deleted: %d\n", result.TotalSucceeded)
fmt.Printf("- Documents failed to delete: %d\n", result.TotalFailed)
```

### Batch Delete Options

```go
// Batch delete with custom options
options := &search.BatchDeleteOptions{
    BatchSize:         500,        // Subjects per batch (default: 100)
    Concurrency:       8,          // Concurrent batches (default: 4)
    WaitForCompletion: true,       // Wait for tasks to complete (default: true)
    PollInterval:      time.Second, // How often to check task status (default: 1s)
    MaxWaitTime:       10 * time.Minute, // Maximum wait time (default: 1h)
    ProgressCallback:  func(current, total int, done bool) {
        percent := float64(current) / float64(total) * 100
        fmt.Printf("\rProgress: %.1f%% (%d/%d)", percent, current, total)
        if done {
            fmt.Println("\nBatch deletion complete!")
        }
    },
}

result, err := client.BatchDeleteDocuments(ctx, "index-id", subjects, options)
```

## Batch Operation Results

The results of batch operations contain detailed statistics:

### Batch Ingest Result

```go
type BatchIngestResult struct {
    TaskIDs         []string   // IDs of created ingest tasks
    TotalSubmitted  int        // Total documents submitted
    TotalSucceeded  int        // Documents successfully ingested
    TotalFailed     int        // Documents that failed to ingest
    FailedDocuments []FailedDocument // Details about failed documents
    ElapsedTime     time.Duration // Total elapsed time
}

type FailedDocument struct {
    Subject   string // Subject of the failed document
    BatchID   int    // Batch ID containing the document
    TaskID    string // Task ID for the batch
    ErrorCode string // Error code if available
    Message   string // Error message
}
```

### Batch Delete Result

```go
type BatchDeleteResult struct {
    TaskIDs         []string   // IDs of created delete tasks
    TotalSubmitted  int        // Total subjects submitted for deletion
    TotalSucceeded  int        // Subjects successfully deleted
    TotalFailed     int        // Subjects that failed to delete
    FailedSubjects  []FailedSubject // Details about failed subjects
    ElapsedTime     time.Duration // Total elapsed time
}

type FailedSubject struct {
    Subject   string // Subject that failed to delete
    BatchID   int    // Batch ID containing the subject
    TaskID    string // Task ID for the batch
    ErrorCode string // Error code if available
    Message   string // Error message
}
```

## Handling Failed Operations

You can examine failed operations to handle errors:

```go
// Process ingest results
result, err := client.BatchIngestDocuments(ctx, "index-id", documents, nil)
if err != nil {
    // Handle error
}

if result.TotalFailed > 0 {
    fmt.Printf("Warning: %d documents failed to ingest\n", result.TotalFailed)
    
    // Log details about failed documents
    for _, failed := range result.FailedDocuments {
        fmt.Printf("Failed document: %s (Batch: %d, Task: %s)\n", 
            failed.Subject, failed.BatchID, failed.TaskID)
        fmt.Printf("Error: %s - %s\n", failed.ErrorCode, failed.Message)
    }
    
    // Maybe retry failed documents
    var failedSubjects []string
    for _, doc := range result.FailedDocuments {
        failedSubjects = append(failedSubjects, doc.Subject)
    }
    
    // Collect documents to retry
    var retryDocuments []search.SearchDocument
    for _, doc := range documents {
        subject := doc["subject"].(string)
        for _, failedSubject := range failedSubjects {
            if subject == failedSubject {
                retryDocuments = append(retryDocuments, doc)
                break
            }
        }
    }
    
    // Retry failed documents
    if len(retryDocuments) > 0 {
        fmt.Printf("Retrying %d failed documents\n", len(retryDocuments))
        retryResult, err := client.BatchIngestDocuments(ctx, "index-id", retryDocuments, nil)
        if err != nil {
            // Handle error
        }
        
        fmt.Printf("Retry results: %d succeeded, %d failed\n", 
            retryResult.TotalSucceeded, retryResult.TotalFailed)
    }
}
```

## Optimizing Batch Operations

### Batch Size

The batch size controls how many documents are included in each request:

```go
// Optimize for smaller documents
options := &search.BatchIngestOptions{
    BatchSize:    500,  // More documents per batch for small documents
    Concurrency:  8,    // Higher concurrency
}

// Optimize for larger documents
options := &search.BatchIngestOptions{
    BatchSize:    50,   // Fewer documents per batch for large documents
    Concurrency:  4,    // Lower concurrency
}
```

### Concurrency

The concurrency controls how many batches are processed in parallel:

```go
// High concurrency for fast networks
options := &search.BatchIngestOptions{
    Concurrency: 16,  // High concurrency for fast networks
}

// Low concurrency for slower networks
options := &search.BatchIngestOptions{
    Concurrency: 2,   // Lower concurrency for slower networks
}
```

### Wait Options

You can control how long to wait for tasks to complete:

```go
// Don't wait for completion
options := &search.BatchIngestOptions{
    WaitForCompletion: false,  // Don't wait for tasks to complete
}

// Custom waiting parameters
options := &search.BatchIngestOptions{
    WaitForCompletion: true,
    PollInterval:      5 * time.Second,  // Check every 5 seconds
    MaxWaitTime:       30 * time.Minute, // Wait up to 30 minutes
}
```

## Progress Tracking

You can track progress using a callback:

```go
// Define a progress callback
progressCallback := func(current, total int, done bool) {
    percent := float64(current) / float64(total) * 100
    fmt.Printf("\rProgress: %.1f%% (%d/%d)", percent, current, total)
    if done {
        fmt.Println("\nOperation complete!")
    }
}

// Use with batch operations
options := &search.BatchIngestOptions{
    ProgressCallback: progressCallback,
}

result, err := client.BatchIngestDocuments(ctx, "index-id", documents, options)
```

## Cancellation

You can cancel batch operations using context cancellation:

```go
// Create a context with cancellation
ctx, cancel := context.WithCancel(context.Background())

// Start a goroutine to cancel after a timeout or user input
go func() {
    // Wait for user input
    fmt.Println("Press Enter to cancel the operation...")
    fmt.Scanln()
    
    // Cancel the context
    cancel()
    fmt.Println("Cancelling batch operation...")
}()

// Use the context with batch operations
result, err := client.BatchIngestDocuments(ctx, "index-id", documents, nil)
if err != nil {
    if errors.Is(err, context.Canceled) {
        fmt.Println("Batch operation was cancelled")
    } else {
        fmt.Println("Error:", err)
    }
}
```

## Common Patterns

### Index Population

Populate an index with a large dataset:

```go
// Load documents from a data source
documents, err := loadDocumentsFromSource() // Your function to load documents
if err != nil {
    // Handle error
}

// Create a progress bar
progressCallback := func(current, total int, done bool) {
    percent := float64(current) / float64(total) * 100
    fmt.Printf("\rIndexing Progress: %.1f%% (%d/%d documents)", percent, current, total)
    if done {
        fmt.Println("\nIndexing complete!")
    }
}

// Ingest documents with progress tracking
result, err := client.BatchIngestDocuments(ctx, "index-id", documents, &search.BatchIngestOptions{
    BatchSize:        200,
    Concurrency:      8,
    ProgressCallback: progressCallback,
})
if err != nil {
    // Handle error
}

fmt.Printf("\nIndexing statistics:\n")
fmt.Printf("- Documents submitted: %d\n", result.TotalSubmitted)
fmt.Printf("- Documents succeeded: %d\n", result.TotalSucceeded)
fmt.Printf("- Documents failed: %d\n", result.TotalFailed)
fmt.Printf("- Elapsed time: %s\n", result.ElapsedTime)

// Check for any failure
if result.TotalFailed > 0 {
    fmt.Printf("Warning: Some documents failed to index. Check logs for details.\n")
}
```

### Data Cleanup

Remove outdated or irrelevant documents:

```go
// Find outdated documents
query := search.NewRangeQuery("updated_date").WithLt("2020-01-01")
request := &search.StructuredSearchRequest{
    IndexID: "index-id",
    Query:   query,
    Options: &search.SearchOptions{
        Fields: []string{"subject"}, // Only retrieve subjects
        Limit:  1000,
    },
}

// Retrieve all matching subjects
var outdatedSubjects []string
allResults, err := client.StructuredSearchAll(ctx, request)
if err != nil {
    // Handle error
}

for _, result := range allResults {
    outdatedSubjects = append(outdatedSubjects, result.Subject)
}

fmt.Printf("Found %d outdated documents\n", len(outdatedSubjects))

// Delete outdated documents
if len(outdatedSubjects) > 0 {
    result, err := client.BatchDeleteDocuments(ctx, "index-id", outdatedSubjects, &search.BatchDeleteOptions{
        WaitForCompletion: true,
    })
    if err != nil {
        // Handle error
    }
    
    fmt.Printf("Cleanup complete: %d/%d documents deleted\n", 
        result.TotalSucceeded, result.TotalSubmitted)
}
```

### Index Migration

Migrate documents from one index to another:

```go
// Create a function to migrate documents between indexes
func migrateIndex(ctx context.Context, client *search.Client, sourceIndex, targetIndex string) error {
    // Retrieve all documents from source index
    request := &search.SearchRequest{
        IndexID: sourceIndex,
        Query:   "*", // Match all documents
        Options: &search.SearchOptions{
            Limit: 1000, // Max per page
        },
    }
    
    allResults, err := client.SearchAll(ctx, request)
    if err != nil {
        return fmt.Errorf("failed to retrieve source documents: %w", err)
    }
    
    fmt.Printf("Found %d documents to migrate\n", len(allResults))
    
    // Convert search results to documents
    var documents []search.SearchDocument
    for _, result := range allResults {
        // Create a new document with the same subject
        doc := search.SearchDocument{
            "subject": result.Subject,
        }
        
        // Copy all content fields
        for key, value := range result.Content {
            doc[key] = value
        }
        
        documents = append(documents, doc)
    }
    
    // Batch ingest to target index
    result, err := client.BatchIngestDocuments(ctx, targetIndex, documents, &search.BatchIngestOptions{
        BatchSize:    200,
        Concurrency:  8,
        ProgressCallback: func(current, total int, done bool) {
            percent := float64(current) / float64(total) * 100
            fmt.Printf("\rMigration Progress: %.1f%% (%d/%d)", percent, current, total)
            if done {
                fmt.Println("\nMigration complete!")
            }
        },
    })
    if err != nil {
        return fmt.Errorf("failed to ingest documents to target index: %w", err)
    }
    
    fmt.Printf("\nMigration results:\n")
    fmt.Printf("- Documents migrated: %d/%d\n", result.TotalSucceeded, result.TotalSubmitted)
    fmt.Printf("- Migration time: %s\n", result.ElapsedTime)
    
    return nil
}

// Use the migration function
err := migrateIndex(ctx, client, "old-index", "new-index")
if err != nil {
    fmt.Printf("Migration failed: %v\n", err)
}
```

## Best Practices

1. **Choose Appropriate Batch Size**: Adjust batch size based on document size (larger for small documents, smaller for large documents)
2. **Set Reasonable Concurrency**: Match concurrency to your network capabilities and API rate limits
3. **Use Progress Callbacks**: Track progress for long-running operations
4. **Handle Failures**: Check for and handle failed documents or subjects
5. **Set Timeouts**: Use context timeouts for the overall operation
6. **Cancel Gracefully**: Implement cancellation if needed for long-running operations
7. **Monitor Results**: Check statistics after completion
8. **Adjust Wait Parameters**: Set appropriate poll intervals and max wait times for your use case
9. **Retry Failed Operations**: Implement retry logic for failed documents
10. **Optimize Memory Usage**: Process very large document sets in chunks to manage memory consumption