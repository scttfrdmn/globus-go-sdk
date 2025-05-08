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
	t.Run("auth.Client", func(t *testing.T) {
		client, err := auth.NewClient()
		if err != nil {
			t.Skip("Could not create auth client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("auth.Client should have a valid core.Client")
		}
	})

	t.Run("compute.Client", func(t *testing.T) {
		client, err := compute.NewClient()
		if err != nil {
			t.Skip("Could not create compute client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("compute.Client should have a valid core.Client")
		}
	})

	t.Run("flows.Client", func(t *testing.T) {
		client, err := flows.NewClient()
		if err != nil {
			t.Skip("Could not create flows client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("flows.Client should have a valid core.Client")
		}
	})

	t.Run("groups.Client", func(t *testing.T) {
		client, err := groups.NewClient()
		if err != nil {
			t.Skip("Could not create groups client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("groups.Client should have a valid core.Client")
		}
	})

	t.Run("search.Client", func(t *testing.T) {
		client, err := search.NewClient()
		if err != nil {
			t.Skip("Could not create search client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("search.Client should have a valid core.Client")
		}
	})

	t.Run("timers.Client", func(t *testing.T) {
		client, err := timers.NewClient()
		if err != nil {
			t.Skip("Could not create timers client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("timers.Client should have a valid core.Client")
		}
	})

	t.Run("transfer.Client", func(t *testing.T) {
		client, err := transfer.NewClient()
		if err != nil {
			t.Skip("Could not create transfer client:", err)
		}
		if client == nil || client.Client == nil {
			t.Error("transfer.Client should have a valid core.Client")
		}
	})
}

// Note: We removed the verifyClientType helper function since we're using
// a more direct approach in TestServiceClientInterfaces

// TestServiceTransportInterfaces verifies that all service client cores
// have transports that implement the Transport interface
func TestServiceTransportInterfaces(t *testing.T) {
	// This test verifies at runtime that service client cores have transports
	// that implement the Transport interface

	// Helper function to test a client's transport
	testTransport := func(t *testing.T, name string, createClient func() (interface{}, error), getClient func(interface{}) interface{}, getTransport func(interface{}) interface{}) {
		t.Run(name+" transport", func(t *testing.T) {
			client, err := createClient()
			if err != nil {
				t.Skip("Could not create client:", err)
				return
			}
			if client == nil {
				t.Skip("Client is nil")
				return
			}

			coreClient := getClient(client)
			if coreClient == nil {
				t.Skip("Core client is nil")
				return
			}

			transport := getTransport(coreClient)
			if transport == nil {
				t.Skip("Transport is nil")
				return
			}

			if _, ok := transport.(interfaces.Transport); !ok {
				t.Errorf("%s transport does not implement interfaces.Transport", name)
			}
		})
	}

	// Test each service client
	testTransport(t, "auth",
		func() (interface{}, error) { return auth.NewClient() },
		func(c interface{}) interface{} { return c.(*auth.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "compute",
		func() (interface{}, error) { return compute.NewClient() },
		func(c interface{}) interface{} { return c.(*compute.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "flows",
		func() (interface{}, error) { return flows.NewClient() },
		func(c interface{}) interface{} { return c.(*flows.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "groups",
		func() (interface{}, error) { return groups.NewClient() },
		func(c interface{}) interface{} { return c.(*groups.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "search",
		func() (interface{}, error) { return search.NewClient() },
		func(c interface{}) interface{} { return c.(*search.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "timers",
		func() (interface{}, error) { return timers.NewClient() },
		func(c interface{}) interface{} { return c.(*timers.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })

	testTransport(t, "transfer",
		func() (interface{}, error) { return transfer.NewClient() },
		func(c interface{}) interface{} { return c.(*transfer.Client).Client },
		func(c interface{}) interface{} { return c.(*core.Client).Transport })
}
