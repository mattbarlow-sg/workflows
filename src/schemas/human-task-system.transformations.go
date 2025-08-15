// Package schemas provides data transformations for the human task system
package schemas

import (
	"fmt"
	"sort"
	"time"
)

// TaskTransformations provides transformation utilities for tasks
type TaskTransformations struct{}

// NewTaskTransformations creates a new transformer instance
func NewTaskTransformations() *TaskTransformations {
	return &TaskTransformations{}
}

// CreateNewTask transforms a CreateTaskRequest into a TaskRecord
func (t *TaskTransformations) CreateNewTask(req CreateTaskRequest) *TaskRecord {
	now := time.Now()
	taskID := generateUUID() // Placeholder - would use actual UUID generator

	task := &TaskRecord{
		ID:                 taskID,
		Type:               req.Type,
		Status:             TaskStatusPending,
		Context:            req.Context,
		CreatedAt:          now,
		UpdatedAt:          now,
		Escalated:          false,
		ClarificationCount: 0,
		RequesterID:        req.RequesterID,
		Deadline:           req.Deadline,
		History: []HistoryEntry{
			{
				Timestamp: now,
				FromState: TaskStatusPending, // Initial state
				ToState:   TaskStatusPending,
				Actor:     req.RequesterID,
				Reason:    "Task created",
				EventType: "state_change",
				Details: map[string]interface{}{
					"type":         req.Type,
					"priorityHint": req.PriorityHint,
				},
			},
		},
	}

	// Add metadata if provided
	if len(req.Metadata) > 0 {
		if task.Context == nil {
			task.Context = make(map[string]interface{})
		}
		task.Context["metadata"] = req.Metadata
	}

	return task
}

// TransitionTaskState transforms a task through a state transition
func (t *TaskTransformations) TransitionTaskState(
	task *TaskRecord,
	newStatus TaskStatus,
	actor string,
	reason string,
) (*TaskRecord, error) {
	// Validate transition
	if err := ValidateStateTransition(task.Status, newStatus); err != nil {
		return nil, err
	}

	// Create history entry
	historyEntry := HistoryEntry{
		Timestamp: time.Now(),
		FromState: task.Status,
		ToState:   newStatus,
		Actor:     actor,
		Reason:    reason,
		EventType: "state_change",
		Details:   make(map[string]interface{}),
	}

	// Apply state-specific transformations
	switch newStatus {
	case TaskStatusInReview:
		task.ReviewingOperator = &actor
		historyEntry.Details["operator"] = actor

	case TaskStatusAwaitingClarification:
		task.ClarificationCount++
		task.ReviewingOperator = nil // Release operator
		historyEntry.Details["clarificationRound"] = task.ClarificationCount

	case TaskStatusCompleted, TaskStatusApproved:
		// Ensure proof is present
		if task.Proof == nil {
			return nil, ErrProofRequired
		}
		historyEntry.Details["proofProvided"] = true

	case TaskStatusRejected, TaskStatusCancelled:
		task.ReviewingOperator = nil
		historyEntry.Details["terminal"] = true
	}

	// Update task
	task.Status = newStatus
	task.UpdatedAt = time.Now()
	task.History = append(task.History, historyEntry)

	return task, nil
}

// ApplyOperatorDecision transforms a task based on operator decision
func (t *TaskTransformations) ApplyOperatorDecision(
	task *TaskRecord,
	decision OperatorDecision,
) (*TaskResponse, error) {
	// Validate decision
	if err := ValidateOperatorDecision(&decision, task); err != nil {
		return nil, err
	}

	var newStatus TaskStatus
	var needsProof bool

	// Map decision to status
	switch decision.Decision {
	case "complete":
		newStatus = TaskStatusCompleted
		needsProof = true
	case "approve":
		newStatus = TaskStatusApproved
	case "reject":
		newStatus = TaskStatusRejected
	case "return":
		newStatus = TaskStatusAwaitingClarification
	case "cancel":
		newStatus = TaskStatusCancelled
	default:
		return nil, fmt.Errorf("unknown decision: %s", decision.Decision)
	}

	// Add proof if needed
	if needsProof && decision.Proof != "" {
		task.Proof = &ProofRecord{
			Content:     decision.Proof,
			SubmittedAt: time.Now(),
		}
	}

	// Apply transition
	updatedTask, err := t.TransitionTaskState(task, newStatus, decision.OperatorID, decision.Reason)
	if err != nil {
		return nil, err
	}

	// Create response
	response := &TaskResponse{
		TaskID:       task.ID,
		Outcome:      newStatus,
		Timestamp:    time.Now(),
		OperatorID:   decision.OperatorID,
		Reason:       decision.Reason,
		WasEscalated: task.Escalated,
	}

	if decision.Proof != "" {
		response.Proof = &decision.Proof
	}

	return response, nil
}

