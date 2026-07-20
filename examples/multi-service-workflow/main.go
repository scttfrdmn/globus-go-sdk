// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

// WorkflowConfig holds configuration for the workflow
type WorkflowConfig struct {
	// Endpoint IDs
	SourceEndpointID      string
	DestinationEndpointID string
	ComputeEndpointID     string
	SearchIndexID         string

	// Paths
	SourcePath      string
	DestinationPath string
	ResultsPath     string

	// Compute
	ComputeFunction string
	ComputeArgs     []string

	// Flow
	FlowName        string
	FlowDescription string

	// Workflow options
	SkipTransfer  bool
	SkipCompute   bool
	SkipSearch    bool
	SkipFlows     bool
	WithDashboard bool
}

// Workflow manages the multi-service workflow
type Workflow struct {
	Config WorkflowConfig
	SDK    *pkg.SDKConfig
	Logger *log.Logger

	// Component states
	TransferTaskID    string
	ComputeTaskID     string
	SearchIndexCount  int
	FlowID            string
	TransferCompleted bool
	ComputeCompleted  bool
	SearchCompleted   bool
	FlowCreated       bool

	// Metrics
	StartTime        time.Time
	EndTime          time.Time
	BytesTransferred int64
	ComputeTime      time.Duration
	ErrorCount       int

	// Synchronization
	Mutex sync.RWMutex
}

// Initialize the workflow with default config
func NewWorkflow() *Workflow {
	return &Workflow{
		Config: WorkflowConfig{
			SourceEndpointID:      os.Getenv("GLOBUS_SOURCE_ENDPOINT"),
			DestinationEndpointID: os.Getenv("GLOBUS_DESTINATION_ENDPOINT"),
			ComputeEndpointID:     os.Getenv("GLOBUS_COMPUTE_ENDPOINT"),
			SearchIndexID:         os.Getenv("GLOBUS_SEARCH_INDEX"),
			SourcePath:            "/data/input/",
			DestinationPath:       "/data/processing/",
			ResultsPath:           "/data/results/",
			ComputeFunction:       "process_data",
			ComputeArgs:           []string{"--format=csv", "--analyze=true"},
			FlowName:              "Data Processing Workflow",
			FlowDescription:       "Automated workflow for data processing and analysis",
			SkipTransfer:          false,
			SkipCompute:           false,
			SkipSearch:            false,
			SkipFlows:             false,
			WithDashboard:         false,
		},
		Logger:    log.New(os.Stdout, "[Multi-Service Workflow] ", log.LstdFlags),
		StartTime: time.Now(),
	}
}

// Parse command line flags to update config
func (w *Workflow) ParseFlags() {
	flag.StringVar(&w.Config.SourcePath, "source-path", w.Config.SourcePath, "Path on source endpoint")
	flag.StringVar(&w.Config.DestinationPath, "destination-path", w.Config.DestinationPath, "Path on destination endpoint")
	flag.StringVar(&w.Config.ComputeFunction, "compute-function", w.Config.ComputeFunction, "Function to execute on compute endpoint")
	flag.BoolVar(&w.Config.SkipTransfer, "skip-transfer", w.Config.SkipTransfer, "Skip the transfer step")
	flag.BoolVar(&w.Config.SkipCompute, "skip-compute", w.Config.SkipCompute, "Skip the compute step")
	flag.BoolVar(&w.Config.SkipSearch, "skip-search", w.Config.SkipSearch, "Skip the search indexing step")
	flag.BoolVar(&w.Config.SkipFlows, "skip-flows", w.Config.SkipFlows, "Skip the flow creation step")
	flag.BoolVar(&w.Config.WithDashboard, "with-dashboard", w.Config.WithDashboard, "Display progress dashboard")
	flag.Parse()

	// Validate required configuration
	w.validateConfig()
}

// Validate required configuration
func (w *Workflow) validateConfig() {
	if !w.Config.SkipTransfer {
		if w.Config.SourceEndpointID == "" {
			w.Logger.Fatal("Source endpoint ID is required. Set GLOBUS_SOURCE_ENDPOINT environment variable or use --skip-transfer.")
		}
		if w.Config.DestinationEndpointID == "" {
			w.Logger.Fatal("Destination endpoint ID is required. Set GLOBUS_DESTINATION_ENDPOINT environment variable or use --skip-transfer.")
		}
	}

	if !w.Config.SkipCompute {
		if w.Config.ComputeEndpointID == "" {
			w.Logger.Fatal("Compute endpoint ID is required. Set GLOBUS_COMPUTE_ENDPOINT environment variable or use --skip-compute.")
		}
	}

	if !w.Config.SkipSearch {
		if w.Config.SearchIndexID == "" {
			w.Logger.Fatal("Search index ID is required. Set GLOBUS_SEARCH_INDEX environment variable or use --skip-search.")
		}
	}
}

