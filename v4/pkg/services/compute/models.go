// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import "time"

// Endpoint represents a Globus compute endpoint
type Endpoint struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created_at,omitempty"`
	Updated     time.Time `json:"updated_at,omitempty"`
	OwnerID     string    `json:"owner_id,omitempty"`
	Public      bool      `json:"public,omitempty"`
}

// EndpointList represents a list of compute endpoints
type EndpointList struct {
	Endpoints []Endpoint `json:"endpoints"`
	Offset    int        `json:"offset"`
	Limit     int        `json:"limit"`
	Total     int        `json:"total"`
}

// ListEndpointsOptions contains options for listing endpoints
type ListEndpointsOptions struct {
	Limit  int
	Offset int
}

// FunctionSubmission represents a function submission
type FunctionSubmission struct {
	FunctionID string                 `json:"function_id"`
	Args       []interface{}          `json:"args,omitempty"`
	Kwargs     map[string]interface{} `json:"kwargs,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
}

// FunctionRun represents a function execution
type FunctionRun struct {
	ID          string                 `json:"id"`
	FunctionID  string                 `json:"function_id"`
	EndpointID  string                 `json:"endpoint_id"`
	Status      string                 `json:"status"`
	Result      interface{}            `json:"result,omitempty"`
	Exception   string                 `json:"exception,omitempty"`
	StartTime   time.Time              `json:"start_time,omitempty"`
	EndTime     time.Time              `json:"completion_time,omitempty"`
	ExecutionTime float64              `json:"execution_time,omitempty"`
}

// FunctionList represents a list of function runs
type FunctionList struct {
	Functions []FunctionRun `json:"functions"`
	Offset    int           `json:"offset"`
	Limit     int           `json:"limit"`
	Total     int           `json:"total"`
}

// ListFunctionsOptions contains options for listing functions
type ListFunctionsOptions struct {
	EndpointID string
	Limit      int
	Offset     int
}

// FunctionDefinition is used to register a new callable function.
type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Serialized  string `json:"function_code"`
}

// FunctionRegistration is the response from RegisterFunction.
type FunctionRegistration struct {
	FunctionID string `json:"function_id"`
}

// FunctionUpdate contains fields that can be changed on a registered function.
type FunctionUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// TaskStatus represents the execution status of a single function task.
type TaskStatus struct {
	TaskID         string      `json:"task_id"`
	Status         string      `json:"status"`
	Result         interface{} `json:"result,omitempty"`
	Exception      string      `json:"exception,omitempty"`
	CompletionTime string      `json:"completion_time,omitempty"`
}

// TaskList is a paginated list of function task statuses.
type TaskList struct {
	Tasks  []TaskStatus `json:"tasks"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

// ListTasksOptions controls which tasks are returned.
type ListTasksOptions struct {
	EndpointID string
	Limit      int
	Offset     int
}

// TaskRequest is a single function invocation within a batch.
type TaskRequest struct {
	FunctionID string                 `json:"function_id"`
	EndpointID string                 `json:"endpoint_id,omitempty"`
	Args       []interface{}          `json:"args,omitempty"`
	Kwargs     map[string]interface{} `json:"kwargs,omitempty"`
}

// BatchTaskRequest represents a request to execute multiple function calls.
type BatchTaskRequest struct {
	Tasks []TaskRequest `json:"tasks"`
}

// BatchTaskResponse represents the response from a batch execution request.
type BatchTaskResponse struct {
	TaskIDs []string `json:"task_ids,omitempty"`
	Message string   `json:"message,omitempty"`
	Status  string   `json:"status,omitempty"`
}

// BatchTaskStatus represents the combined status of multiple tasks.
type BatchTaskStatus struct {
	Tasks     map[string]TaskStatus `json:"tasks,omitempty"`
	Failed    []string              `json:"failed,omitempty"`
	Pending   []string              `json:"pending,omitempty"`
	Completed []string              `json:"completed,omitempty"`
}
