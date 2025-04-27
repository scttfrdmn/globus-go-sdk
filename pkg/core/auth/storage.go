// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrTokenNotFound is returned when a token is not found in storage
	ErrTokenNotFound = errors.New("token not found")

	// ErrStorageCorrupt is returned when token storage is corrupt
	ErrStorageCorrupt = errors.New("token storage is corrupt")
)

// TokenInfo stores information about an OAuth2 token
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
}

// IsValid returns true if the token is still valid (not expired)
func (t TokenInfo) IsValid() bool {
	// Add a buffer of 30 seconds to avoid edge cases
	return time.Now().Add(30 * time.Second).Before(t.ExpiresAt)
}

// CanRefresh returns true if the token can be refreshed
func (t TokenInfo) CanRefresh() bool {
	return t.RefreshToken != ""
}

// TokenStorage defines the interface for storing and retrieving tokens
type TokenStorage interface {
	// StoreToken saves a token for a specific user or resource
	StoreToken(ctx context.Context, key string, token TokenInfo) error

	// GetToken retrieves a token for a specific user or resource
	GetToken(ctx context.Context, key string) (TokenInfo, error)

	// DeleteToken removes a token for a specific user or resource
	DeleteToken(ctx context.Context, key string) error

	// ListTokens returns all stored token keys
	ListTokens(ctx context.Context) ([]string, error)
}

// MemoryTokenStorage is an in-memory implementation of TokenStorage
type MemoryTokenStorage struct {
	tokens map[string]TokenInfo
	mu     sync.RWMutex
}

// NewMemoryTokenStorage creates a new in-memory token storage
func NewMemoryTokenStorage() *MemoryTokenStorage {
	return &MemoryTokenStorage{
		tokens: make(map[string]TokenInfo),
	}
}

// StoreToken saves a token to memory
func (m *MemoryTokenStorage) StoreToken(ctx context.Context, key string, token TokenInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[key] = token
	return nil
}

// GetToken retrieves a token from memory
func (m *MemoryTokenStorage) GetToken(ctx context.Context, key string) (TokenInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	token, ok := m.tokens[key]
	if !ok {
		return TokenInfo{}, ErrTokenNotFound
	}
	return token, nil
}

// DeleteToken removes a token from memory
func (m *MemoryTokenStorage) DeleteToken(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, key)
	return nil
}

// ListTokens returns all token keys from memory
func (m *MemoryTokenStorage) ListTokens(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.tokens))
	for k := range m.tokens {
		keys = append(keys, k)
	}
	return keys, nil
}

// FileTokenStorage is a file-based implementation of TokenStorage
type FileTokenStorage struct {
	directory string
	mu        sync.Mutex
}

// NewFileTokenStorage creates a new file-based token storage
func NewFileTokenStorage(directory string) (*FileTokenStorage, error) {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("failed to create token directory: %w", err)
	}

	return &FileTokenStorage{
		directory: directory,
	}, nil
}

// getFilePath returns the file path for a token key
func (f *FileTokenStorage) getFilePath(key string) string {
	// Ensure the key is safe for use as a filename
	safeKey := filepath.Base(fmt.Sprintf("%s.json", key))
	return filepath.Join(f.directory, safeKey)
}

// StoreToken saves a token to a file
func (f *FileTokenStorage) StoreToken(ctx context.Context, key string, token TokenInfo) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Marshal the token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write the token to a file
	filePath := f.getFilePath(key)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// GetToken retrieves a token from a file
func (f *FileTokenStorage) GetToken(ctx context.Context, key string) (TokenInfo, error) {
	filePath := f.getFilePath(key)

	// Read the token file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return TokenInfo{}, ErrTokenNotFound
		}
		return TokenInfo{}, fmt.Errorf("failed to read token file: %w", err)
	}

	// Unmarshal the token
	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return TokenInfo{}, fmt.Errorf("%w: %v", ErrStorageCorrupt, err)
	}

	return token, nil
}

// DeleteToken removes a token file
func (f *FileTokenStorage) DeleteToken(ctx context.Context, key string) error {
	filePath := f.getFilePath(key)

	// Remove the token file
	err := os.Remove(filePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	return nil
}

// ListTokens returns all token keys from the directory
func (f *FileTokenStorage) ListTokens(ctx context.Context) ([]string, error) {
	// Read the directory
	files, err := os.ReadDir(f.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read token directory: %w", err)
	}

	// Extract the token keys from filenames
	keys := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		// Remove the .json extension to get the key
		key := file.Name()[:len(file.Name())-5]
		keys = append(keys, key)
	}

	return keys, nil
}