// Initialize the Globus SDK
func (w *Workflow) InitSDK() error {
	// Create SDK with custom options
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	w.SDK = config

	return nil
}

// Run the workflow
func (w *Workflow) Run() error {
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		w.Logger.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Log workflow start
	w.Logger.Println("Starting multi-service workflow")
	w.logConfig()

	// Start progress monitoring if dashboard is enabled
	if w.Config.WithDashboard {
		go w.monitorProgress(ctx)
	}

	// Execute workflow steps
	var err error

	// Step 1: Transfer data
	if !w.Config.SkipTransfer {
		w.Logger.Println("Step 1: Starting data transfer...")
		if err = w.executeTransfer(ctx); err != nil {
			w.Logger.Printf("Transfer step failed: %v", err)
			return err
		}
		w.TransferCompleted = true
		w.Logger.Println("Data transfer completed successfully")
	} else {
		w.Logger.Println("Skipping transfer step as requested")
		w.TransferCompleted = true
	}

	// Step 2: Submit compute job
	if !w.Config.SkipCompute {
		w.Logger.Println("Step 2: Submitting compute job...")
		if err = w.executeCompute(ctx); err != nil {
			w.Logger.Printf("Compute step failed: %v", err)
			return err
		}
		w.ComputeCompleted = true
		w.Logger.Println("Compute job completed successfully")
	} else {
		w.Logger.Println("Skipping compute step as requested")
		w.ComputeCompleted = true
	}

	// Step 3: Index results in Search
	if !w.Config.SkipSearch {
		w.Logger.Println("Step 3: Indexing results in Search...")
		if err = w.executeSearch(ctx); err != nil {
			w.Logger.Printf("Search indexing step failed: %v", err)
			return err
		}
		w.SearchCompleted = true
		w.Logger.Println("Search indexing completed successfully")
	} else {
		w.Logger.Println("Skipping search indexing step as requested")
		w.SearchCompleted = true
	}

	// Step 4: Create Flow for automation
	if !w.Config.SkipFlows {
		w.Logger.Println("Step 4: Creating Flow for automation...")
		if err = w.createFlow(ctx); err != nil {
			w.Logger.Printf("Flow creation step failed: %v", err)
			return err
		}
		w.FlowCreated = true
		w.Logger.Println("Flow created successfully")
	} else {
		w.Logger.Println("Skipping flow creation step as requested")
	}

	// Log workflow completion
	w.EndTime = time.Now()
	w.logSummary()

	return nil
}

// Log the current configuration
func (w *Workflow) logConfig() {
	w.Logger.Println("=== Workflow Configuration ===")
	w.Logger.Printf("Source: Endpoint=%s, Path=%s\n",
		w.Config.SourceEndpointID, w.Config.SourcePath)
	w.Logger.Printf("Destination: Endpoint=%s, Path=%s\n",
		w.Config.DestinationEndpointID, w.Config.DestinationPath)
	w.Logger.Printf("Compute: Endpoint=%s, Function=%s\n",
		w.Config.ComputeEndpointID, w.Config.ComputeFunction)
	w.Logger.Printf("Search: Index=%s\n", w.Config.SearchIndexID)
	w.Logger.Printf("Flow: Name=%s\n", w.Config.FlowName)
	w.Logger.Printf("Options: SkipTransfer=%v, SkipCompute=%v, SkipSearch=%v, SkipFlows=%v, WithDashboard=%v\n",
		w.Config.SkipTransfer, w.Config.SkipCompute, w.Config.SkipSearch,
		w.Config.SkipFlows, w.Config.WithDashboard)
	w.Logger.Println("===============================")
}

// Log workflow summary
func (w *Workflow) logSummary() {
	duration := w.EndTime.Sub(w.StartTime)
	w.Logger.Println("\n=== Workflow Summary ===")
	w.Logger.Printf("Total Duration: %v\n", duration.Round(time.Second))
	if w.TransferCompleted && !w.Config.SkipTransfer {
		w.Logger.Printf("Transfer: TaskID=%s, Data Transferred=%d bytes\n",
			w.TransferTaskID, w.BytesTransferred)
	}
	if w.ComputeCompleted && !w.Config.SkipCompute {
		w.Logger.Printf("Compute: TaskID=%s, Compute Time=%v\n",
			w.ComputeTaskID, w.ComputeTime.Round(time.Second))
	}
	if w.SearchCompleted && !w.Config.SkipSearch {
		w.Logger.Printf("Search: Documents Indexed=%d\n", w.SearchIndexCount)
	}
	if w.FlowCreated && !w.Config.SkipFlows {
		w.Logger.Printf("Flow: FlowID=%s\n", w.FlowID)
	}
	if w.ErrorCount > 0 {
		w.Logger.Printf("Errors: %d errors encountered and recovered\n", w.ErrorCount)
	}
	w.Logger.Println("========================")
}

