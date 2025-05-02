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

// DependencyRegistrationRequest represents a request to register a dependency
type DependencyRegistrationRequest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	Public             bool                `json:"public,omitempty"`
	PythonRequirements string              `json:"python_requirements,omitempty"`
	PythonPackages     []PythonPackage     `json:"python_packages,omitempty"`
	CustomDependencies []CustomDependency  `json:"custom_dependencies,omitempty"`
	GitRepo            string              `json:"git_repo,omitempty"`
	GitRef             string              `json:"git_ref,omitempty"`
	Version            string              `json:"version,omitempty"`
}

// DependencyUpdateRequest represents a request to update a dependency
type DependencyUpdateRequest struct {
	Name               string              `json:"name,omitempty"`
	Description        string              `json:"description,omitempty"`
	Public             *bool               `json:"public,omitempty"`
	PythonRequirements string              `json:"python_requirements,omitempty"`
	PythonPackages     []PythonPackage     `json:"python_packages,omitempty"`
	CustomDependencies []CustomDependency  `json:"custom_dependencies,omitempty"`
	GitRepo            string              `json:"git_repo,omitempty"`
	GitRef             string              `json:"git_ref,omitempty"`
	Version            string              `json:"version,omitempty"`
}

// DependencyResponse represents a dependency registered with Globus Compute
type DependencyResponse struct {
	ID                 string              `json:"id,omitempty"`
	Name               string              `json:"name,omitempty"`
	Description        string              `json:"description,omitempty"`
	Owner              string              `json:"owner,omitempty"`
	Public             bool                `json:"public,omitempty"`
	PythonRequirements string              `json:"python_requirements,omitempty"`
	PythonPackages     []PythonPackage     `json:"python_packages,omitempty"`
	CustomDependencies []CustomDependency  `json:"custom_dependencies,omitempty"`
	GitRepo            string              `json:"git_repo,omitempty"`
	GitRef             string              `json:"git_ref,omitempty"`
	Version            string              `json:"version,omitempty"`
	CreatedAt          time.Time           `json:"created_at,omitempty"`
	ModifiedAt         time.Time           `json:"modified_at,omitempty"`
}

// PythonPackage represents a Python package dependency
type PythonPackage struct {
	Name         string   `json:"name"`
	Version      string   `json:"version,omitempty"`
	ExtraIndices []string `json:"extra_indices,omitempty"`
}

// CustomDependency represents a custom dependency
type CustomDependency struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Version  string                 `json:"version,omitempty"`
	Specs    map[string]interface{} `json:"specs,omitempty"`
	Commands []string               `json:"commands,omitempty"`
}

// DependencyList is a list of dependencies
type DependencyList struct {
	Dependencies []DependencyResponse `json:"dependencies,omitempty"`
	Total        int                  `json:"total,omitempty"`
	HasNextPage  bool                 `json:"has_next_page,omitempty"`
	Offset       int                  `json:"offset,omitempty"`
	Limit        int                  `json:"limit,omitempty"`
}

// ListDependenciesOptions are options for listing dependencies
type ListDependenciesOptions struct {
	PerPage int    `url:"per_page,omitempty"`
	Marker  string `url:"marker,omitempty"`
	Search  string `url:"search,omitempty"`
}

// EnvironmentCreateRequest represents a request to create an environment configuration
type EnvironmentCreateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Variables   map[string]string      `json:"variables,omitempty"`
	Secrets     []string               `json:"secrets,omitempty"`
	Resources   map[string]interface{} `json:"resources,omitempty"`
	Public      bool                   `json:"public,omitempty"`
}

// EnvironmentUpdateRequest represents a request to update an environment configuration
type EnvironmentUpdateRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Variables   map[string]string      `json:"variables,omitempty"`
	Secrets     []string               `json:"secrets,omitempty"`
	Resources   map[string]interface{} `json:"resources,omitempty"`
	Public      *bool                  `json:"public,omitempty"`
}

