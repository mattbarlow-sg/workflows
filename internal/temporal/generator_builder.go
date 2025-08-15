// Package temporal provides Temporal workflow infrastructure
package temporal

import (
	"context"
	"fmt"
	"time"
)

// WorkflowBuilder provides a fluent interface for building workflow specifications
type WorkflowBuilder struct {
	spec   WorkflowSpec
	errors []error
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(name, pkg string) *WorkflowBuilder {
	return &WorkflowBuilder{
		spec: WorkflowSpec{
			Name:    name,
			Package: pkg,
			Options: WorkflowOptions{
				TaskQueue: "default",
			},
		},
		errors: []error{},
	}
}

// WithDescription sets the workflow description
func (b *WorkflowBuilder) WithDescription(desc string) *WorkflowBuilder {
	b.spec.Description = desc
	return b
}

// WithTemplate sets the template to use
func (b *WorkflowBuilder) WithTemplate(template string) *WorkflowBuilder {
	b.spec.Template = template
	return b
}

// WithInput sets the input type
func (b *WorkflowBuilder) WithInput(inputType string) *WorkflowBuilder {
	b.spec.InputType = inputType
	return b
}

// WithOutput sets the output type
func (b *WorkflowBuilder) WithOutput(outputType string) *WorkflowBuilder {
	b.spec.OutputType = outputType
	return b
}

// AddActivity adds an activity to the workflow
func (b *WorkflowBuilder) AddActivity(name, desc string) *ActivityBuilder {
	return &ActivityBuilder{
		parent: b,
		activity: ActivitySpec{
			Name:        name,
			Description: desc,
			Timeout:     10 * time.Minute, // Default timeout
		},
	}
}

// AddSignal adds a signal handler to the workflow
func (b *WorkflowBuilder) AddSignal(name, desc, payloadType string) *WorkflowBuilder {
	b.spec.Signals = append(b.spec.Signals, SignalSpec{
		Name:        name,
		Description: desc,
		PayloadType: payloadType,
	})
	return b
}

// AddQuery adds a query handler to the workflow
func (b *WorkflowBuilder) AddQuery(name, desc, responseType string) *WorkflowBuilder {
	b.spec.Queries = append(b.spec.Queries, QuerySpec{
		Name:         name,
		Description:  desc,
		ResponseType: responseType,
	})
	return b
}

// AddHumanTask adds a human task to the workflow
func (b *WorkflowBuilder) AddHumanTask(name, desc string) *HumanTaskBuilder {
	return &HumanTaskBuilder{
		parent: b,
		task: HumanTaskSpec{
			Name:        name,
			Description: desc,
			Priority:    "medium", // Default priority
			Deadline:    24 * time.Hour, // Default deadline
		},
	}
}

// AddChildWorkflow adds a child workflow
func (b *WorkflowBuilder) AddChildWorkflow(name string) *ChildWorkflowBuilder {
	return &ChildWorkflowBuilder{
		parent: b,
		child: ChildWorkflowSpec{
			Name:      name,
			TaskQueue: "default",
		},
	}
}

// WithTaskQueue sets the task queue
func (b *WorkflowBuilder) WithTaskQueue(queue string) *WorkflowBuilder {
	b.spec.Options.TaskQueue = queue
	return b
}

// WithExecutionTimeout sets the workflow execution timeout
func (b *WorkflowBuilder) WithExecutionTimeout(timeout time.Duration) *WorkflowBuilder {
	b.spec.Options.WorkflowExecutionTimeout = timeout
	return b
}

// WithRunTimeout sets the workflow run timeout
func (b *WorkflowBuilder) WithRunTimeout(timeout time.Duration) *WorkflowBuilder {
	b.spec.Options.WorkflowRunTimeout = timeout
	return b
}

// WithTaskTimeout sets the workflow task timeout
func (b *WorkflowBuilder) WithTaskTimeout(timeout time.Duration) *WorkflowBuilder {
	b.spec.Options.WorkflowTaskTimeout = timeout
	return b
}

// WithRetryPolicy sets the workflow retry policy
func (b *WorkflowBuilder) WithRetryPolicy(maxAttempts int, backoff float32) *WorkflowBuilder {
	b.spec.Retries = RetrySpec{
		MaxAttempts:        maxAttempts,
		BackoffCoefficient: backoff,
	}
	return b
}

// Build validates and returns the workflow specification
func (b *WorkflowBuilder) Build() (WorkflowSpec, error) {
	// Validate required fields
	if b.spec.Name == "" {
		b.errors = append(b.errors, fmt.Errorf("workflow name is required"))
	}
	if b.spec.Package == "" {
		b.errors = append(b.errors, fmt.Errorf("package name is required"))
	}
	
	// Set defaults if not specified
	if b.spec.Template == "" {
		b.spec.Template = "basic"
	}
	if b.spec.InputType == "" {
		b.spec.InputType = "interface{}"
	}
	if b.spec.OutputType == "" {
		b.spec.OutputType = "interface{}"
	}
	
	if len(b.errors) > 0 {
		return WorkflowSpec{}, fmt.Errorf("validation errors: %v", b.errors)
	}
	
	return b.spec, nil
}

// Generate builds and generates the workflow code
func (b *WorkflowBuilder) Generate(ctx context.Context, gen *Generator) (*GeneratedCode, error) {
	spec, err := b.Build()
	if err != nil {
		return nil, err
	}
	return gen.GenerateWorkflow(ctx, spec)
}

// ActivityBuilder builds an activity specification
type ActivityBuilder struct {
	parent   *WorkflowBuilder
	activity ActivitySpec
}

// WithInput sets the activity input type
func (a *ActivityBuilder) WithInput(inputType string) *ActivityBuilder {
	a.activity.InputType = inputType
	return a
}

// WithOutput sets the activity output type
func (a *ActivityBuilder) WithOutput(outputType string) *ActivityBuilder {
	a.activity.OutputType = outputType
	return a
}

// WithTimeout sets the activity timeout
func (a *ActivityBuilder) WithTimeout(timeout time.Duration) *ActivityBuilder {
	a.activity.Timeout = timeout
	return a
}

// WithRetryPolicy sets the activity retry policy
func (a *ActivityBuilder) WithRetryPolicy(initial time.Duration, backoff float32, max time.Duration, attempts int32) *ActivityBuilder {
	a.activity.RetryPolicy = RetryPolicy{
		InitialInterval:    initial,
		BackoffCoefficient: backoff,
		MaximumInterval:    max,
		MaximumAttempts:    attempts,
	}
	return a
}

// AsHumanTask marks the activity as a human task
func (a *ActivityBuilder) AsHumanTask() *ActivityBuilder {
	a.activity.IsHumanTask = true
	return a
}

// Done adds the activity to the workflow and returns the workflow builder
func (a *ActivityBuilder) Done() *WorkflowBuilder {
	a.parent.spec.Activities = append(a.parent.spec.Activities, a.activity)
	return a.parent
}

// HumanTaskBuilder builds a human task specification
type HumanTaskBuilder struct {
	parent *WorkflowBuilder
	task   HumanTaskSpec
}

// AssignTo sets who the task is assigned to
func (h *HumanTaskBuilder) AssignTo(assignee string) *HumanTaskBuilder {
	h.task.AssignedTo = assignee
	return h
}

// WithEscalation sets escalation parameters
func (h *HumanTaskBuilder) WithEscalation(escalationTime time.Duration, escalateTo string) *HumanTaskBuilder {
	h.task.EscalationTime = escalationTime
	h.task.EscalationTo = escalateTo
	return h
}

// WithPriority sets the task priority
func (h *HumanTaskBuilder) WithPriority(priority string) *HumanTaskBuilder {
	h.task.Priority = priority
	return h
}

// WithDeadline sets the task deadline
func (h *HumanTaskBuilder) WithDeadline(deadline time.Duration) *HumanTaskBuilder {
	h.task.Deadline = deadline
	return h
}

// Done adds the human task to the workflow and returns the workflow builder
func (h *HumanTaskBuilder) Done() *WorkflowBuilder {
	h.parent.spec.HumanTasks = append(h.parent.spec.HumanTasks, h.task)
	return h.parent
}

// ChildWorkflowBuilder builds a child workflow specification
type ChildWorkflowBuilder struct {
	parent *WorkflowBuilder
	child  ChildWorkflowSpec
}

// WithInput sets the child workflow input type
func (c *ChildWorkflowBuilder) WithInput(inputType string) *ChildWorkflowBuilder {
	c.child.InputType = inputType
	return c
}

// WithOutput sets the child workflow output type
func (c *ChildWorkflowBuilder) WithOutput(outputType string) *ChildWorkflowBuilder {
	c.child.OutputType = outputType
	return c
}

// WithTaskQueue sets the child workflow task queue
func (c *ChildWorkflowBuilder) WithTaskQueue(queue string) *ChildWorkflowBuilder {
	c.child.TaskQueue = queue
	return c
}

// Done adds the child workflow to the parent and returns the workflow builder
func (c *ChildWorkflowBuilder) Done() *WorkflowBuilder {
	c.parent.spec.ChildWorkflows = append(c.parent.spec.ChildWorkflows, c.child)
	return c.parent
}

// QuickGenerators provide pre-configured workflow generators

// GenerateApprovalWorkflow generates a standard approval workflow
func GenerateApprovalWorkflow(name, pkg string, approvers []string) *WorkflowBuilder {
	builder := NewWorkflowBuilder(name, pkg).
		WithTemplate("approval").
		WithDescription("Approval workflow with escalation").
		WithInput("RequestID string\nRequestType string\nRequester string").
		WithOutput("Status ApprovalStatus\nApprovals []ApprovalRecord")
	
	// Add human tasks for each approver
	for i, approver := range approvers {
		priority := "medium"
		if i == 0 {
			priority = "high" // First approver is high priority
		}
		
		builder.AddHumanTask(
			fmt.Sprintf("Approval_%d", i+1),
			fmt.Sprintf("Approval required from %s", approver),
		).
			AssignTo(approver).
			WithPriority(priority).
			WithDeadline(24 * time.Hour).
			WithEscalation(12*time.Hour, "manager").
			Done()
	}
	
	return builder
}

// GenerateScheduledWorkflow generates a scheduled/cron workflow
func GenerateScheduledWorkflow(name, pkg, schedule string) *WorkflowBuilder {
	return NewWorkflowBuilder(name, pkg).
		WithTemplate("scheduled").
		WithDescription(fmt.Sprintf("Scheduled workflow running on: %s", schedule)).
		WithInput(fmt.Sprintf("Schedule string `default:\"%s\"`\nMaxRuns int", schedule)).
		WithOutput("ExecutionCount int\nLastExecution time.Time")
}

// GenerateLongRunningWorkflow generates a long-running workflow with checkpoints
func GenerateLongRunningWorkflow(name, pkg string, maxDuration time.Duration) *WorkflowBuilder {
	return NewWorkflowBuilder(name, pkg).
		WithTemplate("long_running").
		WithDescription("Long-running workflow with checkpoints and pause/resume").
		WithInput("MaxDuration time.Duration").
		WithOutput("TotalDuration time.Duration\nCheckpoints []Checkpoint").
		WithExecutionTimeout(maxDuration).
		AddActivity("Initialize", "Initialize workflow state").
			WithTimeout(30 * time.Minute).
			Done().
		AddActivity("Process", "Main processing logic").
			WithTimeout(2 * time.Hour).
			Done().
		AddActivity("Validate", "Validate results").
			WithTimeout(1 * time.Hour).
			Done().
		AddActivity("Finalize", "Finalize and cleanup").
			WithTimeout(30 * time.Minute).
			Done()
}

// GenerateHumanTaskWorkflow generates a workflow with human tasks
func GenerateHumanTaskWorkflow(name, pkg string) *WorkflowBuilder {
	return NewWorkflowBuilder(name, pkg).
		WithTemplate("human_task").
		WithDescription("Workflow with human task management").
		WithInput("TaskRequests []TaskRequest").
		WithOutput("CompletedTasks []CompletedTask").
		AddSignal("taskComplete", "Signal task completion", "TaskCompletion").
		AddSignal("taskReassign", "Reassign task to another user", "TaskReassignment").
		AddQuery("getTasks", "Get all pending tasks", "[]HumanTask").
		AddQuery("getCompletedTasks", "Get completed tasks", "[]CompletedTask")
}

// GenerateETLWorkflow generates an ETL (Extract, Transform, Load) workflow
func GenerateETLWorkflow(name, pkg string) *WorkflowBuilder {
	return NewWorkflowBuilder(name, pkg).
		WithDescription("ETL workflow for data processing").
		WithInput("SourceConfig SourceConfig\nTargetConfig TargetConfig").
		WithOutput("RecordsProcessed int\nErrors []error").
		AddActivity("Extract", "Extract data from source").
			WithInput("SourceConfig").
			WithOutput("[]RawData").
			WithTimeout(1 * time.Hour).
			WithRetryPolicy(10*time.Second, 2.0, 5*time.Minute, 3).
			Done().
		AddActivity("Transform", "Transform data").
			WithInput("[]RawData").
			WithOutput("[]TransformedData").
			WithTimeout(2 * time.Hour).
			Done().
		AddActivity("Load", "Load data to target").
			WithInput("[]TransformedData").
			WithOutput("LoadResult").
			WithTimeout(1 * time.Hour).
			WithRetryPolicy(30*time.Second, 2.0, 10*time.Minute, 5).
			Done().
		AddActivity("Validate", "Validate loaded data").
			WithInput("LoadResult").
			WithOutput("ValidationResult").
			WithTimeout(30 * time.Minute).
			Done()
}