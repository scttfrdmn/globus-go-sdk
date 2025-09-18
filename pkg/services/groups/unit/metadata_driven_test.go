// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package groups

import (
	"context"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/testhelpers"
)

// TestMetadataDrivenScenarios tests scenarios using the enhanced metadata-driven system
func TestMetadataDrivenScenarios(t *testing.T) {
	// Test loading enhanced test suite format
	testSuite := testhelpers.LoadTestSuite(t, "groups_enhanced")

	if testSuite.Name == "" {
		t.Errorf("Test suite should have a name")
	}

	if len(testSuite.Scenarios) == 0 {
		t.Errorf("Test suite should have scenarios")
	}

	t.Logf("Loaded test suite '%s' with %d scenarios", testSuite.Name, len(testSuite.Scenarios))
}

// TestScenarioValidation tests the scenario validation functionality
func TestScenarioValidation(t *testing.T) {
	validScenario := &testhelpers.TestScenario{
		Name:       "Valid test scenario",
		HTTPStatus: 200,
		Method:     "GET",
		Path:       "groups/test",
	}

	// This should not fail
	testhelpers.ValidateTestScenario(t, validScenario)

	// Test validation with logging - validation system is working correctly
	t.Logf("Scenario validation functionality is working as expected")
}

// TestVariableSubstitution tests the variable substitution functionality
func TestVariableSubstitution(t *testing.T) {
	scenario := &testhelpers.TestScenario{
		GroupID: "test-group-123",
		UserID:  "test-user-456",
		RoleID:  "admin",
		Subscription: map[string]interface{}{
			"subscription_id": "sub-789",
		},
		Expected: map[string]interface{}{
			"name": "Test Group",
		},
	}

	template := "groups/{{group_id}}/members/{{user_id}}?role={{role_id}}&sub={{subscription_id}}&name={{group_name}}"
	result := testhelpers.ApplyVariableSubstitution(template, scenario)

	expected := "groups/test-group-123/members/test-user-456?role=admin&sub=sub-789&name=Test Group"
	if result != expected {
		t.Errorf("Variable substitution failed.\nExpected: %s\nGot: %s", expected, result)
	}
}

// TestTagBasedScenarioLoading tests loading scenarios by tags
func TestTagBasedScenarioLoading(t *testing.T) {
	// Load scenarios with specific tags
	basicScenarios := testhelpers.LoadScenariosByTag(t, "groups_enhanced", "basic")
	errorScenarios := testhelpers.LoadScenariosByTag(t, "groups_enhanced", "error")
	workflowScenarios := testhelpers.LoadScenariosByTag(t, "groups_enhanced", "workflow")

	t.Logf("Found %d basic scenarios", len(basicScenarios))
	t.Logf("Found %d error scenarios", len(errorScenarios))
	t.Logf("Found %d workflow scenarios", len(workflowScenarios))

	// Verify that scenarios are correctly tagged
	for _, scenario := range basicScenarios {
		if !containsTag(scenario.Tags, "basic") {
			t.Errorf("Scenario '%s' should have 'basic' tag", scenario.Name)
		}
	}

	// Load all scenarios (no tag filter)
	allScenarios := testhelpers.LoadScenariosByTag(t, "groups_enhanced")
	if len(allScenarios) == 0 {
		t.Errorf("Should load all scenarios when no tags specified")
	}
}

// TestTemplateApplication tests template application functionality
func TestTemplateApplication(t *testing.T) {
	// Load a scenario that uses templates
	scenario := testhelpers.LoadTestScenario(t, "groups_enhanced", "group_get_success")

	// Verify that template was applied (checking for template effects)
	if scenario.HTTPStatus != 200 {
		t.Errorf("Template should set status code to 200, got %d", scenario.HTTPStatus)
	}
}

// TestGeneratedTestCases tests the test case generation functionality
func TestGeneratedTestCases(t *testing.T) {
	scenario := testhelpers.LoadTestScenario(t, "groups_enhanced", "group_get_success")

	// Create a test handler
	var handlerCalled bool
	handler := func(t *testing.T, s *testhelpers.TestScenario) {
		handlerCalled = true
		if s.Name != scenario.Name {
			t.Errorf("Handler received wrong scenario: %s", s.Name)
		}
	}

	// Generate and run test case
	testCase := testhelpers.GenerateTestCaseFromScenario(scenario, handler)
	testCase(t)

	if !handlerCalled {
		t.Errorf("Test handler should have been called")
	}
}

// TestDependencyResolution tests dependency resolution for scenarios
func TestDependencyResolution(t *testing.T) {
	// Test loading scenarios with dependencies
	scenarios := testhelpers.LoadScenariosByDependencies(t, "groups_enhanced", "complex_workflow_full_lifecycle")

	if len(scenarios) == 0 {
		t.Errorf("Should load at least one scenario")
	}

	// The target scenario should be included
	found := false
	for _, scenario := range scenarios {
		if scenario.Name == "Full group lifecycle workflow" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Target scenario should be included in dependency resolution")
	}
}

// TestEnhancedScenarioExecution runs a complete enhanced scenario test
func TestEnhancedScenarioExecution(t *testing.T) {
	scenario := testhelpers.LoadTestScenario(t, "groups_enhanced", "group_get_success")

	// Create mock response handler
	mockHandler := testhelpers.NewMockResponseHandler()

	// Apply variable substitution to path
	path := testhelpers.ApplyVariableSubstitution(scenario.Path, scenario)

	// Register mock response
	mockHandler.RegisterResponse(scenario.Method, "/"+path, testhelpers.MockResponse{
		StatusCode: scenario.HTTPStatus,
		Body: map[string]interface{}{
			"DATA_TYPE": "group",
			"id":        scenario.GroupID,
			"name":      "Test Group",
		},
	})

	// Create mock client
	client, server, cleanup := testhelpers.MockGroupsClient(t, mockHandler.ServeHTTP)
	defer cleanup()

	// Execute the test
	ctx := context.Background()
	group, err := client.GetGroup(ctx, scenario.GroupID)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if group.ID != scenario.GroupID {
		t.Errorf("Expected group ID %s, got %s", scenario.GroupID, group.ID)
	}

	t.Logf("✅ Enhanced scenario '%s' executed successfully", scenario.Name)
	t.Logf("   Method: %s, Path: %s, Status: %d", scenario.Method, path, scenario.HTTPStatus)

	_ = server // Avoid unused variable warning
}

// Helper functions

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}
