// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TokenSet represents a set of OAuth2 tokens
type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
}

// IsExpired returns true if the token is expired
func (t *TokenSet) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CanRefresh returns true if the token can be refreshed
func (t *TokenSet) CanRefresh() bool {
	return t.RefreshToken != ""
}

// Entry represents a token entry in the storage
type Entry struct {
	Resource     string    `json:"resource"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	TokenSet     *TokenSet `json:"-"` // Used for convenience, not serialized
}

// Storage defines the interface for storing and retrieving tokens
type Storage interface {
	// Store saves a token entry
	Store(entry *Entry) error

	// Lookup retrieves a token entry for a specific resource
	Lookup(resource string) (*Entry, error)

	// Delete removes a token entry for a specific resource
	Delete(resource string) error

	// List returns all stored token resources
	List() ([]string, error)
}

// MemoryStorage implements in-memory token storage
type MemoryStorage struct {
	entries map[string]*Entry
	mutex   sync.RWMutex
}

// NewMemoryStorage creates a new in-memory token storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		entries: make(map[string]*Entry),
	}
}

// Store implements Storage.Store for in-memory storage
func (s *MemoryStorage) Store(entry *Entry) error {
	if entry == nil || entry.Resource == "" {
		return errors.New("invalid token entry")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Copy the entry
	newEntry := &Entry{
		Resource:     entry.Resource,
		AccessToken:  entry.AccessToken,
		RefreshToken: entry.RefreshToken,
		ExpiresAt:    entry.ExpiresAt,
		Scope:        entry.Scope,
	}

	// If TokenSet is available, use its values
	if entry.TokenSet != nil {
		newEntry.AccessToken = entry.TokenSet.AccessToken
		newEntry.RefreshToken = entry.TokenSet.RefreshToken
		newEntry.ExpiresAt = entry.TokenSet.ExpiresAt
		newEntry.Scope = entry.TokenSet.Scope
	}

	// Create a TokenSet for the new entry
	newEntry.TokenSet = &TokenSet{
		AccessToken:  newEntry.AccessToken,
		RefreshToken: newEntry.RefreshToken,
		ExpiresAt:    newEntry.ExpiresAt,
		Scope:        newEntry.Scope,
		ResourceID:   newEntry.Resource,
	}

	s.entries[entry.Resource] = newEntry

	return nil
}

// Lookup implements Storage.Lookup for in-memory storage
func (s *MemoryStorage) Lookup(resource string) (*Entry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, ok := s.entries[resource]
	if !ok {
		return nil, nil // Not found, but not an error
	}

	// Return a copy
	newEntry := &Entry{
		Resource:     entry.Resource,
		AccessToken:  entry.AccessToken,
		RefreshToken: entry.RefreshToken,
		ExpiresAt:    entry.ExpiresAt,
		Scope:        entry.Scope,
	}

	// Create a TokenSet for the new entry
	newEntry.TokenSet = &TokenSet{
		AccessToken:  newEntry.AccessToken,
		RefreshToken: newEntry.RefreshToken,
		ExpiresAt:    newEntry.ExpiresAt,
		Scope:        newEntry.Scope,
		ResourceID:   newEntry.Resource,
	}

	return newEntry, nil
}

// Delete implements Storage.Delete for in-memory storage
func (s *MemoryStorage) Delete(resource string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.entries, resource)
	return nil
}

// List implements Storage.List for in-memory storage
func (s *MemoryStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	resources := make([]string, 0, len(s.entries))
	for resource := range s.entries {
		resources = append(resources, resource)
	}

	return resources, nil
}

// FileStorage implements file-based token storage
type FileStorage struct {
	directory string
	mutex     sync.RWMutex
}

// NewFileStorage creates a new file-based token storage
func NewFileStorage(directory string) (*FileStorage, error) {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		directory: directory,
	}, nil
}

// Store implements Storage.Store for file-based storage
func (s *FileStorage) Store(entry *Entry) error {
	if entry == nil || entry.Resource == "" {
		return errors.New("invalid token entry")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If the entry has a TokenSet, copy its values to the entry fields
	if entry.TokenSet != nil {
		entry.AccessToken = entry.TokenSet.AccessToken
		entry.RefreshToken = entry.TokenSet.RefreshToken
		entry.ExpiresAt = entry.TokenSet.ExpiresAt
		entry.Scope = entry.TokenSet.Scope
	}

	// Marshal the entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal token entry: %w", err)
	}

	// Write to file
	filename := filepath.Join(s.directory, sanitizeFilename(entry.Resource))
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// Lookup implements Storage.Lookup for file-based storage
func (s *FileStorage) Lookup(resource string) (*Entry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Read the file
	filename := filepath.Join(s.directory, sanitizeFilename(resource))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Unmarshal the entry
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token entry: %w", err)
	}

	// Create a TokenSet for convenience
	entry.TokenSet = &TokenSet{
		AccessToken:  entry.AccessToken,
		RefreshToken: entry.RefreshToken,
		ExpiresAt:    entry.ExpiresAt,
		Scope:        entry.Scope,
		ResourceID:   entry.Resource,
	}

	return &entry, nil
}

// Delete implements Storage.Delete for file-based storage
func (s *FileStorage) Delete(resource string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filename := filepath.Join(s.directory, sanitizeFilename(resource))
	err := os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	return nil
}

// List implements Storage.List for file-based storage
func (s *FileStorage) List() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	resources := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			resources = append(resources, entry.Name())
		}
	}

	return resources, nil
}

// sanitizeFilename sanitizes a resource ID for use as a filename
func sanitizeFilename(resource string) string {
	// This is a simple implementation - in a real application,
	// you might want to use a more sophisticated approach
	// such as base64 encoding or a hash function
	return resource
}
