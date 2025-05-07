// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CheckpointState represents the state of a resumable transfer
type CheckpointState struct {
	// CheckpointID is the unique identifier for this checkpoint
	CheckpointID string `json:"checkpoint_id"`

	// TaskInfo contains high-level information about the transfer task
	TaskInfo struct {
		SourceEndpointID      string    `json:"source_endpoint_id"`
		DestinationEndpointID string    `json:"destination_endpoint_id"`
		SourceBasePath        string    `json:"source_base_path"`
		DestinationBasePath   string    `json:"destination_base_path"`
		Label                 string    `json:"label"`
		StartTime             time.Time `json:"start_time"`
		LastUpdated           time.Time `json:"last_updated"`
	} `json:"task_info"`

	// TransferOptions contains the options for the transfer
	TransferOptions ResumableTransferOptions `json:"transfer_options"`

	// CompletedItems contains the items that have been successfully transferred
	CompletedItems []TransferItem `json:"completed_items"`

	// PendingItems contains the items that are yet to be transferred
	PendingItems []TransferItem `json:"pending_items"`

	// FailedItems contains the items that failed to transfer
	FailedItems []FailedTransferItem `json:"failed_items"`

	// CurrentTasks tracks the active task IDs
	CurrentTasks []string `json:"current_tasks"`

	// Stats contains statistics about the transfer
	Stats struct {
		TotalItems          int   `json:"total_items"`
		TotalBytes          int64 `json:"total_bytes"`
		CompletedItems      int   `json:"completed_items"`
		CompletedBytes      int64 `json:"completed_bytes"`
		FailedItems         int   `json:"failed_items"`
		AttemptedRetryItems int   `json:"attempted_retry_items"`
		RemainingItems      int   `json:"remaining_items"`
		RemainingBytes      int64 `json:"remaining_bytes"`
	} `json:"stats"`
}

// FailedTransferItem represents a transfer item that failed
type FailedTransferItem struct {
	Item         TransferItem `json:"item"`
	ErrorMessage string       `json:"error_message"`
	RetryCount   int          `json:"retry_count"`
	LastAttempt  time.Time    `json:"last_attempt"`
}

// ResumableTransferOptions contains options for resumable transfers
type ResumableTransferOptions struct {
	// BatchSize controls how many items will be included in a single transfer task
	BatchSize int `json:"batch_size"`

	// MaxRetries is the maximum number of retries for failed transfers
	MaxRetries int `json:"max_retries"`

	// RetryDelay is the delay between retries (exponential backoff will be applied)
	RetryDelay time.Duration `json:"retry_delay"`

	// CheckpointInterval is how often to save the checkpoint state
	CheckpointInterval time.Duration `json:"checkpoint_interval"`

	// SyncLevel determines how files are compared (0=none, 1=size, 2=mtime, 3=checksum)
	SyncLevel int `json:"sync_level"`

	// VerifyChecksum specifies whether to verify checksums after transfer
	VerifyChecksum bool `json:"verify_checksum"`

	// PreserveMtime specifies whether to preserve file modification times
	PreserveMtime bool `json:"preserve_mtime"`

	// Encrypt specifies whether to encrypt data in transit
	Encrypt bool `json:"encrypt"`

	// DeleteDestinationExtra specifies whether to delete files at the destination that don't exist at the source
	DeleteDestinationExtra bool `json:"delete_destination_extra"`

	// ProgressCallback is called with progress updates
	// This field is not serialized to JSON
	ProgressCallback func(state *CheckpointState) `json:"-"`
}

// DefaultResumableTransferOptions returns default options for resumable transfers
func DefaultResumableTransferOptions() *ResumableTransferOptions {
	return &ResumableTransferOptions{
		BatchSize:          100,
		MaxRetries:         3,
		RetryDelay:         time.Second * 30,
		CheckpointInterval: time.Second * 60,
		SyncLevel:          3, // Checksum
		VerifyChecksum:     true,
		PreserveMtime:      true,
		Encrypt:            true,
	}
}

// CheckpointStorage defines the interface for storing and retrieving checkpoint state
type CheckpointStorage interface {
	// SaveCheckpoint saves the checkpoint state
	SaveCheckpoint(ctx context.Context, state *CheckpointState) error

	// LoadCheckpoint loads the checkpoint state for the given ID
	LoadCheckpoint(ctx context.Context, checkpointID string) (*CheckpointState, error)

	// ListCheckpoints lists all available checkpoint IDs
	ListCheckpoints(ctx context.Context) ([]string, error)

	// DeleteCheckpoint deletes a checkpoint
	DeleteCheckpoint(ctx context.Context, checkpointID string) error
}

// FileCheckpointStorage implements CheckpointStorage using the local filesystem
type FileCheckpointStorage struct {
	// Directory where checkpoint files are stored
	Directory string

	// mutex protects access to files
	mutex sync.Mutex
}

// NewFileCheckpointStorage creates a new file-based checkpoint storage
func NewFileCheckpointStorage(directory string) (*FileCheckpointStorage, error) {
	// If directory is empty, use ~/.globus-sdk/checkpoints
	if directory == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		directory = filepath.Join(home, ".globus-sdk", "checkpoints")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	return &FileCheckpointStorage{
		Directory: directory,
	}, nil
}

// SaveCheckpoint saves the checkpoint state to a file
func (s *FileCheckpointStorage) SaveCheckpoint(ctx context.Context, state *CheckpointState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update last updated time
	state.TaskInfo.LastUpdated = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint state: %w", err)
	}

	// Write to file
	filePath := filepath.Join(s.Directory, state.CheckpointID+".json")
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// LoadCheckpoint loads a checkpoint state from a file
func (s *FileCheckpointStorage) LoadCheckpoint(ctx context.Context, checkpointID string) (*CheckpointState, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Open the file
	filePath := filepath.Join(s.Directory, checkpointID+".json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open checkpoint file: %w", err)
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	// Unmarshal the JSON
	var state CheckpointState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint state: %w", err)
	}

	return &state, nil
}

// ListCheckpoints lists all available checkpoint IDs
func (s *FileCheckpointStorage) ListCheckpoints(ctx context.Context) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// List files in the directory
	files, err := os.ReadDir(s.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoint directory: %w", err)
	}

	// Extract checkpoint IDs from filenames
	var checkpoints []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			checkpointID := file.Name()[:len(file.Name())-5] // Remove .json extension
			checkpoints = append(checkpoints, checkpointID)
		}
	}

	return checkpoints, nil
}

// DeleteCheckpoint deletes a checkpoint file
func (s *FileCheckpointStorage) DeleteCheckpoint(ctx context.Context, checkpointID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Delete the file
	filePath := filepath.Join(s.Directory, checkpointID+".json")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete checkpoint file: %w", err)
	}

	return nil
}