// Monitor workflow progress
func (w *Workflow) monitorProgress(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.displayProgress()
		}
	}
}

// Display workflow progress
func (w *Workflow) displayProgress() {
	w.Mutex.RLock()
	defer w.Mutex.RUnlock()

	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println("=== Workflow Progress Dashboard ===")
	fmt.Printf("Elapsed Time: %v\n\n", time.Since(w.StartTime).Round(time.Second))

	// Transfer progress
	fmt.Println("1. Transfer:")
	if w.Config.SkipTransfer {
		fmt.Println("   [SKIPPED]")
	} else if w.TransferCompleted {
		fmt.Printf("   [COMPLETE] TaskID: %s, Transferred: %d bytes\n",
			w.TransferTaskID, w.BytesTransferred)
	} else if w.TransferTaskID != "" {
		fmt.Printf("   [IN PROGRESS] TaskID: %s\n", w.TransferTaskID)
	} else {
		fmt.Println("   [PENDING]")
	}

	// Compute progress
	fmt.Println("\n2. Compute:")
	if w.Config.SkipCompute {
		fmt.Println("   [SKIPPED]")
	} else if w.ComputeCompleted {
		fmt.Printf("   [COMPLETE] TaskID: %s, Duration: %v\n",
			w.ComputeTaskID, w.ComputeTime.Round(time.Second))
	} else if w.ComputeTaskID != "" {
		fmt.Printf("   [IN PROGRESS] TaskID: %s\n", w.ComputeTaskID)
	} else {
		fmt.Println("   [PENDING]")
	}

	// Search progress
	fmt.Println("\n3. Search Indexing:")
	if w.Config.SkipSearch {
		fmt.Println("   [SKIPPED]")
	} else if w.SearchCompleted {
		fmt.Printf("   [COMPLETE] Indexed: %d documents\n", w.SearchIndexCount)
	} else if w.SearchIndexCount > 0 {
		fmt.Printf("   [IN PROGRESS] Indexed: %d documents so far\n", w.SearchIndexCount)
	} else {
		fmt.Println("   [PENDING]")
	}

	// Flow progress
	fmt.Println("\n4. Flow Creation:")
	if w.Config.SkipFlows {
		fmt.Println("   [SKIPPED]")
	} else if w.FlowCreated {
		fmt.Printf("   [COMPLETE] FlowID: %s\n", w.FlowID)
	} else {
		fmt.Println("   [PENDING]")
	}

	// Error count
	if w.ErrorCount > 0 {
		fmt.Printf("\nErrors: %d (recovered)\n", w.ErrorCount)
	}

	fmt.Println("\nPress Ctrl+C to cancel the workflow")
	fmt.Println("===================================")
}

