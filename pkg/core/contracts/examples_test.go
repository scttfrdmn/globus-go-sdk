// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/contracts"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// ExampleClient demonstrates how to use the contract tests
// in your own implementation tests
type ExampleClient struct {
	baseURL   string
	userAgent string
	client    *http.Client
	logger    interfaces.Logger
}

func NewExampleClient() *ExampleClient {
	return &ExampleClient{
		baseURL:   "https://example.com",
		userAgent: "example-client/1.0",
		client:    &http.Client{},
		logger:    contracts.NewMockLogger(),
	}
}

func (c *ExampleClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// continue
	}

	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}, nil
}

func (c *ExampleClient) GetHTTPClient() *http.Client {
	return c.client
}

func (c *ExampleClient) GetBaseURL() string {
	return c.baseURL
}

func (c *ExampleClient) GetUserAgent() string {
	return c.userAgent
}

func (c *ExampleClient) GetLogger() interfaces.Logger {
	return c.logger
}

// TestExampleClientContract demonstrates how to use the contract tests
// in your own implementation tests
func TestExampleClientContract(t *testing.T) {
	client := NewExampleClient()

	// Verify that the ExampleClient satisfies the ClientInterface contract
	contracts.VerifyClientContract(t, client)

	// You can also add your own custom tests for implementation-specific behavior
	t.Run("Custom behavior", func(t *testing.T) {
		// Test implementation-specific behavior not covered by the contract
		if client.GetUserAgent() != "example-client/1.0" {
			t.Errorf("Expected user agent example-client/1.0, got %s", client.GetUserAgent())
		}
	})
}

// ExampleVerifyAllContracts demonstrates how to use VerifyAllContracts
// to test multiple implementations at once
func TestExampleVerifyAllContracts(t *testing.T) {
	client := NewExampleClient()
	logger := contracts.NewMockLogger()
	transport := contracts.NewMockTransport()

	// Create a map of implementations to test
	implementations := map[string]interface{}{
		"ExampleClient": client,
		"MockLogger":    logger,
		"MockTransport": transport,
	}

	// Verify all implementations at once
	contracts.VerifyAllContracts(t, implementations)
}

// ExampleCustomContract demonstrates how to create a custom contract test
// for your own interfaces
type CustomServiceInterface interface {
	Operation1(ctx context.Context, input string) (string, error)
	Operation2(ctx context.Context, input int) (int, error)
	GetStatus() string
}

// VerifyCustomServiceContract verifies that a CustomServiceInterface implementation
// satisfies the behavioral contract of the interface
func VerifyCustomServiceContract(t *testing.T, service CustomServiceInterface) {
	t.Helper()

	t.Run("Operation1", func(t *testing.T) {
		ctx := context.Background()
		result, err := service.Operation1(ctx, "test")
		if err != nil {
			t.Errorf("Operation1 failed: %v", err)
		}
		if !strings.Contains(result, "test") {
			t.Errorf("Operation1 result %q doesn't contain input 'test'", result)
		}

		// Test with canceled context
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = service.Operation1(canceledCtx, "test")
		if err == nil {
			t.Error("Operation1 with canceled context should fail")
		}
	})

	t.Run("Operation2", func(t *testing.T) {
		ctx := context.Background()
		result, err := service.Operation2(ctx, 5)
		if err != nil {
			t.Errorf("Operation2 failed: %v", err)
		}
		if result <= 0 {
			t.Errorf("Operation2 result %d should be positive", result)
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		status := service.GetStatus()
		if status == "" {
			t.Error("GetStatus returned empty string")
		}
	})
}

// MockCustomService is a custom service implementation for testing
type MockCustomService struct {
	status string
}

func NewMockCustomService() *MockCustomService {
	return &MockCustomService{
		status: "ready",
	}
}

func (s *MockCustomService) Operation1(ctx context.Context, input string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// continue
	}

	return "processed: " + input, nil
}

func (s *MockCustomService) Operation2(ctx context.Context, input int) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// continue
	}

	return input * 2, nil
}

func (s *MockCustomService) GetStatus() string {
	return s.status
}

// TestCustomServiceContract demonstrates how to use a custom contract test
func TestCustomServiceContract(t *testing.T) {
	service := NewMockCustomService()
	VerifyCustomServiceContract(t, service)
}