// AddClarificationRequest adds a clarification request to a task
func (t *TaskTransformations) AddClarificationRequest(
	task *TaskRecord,
	question string,
	operatorID string,
) (*TaskRecord, error) {
	// Validate current state
	if task.Status != TaskStatusInReview {
		return nil, fmt.Errorf("can only request clarification from in_review state, current: %s", task.Status)
	}

	// Create clarification details
	clarification := &ClarificationDetails{
		Question:    question,
		RequestedAt: time.Now(),
		RequestedBy: operatorID,
		RoundNumber: task.ClarificationCount + 1,
	}

	// Transition to awaiting_clarification
	updatedTask, err := t.TransitionTaskState(
		task,
		TaskStatusAwaitingClarification,
		operatorID,
		fmt.Sprintf("Clarification requested: %s", question),
	)
	if err != nil {
		return nil, err
	}

	updatedTask.ClarificationRequest = clarification
	return updatedTask, nil
}

// ProvideClarificationResponse adds a clarification response to a task
func (t *TaskTransformations) ProvideClarificationResponse(
	task *TaskRecord,
	response string,
) (*TaskRecord, error) {
	// Validate current state
	if task.Status != TaskStatusAwaitingClarification {
		return nil, fmt.Errorf("task is not awaiting clarification, current: %s", task.Status)
	}

	if task.ClarificationRequest == nil {
		return nil, fmt.Errorf("no clarification request found")
	}

	// Add response
	now := time.Now()
	task.ClarificationRequest.Response = &response
	task.ClarificationRequest.RespondedAt = &now

	// Add to history
	task.History = append(task.History, HistoryEntry{
		Timestamp: now,
		FromState: TaskStatusAwaitingClarification,
		ToState:   TaskStatusAwaitingClarification,
		Actor:     "system",
		Reason:    "Clarification response provided",
		EventType: "clarification",
		Details: map[string]interface{}{
			"roundNumber": task.ClarificationRequest.RoundNumber,
			"response":    response,
		},
	})

	task.UpdatedAt = now
	return task, nil
}

// EscalateTask marks a task as escalated
func (t *TaskTransformations) EscalateTask(task *TaskRecord) (*TaskRecord, error) {
	// Check if already escalated
	if task.Escalated {
		return task, nil // Idempotent
	}

	// Check age threshold
	age := time.Since(task.CreatedAt)
	escalationThreshold := 14 * 24 * time.Hour

	if age < escalationThreshold {
		return nil, fmt.Errorf("task too young for escalation: %v < %v", age, escalationThreshold)
	}

	// Mark as escalated
	now := time.Now()
	task.Escalated = true
	task.EscalatedAt = &now
	task.UpdatedAt = now

	// Add history entry
	task.History = append(task.History, HistoryEntry{
		Timestamp: now,
		FromState: task.Status,
		ToState:   task.Status, // Status doesn't change
		Actor:     "system",
		Reason:    fmt.Sprintf("Task escalated after %v", age),
		EventType: "escalation",
		Details: map[string]interface{}{
			"taskAge":   age.String(),
			"threshold": escalationThreshold.String(),
		},
	})

	return task, nil
}

// AddProofValidation adds AI validation result to a proof
func (t *TaskTransformations) AddProofValidation(
	task *TaskRecord,
	validation *AIValidation,
) (*TaskRecord, error) {
	if task.Proof == nil {
		return nil, fmt.Errorf("no proof to validate")
	}

	task.Proof.ValidationResult = validation
	task.UpdatedAt = time.Now()

	// Add history entry
	task.History = append(task.History, HistoryEntry{
		Timestamp: time.Now(),
		FromState: task.Status,
		ToState:   task.Status,
		Actor:     "ai-validator",
		Reason:    fmt.Sprintf("Proof validated: %v (confidence: %.2f)", validation.Valid, validation.Confidence),
		EventType: "system",
		Details: map[string]interface{}{
			"valid":      validation.Valid,
			"confidence": validation.Confidence,
			"reasoning":  validation.Reasoning,
		},
	})

	return task, nil
}

