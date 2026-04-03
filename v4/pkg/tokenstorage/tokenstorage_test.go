// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package tokenstorage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleToken(rs string) *tokenstorage.TokenData {
	return &tokenstorage.TokenData{
		ResourceServer: rs,
		AccessToken:    "at-" + rs,
		RefreshToken:   "rt-" + rs,
		Scope:          "all",
		TokenType:      "Bearer",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
}

func runStorageTests(t *testing.T, s tokenstorage.TokenStorage) {
	t.Helper()

	t.Run("store and get", func(t *testing.T) {
		td := sampleToken("transfer.api.globus.org")
		require.NoError(t, s.Store(td))

		got, err := s.Get("transfer.api.globus.org")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "at-transfer.api.globus.org", got.AccessToken)
	})

	t.Run("get missing returns nil", func(t *testing.T) {
		got, err := s.Get("missing.example.com")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("get all", func(t *testing.T) {
		_ = s.Store(sampleToken("groups.api.globus.org"))
		all, err := s.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)
	})

	t.Run("remove", func(t *testing.T) {
		_ = s.Store(sampleToken("search.api.globus.org"))
		require.NoError(t, s.Remove("search.api.globus.org"))
		got, err := s.Get("search.api.globus.org")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("remove missing is no-op", func(t *testing.T) {
		assert.NoError(t, s.Remove("nonexistent.example.com"))
	})

	t.Run("close", func(t *testing.T) {
		assert.NoError(t, s.Close())
	})
}

func TestMemoryTokenStorage(t *testing.T) {
	s := tokenstorage.NewMemoryTokenStorage()
	runStorageTests(t, s)
}

func TestJSONTokenStorage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	s, err := tokenstorage.NewJSONTokenStorage(path)
	require.NoError(t, err)
	runStorageTests(t, s)
}

func TestJSONTokenStoragePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	// Write in one instance.
	s1, err := tokenstorage.NewJSONTokenStorage(path)
	require.NoError(t, err)
	require.NoError(t, s1.Store(sampleToken("transfer.api.globus.org")))

	// Read in a new instance.
	s2, err := tokenstorage.NewJSONTokenStorage(path)
	require.NoError(t, err)
	got, err := s2.Get("transfer.api.globus.org")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "at-transfer.api.globus.org", got.AccessToken)
}

func TestJSONTokenStorageNamespaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	sA, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, "app-a")
	require.NoError(t, err)
	sB, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, "app-b")
	require.NoError(t, err)

	require.NoError(t, sA.Store(sampleToken("transfer.api.globus.org")))

	// app-b should not see app-a's tokens.
	got, err := sB.Get("transfer.api.globus.org")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestJSONTokenStorageAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	s, err := tokenstorage.NewJSONTokenStorage(path)
	require.NoError(t, err)
	require.NoError(t, s.Store(sampleToken("transfer.api.globus.org")))

	// Temp file should be gone after a successful write.
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestTokenDataHelpers(t *testing.T) {
	future := &tokenstorage.TokenData{ExpiresAt: time.Now().Add(time.Hour)}
	assert.False(t, future.IsExpired())
	assert.Greater(t, future.ExpiresIn(), time.Duration(0))

	past := &tokenstorage.TokenData{ExpiresAt: time.Now().Add(-time.Minute)}
	assert.True(t, past.IsExpired())
	assert.Equal(t, time.Duration(0), past.ExpiresIn())
}
