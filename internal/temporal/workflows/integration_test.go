// Package workflows provides integration tests for the complete workflow library
package workflows

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// WorkflowLibraryIntegrationTestSuite provides integration tests for the entire workflow library
type WorkflowLibraryIntegrationTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
	
	// Test infrastructure
	versionManager *VersionManager
	registry       *WorkflowRegistry
}

// SetupTest sets up test environment before each test
func (s *WorkflowLibraryIntegrationTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.versionManager = NewVersionManager()
	s.registry = NewWorkflowRegistry()
	
	// Register all activities
	s.env.RegisterActivity(CreateHumanTaskActivity)
	s.env.RegisterActivity(CreateHumanTaskForWorkflowActivity)
	s.env.RegisterActivity(CleanupHumanTaskActivity)
	s.env.RegisterActivity(ScheduledTaskActivity)
	s.env.RegisterActivity(DataProcessingActivity)
	s.env.RegisterActivity(CleanupActivity)
	s.env.RegisterActivity(ReassignHumanTaskActivity)
	s.env.RegisterActivity(EscalateHumanTaskActivity)
	s.env.RegisterActivity(EnrollStudentActivity)
	s.env.RegisterActivity(CheckPrerequisitesActivity)
	s.env.RegisterActivity(StartModuleActivity)
	s.env.RegisterActivity(ProcessAssessmentActivity)
	s.env.RegisterActivity(GenerateCertificateActivity)
	s.env.RegisterActivity(SendProgressUpdateActivity)
}

// AfterTest cleans up after each test
func (s *WorkflowLibraryIntegrationTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestCompleteWorkflowLibrarySetup tests that all workflows can be created and registered
func (s *WorkflowLibraryIntegrationTestSuite) TestCompleteWorkflowLibrarySetup() {
	// Create all workflow types
	approvalWorkflow := NewApprovalWorkflow()
	scheduledWorkflow := NewScheduledWorkflow()
	humanTaskWorkflow := NewHumanTaskWorkflow()
	educationalWorkflow := NewEducationalWorkflow()
	
	// Verify all workflows implement BaseWorkflow
	var baseWorkflows []BaseWorkflow
	baseWorkflows = append(baseWorkflows, approvalWorkflow)
	baseWorkflows = append(baseWorkflows, scheduledWorkflow)
	baseWorkflows = append(baseWorkflows, humanTaskWorkflow)
	baseWorkflows = append(baseWorkflows, educationalWorkflow)
	
	// Register all workflows
	for _, wf := range baseWorkflows {
		err := s.registry.Register(wf)
		assert.NoError(s.T(), err)
	}
	
	// Verify registration
	workflows := s.registry.ListWorkflows()
	assert.Len(s.T(), workflows, 4)
	assert.Contains(s.T(), workflows, "ApprovalWorkflow")
	assert.Contains(s.T(), workflows, "ScheduledWorkflow")
	assert.Contains(s.T(), workflows, "HumanTaskWorkflow")
	assert.Contains(s.T(), workflows, "EducationalWorkflow")
}

// TestWorkflowVersioningIntegration tests versioning with actual workflows
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowVersioningIntegration() {
	// Create different versions of the same workflow
	v1_0_0 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0, ReleaseDate: time.Now()}
	v1_1_0 := WorkflowVersion{Major: 1, Minor: 1, Patch: 0, ReleaseDate: time.Now()}
	v2_0_0 := WorkflowVersion{Major: 2, Minor: 0, Patch: 0, ReleaseDate: time.Now()}
	
	// Create workflows with versioning
	baseWorkflow1 := NewApprovalWorkflow()
	versionedWorkflow1 := WithVersioning(baseWorkflow1, v1_0_0)
	
	baseWorkflow2 := NewApprovalWorkflow()
	versionedWorkflow2 := WithVersioning(baseWorkflow2, v1_1_0)
	
	baseWorkflow3 := NewApprovalWorkflow()
	versionedWorkflow3 := WithVersioning(baseWorkflow3, v2_0_0)
	
	// Register versioned workflows
	err := s.versionManager.RegisterVersionedWorkflow(versionedWorkflow1)
	assert.NoError(s.T(), err)
	
	err = s.versionManager.RegisterVersionedWorkflow(versionedWorkflow2)
	assert.NoError(s.T(), err)
	
	err = s.versionManager.RegisterVersionedWorkflow(versionedWorkflow3)
	assert.NoError(s.T(), err)
	
	// Test version retrieval
	latest, err := s.versionManager.GetLatestWorkflow("ApprovalWorkflow")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), v2_0_0, latest.GetWorkflowVersion())
	
	// Test compatibility
	compatible, err := s.versionManager.GetCompatibleWorkflow("ApprovalWorkflow", v1_0_0)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), v1_1_0, compatible.GetWorkflowVersion()) // Should get the highest compatible version
}

