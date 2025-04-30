// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"time"
)

// ComputeEndpoint represents a Globus Compute endpoint
type ComputeEndpoint struct {
	ID           string                  `json:"id,omitempty"`
	UUID         string                  `json:"uuid,omitempty"`
	Status       string                  `json:"status,omitempty"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description,omitempty"`
	Owner        string                  `json:"owner,omitempty"`
	CreatedAt    time.Time               `json:"created_at,omitempty"`
	LastModified time.Time               `json:"last_modified,omitempty"`
	Connected    bool                    `json:"connected,omitempty"`
	Type         string                  `json:"type,omitempty"`
	Public       bool                    `json:"public,omitempty"`
	Metrics      *ComputeEndpointMetrics `json:"metrics,omitempty"`
}

// ComputeEndpointMetrics contains metrics for a Compute endpoint
type ComputeEndpointMetrics struct {
	OutstandingCounts map[string]int `json:"outstanding_counts,omitempty"`
	RunningCounts     map[string]int `json:"running_counts,omitempty"`
	Utilization       float64        `json:"utilization,omitempty"`
}

// ComputeEndpointList is a list of Compute endpoints
// The API sometimes returns an array of endpoints directly instead of a structured response
type ComputeEndpointList struct {
	Endpoints []ComputeEndpoint 
}

// ListEndpointsOptions are options for listing Compute endpoints
type ListEndpointsOptions struct {
	PerPage      int    `url:"per_page,omitempty"`
	Marker       string `url:"marker,omitempty"`
	OrderBy      string `url:"orderby,omitempty"`
	Search       string `url:"search,omitempty"`
	FilterScope  string `url:"filter_scope,omitempty"`
	FilterStatus string `url:"filter_status,omitempty"`
	IncludeInfo  bool   `url:"include_info,omitempty"`
}

// FunctionResponse represents a function registered with Globus Compute
type FunctionResponse struct {
	ID          string            `json:"id,omitempty"`
	Function    string            `json:"function,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status,omitempty"`
	Detail      string            `json:"detail,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Public      bool              `json:"public,omitempty"`
	Container   *Container        `json:"container,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	ModifiedAt  time.Time         `json:"modified_at,omitempty"`
}

// FunctionList is a list of functions
type FunctionList struct {
	Functions   []FunctionResponse `json:"functions,omitempty"`
	Total       int                `json:"total,omitempty"`
	HasNextPage bool               `json:"has_next_page,omitempty"`
	Offset      int                `json:"offset,omitempty"`
	Limit       int                `json:"limit,omitempty"`
}

// ListFunctionsOptions are options for listing functions
type ListFunctionsOptions struct {
	PerPage     int    `url:"per_page,omitempty"`
	Marker      string `url:"marker,omitempty"`
	OrderBy     string `url:"orderby,omitempty"`
	Search      string `url:"search,omitempty"`
	FilterScope string `url:"filter_scope,omitempty"`
}

// FunctionRegisterRequest represents a request to register a function
type FunctionRegisterRequest struct {
	Function    string            `json:"function"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Public      bool              `json:"public,omitempty"`
	Container   *Container        `json:"container,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// FunctionUpdateRequest represents a request to update a function
type FunctionUpdateRequest struct {
	Function    string            `json:"function,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Public      *bool             `json:"public,omitempty"`
	Container   *Container        `json:"container,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Container represents a container configuration for a function
type Container struct {
	Type      string            `json:"type,omitempty"`
	Image     string            `json:"image,omitempty"`
	Unpack    bool              `json:"unpack,omitempty"`
	Arguments []string          `json:"arguments,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}

// ContainerRegistrationRequest represents a request to register a container
type ContainerRegistrationRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image"`
	Type        string            `json:"type,omitempty"`
	Registry    string            `json:"registry,omitempty"`
	Public      bool              `json:"public,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
}

// ContainerUpdateRequest represents a request to update a container
type ContainerUpdateRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image,omitempty"`
	Type        string            `json:"type,omitempty"`
	Registry    string            `json:"registry,omitempty"`
	Public      *bool             `json:"public,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
}

// ContainerResponse represents a container registered with Globus Compute
type ContainerResponse struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image,omitempty"`
	Type        string            `json:"type,omitempty"`
	Registry    string            `json:"registry,omitempty"`
	Public      bool              `json:"public,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	ModifiedAt  time.Time         `json:"modified_at,omitempty"`
}

// ContainerList is a list of containers
type ContainerList struct {
	Containers  []ContainerResponse `json:"containers,omitempty"`
	Total       int                 `json:"total,omitempty"`
	HasNextPage bool                `json:"has_next_page,omitempty"`
	Offset      int                 `json:"offset,omitempty"`
	Limit       int                 `json:"limit,omitempty"`
}

// ListContainersOptions are options for listing containers
type ListContainersOptions struct {
	PerPage int    `url:"per_page,omitempty"`
	Marker  string `url:"marker,omitempty"`
	Search  string `url:"search,omitempty"`
}

// ContainerTaskRequest represents a request to execute code within a container
type ContainerTaskRequest struct {
	EndpointID  string            `json:"endpoint_id"`
	ContainerID string            `json:"container_id"`
	FunctionID  string            `json:"function_id,omitempty"`
	Code        string            `json:"code,omitempty"`
	Args        []any             `json:"args,omitempty"`
	Kwargs      map[string]any    `json:"kwargs,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	ExecMode    string            `json:"exec_mode,omitempty"`
}

// TaskRequest represents a request to execute a function
type TaskRequest struct {
	FunctionID string         `json:"function_id"`
	EndpointID string         `json:"endpoint_id"`
	Args       []any          `json:"args,omitempty"`
	Kwargs     map[string]any `json:"kwargs,omitempty"`
	Priority   int            `json:"priority,omitempty"`
	ExecMode   string         `json:"exec_mode,omitempty"`
}

// TaskResponse represents the response from a task execution request
type TaskResponse struct {
	TaskID  string `json:"task_id,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// BatchTaskRequest represents a request to execute multiple function calls
type BatchTaskRequest struct {
	Tasks []TaskRequest `json:"tasks"`
}

// BatchTaskResponse represents the response from a batch execution request
type BatchTaskResponse struct {
	TaskIDs []string `json:"task_ids,omitempty"`
	Message string   `json:"message,omitempty"`
	Status  string   `json:"status,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	TaskID      string      `json:"task_id,omitempty"`
	Status      string      `json:"status,omitempty"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Exception   string      `json:"exception,omitempty"`
}

// BatchTaskStatus represents the status of multiple tasks
type BatchTaskStatus struct {
	Tasks     map[string]TaskStatus `json:"tasks,omitempty"`
	Message   string                `json:"message,omitempty"`
	Failed    []string              `json:"failed,omitempty"`
	Pending   []string              `json:"pending,omitempty"`
	Completed []string              `json:"completed,omitempty"`
}

// TaskListOptions are options for listing tasks
type TaskListOptions struct {
	PerPage    int    `url:"per_page,omitempty"`
	Marker     string `url:"marker,omitempty"`
	Status     string `url:"status,omitempty"`
	EndpointID string `url:"endpoint_id,omitempty"`
	FunctionID string `url:"function_id,omitempty"`
}

// TaskList is a list of tasks
type TaskList struct {
	Tasks       []string `json:"tasks,omitempty"`
	Total       int      `json:"total,omitempty"`
	HasNextPage bool     `json:"has_next_page,omitempty"`
	Offset      int      `json:"offset,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}