// Execute the transfer step
func (w *Workflow) executeTransfer(ctx context.Context) error {
	w.Logger.Println("Initiating transfer from source to destination")

	// Create transfer client
	var transferClient *transfer.Client
	var err error
	transferClient, err = w.SDK.NewTransferClient(os.Getenv("GLOBUS_TRANSFER_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Submit the transfer
	w.Logger.Println("Submitting transfer task...")
	result, err := transferClient.SubmitTransfer(
		ctx,
		w.Config.SourceEndpointID,
		w.Config.SourcePath,
		w.Config.DestinationEndpointID,
		w.Config.DestinationPath,
		"Multi-Service Workflow Data Transfer",
		map[string]interface{}{
			"recursive":       true,
			"verify_checksum": true,
			"sync_level":      3, // Checksum sync level
			"preserve_mtime":  true,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to submit transfer task: %w", err)
	}

	w.Mutex.Lock()
	w.TransferTaskID = result.TaskID
	w.Mutex.Unlock()

	w.Logger.Printf("Transfer task submitted with ID: %s", result.TaskID)

	// Wait for transfer to complete
	w.Logger.Println("Waiting for transfer to complete...")

	// Poll every 5 seconds for completion
	pollInterval := 5 * time.Second
	completion, err := transferClient.WaitForTaskCompletion(ctx, result.TaskID, pollInterval)
	if err != nil {
		return fmt.Errorf("error waiting for transfer task: %w", err)
	}

	// Check if the task completed successfully
	if completion.Status != "SUCCEEDED" {
		return fmt.Errorf("transfer failed with status: %s", completion.Status)
	}

	// Get task details for metrics
	taskInfo, err := transferClient.GetTask(ctx, result.TaskID)
	if err != nil {
		w.Logger.Printf("Warning: Could not get task details: %v", err)
	} else {
		w.Mutex.Lock()
		w.BytesTransferred = taskInfo.BytesTransferred
		w.Mutex.Unlock()
	}

	w.Logger.Printf("Transfer completed successfully. Transferred %d bytes.", taskInfo.BytesTransferred)
	return nil
}

// Execute the compute step
func (w *Workflow) executeCompute(ctx context.Context) error {
	w.Logger.Println("Initiating compute job")

	// Create compute client
	computeClient, err := w.SDK.NewComputeClient(os.Getenv("GLOBUS_COMPUTE_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	// Set up the compute job
	computeStart := time.Now()

	// Submit a task batch (POST /v2/submit). The document shape is defined by the
	// Compute API; the Go client sends it as a passthrough document.
	w.Logger.Println("Submitting compute batch...")
	submitDoc := map[string]interface{}{
		"tasks": map[string]interface{}{
			w.Config.ComputeEndpointID: []interface{}{
				map[string]interface{}{
					"function": w.Config.ComputeFunction,
					"args":     convertToAnySlice(w.Config.ComputeArgs),
					"kwargs": map[string]interface{}{
						"input_path":  w.Config.DestinationPath,
						"output_path": w.Config.ResultsPath,
					},
				},
			},
		},
	}
	result, err := computeClient.Submit(ctx, submitDoc)
	if err != nil {
		return fmt.Errorf("failed to submit compute batch: %w", err)
	}

	taskIDs := extractTaskIDs(result)
	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs returned from submission")
	}
	w.Mutex.Lock()
	w.ComputeTaskID = taskIDs[0]
	w.Mutex.Unlock()
	w.Logger.Printf("Compute batch submitted with %d tasks. Primary task ID: %s", len(taskIDs), w.ComputeTaskID)

	// Poll for completion via batch status (passthrough response).
	w.Logger.Println("Waiting for compute job to complete...")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			status, err := computeClient.GetBatchStatus(ctx, taskIDs)
			if err != nil {
				w.Logger.Printf("Warning: Error checking batch status: %v, retrying...", err)
				w.Mutex.Lock()
				w.ErrorCount++
				w.Mutex.Unlock()
				continue
			}
			if allTasksComplete(status, taskIDs) {
				computeEnd := time.Now()
				w.Mutex.Lock()
				w.ComputeTime = computeEnd.Sub(computeStart)
				w.Mutex.Unlock()
				w.Logger.Printf("Compute job completed in %v", w.ComputeTime.Round(time.Second))
				return nil
			}
		}
	}
}

// extractTaskIDs pulls task IDs from a passthrough submit response.
func extractTaskIDs(result map[string]interface{}) []string {
	raw, ok := result["task_ids"].([]interface{})
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			ids = append(ids, s)
		}
	}
	return ids
}

// allTasksComplete reports whether every task in the batch-status response has a
// terminal status.
func allTasksComplete(status map[string]interface{}, taskIDs []string) bool {
	results, ok := status["results"].(map[string]interface{})
	if !ok {
		return false
	}
	for _, id := range taskIDs {
		t, ok := results[id].(map[string]interface{})
		if !ok {
			return false
		}
		if s, _ := t["status"].(string); s != "success" && s != "failed" {
			return false
		}
	}
	return true
}

