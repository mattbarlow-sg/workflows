// Package workflows provides human task workflow implementation
package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// HumanTaskWorkflow implements a human task workflow with escalation
type HumanTaskWorkflow struct {
	*BaseWorkflowImpl
	tasks               []HumanTask
	completedTasks      []CompletedTask
	escalatedTasks      []EscalatedTask
	activeAssignments   map[string]string // taskID -> assignee
}

// HumanTaskInput defines the input for the human task workflow
type HumanTaskInput struct {
	WorkflowID      string                 `json:"workflow_id"`
	Tasks           []HumanTaskDefinition  `json:"tasks"`
	GlobalSettings  GlobalTaskSettings     `json:"global_settings"`
	EscalationChain []string               `json:"escalation_chain"`
	NotificationConfig NotificationConfig  `json:"notification_config"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

// HumanTaskOutput defines the output of the human task workflow
type HumanTaskOutput struct {
	WorkflowID       string          `json:"workflow_id"`
	CompletedTasks   []CompletedTask `json:"completed_tasks"`
	EscalatedTasks   []EscalatedTask `json:"escalated_tasks"`
	PendingTasks     []HumanTask     `json:"pending_tasks"`
	Status           HumanTaskWorkflowStatus `json:"status"`
	StartTime        time.Time       `json:"start_time"`
	EndTime          time.Time       `json:"end_time"`
	TotalDuration    time.Duration   `json:"total_duration"`
}

// HumanTaskDefinition defines a human task to be created
type HumanTaskDefinition struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	AssignedTo       string                 `json:"assigned_to"`
	Priority         Priority               `json:"priority"`
	Deadline         time.Duration          `json:"deadline"`
	EscalationTime   time.Duration          `json:"escalation_time"`
	RequiredSkills   []string               `json:"required_skills,omitempty"`
	FormData         map[string]interface{} `json:"form_data,omitempty"`
	Dependencies     []string               `json:"dependencies,omitempty"` // IDs of tasks that must complete first
	MaxRetries       int                    `json:"max_retries"`
	AllowReassignment bool                  `json:"allow_reassignment"`
}

// TaskStatus defines task status
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusEscalated  TaskStatus = "escalated"
	TaskStatusFailed     TaskStatus = "failed"
)

// HumanTask represents an active human task
type HumanTask struct {
	Definition    HumanTaskDefinition    `json:"definition"`
	Status        TaskStatus             `json:"status"`
	AssignedTo    string                 `json:"assigned_to"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Deadline      time.Time              `json:"deadline"`
	RetryCount    int                    `json:"retry_count"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CompletedTask represents a completed human task
type CompletedTask struct {
	Task          HumanTask              `json:"task"`
	CompletedBy   string                 `json:"completed_by"`
	CompletedAt   time.Time              `json:"completed_at"`
	Result        TaskResult             `json:"result"`
	Duration      time.Duration          `json:"duration"`
	Notes         string                 `json:"notes,omitempty"`
}

// EscalatedTask represents an escalated human task
type EscalatedTask struct {
	Task           HumanTask `json:"task"`
	EscalatedAt    time.Time `json:"escalated_at"`
	EscalatedTo    string    `json:"escalated_to"`
	EscalationLevel int      `json:"escalation_level"`
	Reason         string    `json:"reason"`
	OriginalAssignee string  `json:"original_assignee"`
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	Status    TaskResultStatus       `json:"status"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Files     []TaskFile             `json:"files,omitempty"`
	Comments  string                 `json:"comments,omitempty"`
}

// TaskFile represents a file attached to a task result
type TaskFile struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// TaskResultStatus represents the status of a task result
type TaskResultStatus string

const (
	TaskResultStatusCompleted TaskResultStatus = "completed"
	TaskResultStatusSkipped   TaskResultStatus = "skipped"
	TaskResultStatusFailed    TaskResultStatus = "failed"
	TaskResultStatusRejected  TaskResultStatus = "rejected"
)

