// Package workflows provides reusable Temporal workflow patterns
package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// BaseWorkflow provides common functionality for all workflows
type BaseWorkflow interface {
	// Execute runs the workflow with the given context and input
	Execute(ctx workflow.Context, input interface{}) (interface{}, error)
	
	// GetName returns the workflow name
	GetName() string
	
	// GetVersion returns the workflow version
	GetVersion() string
	
	// Validate validates the workflow input
	Validate(input interface{}) error
}

// WorkflowMetadata contains metadata about a workflow
type WorkflowMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags"`
	Author      string            `json:"author"`
	CreatedAt   time.Time         `json:"created_at"`
}

// WorkflowState represents the current state of a workflow
type WorkflowState struct {
	Phase       string                 `json:"phase"`
	Progress    int                    `json:"progress"`
	Status      WorkflowStatus         `json:"status"`
	Data        map[string]interface{} `json:"data"`
	Error       string                 `json:"error,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowStatus represents the workflow execution status
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusRunning    WorkflowStatus = "running"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"
	WorkflowStatusPaused     WorkflowStatus = "paused"
	WorkflowStatusEscalated  WorkflowStatus = "escalated"
)

// BaseWorkflowImpl provides a base implementation for all workflows
type BaseWorkflowImpl struct {
	metadata WorkflowMetadata
	state    WorkflowState
}

// NewBaseWorkflow creates a new base workflow implementation
func NewBaseWorkflow(metadata WorkflowMetadata) *BaseWorkflowImpl {
	return &BaseWorkflowImpl{
		metadata: metadata,
		state: WorkflowState{
			Phase:     "initialization",
			Progress:  0,
			Status:    WorkflowStatusPending,
			Data:      make(map[string]interface{}),
			UpdatedAt: time.Now(),
		},
	}
}

// GetName returns the workflow name
func (b *BaseWorkflowImpl) GetName() string {
	return b.metadata.Name
}

// GetVersion returns the workflow version
func (b *BaseWorkflowImpl) GetVersion() string {
	return b.metadata.Version
}

// GetMetadata returns the workflow metadata
func (b *BaseWorkflowImpl) GetMetadata() WorkflowMetadata {
	return b.metadata
}

// GetState returns the current workflow state
func (b *BaseWorkflowImpl) GetState() WorkflowState {
	return b.state
}

// UpdateState updates the workflow state
func (b *BaseWorkflowImpl) UpdateState(phase string, progress int, status WorkflowStatus) {
	b.state.Phase = phase
	b.state.Progress = progress
	b.state.Status = status
	b.state.UpdatedAt = time.Now()
}

// SetStateData sets data in the workflow state
func (b *BaseWorkflowImpl) SetStateData(key string, value interface{}) {
	if b.state.Data == nil {
		b.state.Data = make(map[string]interface{})
	}
	b.state.Data[key] = value
	b.state.UpdatedAt = time.Now()
}

// GetStateData gets data from the workflow state
func (b *BaseWorkflowImpl) GetStateData(key string) (interface{}, bool) {
	if b.state.Data == nil {
		return nil, false
	}
	value, exists := b.state.Data[key]
	return value, exists
}

// SetError sets an error on the workflow state
func (b *BaseWorkflowImpl) SetError(err error) {
	if err != nil {
		b.state.Error = err.Error()
		b.state.Status = WorkflowStatusFailed
		b.state.UpdatedAt = time.Now()
	}
}

// Execute provides default implementation - should be overridden
func (b *BaseWorkflowImpl) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Default implementation - should be overridden by concrete workflows
	return nil, fmt.Errorf("Execute method not implemented for workflow %s", b.GetName())
}

// Validate provides default validation (can be overridden)
func (b *BaseWorkflowImpl) Validate(input interface{}) error {
	// Default implementation does no validation
	return nil
}

// SetupBaseQueries sets up standard query handlers that all workflows should have
func (b *BaseWorkflowImpl) SetupBaseQueries(ctx workflow.Context) error {
	// Query handler for workflow state
	err := workflow.SetQueryHandler(ctx, "getState", func() (WorkflowState, error) {
		return b.state, nil
	})
	if err != nil {
		return err
	}

	// Query handler for workflow metadata
	err = workflow.SetQueryHandler(ctx, "getMetadata", func() (WorkflowMetadata, error) {
		return b.metadata, nil
	})
	if err != nil {
		return err
	}

	// Query handler for workflow progress
	err = workflow.SetQueryHandler(ctx, "getProgress", func() (int, error) {
		return b.state.Progress, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// SetupBaseSignals sets up standard signal handlers that all workflows should support
func (b *BaseWorkflowImpl) SetupBaseSignals(ctx workflow.Context) {
	// Signal handler for pausing workflow
	pauseSignal := workflow.GetSignalChannel(ctx, "pause")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var reason string
			if more := pauseSignal.Receive(ctx, &reason); !more {
				return
			}
			b.state.Status = WorkflowStatusPaused
			b.SetStateData("pause_reason", reason)
			workflow.GetLogger(ctx).Info("Workflow paused", "reason", reason)
		}
	})

	// Signal handler for resuming workflow
	resumeSignal := workflow.GetSignalChannel(ctx, "resume")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var reason string
			if more := resumeSignal.Receive(ctx, &reason); !more {
				return
			}
			b.state.Status = WorkflowStatusRunning
			b.SetStateData("resume_reason", reason)
			workflow.GetLogger(ctx).Info("Workflow resumed", "reason", reason)
		}
	})

	// Signal handler for updating workflow data
	dataUpdateSignal := workflow.GetSignalChannel(ctx, "updateData")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var update DataUpdate
			if more := dataUpdateSignal.Receive(ctx, &update); !more {
				return
			}
			b.SetStateData(update.Key, update.Value)
			workflow.GetLogger(ctx).Info("Workflow data updated", "key", update.Key, "value", update.Value)
		}
	})
}

// DataUpdate represents a data update signal payload
type DataUpdate struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// DefaultActivityOptions returns default activity options for workflows
func DefaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    100 * time.Second,
			MaximumAttempts:    3,
		},
	}
}

// LongRunningActivityOptions returns activity options for long-running activities
func LongRunningActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 24 * time.Hour,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    30 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Minute,
			MaximumAttempts:    5,
		},
	}
}

// HumanTaskActivityOptions returns activity options for human tasks
func HumanTaskActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 7 * 24 * time.Hour, // 7 days for human tasks
		HeartbeatTimeout:    time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    10 * time.Minute,
			BackoffCoefficient: 1.5,
			MaximumInterval:    time.Hour,
			MaximumAttempts:    3,
		},
	}
}

// WorkflowBuilder provides a fluent interface for building workflows
type WorkflowBuilder struct {
	metadata   WorkflowMetadata
	activities []ActivityBuilder
	signals    []SignalBuilder
	queries    []QueryBuilder
	options    WorkflowOptions
}

// ActivityBuilder defines how to build an activity
type ActivityBuilder struct {
	Name        string
	Function    interface{}
	Options     workflow.ActivityOptions
	Description string
}

// SignalBuilder defines how to build a signal handler
type SignalBuilder struct {
	Name        string
	Handler     interface{}
	Description string
}

// QueryBuilder defines how to build a query handler
type QueryBuilder struct {
	Name        string
	Handler     interface{}
	Description string
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(name, version string) *WorkflowBuilder {
	return &WorkflowBuilder{
		metadata: WorkflowMetadata{
			Name:      name,
			Version:   version,
			CreatedAt: time.Now(),
			Tags:      make(map[string]string),
		},
		activities: []ActivityBuilder{},
		signals:    []SignalBuilder{},
		queries:    []QueryBuilder{},
	}
}

// WithDescription sets the workflow description
func (wb *WorkflowBuilder) WithDescription(description string) *WorkflowBuilder {
	wb.metadata.Description = description
	return wb
}

// WithAuthor sets the workflow author
func (wb *WorkflowBuilder) WithAuthor(author string) *WorkflowBuilder {
	wb.metadata.Author = author
	return wb
}

// WithTag adds a tag to the workflow
func (wb *WorkflowBuilder) WithTag(key, value string) *WorkflowBuilder {
	wb.metadata.Tags[key] = value
	return wb
}

// WithActivity adds an activity to the workflow
func (wb *WorkflowBuilder) WithActivity(name string, fn interface{}, opts workflow.ActivityOptions, description string) *WorkflowBuilder {
	wb.activities = append(wb.activities, ActivityBuilder{
		Name:        name,
		Function:    fn,
		Options:     opts,
		Description: description,
	})
	return wb
}

// WithSignal adds a signal handler to the workflow
func (wb *WorkflowBuilder) WithSignal(name string, handler interface{}, description string) *WorkflowBuilder {
	wb.signals = append(wb.signals, SignalBuilder{
		Name:        name,
		Handler:     handler,
		Description: description,
	})
	return wb
}

// WithQuery adds a query handler to the workflow
func (wb *WorkflowBuilder) WithQuery(name string, handler interface{}, description string) *WorkflowBuilder {
	wb.queries = append(wb.queries, QueryBuilder{
		Name:        name,
		Handler:     handler,
		Description: description,
	})
	return wb
}

// WithOptions sets workflow execution options
func (wb *WorkflowBuilder) WithOptions(opts WorkflowOptions) *WorkflowBuilder {
	wb.options = opts
	return wb
}

// WorkflowOptions defines workflow execution options (simplified version)
type WorkflowOptions struct {
	TaskQueue           string
	WorkflowTimeout     time.Duration
	WorkflowTaskTimeout time.Duration
}

// Build returns the workflow metadata and configuration
func (wb *WorkflowBuilder) Build() (WorkflowMetadata, []ActivityBuilder, []SignalBuilder, []QueryBuilder, WorkflowOptions) {
	return wb.metadata, wb.activities, wb.signals, wb.queries, wb.options
}

// Composable workflow interfaces for different patterns
type ComposableWorkflow interface {
	BaseWorkflow
	// Compose allows workflows to be composed with other workflows
	Compose(other ComposableWorkflow) ComposableWorkflow
}

// EventDrivenWorkflow supports event-driven patterns
type EventDrivenWorkflow interface {
	BaseWorkflow
	// HandleEvent processes workflow events
	HandleEvent(ctx workflow.Context, event WorkflowEvent) error
}

// WorkflowEvent represents an event in the workflow
type WorkflowEvent struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// StateMachineWorkflow supports state machine patterns
type StateMachineWorkflow interface {
	BaseWorkflow
	// GetCurrentState returns the current state
	GetCurrentState() string
	// TransitionTo transitions to a new state
	TransitionTo(ctx workflow.Context, newState string) error
	// IsValidTransition checks if a transition is valid
	IsValidTransition(from, to string) bool
}

// WorkflowRegistry manages workflow registrations and provides discovery
type WorkflowRegistry struct {
	workflows map[string]BaseWorkflow
	versions  map[string][]string
}

// NewWorkflowRegistry creates a new workflow registry
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]BaseWorkflow),
		versions:  make(map[string][]string),
	}
}

// Register registers a workflow with the registry
func (wr *WorkflowRegistry) Register(workflow BaseWorkflow) error {
	name := workflow.GetName()
	version := workflow.GetVersion()
	
	key := name + "@" + version
	wr.workflows[key] = workflow
	
	// Track versions for this workflow
	if versions, exists := wr.versions[name]; exists {
		// Check if this version already exists
		for _, v := range versions {
			if v == version {
				return nil // Already registered
			}
		}
		wr.versions[name] = append(versions, version)
	} else {
		wr.versions[name] = []string{version}
	}
	
	return nil
}

// Get retrieves a workflow by name and version
func (wr *WorkflowRegistry) Get(name, version string) (BaseWorkflow, bool) {
	key := name + "@" + version
	workflow, exists := wr.workflows[key]
	return workflow, exists
}

// GetLatest retrieves the latest version of a workflow by name
func (wr *WorkflowRegistry) GetLatest(name string) (BaseWorkflow, bool) {
	versions, exists := wr.versions[name]
	if !exists || len(versions) == 0 {
		return nil, false
	}
	
	// For now, just return the last registered version
	// In a real implementation, you'd want to use semantic versioning
	latestVersion := versions[len(versions)-1]
	return wr.Get(name, latestVersion)
}

// ListVersions returns all versions for a workflow
func (wr *WorkflowRegistry) ListVersions(name string) []string {
	if versions, exists := wr.versions[name]; exists {
		// Return a copy to prevent modification
		result := make([]string, len(versions))
		copy(result, versions)
		return result
	}
	return []string{}
}

// ListWorkflows returns all workflow names
func (wr *WorkflowRegistry) ListWorkflows() []string {
	var names []string
	for name := range wr.versions {
		names = append(names, name)
	}
	return names
}