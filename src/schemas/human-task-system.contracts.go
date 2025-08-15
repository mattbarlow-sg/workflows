// Package schemas provides validation contracts for the human task system
package schemas

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Validation errors
var (
	// ErrInvalidTaskType indicates an unrecognized task type
	ErrInvalidTaskType = errors.New("invalid task type")

	// ErrContextMissing indicates required context fields are missing
	ErrContextMissing = errors.New("required context fields missing")

	// ErrTaskNotFound indicates the task does not exist
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidState indicates an invalid state transition was attempted
	ErrInvalidState = errors.New("invalid state transition")

	// ErrProofRequired indicates proof is required for completion
	ErrProofRequired = errors.New("proof required for completion")

	// ErrValidationFailed indicates AI validation of proof failed
	ErrValidationFailed = errors.New("proof validation failed")

	// ErrOperatorBusy indicates operator is already reviewing another task
	ErrOperatorBusy = errors.New("operator already busy with another task")

	// ErrQueueEmpty indicates no tasks are available in the queue
	ErrQueueEmpty = errors.New("no tasks available in queue")

	// ErrInvalidTransition indicates a forbidden state transition
	ErrInvalidTransition = errors.New("forbidden state transition")

	// ErrTaskLocked indicates task is locked by another operator
	ErrTaskLocked = errors.New("task is locked by another operator")

	// ErrEscalationFailed indicates escalation process failed
	ErrEscalationFailed = errors.New("escalation process failed")

	// ErrClarificationLimit indicates maximum clarification rounds exceeded
	ErrClarificationLimit = errors.New("maximum clarification rounds exceeded")
)

// TaskQueueContract defines the contract for task queue operations
type TaskQueueContract interface {
	// CreateTask adds a new task to the queue (LIFO)
	CreateTask(ctx context.Context, req CreateTaskRequest) (*TaskRecord, error)

	// GetNextTask retrieves the next task for an operator (LIFO order)
	GetNextTask(ctx context.Context, operatorID string) (*TaskRecord, error)

	// SubmitDecision submits an operator's decision on a task
	SubmitDecision(ctx context.Context, decision OperatorDecision) (*TaskResponse, error)

	// RequestClarification requests additional information for a task
	RequestClarification(ctx context.Context, taskID string, question string, operatorID string) error

	// ProvideClarification provides response to a clarification request
	ProvideClarification(ctx context.Context, taskID string, response string) error

	// GetQueueState returns the current state of the queue
	GetQueueState(ctx context.Context) (*QueueState, error)

	// GetTask retrieves a specific task by ID
	GetTask(ctx context.Context, taskID string) (*TaskRecord, error)

	// FilterTasks retrieves tasks matching the filter criteria
	FilterTasks(ctx context.Context, filter TaskFilter) ([]TaskRecord, error)

	// CancelTask cancels a task with reason
	CancelTask(ctx context.Context, taskID string, reason string, operatorID string) error
}

// ProofValidatorContract defines the contract for AI proof validation
type ProofValidatorContract interface {
	// ValidateProof validates proof sufficiency using AI
	ValidateProof(ctx context.Context, taskType string, proof string) (*AIValidation, error)

	// GetValidationConfidenceThreshold returns the minimum confidence required
	GetValidationConfidenceThreshold() float64

	// IsProofRequired determines if proof is required for task type
	IsProofRequired(taskType string) bool
}

// EscalationManagerContract defines the contract for escalation management
type EscalationManagerContract interface {
	// CheckEscalations checks and marks tasks for escalation
	CheckEscalations(ctx context.Context) ([]string, error)

	// GetEscalationThreshold returns the duration before escalation
	GetEscalationThreshold() time.Duration

	// NotifyEscalation sends notifications for escalated tasks
	NotifyEscalation(ctx context.Context, taskID string) error

	// GetEscalatedTasks returns all currently escalated tasks
	GetEscalatedTasks(ctx context.Context) ([]TaskRecord, error)
}

// StateValidatorContract defines the contract for state transition validation
type StateValidatorContract interface {
	// ValidateTransition checks if a state transition is allowed
	ValidateTransition(from TaskStatus, to TaskStatus) error

	// GetAllowedTransitions returns valid transitions from a state
	GetAllowedTransitions(from TaskStatus) []TaskStatus

	// IsTerminalState checks if a state is terminal
	IsTerminalState(status TaskStatus) bool

	// RequiresProof checks if transition requires proof
	RequiresProof(from TaskStatus, to TaskStatus) bool
}