// TestWorkflowMigrationFlow tests complete migration workflow
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowMigrationFlow() {
	// Create versions
	v1_0_0 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	v1_1_0 := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	
	// Add migration rule
	migrationRule := MigrationRule{
		FromVersion: v1_0_0,
		ToVersion:   v1_1_0,
		Handler: func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
			stateMap := fromState.(map[string]interface{})
			stateMap["migrationApplied"] = true
			stateMap["migratedAt"] = workflow.Now(ctx)
			return stateMap, nil
		},
		Description: "Migration from 1.0.0 to 1.1.0",
		Required:    true,
	}
	
	err := s.versionManager.AddMigrationRule("TestWorkflow", migrationRule)
	assert.NoError(s.T(), err)
	
	// Test migration application
	testMigrationWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		initialState := map[string]interface{}{
			"data": "original",
		}
		
		result, err := s.versionManager.ApplyMigration(ctx, "TestWorkflow", v1_0_0, v1_1_0, initialState)
		return result, err
	}
	
	s.env.RegisterWorkflow(testMigrationWorkflow)
	s.env.ExecuteWorkflow(testMigrationWorkflow, nil)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var result map[string]interface{}
	s.NoError(s.env.GetWorkflowResult(&result))
	
	assert.Equal(s.T(), "original", result["data"])
	assert.Equal(s.T(), true, result["migrationApplied"])
	assert.NotNil(s.T(), result["migratedAt"])
}

// TestApprovalWorkflowCompleteFlow tests end-to-end approval workflow
func (s *WorkflowLibraryIntegrationTestSuite) TestApprovalWorkflowCompleteFlow() {
	approvalWf := NewApprovalWorkflow()
	s.env.RegisterWorkflow(approvalWf.Execute)
	
	input := ApprovalInput{
		RequestID:   "INT-001",
		RequestType: "Budget Approval",
		Requester:   "requester@company.com",
		Description: "Q4 Marketing Budget",
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
		EscalationChain: []string{"vp@company.com", "ceo@company.com"},
		Priority:        PriorityHigh,
		Deadline:        72 * time.Hour,
	}
	
	// Mock all activity calls
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything).Return(nil)
	
	// Send approval signals in sequence
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Manager Approval",
			Approver: "manager@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Budget looks reasonable",
		})
	}, 50*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Director Approval",
			Approver: "director@company.com",
			Decision: ApprovalDecisionApproved,
			Comments: "Approved for Q4",
		})
	}, 100*time.Millisecond)
	
	// Test queries during execution
	s.env.RegisterDelayedCallback(func() {
		// Query current state
		result, err := s.env.QueryWorkflow("getState")
		s.NoError(err)
		
		var state WorkflowState
		err = result.Get(&state)
		s.NoError(err)
		s.Equal(WorkflowStatusRunning, state.Status)
		
		// Query progress
		progressResult, err := s.env.QueryWorkflow("getProgress")
		s.NoError(err)
		
		var progress int
		err = progressResult.Get(&progress)
		s.NoError(err)
		s.GreaterOrEqual(progress, 0)
	}, 75*time.Millisecond)
	
	s.env.ExecuteWorkflow(approvalWf.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ApprovalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), ApprovalStatusApproved, output.Status)
	assert.Len(s.T(), output.Approvals, 2)
	assert.Equal(s.T(), "director@company.com", output.FinalApprover)
	assert.Greater(s.T(), output.TotalDuration, time.Duration(0))
}