// HumanTaskWorkflowStatus represents the status of the human task workflow
type HumanTaskWorkflowStatus string

const (
	HumanTaskWorkflowStatusRunning   HumanTaskWorkflowStatus = "running"
	HumanTaskWorkflowStatusCompleted HumanTaskWorkflowStatus = "completed"
	HumanTaskWorkflowStatusFailed    HumanTaskWorkflowStatus = "failed"
	HumanTaskWorkflowStatusCancelled HumanTaskWorkflowStatus = "cancelled"
)

// GlobalTaskSettings defines global settings for all tasks
type GlobalTaskSettings struct {
	DefaultTimeout     time.Duration `json:"default_timeout"`
	DefaultEscalation  time.Duration `json:"default_escalation"`
	MaxRetries         int           `json:"max_retries"`
	AutoReassignOnFail bool          `json:"auto_reassign_on_fail"`
	RequireComments    bool          `json:"require_comments"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	OnTaskCreated    []string `json:"on_task_created,omitempty"`
	OnTaskCompleted  []string `json:"on_task_completed,omitempty"`
	OnTaskEscalated  []string `json:"on_task_escalated,omitempty"`
	OnTaskOverdue    []string `json:"on_task_overdue,omitempty"`
	ReminderIntervals []time.Duration `json:"reminder_intervals,omitempty"`
}

// NewHumanTaskWorkflow creates a new human task workflow
func NewHumanTaskWorkflow() *HumanTaskWorkflow {
	metadata := WorkflowMetadata{
		Name:        "HumanTaskWorkflow",
		Version:     "1.0.0",
		Description: "Human task workflow with escalation and dependency management",
		Tags: map[string]string{
			"type":     "human-task",
			"category": "interactive",
		},
		Author:    "Temporal Workflows",
		CreatedAt: time.Now(),
	}

	return &HumanTaskWorkflow{
		BaseWorkflowImpl:  NewBaseWorkflow(metadata),
		tasks:             []HumanTask{},
		completedTasks:    []CompletedTask{},
		escalatedTasks:    []EscalatedTask{},
		activeAssignments: make(map[string]string),
	}
}

// Execute runs the human task workflow
func (htw *HumanTaskWorkflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Type assertion for input
	taskInput, ok := input.(HumanTaskInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for HumanTaskWorkflow")
	}

	// Validate input
	if err := htw.Validate(taskInput); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting human task workflow", "workflowID", taskInput.WorkflowID, "taskCount", len(taskInput.Tasks))

	startTime := workflow.Now(ctx)
	htw.UpdateState("initialization", 0, WorkflowStatusRunning)

	// Setup base queries and signals
	if err := htw.SetupBaseQueries(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup base queries: %w", err)
	}
	htw.SetupBaseSignals(ctx)

	// Setup human task specific queries and signals
	htw.setupTaskQueries(ctx)
	taskCompleteSignal, taskReassignSignal, cancelSignal := htw.setupTaskSignals(ctx)

	// Configure activity options
	ao := HumanTaskActivityOptions()
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Create dependency graph
	dependencyGraph := htw.buildDependencyGraph(taskInput.Tasks)
	
	// Initialize tasks
	for _, taskDef := range taskInput.Tasks {
		task := HumanTask{
			Definition: taskDef,
			Status:     TaskStatusPending,
			AssignedTo: taskDef.AssignedTo,
			CreatedAt:  startTime,
			UpdatedAt:  startTime,
			Deadline:   startTime.Add(htw.getTaskDeadline(taskDef, taskInput.GlobalSettings)),
			RetryCount: 0,
		}
		htw.tasks = append(htw.tasks, task)
		htw.activeAssignments[task.Definition.ID] = task.AssignedTo
	}

	// Main task processing loop
	for len(htw.completedTasks) < len(taskInput.Tasks) {
		// Find tasks ready to execute (no pending dependencies)
		readyTasks := htw.getReadyTasks(dependencyGraph)
		
		if len(readyTasks) == 0 {
			// No tasks ready, check if we're deadlocked
			if htw.isDeadlocked() {
				return htw.createOutput(taskInput, startTime, HumanTaskWorkflowStatusFailed), 
					fmt.Errorf("workflow deadlocked: no tasks can proceed")
			}
			
			// Wait for task completion or other signals
			htw.waitForTaskEvents(ctx, taskCompleteSignal, taskReassignSignal, cancelSignal)
			continue
		}

		// Create/update tasks that are ready
		for _, task := range readyTasks {
			if task.Status == TaskStatusPending {
				if err := htw.createTask(ctx, taskInput, task); err != nil {
					logger.Error("Failed to create task", "taskID", task.Definition.ID, "error", err)
					continue
				}
				htw.updateTaskStatus(task.Definition.ID, TaskStatusAssigned)
			}
		}

		// Update progress
		progress := (len(htw.completedTasks) * 100) / len(taskInput.Tasks)
		htw.UpdateState(fmt.Sprintf("processing_tasks_%d_%d", len(htw.completedTasks), len(taskInput.Tasks)), 
			progress, WorkflowStatusRunning)

		// Wait for task events with timeout
		htw.waitForTaskEventsWithTimeout(ctx, taskCompleteSignal, taskReassignSignal, cancelSignal, 
			taskInput.GlobalSettings.DefaultTimeout)

		// Check for escalations
		htw.checkForEscalations(ctx, taskInput)
		
		// Check for cancellation
		if htw.GetState().Status == WorkflowStatusCancelled {
			return htw.createOutput(taskInput, startTime, HumanTaskWorkflowStatusCancelled), nil
		}
	}

	// All tasks completed
	endTime := workflow.Now(ctx)
	htw.UpdateState("completed", 100, WorkflowStatusCompleted)
	
	logger.Info("Human task workflow completed successfully", 
		"completedTasks", len(htw.completedTasks), 
		"escalatedTasks", len(htw.escalatedTasks),
		"totalDuration", endTime.Sub(startTime))

	return htw.createOutput(taskInput, startTime, HumanTaskWorkflowStatusCompleted), nil
}

// Validate validates the human task input
func (htw *HumanTaskWorkflow) Validate(input interface{}) error {
	taskInput, ok := input.(HumanTaskInput)
	if !ok {
		return fmt.Errorf("input must be of type HumanTaskInput")
	}

	if taskInput.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if len(taskInput.Tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	// Validate each task
	taskIDs := make(map[string]bool)
	for i, task := range taskInput.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task %d ID is required", i)
		}

		// Check for duplicate task IDs
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		if task.Title == "" {
			return fmt.Errorf("task %d title is required", i)
		}

		if task.AssignedTo == "" {
			return fmt.Errorf("task %d must be assigned to someone", i)
		}

		if task.Deadline <= 0 {
			return fmt.Errorf("task %d deadline must be positive", i)
		}

		// Validate dependencies exist
		for _, depID := range task.Dependencies {
			if !taskIDs[depID] && depID != task.ID {
				return fmt.Errorf("task %s has invalid dependency: %s", task.ID, depID)
			}
		}
	}

	// Check for circular dependencies (simplified check)
	if htw.hasCircularDependencies(taskInput.Tasks) {
		return fmt.Errorf("circular dependencies detected")
	}

	return nil
}

// buildDependencyGraph builds a dependency graph from tasks
func (htw *HumanTaskWorkflow) buildDependencyGraph(tasks []HumanTaskDefinition) map[string][]string {
	graph := make(map[string][]string)
	
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}
	
	return graph
}

// getReadyTasks returns tasks that have no pending dependencies
func (htw *HumanTaskWorkflow) getReadyTasks(dependencyGraph map[string][]string) []HumanTask {
	var readyTasks []HumanTask
	completedTaskIDs := htw.getCompletedTaskIDs()
	
	for i, task := range htw.tasks {
		if task.Status != TaskStatusPending && task.Status != TaskStatusAssigned {
			continue
		}
		
		// Check if all dependencies are completed
		allDepsCompleted := true
		for _, depID := range dependencyGraph[task.Definition.ID] {
			if !completedTaskIDs[depID] {
				allDepsCompleted = false
				break
			}
		}
		
		if allDepsCompleted {
			readyTasks = append(readyTasks, htw.tasks[i])
		}
	}
	
	return readyTasks
}

// getCompletedTaskIDs returns a set of completed task IDs
func (htw *HumanTaskWorkflow) getCompletedTaskIDs() map[string]bool {
	completed := make(map[string]bool)
	for _, task := range htw.completedTasks {
		completed[task.Task.Definition.ID] = true
	}
	return completed
}

// createTask creates a human task in the external system
func (htw *HumanTaskWorkflow) createTask(ctx workflow.Context, input HumanTaskInput, task HumanTask) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Creating human task", "taskID", task.Definition.ID, "assignedTo", task.AssignedTo)

	taskInput := HumanTaskCreationInput{
		TaskID:      task.Definition.ID,
		TaskType:    "human_task",
		AssignedTo:  task.AssignedTo,
		Title:       task.Definition.Title,
		Description: task.Definition.Description,
		Priority:    string(task.Definition.Priority),
		Deadline:    task.Deadline,
		Data: map[string]interface{}{
			"workflowID":     input.WorkflowID,
			"formData":       task.Definition.FormData,
			"requiredSkills": task.Definition.RequiredSkills,
		},
	}

	return workflow.ExecuteActivity(ctx, CreateHumanTaskForWorkflowActivity, taskInput).Get(ctx, nil)
}

// waitForTaskEvents waits for task completion, reassignment, or cancellation signals
func (htw *HumanTaskWorkflow) waitForTaskEvents(ctx workflow.Context, 
	taskCompleteSignal, taskReassignSignal, cancelSignal workflow.ReceiveChannel) {
	
	selector := workflow.NewSelector(ctx)
	
	// Task completion handler
	selector.AddReceive(taskCompleteSignal, func(c workflow.ReceiveChannel, more bool) {
		var completion TaskCompletionSignal
		c.Receive(ctx, &completion)
		htw.handleTaskCompletion(ctx, completion)
	})
	
	// Task reassignment handler
	selector.AddReceive(taskReassignSignal, func(c workflow.ReceiveChannel, more bool) {
		var reassignment TaskReassignmentSignal
		c.Receive(ctx, &reassignment)
		htw.handleTaskReassignment(ctx, reassignment)
	})
	
	// Cancellation handler
	selector.AddReceive(cancelSignal, func(c workflow.ReceiveChannel, more bool) {
		var reason string
		c.Receive(ctx, &reason)
		htw.UpdateState("cancelled", htw.GetState().Progress, WorkflowStatusCancelled)
		workflow.GetLogger(ctx).Info("Human task workflow cancelled", "reason", reason)
	})
	
	selector.Select(ctx)
}

// waitForTaskEventsWithTimeout waits for task events with a timeout
func (htw *HumanTaskWorkflow) waitForTaskEventsWithTimeout(ctx workflow.Context,
	taskCompleteSignal, taskReassignSignal, cancelSignal workflow.ReceiveChannel,
	timeout time.Duration) {
	
	selector := workflow.NewSelector(ctx)
	timer := workflow.NewTimer(ctx, timeout)
	
	// Add event handlers (same as above)
	htw.addTaskEventHandlers(ctx, selector, taskCompleteSignal, taskReassignSignal, cancelSignal)
	
	// Add timeout handler
	selector.AddFuture(timer, func(f workflow.Future) {
		// Timeout reached, continue processing
		workflow.GetLogger(ctx).Debug("Task event timeout reached")
	})
	
	selector.Select(ctx)
}

// addTaskEventHandlers adds task event handlers to a selector
func (htw *HumanTaskWorkflow) addTaskEventHandlers(ctx workflow.Context, selector workflow.Selector,
	taskCompleteSignal, taskReassignSignal, cancelSignal workflow.ReceiveChannel) {
	
	selector.AddReceive(taskCompleteSignal, func(c workflow.ReceiveChannel, more bool) {
		var completion TaskCompletionSignal
		c.Receive(ctx, &completion)
		htw.handleTaskCompletion(ctx, completion)
	})
	
	selector.AddReceive(taskReassignSignal, func(c workflow.ReceiveChannel, more bool) {
		var reassignment TaskReassignmentSignal
		c.Receive(ctx, &reassignment)
		htw.handleTaskReassignment(ctx, reassignment)
	})
	
	selector.AddReceive(cancelSignal, func(c workflow.ReceiveChannel, more bool) {
		var reason string
		c.Receive(ctx, &reason)
		htw.UpdateState("cancelled", htw.GetState().Progress, WorkflowStatusCancelled)
		workflow.GetLogger(ctx).Info("Human task workflow cancelled", "reason", reason)
	})
}

// handleTaskCompletion handles task completion signals
func (htw *HumanTaskWorkflow) handleTaskCompletion(ctx workflow.Context, completion TaskCompletionSignal) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Task completed", "taskID", completion.TaskID, "completedBy", completion.CompletedBy)

	// Find the task
	for i, task := range htw.tasks {
		if task.Definition.ID == completion.TaskID {
			completedTask := CompletedTask{
				Task:        task,
				CompletedBy: completion.CompletedBy,
				CompletedAt: workflow.Now(ctx),
				Result:      completion.Result,
				Duration:    workflow.Now(ctx).Sub(task.CreatedAt),
				Notes:       completion.Notes,
			}
			
			htw.completedTasks = append(htw.completedTasks, completedTask)
			htw.tasks[i].Status = TaskStatusCompleted
			htw.tasks[i].UpdatedAt = workflow.Now(ctx)
			
			// Remove from active assignments
			delete(htw.activeAssignments, completion.TaskID)
			break
		}
	}
}

// handleTaskReassignment handles task reassignment signals
func (htw *HumanTaskWorkflow) handleTaskReassignment(ctx workflow.Context, reassignment TaskReassignmentSignal) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Task reassigned", "taskID", reassignment.TaskID, "newAssignee", reassignment.NewAssignee)

	// Update task assignment
	for i, task := range htw.tasks {
		if task.Definition.ID == reassignment.TaskID {
			oldAssignee := htw.tasks[i].AssignedTo
			htw.tasks[i].AssignedTo = reassignment.NewAssignee
			htw.tasks[i].UpdatedAt = workflow.Now(ctx)
			htw.activeAssignments[reassignment.TaskID] = reassignment.NewAssignee
			
			// Update task in external system
			workflow.ExecuteActivity(ctx, ReassignHumanTaskActivity, ReassignTaskInput{
				TaskID:      reassignment.TaskID,
				OldAssignee: oldAssignee,
				NewAssignee: reassignment.NewAssignee,
				Reason:      reassignment.Reason,
			}).Get(ctx, nil)
			break
		}
	}
}

// checkForEscalations checks for tasks that need to be escalated
func (htw *HumanTaskWorkflow) checkForEscalations(ctx workflow.Context, input HumanTaskInput) {
	currentTime := workflow.Now(ctx)
	
	for i, task := range htw.tasks {
		if task.Status != TaskStatusAssigned && task.Status != TaskStatusInProgress {
			continue
		}
		
		// Check if escalation time has passed
		escalationTime := task.CreatedAt.Add(htw.getTaskEscalationTime(task.Definition, input.GlobalSettings))
		if currentTime.After(escalationTime) {
			htw.escalateTask(ctx, input, i)
		}
	}
}

// escalateTask escalates a task to the next level
func (htw *HumanTaskWorkflow) escalateTask(ctx workflow.Context, input HumanTaskInput, taskIndex int) {
	logger := workflow.GetLogger(ctx)
	task := htw.tasks[taskIndex]
	
	// Find next escalation level
	escalationLevel := 0
	for _, escalated := range htw.escalatedTasks {
		if escalated.Task.Definition.ID == task.Definition.ID {
			if escalated.EscalationLevel > escalationLevel {
				escalationLevel = escalated.EscalationLevel
			}
		}
	}
	escalationLevel++
	
	if escalationLevel > len(input.EscalationChain) {
		logger.Warn("No more escalation levels available", "taskID", task.Definition.ID)
		return
	}
	
	escalateTo := input.EscalationChain[escalationLevel-1]
	
	escalatedTask := EscalatedTask{
		Task:            task,
		EscalatedAt:     workflow.Now(ctx),
		EscalatedTo:     escalateTo,
		EscalationLevel: escalationLevel,
		Reason:          "Timeout",
		OriginalAssignee: task.AssignedTo,
	}
	
	htw.escalatedTasks = append(htw.escalatedTasks, escalatedTask)
	
	// Reassign task
	htw.tasks[taskIndex].AssignedTo = escalateTo
	htw.tasks[taskIndex].Status = TaskStatusEscalated
	htw.tasks[taskIndex].UpdatedAt = workflow.Now(ctx)
	htw.activeAssignments[task.Definition.ID] = escalateTo
	
	logger.Info("Task escalated", "taskID", task.Definition.ID, "escalatedTo", escalateTo, "level", escalationLevel)
	
	// Notify external system
	workflow.ExecuteActivity(ctx, EscalateHumanTaskActivity, EscalateTaskInput{
		TaskID:     task.Definition.ID,
		EscalateTo: escalateTo,
		Level:      escalationLevel,
		Reason:     "Timeout escalation",
	}).Get(ctx, nil)
}

// Helper methods

func (htw *HumanTaskWorkflow) getTaskDeadline(task HumanTaskDefinition, settings GlobalTaskSettings) time.Duration {
	if task.Deadline > 0 {
		return task.Deadline
	}
	return settings.DefaultTimeout
}

func (htw *HumanTaskWorkflow) getTaskEscalationTime(task HumanTaskDefinition, settings GlobalTaskSettings) time.Duration {
	if task.EscalationTime > 0 {
		return task.EscalationTime
	}
	return settings.DefaultEscalation
}

func (htw *HumanTaskWorkflow) updateTaskStatus(taskID string, status TaskStatus) {
	for i, task := range htw.tasks {
		if task.Definition.ID == taskID {
			htw.tasks[i].Status = status
			htw.tasks[i].UpdatedAt = time.Now()
			break
		}
	}
}

func (htw *HumanTaskWorkflow) isDeadlocked() bool {
	// Simple deadlock detection - no tasks are in progress and none can start
	hasActiveTask := false
	for _, task := range htw.tasks {
		if task.Status == TaskStatusAssigned || task.Status == TaskStatusInProgress {
			hasActiveTask = true
			break
		}
	}
	return !hasActiveTask && len(htw.completedTasks) < len(htw.tasks)
}

func (htw *HumanTaskWorkflow) hasCircularDependencies(tasks []HumanTaskDefinition) bool {
	// Simplified circular dependency detection
	// In practice, you'd implement a proper topological sort
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	var hasCycle func(string) bool
	hasCycle = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true
		
		// Find the task
		var task *HumanTaskDefinition
		for _, t := range tasks {
			if t.ID == taskID {
				task = &t
				break
			}
		}
		
		if task == nil {
			return false
		}
		
		for _, dep := range task.Dependencies {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}
		
		recStack[taskID] = false
		return false
	}
	
	for _, task := range tasks {
		if !visited[task.ID] {
			if hasCycle(task.ID) {
				return true
			}
		}
	}
	
	return false
}

func (htw *HumanTaskWorkflow) createOutput(input HumanTaskInput, startTime time.Time, status HumanTaskWorkflowStatus) HumanTaskOutput {
	endTime := time.Now() // Use regular time.Now() since this is not in workflow context
	
	var pendingTasks []HumanTask
	for _, task := range htw.tasks {
		if task.Status == TaskStatusPending || task.Status == TaskStatusAssigned || task.Status == TaskStatusInProgress {
			pendingTasks = append(pendingTasks, task)
		}
	}
	
	return HumanTaskOutput{
		WorkflowID:     input.WorkflowID,
		CompletedTasks: htw.completedTasks,
		EscalatedTasks: htw.escalatedTasks,
		PendingTasks:   pendingTasks,
		Status:         status,
		StartTime:      startTime,
		EndTime:        endTime,
		TotalDuration:  endTime.Sub(startTime),
	}
}

// Query and signal setup methods

func (htw *HumanTaskWorkflow) setupTaskQueries(ctx workflow.Context) {
	workflow.SetQueryHandler(ctx, "getTasks", func() ([]HumanTask, error) {
		return htw.tasks, nil
	})
	
	workflow.SetQueryHandler(ctx, "getCompletedTasks", func() ([]CompletedTask, error) {
		return htw.completedTasks, nil
	})
	
	workflow.SetQueryHandler(ctx, "getEscalatedTasks", func() ([]EscalatedTask, error) {
		return htw.escalatedTasks, nil
	})
	
	workflow.SetQueryHandler(ctx, "getTaskStatus", func(taskID string) (*HumanTask, error) {
		for _, task := range htw.tasks {
			if task.Definition.ID == taskID {
				return &task, nil
			}
		}
		return nil, fmt.Errorf("task not found: %s", taskID)
	})
}

func (htw *HumanTaskWorkflow) setupTaskSignals(ctx workflow.Context) (workflow.ReceiveChannel, workflow.ReceiveChannel, workflow.ReceiveChannel) {
	taskCompleteSignal := workflow.GetSignalChannel(ctx, "taskComplete")
	taskReassignSignal := workflow.GetSignalChannel(ctx, "taskReassign")
	cancelSignal := workflow.GetSignalChannel(ctx, "cancel")
	
	return taskCompleteSignal, taskReassignSignal, cancelSignal
}

// Signal payload types

type TaskCompletionSignal struct {
	TaskID      string     `json:"task_id"`
	CompletedBy string     `json:"completed_by"`
	Result      TaskResult `json:"result"`
	Notes       string     `json:"notes,omitempty"`
}

type TaskReassignmentSignal struct {
	TaskID      string `json:"task_id"`
	NewAssignee string `json:"new_assignee"`
	Reason      string `json:"reason"`
}

// Activity input types

// HumanTaskCreationInput represents input for creating an individual human task
type HumanTaskCreationInput struct {
	TaskID      string                 `json:"task_id"`
	TaskType    string                 `json:"task_type"`
	AssignedTo  string                 `json:"assigned_to"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Deadline    time.Time              `json:"deadline"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

type ReassignTaskInput struct {
	TaskID      string `json:"task_id"`
	OldAssignee string `json:"old_assignee"`
	NewAssignee string `json:"new_assignee"`
	Reason      string `json:"reason"`
}

type EscalateTaskInput struct {
	TaskID     string `json:"task_id"`
	EscalateTo string `json:"escalate_to"`
	Level      int    `json:"level"`
	Reason     string `json:"reason"`
}

// Activity implementations (stubs)

func CreateHumanTaskForWorkflowActivity(ctx context.Context, input HumanTaskCreationInput) error {
	// Implementation would create task in external system
	return nil
}

func ReassignHumanTaskActivity(ctx context.Context, input ReassignTaskInput) error {
	// Implementation would reassign task in external system
	return nil
}

func EscalateHumanTaskActivity(ctx context.Context, input EscalateTaskInput) error {
	// Implementation would escalate task in external system
	return nil
}