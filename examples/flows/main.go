// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
	// Setup placeholder client
	client := flows.NewClient("fake-token", core.WithLogLevel(core.LogLevelDebug))
	ctx := context.Background()

	// Try using the problematic types
	flowIDs := []string{"flow1", "flow2"}

	// Example of BatchGetFlows
	fmt.Println("Batch retrieving flows...")
	batchFlowsResp := client.BatchGetFlows(ctx, &flows.BatchFlowsRequest{
		FlowIDs: flowIDs,
		Options: &flows.BatchOptions{
			Concurrency: 5,
		},
	})

	fmt.Printf("Batch flow response: %v\n", batchFlowsResp)

	// Create multiple run requests
	runRequests := []*flows.RunRequest{
		{
			FlowID: "flow1",
			Label:  "Batch run 1",
			Input: map[string]interface{}{
				"message": "Batch message 1",
			},
		},
	}

	fmt.Println("Starting batch flow runs...")
	batchRunResp := client.BatchRunFlows(ctx, &flows.BatchRunFlowsRequest{
		Requests: runRequests,
		Options: &flows.BatchOptions{
			Concurrency: 2,
		},
	})

	fmt.Printf("Batch run response: %v\n", batchRunResp)

	// Example of batch canceling runs
	runIDs := []string{"run1", "run2"}
	cancelResp := client.BatchCancelRuns(ctx, &flows.BatchCancelRunsRequest{
		RunIDs: runIDs,
		Options: &flows.BatchOptions{
			Concurrency: 2,
		},
	})

	fmt.Printf("Cancel response: %v\n", cancelResp)
}