// EnvironmentResponse represents an environment configuration
type EnvironmentResponse struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Owner       string                 `json:"owner,omitempty"`
	Variables   map[string]string      `json:"variables,omitempty"`
	Secrets     []string               `json:"secrets,omitempty"`
	Resources   map[string]interface{} `json:"resources,omitempty"`
	Public      bool                   `json:"public,omitempty"`
	CreatedAt   time.Time              `json:"created_at,omitempty"`
	ModifiedAt  time.Time              `json:"modified_at,omitempty"`
}

// EnvironmentList is a list of environment configurations
type EnvironmentList struct {
	Environments []EnvironmentResponse `json:"environments,omitempty"`
	Total        int                   `json:"total,omitempty"`
	HasNextPage  bool                  `json:"has_next_page,omitempty"`
	Offset       int                   `json:"offset,omitempty"`
	Limit        int                   `json:"limit,omitempty"`
}

// ListEnvironmentsOptions are options for listing environments
type ListEnvironmentsOptions struct {
	PerPage int    `url:"per_page,omitempty"`
	Marker  string `url:"marker,omitempty"`
	Search  string `url:"search,omitempty"`
}

// SecretCreateRequest represents a request to create a secret
type SecretCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value"`
}

// SecretResponse represents a secret (note that the value is never returned)
type SecretResponse struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Owner       string    `json:"owner,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	ModifiedAt  time.Time `json:"modified_at,omitempty"`
}

// WorkflowCreateRequest represents a request to create a workflow
type WorkflowCreateRequest struct {
	Name         string                `json:"name"`
	Description  string                `json:"description,omitempty"`
	Tasks        []WorkflowTask        `json:"tasks"`
	Dependencies map[string][]string   `json:"dependencies,omitempty"`
	ErrorHandling string               `json:"error_handling,omitempty"`
	RetryPolicy  *RetryPolicy          `json:"retry_policy,omitempty"`
	Metadata     map[string]string     `json:"metadata,omitempty"`
	Public       bool                  `json:"public,omitempty"`
}

// WorkflowUpdateRequest represents a request to update a workflow
type WorkflowUpdateRequest struct {
	Name         string                `json:"name,omitempty"`
	Description  string                `json:"description,omitempty"`
	Tasks        []WorkflowTask        `json:"tasks,omitempty"`
	Dependencies map[string][]string   `json:"dependencies,omitempty"`
	ErrorHandling string               `json:"error_handling,omitempty"`
	RetryPolicy  *RetryPolicy          `json:"retry_policy,omitempty"`
	Metadata     map[string]string     `json:"metadata,omitempty"`
	Public       *bool                 `json:"public,omitempty"`
}

// WorkflowList represents a list of workflows
type WorkflowList struct {
	Workflows    []WorkflowResponse   `json:"workflows,omitempty"`
	Total        int                  `json:"total,omitempty"`
	HasNextPage  bool                 `json:"has_next_page,omitempty"`
	Offset       int                  `json:"offset,omitempty"`
	Limit        int                  `json:"limit,omitempty"`
}

// ListWorkflowsOptions are options for listing workflows
type ListWorkflowsOptions struct {
	PerPage     int    `url:"per_page,omitempty"`
	Marker      string `url:"marker,omitempty"`
	Search      string `url:"search,omitempty"`
	FilterScope string `url:"filter_scope,omitempty"`
}

// WorkflowTask represents a task in a workflow
type WorkflowTask struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name,omitempty"`
	FunctionID  string                      `json:"function_id"`
	EndpointID  string                      `json:"endpoint_id"`
	Args        []any                       `json:"args,omitempty"`
	Kwargs      map[string]any              `json:"kwargs,omitempty"`
	Environment string                      `json:"environment,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
	Timeout     int                         `json:"timeout,omitempty"`
}

// RetryPolicy represents a retry policy for tasks or workflows
type RetryPolicy struct {
	MaxRetries      int       `json:"max_retries"`
	RetryInterval   int       `json:"retry_interval,omitempty"`
	BackoffFactor   float64   `json:"backoff_factor,omitempty"`
	MaxInterval     int       `json:"max_interval,omitempty"`
	RetryConditions []string  `json:"retry_conditions,omitempty"`
}

// WorkflowResponse represents a workflow
type WorkflowResponse struct {
	ID          string                      `json:"id,omitempty"`
	Name        string                      `json:"name,omitempty"`
	Description string                      `json:"description,omitempty"`
	Owner       string                      `json:"owner,omitempty"`
	Tasks       []WorkflowTask              `json:"tasks,omitempty"`
	Dependencies map[string][]string        `json:"dependencies,omitempty"`
	ErrorHandling string                    `json:"error_handling,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
	Metadata    map[string]string           `json:"metadata,omitempty"`
	Public      bool                        `json:"public,omitempty"`
	CreatedAt   time.Time                   `json:"created_at,omitempty"`
	ModifiedAt  time.Time                   `json:"modified_at,omitempty"`
}

