// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestParseErrorResponse(t *testing.T) {
	// Test flow not found error
	flowNotFoundJSON := `{"code":"NotFound","message":"Flow not found","request_id":"123456"}`
	err := ParseErrorResponse([]byte(flowNotFoundJSON), http.StatusNotFound, "flow-123", "flow")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	flowNotFoundErr, ok := err.(*FlowNotFoundError)
	if !ok {
		t.Errorf("Expected FlowNotFoundError, got %T", err)
	} else {
		if flowNotFoundErr.FlowID != "flow-123" {
			t.Errorf("Expected flow ID flow-123, got %s", flowNotFoundErr.FlowID)
		}
		if flowNotFoundErr.Code != "NotFound" {
			t.Errorf("Expected error code NotFound, got %s", flowNotFoundErr.Code)
		}
		if flowNotFoundErr.RequestID != "123456" {
			t.Errorf("Expected request ID 123456, got %s", flowNotFoundErr.RequestID)
		}
	}
	
	// Test run not found error
	runNotFoundJSON := `{"code":"NotFound","message":"Run not found","request_id":"123456"}`
	err = ParseErrorResponse([]byte(runNotFoundJSON), http.StatusNotFound, "run-123", "run")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	runNotFoundErr, ok := err.(*RunNotFoundError)
	if !ok {
		t.Errorf("Expected RunNotFoundError, got %T", err)
	} else {
		if runNotFoundErr.RunID != "run-123" {
			t.Errorf("Expected run ID run-123, got %s", runNotFoundErr.RunID)
		}
	}
	
	// Test action provider not found error
	providerNotFoundJSON := `{"code":"NotFound","message":"Action provider not found","request_id":"123456"}`
	err = ParseErrorResponse([]byte(providerNotFoundJSON), http.StatusNotFound, "provider-123", "action_provider")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	providerNotFoundErr, ok := err.(*ActionProviderNotFoundError)
	if !ok {
		t.Errorf("Expected ActionProviderNotFoundError, got %T", err)
	} else {
		if providerNotFoundErr.ProviderID != "provider-123" {
			t.Errorf("Expected provider ID provider-123, got %s", providerNotFoundErr.ProviderID)
		}
	}
	
	// Test action role not found error
	roleNotFoundJSON := `{"code":"NotFound","message":"Action role not found","request_id":"123456"}`
	err = ParseErrorResponse([]byte(roleNotFoundJSON), http.StatusNotFound, "provider-123:role-123", "action_role")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	roleNotFoundErr, ok := err.(*ActionRoleNotFoundError)
	if !ok {
		t.Errorf("Expected ActionRoleNotFoundError, got %T", err)
	} else {
		if roleNotFoundErr.ProviderID != "provider-123" {
			t.Errorf("Expected provider ID provider-123, got %s", roleNotFoundErr.ProviderID)
		}
		if roleNotFoundErr.RoleID != "role-123" {
			t.Errorf("Expected role ID role-123, got %s", roleNotFoundErr.RoleID)
		}
	}
	
	// Test forbidden error
	forbiddenJSON := `{"code":"Forbidden","message":"Not authorized to access this resource"}`
	err = ParseErrorResponse([]byte(forbiddenJSON), http.StatusForbidden, "", "")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	forbiddenErr, ok := err.(*ForbiddenError)
	if !ok {
		t.Errorf("Expected ForbiddenError, got %T", err)
	} else {
		if forbiddenErr.Code != "Forbidden" {
			t.Errorf("Expected error code Forbidden, got %s", forbiddenErr.Code)
		}
	}
	
	// Test validation error
	validationJSON := `{"code":"BadRequest","message":"Invalid request parameters"}`
	err = ParseErrorResponse([]byte(validationJSON), http.StatusBadRequest, "", "")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	} else {
		if validationErr.Code != "BadRequest" {
			t.Errorf("Expected error code BadRequest, got %s", validationErr.Code)
		}
	}
	
	// Test generic error
	genericJSON := `{"code":"InternalError","message":"Something went wrong"}`
	err = ParseErrorResponse([]byte(genericJSON), http.StatusInternalServerError, "", "")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	genericErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	} else {
		if genericErr.Code != "InternalError" {
			t.Errorf("Expected error code InternalError, got %s", genericErr.Code)
		}
	}
	
	// Test invalid JSON
	invalidJSON := `{"code":"NotFound",}`
	err = ParseErrorResponse([]byte(invalidJSON), http.StatusNotFound, "flow-123", "flow")
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	genericErr, ok = err.(*ErrorResponse)
	if !ok {
		t.Errorf("Expected ErrorResponse, got %T", err)
	} else {
		if genericErr.Code != "HTTP404" {
			t.Errorf("Expected error code HTTP404, got %s", genericErr.Code)
		}
	}
}

