// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

const (
	// Retry configuration
	maxRetries        = 5
	initialBackoffSec = 1
	maxBackoffSec     = 60
	backoffFactor     = 2.0

	// Monitoring configuration
	progressUpdateInterval = 2 * time.Second
	metricsOutputInterval  = 10 * time.Second
)

// Environment variables for configuration
type config struct {
	// Authentication
	ClientID     string
	ClientSecret string

	// Transfer configuration
	SourceEndpointID      string
	SourcePath            string
	DestinationEndpointID string
	DestinationPath       string

	// Search configuration
	SearchIndexID string

	// Compute configuration
	ComputeEndpointID string
	ContainerImage    string

	// Flow configuration
	FlowID string
}

// Pipeline is our main application struct that orchestrates all services
type Pipeline struct {
	config config
	sdk    *pkg.SDKConfig

	// Service clients
	authClient     *auth.Client
	transferClient *transfer.Client
	searchClient   *search.Client
	computeClient  *compute.Client
	flowsClient    *flows.Client

	// Metrics collector
	metrics *Storage

	// Logger
	logger *log.Logger

	// Context for cancellation
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Initialize the pipeline with all required services
func NewPipeline() (*Pipeline, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Continue even if .env file is not found
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create logger
	logger := log.New(os.Stdout, "[PIPELINE] ", log.LstdFlags|log.Lshortfile)

	// Initialize the SDK with logging enabled
	sdk := pkg.NewConfig().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Initialize metrics storage
	metricsStorage := NewMetricsStorage()

	// Create the pipeline
	p := &Pipeline{
		config:     cfg,
		sdk:        sdk,
		logger:     logger,
		ctx:        ctx,
		cancelFunc: cancel,
		metrics:    metricsStorage,
	}

	// Initialize all required service clients
	if err := p.initializeClients(); err != nil {
		p.Shutdown()
		return nil, err
	}

	return p, nil
}

// Initialize all service clients with appropriate authorizer
func (p *Pipeline) initializeClients() error {
	// Create client credentials authorizer
	// Create a simple wrapper function for auth
	authFunction := func(ctx context.Context, clientID, clientSecret string, scopes []string) (string, time.Time, error) {
		// Return a mock token - in a real app, this would contact the auth service
		return "mock-access-token", time.Now().Add(1 * time.Hour), nil
	}

	// Create the scopes list
	scopes := []string{
		"urn:globus:auth:scope:transfer.api.globus.org:all",
		"urn:globus:auth:scope:search.api.globus.org:all",
		"urn:globus:auth:scope:flows.globus.org:all",
		"urn:globus:auth:scope:compute.api.globus.org:all",
	}

	// Create the authorizer
	clientAuthorizer := authorizers.NewClientCredentialsAuthorizer(
		p.config.ClientID,
		p.config.ClientSecret,
		scopes,
		authFunction,
	)

	// Initialize service clients
	var err error
	p.authClient, err = p.sdk.NewAuthClient()
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Get the access token from the authorizer
	token, err := clientAuthorizer.GetAuthorizationHeader(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to get authorization token: %w", err)
	}
	// Remove "Bearer " prefix if present
	accessToken := token
	if len(token) > 7 && token[:7] == "Bearer " {
		accessToken = token[7:]
	}

	p.transferClient, err = p.sdk.NewTransferClient(accessToken)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	p.searchClient, err = p.sdk.NewSearchClient(accessToken)
	if err != nil {
		return fmt.Errorf("failed to create search client: %w", err)
	}

	p.computeClient, err = p.sdk.NewComputeClient(accessToken)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	p.flowsClient, err = p.sdk.NewFlowsClient(accessToken)
	if err != nil {
		return fmt.Errorf("failed to create flows client: %w", err)
	}

	return nil
}

// Execute runs the complete data pipeline
func (p *Pipeline) Execute() error {
	p.logger.Println("Starting data pipeline execution")
	startTime := time.Now()

	// Set up metrics reporting
	go p.reportMetrics()

	// 1. Perform data transfer with resumable capability
	transferTaskID, err := p.executeTransfer()
	if err != nil {
		return fmt.Errorf("transfer stage failed: %w", err)
	}

	// 2. Index the transferred data in Search
	ingestID, err := p.indexTransferredData(transferTaskID)
	if err != nil {
		return fmt.Errorf("indexing stage failed: %w", err)
	}

	// 3. Process the data with Compute
	taskID, err := p.processData()
	if err != nil {
		return fmt.Errorf("compute processing stage failed: %w", err)
	}

	// 4. Create or run a Flow to orchestrate future executions
	flowID, runID, err := p.setupFlow(transferTaskID, ingestID, taskID)
	if err != nil {
		return fmt.Errorf("flow orchestration stage failed: %w", err)
	}

	// 5. Wait for all processes to complete
	if err := p.waitForCompletion(transferTaskID, ingestID, taskID, runID); err != nil {
		return fmt.Errorf("waiting for completion failed: %w", err)
	}

	// Record final metrics
	duration := time.Since(startTime)
	p.metrics.RecordValue("pipeline.total_duration_ms", float64(duration.Milliseconds()))
	p.metrics.RecordValue("pipeline.success", 1.0)

	p.logger.Printf("Pipeline completed successfully in %s\n", duration)
	p.logger.Printf("Flow ID: %s, Run ID: %s\n", flowID, runID)

	return nil
}

// TransferItem struct definition
type TransferItem struct {
	SourcePath      string
	DestinationPath string
	Recursive       bool
}

// TransferTask struct definition
type TransferTask struct {
	TaskID              string
	Label               string
	SourceEndpoint      string
	DestinationEndpoint string
	Sync                bool
	VerifyChecksums     bool
	Items               []TransferItem
	Status              string
}

// TaskProgress struct definition
type TaskProgress struct {
	FilesTransferred int64
	FilesTotal       int64
	BytesTransferred int64
	BytesTotal       int64
	Status           string
	StartTime        *time.Time
}

// Execute a transfer with resumable capabilities and progress monitoring
func (p *Pipeline) executeTransfer() (string, error) {
	p.logger.Println("Starting file transfer stage")

	// Create a transfer task
	task := TransferTask{
		Label:               "Data Pipeline Transfer",
		SourceEndpoint:      p.config.SourceEndpointID,
		DestinationEndpoint: p.config.DestinationEndpointID,
		Sync:                true, // Ensure destination matches source
		VerifyChecksums:     true,
	}

	// Add a transfer item
	item := TransferItem{
		SourcePath:      p.config.SourcePath,
		DestinationPath: p.config.DestinationPath,
		Recursive:       true,
	}
	task.Items = append(task.Items, item)

	// Set up progress monitoring
	progressChan := make(chan TaskProgress)
	progressCtx, progressCancel := context.WithCancel(p.ctx)
	defer progressCancel()

	// Start progress monitoring in a goroutine
	go func() {
		ticker := time.NewTicker(progressUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-progressCtx.Done():
				return
			case <-ticker.C:
				// Mock getting task progress
				startTime := time.Now().Add(-5 * time.Minute)
				progress := TaskProgress{
					FilesTransferred: 1,
					FilesTotal:       2,
					BytesTransferred: 1024,
					BytesTotal:       2048,
					Status:           "ACTIVE",
					StartTime:        &startTime,
				}
				progressChan <- progress
			}
		}
	}()

	// Mock submission of transfer (in a real app this would call the API)
	taskID := "mock-task-" + time.Now().Format(time.RFC3339)
	task.TaskID = taskID
	task.Status = "ACTIVE"

	p.logger.Printf("Submitted transfer task with ID: %s", taskID)

	// Mock checkpoint manager for resumable transfers
	type CheckpointManager struct {
		TaskID         string
		SaveCheckpoint func(ctx context.Context) error
	}

	checkpointManager := &CheckpointManager{TaskID: taskID}

	// Add a SaveCheckpoint method to the CheckpointManager
	checkpointManager.SaveCheckpoint = func(ctx context.Context) error {
		p.logger.Printf("Saving checkpoint for task %s", taskID)
		return nil
	}

	// Handle interrupts for resumable transfers
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-interruptChan
		p.logger.Println("Received interrupt, saving checkpoint and exiting...")
		if err := checkpointManager.SaveCheckpoint(p.ctx); err != nil {
			p.logger.Printf("Error saving checkpoint: %v", err)
		}
		p.Shutdown()
		os.Exit(1)
	}()

	// Start monitoring progress
	go func() {
		for progress := range progressChan {
			p.logger.Printf("Transfer Progress: %d/%d files, %d/%d bytes, %s status",
				progress.FilesTransferred, progress.FilesTotal,
				progress.BytesTransferred, progress.BytesTotal,
				progress.Status)

			// Record metrics
			p.metrics.RecordValue("transfer.files_transferred", float64(progress.FilesTransferred))
			p.metrics.RecordValue("transfer.bytes_transferred", float64(progress.BytesTransferred))

			// Calculate and record transfer rate
			if progress.BytesTransferred > 0 && progress.StartTime != nil {
				elapsed := time.Since(*progress.StartTime).Seconds()
				if elapsed > 0 {
					transferRate := float64(progress.BytesTransferred) / elapsed
					p.metrics.RecordValue("transfer.bytes_per_second", transferRate)
				}
			}
		}
	}()

	// Wait for the transfer to complete
	result, err := p.waitForTaskWithRetry(taskID)
	if err != nil {
		// If transfer fails, save checkpoint for later resumption
		if err := checkpointManager.SaveCheckpoint(p.ctx); err != nil {
			p.logger.Printf("Error saving checkpoint: %v", err)
		}
		return "", fmt.Errorf("transfer failed: %w", err)
	}

	// Close the progress channel
	progressCancel()

	p.logger.Printf("Transfer completed with %d successful files, %d skipped files, %d failed files",
		result.Successful, result.Skipped, result.Failed)

	// Record final transfer metrics
	p.metrics.RecordValue("transfer.successful_files", float64(result.Successful))
	p.metrics.RecordValue("transfer.failed_files", float64(result.Failed))

	return taskID, nil
}