// TestScheduledWorkflowCompleteFlow tests end-to-end scheduled workflow
func (s *WorkflowLibraryIntegrationTestSuite) TestScheduledWorkflowCompleteFlow() {
	scheduledWf := NewScheduledWorkflow()
	s.env.RegisterWorkflow(scheduledWf.Execute)
	
	input := ScheduledInput{
		Schedule:    "@hourly",
		MaxRuns:     3,
		MaxFailures: 1,
		TaskConfig: TaskConfiguration{
			ActivityName: "ScheduledTaskActivity",
			InputData:    map[string]interface{}{"message": "test"},
			Timeout:      5 * time.Minute,
		},
		TimeZone: "UTC",
	}
	
	// Mock activity calls - make them succeed
	s.env.OnActivity("ScheduledTaskActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"status":    "completed",
			"timestamp": time.Now(),
			"message":   "Task completed successfully",
		}, nil)
	
	// Speed up time for testing
	s.env.SetStartTime(time.Now())
	
	s.env.ExecuteWorkflow(scheduledWf.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output ScheduledOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), ScheduledStatusCompleted, output.Status)
	assert.Equal(s.T(), 3, output.TotalExecutions)
	assert.Equal(s.T(), 3, output.SuccessfulRuns)
	assert.Equal(s.T(), 0, output.FailedRuns)
	assert.Len(s.T(), output.ExecutionHistory, 3)
}

