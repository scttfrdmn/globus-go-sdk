// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/benchmark"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Parse command line flags
	srcEndpointID := flag.String("src", "", "Source endpoint ID")
	destEndpointID := flag.String("dest", "", "Destination endpoint ID")
	srcPath := flag.String("src-path", "~/", "Source path")
	destPath := flag.String("dest-path", "~/", "Destination path")
	fileSizeMB := flag.Float64("file-size", 10.0, "Size of each file in MB")
	fileCount := flag.Int("file-count", 10, "Number of files to transfer")
	parallelism := flag.Int("parallel", 4, "Transfer parallelism")
	useRecursive := flag.Bool("recursive", true, "Use recursive transfer")
	generateData := flag.Bool("generate", true, "Generate test data")
	deleteAfter := flag.Bool("delete", true, "Delete test data after benchmark")
	accessToken := flag.String("token", "", "Globus access token (if not provided, will use auth flow)")

	// Flags for specific benchmark suites
	runSize := flag.Bool("size", false, "Run file size benchmark suite")
	runParallelism := flag.Bool("parallelism-test", false, "Run parallelism benchmark suite")

	flag.Parse()

	// Check required parameters
	if *srcEndpointID == "" || *destEndpointID == "" {
		fmt.Println("Error: Source and destination endpoint IDs are required")
		flag.Usage()
		os.Exit(1)
	}

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get access token
	accessTokenStr := *accessToken
	if accessTokenStr == "" {
		token, err := getAccessToken()
		if err != nil {
			log.Fatalf("Error getting access token: %v", err)
		}
		accessTokenStr = token
	}

	// Create transfer client with proper authorizer
	authorizer := &simpleAuthorizer{token: accessTokenStr}
	client, err := transfer.NewClient(transfer.WithAuthorizer(authorizer))
	if err != nil {
		log.Fatalf("Error creating transfer client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel()

	// Create benchmark config
	config := &benchmark.TransferBenchmarkConfig{
		FileSizeMB:       *fileSizeMB,
		FileCount:        *fileCount,
		SourceEndpoint:   *srcEndpointID,
		DestEndpoint:     *destEndpointID,
		SourcePath:       *srcPath,
		DestPath:         *destPath,
		Parallelism:      *parallelism,
		UseRecursive:     *useRecursive,
		GenerateTestData: *generateData,
		DeleteAfter:      *deleteAfter,
	}

	// Run requested benchmark
	if *runSize {
		runFileSizeBenchmarkSuite(ctx, client, config)
	} else if *runParallelism {
		runParallelismBenchmarkSuite(ctx, client, config)
	} else {
		// Run single benchmark
		fmt.Printf("Running single transfer benchmark with:\n")
		fmt.Printf("  Source Endpoint:    %s\n", *srcEndpointID)
		fmt.Printf("  Destination Endpoint: %s\n", *destEndpointID)
		fmt.Printf("  File Size:          %.2f MB\n", *fileSizeMB)
		fmt.Printf("  File Count:         %d\n", *fileCount)
		fmt.Printf("  Total Size:         %.2f MB\n", *fileSizeMB*float64(*fileCount))
		fmt.Printf("  Parallelism:        %d\n", *parallelism)
		fmt.Printf("  Use Recursive:      %v\n", *useRecursive)
		fmt.Printf("  Generate Test Data: %v\n", *generateData)
		fmt.Printf("  Delete After:       %v\n", *deleteAfter)
		fmt.Println()

		// Start memory sampler
		memorySampler := benchmark.NewMemorySampler(500 * time.Millisecond)
		memorySampler.Start()
		defer memorySampler.Stop()

		// Run benchmark
		result, err := benchmark.BenchmarkTransfer(ctx, client, config, os.Stdout)
		if err != nil {
			log.Fatalf("Benchmark failed: %v", err)
		}

		// Update result with memory usage
		result.MemoryPeakMB = memorySampler.GetPeakMemory()

		// Print memory usage summary
		memorySampler.PrintSummary()
	}
}

// getAccessToken obtains a Globus access token using the auth flow
func getAccessToken() (string, error) {
	// Get client credentials from environment
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables are required")
	}

	// Create auth client
	authClient := auth.NewClient(clientID, clientSecret)
	authClient.SetRedirectURL("http://localhost:8080/callback")

	// Get authorization URL
	state := "benchmark-state"
	authURL := authClient.GetAuthorizationURL(
		state,
		transfer.TransferScope,
	)

	// Instruct user to visit the URL
	fmt.Printf("Please visit the following URL to authorize this application:\n\n%s\n\n", authURL)
	fmt.Printf("After authorization, you will be redirected to a localhost URL.\n")
	fmt.Printf("Copy the authorization code from the URL and paste it here: ")

	// Read authorization code
	var code string
	fmt.Scanln(&code)

	// Exchange code for tokens
	ctx := context.Background()
	tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

// runFileSizeBenchmarkSuite runs benchmarks with different file sizes
func runFileSizeBenchmarkSuite(ctx context.Context, client *transfer.Client, baseConfig *benchmark.TransferBenchmarkConfig) {
	fmt.Println("Running File Size Benchmark Suite")
	fmt.Println("=================================")

	// Define file size test cases
	testCases := []struct {
		name     string
		fileSize float64
		count    int
	}{
		{"Small Files", 1.0, 20},
		{"Medium Files", 10.0, 10},
		{"Large Files", 100.0, 2},
		{"Very Large File", 500.0, 1},
	}

	results := make([]*benchmark.BenchmarkResult, 0, len(testCases))

	for _, tc := range testCases {
		fmt.Printf("\n====== Benchmark: %s (%.1f MB x %d files) ======\n\n",
			tc.name, tc.fileSize, tc.count)

		config := *baseConfig
		config.FileSizeMB = tc.fileSize
		config.FileCount = tc.count

		// Start memory sampler
		memorySampler := benchmark.NewMemorySampler(500 * time.Millisecond)
		memorySampler.Start()

		// Run benchmark
		result, err := benchmark.BenchmarkTransfer(ctx, client, &config, os.Stdout)

		// Stop memory sampler
		memorySampler.Stop()
		memorySampler.PrintSummary()

		if err != nil {
			fmt.Printf("Error running benchmark %s: %v\n", tc.name, err)
			continue
		}

		// Update result with memory usage
		result.MemoryPeakMB = memorySampler.GetPeakMemory()
		results = append(results, result)

		// Add a small delay between benchmarks
		time.Sleep(5 * time.Second)
	}

	// Print comparison table
	fmt.Printf("\n====== File Size Benchmark Summary ======\n\n")
	fmt.Printf("| %-15s | %-10s | %-10s | %-15s | %-15s | %-15s |\n",
		"Benchmark", "Size/File", "Files", "Total Size", "Time", "Speed (MB/s)")
	fmt.Printf("|%-15s-|%-10s-|%-10s-|%-15s-|%-15s-|%-15s-|\n",
		"---------------", "----------", "----------", "---------------", "---------------", "---------------")

	for i, result := range results {
		tc := testCases[i]
		fmt.Printf("| %-15s | %-10.1f | %-10d | %-15.1f | %-15s | %-15.2f |\n",
			tc.name, tc.fileSize, tc.count, result.TotalSizeMB,
			result.ElapsedTime.Round(time.Millisecond), result.TransferSpeedMBs)
	}
}

// runParallelismBenchmarkSuite runs benchmarks with different parallelism settings
func runParallelismBenchmarkSuite(ctx context.Context, client *transfer.Client, baseConfig *benchmark.TransferBenchmarkConfig) {
	fmt.Println("Running Parallelism Benchmark Suite")
	fmt.Println("==================================")

	// Define parallelism test cases
	testCases := []struct {
		name        string
		parallelism int
	}{
		{"Sequential", 1},
		{"Low Parallelism", 2},
		{"Medium Parallelism", 4},
		{"High Parallelism", 8},
		{"Very High Parallelism", 16},
	}

	results := make([]*benchmark.BenchmarkResult, 0, len(testCases))

	// Use consistent file size for all tests
	baseConfig.FileSizeMB = 10.0
	baseConfig.FileCount = 10

	for _, tc := range testCases {
		fmt.Printf("\n====== Benchmark: %s (Parallelism: %d) ======\n\n",
			tc.name, tc.parallelism)

		config := *baseConfig
		config.Parallelism = tc.parallelism

		// Start memory sampler
		memorySampler := benchmark.NewMemorySampler(500 * time.Millisecond)
		memorySampler.Start()

		// Run benchmark
		result, err := benchmark.BenchmarkTransfer(ctx, client, &config, os.Stdout)

		// Stop memory sampler
		memorySampler.Stop()
		memorySampler.PrintSummary()

		if err != nil {
			fmt.Printf("Error running benchmark %s: %v\n", tc.name, err)
			continue
		}

		// Update result with memory usage
		result.MemoryPeakMB = memorySampler.GetPeakMemory()
		results = append(results, result)

		// Add a small delay between benchmarks
		time.Sleep(5 * time.Second)
	}

	// Print comparison table
	fmt.Printf("\n====== Parallelism Benchmark Summary ======\n\n")
	fmt.Printf("| %-20s | %-12s | %-15s | %-15s | %-15s |\n",
		"Benchmark", "Parallelism", "Time", "Speed (MB/s)", "Memory (MB)")
	fmt.Printf("|%-20s-|%-12s-|%-15s-|%-15s-|%-15s-|\n",
		"--------------------", "------------", "---------------", "---------------", "---------------")

	for i, result := range results {
		tc := testCases[i]
		fmt.Printf("| %-20s | %-12d | %-15s | %-15.2f | %-15.2f |\n",
			tc.name, tc.parallelism,
			result.ElapsedTime.Round(time.Millisecond),
			result.TransferSpeedMBs, result.MemoryPeakMB)
	}
}

// simpleAuthorizer is a simple implementation of the auth.Authorizer interface
type simpleAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *simpleAuthorizer) GetAuthorizationHeader(_ ...context.Context) (string, error) {
	if a.token == "" {
		return "", nil
	}
	return "Bearer " + a.token, nil
}