// SearchEntry - custom definition to avoid search package dependency
type SearchEntry struct {
	ID         string
	Content    map[string]interface{}
	Visible_to []string
}

// SearchBatch - custom definition to avoid search package dependency
type SearchBatch struct {
	Entries []SearchEntry
}

// Index the transferred data in the Search service
func (p *Pipeline) indexTransferredData(transferTaskID string) (string, error) {
	p.logger.Println("Starting data indexing stage")

	// Mock task details
	task := TransferTask{
		TaskID: transferTaskID,
		Items: []TransferItem{
			{
				SourcePath:      p.config.SourcePath,
				DestinationPath: p.config.DestinationPath,
				Recursive:       true,
			},
		},
	}

	// Prepare entries for indexing
	entries := []SearchEntry{}
	for _, item := range task.Items {
		// Create a unique ID for this item
		entryID := fmt.Sprintf("%s:%s", filepath.Base(item.DestinationPath), time.Now().Format(time.RFC3339))

		// Create entry with metadata
		entry := SearchEntry{
			ID: entryID,
			Content: map[string]interface{}{
				"path":          item.DestinationPath,
				"source_path":   item.SourcePath,
				"endpoint_id":   p.config.DestinationEndpointID,
				"transfer_id":   transferTaskID,
				"transfer_time": time.Now().Format(time.RFC3339),
				"file_name":     filepath.Base(item.DestinationPath),
				"directory":     filepath.Dir(item.DestinationPath),
				"pipeline_run":  true,
			},
			Visible_to: []string{"public"},
		}
		entries = append(entries, entry)
	}

	// Create batch for ingest (in a real implementation we would use this batch)
	_ = SearchBatch{
		Entries: entries,
	}

	// Submit the ingest - mock response
	ingestID := "ingest-" + time.Now().Format(time.RFC3339)
	p.logger.Printf("Indexed %d entries with ingest ID: %s", len(entries), ingestID)

	// Record metrics
	p.metrics.RecordValue("search.entries_indexed", float64(len(entries)))

	return ingestID, nil
}

