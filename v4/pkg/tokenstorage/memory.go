// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokenstorage

import "sync"

// MemoryTokenStorage is an in-memory implementation of TokenStorage.
// Token data is lost when the process exits. Safe for concurrent use.
type MemoryTokenStorage struct {
	mu   sync.RWMutex
	data map[string]*TokenData
}

// NewMemoryTokenStorage creates a new empty MemoryTokenStorage.
func NewMemoryTokenStorage() *MemoryTokenStorage {
	return &MemoryTokenStorage{
		data: make(map[string]*TokenData),
	}
}

// Store saves or replaces token data for the given resource server.
func (m *MemoryTokenStorage) Store(data *TokenData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Copy to avoid external mutation.
	copy := *data
	m.data[data.ResourceServer] = &copy
	return nil
}

// Get retrieves token data for the given resource server.
// Returns nil, nil if not found.
func (m *MemoryTokenStorage) Get(resourceServer string) (*TokenData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	td, ok := m.data[resourceServer]
	if !ok {
		return nil, nil
	}
	copy := *td
	return &copy, nil
}

// Remove deletes token data for the given resource server.
func (m *MemoryTokenStorage) Remove(resourceServer string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, resourceServer)
	return nil
}

// GetAll returns all stored token data.
func (m *MemoryTokenStorage) GetAll() ([]*TokenData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*TokenData, 0, len(m.data))
	for _, td := range m.data {
		copy := *td
		result = append(result, &copy)
	}
	return result, nil
}

// Close is a no-op for in-memory storage.
func (m *MemoryTokenStorage) Close() error {
	return nil
}
