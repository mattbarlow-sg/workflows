// Package workflows provides scheduled workflow implementation
package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ScheduledWorkflow implements a scheduled/cron workflow
type ScheduledWorkflow struct {
	*BaseWorkflowImpl
	executionCount    int
	lastExecutionTime time.Time
	schedule          *CronSchedule
	failedExecutions  int
}

// ScheduledInput defines the input for the scheduled workflow
type ScheduledInput struct {
	Schedule         string                 `json:"schedule"` // Cron expression
	MaxRuns          int                    `json:"max_runs"` // 0 = unlimited
	MaxFailures      int                    `json:"max_failures"` // Max consecutive failures before stopping
	TaskConfig       TaskConfiguration      `json:"task_config"`
	RetryPolicy      *temporal.RetryPolicy  `json:"retry_policy,omitempty"`
	TimeZone         string                 `json:"time_zone,omitempty"` // Default: UTC
	StartTime        *time.Time             `json:"start_time,omitempty"` // When to start (default: now)
	EndTime          *time.Time             `json:"end_time,omitempty"`   // When to stop
	Data             map[string]interface{} `json:"data,omitempty"`
}

// ScheduledOutput defines the output of the scheduled workflow
type ScheduledOutput struct {
	TotalExecutions    int                      `json:"total_executions"`
	SuccessfulRuns     int                      `json:"successful_runs"`
	FailedRuns         int                      `json:"failed_runs"`
	LastExecutionTime  time.Time                `json:"last_execution_time"`
	LastExecutionResult *ExecutionResult        `json:"last_execution_result,omitempty"`
	NextExecutionTime  *time.Time               `json:"next_execution_time,omitempty"`
	Status             ScheduledWorkflowStatus  `json:"status"`
	ExecutionHistory   []ExecutionResult        `json:"execution_history"`
}

// TaskConfiguration defines what task to execute
type TaskConfiguration struct {
	ActivityName string                 `json:"activity_name"`
	InputData    map[string]interface{} `json:"input_data"`
	Timeout      time.Duration          `json:"timeout"`
}

