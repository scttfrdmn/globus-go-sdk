// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// streams-tunnels demonstrates the Globus Streams/Tunnels API, which provides
// persistent channels for streaming data between endpoints.
//
// The Streams/Tunnels API was added in Python SDK v4.3.0 (tunnels) and
// v4.4.0 (tunnel events).
//
// Required environment variables:
//
//	GLOBUS_ACCESS_TOKEN  - A valid Globus Transfer access token
//
// Optional environment variables:
//
//	GLOBUS_TUNNEL_ID     - An existing tunnel ID to query events for
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
	// Retrieve the access token from the environment.
	// In a real application you would obtain this via an OAuth2 flow.
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_ACCESS_TOKEN environment variable is required")
	}

	// Create a TransferClient using the SDK configuration helper.
	transferClient, err := pkg.NewConfig().NewTransferClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create transfer client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// ------------------------------------------------------------------
	// Step 1: List existing tunnels
	// ------------------------------------------------------------------
	fmt.Println("=== Listing Tunnels ===")

	tunnelList, err := transferClient.ListTunnels(ctx, &transfer.ListTunnelsOptions{
		Limit: 10,
	})
	if err != nil {
		log.Fatalf("Failed to list tunnels: %v", err)
	}

	fmt.Printf("Total tunnels: %d\n", tunnelList.Total)

	if len(tunnelList.Tunnels) == 0 {
		fmt.Println("No tunnels found for this account.")
	} else {
		for i, t := range tunnelList.Tunnels {
			fmt.Printf("\nTunnel %d:\n", i+1)
			fmt.Printf("  ID:           %s\n", t.ID)
			fmt.Printf("  Display Name: %s\n", t.DisplayName)
			fmt.Printf("  Status:       %s\n", t.Status)
			fmt.Printf("  Owner:        %s\n", t.Owner)
			if t.SourceEndpointID != "" {
				fmt.Printf("  Source Endpoint: %s\n", t.SourceEndpointID)
				fmt.Printf("  Source Path:     %s\n", t.SourcePath)
			}
			if t.CreatedAt != nil {
				fmt.Printf("  Created At:   %s\n", t.CreatedAt.Format(time.RFC3339))
			}
			if t.ExpiresAt != nil {
				fmt.Printf("  Expires At:   %s\n", t.ExpiresAt.Format(time.RFC3339))
			}
		}
	}

	if tunnelList.HasMore {
		fmt.Printf("\n(More tunnels available; next marker: %s)\n", tunnelList.Marker)
	}

	// ------------------------------------------------------------------
	// Step 2: Demonstrate creating a tunnel
	//
	// NOTE: Creating a tunnel requires a real, active Globus endpoint ID
	// and an appropriate Transfer scope with write permissions.  The code
	// below is shown for reference but is intentionally gated behind a
	// placeholder check so the example can be run without side-effects.
	// ------------------------------------------------------------------
	fmt.Println("\n=== Creating a Tunnel (demonstration) ===")

	// Replace these values with a real endpoint UUID and path to actually
	// create a tunnel.
	const exampleEndpointID = "YOUR-ENDPOINT-UUID-HERE"
	const examplePath = "/~/streams-example/"

	if exampleEndpointID == "YOUR-ENDPOINT-UUID-HERE" {
		fmt.Println("NOTE: Tunnel creation is skipped in this example because it requires")
		fmt.Println("      a real Globus endpoint ID.  To create a tunnel, set a valid")
		fmt.Println("      endpoint UUID and path in the source code (or pass them as")
		fmt.Println("      environment variables) and remove this guard.")
		fmt.Println()
		fmt.Println("Example CreateTunnelData:")
		fmt.Println("  transfer.CreateTunnelData{")
		fmt.Printf("    DisplayName:      \"Example Streaming Tunnel\",\n")
		fmt.Printf("    SourceEndpointID: \"<endpoint-uuid>\",\n")
		fmt.Printf("    SourcePath:       \"/~/data/\",\n")
		fmt.Println("  }")
	} else {
		// Actual creation call – only reached when a real endpoint ID is set.
		createData := &transfer.CreateTunnelData{
			DisplayName:      "Example Streaming Tunnel " + time.Now().Format("20060102_150405"),
			SourceEndpointID: exampleEndpointID,
			SourcePath:       examplePath,
		}

		newTunnel, err := transferClient.CreateTunnel(ctx, createData)
		if err != nil {
			log.Fatalf("Failed to create tunnel: %v", err)
		}

		fmt.Printf("Created tunnel: %s (%s)\n", newTunnel.DisplayName, newTunnel.ID)
		fmt.Printf("  Status: %s\n", newTunnel.Status)
	}

	// ------------------------------------------------------------------
	// Step 3: Get tunnel events for an existing tunnel
	//
	// Tunnel events were added in Python SDK v4.4.0.
	// ------------------------------------------------------------------
	fmt.Println("\n=== Tunnel Events ===")

	tunnelID := os.Getenv("GLOBUS_TUNNEL_ID")

	// If no explicit tunnel ID was provided, try to use the first tunnel
	// returned by ListTunnels.
	if tunnelID == "" && len(tunnelList.Tunnels) > 0 {
		tunnelID = tunnelList.Tunnels[0].ID
		fmt.Printf("Using first listed tunnel: %s\n", tunnelID)
	}

	if tunnelID == "" {
		fmt.Println("No tunnel ID available.  Set GLOBUS_TUNNEL_ID to fetch events")
		fmt.Println("for a specific tunnel, or ensure at least one tunnel exists.")
		fmt.Println()
		fmt.Println("Example GetTunnelEvents call:")
		fmt.Println("  events, err := transferClient.GetTunnelEvents(ctx, \"<tunnel-id>\",")
		fmt.Println("      &transfer.ListTunnelEventsOptions{Limit: 10})")
	} else {
		events, err := transferClient.GetTunnelEvents(ctx, tunnelID, &transfer.ListTunnelEventsOptions{
			Limit: 10,
		})
		if err != nil {
			log.Fatalf("Failed to get tunnel events for %s: %v", tunnelID, err)
		}

		fmt.Printf("Tunnel %s has %d event(s) (total: %d):\n", tunnelID, len(events.Events), events.Total)

		if len(events.Events) == 0 {
			fmt.Println("  No events recorded for this tunnel yet.")
		} else {
			for i, ev := range events.Events {
				fmt.Printf("\n  Event %d:\n", i+1)
				fmt.Printf("    ID:          %s\n", ev.ID)
				fmt.Printf("    Code:        %s\n", ev.Code)
				if ev.Description != "" {
					fmt.Printf("    Description: %s\n", ev.Description)
				}
				if ev.OccurredAt != nil {
					fmt.Printf("    Occurred At: %s\n", ev.OccurredAt.Format(time.RFC3339))
				}
				if len(ev.Details) > 0 {
					fmt.Printf("    Details:     %v\n", ev.Details)
				}
			}
		}

		if events.HasMore {
			fmt.Printf("\n(More events available; next marker: %s)\n", events.Marker)
		}
	}

	// ------------------------------------------------------------------
	// Step 4: Retrieve a Stream Access Point by ID
	//
	// Stream Access Points provide URL-based access to real-time data
	// streams.  They are created by the Globus service when a tunnel
	// becomes active.  A GetStreamAccessPoint call is shown here; a
	// ListStreamAccessPoints endpoint is not yet part of the Transfer API
	// surface exposed by this SDK (access points are retrieved per-ID).
	// ------------------------------------------------------------------
	fmt.Println("\n=== Stream Access Points ===")
	fmt.Println("Stream Access Points are associated with active tunnels.")
	fmt.Println("Use GetStreamAccessPoint(ctx, accessPointID) to retrieve a")
	fmt.Println("specific access point once you have its ID.")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  ap, err := transferClient.GetStreamAccessPoint(ctx, \"<access-point-id>\")")
	fmt.Println("  if err != nil { ... }")
	fmt.Println("  fmt.Println(\"Access URL:\", ap.AccessURL)")

	fmt.Println("\nStreams/Tunnels example complete.")
}
