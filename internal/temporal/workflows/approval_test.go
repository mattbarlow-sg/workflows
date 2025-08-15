// Package workflows provides tests for approval workflow functionality
package workflows

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// ApprovalWorkflowTestSuite provides test suite for approval workflow
type ApprovalWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up test environment before each test
func (s *ApprovalWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	
	// Register activities
	s.env.RegisterActivity(CreateHumanTaskActivity)
	s.env.RegisterActivity(CleanupHumanTaskActivity)
}

// AfterTest cleans up after each test
func (s *ApprovalWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestApprovalWorkflowCreation tests approval workflow creation
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowCreation() {
	workflow := NewApprovalWorkflow()
	
	assert.NotNil(s.T(), workflow)
	assert.Equal(s.T(), "ApprovalWorkflow", workflow.GetName())
	assert.Equal(s.T(), "1.0.0", workflow.GetVersion())
	assert.Equal(s.T(), "Human approval workflow with escalation support", workflow.GetMetadata().Description)
	assert.Equal(s.T(), "approval", workflow.GetMetadata().Tags["type"])
	assert.Equal(s.T(), "human-task", workflow.GetMetadata().Tags["category"])
}

// TestApprovalInputValidation tests input validation
func (s *ApprovalWorkflowTestSuite) TestApprovalInputValidation() {
	workflow := NewApprovalWorkflow()
	
	// Test valid input
	validInput := ApprovalInput{
		RequestID:   "REQ-001",
		RequestType: "Purchase Order",
		Requester:   "john.doe@company.com",
		Description: "Need approval for office supplies",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
		},
		EscalationChain: []string{"director@company.com"},
		Priority:        PriorityMedium,
		Deadline:        48 * time.Hour,
	}
	
	err := workflow.Validate(validInput)
	assert.NoError(s.T(), err)
	
	// Test invalid inputs
	testCases := []struct {
		name     string
		input    ApprovalInput
		expected string
	}{
		{
			name:     "empty request ID",
			input:    ApprovalInput{},
			expected: "request ID is required",
		},
		{
			name: "empty request type",
			input: ApprovalInput{
				RequestID: "REQ-001",
			},
			expected: "request type is required",
		},
		{
			name: "empty requester",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
			},
			expected: "requester is required",
		},
		{
			name: "no approval steps",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
				Requester:   "john.doe@company.com",
			},
			expected: "at least one approval step is required",
		},
		{
			name: "empty step name",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
				Requester:   "john.doe@company.com",
				ApprovalSteps: []ApprovalStep{
					{
						Name:      "",
						Approvers: []string{"manager@company.com"},
						Timeout:   24 * time.Hour,
					},
				},
			},
			expected: "approval step 0 name is required",
		},
		{
			name: "no approvers",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
				Requester:   "john.doe@company.com",
				ApprovalSteps: []ApprovalStep{
					{
						Name:    "Manager Approval",
						Timeout: 24 * time.Hour,
					},
				},
			},
			expected: "approval step 0 must have at least one approver",
		},
		{
			name: "invalid required count",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
				Requester:   "john.doe@company.com",
				ApprovalSteps: []ApprovalStep{
					{
						Name:      "Manager Approval",
						Approvers: []string{"manager@company.com"},
						Required:  2, // More than available approvers
						Timeout:   24 * time.Hour,
					},
				},
			},
			expected: "approval step 0 required count is invalid",
		},
		{
			name: "zero timeout",
			input: ApprovalInput{
				RequestID:   "REQ-001",
				RequestType: "Purchase Order",
				Requester:   "john.doe@company.com",
				ApprovalSteps: []ApprovalStep{
					{
						Name:      "Manager Approval",
						Approvers: []string{"manager@company.com"},
						Timeout:   0,
					},
				},
			},
			expected: "approval step 0 timeout must be positive",
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := workflow.Validate(tc.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expected)
		})
	}
}

// TestApprovalWorkflowSuccess tests successful approval workflow execution
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowSuccess() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-001",
		RequestType: "Purchase Order",
		Requester:   "john.doe@company.com",
		Description: "Need approval for office supplies",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
		},
		EscalationChain: []string{"director@company.com"},
		Priority:        PriorityMedium,
		Deadline:        48 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send approval signal during execution
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName:  "Manager Approval",
			Approver:  "manager@company.com",
			Decision:  ApprovalDecisionApproved,
			Comments:  "Looks good to me",
		})
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), input.RequestID, output.RequestID)
	assert.Equal(s.T(), ApprovalStatusApproved, output.Status)
	assert.Len(s.T(), output.Approvals, 1)
	assert.Equal(s.T(), ApprovalDecisionApproved, output.Approvals[0].Decision)
	assert.Equal(s.T(), "manager@company.com", output.Approvals[0].Approver)
	assert.Equal(s.T(), "manager@company.com", output.FinalApprover)
}

