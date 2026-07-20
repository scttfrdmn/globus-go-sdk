// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
)

// Demonstrates submitting a task batch to Globus Compute. The submit and
// batch-status bodies/results are open-ended documents (map[string]interface{}).
func main() {
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	endpointID := os.Getenv("GLOBUS_COMPUTE_ENDPOINT")
	if accessToken == "" || endpointID == "" {
		fmt.Println("Set GLOBUS_ACCESS_TOKEN and GLOBUS_COMPUTE_ENDPOINT")
		os.Exit(1)
	}

	config := pkg.NewConfigFromEnvironment()
	computeClient, err := config.NewComputeClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}
	ctx := context.Background()

	// Submit a batch of tasks (POST /v2/submit). The document shape is defined by
	// the Compute API; a minimal placeholder is shown here.
	submitDoc := map[string]interface{}{
		"tasks": map[string]interface{}{
			endpointID: []interface{}{},
		},
	}
	result, err := computeClient.Submit(ctx, submitDoc)
	if err != nil {
		log.Fatalf("Submit failed: %v", err)
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Submit result: %s\n", b)

	// Query the status of returned task IDs via POST /v2/batch_status.
	if taskIDs, ok := result["task_ids"].([]interface{}); ok && len(taskIDs) > 0 {
		ids := make([]string, 0, len(taskIDs))
		for _, id := range taskIDs {
			if s, ok := id.(string); ok {
				ids = append(ids, s)
			}
		}
		status, err := computeClient.GetBatchStatus(ctx, ids)
		if err != nil {
			log.Printf("GetBatchStatus failed: %v", err)
		} else {
			sb, _ := json.MarshalIndent(status, "", "  ")
			fmt.Printf("Batch status: %s\n", sb)
		}
	}
	fmt.Println("Done.")
}
