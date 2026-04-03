// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package main demonstrates the unified systems in Globus Go SDK v3.60.0
//
// This example shows how to use the new unified error handling,
// response wrappers, and deprecation system introduced in v3.60.0.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/client"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/deprecation"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/response"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

func main() {
	// Enable deprecation warnings for demonstration
	deprecation.Enable()

	fmt.Println("=== Globus Go SDK v3.60.0 - Unified Systems Demo ===")
	fmt.Println()

	// Demonstrate deprecation warnings
	fmt.Println("1. Deprecation System Demo:")
	demonstrateDeprecationWarnings()
	fmt.Println()

	// Demonstrate unified client configuration
	fmt.Println("2. Unified Client Configuration Demo:")
	demonstrateUnifiedClientConfig()
	fmt.Println()

	// Demonstrate unified error handling
	fmt.Println("3. Unified Error Handling Demo:")
	demonstrateUnifiedErrorHandling()
	fmt.Println()

	// Demonstrate unified response system
	fmt.Println("4. Unified Response System Demo:")
	demonstrateUnifiedResponseSystem()
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("All unified systems working correctly!")
}

func demonstrateDeprecationWarnings() {
	// This will trigger a deprecation warning
	fmt.Println("   Calling deprecated method...")

	// Create a client and call the deprecated method
	authClient, err := auth.NewClient(
		auth.WithClientID("demo-client"),
		auth.WithClientSecret("demo-secret"),
	)
	if err != nil {
		log.Printf("Failed to create auth client: %v", err)
		return
	}

	// This will trigger a deprecation warning
	authClient.SetRedirectURL("https://example.com/callback")

	fmt.Println("   ✓ Deprecation warning issued")
}

func demonstrateUnifiedClientConfig() {
	// Demonstrate new unified client configuration
	fmt.Println("   Creating clients with unified configuration...")

	// Auth client with unified config
	authConfig, err := client.AuthConfig(
		client.WithAccessToken("demo-token"),
		client.WithTimeout(30*time.Second),
		client.WithMaxRetries(3),
	)
	if err != nil {
		log.Printf("Failed to create auth config: %v", err)
		return
	}

	fmt.Printf("   ✓ Auth config created: %s\n", authConfig.BaseURL)

	// Transfer client with unified config
	transferConfig, err := client.TransferConfig(
		client.WithAccessToken("demo-token"),
		client.WithTimeout(60*time.Second),
		client.WithMaxRetries(5),
	)
	if err != nil {
		log.Printf("Failed to create transfer config: %v", err)
		return
	}

	fmt.Printf("   ✓ Transfer config created: %s\n", transferConfig.BaseURL)

	// Groups client with unified config
	groupsConfig, err := client.GroupsConfig(
		client.WithAccessToken("demo-token"),
		client.WithDebug(true),
	)
	if err != nil {
		log.Printf("Failed to create groups config: %v", err)
		return
	}

	fmt.Printf("   ✓ Groups config created: %s\n", groupsConfig.BaseURL)

	fmt.Println("   ✓ All clients created with unified configuration")
}

func demonstrateUnifiedErrorHandling() {
	// Demonstrate unified error handling
	fmt.Println("   Creating and handling unified errors...")

	// Create different types of errors
	authError := errors.NewAuthError("invalid_token", "The provided token is invalid")
	transferError := errors.NewTransferError("TaskNotFound", "Transfer task not found")
	groupsError := errors.NewGroupsError("GroupNotFound", "Group not found")

	// Add context to errors
	authError.WithContext("endpoint", "/oauth2/token").WithRequestID("req-123")
	transferError.WithContext("task_id", "abc-123").WithDetail("Task was canceled")
	groupsError.WithContext("group_id", "group-456").WithRequestID("req-456")

	// Handle errors consistently
	handleError(authError, "auth operation")
	handleError(transferError, "transfer operation")
	handleError(groupsError, "groups operation")

	fmt.Println("   ✓ All errors handled consistently")
}

func handleError(err error, operation string) {
	globusErr, ok := err.(*errors.GlobusError)
	if !ok {
		fmt.Printf("   - %s: Non-Globus error: %v\n", operation, err)
		return
	}

	fmt.Printf("   - %s: [%s] %s", operation, globusErr.Service, globusErr.Code)
	if globusErr.RequestID != "" {
		fmt.Printf(" (Request: %s)", globusErr.RequestID)
	}
	fmt.Println()

	// Check error type
	if globusErr.IsAuthenticationError() {
		fmt.Printf("     → Authentication error detected\n")
	} else if globusErr.IsNotFoundError() {
		fmt.Printf("     → Not found error detected\n")
	} else if globusErr.IsRetryable() {
		fmt.Printf("     → Retryable error detected\n")
	}
}

func demonstrateUnifiedResponseSystem() {
	// Demonstrate unified response system
	fmt.Println("   Creating unified responses...")

	// Create sample data
	tokenData := auth.TokenResponse{
		AccessToken: "sample-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "openid profile email",
	}

	// Create unified response
	authResponse := response.NewAuthResponse(tokenData)
	authResponse.WithRequestID("req-789")

	// Add metadata
	metadata := response.ResponseMetadata{
		Service:    "auth",
		APIVersion: "v2",
		HTTPStatus: 200,
		Timestamp:  time.Now(),
	}
	authResponse.WithMetadata(metadata)

	fmt.Printf("   ✓ Auth response created with request ID: %s\n", authResponse.RequestID)
	fmt.Printf("   ✓ Service: %s, API Version: %s\n",
		authResponse.Metadata.Service, authResponse.Metadata.APIVersion)

	// Create paginated response
	sampleData := []string{"item1", "item2", "item3"}
	paginatedResponse := response.NewPaginatedResponse(sampleData, "transfer")
	paginatedResponse.WithRequestID("req-101")

	// Add pagination info
	paginationInfo := response.PaginationInfo{
		HasMore:   true,
		NextToken: "next-page-token",
		Limit:     10,
		Total:     25,
		PageSize:  3,
	}
	paginatedResponse.WithPagination(paginationInfo)

	fmt.Printf("   ✓ Paginated response created with %d items\n", len(paginatedResponse.Data))
	fmt.Printf("   ✓ Has more pages: %v, Next token: %s\n",
		paginatedResponse.Pagination.HasMore, paginatedResponse.Pagination.NextToken)

	// Create iterator from paginated response
	iterator := response.NewIterator(paginatedResponse, func(nextToken string) (*response.PaginatedResponse[string], error) {
		// This would normally fetch the next page
		return &response.PaginatedResponse[string]{
			Data: []string{"item4", "item5"},
			Pagination: response.PaginationInfo{
				HasMore:   false,
				NextToken: "",
				PageSize:  2,
			},
		}, nil
	})

	fmt.Printf("   ✓ Iterator created for paginated data\n")

	// Demonstrate iterator usage
	count := 0
	for {
		item, ok := iterator.Next()
		if !ok {
			break
		}
		count++
		fmt.Printf("     - Item %d: %s\n", count, item)
	}

	fmt.Printf("   ✓ Iterator processed %d items\n", count)

	if err := iterator.Error(); err != nil {
		fmt.Printf("   ✗ Iterator error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Iterator completed successfully\n")
	}
}
