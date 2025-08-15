// Package schemas provides type definitions for the human task management system
package schemas

import (
	"time"
)

// TaskRecord represents a complete task record with all metadata
type TaskRecord struct {
	// Unique task identifier (UUID)
	ID string `json:"id" validate:"required,uuid"`

	// Task type category
	Type string `json:"type" validate:"required,oneof=approval review data_entry decision other"`

	// Current task status
	Status TaskStatus `json:"status" validate:"required"`

	// Task-specific context data
	Context map[string]interface{} `json:"context" validate:"required"`

	// Proof of completion if provided
	Proof *ProofRecord `json:"proof,omitempty"`

	// Task creation timestamp
	CreatedAt time.Time `json:"createdAt" validate:"required"`

	// Whether task is escalated
	Escalated bool `json:"escalated"`

	// Escalation timestamp if escalated
	EscalatedAt *time.Time `json:"escalatedAt,omitempty"`

	// Currently reviewing operator
	ReviewingOperator *string `json:"reviewingOperator,omitempty"`

	// Number of clarification cycles
	ClarificationCount int `json:"clarificationCount" validate:"min=0"`

	// Immutable history of all state changes
	History []HistoryEntry `json:"history" validate:"required,dive"`

	// Optional deadline for tracking
	Deadline *time.Time `json:"deadline,omitempty"`

	// ID of task requester
	RequesterID string `json:"requesterId,omitempty"`

	// Last updated timestamp
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`

	// Clarification details if awaiting
	ClarificationRequest *ClarificationDetails `json:"clarificationRequest,omitempty"`
}

// TaskStatus represents the enumeration of task states
type TaskStatus string

const (
	// TaskStatusPending - Task waiting in queue for operator
	TaskStatusPending TaskStatus = "pending"

	// TaskStatusInReview - Task being actively reviewed by operator
	TaskStatusInReview TaskStatus = "in_review"

	// TaskStatusAwaitingClarification - Task requires additional information
	TaskStatusAwaitingClarification TaskStatus = "awaiting_clarification"

	// TaskStatusCompleted - Task successfully completed with proof
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusApproved - Task approved by operator
	TaskStatusApproved TaskStatus = "approved"

	// TaskStatusRejected - Task rejected by operator
	TaskStatusRejected TaskStatus = "rejected"

	// TaskStatusCancelled - Task cancelled
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ProofRecord represents proof submission with validation
type ProofRecord struct {
	// Proof content (URL, text, etc.)
	Content string `json:"content" validate:"required,min=1"`

	// AI validation result
	ValidationResult *AIValidation `json:"validationResult,omitempty"`

	// Proof submission timestamp
	SubmittedAt time.Time `json:"submittedAt" validate:"required"`

	// Type of proof provided
	ProofType string `json:"proofType,omitempty" validate:"omitempty,oneof=url screenshot document certificate other"`
}

// AIValidation represents AI validation result for proof
type AIValidation struct {
	// Whether proof is sufficient
	Valid bool `json:"valid"`

	// Confidence score (0-1)
	Confidence float64 `json:"confidence" validate:"min=0,max=1"`

	// AI's reasoning for decision
	Reasoning string `json:"reasoning,omitempty"`

	// Validation timestamp
	ValidatedAt time.Time `json:"validatedAt" validate:"required"`

	// Model version used for validation
	ModelVersion string `json:"modelVersion,omitempty"`
}

// QueueState represents the current state of the task queue
type QueueState struct {
	// All tasks in queue (LIFO order)
	Tasks []TaskRecord `json:"tasks" validate:"required,dive"`

	// Current operator status
	OperatorStatus string `json:"operatorStatus" validate:"required,oneof=available busy offline"`

	// Queue statistics
	Stats QueueStats `json:"stats" validate:"required"`

	// Currently active task ID
	ActiveTaskID *string `json:"activeTaskId,omitempty"`

	// Last activity timestamp
	LastActivity time.Time `json:"lastActivity"`
}

// QueueStats provides statistics about the queue
type QueueStats struct {
	// Total tasks in queue
	TotalTasks int `json:"totalTasks"`

	// Tasks awaiting review
	PendingTasks int `json:"pendingTasks"`

	// Number of escalated tasks
	EscalatedTasks int `json:"escalatedTasks"`

	// Tasks in review
	InReviewTasks int `json:"inReviewTasks"`

	// Tasks awaiting clarification
	AwaitingClarificationTasks int `json:"awaitingClarificationTasks"`

	// Average task age
	AverageAge time.Duration `json:"averageAge"`

	// Creation time of oldest task
	OldestTask *time.Time `json:"oldestTask,omitempty"`

	// Creation time of newest task
	NewestTask *time.Time `json:"newestTask,omitempty"`

	// Completed tasks in last 24 hours
	CompletedLast24h int `json:"completedLast24h"`

	// Average completion time
	AverageCompletionTime time.Duration `json:"averageCompletionTime"`
}

// CreateTaskRequest represents a request to create a new task
type CreateTaskRequest struct {
	// Task type
	Type string `json:"type" validate:"required,oneof=approval review data_entry decision other"`

	// Task-specific data
	Context map[string]interface{} `json:"context" validate:"required"`

	// ID of task requester
	RequesterID string `json:"requesterId,omitempty"`

	// Optional deadline (for tracking only)
	Deadline *time.Time `json:"deadline,omitempty"`

	// Priority hint (not used for ordering, informational only)
	PriorityHint string `json:"priorityHint,omitempty" validate:"omitempty,oneof=low medium high critical"`

	// Additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TaskResponse represents the response after task decision
type TaskResponse struct {
	// Task identifier
	TaskID string `json:"taskId" validate:"required,uuid"`

	// Final task status
	Outcome TaskStatus `json:"outcome" validate:"required"`

	// Proof if provided
	Proof *string `json:"proof,omitempty"`

	// Decision timestamp
	Timestamp time.Time `json:"timestamp" validate:"required"`

	// Operator who made the decision
	OperatorID string `json:"operatorId" validate:"required"`

	// Reason for decision
	Reason string `json:"reason,omitempty"`

	// Whether task was escalated when decided
	WasEscalated bool `json:"wasEscalated"`
}

// OperatorDecision represents an operator's decision on a task
type OperatorDecision struct {
	// Task to decide on
	TaskID string `json:"taskId" validate:"required,uuid"`

	// Decision type
	Decision string `json:"decision" validate:"required,oneof=complete approve reject return cancel"`

	// Proof for completion
	Proof string `json:"proof,omitempty"`

	// Reason for decision
	Reason string `json:"reason,omitempty"`

	// Deciding operator
	OperatorID string `json:"operatorId" validate:"required"`

	// Additional decision metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// HistoryEntry represents an immutable history record
type HistoryEntry struct {
	// When the event occurred
	Timestamp time.Time `json:"timestamp" validate:"required"`

	// Previous state
	FromState TaskStatus `json:"fromState" validate:"required"`

	// New state
	ToState TaskStatus `json:"toState" validate:"required"`

	// Who triggered the change
	Actor string `json:"actor,omitempty"`

	// Why the change occurred
	Reason string `json:"reason,omitempty"`

	// Additional event details
	Details map[string]interface{} `json:"details,omitempty"`

	// Event type for categorization
	EventType string `json:"eventType" validate:"required,oneof=state_change escalation clarification decision system"`
}

// ClarificationDetails contains information about clarification requests
type ClarificationDetails struct {
	// Clarification question
	Question string `json:"question" validate:"required,min=1"`

	// When clarification was requested
	RequestedAt time.Time `json:"requestedAt" validate:"required"`

	// Who requested clarification
	RequestedBy string `json:"requestedBy" validate:"required"`

	// Response to clarification
	Response *string `json:"response,omitempty"`

	// When response was provided
	RespondedAt *time.Time `json:"respondedAt,omitempty"`

	// Current clarification round number
	RoundNumber int `json:"roundNumber" validate:"min=1"`
}

// EscalationEvent represents a task escalation event
type EscalationEvent struct {
	// Task that was escalated
	TaskID string `json:"taskId" validate:"required,uuid"`

	// When escalation occurred
	EscalatedAt time.Time `json:"escalatedAt" validate:"required"`

	// Task age when escalated
	TaskAge time.Duration `json:"taskAge" validate:"required"`

	// Reason for escalation
	Reason string `json:"reason" validate:"required"`

	// Notifications sent
	NotificationsSent []NotificationRecord `json:"notificationsSent,omitempty"`
}

// NotificationRecord tracks notifications sent
type NotificationRecord struct {
	// Recipient of notification
	Recipient string `json:"recipient" validate:"required"`

	// Type of notification
	Type string `json:"type" validate:"required,oneof=email slack webhook"`

	// When notification was sent
	SentAt time.Time `json:"sentAt" validate:"required"`

	// Whether notification was successful
	Success bool `json:"success"`

	// Error if notification failed
	Error string `json:"error,omitempty"`
}

// TaskFilter provides filtering options for task queries
type TaskFilter struct {
	// Filter by status
	Status []TaskStatus `json:"status,omitempty"`

	// Filter by type
	Type []string `json:"type,omitempty"`

	// Filter by escalation status
	Escalated *bool `json:"escalated,omitempty"`

	// Filter by operator
	ReviewingOperator *string `json:"reviewingOperator,omitempty"`

	// Filter by creation date range
	CreatedAfter *time.Time `json:"createdAfter,omitempty"`
	CreatedBefore *time.Time `json:"createdBefore,omitempty"`

	// Filter by requester
	RequesterID *string `json:"requesterId,omitempty"`

	// Maximum results to return
	Limit int `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// TaskBatch represents a batch of tasks for bulk operations
type TaskBatch struct {
	// Tasks in the batch
	Tasks []TaskRecord `json:"tasks" validate:"required,min=1,dive"`

	// Batch operation type
	Operation string `json:"operation" validate:"required,oneof=cancel escalate archive"`

	// Reason for batch operation
	Reason string `json:"reason" validate:"required"`

	// Who initiated the batch operation
	InitiatedBy string `json:"initiatedBy" validate:"required"`

	// When batch operation was initiated
	InitiatedAt time.Time `json:"initiatedAt" validate:"required"`
}

// QueueConfiguration defines queue behavior configuration
type QueueConfiguration struct {
	// Escalation threshold duration
	EscalationThreshold time.Duration `json:"escalationThreshold" validate:"required"`

	// Maximum clarification rounds allowed
	MaxClarificationRounds int `json:"maxClarificationRounds" validate:"min=0"`

	// Enable AI validation
	EnableAIValidation bool `json:"enableAIValidation"`

	// AI validation confidence threshold
	AIConfidenceThreshold float64 `json:"aiConfidenceThreshold" validate:"min=0,max=1"`

	// Queue processing mode
	ProcessingMode string `json:"processingMode" validate:"required,oneof=lifo fifo"`

	// Enable notifications
	EnableNotifications bool `json:"enableNotifications"`

	// Notification channels
	NotificationChannels []string `json:"notificationChannels,omitempty"`
}

// OperatorSession represents an operator's work session
type OperatorSession struct {
	// Session ID
	SessionID string `json:"sessionId" validate:"required,uuid"`

	// Operator ID
	OperatorID string `json:"operatorId" validate:"required"`

	// Session start time
	StartedAt time.Time `json:"startedAt" validate:"required"`

	// Session end time
	EndedAt *time.Time `json:"endedAt,omitempty"`

	// Tasks processed in session
	TasksProcessed []string `json:"tasksProcessed"`

	// Session statistics
	Stats SessionStats `json:"stats"`
}

// SessionStats provides statistics for an operator session
type SessionStats struct {
	// Total tasks reviewed
	TasksReviewed int `json:"tasksReviewed"`

	// Tasks completed
	TasksCompleted int `json:"tasksCompleted"`

	// Tasks approved
	TasksApproved int `json:"tasksApproved"`

	// Tasks rejected
	TasksRejected int `json:"tasksRejected"`

	// Tasks cancelled
	TasksCancelled int `json:"tasksCancelled"`

	// Clarifications requested
	ClarificationsRequested int `json:"clarificationsRequested"`

	// Average review time per task
	AverageReviewTime time.Duration `json:"averageReviewTime"`

	// Total session duration
	SessionDuration time.Duration `json:"sessionDuration"`
}