// SortTasksLIFO sorts tasks in LIFO order (newest first)
func (t *TaskTransformations) SortTasksLIFO(tasks []TaskRecord) []TaskRecord {
	sorted := make([]TaskRecord, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		// Newer tasks come first (LIFO)
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	return sorted
}

// FilterTasksByStatus filters tasks by their status
func (t *TaskTransformations) FilterTasksByStatus(tasks []TaskRecord, statuses ...TaskStatus) []TaskRecord {
	if len(statuses) == 0 {
		return tasks
	}

	statusMap := make(map[TaskStatus]bool)
	for _, s := range statuses {
		statusMap[s] = true
	}

	filtered := make([]TaskRecord, 0, len(tasks))
	for _, task := range tasks {
		if statusMap[task.Status] {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// CalculateQueueStats calculates statistics for a queue of tasks
func (t *TaskTransformations) CalculateQueueStats(tasks []TaskRecord) QueueStats {
	stats := QueueStats{
		TotalTasks: len(tasks),
	}

	if len(tasks) == 0 {
		return stats
	}

	var (
		totalAge          time.Duration
		completedLast24h  int
		totalCompletionTime time.Duration
		completionCount   int
	)

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	for _, task := range tasks {
		// Count by status
		switch task.Status {
		case TaskStatusPending:
			stats.PendingTasks++
		case TaskStatusInReview:
			stats.InReviewTasks++
		case TaskStatusAwaitingClarification:
			stats.AwaitingClarificationTasks++
		}

		// Count escalated
		if task.Escalated {
			stats.EscalatedTasks++
		}

		// Calculate age
		age := now.Sub(task.CreatedAt)
		totalAge += age

		// Track oldest/newest
		if stats.OldestTask == nil || task.CreatedAt.Before(*stats.OldestTask) {
			t := task.CreatedAt
			stats.OldestTask = &t
		}
		if stats.NewestTask == nil || task.CreatedAt.After(*stats.NewestTask) {
			t := task.CreatedAt
			stats.NewestTask = &t
		}

		// Count recent completions
		if IsTerminalState(task.Status) && task.UpdatedAt.After(last24h) {
			completedLast24h++
			if task.Status == TaskStatusCompleted || task.Status == TaskStatusApproved {
				completionTime := task.UpdatedAt.Sub(task.CreatedAt)
				totalCompletionTime += completionTime
				completionCount++
			}
		}
	}

	// Calculate averages
	if len(tasks) > 0 {
		stats.AverageAge = totalAge / time.Duration(len(tasks))
	}

	stats.CompletedLast24h = completedLast24h

	if completionCount > 0 {
		stats.AverageCompletionTime = totalCompletionTime / time.Duration(completionCount)
	}

	return stats
}

// CreateSessionStats creates session statistics from processed tasks
func (t *TaskTransformations) CreateSessionStats(
	session *OperatorSession,
	tasks []TaskRecord,
) SessionStats {
	stats := SessionStats{}

	if session.EndedAt != nil {
		stats.SessionDuration = session.EndedAt.Sub(session.StartedAt)
	} else {
		stats.SessionDuration = time.Since(session.StartedAt)
	}

	var totalReviewTime time.Duration
	reviewCount := 0

	for _, taskID := range session.TasksProcessed {
		// Find task in list
		for _, task := range tasks {
			if task.ID == taskID {
				stats.TasksReviewed++

				switch task.Status {
				case TaskStatusCompleted:
					stats.TasksCompleted++
				case TaskStatusApproved:
					stats.TasksApproved++
				case TaskStatusRejected:
					stats.TasksRejected++
				case TaskStatusCancelled:
					stats.TasksCancelled++
				}

				// Count clarifications
				stats.ClarificationsRequested += task.ClarificationCount

				// Calculate review time from history
				reviewTime := t.calculateReviewTime(task.History)
				if reviewTime > 0 {
					totalReviewTime += reviewTime
					reviewCount++
				}

				break
			}
		}
	}

	if reviewCount > 0 {
		stats.AverageReviewTime = totalReviewTime / time.Duration(reviewCount)
	}

	return stats
}

// calculateReviewTime calculates time spent in review from history
func (t *TaskTransformations) calculateReviewTime(history []HistoryEntry) time.Duration {
	var totalTime time.Duration
	var reviewStarted *time.Time

	for _, entry := range history {
		if entry.ToState == TaskStatusInReview {
			reviewStarted = &entry.Timestamp
		} else if reviewStarted != nil && entry.FromState == TaskStatusInReview {
			totalTime += entry.Timestamp.Sub(*reviewStarted)
			reviewStarted = nil
		}
	}

	// If still in review, add time until now
	if reviewStarted != nil {
		totalTime += time.Since(*reviewStarted)
	}

	return totalTime
}

// ConvertToTaskBatch creates a batch from multiple tasks
func (t *TaskTransformations) ConvertToTaskBatch(
	tasks []TaskRecord,
	operation string,
	reason string,
	initiatedBy string,
) (*TaskBatch, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("cannot create batch with no tasks")
	}

	// Validate operation
	validOps := map[string]bool{
		"cancel": true, "escalate": true, "archive": true,
	}
	if !validOps[operation] {
		return nil, fmt.Errorf("invalid batch operation: %s", operation)
	}

	batch := &TaskBatch{
		Tasks:       tasks,
		Operation:   operation,
		Reason:      reason,
		InitiatedBy: initiatedBy,
		InitiatedAt: time.Now(),
	}

	return batch, nil
}

// Helper function to generate UUID (placeholder)
func generateUUID() string {
	// In production, use a proper UUID library
	return fmt.Sprintf("%d-%d", time.Now().Unix(), time.Now().Nanosecond())
}