// Package workflows provides approval workflow implementation
package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ApprovalWorkflow implements a human approval workflow with escalation
type ApprovalWorkflow struct {
	*BaseWorkflowImpl
	approvals     []ApprovalRecord
	currentStep   int
	escalationCount int
}

// ApprovalInput defines the input for the approval workflow
type ApprovalInput struct {
	RequestID       string        `json:"request_id"`
	RequestType     string        `json:"request_type"`
	Requester       string        `json:"requester"`
	Description     string        `json:"description"`
	ApprovalSteps   []ApprovalStep `json:"approval_steps"`
	EscalationChain []string      `json:"escalation_chain"`
	Priority        Priority      `json:"priority"`
	Deadline        time.Duration `json:"deadline"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

// ApprovalOutput defines the output of the approval workflow
type ApprovalOutput struct {
	RequestID      string           `json:"request_id"`
	Status         ApprovalStatus   `json:"status"`
	Approvals      []ApprovalRecord `json:"approvals"`
	FinalApprover  string           `json:"final_approver"`
	CompletedAt    time.Time        `json:"completed_at"`
	TotalDuration  time.Duration    `json:"total_duration"`
}

// ApprovalStep defines a step in the approval process
type ApprovalStep struct {
	Name       string        `json:"name"`
	Approvers  []string      `json:"approvers"`
	Required   int           `json:"required"` // Number of approvers required (0 = all)
	Timeout    time.Duration `json:"timeout"`
	Sequential bool          `json:"sequential"` // If true, approvers must approve in order
}

// ApprovalRecord tracks an individual approval decision
type ApprovalRecord struct {
	StepName     string            `json:"step_name"`
	Approver     string            `json:"approver"`
	Decision     ApprovalDecision  `json:"decision"`
	Timestamp    time.Time         `json:"timestamp"`
	Comments     string            `json:"comments"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ApprovalDecision represents the approval decision
type ApprovalDecision string

const (
	ApprovalDecisionApproved ApprovalDecision = "approved"
	ApprovalDecisionRejected ApprovalDecision = "rejected"
	ApprovalDecisionSkipped  ApprovalDecision = "skipped"
)

// ApprovalStatus represents the overall approval status
type ApprovalStatus string

const (
	ApprovalStatusPending    ApprovalStatus = "pending"
	ApprovalStatusApproved   ApprovalStatus = "approved"
	ApprovalStatusRejected   ApprovalStatus = "rejected"
	ApprovalStatusEscalated  ApprovalStatus = "escalated"
	ApprovalStatusExpired    ApprovalStatus = "expired"
	ApprovalStatusCancelled  ApprovalStatus = "cancelled"
)

// Priority defines the priority levels for approvals
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// NewApprovalWorkflow creates a new approval workflow
func NewApprovalWorkflow() *ApprovalWorkflow {
	metadata := WorkflowMetadata{
		Name:        "ApprovalWorkflow",
		Version:     "1.0.0",
		Description: "Human approval workflow with escalation support",
		Tags: map[string]string{
			"type":     "approval",
			"category": "human-task",
		},
		Author:    "Temporal Workflows",
		CreatedAt: time.Now(),
	}

	return &ApprovalWorkflow{
		BaseWorkflowImpl: NewBaseWorkflow(metadata),
		approvals:        []ApprovalRecord{},
		currentStep:      0,
		escalationCount:  0,
	}
}

// Execute runs the approval workflow
func (aw *ApprovalWorkflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Type assertion for input
	approvalInput, ok := input.(ApprovalInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for ApprovalWorkflow")
	}

	// Validate input
	if err := aw.Validate(approvalInput); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting approval workflow", "requestID", approvalInput.RequestID, "requestType", approvalInput.RequestType)

	startTime := workflow.Now(ctx)
	aw.UpdateState("initialization", 0, WorkflowStatusRunning)

	// Setup base queries and signals
	if err := aw.SetupBaseQueries(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup base queries: %w", err)
	}
	aw.SetupBaseSignals(ctx)

	// Setup approval-specific query handlers
	err := workflow.SetQueryHandler(ctx, "getApprovals", func() ([]ApprovalRecord, error) {
		return aw.approvals, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup approvals query handler: %w", err)
	}

	// Setup approval signal channel
	approvalSignal := workflow.GetSignalChannel(ctx, "approval")

	// Setup cancellation signal channel
	cancelSignal := workflow.GetSignalChannel(ctx, "cancel")

	// Configure activity options based on priority
	ao := aw.getActivityOptions(approvalInput.Priority)
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Process each approval step
	for stepIndex, step := range approvalInput.ApprovalSteps {
		aw.currentStep = stepIndex
		aw.UpdateState(fmt.Sprintf("step_%d_%s", stepIndex, step.Name), 
			(stepIndex*100)/len(approvalInput.ApprovalSteps), WorkflowStatusRunning)

		logger.Info("Processing approval step", "step", step.Name, "approvers", step.Approvers)

		// Create human tasks for this step
		taskIDs, err := aw.createApprovalTasks(ctx, approvalInput, step)
		if err != nil {
			aw.SetError(err)
			return aw.createErrorOutput(approvalInput, err), err
		}

		// Wait for approvals for this step
		stepResult, err := aw.waitForStepApprovals(ctx, step, approvalSignal, cancelSignal)
		if err != nil {
			aw.SetError(err)
			return aw.createErrorOutput(approvalInput, err), err
		}

		// Check if step was cancelled
		if stepResult.Status == ApprovalStatusCancelled {
			logger.Info("Approval workflow cancelled")
			return ApprovalOutput{
				RequestID:     approvalInput.RequestID,
				Status:        ApprovalStatusCancelled,
				Approvals:     aw.approvals,
				CompletedAt:   workflow.Now(ctx),
				TotalDuration: workflow.Now(ctx).Sub(startTime),
			}, nil
		}

		// Check if step was rejected
		if stepResult.Status == ApprovalStatusRejected {
			logger.Info("Approval workflow rejected at step", "step", step.Name)
			return ApprovalOutput{
				RequestID:     approvalInput.RequestID,
				Status:        ApprovalStatusRejected,
				Approvals:     aw.approvals,
				CompletedAt:   workflow.Now(ctx),
				TotalDuration: workflow.Now(ctx).Sub(startTime),
			}, nil
		}

		// Clean up tasks for this step
		if err := aw.cleanupTasks(ctx, taskIDs); err != nil {
			logger.Warn("Failed to cleanup tasks", "error", err)
		}
	}

	// All steps completed successfully
	completedAt := workflow.Now(ctx)
	aw.UpdateState("completed", 100, WorkflowStatusCompleted)

	logger.Info("Approval workflow completed successfully")
	return ApprovalOutput{
		RequestID:     approvalInput.RequestID,
		Status:        ApprovalStatusApproved,
		Approvals:     aw.approvals,
		FinalApprover: aw.getFinalApprover(),
		CompletedAt:   completedAt,
		TotalDuration: completedAt.Sub(startTime),
	}, nil
}

// Validate validates the approval input
func (aw *ApprovalWorkflow) Validate(input interface{}) error {
	approvalInput, ok := input.(ApprovalInput)
	if !ok {
		return fmt.Errorf("input must be of type ApprovalInput")
	}

	if approvalInput.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}

	if approvalInput.RequestType == "" {
		return fmt.Errorf("request type is required")
	}

	if approvalInput.Requester == "" {
		return fmt.Errorf("requester is required")
	}

	if len(approvalInput.ApprovalSteps) == 0 {
		return fmt.Errorf("at least one approval step is required")
	}

	// Validate each approval step
	for i, step := range approvalInput.ApprovalSteps {
		if step.Name == "" {
			return fmt.Errorf("approval step %d name is required", i)
		}

		if len(step.Approvers) == 0 {
			return fmt.Errorf("approval step %d must have at least one approver", i)
		}

		if step.Required < 0 || (step.Required > 0 && step.Required > len(step.Approvers)) {
			return fmt.Errorf("approval step %d required count is invalid", i)
		}

		if step.Timeout <= 0 {
			return fmt.Errorf("approval step %d timeout must be positive", i)
		}
	}

	return nil
}

// createApprovalTasks creates human tasks for the approval step
func (aw *ApprovalWorkflow) createApprovalTasks(ctx workflow.Context, input ApprovalInput, step ApprovalStep) ([]string, error) {
	var taskIDs []string

	for _, approver := range step.Approvers {
		taskID := fmt.Sprintf("approval-%s-%s-%s", input.RequestID, step.Name, approver)
		
		task := ApprovalHumanTaskInput{
			TaskID:      taskID,
			TaskType:    "approval",
			AssignedTo:  approver,
			Title:       fmt.Sprintf("Approval Required: %s", input.RequestType),
			Description: fmt.Sprintf("Please review and approve/reject request %s: %s", input.RequestID, input.Description),
			Priority:    string(input.Priority),
			Deadline:    workflow.Now(ctx).Add(step.Timeout),
			Data: map[string]interface{}{
				"requestID":   input.RequestID,
				"requestType": input.RequestType,
				"requester":   input.Requester,
				"stepName":    step.Name,
				"inputData":   input.Data,
			},
		}

		err := workflow.ExecuteActivity(ctx, CreateHumanTaskActivity, task).Get(ctx, nil)
		if err != nil {
			return taskIDs, fmt.Errorf("failed to create task for %s: %w", approver, err)
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

// waitForStepApprovals waits for approvals for a single step
func (aw *ApprovalWorkflow) waitForStepApprovals(ctx workflow.Context, step ApprovalStep, approvalSignal workflow.ReceiveChannel, cancelSignal workflow.ReceiveChannel) (*StepResult, error) {
	logger := workflow.GetLogger(ctx)
	stepApprovals := make(map[string]ApprovalRecord)
	requiredApprovals := step.Required
	if requiredApprovals == 0 {
		requiredApprovals = len(step.Approvers) // All approvers required
	}

	// Setup timeout
	timerCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()
	timer := workflow.NewTimer(timerCtx, step.Timeout)

	for {
		selector := workflow.NewSelector(ctx)
		
		// Handle approval signals
		selector.AddReceive(approvalSignal, func(c workflow.ReceiveChannel, more bool) {
			var approval ApprovalSignal
			c.Receive(ctx, &approval)

			// Check if this approval is for the current step
			if approval.StepName == step.Name {
				record := ApprovalRecord{
					StepName:  step.Name,
					Approver:  approval.Approver,
					Decision:  approval.Decision,
					Timestamp: workflow.Now(ctx),
					Comments:  approval.Comments,
					Metadata:  approval.Metadata,
				}

				stepApprovals[approval.Approver] = record
				aw.approvals = append(aw.approvals, record)

				logger.Info("Received approval", "approver", approval.Approver, "decision", approval.Decision, "step", step.Name)
			}
		})

		// Handle cancellation signals
		selector.AddReceive(cancelSignal, func(c workflow.ReceiveChannel, more bool) {
			var reason string
			c.Receive(ctx, &reason)
			logger.Info("Approval workflow cancelled", "reason", reason)
		})

		// Handle timeout
		selector.AddFuture(timer, func(f workflow.Future) {
			logger.Warn("Approval step timeout", "step", step.Name)
			
			// Escalate if escalation chain exists
			if len(aw.getEscalationChain()) > aw.escalationCount {
				aw.escalateStep(ctx, step)
				return
			}
		})

		selector.Select(ctx)

		// Check if we received a cancellation
		var cancelled bool
		cancelSignal.ReceiveAsync(&cancelled)
		if cancelled {
			return &StepResult{Status: ApprovalStatusCancelled}, nil
		}

		// Check if we have enough approvals
		approvedCount := 0
		rejectedCount := 0
		
		for _, record := range stepApprovals {
			switch record.Decision {
			case ApprovalDecisionApproved:
				approvedCount++
			case ApprovalDecisionRejected:
				rejectedCount++
			}
		}

		// If anyone rejected, the step is rejected
		if rejectedCount > 0 {
			return &StepResult{
				Status:    ApprovalStatusRejected,
				Approvals: stepApprovals,
			}, nil
		}

		// If we have enough approvals, the step is approved
		if approvedCount >= requiredApprovals {
			return &StepResult{
				Status:    ApprovalStatusApproved,
				Approvals: stepApprovals,
			}, nil
		}

		// Continue waiting for more approvals
	}
}

// escalateStep escalates the approval step to the next level
func (aw *ApprovalWorkflow) escalateStep(ctx workflow.Context, step ApprovalStep) {
	logger := workflow.GetLogger(ctx)
	escalationChain := aw.getEscalationChain()
	
	if aw.escalationCount < len(escalationChain) {
		escalateTo := escalationChain[aw.escalationCount]
		aw.escalationCount++
		
		logger.Info("Escalating approval step", "step", step.Name, "escalateTo", escalateTo)
		
		requestID, _ := aw.GetStateData("requestID")
		escalationTask := ApprovalHumanTaskInput{
			TaskID:      fmt.Sprintf("escalation-%s-%s-%d", requestID, step.Name, aw.escalationCount),
			TaskType:    "escalation",
			AssignedTo:  escalateTo,
			Title:       fmt.Sprintf("Escalated Approval: %s", step.Name),
			Description: fmt.Sprintf("This approval has been escalated due to timeout. Please review urgently."),
			Priority:    string(PriorityCritical),
			Deadline:    workflow.Now(ctx).Add(step.Timeout / 2), // Shorter deadline for escalation
		}
		
		workflow.ExecuteActivity(ctx, CreateHumanTaskActivity, escalationTask).Get(ctx, nil)
	}
}

// cleanupTasks cleans up completed tasks
func (aw *ApprovalWorkflow) cleanupTasks(ctx workflow.Context, taskIDs []string) error {
	for _, taskID := range taskIDs {
		err := workflow.ExecuteActivity(ctx, CleanupHumanTaskActivity, taskID).Get(ctx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// getActivityOptions returns activity options based on priority
func (aw *ApprovalWorkflow) getActivityOptions(priority Priority) workflow.ActivityOptions {
	switch priority {
	case PriorityCritical:
		return workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Minute,
			HeartbeatTimeout:    30 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    5 * time.Second,
				BackoffCoefficient: 1.5,
				MaximumInterval:    30 * time.Second,
				MaximumAttempts:    5,
			},
		}
	case PriorityHigh:
		return workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Minute,
			HeartbeatTimeout:    time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    10 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    time.Minute,
				MaximumAttempts:    3,
			},
		}
	default:
		return HumanTaskActivityOptions()
	}
}

// getEscalationChain returns the escalation chain
func (aw *ApprovalWorkflow) getEscalationChain() []string {
	if chain, exists := aw.GetStateData("escalationChain"); exists {
		if escalationChain, ok := chain.([]string); ok {
			return escalationChain
		}
	}
	return []string{}
}

// getFinalApprover returns the final approver
func (aw *ApprovalWorkflow) getFinalApprover() string {
	if len(aw.approvals) == 0 {
		return ""
	}
	// Return the last approver who approved
	for i := len(aw.approvals) - 1; i >= 0; i-- {
		if aw.approvals[i].Decision == ApprovalDecisionApproved {
			return aw.approvals[i].Approver
		}
	}
	return ""
}

// createErrorOutput creates an error output
func (aw *ApprovalWorkflow) createErrorOutput(input ApprovalInput, err error) ApprovalOutput {
	return ApprovalOutput{
		RequestID:     input.RequestID,
		Status:        ApprovalStatusExpired,
		Approvals:     aw.approvals,
		CompletedAt:   time.Now(),
		TotalDuration: 0,
	}
}

// StepResult represents the result of a single approval step
type StepResult struct {
	Status    ApprovalStatus
	Approvals map[string]ApprovalRecord
}

// ApprovalSignal represents an approval signal payload
type ApprovalSignal struct {
	StepName  string                     `json:"step_name"`
	Approver  string                     `json:"approver"`
	Decision  ApprovalDecision           `json:"decision"`
	Comments  string                     `json:"comments"`
	Metadata  map[string]interface{}     `json:"metadata,omitempty"`
}

// Human Task Activities
type ApprovalHumanTaskInput struct {
	TaskID      string                 `json:"task_id"`
	TaskType    string                 `json:"task_type"`
	AssignedTo  string                 `json:"assigned_to"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Deadline    time.Time              `json:"deadline"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// CreateHumanTaskActivity creates a human task
func CreateHumanTaskActivity(ctx context.Context, input ApprovalHumanTaskInput) error {
	// Implementation would integrate with external task management system
	// For now, return nil (stub implementation)
	return nil
}

// CleanupHumanTaskActivity cleans up a completed human task
func CleanupHumanTaskActivity(ctx context.Context, taskID string) error {
	// Implementation would clean up task from external system
	// For now, return nil (stub implementation)
	return nil
}