// Execute the search indexing step
func (w *Workflow) executeSearch(ctx context.Context) error {
	w.Logger.Println("Initiating search indexing of results")

	// Create search client
	searchClient, err := w.SDK.NewSearchClient(os.Getenv("GLOBUS_SEARCH_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create search client: %w", err)
	}

	// Simulate getting result files
	resultFiles := []string{
		"result1.csv",
		"result2.csv",
		"analysis_summary.json",
		"visualization.png",
	}

	// Create search documents for each result file
	var documents []search.SearchDocument
	for i, file := range resultFiles {
		// Create a document with metadata
		doc := search.SearchDocument{
			Subject: fmt.Sprintf("workflow-result-%d", i),
			Content: map[string]interface{}{
				"filename":     file,
				"path":         fmt.Sprintf("%s/%s", w.Config.ResultsPath, file),
				"workflow_id":  fmt.Sprintf("multi-service-workflow-%d", time.Now().Unix()),
				"processed_at": time.Now().Format(time.RFC3339),
				"source_data":  w.Config.SourcePath,
				"file_type":    getFileExtension(file),
				"size":         1024 * (i + 1), // Simulated file size
				"description":  fmt.Sprintf("Result file from compute job %s", w.ComputeTaskID),
			},
		}
		documents = append(documents, doc)

		// Update our count for progress tracking
		w.Mutex.Lock()
		w.SearchIndexCount++
		w.Mutex.Unlock()
	}

	// Index the documents
	w.Logger.Printf("Indexing %d documents in search index %s...", len(documents), w.Config.SearchIndexID)
	// Create an ingest request
	ingestRequest := &search.IngestRequest{
		IndexID:   w.Config.SearchIndexID,
		Documents: documents,
	}
	_, err = searchClient.IngestDocuments(ctx, ingestRequest)
	if err != nil {
		return fmt.Errorf("failed to index documents: %w", err)
	}

	w.Logger.Printf("Successfully indexed %d documents in search", len(documents))
	return nil
}

// Create a flow for automation
func (w *Workflow) createFlow(ctx context.Context) error {
	w.Logger.Println("Creating flow for workflow automation")

	// Create flows client
	flowsClient, err := w.SDK.NewFlowsClient(os.Getenv("GLOBUS_FLOWS_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create flows client: %w", err)
	}

	// Define flow
	flow := &flows.FlowCreateRequest{
		Title:       w.Config.FlowName,
		Description: w.Config.FlowDescription,
		Definition: map[string]interface{}{
			"StartAt": "TransferData",
			"States": map[string]interface{}{
				"TransferData": map[string]interface{}{
					"Type":      "Action",
					"ActionURL": "https://actions.globus.org/transfer/transfer",
					"Parameters": map[string]interface{}{
						"source_endpoint_id":      w.Config.SourceEndpointID,
						"destination_endpoint_id": w.Config.DestinationEndpointID,
						"transfer_items": []map[string]interface{}{
							{
								"source_path":      w.Config.SourcePath,
								"destination_path": w.Config.DestinationPath,
								"recursive":        true,
							},
						},
						"sync_level": 1,
					},
					"ResultPath": "$.TransferResult",
					"Next":       "ProcessData",
				},
				"ProcessData": map[string]interface{}{
					"Type":      "Action",
					"ActionURL": "https://actions.globus.org/compute/batch",
					"Parameters": map[string]interface{}{
						"endpoint":   w.Config.ComputeEndpointID,
						"batch_type": "single",
						"commands": []map[string]interface{}{
							{
								"command_line": []string{w.Config.ComputeFunction},
								"working_dir":  w.Config.DestinationPath,
							},
						},
					},
					"ResultPath": "$.ComputeResult",
					"Next":       "IndexResults",
				},
				"IndexResults": map[string]interface{}{
					"Type":      "Action",
					"ActionURL": "https://actions.globus.org/search/ingest",
					"Parameters": map[string]interface{}{
						"search_index": w.Config.SearchIndexID,
						"documents": []map[string]interface{}{
							{
								"id": "$.id",
								"content": map[string]interface{}{
									"workflow_id":  "$.id",
									"processed_at": "$.startTime",
									"result_path":  w.Config.ResultsPath,
								},
							},
						},
					},
					"ResultPath": "$.SearchResult",
					"Next":       "NotifyCompletion",
				},
				"NotifyCompletion": map[string]interface{}{
					"Type":      "Action",
					"ActionURL": "https://actions.globus.org/notification",
					"Parameters": map[string]interface{}{
						"body":    "The workflow has completed successfully.",
						"subject": "Workflow Completed: " + w.Config.FlowName,
					},
					"End": true,
				},
			},
		},
	}

	// Create the flow
	w.Logger.Println("Submitting flow creation request...")
	result, err := flowsClient.CreateFlow(ctx, flow)
	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	w.Mutex.Lock()
	w.FlowID = result.ID
	w.Mutex.Unlock()

	w.Logger.Printf("Flow created successfully with ID: %s", result.ID)
	return nil
}

// Get file extension
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}

func main() {
	// Create and run the workflow
	workflow := NewWorkflow()
	workflow.ParseFlags()

	err := workflow.InitSDK()
	if err != nil {
		log.Fatalf("Failed to initialize SDK: %v", err)
	}

	err = workflow.Run()
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	log.Println("Workflow completed successfully")
}

// Helper to convert string slice to interface{} slice
func convertToAnySlice(strSlice []string) []interface{} {
	result := make([]interface{}, len(strSlice))
	for i, v := range strSlice {
		result[i] = v
	}
	return result
}
