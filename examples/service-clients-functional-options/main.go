// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("=== Globus Go SDK v0.9.0 Service Clients with Functional Options ===")
	fmt.Printf("SDK Version: %s\n\n", pkg.Version)

	// Check for Globus credentials
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")

	if clientID == "" || clientSecret == "" {
		fmt.Println("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables must be set")
		fmt.Println("This example will only demonstrate client creation, not actual API calls")
	}

	if accessToken == "" {
		fmt.Println("GLOBUS_ACCESS_TOKEN environment variable not set")
		fmt.Println("Using a placeholder access token for demonstration")
		accessToken = "placeholder-access-token"
	}

	// Enable HTTP debugging if requested
	enableDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"
	enableTracing := os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1"

	// Create SDK config
	config := pkg.NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Example 1: Create Auth client with functional options
	fmt.Println("1. Creating Auth client with functional options")
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
		auth.WithHTTPDebugging(enableDebug),
		auth.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Printf("Auth client created successfully: %T\n", authClient)

	// Example 2: Create Auth client using SDK config
	fmt.Println("\n2. Creating Auth client using SDK config")
	authClient2, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Printf("Auth client created successfully via SDK config: %T\n", authClient2)

	// Example 3: Create Flows client with functional options
	fmt.Println("\n3. Creating Flows client with functional options")
	flowsClient, err := flows.NewClient(
		flows.WithAccessToken(accessToken),
		flows.WithHTTPDebugging(enableDebug),
		flows.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create flows client: %v", err)
	}

	fmt.Printf("Flows client created successfully: %T\n", flowsClient)

	// Example 4: Create Flows client using SDK config
	fmt.Println("\n4. Creating Flows client using SDK config")
	flowsClient2, err := config.NewFlowsClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create flows client: %v", err)
	}

	fmt.Printf("Flows client created successfully via SDK config: %T\n", flowsClient2)

	// Example 5: Create Transfer client with functional options
	fmt.Println("\n5. Creating Transfer client with functional options")

	// Create an authorizer from the access token
	authorizer := authorizers.StaticTokenAuthorizerWithCoreAuthorizer(accessToken)

	transferClient, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizer),
		transfer.WithHTTPDebugging(enableDebug),
		transfer.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create transfer client: %v", err)
	}

	fmt.Printf("Transfer client created successfully: %T\n", transferClient)

	// Example 6: Create Transfer client using SDK config
	fmt.Println("\n6. Creating Transfer client using SDK config")
	transferClient2, err := config.NewTransferClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create transfer client: %v", err)
	}

	fmt.Printf("Transfer client created successfully via SDK config: %T\n", transferClient2)

	// Example 7: Create Search client with functional options
	fmt.Println("\n7. Creating Search client with functional options")
	searchClient, err := search.NewClient(
		search.WithAccessToken(accessToken),
		search.WithHTTPDebugging(enableDebug),
		search.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create search client: %v", err)
	}

	fmt.Printf("Search client created successfully: %T\n", searchClient)

	// Example 8: Create Search client using SDK config
	fmt.Println("\n8. Creating Search client using SDK config")
	searchClient2, err := config.NewSearchClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create search client: %v", err)
	}

	fmt.Printf("Search client created successfully via SDK config: %T\n", searchClient2)

	// Example 9: Create Groups client with functional options
	fmt.Println("\n9. Creating Groups client with functional options")
	groupsClient, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithHTTPDebugging(enableDebug),
		groups.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create groups client: %v", err)
	}

	fmt.Printf("Groups client created successfully: %T\n", groupsClient)

	// Example 10: Create Groups client using SDK config
	fmt.Println("\n10. Creating Groups client using SDK config")
	groupsClient2, err := config.NewGroupsClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create groups client: %v", err)
	}

	fmt.Printf("Groups client created successfully via SDK config: %T\n", groupsClient2)

	// Example 11: Create Compute client with functional options
	fmt.Println("\n11. Creating Compute client with functional options")
	computeClient, err := compute.NewClient(
		compute.WithAccessToken(accessToken),
		compute.WithHTTPDebugging(enableDebug),
		compute.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}

	fmt.Printf("Compute client created successfully: %T\n", computeClient)

	// Example 12: Create Compute client using SDK config
	fmt.Println("\n12. Creating Compute client using SDK config")
	computeClient2, err := config.NewComputeClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}

	fmt.Printf("Compute client created successfully via SDK config: %T\n", computeClient2)

	// Example 13: Create Timers client with functional options
	fmt.Println("\n13. Creating Timers client with functional options")
	timersClient, err := timers.NewClient(
		timers.WithAccessToken(accessToken),
		timers.WithHTTPDebugging(enableDebug),
		timers.WithHTTPTracing(enableTracing),
	)
	if err != nil {
		log.Fatalf("Failed to create timers client: %v", err)
	}

	fmt.Printf("Timers client created successfully: %T\n", timersClient)

	// Example 14: Create Timers client using SDK config
	fmt.Println("\n14. Creating Timers client using SDK config")
	timersClient2, err := config.NewTimersClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create timers client: %v", err)
	}

	fmt.Printf("Timers client created successfully via SDK config: %T\n", timersClient2)

	// Example 15: Create Token Manager with functional options
	fmt.Println("\n15. Creating Token Manager with functional options")
	storage := tokens.NewMemoryStorage()

	tokenManager, err := tokens.NewManager(
		tokens.WithStorage(storage),
		tokens.WithAuthClient(authClient),
		tokens.WithRefreshThreshold(15*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created successfully: %T\n", tokenManager)

	// Example 16: Create Token Manager using SDK config
	fmt.Println("\n16. Creating Token Manager using SDK config")
	tokenManager2, err := config.NewTokenManager(
		tokens.WithStorage(storage),
		tokens.WithAuthClient(authClient),
	)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created successfully via SDK config: %T\n", tokenManager2)

	// Example 17: Advanced configuration with core client options
	fmt.Println("\n17. Advanced configuration with core client options")

	// Create custom core client options
	coreOptions := []core.ClientOption{
		core.WithAuthorizer(authorizer),
		core.WithUserAgent("globus-go-sdk-example/1.0"),
		core.WithRequestTimeout(30 * time.Second),
	}

	// Apply core options to flows client
	advancedFlowsClient, err := flows.NewClient(
		flows.WithCoreOptions(coreOptions...),
		flows.WithHTTPDebugging(enableDebug),
	)
	if err != nil {
		log.Fatalf("Failed to create advanced flows client: %v", err)
	}

	fmt.Printf("Advanced flows client created successfully: %T\n", advancedFlowsClient)

	fmt.Println("\n=== Functional Options Pattern Demonstration Complete ===")
	fmt.Println("All service clients in the Globus Go SDK now use a consistent")
	fmt.Println("functional options pattern for configuration and initialization.")
}