// Process the transferred data using Compute
func (p *Pipeline) processData() (string, error) {
	p.logger.Println("Starting data processing stage")

	// Create a container function - mock response
	funcID := "func-" + time.Now().Format(time.RFC3339)
	p.logger.Printf("Registered compute function with ID: %s", funcID)

	// Submit the task - mock response
	taskID := "task-" + time.Now().Format(time.RFC3339)
	p.logger.Printf("Submitted compute task with ID: %s", taskID)

	// Record metrics
	p.metrics.RecordValue("compute.tasks_submitted", 1.0)

	return taskID, nil
}

// Set up a Flow to orchestrate the pipeline
func (p *Pipeline) setupFlow(transferTaskID, ingestID, computeTaskID string) (string, string, error) {
	p.logger.Println("Starting flow orchestration stage")

	var flowID string

	// Use existing flow or create a new one
	if p.config.FlowID != "" {
		flowID = p.config.FlowID
		p.logger.Printf("Using existing flow with ID: %s", flowID)
	} else {
		// Mock flow creation
		flowID = "flow-" + time.Now().Format(time.RFC3339)
		p.logger.Printf("Created new flow with ID: %s", flowID)
	}

	// Mock run creation
	runID := "run-" + time.Now().Format(time.RFC3339)
	p.logger.Printf("Started flow run with ID: %s", runID)

	// Record metrics
	p.metrics.RecordValue("flows.runs_started", 1.0)

	return flowID, runID, nil
}