// TestApprovalWorkflowRejection tests approval workflow rejection
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowRejection() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-002",
		RequestType: "Purchase Order",
		Requester:   "john.doe@company.com",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
		},
		Priority: PriorityMedium,
		Deadline: 48 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send rejection signal during execution
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName:  "Manager Approval",
			Approver:  "manager@company.com",
			Decision:  ApprovalDecisionRejected,
			Comments:  "Budget exceeded",
		})
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), input.RequestID, output.RequestID)
	assert.Equal(s.T(), ApprovalStatusRejected, output.Status)
	assert.Len(s.T(), output.Approvals, 1)
	assert.Equal(s.T(), ApprovalDecisionRejected, output.Approvals[0].Decision)
	assert.Equal(s.T(), "Budget exceeded", output.Approvals[0].Comments)
}

// TestApprovalWorkflowCancellation tests approval workflow cancellation
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowCancellation() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-003",
		RequestType: "Purchase Order",
		Requester:   "john.doe@company.com",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
		},
		Priority: PriorityMedium,
		Deadline: 48 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send cancellation signal during execution
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("cancel", "Request no longer needed")
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), input.RequestID, output.RequestID)
	assert.Equal(s.T(), ApprovalStatusCancelled, output.Status)
}

// TestApprovalWorkflowMultipleSteps tests multiple approval steps
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowMultipleSteps() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-004",
		RequestType: "Large Purchase",
		Requester:   "john.doe@company.com",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
			{
				Name:      "Director Approval",
				Approvers: []string{"director@company.com"},
				Required:  1,
				Timeout:   48 * time.Hour,
			},
		},
		Priority: PriorityHigh,
		Deadline: 72 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send approval signals for both steps
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Manager Approval",
			Approver: "manager@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Manager approved",
		})
	}, 100*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Director Approval",
			Approver: "director@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Director approved",
		})
	}, 200*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), ApprovalStatusApproved, output.Status)
	assert.Len(s.T(), output.Approvals, 2)
	assert.Equal(s.T(), "director@company.com", output.FinalApprover)
}

// TestApprovalWorkflowMultipleApprovers tests multiple approvers for one step
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowMultipleApprovers() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-005",
		RequestType: "Budget Request",
		Requester:   "john.doe@company.com",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Committee Approval",
				Approvers: []string{"member1@company.com", "member2@company.com", "member3@company.com"},
				Required:  2, // Need 2 out of 3 approvals
				Timeout:   48 * time.Hour,
			},
		},
		Priority: PriorityHigh,
		Deadline: 72 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send approval signals from 2 members
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Committee Approval",
			Approver: "member1@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Member 1 approved",
		})
	}, 100*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Committee Approval",
			Approver: "member2@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Member 2 approved",
		})
	}, 150*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), ApprovalStatusApproved, output.Status)
	assert.Len(s.T(), output.Approvals, 2) // Should have 2 approvals
	
	// Verify both approvers
	approvers := make(map[string]bool)
	for _, approval := range output.Approvals {
		approvers[approval.Approver] = true
	}
	assert.True(s.T(), approvers["member1@company.com"])
	assert.True(s.T(), approvers["member2@company.com"])
}

