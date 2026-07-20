// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package tokenstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const fileVersion = "2.0"

// jsonFileFormat is the on-disk representation.
type jsonFileFormat struct {
	Version string                           `json:"version"`
	ByRS    map[string]map[string]*TokenData `json:"by_rs"`
}

// JSONTokenStorage is a file-backed implementation of TokenStorage.
// Token data is stored as a JSON file and survives process restarts.
// Writes are atomic (write-then-rename). Safe for concurrent use.
type JSONTokenStorage struct {
	mu        sync.Mutex
	path      string
	namespace string
}

// NewJSONTokenStorage creates a JSONTokenStorage backed by the given file path,
// using the default namespace "default".
func NewJSONTokenStorage(path string) (*JSONTokenStorage, error) {
	return NewJSONTokenStorageWithNamespace(path, "default")
}

// NewJSONTokenStorageWithNamespace creates a JSONTokenStorage backed by the
// given file path. The namespace allows multiple independent token sets in
// the same file (e.g. one per application or environment).
func NewJSONTokenStorageWithNamespace(path, namespace string) (*JSONTokenStorage, error) {
	if path == "" {
		return nil, fmt.Errorf("tokenstorage: file path is required")
	}
	if namespace == "" {
		namespace = "default"
	}
	s := &JSONTokenStorage{path: path, namespace: namespace}
	// Create the file if it does not exist so we can fail early on bad paths.
	if err := s.initFile(); err != nil {
		return nil, err
	}
	return s, nil
}

// Store saves or replaces token data for the given resource server.
func (s *JSONTokenStorage) Store(data *TokenData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := s.load()
	if err != nil {
		return err
	}

	ns := s.ensureNamespace(f)
	copy := *data
	ns[data.ResourceServer] = &copy

	return s.save(f)
}

// Get retrieves token data for the given resource server.
// Returns nil, nil if not found.
func (s *JSONTokenStorage) Get(resourceServer string) (*TokenData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := s.load()
	if err != nil {
		return nil, err
	}

	ns, ok := f.ByRS[s.namespace]
	if !ok {
		return nil, nil
	}
	td, ok := ns[resourceServer]
	if !ok {
		return nil, nil
	}
	copy := *td
	return &copy, nil
}

// Remove deletes token data for the given resource server.
func (s *JSONTokenStorage) Remove(resourceServer string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := s.load()
	if err != nil {
		return err
	}

	ns, ok := f.ByRS[s.namespace]
	if !ok {
		return nil
	}
	delete(ns, resourceServer)
	return s.save(f)
}

// GetAll returns all stored token data for the current namespace.
func (s *JSONTokenStorage) GetAll() ([]*TokenData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := s.load()
	if err != nil {
		return nil, err
	}

	ns, ok := f.ByRS[s.namespace]
	if !ok {
		return []*TokenData{}, nil
	}

	result := make([]*TokenData, 0, len(ns))
	for _, td := range ns {
		copy := *td
		result = append(result, &copy)
	}
	return result, nil
}

// Close is a no-op; the file handle is not kept open between operations.
func (s *JSONTokenStorage) Close() error {
	return nil
}

// initFile creates the storage file with an empty structure if it does not exist.
func (s *JSONTokenStorage) initFile() error {
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		f := &jsonFileFormat{Version: fileVersion, ByRS: map[string]map[string]*TokenData{}}
		return s.writeFile(f)
	}
	return nil
}

// load reads and parses the storage file.
func (s *JSONTokenStorage) load() (*jsonFileFormat, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &jsonFileFormat{Version: fileVersion, ByRS: map[string]map[string]*TokenData{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("tokenstorage: read %s: %w", s.path, err)
	}

	var f jsonFileFormat
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("tokenstorage: parse %s: %w", s.path, err)
	}
	if f.ByRS == nil {
		f.ByRS = map[string]map[string]*TokenData{}
	}
	return &f, nil
}

// save atomically writes the file format to disk.
func (s *JSONTokenStorage) save(f *jsonFileFormat) error {
	return s.writeFile(f)
}

// writeFile serializes f and atomically replaces s.path.
func (s *JSONTokenStorage) writeFile(f *jsonFileFormat) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("tokenstorage: marshal: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("tokenstorage: write temp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("tokenstorage: rename to %s: %w", s.path, err)
	}
	return nil
}

// ensureNamespace returns the namespace map, creating it if absent.
func (s *JSONTokenStorage) ensureNamespace(f *jsonFileFormat) map[string]*TokenData {
	ns, ok := f.ByRS[s.namespace]
	if !ok {
		ns = map[string]*TokenData{}
		f.ByRS[s.namespace] = ns
	}
	return ns
}