// WorkflowRunRequest represents a request to run a workflow
type WorkflowRunRequest struct {
	GlobalArgs  map[string]any              `json:"global_args,omitempty"`
	TaskArgs    map[string]map[string]any   `json:"task_args,omitempty"`
	Priority    int                         `json:"priority,omitempty"`
	Description string                      `json:"description,omitempty"`
	RunLabel    string                      `json:"run_label,omitempty"`
}

// WorkflowRunResponse represents the response from a workflow run request
type WorkflowRunResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	WorkflowID  string                      `json:"workflow_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	Message     string                      `json:"message,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
}

// WorkflowStatusResponse represents the status of a workflow run
type WorkflowStatusResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	WorkflowID  string                      `json:"workflow_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	TaskStatus  map[string]TaskStatusInfo   `json:"task_status,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
	CompletedAt time.Time                   `json:"completed_at,omitempty"`
	Error       string                      `json:"error,omitempty"`
	Progress    WorkflowProgressInfo        `json:"progress,omitempty"`
}

// TaskStatusInfo represents the status info of a task in a workflow
type TaskStatusInfo struct {
	Status      string                      `json:"status,omitempty"`
	TaskID      string                      `json:"task_id,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
	CompletedAt time.Time                   `json:"completed_at,omitempty"`
	Error       string                      `json:"error,omitempty"`
	Result      interface{}                 `json:"result,omitempty"`
}

// WorkflowProgressInfo represents progress information for a workflow
type WorkflowProgressInfo struct {
	TotalTasks   int                        `json:"total_tasks,omitempty"`
	Completed    int                        `json:"completed,omitempty"`
	Running      int                        `json:"running,omitempty"`
	Pending      int                        `json:"pending,omitempty"`
	Failed       int                        `json:"failed,omitempty"`
	PercentDone  float64                    `json:"percent_done,omitempty"`
}

// TaskGroupCreateRequest represents a request to create a task group
type TaskGroupCreateRequest struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Tasks       []TaskRequest               `json:"tasks"`
	Concurrency int                         `json:"concurrency,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
	Public      bool                        `json:"public,omitempty"`
}

// TaskGroupResponse represents a task group
type TaskGroupResponse struct {
	ID          string                      `json:"id,omitempty"`
	Name        string                      `json:"name,omitempty"`
	Description string                      `json:"description,omitempty"`
	Owner       string                      `json:"owner,omitempty"`
	Tasks       []TaskRequest               `json:"tasks,omitempty"`
	Concurrency int                         `json:"concurrency,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
	Public      bool                        `json:"public,omitempty"`
	CreatedAt   time.Time                   `json:"created_at,omitempty"`
	ModifiedAt  time.Time                   `json:"modified_at,omitempty"`
}

// TaskGroupUpdateRequest represents a request to update a task group
type TaskGroupUpdateRequest struct {
	Name        string                      `json:"name,omitempty"`
	Description string                      `json:"description,omitempty"`
	Tasks       []TaskRequest               `json:"tasks,omitempty"`
	Concurrency int                         `json:"concurrency,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
	Public      *bool                       `json:"public,omitempty"`
}

// TaskGroupList represents a list of task groups
type TaskGroupList struct {
	TaskGroups   []TaskGroupResponse       `json:"task_groups,omitempty"`
	Total        int                       `json:"total,omitempty"`
	HasNextPage  bool                      `json:"has_next_page,omitempty"`
	Offset       int                       `json:"offset,omitempty"`
	Limit        int                       `json:"limit,omitempty"`
}

