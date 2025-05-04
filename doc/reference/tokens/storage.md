# Tokens Package: Storage

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The tokens package provides a storage interface for persisting OAuth2 tokens, with implementations for memory-based and file-based storage.

## Storage Interface

```go
type Storage interface {
    // Store stores a token entry
    Store(entry *Entry) error
    
    // Lookup retrieves a token entry by resource ID
    Lookup(resource string) (*Entry, error)
    
    // List returns all stored resource IDs
    List() ([]string, error)
    
    // Delete removes a token entry by resource ID
    Delete(resource string) error
}
```

## Memory Storage

Memory storage keeps tokens in memory, which is useful for short-lived applications or testing.

### Creating Memory Storage

```go
// Create memory storage
storage := tokens.NewMemoryStorage()
```

### Characteristics

- **Persistence**: Tokens are lost when the application terminates
- **Concurrency**: Thread-safe for concurrent access
- **Performance**: Fast, with O(1) access time
- **Use Cases**: Testing, short-lived applications, or when tokens are only needed for the current session

## File Storage

File storage persists tokens to the filesystem, which is useful for long-lived applications or CLIs.

### Creating File Storage

```go
// Create file storage
storage, err := tokens.NewFileStorage("~/.globus-tokens")
if err != nil {
    // Handle error
}
```

The path can be:
- An absolute path (e.g., `/home/user/.globus-tokens`)
- A path with `~` expansion (e.g., `~/.globus-tokens`)
- A relative path (e.g., `.globus-tokens`)

### Characteristics

- **Persistence**: Tokens are stored on disk and persist across application restarts
- **Concurrency**: Thread-safe for concurrent access
- **Security**: Tokens are stored in JSON format; ensure the directory has appropriate permissions
- **Use Cases**: CLI tools, daemon applications, or any application that needs tokens to persist across restarts

## Token Entry

The `Entry` struct represents a token entry in storage:

```go
type Entry struct {
    Resource     string
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scope        string
    TokenSet     *TokenSet
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Resource` | `string` | A unique identifier for the token |
| `AccessToken` | `string` | The OAuth2 access token |
| `RefreshToken` | `string` | The OAuth2 refresh token |
| `ExpiresAt` | `time.Time` | When the access token expires |
| `Scope` | `string` | The scope of the token |
| `TokenSet` | `*TokenSet` | A convenience wrapper for the token |

Note: The `TokenSet` field is a convenience wrapper that will be populated automatically if it's not set.

## Using Storage Directly

While it's generally recommended to use the Token Manager for handling tokens, you can also use the storage implementations directly:

### Storing a Token

```go
// Create a token entry
entry := &tokens.Entry{
    Resource:     "example-resource",
    AccessToken:  "access-token",
    RefreshToken: "refresh-token",
    ExpiresAt:    time.Now().Add(1 * time.Hour),
    Scope:        "example-scope",
}

// Store the token
err := storage.Store(entry)
if err != nil {
    // Handle error
}
```

### Looking Up a Token

```go
// Look up a token
entry, err := storage.Lookup("example-resource")
if err != nil {
    // Handle error
}
if entry == nil {
    // Token not found
} else {
    // Use entry.AccessToken
}
```

### Listing All Resources

```go
// List all resources
resources, err := storage.List()
if err != nil {
    // Handle error
}
for _, resource := range resources {
    fmt.Println(resource)
}
```

### Deleting a Token

```go
// Delete a token
err := storage.Delete("example-resource")
if err != nil {
    // Handle error
}
```

## Best Practices

1. **Use file storage for CLI tools**: CLI tools should use file storage to persist tokens across invocations
2. **Use memory storage for tests**: Memory storage is ideal for unit tests
3. **Consider security**: File storage stores tokens on disk; ensure the directory has appropriate permissions
4. **Handle errors**: Always check for errors when performing storage operations
5. **Use the Token Manager**: Instead of using storage directly, use the Token Manager for a higher-level API

## Error Handling

Storage implementations can return the following errors:

### Memory Storage

- No specific errors; memory storage operations always succeed

### File Storage

- `"failed to create directory: %w"`: When creating the storage directory fails
- `"failed to marshal token: %w"`: When converting a token to JSON fails
- `"failed to write token: %w"`: When writing a token to disk fails
- `"failed to list files: %w"`: When listing token files fails
- `"failed to unmarshal token: %w"`: When converting JSON to a token fails
- `"failed to read token: %w"`: When reading a token file fails
- `"failed to remove token file: %w"`: When deleting a token file fails

## Implementation Details

### Memory Storage

```go
type MemoryStorage struct {
    tokens map[string]*Entry
    mu     sync.RWMutex
}
```

Memory storage uses a mutex-protected map to store tokens in memory.

### File Storage

```go
type FileStorage struct {
    dir string
    mu  sync.RWMutex
}
```

File storage uses a directory on disk to store tokens, with each token stored in a separate file named after its resource ID. A mutex is used to synchronize file operations.

## Extending Storage

You can create your own storage implementation by implementing the `Storage` interface:

```go
type MyStorage struct {
    // Your fields here
}

func (s *MyStorage) Store(entry *Entry) error {
    // Your implementation here
}

func (s *MyStorage) Lookup(resource string) (*Entry, error) {
    // Your implementation here
}

func (s *MyStorage) List() ([]string, error) {
    // Your implementation here
}

func (s *MyStorage) Delete(resource string) error {
    // Your implementation here
}
```

Common storage backends you might want to implement:

- Database storage (e.g., SQL, NoSQL)
- Encrypted file storage
- Remote storage (e.g., Redis, etcd)