// TestHumanTaskWorkflowCompleteFlow tests end-to-end human task workflow
func (s *WorkflowLibraryIntegrationTestSuite) TestHumanTaskWorkflowCompleteFlow() {
	humanTaskWf := NewHumanTaskWorkflow()
	s.env.RegisterWorkflow(humanTaskWf.Execute)
	
	input := HumanTaskInput{
		WorkflowID: "HUMAN-001",
		Tasks: []HumanTaskDefinition{
			{
				ID:          "task1",
				Title:       "Review Document",
				Description: "Please review the attached document",
				AssignedTo:  "reviewer@company.com",
				Priority:    PriorityMedium,
				Deadline:    24 * time.Hour,
				EscalationTime: 12 * time.Hour,
			},
			{
				ID:           "task2",
				Title:        "Final Approval",
				Description:  "Provide final approval",
				AssignedTo:   "approver@company.com",
				Priority:     PriorityHigh,
				Deadline:     48 * time.Hour,
				Dependencies: []string{"task1"},
			},
		},
		GlobalSettings: GlobalTaskSettings{
			DefaultTimeout:     24 * time.Hour,
			DefaultEscalation:  12 * time.Hour,
			MaxRetries:         3,
			RequireComments:    true,
		},
		EscalationChain: []string{"manager@company.com"},
	}
	
	// Mock activity calls
	s.env.OnActivity(CreateHumanTaskForWorkflowActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(ReassignHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(EscalateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	
	// Send task completion signals
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("taskComplete", TaskCompletionSignal{
			TaskID:      "task1",
			CompletedBy: "reviewer@company.com",
			Result: TaskResult{
				Status:   TaskResultStatusCompleted,
				Comments: "Document looks good",
			},
			Notes: "Reviewed and approved",
		})
	}, 50*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("taskComplete", TaskCompletionSignal{
			TaskID:      "task2",
			CompletedBy: "approver@company.com",
			Result: TaskResult{
				Status:   TaskResultStatusCompleted,
				Comments: "Final approval granted",
			},
			Notes: "All requirements met",
		})
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(humanTaskWf.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output HumanTaskOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), HumanTaskWorkflowStatusCompleted, output.Status)
	assert.Len(s.T(), output.CompletedTasks, 2)
	assert.Empty(s.T(), output.PendingTasks)
	assert.Greater(s.T(), output.TotalDuration, time.Duration(0))
}

// TestEducationalWorkflowCompleteFlow tests end-to-end educational workflow
func (s *WorkflowLibraryIntegrationTestSuite) TestEducationalWorkflowCompleteFlow() {
	educationalWf := NewEducationalWorkflow()
	s.env.RegisterWorkflow(educationalWf.Execute)
	
	input := EducationalInput{
		CourseID:   "COURSE-001",
		StudentID:  "student@university.edu",
		CourseName: "Introduction to Programming",
		Modules: []Module{
			{
				ID:                "module1",
				Name:              "Basic Concepts",
				Type:              ModuleTypeLesson,
				EstimatedDuration: 2 * time.Hour,
				Required:          true,
				Weight:            0.3,
				Assessment: &Assessment{
					ID:           "quiz1",
					Name:         "Basic Concepts Quiz",
					Type:         AssessmentTypeQuiz,
					TimeLimit:    &[]time.Duration{30 * time.Minute}[0],
					MaxAttempts:  2,
					PassingScore: 70,
					Weight:       0.2,
				},
			},
			{
				ID:                "module2",
				Name:              "Advanced Topics",
				Type:              ModuleTypeLesson,
				EstimatedDuration: 3 * time.Hour,
				Prerequisites:     []string{"module1"},
				Required:          true,
				Weight:            0.5,
				Assessment: &Assessment{
					ID:           "final_exam",
					Name:         "Final Exam",
					Type:         AssessmentTypeExam,
					TimeLimit:    &[]time.Duration{2 * time.Hour}[0],
					MaxAttempts:  1,
					PassingScore: 75,
					Weight:       0.3,
				},
			},
		},
		Settings: CourseSettings{
			MinPassingGrade:   70,
			ModuleTimeout:     24 * time.Hour,
			AssessmentTimeout: 2 * time.Hour,
		},
		Instructor: "professor@university.edu",
	}
	
	// Mock activity calls
	s.env.OnActivity(EnrollStudentActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CheckPrerequisitesActivity, mock.Anything, mock.Anything).Return(
		map[string]interface{}{"met": true}, nil)
	s.env.OnActivity(StartModuleActivity, mock.Anything, mock.Anything).Return(
		map[string]interface{}{"started": true}, nil)
	s.env.OnActivity(ProcessAssessmentActivity, mock.Anything, mock.Anything).Return(
		AssessmentResult{
			Score:      85,
			MaxScore:   100,
			Percentage: 85,
			Passed:     true,
		}, nil)
	s.env.OnActivity(GenerateCertificateActivity, mock.Anything, mock.Anything).Return(
		Certificate{
			ID:       "CERT-001",
			IssuedAt: time.Now(),
		}, nil)
	s.env.OnActivity(SendProgressUpdateActivity, mock.Anything, mock.Anything).Return(nil)
	
	// Send module completion signals
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("moduleComplete", ModuleCompletionSignal{
			ModuleID: "module1",
			Score:    85,
			CompletedItems: []string{"lesson1", "quiz1"},
		})
	}, 50*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("moduleComplete", ModuleCompletionSignal{
			ModuleID: "module2",
			Score:    90,
			CompletedItems: []string{"lesson2", "final_exam"},
		})
	}, 100*time.Millisecond)
	
	s.env.ExecuteWorkflow(educationalWf.Execute, input)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var output EducationalOutput
	s.NoError(s.env.GetWorkflowResult(&output))
	
	assert.Equal(s.T(), EducationalStatusCompleted, output.Status)
	assert.Len(s.T(), output.CompletedModules, 2)
	assert.NotNil(s.T(), output.Grade)
	assert.True(s.T(), output.Grade.Passed)
	assert.NotNil(s.T(), output.Certificate)
	assert.Greater(s.T(), output.TotalDuration, time.Duration(0))
}

// TestWorkflowBuilderIntegration tests the workflow builder with actual workflows
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowBuilderIntegration() {
	// Build a workflow using the builder pattern
	builder := NewWorkflowBuilder("IntegratedWorkflow", "1.0.0")
	
	metadata, activities, signals, queries, options := builder.
		WithDescription("Integration test workflow").
		WithAuthor("Test Suite").
		WithTag("environment", "test").
		WithTag("type", "integration").
		WithActivity("testActivity", func() {}, DefaultActivityOptions(), "Test activity").
		WithSignal("testSignal", func() {}, "Test signal").
		WithQuery("testQuery", func() {}, "Test query").
		Build()
	
	// Create a workflow using the metadata
	baseWorkflow := NewBaseWorkflow(metadata)
	
	// Verify the workflow was built correctly
	assert.Equal(s.T(), "IntegratedWorkflow", baseWorkflow.GetName())
	assert.Equal(s.T(), "1.0.0", baseWorkflow.GetVersion())
	assert.Equal(s.T(), "Integration test workflow", baseWorkflow.GetMetadata().Description)
	assert.Equal(s.T(), "Test Suite", baseWorkflow.GetMetadata().Author)
	assert.Equal(s.T(), "test", baseWorkflow.GetMetadata().Tags["environment"])
	
	assert.Len(s.T(), activities, 1)
	assert.Len(s.T(), signals, 1)
	assert.Len(s.T(), queries, 1)
	assert.NotNil(s.T(), options)
	
	// Register with registry
	err := s.registry.Register(baseWorkflow)
	assert.NoError(s.T(), err)
	
	// Verify retrieval
	retrieved, exists := s.registry.Get("IntegratedWorkflow", "1.0.0")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), baseWorkflow, retrieved)
}