// TaskCompletionResult represents the result of a completed transfer task
type TaskCompletionResult struct {
	Successful int
	Failed     int
	Skipped    int
}

// Wait for all pipeline components to complete
func (p *Pipeline) waitForCompletion(transferTaskID, ingestID, computeTaskID, flowRunID string) error {
	p.logger.Println("Waiting for all pipeline components to complete")

	// Wait for a short time to simulate work
	time.Sleep(2 * time.Second)

	p.logger.Println("All pipeline components completed successfully")
	return nil
}

// Wait for a task to complete with retry capability
func (p *Pipeline) waitForTaskWithRetry(taskID string) (*TaskCompletionResult, error) {
	backoff := initialBackoffSec

	for attempts := 0; attempts < maxRetries; attempts++ {
		// This is a mock implementation - simulate completion after first attempt
		if attempts == 0 {
			// Simulate work
			time.Sleep(1 * time.Second)

			// Return successful result
			return &TaskCompletionResult{
				Successful: 1,
				Failed:     0,
				Skipped:    0,
			}, nil
		}

		// Log retry attempt
		p.logger.Printf("Retry %d/%d: Task completion check failed. Retrying in %d seconds...",
			attempts+1, maxRetries, backoff)

		// Wait before retrying
		select {
		case <-time.After(time.Duration(backoff) * time.Second):
			// Exponential backoff with type conversion
			backoff = int(min(float64(backoff)*backoffFactor, float64(maxBackoffSec)))
		case <-p.ctx.Done():
			// Get the context error and return it
			ctxErr := p.ctx.Err()
			return nil, fmt.Errorf("context canceled while waiting for task: %w", ctxErr)
		}
	}

	// Maximum retries reached
	return nil, errors.New("maximum retry attempts reached")
}

// Report metrics periodically
func (p *Pipeline) reportMetrics() {
	ticker := time.NewTicker(metricsOutputInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Get all metrics
			allMetrics := p.metrics.GetAllMetrics()

			if len(allMetrics) > 0 {
				p.logger.Println("Current Pipeline Metrics:")
				for key, value := range allMetrics {
					p.logger.Printf("  %s: %.2f", key, value)
				}
			}
		}
	}
}

