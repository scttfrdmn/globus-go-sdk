// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenSetIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Minute),
			want:      true,
		},
		{
			name:      "valid token",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &TokenSet{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresAt:    tt.expiresAt,
			}
			if got := token.IsExpired(); got != tt.want {
				t.Errorf("TokenSet.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenSetCanRefresh(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		want         bool
	}{
		{
			name:         "can refresh",
			refreshToken: "test-refresh-token",
			want:         true,
		},
		{
			name:         "cannot refresh",
			refreshToken: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &TokenSet{
				AccessToken:  "test-access-token",
				RefreshToken: tt.refreshToken,
			}
			if got := token.CanRefresh(); got != tt.want {
				t.Errorf("TokenSet.CanRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	// Create a test entry
	entry := &Entry{
		Resource: "test-resource",
		TokenSet: &TokenSet{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "test-scope",
		},
	}

	// Test Store
	err := storage.Store(entry)
	if err != nil {
		t.Fatalf("MemoryStorage.Store() error = %v", err)
	}

	// Test Lookup
	got, err := storage.Lookup("test-resource")
	if err != nil {
		t.Fatalf("MemoryStorage.Lookup() error = %v", err)
	}
	if got == nil {
		t.Fatalf("MemoryStorage.Lookup() = nil, want entry")
	}
	if got.AccessToken != entry.TokenSet.AccessToken {
		t.Errorf("MemoryStorage.Lookup().AccessToken = %v, want %v", got.AccessToken, entry.TokenSet.AccessToken)
	}

	// Test List
	resources, err := storage.List()
	if err != nil {
		t.Fatalf("MemoryStorage.List() error = %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("MemoryStorage.List() = %v, want %v", len(resources), 1)
	}
	if resources[0] != "test-resource" {
		t.Errorf("MemoryStorage.List()[0] = %v, want %v", resources[0], "test-resource")
	}

	// Test Delete
	err = storage.Delete("test-resource")
	if err != nil {
		t.Fatalf("MemoryStorage.Delete() error = %v", err)
	}

	// Verify deletion
	got, err = storage.Lookup("test-resource")
	if err != nil {
		t.Fatalf("MemoryStorage.Lookup() error = %v", err)
	}
	if got != nil {
		t.Errorf("MemoryStorage.Lookup() = %v, want nil", got)
	}
}

func TestFileStorage(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "tokens-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}

	// Create a test entry
	entry := &Entry{
		Resource: "test-resource",
		TokenSet: &TokenSet{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scope:        "test-scope",
		},
	}

	// Test Store
	err = storage.Store(entry)
	if err != nil {
		t.Fatalf("FileStorage.Store() error = %v", err)
	}

	// Verify file was created
	_, err = os.Stat(filepath.Join(tempDir, "test-resource"))
	if os.IsNotExist(err) {
		t.Fatalf("FileStorage.Store() did not create file")
	}

	// Test Lookup
	got, err := storage.Lookup("test-resource")
	if err != nil {
		t.Fatalf("FileStorage.Lookup() error = %v", err)
	}
	if got == nil {
		t.Fatalf("FileStorage.Lookup() = nil, want entry")
	}
	if got.AccessToken != entry.TokenSet.AccessToken {
		t.Errorf("FileStorage.Lookup().AccessToken = %v, want %v", got.AccessToken, entry.TokenSet.AccessToken)
	}

	// Test List
	resources, err := storage.List()
	if err != nil {
		t.Fatalf("FileStorage.List() error = %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("FileStorage.List() = %v, want %v", len(resources), 1)
	}
	if resources[0] != "test-resource" {
		t.Errorf("FileStorage.List()[0] = %v, want %v", resources[0], "test-resource")
	}

	// Test Delete
	err = storage.Delete("test-resource")
	if err != nil {
		t.Fatalf("FileStorage.Delete() error = %v", err)
	}

	// Verify deletion
	got, err = storage.Lookup("test-resource")
	if err != nil {
		t.Fatalf("FileStorage.Lookup() error = %v", err)
	}
	if got != nil {
		t.Errorf("FileStorage.Lookup() = %v, want nil", got)
	}

	// Verify file was removed
	_, err = os.Stat(filepath.Join(tempDir, "test-resource"))
	if !os.IsNotExist(err) {
		t.Fatalf("FileStorage.Delete() did not remove file")
	}
}