// TestCrossWorkflowIntegration tests workflows working together
func (s *WorkflowLibraryIntegrationTestSuite) TestCrossWorkflowIntegration() {
	// Create a composite workflow that uses multiple workflow types
	compositeWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		logger := workflow.GetLogger(ctx)
		logger.Info("Starting composite workflow")
		
		// First, create some scheduled tasks
		scheduledInput := ScheduledInput{
			Schedule: "@hourly",
			MaxRuns:  1,
			TaskConfig: TaskConfiguration{
				ActivityName: "DataProcessingActivity",
				InputData:    map[string]interface{}{"data": "test"},
				Timeout:      5 * time.Minute,
			},
		}
		
		scheduledWf := NewScheduledWorkflow()
		var scheduledOutput ScheduledOutput
		err := workflow.ExecuteChildWorkflow(ctx, scheduledWf.Execute, scheduledInput).Get(ctx, &scheduledOutput)
		if err != nil {
			return nil, err
		}
		
		// Then, create a human task based on the results
		if scheduledOutput.SuccessfulRuns > 0 {
			humanTaskInput := HumanTaskInput{
				WorkflowID: "COMPOSITE-001",
				Tasks: []HumanTaskDefinition{
					{
						ID:          "review_results",
						Title:       "Review Scheduled Task Results",
						Description: "Please review the results of the scheduled data processing",
						AssignedTo:  "reviewer@company.com",
						Priority:    PriorityMedium,
						Deadline:    24 * time.Hour,
					},
				},
				GlobalSettings: GlobalTaskSettings{
					DefaultTimeout: 24 * time.Hour,
				},
			}
			
			humanTaskWf := NewHumanTaskWorkflow()
			var humanTaskOutput HumanTaskOutput
			err = workflow.ExecuteChildWorkflow(ctx, humanTaskWf.Execute, humanTaskInput).Get(ctx, &humanTaskOutput)
			if err != nil {
				return nil, err
			}
			
			// Finally, if human task completed, start approval workflow
			if humanTaskOutput.Status == HumanTaskWorkflowStatusCompleted {
				approvalInput := ApprovalInput{
					RequestID:   "COMPOSITE-APPROVAL-001",
					RequestType: "Data Processing Approval",
					Requester:   "system@company.com",
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
				
				approvalWf := NewApprovalWorkflow()
				var approvalOutput ApprovalOutput
				err = workflow.ExecuteChildWorkflow(ctx, approvalWf.Execute, approvalInput).Get(ctx, &approvalOutput)
				if err != nil {
					return nil, err
				}
				
				return map[string]interface{}{
					"scheduled": scheduledOutput,
					"humanTask": humanTaskOutput,
					"approval":  approvalOutput,
				}, nil
			}
		}
		
		return map[string]interface{}{
			"scheduled": scheduledOutput,
		}, nil
	}
	
	s.env.RegisterWorkflow(compositeWorkflow)
	
	// Mock all activities
	s.env.OnActivity("DataProcessingActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{"processed": true}, nil)
	s.env.OnActivity(CreateHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(CleanupHumanTaskActivity, mock.Anything, mock.Anything).Return(nil)
	
	// Send completion signals for the human task and approval
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("taskComplete", TaskCompletionSignal{
			TaskID:      "review_results",
			CompletedBy: "reviewer@company.com",
			Result: TaskResult{
				Status: TaskResultStatusCompleted,
			},
		})
	}, 100*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval", ApprovalSignal{
			StepName: "Manager Approval",
			Approver: "manager@company.com",
			Decision: ApprovalDecisionApproved,
		})
	}, 150*time.Millisecond)
	
	s.env.ExecuteWorkflow(compositeWorkflow, nil)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var result map[string]interface{}
	s.NoError(s.env.GetWorkflowResult(&result))
	
	// Verify all workflow types completed
	assert.Contains(s.T(), result, "scheduled")
	assert.Contains(s.T(), result, "humanTask")
	assert.Contains(s.T(), result, "approval")
}

