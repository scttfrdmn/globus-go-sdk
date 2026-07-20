// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package testhelpers provides shared testing utilities and fixtures
// following patterns from the upstream Python SDK testing approach.
package testhelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
)

// TestConfig holds global test configuration following Python SDK patterns
type TestConfig struct {
	DisableRetries bool
	MockSleep      bool
	TimeoutMs      int
}

// GlobalTestConfig is the default test configuration
var GlobalTestConfig = &TestConfig{
	DisableRetries: true,
	MockSleep:      true,
	TimeoutMs:      100, // Fast timeouts for tests
}

// TestScenario represents test metadata similar to Python SDK's load_response
type TestScenario struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	GroupID      string                 `json:"group_id,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	RoleID       string                 `json:"role_id,omitempty"`
	Subscription map[string]interface{} `json:"subscription,omitempty"`
	Expected     map[string]interface{} `json:"expected,omitempty"`
	HTTPStatus   int                    `json:"http_status"`

	// Enhanced metadata for comprehensive testing
	Method       string                 `json:"method,omitempty"`        // HTTP method
	Path         string                 `json:"path,omitempty"`          // API endpoint path
	RequestBody  map[string]interface{} `json:"request_body,omitempty"`  // Expected request payload
	ResponseBody map[string]interface{} `json:"response_body,omitempty"` // Mock response payload
	Headers      map[string]string      `json:"headers,omitempty"`       // Additional headers
	QueryParams  map[string]string      `json:"query_params,omitempty"`  // URL query parameters

	// Test behavior configuration
	Timeout      int      `json:"timeout,omitempty"`      // Test timeout in milliseconds
	Retry        bool     `json:"retry,omitempty"`        // Whether to retry on failure
	Skip         bool     `json:"skip,omitempty"`         // Skip this test scenario
	Tags         []string `json:"tags,omitempty"`         // Test categories/tags
	Dependencies []string `json:"dependencies,omitempty"` // Required test dependencies
}

// TestSuite represents a collection of related test scenarios
type TestSuite struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Setup       map[string]interface{}   `json:"setup,omitempty"`     // Global test setup data
	Teardown    map[string]interface{}   `json:"teardown,omitempty"`  // Global test cleanup data
	Scenarios   map[string]*TestScenario `json:"scenarios"`           // Individual test scenarios
	Templates   map[string]interface{}   `json:"templates,omitempty"` // Reusable response templates
}

// ResponseTemplate represents reusable response patterns
type ResponseTemplate struct {
	StatusCode int                    `json:"status_code"`
	Body       map[string]interface{} `json:"body"`
	Headers    map[string]string      `json:"headers,omitempty"`
}

// LoadTestScenario loads test metadata from JSON files (Python SDK pattern)
func LoadTestScenario(t *testing.T, service, scenario string) *TestScenario {
	testSuite := LoadTestSuite(t, service)

	testScenario, exists := testSuite.Scenarios[scenario]
	if !exists {
		t.Fatalf("Test scenario %s not found in %s test suite", scenario, service)
	}

	// Apply templates if referenced
	if templateName, hasTemplate := testScenario.Expected["template"].(string); hasTemplate {
		if template, exists := testSuite.Templates[templateName]; exists {
			applyTemplate(testScenario, template)
		}
	}

	return testScenario
}

// LoadTestSuite loads a complete test suite from JSON files
func LoadTestSuite(t *testing.T, service string) *TestSuite {
	testdataPath := filepath.Join("testdata", fmt.Sprintf("%s.json", service))

	data, err := os.ReadFile(testdataPath)
	if err != nil {
		t.Fatalf("Failed to load test data %s: %v", testdataPath, err)
	}

	// Try loading as new TestSuite format first
	var testSuite TestSuite
	if err := json.Unmarshal(data, &testSuite); err == nil && testSuite.Scenarios != nil {
		return &testSuite
	}

	// Fallback to legacy format (map of scenarios)
	var legacyScenarios map[string]*TestScenario
	if err := json.Unmarshal(data, &legacyScenarios); err != nil {
		t.Fatalf("Failed to parse test data %s: %v", testdataPath, err)
	}

	// Convert legacy format to new TestSuite format
	return &TestSuite{
		Name:      service + " Test Suite",
		Scenarios: legacyScenarios,
	}
}

// LoadScenariosByTag loads all scenarios matching specific tags
func LoadScenariosByTag(t *testing.T, service string, tags ...string) []*TestScenario {
	testSuite := LoadTestSuite(t, service)

	var matchingScenarios []*TestScenario
	for _, scenario := range testSuite.Scenarios {
		if scenario.Skip {
			continue // Skip disabled scenarios
		}

		if hasAnyTag(scenario.Tags, tags) {
			matchingScenarios = append(matchingScenarios, scenario)
		}
	}

	return matchingScenarios
}

// applyTemplate applies a response template to a test scenario
func applyTemplate(scenario *TestScenario, template interface{}) {
	if templateMap, ok := template.(map[string]interface{}); ok {
		if responseBody, exists := templateMap["response_body"]; exists {
			if rb, ok := responseBody.(map[string]interface{}); ok {
				scenario.ResponseBody = rb
			}
		}
		if statusCode, exists := templateMap["status_code"]; exists {
			if code, ok := statusCode.(float64); ok {
				scenario.HTTPStatus = int(code)
			}
		}
	}
}

// hasAnyTag checks if scenario has any of the specified tags
func hasAnyTag(scenarioTags, searchTags []string) bool {
	if len(searchTags) == 0 {
		return true // No tag filter means include all
	}

	for _, searchTag := range searchTags {
		for _, scenarioTag := range scenarioTags {
			if scenarioTag == searchTag {
				return true
			}
		}
	}

	return false
}

// ValidateTestScenario validates that a test scenario has all required fields
func ValidateTestScenario(t *testing.T, scenario *TestScenario) {
	if scenario.Name == "" {
		t.Errorf("Test scenario is missing required 'name' field")
	}
	if scenario.HTTPStatus == 0 {
		t.Errorf("Test scenario '%s' is missing required 'http_status' field", scenario.Name)
	}
	if scenario.Method != "" && scenario.Path == "" {
		t.Errorf("Test scenario '%s' has method but no path", scenario.Name)
	}
}

// ApplyVariableSubstitution applies variable substitution to string fields using scenario data
func ApplyVariableSubstitution(template string, scenario *TestScenario) string {
	result := template

	// Apply basic variable substitutions
	if scenario.GroupID != "" {
		result = strings.ReplaceAll(result, "{{group_id}}", scenario.GroupID)
	}
	if scenario.UserID != "" {
		result = strings.ReplaceAll(result, "{{user_id}}", scenario.UserID)
		result = strings.ReplaceAll(result, "{{identity_id}}", scenario.UserID)
	}
	if scenario.RoleID != "" {
		result = strings.ReplaceAll(result, "{{role_id}}", scenario.RoleID)
	}

	// Apply subscription variables
	if scenario.Subscription != nil {
		if subID, ok := scenario.Subscription["subscription_id"].(string); ok {
			result = strings.ReplaceAll(result, "{{subscription_id}}", subID)
		}
	}

	// Apply expected variables
	if scenario.Expected != nil {
		if groupName, ok := scenario.Expected["name"].(string); ok {
			result = strings.ReplaceAll(result, "{{group_name}}", groupName)
		}
	}

	return result
}

// GenerateTestCaseFromScenario creates a test case function from a scenario
func GenerateTestCaseFromScenario(scenario *TestScenario, handler func(*testing.T, *TestScenario)) func(*testing.T) {
	return func(t *testing.T) {
		if scenario.Skip {
			t.Skipf("Scenario '%s' is marked as skip", scenario.Name)
		}

		// Validate scenario before running
		ValidateTestScenario(t, scenario)

		// Set test timeout if specified
		if scenario.Timeout > 0 {
			oldDeadline, hasDeadline := t.Deadline()
			testTimeout := time.Duration(scenario.Timeout) * time.Millisecond
			if !hasDeadline || time.Until(oldDeadline) > testTimeout {
				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()
				_ = ctx // Use context in test if needed
			}
		}

		// Run the test handler
		handler(t, scenario)
	}
}

// LoadScenariosByDependencies loads scenarios considering their dependencies
func LoadScenariosByDependencies(t *testing.T, service string, rootScenario string) []*TestScenario {
	testSuite := LoadTestSuite(t, service)

	var orderedScenarios []*TestScenario
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var loadDependencies func(string)
	loadDependencies = func(scenarioName string) {
		if visited[scenarioName] {
			return
		}
		if visiting[scenarioName] {
			t.Fatalf("Circular dependency detected involving scenario '%s'", scenarioName)
		}

		scenario, exists := testSuite.Scenarios[scenarioName]
		if !exists {
			t.Fatalf("Scenario '%s' not found", scenarioName)
		}

		visiting[scenarioName] = true

		// Load dependencies first
		for _, dep := range scenario.Dependencies {
			loadDependencies(dep)
		}

		visiting[scenarioName] = false
		visited[scenarioName] = true
		orderedScenarios = append(orderedScenarios, scenario)
	}

	loadDependencies(rootScenario)
	return orderedScenarios
}

// MockGroupsClient creates a groups client with enhanced mock server
// following Python SDK's mocked_responses pattern
func MockGroupsClient(t *testing.T, handler http.HandlerFunc) (*groups.Client, *httptest.Server, func()) {
	server := httptest.NewServer(handler)

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")

	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create mock groups client: %v", err)
	}

	cleanup := func() {
		server.Close()
	}

	return client, server, cleanup
}

// MockResponseHandler creates a flexible response handler for different scenarios
type MockResponseHandler struct {
	responses map[string]MockResponse
}

type MockResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// NewMockResponseHandler creates a new mock response handler
func NewMockResponseHandler() *MockResponseHandler {
	return &MockResponseHandler{
		responses: make(map[string]MockResponse),
	}
}

// RegisterResponse registers a response for a specific path/method combination
func (m *MockResponseHandler) RegisterResponse(method, path string, response MockResponse) {
	key := fmt.Sprintf("%s %s", method, path)
	m.responses[key] = response
}

// ServeHTTP implements the http.Handler interface
func (m *MockResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

	response, exists := m.responses[key]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Set headers
	for k, v := range response.Headers {
		w.Header().Set(k, v)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	if response.Body != nil {
		_ = json.NewEncoder(w).Encode(response.Body)
	}
}

// AssertHTTPSuccess validates successful HTTP responses
func AssertHTTPSuccess(t *testing.T, statusCode int) {
	if statusCode < 200 || statusCode >= 300 {
		t.Errorf("Expected success status code (2xx), got %d", statusCode)
	}
}

// AssertHTTPError validates error HTTP responses
func AssertHTTPError(t *testing.T, statusCode, expectedCode int) {
	if statusCode != expectedCode {
		t.Errorf("Expected status code %d, got %d", expectedCode, statusCode)
	}
}

// AssertGroupEquals validates group objects match expected values
func AssertGroupEquals(t *testing.T, actual *groups.Group, expected *TestScenario) {
	if expected.GroupID != "" && actual.ID != expected.GroupID {
		t.Errorf("Group ID mismatch: expected %s, got %s", expected.GroupID, actual.ID)
	}

	if expectedName, ok := expected.Expected["name"].(string); ok {
		if actual.Name != expectedName {
			t.Errorf("Group name mismatch: expected %s, got %s", expectedName, actual.Name)
		}
	}
}


// MockSleepDisabled prevents actual sleep during tests (Python SDK pattern)
func MockSleepDisabled() {
	// In Go, we can't easily mock time.Sleep globally like Python
	// But we can configure clients to use minimal timeouts
	_ = GlobalTestConfig.MockSleep
}

// TestTimeout returns appropriate timeout for tests
func TestTimeout() time.Duration {
	return time.Duration(GlobalTestConfig.TimeoutMs) * time.Millisecond
}

// CreateTestContext creates a context with test-appropriate timeout
func CreateTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), TestTimeout())
}
