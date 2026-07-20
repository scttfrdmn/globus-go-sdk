// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package compute provides a client for interacting with the Globus Compute service.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees.

The Globus Compute web service (as of Python globus-sdk 3.65.0) defines no
request/response models and no pagination. This client mirrors that: request
bodies and object responses are passthrough map[string]interface{} documents.
Two endpoints do not return JSON objects, so their methods return their real
shape — GetEndpoints returns []map[string]interface{} (a top-level array) and
GetVersion returns interface{} (a bare string, or an object when a service is
given).

The client folds the upstream ComputeClientV2 and ComputeClientV3 into one type;
the v3 methods carry a "V3" suffix (RegisterEndpointV3, UpdateEndpointV3,
LockEndpointV3, GetEndpointAllowlistV3, RegisterFunctionV3, SubmitV3).

# Compatibility Guarantees

  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices and
    maintained for at least one major release cycle
  - Breaking changes only occur in major version bumps (e.g., v3.x to v4.x)

# Synchronized Versioning

This package follows synchronized versioning with the Globus Python SDK to keep
API compatibility and feature parity across language implementations.

# Basic Usage

Create a new compute client:

	computeClient, err := compute.NewClient(
		compute.WithAuthorizer(authorizer),
	)
	if err != nil {
		// Handle error
	}

Service information:

	// GetVersion returns a bare string with no service argument, or an object
	// when a service is named.
	version, err := computeClient.GetVersion(ctx, "")
	if err != nil {
		// Handle error
	}
	fmt.Printf("Compute API version: %v\n", version)

Endpoints:

	// GetEndpoints returns a top-level array of endpoint documents.
	endpoints, err := computeClient.GetEndpoints(ctx, &compute.GetEndpointsOptions{Role: "owner"})
	if err != nil {
		// Handle error
	}
	for _, ep := range endpoints {
		fmt.Printf("Endpoint: %v (%v)\n", ep["name"], ep["uuid"])
	}

	// Register / inspect / delete an endpoint (passthrough documents).
	ep, err := computeClient.RegisterEndpoint(ctx, map[string]interface{}{
		"display_name": "my-endpoint",
	})
	status, err := computeClient.GetEndpointStatus(ctx, "endpoint-id")
	_, err = computeClient.DeleteEndpoint(ctx, "endpoint-id")

Functions:

	// Register a function. The document shape is defined by the Compute API; a
	// usable function requires the service's serialization envelope.
	fn, err := computeClient.RegisterFunction(ctx, map[string]interface{}{
		"function_name": "example",
		"function_code": "def example(x, y):\n    return x + y\n",
	})
	if err != nil {
		// Handle error
	}
	functionID, _ := fn["function_uuid"].(string)

	got, err := computeClient.GetFunction(ctx, functionID)
	_, err = computeClient.DeleteFunction(ctx, functionID)

Task submission and status:

	// Submit a task batch (POST /v2/submit). The document shape is defined by
	// the Compute API; keys map an endpoint ID to its list of tasks.
	result, err := computeClient.Submit(ctx, map[string]interface{}{
		"tasks": map[string]interface{}{},
	})
	if err != nil {
		// Handle error
	}

	// Poll individual tasks or a batch.
	task, err := computeClient.GetTask(ctx, "task-id")
	batch, err := computeClient.GetBatchStatus(ctx, []string{"t1", "t2"})

# V3 API

The v3 endpoint, function, and submit routes are available through the
V3-suffixed methods, e.g.:

	_, err := computeClient.SubmitV3(ctx, "endpoint-id", map[string]interface{}{
		"tasks": []interface{}{},
	})
*/
package compute