// ValidateTaskRecord validates a task record for consistency
func ValidateTaskRecord(task *TaskRecord) error {
	if task == nil {
		return errors.New("task record is nil")
	}

	// Validate required fields
	if task.ID == "" {
		return errors.New("task ID is required")
	}

	if task.Type == "" {
		return ErrInvalidTaskType
	}

	if task.Context == nil {
		return ErrContextMissing
	}

	// Validate status
	if !IsValidTaskStatus(task.Status) {
		return fmt.Errorf("invalid task status: %s", task.Status)
	}

	// Validate escalation consistency
	if task.Escalated && task.EscalatedAt == nil {
		return errors.New("escalated task must have escalation timestamp")
	}

	if !task.Escalated && task.EscalatedAt != nil {
		return errors.New("non-escalated task should not have escalation timestamp")
	}

	// Validate clarification count
	if task.ClarificationCount < 0 {
		return errors.New("clarification count cannot be negative")
	}

	// Validate history consistency
	if len(task.History) == 0 {
		return errors.New("task must have at least one history entry")
	}

	// Validate proof for completed/approved tasks
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusApproved {
		if task.Proof == nil || task.Proof.Content == "" {
			return ErrProofRequired
		}
	}

	return nil
}

// IsValidTaskStatus checks if a task status is valid
func IsValidTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending,
		TaskStatusInReview,
		TaskStatusAwaitingClarification,
		TaskStatusCompleted,
		TaskStatusApproved,
		TaskStatusRejected,
		TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidTaskType checks if a task type is valid
func IsValidTaskType(taskType string) bool {
	switch taskType {
	case "approval", "review", "data_entry", "decision", "other":
		return true
	default:
		return false
	}
}

// ValidateStateTransition validates a state transition
func ValidateStateTransition(from, to TaskStatus) error {
	// Define allowed transitions
	allowedTransitions := map[TaskStatus][]TaskStatus{
		TaskStatusPending: {
			TaskStatusInReview,
		},
		TaskStatusInReview: {
			TaskStatusCompleted,
			TaskStatusApproved,
			TaskStatusRejected,
			TaskStatusCancelled,
			TaskStatusAwaitingClarification,
		},
		TaskStatusAwaitingClarification: {
			TaskStatusInReview,
		},
		TaskStatusApproved: {
			TaskStatusCancelled,
		},
		// Terminal states - no transitions allowed
		TaskStatusCompleted: {},
		TaskStatusRejected:  {},
		TaskStatusCancelled: {},
	}

	allowed, exists := allowedTransitions[from]
	if !exists {
		return fmt.Errorf("%w: unknown state %s", ErrInvalidState, from)
	}

	for _, validTo := range allowed {
		if validTo == to {
			return nil
		}
	}

	// Check for explicitly forbidden transitions
	if from == TaskStatusCancelled {
		return fmt.Errorf("%w: cancelled tasks cannot be reopened", ErrInvalidTransition)
	}

	if from == TaskStatusApproved && to == TaskStatusRejected {
		return fmt.Errorf("%w: approved tasks cannot be rejected", ErrInvalidTransition)
	}

	if IsTerminalState(from) {
		return fmt.Errorf("%w: cannot transition from terminal state %s", ErrInvalidTransition, from)
	}

	return fmt.Errorf("%w: transition from %s to %s not allowed", ErrInvalidState, from, to)
}