// TestApprovalWorkflowQueries tests query handlers
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowQueries() {
	workflow := NewApprovalWorkflow()
	s.env.RegisterWorkflow(workflow.Execute)
	
	input := ApprovalInput{
		RequestID:   "REQ-006",
		RequestType: "Purchase Order",
		Requester:   "john.doe@company.com",
		ApprovalSteps: []ApprovalStep{
			{
				Name:      "Manager Approval",
				Approvers: []string{"manager@company.com"},
				Required:  1,
				Timeout:   24 * time.Hour,
			},
		},
		Priority: PriorityMedium,
		Deadline: 48 * time.Hour,
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Test queries during execution
	s.env.RegisterDelayedCallback(func() {
		// Test getState query
		result, err := s.env.QueryWorkflow("getState")
		s.NoError(err)
		
		var state WorkflowState
		err = result.Get(&state)
		s.NoError(err)
		s.Equal(WorkflowStatusRunning, state.Status)
		
		// Test getApprovals query
		approvalsResult, err := s.env.QueryWorkflow("getApprovals")
		s.NoError(err)
		
		var approvals []ApprovalRecord
		err = approvalsResult.Get(&approvals)
		s.NoError(err)
		// Should be empty initially
		
		// Send approval
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Manager Approval",
			Approver: "manager@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Approved by manager",
		})
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(workflow.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestApprovalStatus tests approval status enum
func (s *ApprovalWorkflowTestSuite) TestApprovalStatus() {
	statuses := []ApprovalStatus{
		ApprovalStatusPending,
		ApprovalStatusApproved,
		ApprovalStatusRejected,
		ApprovalStatusEscalated,
		ApprovalStatusExpired,
		ApprovalStatusCancelled,
	}
	
	expectedValues := []string{
		"pending",
		"approved",
		"rejected",
		"escalated",
		"expired",
		"cancelled",
	}
	
	for i, status := range statuses {
		assert.Equal(s.T(), expectedValues[i], string(status))
	}
}

// TestApprovalDecision tests approval decision enum
func (s *ApprovalWorkflowTestSuite) TestApprovalDecision() {
	decisions := []ApprovalDecision{
		ApprovalDecisionApproved,
		ApprovalDecisionRejected,
		ApprovalDecisionSkipped,
	}
	
	expectedValues := []string{
		"approved",
		"rejected",
		"skipped",
	}
	
	for i, decision := range decisions {
		assert.Equal(s.T(), expectedValues[i], string(decision))
	}
}

// TestApprovalSignal tests approval signal structure
func (s *ApprovalWorkflowTestSuite) TestApprovalSignal() {
	signal := ApprovalSignal{
		StepName: "Test Step",
		Approver: "test@company.com",
		Decision: ApprovalDecisionApproved,
		Comments: "Test approval",
		Metadata: map[string]interface{}{
			"source": "mobile_app",
			"ip":     "192.168.1.1",
		},
	}
	
	assert.Equal(s.T(), "Test Step", signal.StepName)
	assert.Equal(s.T(), "test@company.com", signal.Approver)
	assert.Equal(s.T(), ApprovalDecisionApproved, signal.Decision)
	assert.Equal(s.T(), "Test approval", signal.Comments)
	assert.Equal(s.T(), "mobile_app", signal.Metadata["source"])
}

// TestApprovalHumanTaskInput tests approval human task input structure
func (s *ApprovalWorkflowTestSuite) TestApprovalHumanTaskInput() {
	deadline := time.Now().Add(24 * time.Hour)
	task := ApprovalHumanTaskInput{
		TaskID:      "task-001",
		TaskType:    "approval",
		AssignedTo:  "user@company.com",
		Title:       "Approval Required",
		Description: "Please review and approve",
		Priority:    string(PriorityHigh),
		Deadline:    deadline,
		Data: map[string]interface{}{
			"requestID": "REQ-001",
			"amount":    1000.00,
		},
	}
	
	assert.Equal(s.T(), "task-001", task.TaskID)
	assert.Equal(s.T(), "approval", task.TaskType)
	assert.Equal(s.T(), "user@company.com", task.AssignedTo)
	assert.Equal(s.T(), "Approval Required", task.Title)
	assert.Equal(s.T(), "high", task.Priority)
	assert.Equal(s.T(), deadline, task.Deadline)
	assert.Equal(s.T(), "REQ-001", task.Data["requestID"])
	assert.Equal(s.T(), 1000.00, task.Data["amount"])
}

// TestActivityOptions tests activity options for different priorities
func (s *ApprovalWorkflowTestSuite) TestActivityOptions() {
	workflow := NewApprovalWorkflow()
	
	// Test critical priority
	criticalOpts := workflow.getActivityOptions(PriorityCritical)
	assert.Equal(s.T(), 5*time.Minute, criticalOpts.StartToCloseTimeout)
	assert.Equal(s.T(), 5*time.Second, criticalOpts.RetryPolicy.InitialInterval)
	
	// Test high priority
	highOpts := workflow.getActivityOptions(PriorityHigh)
	assert.Equal(s.T(), 10*time.Minute, highOpts.StartToCloseTimeout)
	assert.Equal(s.T(), 10*time.Second, highOpts.RetryPolicy.InitialInterval)
	
	// Test default (medium/low)
	defaultOpts := workflow.getActivityOptions(PriorityMedium)
	humanTaskOpts := HumanTaskActivityOptions()
	assert.Equal(s.T(), humanTaskOpts.StartToCloseTimeout, defaultOpts.StartToCloseTimeout)
}

// TestApprovalWorkflowInvalidInput tests workflow with invalid input type
func (s *ApprovalWorkflowTestSuite) TestApprovalWorkflowInvalidInput() {
	workflow := NewApprovalWorkflow()
	
	// Test with invalid input type
	ctx := context.Background()
	_, err := workflow.Execute(ctx, "invalid input")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid input type")
}

// Run the test suite
func TestApprovalWorkflowSuite(t *testing.T) {
	suite.Run(t, new(ApprovalWorkflowTestSuite))
}