func TestErrorTypeChecking(t *testing.T) {
	// Create error instances
	flowErr := &FlowNotFoundError{
		FlowID: "flow-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Flow not found",
		},
	}
	
	runErr := &RunNotFoundError{
		RunID: "run-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Run not found",
		},
	}
	
	providerErr := &ActionProviderNotFoundError{
		ProviderID: "provider-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Action provider not found",
		},
	}
	
	roleErr := &ActionRoleNotFoundError{
		ProviderID: "provider-123",
		RoleID:     "role-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Action role not found",
		},
	}
	
	forbiddenErr := &ForbiddenError{
		ErrorResponse: &ErrorResponse{
			Code:    "Forbidden",
			Message: "Not authorized",
		},
	}
	
	validationErr := &ValidationError{
		ErrorResponse: &ErrorResponse{
			Code:    "BadRequest",
			Message: "Invalid request",
		},
	}
	
	genericErr := &ErrorResponse{
		Code:    "InternalError",
		Message: "Something went wrong",
	}
	
	// Test IsFlowNotFoundError
	if !IsFlowNotFoundError(flowErr) {
		t.Error("Expected IsFlowNotFoundError to return true for FlowNotFoundError")
	}
	if IsFlowNotFoundError(runErr) {
		t.Error("Expected IsFlowNotFoundError to return false for RunNotFoundError")
	}
	
	// Test IsRunNotFoundError
	if !IsRunNotFoundError(runErr) {
		t.Error("Expected IsRunNotFoundError to return true for RunNotFoundError")
	}
	if IsRunNotFoundError(flowErr) {
		t.Error("Expected IsRunNotFoundError to return false for FlowNotFoundError")
	}
	
	// Test IsActionProviderNotFoundError
	if !IsActionProviderNotFoundError(providerErr) {
		t.Error("Expected IsActionProviderNotFoundError to return true for ActionProviderNotFoundError")
	}
	if IsActionProviderNotFoundError(flowErr) {
		t.Error("Expected IsActionProviderNotFoundError to return false for FlowNotFoundError")
	}
	
	// Test IsActionRoleNotFoundError
	if !IsActionRoleNotFoundError(roleErr) {
		t.Error("Expected IsActionRoleNotFoundError to return true for ActionRoleNotFoundError")
	}
	if IsActionRoleNotFoundError(providerErr) {
		t.Error("Expected IsActionRoleNotFoundError to return false for ActionProviderNotFoundError")
	}
	
	// Test IsForbiddenError
	if !IsForbiddenError(forbiddenErr) {
		t.Error("Expected IsForbiddenError to return true for ForbiddenError")
	}
	if IsForbiddenError(flowErr) {
		t.Error("Expected IsForbiddenError to return false for FlowNotFoundError")
	}
	
	// Test IsValidationError
	if !IsValidationError(validationErr) {
		t.Error("Expected IsValidationError to return true for ValidationError")
	}
	if IsValidationError(genericErr) {
		t.Error("Expected IsValidationError to return false for ErrorResponse")
	}
}

func TestErrorFormattingAndMessages(t *testing.T) {
	// Test ErrorResponse.Error()
	errResp := &ErrorResponse{
		Code:      "TestError",
		Message:   "Test error message",
		RequestID: "123456",
	}
	expected := "flows error [TestError] Test error message (request_id: 123456)"
	if errResp.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, errResp.Error())
	}
	
	// Without request ID
	errResp.RequestID = ""
	expected = "flows error [TestError] Test error message"
	if errResp.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, errResp.Error())
	}
	
	// Test FlowNotFoundError.Error()
	flowErr := &FlowNotFoundError{
		FlowID: "flow-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Flow not found",
		},
	}
	expected = "flow not found: flow-123"
	if flowErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, flowErr.Error())
	}
	
	// Test RunNotFoundError.Error()
	runErr := &RunNotFoundError{
		RunID: "run-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Run not found",
		},
	}
	expected = "run not found: run-123"
	if runErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, runErr.Error())
	}
	
	// Test ActionProviderNotFoundError.Error()
	providerErr := &ActionProviderNotFoundError{
		ProviderID: "provider-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Action provider not found",
		},
	}
	expected = "action provider not found: provider-123"
	if providerErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, providerErr.Error())
	}
	
	// Test ActionRoleNotFoundError.Error()
	roleErr := &ActionRoleNotFoundError{
		ProviderID: "provider-123",
		RoleID:     "role-123",
		ErrorResponse: &ErrorResponse{
			Code:    "NotFound",
			Message: "Action role not found",
		},
	}
	expected = "action role not found: role-123 (provider: provider-123)"
	if roleErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, roleErr.Error())
	}
	
	// Test ForbiddenError.Error()
	forbiddenErr := &ForbiddenError{
		ErrorResponse: &ErrorResponse{
			Code:    "Forbidden",
			Message: "Not authorized",
		},
	}
	expected = "forbidden: Not authorized"
	if forbiddenErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, forbiddenErr.Error())
	}
	
	// Test ValidationError.Error()
	validationErr := &ValidationError{
		ErrorResponse: &ErrorResponse{
			Code:    "BadRequest",
			Message: "Invalid request",
		},
	}
	expected = "validation error: Invalid request"
	if validationErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, validationErr.Error())
	}
}

func TestParseResourceIDs(t *testing.T) {
	// Test with colon separator
	providerID, roleID := parseResourceIDs("provider-123:role-123")
	if providerID != "provider-123" {
		t.Errorf("Expected provider ID provider-123, got %s", providerID)
	}
	if roleID != "role-123" {
		t.Errorf("Expected role ID role-123, got %s", roleID)
	}
	
	// Test without colon
	providerID, roleID = parseResourceIDs("provider-123")
	if providerID != "provider-123" {
		t.Errorf("Expected provider ID provider-123, got %s", providerID)
	}
	if roleID != "" {
		t.Errorf("Expected empty role ID, got %s", roleID)
	}
	
	// Test with multiple colons
	providerID, roleID = parseResourceIDs("provider-123:role-123:extra")
	if providerID != "provider-123" {
		t.Errorf("Expected provider ID provider-123, got %s", providerID)
	}
	if roleID != "role-123:extra" {
		t.Errorf("Expected role ID role-123:extra, got %s", roleID)
	}
}