// ListTaskGroupsOptions are options for listing task groups
type ListTaskGroupsOptions struct {
	PerPage     int    `url:"per_page,omitempty"`
	Marker      string `url:"marker,omitempty"`
	Search      string `url:"search,omitempty"`
	FilterScope string `url:"filter_scope,omitempty"`
}

// TaskGroupRunRequest represents a request to run a task group
type TaskGroupRunRequest struct {
	Priority    int                         `json:"priority,omitempty"`
	Description string                      `json:"description,omitempty"`
	RunLabel    string                      `json:"run_label,omitempty"`
}

// TaskGroupRunResponse represents the response from a task group run request
type TaskGroupRunResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	TaskGroupID string                      `json:"task_group_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	Message     string                      `json:"message,omitempty"`
	TaskIDs     []string                    `json:"task_ids,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
}

// TaskGroupStatusResponse represents the status of a task group run
type TaskGroupStatusResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	TaskGroupID string                      `json:"task_group_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	TaskStatus  map[string]TaskStatusInfo   `json:"task_status,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
	CompletedAt time.Time                   `json:"completed_at,omitempty"`
	Error       string                      `json:"error,omitempty"`
	Progress    TaskGroupProgressInfo       `json:"progress,omitempty"`
}

// TaskGroupProgressInfo represents progress information for a task group
type TaskGroupProgressInfo struct {
	TotalTasks   int                        `json:"total_tasks,omitempty"`
	Completed    int                        `json:"completed,omitempty"`
	Running      int                        `json:"running,omitempty"`
	Pending      int                        `json:"pending,omitempty"`
	Failed       int                        `json:"failed,omitempty"`
	PercentDone  float64                    `json:"percent_done,omitempty"`
}

// DependencyGraphNode represents a node in a dependency graph
type DependencyGraphNode struct {
	Task         TaskRequest                `json:"task"`
	Dependencies []string                   `json:"dependencies,omitempty"`
	Condition    string                     `json:"condition,omitempty"`
	RetryPolicy  *RetryPolicy               `json:"retry_policy,omitempty"`
	ErrorHandler *ErrorHandler              `json:"error_handler,omitempty"`
}

// ErrorHandler represents an error handler for a node in a dependency graph
type ErrorHandler struct {
	Strategy    string                      `json:"strategy"`
	FallbackID  string                      `json:"fallback_id,omitempty"`
	RetryPolicy *RetryPolicy                `json:"retry_policy,omitempty"`
}

// DependencyGraphRequest represents a request to run a dependency graph
type DependencyGraphRequest struct {
	Nodes       map[string]DependencyGraphNode `json:"nodes"`
	Description string                      `json:"description,omitempty"`
	ErrorPolicy string                      `json:"error_policy,omitempty"`
	RunLabel    string                      `json:"run_label,omitempty"`
}

// DependencyGraphResponse represents the response from a dependency graph run request
type DependencyGraphResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	Message     string                      `json:"message,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
}

// DependencyGraphStatusResponse represents the status of a dependency graph run
type DependencyGraphStatusResponse struct {
	RunID       string                      `json:"run_id,omitempty"`
	Status      string                      `json:"status,omitempty"`
	NodeStatus  map[string]TaskStatusInfo   `json:"node_status,omitempty"`
	StartedAt   time.Time                   `json:"started_at,omitempty"`
	CompletedAt time.Time                   `json:"completed_at,omitempty"`
	Error       string                      `json:"error,omitempty"`
	Progress    DependencyGraphProgressInfo `json:"progress,omitempty"`
}

// DependencyGraphProgressInfo represents progress information for a dependency graph
type DependencyGraphProgressInfo struct {
	TotalNodes   int                        `json:"total_nodes,omitempty"`
	Completed    int                        `json:"completed,omitempty"`
	Running      int                        `json:"running,omitempty"`
	Pending      int                        `json:"pending,omitempty"`
	Failed       int                        `json:"failed,omitempty"`
	Skipped      int                        `json:"skipped,omitempty"`
	PercentDone  float64                    `json:"percent_done,omitempty"`
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
