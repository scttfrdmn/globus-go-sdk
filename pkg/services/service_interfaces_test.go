// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package services

import (
	"testing"
	
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// TestServiceClientInterfaces verifies that all service clients use the core.Client
// type and that their transport members implement the Transport interface
func TestServiceClientInterfaces(t *testing.T) {
	// Verify that each service client embeds a core.Client
	verifyClientType(t, "auth.Client", func() *core.Client {
		client, _ := auth.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "compute.Client", func() *core.Client {
		client, _ := compute.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "flows.Client", func() *core.Client {
		client, _ := flows.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "groups.Client", func() *core.Client {
		client, _ := groups.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "search.Client", func() *core.Client {
		client, _ := search.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "timers.Client", func() *core.Client {
		client, _ := timers.NewClient()
		return client.Client
	})
	
	verifyClientType(t, "transfer.Client", func() *core.Client {
		client, _ := transfer.NewClient()
		return client.Client
	})
}

// verifyClientType is a helper function that verifies that the returned client
// is a valid *core.Client
func verifyClientType(t *testing.T, name string, getClient func() *core.Client) {
	t.Run(name, func(t *testing.T) {
		client := getClient()
		if client == nil {
			t.Fatalf("%s should have a valid core.Client", name)
		}
		
		// Verify that it has a VersionCheck field
		if client.VersionCheck == nil {
			t.Fatalf("%s's Client should have a VersionCheck field", name)
		}
	})
}

// TestServiceTransportInterfaces verifies that all service transports
// implement the Transport interface
func TestServiceTransportInterfaces(t *testing.T) {
	// This test verifies at compile time that service transports implement 
	// the Transport interface
	
	// For each service that has a Transport member, verify it implements interfaces.Transport
	var _ interfaces.Transport = &auth.Client{}.Transport
	var _ interfaces.Transport = &compute.Client{}.Transport
	var _ interfaces.Transport = &flows.Client{}.Transport
	var _ interfaces.Transport = &groups.Client{}.Transport
	var _ interfaces.Transport = &search.Client{}.Transport
	var _ interfaces.Transport = &timers.Client{}.Transport
	var _ interfaces.Transport = &transfer.Client{}.Transport
	
	t.Skip("Test is compile-time only, not meant to be executed")
}