// Shutdown gracefully closes the pipeline
func (p *Pipeline) Shutdown() {
	p.logger.Println("Shutting down pipeline")
	p.cancelFunc()
}

// Helper function to check if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for specific retryable errors
	knownRetryableErrors := []string{
		"connection reset",
		"timeout",
		"temporary",
		"deadline exceeded",
		"too many requests",
		"rate limit",
		"503",
		"429",
	}

	errStr := strings.ToLower(err.Error())
	for _, retryable := range knownRetryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// Helper function to load configuration from environment variables
func loadConfig() (config, error) {
	cfg := config{
		// Authentication
		ClientID:     os.Getenv("GLOBUS_CLIENT_ID"),
		ClientSecret: os.Getenv("GLOBUS_CLIENT_SECRET"),

		// Transfer configuration
		SourceEndpointID:      os.Getenv("SOURCE_ENDPOINT_ID"),
		SourcePath:            os.Getenv("SOURCE_PATH"),
		DestinationEndpointID: os.Getenv("DESTINATION_ENDPOINT_ID"),
		DestinationPath:       os.Getenv("DESTINATION_PATH"),

		// Search configuration
		SearchIndexID: os.Getenv("SEARCH_INDEX_ID"),

		// Compute configuration
		ComputeEndpointID: os.Getenv("COMPUTE_ENDPOINT_ID"),
		ContainerImage:    os.Getenv("CONTAINER_IMAGE"),

		// Flow configuration (optional)
		FlowID: os.Getenv("FLOW_ID"),
	}

	// Validate required configuration
	var missingVars []string

	if cfg.ClientID == "" {
		missingVars = append(missingVars, "GLOBUS_CLIENT_ID")
	}
	if cfg.ClientSecret == "" {
		missingVars = append(missingVars, "GLOBUS_CLIENT_SECRET")
	}
	if cfg.SourceEndpointID == "" {
		missingVars = append(missingVars, "SOURCE_ENDPOINT_ID")
	}
	if cfg.SourcePath == "" {
		missingVars = append(missingVars, "SOURCE_PATH")
	}
	if cfg.DestinationEndpointID == "" {
		missingVars = append(missingVars, "DESTINATION_ENDPOINT_ID")
	}
	if cfg.DestinationPath == "" {
		missingVars = append(missingVars, "DESTINATION_PATH")
	}
	if cfg.SearchIndexID == "" {
		missingVars = append(missingVars, "SEARCH_INDEX_ID")
	}
	if cfg.ComputeEndpointID == "" {
		missingVars = append(missingVars, "COMPUTE_ENDPOINT_ID")
	}
	if cfg.ContainerImage == "" {
		missingVars = append(missingVars, "CONTAINER_IMAGE")
	}

	if len(missingVars) > 0 {
		return cfg, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return cfg, nil
}

// min returns the smaller of x or y
func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

// Storage is a simple metrics storage implementation
type Storage struct {
	values map[string]float64
	mu     sync.RWMutex
}

// NewMetricsStorage creates a new metrics storage instance
func NewMetricsStorage() *Storage {
	return &Storage{
		values: make(map[string]float64),
	}
}

// RecordValue records a metric value
func (s *Storage) RecordValue(key string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

// GetValue gets a metric value
func (s *Storage) GetValue(key string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.values[key]
	return value, exists
}

// GetAllMetrics returns all metrics
func (s *Storage) GetAllMetrics() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Make a copy to avoid concurrent map access
	result := make(map[string]float64, len(s.values))
	for k, v := range s.values {
		result[k] = v
	}
	return result
}

func main() {
	// Create the pipeline
	pipeline, err := NewPipeline()
	if err != nil {
		log.Fatalf("Failed to initialize pipeline: %v", err)
	}
	defer pipeline.Shutdown()

	// Execute the pipeline
	if err := pipeline.Execute(); err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	log.Println("Data pipeline completed successfully")
}
