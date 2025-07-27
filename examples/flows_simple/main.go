// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
)

func main() {
	// Just create some structs to check for compilation
	batchOpt := &flows.BatchOptions{
		Concurrency: 5,
	}

	batchReq := &flows.BatchFlowsRequest{
		FlowIDs: []string{"test"},
		Options: batchOpt,
	}

	fmt.Printf("Batch request: %v\n", batchReq)
}