// ExecutionResult represents the result of a single execution
type ExecutionResult struct {
	ExecutionNumber int                    `json:"execution_number"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Success         bool                   `json:"success"`
	Error           string                 `json:"error,omitempty"`
	Output          map[string]interface{} `json:"output,omitempty"`
}

// ScheduledWorkflowStatus represents the status of the scheduled workflow
type ScheduledWorkflowStatus string

const (
	ScheduledStatusRunning   ScheduledWorkflowStatus = "running"
	ScheduledStatusCompleted ScheduledWorkflowStatus = "completed"
	ScheduledStatusFailed    ScheduledWorkflowStatus = "failed"
	ScheduledStatusStopped   ScheduledWorkflowStatus = "stopped"
	ScheduledStatusPaused    ScheduledWorkflowStatus = "paused"
)

// CronSchedule represents a parsed cron schedule
type CronSchedule struct {
	Expression string    `json:"expression"`
	TimeZone   string    `json:"time_zone"`
	NextRun    time.Time `json:"next_run"`
}

// NewScheduledWorkflow creates a new scheduled workflow
func NewScheduledWorkflow() *ScheduledWorkflow {
	metadata := WorkflowMetadata{
		Name:        "ScheduledWorkflow",
		Version:     "1.0.0",
		Description: "Scheduled/cron workflow for recurring tasks",
		Tags: map[string]string{
			"type":     "scheduled",
			"category": "recurring",
		},
		Author:    "Temporal Workflows",
		CreatedAt: time.Now(),
	}

	return &ScheduledWorkflow{
		BaseWorkflowImpl: NewBaseWorkflow(metadata),
		executionCount:   0,
		failedExecutions: 0,
	}
}

// Execute runs the scheduled workflow
func (sw *ScheduledWorkflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Type assertion for input
	scheduledInput, ok := input.(ScheduledInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for ScheduledWorkflow")
	}

	// Validate input
	if err := sw.Validate(scheduledInput); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting scheduled workflow", "schedule", scheduledInput.Schedule, "maxRuns", scheduledInput.MaxRuns)

	// Initialize workflow state
	sw.UpdateState("initialization", 0, WorkflowStatusRunning)
	sw.SetStateData("schedule", scheduledInput.Schedule)
	sw.SetStateData("maxRuns", scheduledInput.MaxRuns)
	sw.SetStateData("maxFailures", scheduledInput.MaxFailures)

	// Parse schedule
	schedule, err := sw.parseSchedule(scheduledInput)
	if err != nil {
		sw.SetError(err)
		return nil, fmt.Errorf("failed to parse schedule: %w", err)
	}
	sw.schedule = schedule

	// Setup base queries and signals
	if err := sw.SetupBaseQueries(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup base queries: %w", err)
	}
	sw.SetupBaseSignals(ctx)

	// Setup scheduled-specific query handlers
	sw.setupScheduledQueries(ctx)
	
	// Setup scheduled-specific signal handlers
	stopSignal := sw.setupScheduledSignals(ctx)

	// Configure activity options
	ao := sw.getActivityOptions(scheduledInput)
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Initialize output
	output := &ScheduledOutput{
		Status:           ScheduledStatusRunning,
		ExecutionHistory: []ExecutionResult{},
	}

	// Determine start time
	startTime := workflow.Now(ctx)
	if scheduledInput.StartTime != nil {
		startTime = *scheduledInput.StartTime
	}

	// Calculate first execution time
	nextExecution := sw.calculateNextExecution(startTime)
	if scheduledInput.StartTime != nil && nextExecution.Before(*scheduledInput.StartTime) {
		nextExecution = *scheduledInput.StartTime
	}

	// Main execution loop
	for {
		// Check if we should stop
		if sw.shouldStop(scheduledInput, output) {
			break
		}

		// Wait until next execution time
		currentTime := workflow.Now(ctx)
		if nextExecution.After(currentTime) {
			sleepDuration := nextExecution.Sub(currentTime)
			logger.Info("Sleeping until next execution", "duration", sleepDuration, "nextExecution", nextExecution)
			
			// Wait with ability to be interrupted by stop signal
			selector := workflow.NewSelector(ctx)
			timer := workflow.NewTimer(ctx, sleepDuration)
			
			selector.AddFuture(timer, func(f workflow.Future) {
				// Timer completed, continue with execution
			})
			
			selector.AddReceive(stopSignal, func(c workflow.ReceiveChannel, more bool) {
				var reason string
				c.Receive(ctx, &reason)
				logger.Info("Received stop signal", "reason", reason)
				output.Status = ScheduledStatusStopped
			})
			
			selector.Select(ctx)
			
			// Check if we received stop signal
			if output.Status == ScheduledStatusStopped {
				break
			}
		}

		// Execute the scheduled task
		sw.executionCount++
		progress := 0
		if scheduledInput.MaxRuns > 0 {
			progress = (sw.executionCount * 100) / scheduledInput.MaxRuns
		}
		
		sw.UpdateState(fmt.Sprintf("execution_%d", sw.executionCount), progress, WorkflowStatusRunning)
		
		executionResult := sw.executeScheduledTask(ctx, scheduledInput, sw.executionCount)
		output.ExecutionHistory = append(output.ExecutionHistory, executionResult)
		output.LastExecutionResult = &executionResult
		output.LastExecutionTime = executionResult.EndTime
		output.TotalExecutions = sw.executionCount

		if executionResult.Success {
			output.SuccessfulRuns++
			sw.failedExecutions = 0 // Reset failure count on success
			logger.Info("Scheduled task executed successfully", "execution", sw.executionCount)
		} else {
			output.FailedRuns++
			sw.failedExecutions++
			logger.Error("Scheduled task failed", "execution", sw.executionCount, "error", executionResult.Error)
			
			// Check if we've exceeded max failures
			if scheduledInput.MaxFailures > 0 && sw.failedExecutions >= scheduledInput.MaxFailures {
				logger.Error("Max consecutive failures reached, stopping workflow", "failures", sw.failedExecutions)
				output.Status = ScheduledStatusFailed
				break
			}
		}

		// Check end time constraint
		if scheduledInput.EndTime != nil && workflow.Now(ctx).After(*scheduledInput.EndTime) {
			logger.Info("End time reached, stopping workflow")
			break
		}

		// Calculate next execution time
		nextExecution = sw.calculateNextExecution(workflow.Now(ctx))
		output.NextExecutionTime = &nextExecution
	}

	// Finalize output
	if output.Status == ScheduledStatusRunning {
		output.Status = ScheduledStatusCompleted
	}
	output.NextExecutionTime = nil // No next execution since we're done

	sw.UpdateState("completed", 100, WorkflowStatusCompleted)
	logger.Info("Scheduled workflow completed", "totalExecutions", output.TotalExecutions, "status", output.Status)

	return *output, nil
}

// Validate validates the scheduled input
func (sw *ScheduledWorkflow) Validate(input interface{}) error {
	scheduledInput, ok := input.(ScheduledInput)
	if !ok {
		return fmt.Errorf("input must be of type ScheduledInput")
	}

	if scheduledInput.Schedule == "" {
		return fmt.Errorf("schedule expression is required")
	}

	if scheduledInput.TaskConfig.ActivityName == "" {
		return fmt.Errorf("activity name is required")
	}

	if scheduledInput.TaskConfig.Timeout <= 0 {
		return fmt.Errorf("task timeout must be positive")
	}

	if scheduledInput.MaxRuns < 0 {
		return fmt.Errorf("max runs cannot be negative")
	}

	if scheduledInput.MaxFailures < 0 {
		return fmt.Errorf("max failures cannot be negative")
	}

	// Validate time zone if provided
	if scheduledInput.TimeZone != "" {
		if _, err := time.LoadLocation(scheduledInput.TimeZone); err != nil {
			return fmt.Errorf("invalid time zone: %w", err)
		}
	}

	// Validate that start time is before end time
	if scheduledInput.StartTime != nil && scheduledInput.EndTime != nil {
		if scheduledInput.StartTime.After(*scheduledInput.EndTime) {
			return fmt.Errorf("start time must be before end time")
		}
	}

	return nil
}

// parseSchedule parses the cron schedule
func (sw *ScheduledWorkflow) parseSchedule(input ScheduledInput) (*CronSchedule, error) {
	timeZone := input.TimeZone
	if timeZone == "" {
		timeZone = "UTC"
	}

	// Basic cron validation (in a real implementation, use a proper cron library)
	if err := sw.validateCronExpression(input.Schedule); err != nil {
		return nil, err
	}

	return &CronSchedule{
		Expression: input.Schedule,
		TimeZone:   timeZone,
	}, nil
}

// validateCronExpression validates a cron expression (simplified)
func (sw *ScheduledWorkflow) validateCronExpression(expr string) error {
	// This is a simplified validation - in practice, use a proper cron library
	if expr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}
	
	// Basic format check for standard 5-field cron expressions
	// Real implementation would use github.com/robfig/cron or similar
	return nil
}

// calculateNextExecution calculates the next execution time
func (sw *ScheduledWorkflow) calculateNextExecution(from time.Time) time.Time {
	// This is a simplified implementation
	// In practice, you'd use a proper cron library like github.com/robfig/cron
	
	// For demonstration, assume a simple interval-based schedule
	// Parse common patterns like "@every 1h", "@hourly", "@daily", etc.
	
	switch sw.schedule.Expression {
	case "@hourly", "0 * * * *":
		return from.Truncate(time.Hour).Add(time.Hour)
	case "@daily", "0 0 * * *":
		return from.Truncate(24 * time.Hour).Add(24 * time.Hour)
	case "@weekly":
		return from.AddDate(0, 0, 7)
	case "@monthly":
		return from.AddDate(0, 1, 0)
	default:
		// Default to hourly for unknown expressions
		return from.Add(time.Hour)
	}
}

// shouldStop determines if the workflow should stop
func (sw *ScheduledWorkflow) shouldStop(input ScheduledInput, output *ScheduledOutput) bool {
	// Check if we've reached max runs
	if input.MaxRuns > 0 && sw.executionCount >= input.MaxRuns {
		return true
	}

	// Check if we've been stopped
	if output.Status == ScheduledStatusStopped || output.Status == ScheduledStatusFailed {
		return true
	}

	// Check if we're paused (should pause, not stop)
	if sw.GetState().Status == WorkflowStatusPaused {
		return false // Don't stop, just pause
	}

	return false
}

// executeScheduledTask executes the scheduled task
func (sw *ScheduledWorkflow) executeScheduledTask(ctx workflow.Context, input ScheduledInput, executionNumber int) ExecutionResult {
	logger := workflow.GetLogger(ctx)
	startTime := workflow.Now(ctx)

	result := ExecutionResult{
		ExecutionNumber: executionNumber,
		StartTime:       startTime,
		Success:         false,
	}

	// Prepare activity input
	activityInput := map[string]interface{}{
		"executionNumber": executionNumber,
		"scheduledInput":  input.Data,
		"taskConfig":      input.TaskConfig,
	}

	// Merge task-specific input data
	for k, v := range input.TaskConfig.InputData {
		activityInput[k] = v
	}

	logger.Info("Executing scheduled task", "activityName", input.TaskConfig.ActivityName, "execution", executionNumber)

	// Execute the activity
	var output map[string]interface{}
	err := workflow.ExecuteActivity(ctx, input.TaskConfig.ActivityName, activityInput).Get(ctx, &output)
	
	endTime := workflow.Now(ctx)
	result.EndTime = endTime
	result.Duration = endTime.Sub(startTime)

	if err != nil {
		result.Error = err.Error()
		logger.Error("Scheduled task failed", "error", err, "execution", executionNumber)
	} else {
		result.Success = true
		result.Output = output
		logger.Info("Scheduled task completed successfully", "execution", executionNumber, "duration", result.Duration)
	}

	return result
}

// setupScheduledQueries sets up scheduled-specific query handlers
func (sw *ScheduledWorkflow) setupScheduledQueries(ctx workflow.Context) {
	workflow.SetQueryHandler(ctx, "getExecutionCount", func() (int, error) {
		return sw.executionCount, nil
	})

	workflow.SetQueryHandler(ctx, "getLastExecutionTime", func() (time.Time, error) {
		return sw.lastExecutionTime, nil
	})

	workflow.SetQueryHandler(ctx, "getNextExecutionTime", func() (*time.Time, error) {
		if sw.schedule != nil {
			next := sw.calculateNextExecution(workflow.Now(ctx))
			return &next, nil
		}
		return nil, nil
	})

	workflow.SetQueryHandler(ctx, "getFailedExecutions", func() (int, error) {
		return sw.failedExecutions, nil
	})
}

// setupScheduledSignals sets up scheduled-specific signal handlers
func (sw *ScheduledWorkflow) setupScheduledSignals(ctx workflow.Context) workflow.ReceiveChannel {
	// Stop signal
	stopSignal := workflow.GetSignalChannel(ctx, "stop")

	// Skip next execution signal
	skipSignal := workflow.GetSignalChannel(ctx, "skipNext")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var reason string
			if more := skipSignal.Receive(ctx, &reason); !more {
				return
			}
			workflow.GetLogger(ctx).Info("Skipping next execution", "reason", reason)
			// Implementation would set a flag to skip the next execution
		}
	})

	// Update schedule signal
	updateScheduleSignal := workflow.GetSignalChannel(ctx, "updateSchedule")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var newSchedule string
			if more := updateScheduleSignal.Receive(ctx, &newSchedule); !more {
				return
			}
			workflow.GetLogger(ctx).Info("Updating schedule", "newSchedule", newSchedule)
			// Implementation would update the schedule
			sw.schedule.Expression = newSchedule
		}
	})

	return stopSignal
}

// getActivityOptions returns activity options for scheduled tasks
func (sw *ScheduledWorkflow) getActivityOptions(input ScheduledInput) workflow.ActivityOptions {
	opts := DefaultActivityOptions()
	
	// Use custom timeout if provided
	if input.TaskConfig.Timeout > 0 {
		opts.StartToCloseTimeout = input.TaskConfig.Timeout
	}

	// Use custom retry policy if provided
	if input.RetryPolicy != nil {
		opts.RetryPolicy = input.RetryPolicy
	}

	return opts
}

// ScheduledTaskActivity is a sample activity that can be scheduled
func ScheduledTaskActivity(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// This is a sample implementation - real activities would be provided by the user
	
	executionNumber, ok := input["executionNumber"].(int)
	if !ok {
		executionNumber = 0
	}

	// Simulate some work
	time.Sleep(1 * time.Second)

	output := map[string]interface{}{
		"execution":   executionNumber,
		"timestamp":   time.Now(),
		"status":      "completed",
		"message":     fmt.Sprintf("Scheduled task execution %d completed successfully", executionNumber),
	}

	return output, nil
}

// DataProcessingActivity is another sample scheduled activity
func DataProcessingActivity(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Sample data processing activity
	
	recordsProcessed := 0
	if data, exists := input["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if records, exists := dataMap["records"]; exists {
				if recordList, ok := records.([]interface{}); ok {
					recordsProcessed = len(recordList)
				}
			}
		}
	}

	output := map[string]interface{}{
		"recordsProcessed": recordsProcessed,
		"timestamp":        time.Now(),
		"status":           "completed",
	}

	return output, nil
}

// CleanupActivity is a sample cleanup activity
func CleanupActivity(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Sample cleanup activity
	
	output := map[string]interface{}{
		"cleanupCompleted": true,
		"timestamp":        time.Now(),
		"itemsCleanedUp":   42, // Sample number
	}

	return output, nil
}