// IsTerminalState checks if a status is terminal
func IsTerminalState(status TaskStatus) bool {
	switch status {
	case TaskStatusCompleted, TaskStatusRejected, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// ValidateOperatorDecision validates an operator's decision
func ValidateOperatorDecision(decision *OperatorDecision, task *TaskRecord) error {
	if decision == nil {
		return errors.New("decision is nil")
	}

	if task == nil {
		return errors.New("task is nil")
	}

	// Validate decision fields
	if decision.TaskID == "" {
		return errors.New("task ID is required in decision")
	}

	if decision.TaskID != task.ID {
		return errors.New("decision task ID does not match task")
	}

	if decision.OperatorID == "" {
		return errors.New("operator ID is required")
	}

	// Validate decision type
	switch decision.Decision {
	case "complete":
		if decision.Proof == "" {
			return ErrProofRequired
		}
		return ValidateStateTransition(task.Status, TaskStatusCompleted)

	case "approve":
		return ValidateStateTransition(task.Status, TaskStatusApproved)

	case "reject":
		if decision.Reason == "" {
			return errors.New("reason required for rejection")
		}
		return ValidateStateTransition(task.Status, TaskStatusRejected)

	case "return":
		if decision.Reason == "" {
			return errors.New("clarification question required")
		}
		return ValidateStateTransition(task.Status, TaskStatusAwaitingClarification)

	case "cancel":
		if decision.Reason == "" {
			return errors.New("reason required for cancellation")
		}
		return ValidateStateTransition(task.Status, TaskStatusCancelled)

	default:
		return fmt.Errorf("invalid decision type: %s", decision.Decision)
	}
}

// ValidateCreateTaskRequest validates a task creation request
func ValidateCreateTaskRequest(req *CreateTaskRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}

	if !IsValidTaskType(req.Type) {
		return fmt.Errorf("%w: %s", ErrInvalidTaskType, req.Type)
	}

	if req.Context == nil || len(req.Context) == 0 {
		return ErrContextMissing
	}

	// Validate deadline if provided
	if req.Deadline != nil && req.Deadline.Before(time.Now()) {
		return errors.New("deadline cannot be in the past")
	}

	// Validate priority hint if provided
	if req.PriorityHint != "" {
		validPriorities := map[string]bool{
			"low": true, "medium": true, "high": true, "critical": true,
		}
		if !validPriorities[req.PriorityHint] {
			return fmt.Errorf("invalid priority hint: %s", req.PriorityHint)
		}
	}

	return nil
}

// ValidateProofRecord validates a proof record
func ValidateProofRecord(proof *ProofRecord) error {
	if proof == nil {
		return errors.New("proof record is nil")
	}

	if proof.Content == "" {
		return errors.New("proof content cannot be empty")
	}

	// Validate proof type if specified
	if proof.ProofType != "" {
		validTypes := map[string]bool{
			"url": true, "screenshot": true, "document": true,
			"certificate": true, "other": true,
		}
		if !validTypes[proof.ProofType] {
			return fmt.Errorf("invalid proof type: %s", proof.ProofType)
		}
	}

	// Validate AI validation if present
	if proof.ValidationResult != nil {
		if proof.ValidationResult.Confidence < 0 || proof.ValidationResult.Confidence > 1 {
			return errors.New("validation confidence must be between 0 and 1")
		}
	}

	return nil
}

// ValidateQueueInvariant validates queue LIFO ordering invariant
func ValidateQueueInvariant(tasks []TaskRecord) error {
	if len(tasks) <= 1 {
		return nil
	}

	// Check LIFO ordering: newer tasks should come first
	for i := 0; i < len(tasks)-1; i++ {
		if tasks[i].CreatedAt.Before(tasks[i+1].CreatedAt) {
			return fmt.Errorf("queue invariant violated: task %s (created %v) comes before task %s (created %v)",
				tasks[i].ID, tasks[i].CreatedAt,
				tasks[i+1].ID, tasks[i+1].CreatedAt)
		}
	}

	return nil
}

// ValidateSingleOperatorInvariant validates single operator constraint
func ValidateSingleOperatorInvariant(tasks []TaskRecord) error {
	var operatorBusy *string

	for _, task := range tasks {
		if task.Status == TaskStatusInReview {
			if operatorBusy != nil {
				return fmt.Errorf("multiple tasks in review: %s and previous task", task.ID)
			}
			if task.ReviewingOperator == nil {
				return fmt.Errorf("task %s in review but no operator assigned", task.ID)
			}
			operatorBusy = task.ReviewingOperator
		}
	}

	return nil
}

// ValidateEscalationInvariant validates escalation persistence
func ValidateEscalationInvariant(task *TaskRecord) error {
	if !task.Escalated {
		return nil
	}

	if task.EscalatedAt == nil {
		return errors.New("escalated task must have escalation timestamp")
	}

	// Check that escalation timestamp is reasonable
	age := time.Since(task.CreatedAt)
	escalationThreshold := 14 * 24 * time.Hour // 14 days

	if age < escalationThreshold {
		return fmt.Errorf("task escalated too early: age %v < threshold %v", age, escalationThreshold)
	}

	return nil
}

// ValidateHistoryImmutability validates history immutability
func ValidateHistoryImmutability(history []HistoryEntry) error {
	if len(history) == 0 {
		return errors.New("history cannot be empty")
	}

	// Check chronological ordering
	for i := 0; i < len(history)-1; i++ {
		if history[i].Timestamp.After(history[i+1].Timestamp) {
			return fmt.Errorf("history not in chronological order at index %d", i)
		}
	}

	// Check state transition consistency
	for i := 0; i < len(history)-1; i++ {
		if history[i].ToState != history[i+1].FromState {
			return fmt.Errorf("history state mismatch at index %d: %s != %s",
				i, history[i].ToState, history[i+1].FromState)
		}
	}

	return nil
}