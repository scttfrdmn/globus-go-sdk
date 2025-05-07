// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

// SyncLevelProvider is an interface that ensures all required sync level constants are defined
// By having this interface, we can use it in tests to verify that all required constants exist
type SyncLevelProvider interface {
	// Methods that return the sync level constants
	GetSyncLevelExists() int
	GetSyncLevelSize() int
	GetSyncLevelModified() int
	GetSyncLevelChecksum() int
	GetSyncChecksum() int
}

// syncLevelProviderImpl is a concrete implementation of SyncLevelProvider
// that uses the package constants
type syncLevelProviderImpl struct{}

func (s *syncLevelProviderImpl) GetSyncLevelExists() int   { return SyncLevelExists }
func (s *syncLevelProviderImpl) GetSyncLevelSize() int     { return SyncLevelSize }
func (s *syncLevelProviderImpl) GetSyncLevelModified() int { return SyncLevelModified }
func (s *syncLevelProviderImpl) GetSyncLevelChecksum() int { return SyncLevelChecksum }
func (s *syncLevelProviderImpl) GetSyncChecksum() int      { return SyncChecksum }

// Verify at compile time that syncLevelProviderImpl implements SyncLevelProvider
var _ SyncLevelProvider = &syncLevelProviderImpl{}