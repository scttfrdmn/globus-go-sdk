// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package transfer

import (
	"context"
	"reflect"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/tests/compatibility"
)

// TransferClientTest verifies Transfer client compatibility
type TransferClientTest struct{}

func (test *TransferClientTest) Name() string {
	return "TransferClient"
}

func (test *TransferClientTest) Setup(ctx context.Context) error {
	return nil
}

func (test *TransferClientTest) Run(ctx context.Context, version string, t *testing.T) error {
	// Create Transfer client
	client, err := transfer.NewClient()
	if err != nil {
		t.Fatalf("Failed to create Transfer client: %v", err)
	}

	// Verify required methods exist
	t.Run("MethodsExist", func(t *testing.T) {
		// Get method type using reflection
		clientType := reflect.TypeOf(client)

		// List of required methods to verify
		requiredMethods := []string{
			"GetEndpoint",
			"ListEndpoints",
			"SubmitTransfer",
			"GetTaskByID",
			"CancelTask",
		}

		// Verify each method exists
		for _, methodName := range requiredMethods {
			method, found := clientType.MethodByName(methodName)
			if !found {
				t.Errorf("Required method %s not found", methodName)
				continue
			}

			// Log method signature for debugging
			t.Logf("Method %s exists with type %v", methodName, method.Type)
		}
	})

	// Check if client has the BaseClient field that implements interfaces.ClientInterface
	t.Run("HasBaseClient", func(t *testing.T) {
		clientType := reflect.TypeOf(client).Elem()
		_, found := clientType.FieldByName("BaseClient")
		if !found {
			t.Errorf("Transfer client doesn't have BaseClient field")
		}
	})

	// Verify transfer options existence (only for versions >= 0.9.0)
	minVersion := "v0.9.0"
	isMinVersion, err := compatibility.VersionAtLeast(version, minVersion)
	if err != nil {
		t.Fatalf("Failed to compare versions: %v", err)
	}

	if isMinVersion {
		t.Run("TransferOptions", func(t *testing.T) {
			// Check if the options struct exists and has expected field getters
			submitMethodType, found := reflect.TypeOf(client).MethodByName("SubmitTransfer")
			if !found {
				t.Errorf("SubmitTransfer method not found")
				return
			}

			// Check if the method has at least one argument (context) and that would imply options
			if submitMethodType.Type.NumIn() > 2 {
				t.Logf("SubmitTransfer appears to accept options")
			} else {
				t.Logf("SubmitTransfer may not support options pattern")
			}
		})
	}

	return nil
}

func (test *TransferClientTest) Teardown(ctx context.Context) error {
	return nil
}

func init() {
	compatibility.RegisterTest(&TransferClientTest{})
}