// TestWorkflowLibraryErrorHandling tests error handling across all workflows
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowLibraryErrorHandling() {
	// Test that all workflows handle invalid input gracefully
	workflows := map[string]BaseWorkflow{
		"approval":     NewApprovalWorkflow(),
		"scheduled":    NewScheduledWorkflow(),
		"human_task":   NewHumanTaskWorkflow(),
		"educational":  NewEducationalWorkflow(),
	}
	
	invalidInputs := []interface{}{
		nil,
		"invalid string",
		123,
		[]string{"invalid", "array"},
		map[string]interface{}{"invalid": "map"},
	}
	
	for workflowName, wf := range workflows {
		for i, invalidInput := range invalidInputs {
			s.T().Run(fmt.Sprintf("%s_invalid_input_%d", workflowName, i), func(t *testing.T) {
				err := wf.Validate(invalidInput)
				assert.Error(t, err, "Workflow %s should reject invalid input %v", workflowName, invalidInput)
			})
		}
	}
}

// TestWorkflowLibraryQueryConsistency tests that all workflows support base queries
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowLibraryQueryConsistency() {
	// Create a test workflow that verifies all base queries are available
	testQueryWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		baseWorkflow := NewBaseWorkflow(WorkflowMetadata{
			Name:    "QueryTest",
			Version: "1.0.0",
		})
		
		// Setup base queries
		err := baseWorkflow.SetupBaseQueries(ctx)
		if err != nil {
			return nil, err
		}
		
		// Update state for testing
		baseWorkflow.UpdateState("testing", 50, WorkflowStatusRunning)
		
		// Short delay to allow queries to be processed
		workflow.Sleep(ctx, 100*time.Millisecond)
		
		return "query_setup_complete", nil
	}
	
	s.env.RegisterWorkflow(testQueryWorkflow)
	
	// Test all base queries
	s.env.RegisterDelayedCallback(func() {
		// Test each base query
		baseQueries := []string{"getState", "getMetadata", "getProgress"}
		
		for _, queryName := range baseQueries {
			result, err := s.env.QueryWorkflow(queryName)
			s.NoError(err, "Query %s should not error", queryName)
			s.NotNil(result, "Query %s should return a result", queryName)
		}
	}, 50*time.Millisecond)
	
	s.env.ExecuteWorkflow(testQueryWorkflow, nil)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestWorkflowLibrarySignalConsistency tests that all workflows support base signals
func (s *WorkflowLibraryIntegrationTestSuite) TestWorkflowLibrarySignalConsistency() {
	// Create a test workflow that verifies all base signals work
	testSignalWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		baseWorkflow := NewBaseWorkflow(WorkflowMetadata{
			Name:    "SignalTest",
			Version: "1.0.0",
		})
		
		// Setup base signals
		baseWorkflow.SetupBaseSignals(ctx)
		
		// Wait for signals
		workflow.Sleep(ctx, 200*time.Millisecond)
		
		return baseWorkflow.GetState(), nil
	}
	
	s.env.RegisterWorkflow(testSignalWorkflow)
	
	// Send base signals
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("pause", "test pause")
	}, 50*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("updateData", DataUpdate{
			Key:   "testSignal",
			Value: "signalReceived",
		})
	}, 100*time.Millisecond)
	
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("resume", "test resume")
	}, 150*time.Millisecond)
	
	s.env.ExecuteWorkflow(testSignalWorkflow, nil)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var result WorkflowState
	s.NoError(s.env.GetWorkflowResult(&result))
	
	// Verify signal was processed (status should be running, not paused)
	assert.Equal(s.T(), WorkflowStatusRunning, result.Status)
}

// Run the test suite
func TestWorkflowLibraryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WorkflowLibraryIntegrationTestSuite))
}