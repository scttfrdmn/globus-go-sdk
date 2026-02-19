// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// flows-auth-policy demonstrates how to use FlowAuthenticationPolicy when
// creating and updating Globus Flows.  FlowAuthenticationPolicy was added in
// Python SDK v4.1.0 and allows a flow owner to require specific authentication
// levels (high-assurance, MFA, or named session policies) before a run is
// permitted.
//
// Required environment variables:
//   GLOBUS_CLIENT_ID      - Globus application client ID
//   GLOBUS_CLIENT_SECRET  - Globus application client secret
//
// Optional environment variables:
//   GLOBUS_FLOW_ID        - An existing flow ID to query and run
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
)

// boolPtr is a small helper that returns a pointer to a bool literal.
// FlowAuthenticationPolicy fields use *bool so that "false" can be
// distinguished from "omitted".
func boolPtr(b bool) *bool {
	return &b
}

func main() {
	// Build SDK config from environment.  Client credentials are needed to
	// obtain an access token via the client-credentials grant.
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create an Auth client to fetch a Flows access token.
	authClient, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	ctx := context.Background()

	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.FlowsScope)
	if err != nil {
		log.Fatalf("Failed to get Flows access token: %v", err)
	}
	fmt.Printf("Obtained Flows access token (expires in %d seconds)\n", tokenResp.ExpiresIn)

	// Create the Flows client.
	flowsClient, err := config.NewFlowsClient(tokenResp.AccessToken)
	if err != nil {
		log.Fatalf("Failed to create flows client: %v", err)
	}

	// ------------------------------------------------------------------
	// Step 1: Create a flow that includes a FlowAuthenticationPolicy
	// ------------------------------------------------------------------
	fmt.Println("\n=== Creating Flow with AuthenticationPolicy ===")

	// A minimal flow definition: a single Pass state that echoes its input.
	timestamp := time.Now().Format("20060102_150405")
	flowTitle := fmt.Sprintf("Auth-Policy Example Flow %s", timestamp)

	flowDefinition := map[string]interface{}{
		"Comment": "Minimal example flow demonstrating authentication policy",
		"StartAt": "Echo",
		"States": map[string]interface{}{
			"Echo": map[string]interface{}{
				"Type":       "Pass",
				"Result":     "$.message",
				"ResultPath": "$.output",
				"End":        true,
			},
		},
	}

	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "A message to echo back as the flow output",
			},
		},
		"required":             []string{"message"},
		"additionalProperties": false,
	}

	// FlowAuthenticationPolicy lets the flow owner specify that callers
	// must satisfy certain authentication requirements before a run is
	// permitted.  All three fields are optional and independent:
	//
	//   HighAssurance  – requires a high-assurance (HA) authentication session
	//   RequiredMFA    – requires multi-factor authentication
	//   SessionPolicies – one or more named Globus Auth session policies that
	//                     the caller's session must satisfy
	//
	// Here we enable HighAssurance and RequiredMFA to illustrate their use.
	// In a real flow you would choose only the constraints appropriate for
	// the data being handled.
	authPolicy := &flows.FlowAuthenticationPolicy{
		HighAssurance: boolPtr(true),
		RequiredMFA:   boolPtr(true),
		// SessionPolicies can reference named policies registered in Globus Auth.
		// SessionPolicies: []string{"urn:globus:auth:policies:example-policy"},
	}

	createRequest := &flows.FlowCreateRequest{
		Title:                flowTitle,
		Description:          "Example flow with authentication policy (created by Globus Go SDK)",
		Definition:           flowDefinition,
		InputSchema:          inputSchema,
		Keywords:             []string{"example", "auth-policy", "go-sdk"},
		AuthenticationPolicy: authPolicy,
	}

	newFlow, err := flowsClient.CreateFlow(ctx, createRequest)
	if err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}

	fmt.Printf("Created flow: %s\n", newFlow.Title)
	fmt.Printf("  ID:    %s\n", newFlow.ID)
	fmt.Printf("  Owner: %s\n", newFlow.FlowOwner)
	fmt.Println()

	// Print the full create request as JSON so the policy structure is clear.
	createJSON, _ := json.MarshalIndent(createRequest, "", "  ")
	fmt.Println("FlowCreateRequest (with authentication_policy):")
	fmt.Println(string(createJSON))

	flowID := newFlow.ID

	// ------------------------------------------------------------------
	// Step 2: Update the flow to change its authentication policy
	//
	// UpdateFlow accepts a FlowUpdateRequest that also carries an optional
	// AuthenticationPolicy.  Here we clear RequiredMFA and add a session
	// policy name to show how the policy can be modified post-creation.
	// ------------------------------------------------------------------
	fmt.Println("\n=== Updating Flow AuthenticationPolicy ===")

	updatedPolicy := &flows.FlowAuthenticationPolicy{
		HighAssurance: boolPtr(true),
		RequiredMFA:   boolPtr(false), // Relax MFA requirement
		SessionPolicies: []string{
			// Replace with a real policy URN from your Globus Auth configuration.
			"urn:globus:auth:policies:example-placeholder",
		},
	}

	updateRequest := &flows.FlowUpdateRequest{
		Description:          "Updated description – auth policy relaxed to HA-only with session policy",
		AuthenticationPolicy: updatedPolicy,
	}

	updatedFlow, err := flowsClient.UpdateFlow(ctx, flowID, updateRequest)
	if err != nil {
		log.Fatalf("Failed to update flow: %v", err)
	}

	fmt.Printf("Updated flow: %s (%s)\n", updatedFlow.Title, updatedFlow.ID)
	updateJSON, _ := json.MarshalIndent(updateRequest, "", "  ")
	fmt.Println("FlowUpdateRequest sent:")
	fmt.Println(string(updateJSON))

	// ------------------------------------------------------------------
	// Step 3: List flows (to confirm the newly created flow appears)
	// ------------------------------------------------------------------
	fmt.Println("\n=== Listing Flows ===")

	flowList, err := flowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
		Limit:   5,
		OrderBy: "updated_at DESC",
	})
	if err != nil {
		log.Fatalf("Failed to list flows: %v", err)
	}

	fmt.Printf("Found %d flow(s) (page of %d):\n", len(flowList.Flows), flowList.Limit)
	for i, f := range flowList.Flows {
		marker := " "
		if f.ID == flowID {
			marker = "*" // mark our newly created flow
		}
		fmt.Printf("%s %d. %s (%s)\n", marker, i+1, f.Title, f.ID)
	}

	// ------------------------------------------------------------------
	// Step 4: Run the flow
	//
	// NOTE: If the authentication policy requires HA or MFA and the token
	// used here does not satisfy those requirements, RunFlow will return a
	// 403 Forbidden error.  In a real application you would direct the user
	// through an appropriate login flow first.
	// ------------------------------------------------------------------
	fmt.Println("\n=== Running Flow ===")

	runRequest := &flows.RunRequest{
		FlowID: flowID,
		Label:  "Auth-Policy SDK Example Run " + timestamp,
		Tags:   []string{"example", "auth-policy", "go-sdk"},
		Input: map[string]interface{}{
			"message": "Hello from the Globus Go SDK auth-policy example!",
		},
	}

	run, err := flowsClient.RunFlow(ctx, runRequest)
	if err != nil {
		// A 403 here likely means the token does not satisfy the policy.
		log.Fatalf("Failed to run flow: %v\n"+
			"(If the error is 403 Forbidden, the access token may not satisfy the\n"+
			" authentication policy attached to the flow.  Use a token obtained via\n"+
			" a high-assurance or MFA-enabled session.)", err)
	}

	fmt.Printf("Flow run started:\n")
	fmt.Printf("  Run ID:    %s\n", run.RunID)
	fmt.Printf("  Status:    %s\n", run.Status)
	fmt.Printf("  Created:   %s\n", run.CreatedAt.Format(time.RFC3339))

	// Poll briefly for a terminal status.
	fmt.Println("\nWaiting up to 30 seconds for the run to complete...")
	waitCtx, waitCancel := context.WithTimeout(ctx, 30*time.Second)
	defer waitCancel()

	finalRun, err := flowsClient.WaitForRun(waitCtx, run.RunID, 3*time.Second)
	if err != nil {
		// Context deadline exceeded just means the run is still in progress.
		fmt.Printf("Run did not complete within the wait window: %v\n", err)
		fmt.Printf("Poll status manually using: GET /runs/%s\n", run.RunID)
	} else {
		fmt.Printf("\n=== Run Completed ===\n")
		fmt.Printf("  Status:   %s\n", finalRun.Status)
		if !finalRun.CompletedAt.IsZero() {
			fmt.Printf("  Duration: %s\n", finalRun.CompletedAt.Sub(finalRun.CreatedAt))
		}
		if finalRun.Output != nil {
			outputJSON, _ := json.MarshalIndent(finalRun.Output, "", "  ")
			fmt.Printf("  Output:\n%s\n", outputJSON)
		}
	}

	fmt.Println("\nFlows auth-policy example complete.")
	fmt.Printf("\nFlow %s remains in your account.  Delete it via:\n", flowID)
	fmt.Printf("  flowsClient.DeleteFlow(ctx, %q)\n", flowID)
}
