// Package testing provides integration tests for the Temporal testing framework
package testing

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sample workflows and activities for testing

// SimpleWorkflow is a simple workflow for testing
func SimpleWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SimpleActivity, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// SimpleActivity is a simple activity for testing
func SimpleActivity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}

// LongRunningWorkflow is a workflow that runs for a long time
func LongRunningWorkflow(ctx workflow.Context, duration time.Duration) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting long-running workflow", "duration", duration)

	// Simulate long-running work with periodic checkpoints
	checkpoints := 10
	interval := duration / time.Duration(checkpoints)

	for i := 0; i < checkpoints; i++ {
		logger.Info("Checkpoint", "number", i+1)
		
		// Sleep for interval
		if err := workflow.Sleep(ctx, interval); err != nil {
			return err
		}

		// Execute checkpoint activity
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		}
		ctx := workflow.WithActivityOptions(ctx, ao)
		
		var result string
		err := workflow.ExecuteActivity(ctx, CheckpointActivity, i+1).Get(ctx, &result)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckpointActivity records a checkpoint
func CheckpointActivity(ctx context.Context, checkpoint int) (string, error) {
	return fmt.Sprintf("Checkpoint %d completed", checkpoint), nil
}

// SignalWorkflow is a workflow that handles signals
func SignalWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)
	
	var signalVal string
	signalChan := workflow.GetSignalChannel(ctx, "config-signal")

	s := workflow.NewSelector(ctx)
	s.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalVal)
		logger.Info("Received signal", "value", signalVal)
	})
	s.Select(ctx)

	if signalVal == "" {
		return "", errors.New("no signal received")
	}

	return fmt.Sprintf("Processed signal: %s", signalVal), nil
}

// QueryWorkflow is a workflow that handles queries
func QueryWorkflow(ctx workflow.Context, initialState string) error {
	state := initialState
	
	err := workflow.SetQueryHandler(ctx, "state", func() (string, error) {
		return state, nil
	})
	if err != nil {
		return err
	}

	// Update state over time
	for i := 0; i < 5; i++ {
		workflow.Sleep(ctx, time.Second)
		state = fmt.Sprintf("state-%d", i+1)
	}

	return nil
}

// HumanTaskWorkflow is a workflow that includes human tasks
func HumanTaskWorkflow(ctx workflow.Context, taskID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	
	// Create human task
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 24 * time.Hour, // Long timeout for human tasks
		HeartbeatTimeout:    time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No retries for human tasks
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var taskResult string
	future := workflow.ExecuteActivity(ctx, CreateHumanTaskActivity, taskID)
	
	// Wait for human task completion signal
	signalChan := workflow.GetSignalChannel(ctx, "HumanTaskCompleted")
	
	selector := workflow.NewSelector(ctx)
	
	var completionData map[string]interface{}
	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &completionData)
		logger.Info("Human task completed", "data", completionData)
	})
	
	selector.AddFuture(future, func(f workflow.Future) {
		err := f.Get(ctx, &taskResult)
		if err != nil {
			logger.Error("Human task activity failed", "error", err)
		}
	})
	
	selector.Select(ctx)
	
	if response, ok := completionData["response"].(string); ok {
		return response, nil
	}
	
	return taskResult, nil
}

// CreateHumanTaskActivity creates a human task
func CreateHumanTaskActivity(ctx context.Context, taskID string) (string, error) {
	// In real implementation, this would create a task in an external system
	return fmt.Sprintf("Task %s created", taskID), nil
}

// RetryWorkflow is a workflow that tests retry behavior
func RetryWorkflow(ctx workflow.Context, maxAttempts int) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    int32(maxAttempts),
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, RetryableActivity, maxAttempts).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// RetryableActivity is an activity that fails initially then succeeds
func RetryableActivity(ctx context.Context, succeedAfter int) (string, error) {
	// This would be mocked in tests to control failure behavior
	return "Success", nil
}

// Integration Tests

func TestSimpleWorkflow(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(SimpleWorkflow)
	env.RegisterActivity(SimpleActivity)

	env.ExecuteWorkflow(SimpleWorkflow, "World")

	env.AssertWorkflowCompleted(5 * time.Second)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestWorkflowWithMockedActivity(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(SimpleWorkflow)
	
	// Mock the activity
	mock := env.MockActivity("SimpleActivity")
	mock.Returns("Mocked Result")

	env.ExecuteWorkflow(SimpleWorkflow, "Test")

	env.AssertWorkflowCompleted(5 * time.Second)
	env.AssertActivityCalled("SimpleActivity", 1)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "Mocked Result" {
		t.Errorf("Expected mocked result, got %s", result)
	}
}

func TestLongRunningWorkflowWithTimeSkipping(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(LongRunningWorkflow)
	
	// Mock checkpoint activity
	mock := env.MockActivity("CheckpointActivity")
	mock.Returns("Checkpoint completed")

	// Create time-skipping helper
	timeSkipper := NewTestTimeSkipping(env)
	
	// Add checkpoints for verification
	for i := 1; i <= 10; i++ {
		checkpoint := i
		timeSkipper.AddCheckpoint(1*time.Hour, func() error {
			// Verify checkpoint was called
			if mock.CallCount < checkpoint {
				return fmt.Errorf("Expected checkpoint %d to be called", checkpoint)
			}
			return nil
		})
	}

	// Execute workflow with 10 hour duration
	env.ExecuteWorkflow(LongRunningWorkflow, 10*time.Hour)

	// Run through all checkpoints
	if err := timeSkipper.RunAll(); err != nil {
		t.Fatal(err)
	}

	env.AssertWorkflowCompleted(1 * time.Second)
	env.AssertActivityCalled("CheckpointActivity", 10)
}

func TestSignalHandling(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(SignalWorkflow)

	// Setup signal to be sent after 1 second
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("config-signal", "test-config")
	}, 1*time.Second)

	env.ExecuteWorkflow(SignalWorkflow)

	// Skip time to allow signal to be processed
	env.SkipTime(2 * time.Second)

	env.AssertWorkflowCompleted(5 * time.Second)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	expected := "Processed signal: test-config"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestQueryHandling(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(QueryWorkflow)

	env.ExecuteWorkflow(QueryWorkflow, "initial")

	// Query initial state
	var state string
	result, err := env.QueryWorkflow("state")
	if err == nil && result != nil {
		state = fmt.Sprintf("%v", result)
	}
	if err != nil {
		t.Fatal(err)
	}

	if state != "initial" {
		t.Errorf("Expected initial state, got %v", state)
	}

	// Skip time and query again
	env.SkipTime(3 * time.Second)

	result, err = env.QueryWorkflow("state")
	if err == nil && result != nil {
		state = fmt.Sprintf("%v", result)
	}
	if err != nil {
		t.Fatal(err)
	}

	// State should have been updated
	if state == "initial" {
		t.Error("State should have been updated")
	}
}

func TestHumanTaskSimulation(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(HumanTaskWorkflow)
	env.RegisterActivity(CreateHumanTaskActivity)

	// Create human task simulation
	task := env.CreateHumanTask("task-123", "approval", "user@example.com")
	task.ResponseDelay = 5 * time.Second
	task.AutoComplete = true
	task.Response = "approved"

	env.ExecuteWorkflow(HumanTaskWorkflow, "task-123")

	// Skip time to allow human task to complete
	env.SkipTime(6 * time.Second)

	env.AssertWorkflowCompleted(10 * time.Second)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "approved" {
		t.Errorf("Expected approved, got %s", result)
	}
}

func TestHumanTaskWithEscalation(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(HumanTaskWorkflow)
	env.RegisterActivity(CreateHumanTaskActivity)

	// Create human task with escalation
	task := env.CreateHumanTask("task-456", "approval", "user@example.com")
	task.AutoComplete = false
	task.EscalationPolicy = &EscalationPolicy{
		EscalateAfter:   2 * time.Second,
		EscalateTo:      "manager@example.com",
		MaxEscalations:  2,
		EscalationChain: []string{"user@example.com", "manager@example.com", "director@example.com"},
	}

	env.ExecuteWorkflow(HumanTaskWorkflow, "task-456")

	// Skip time to trigger escalation
	env.SkipTime(3 * time.Second)

	// Escalate the task
	err := env.EscalateHumanTask("task-456")
	if err != nil {
		t.Fatal(err)
	}

	// Complete the task after escalation
	env.CompleteHumanTask("task-456", "approved-by-manager", nil)

	env.AssertWorkflowCompleted(10 * time.Second)

	var result string
	err = env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "approved-by-manager" {
		t.Errorf("Expected approved-by-manager, got %s", result)
	}
}

func TestRetryBehavior(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(RetryWorkflow)

	// Setup retry helper
	retryHelper := NewRetryTestHelper(env, "RetryableActivity")
	retryHelper.SetSuccessAfter(3)
	retryHelper.Setup()

	env.ExecuteWorkflow(RetryWorkflow, 5)

	env.AssertWorkflowCompleted(30 * time.Second)
	retryHelper.AssertRetriedExpectedTimes(t)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "success" {
		t.Errorf("Expected success, got %s", result)
	}
}

func TestWorkflowBuilderPattern(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Use builder pattern to set up test
	builder := NewWorkflowBuilder(env)
	
	mockActivity := &MockActivity{
		Name:         "SimpleActivity",
		ReturnValues: []interface{}{"Built with builder"},
	}
	
	builder.
		WithWorkflow(SimpleWorkflow).
		WithMockedActivity("SimpleActivity", mockActivity).
		WithSignalHandler("test-signal", func(signalName string, signalArg interface{}) {
			// Handle signal
		}).
		WithQueryHandler("test-query", func(queryType string) (interface{}, error) {
			return "query-result", nil
		})

	builder.Execute("Builder Test")

	env.AssertWorkflowCompleted(5 * time.Second)

	var result string
	err := env.GetWorkflowResult(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "Built with builder" {
		t.Errorf("Expected 'Built with builder', got %s", result)
	}
}

func TestDataDrivenWorkflow(t *testing.T) {
	tests := []DataDrivenTest{
		{
			Name: "SimpleWorkflow",
			TestData: []TestData{
				{
					Name:           "WithWorld",
					Input:          "World",
					ExpectedOutput: "Hello, World!",
				},
				{
					Name:           "WithTemporal",
					Input:          "Temporal",
					ExpectedOutput: "Hello, Temporal!",
				},
				{
					Name:           "WithEmpty",
					Input:          "",
					ExpectedOutput: "Hello, !",
				},
			},
			TestFunc: func(data TestData) error {
				env := NewTestEnvironment(t)
				defer env.Cleanup()

				env.RegisterWorkflow(SimpleWorkflow)
				env.RegisterActivity(SimpleActivity)

				env.ExecuteWorkflow(SimpleWorkflow, data.Input)
				env.AssertWorkflowCompleted(5 * time.Second)

				var result string
				if err := env.GetWorkflowResult(&result); err != nil {
					return err
				}

				if result != data.ExpectedOutput {
					return fmt.Errorf("expected %s, got %s", data.ExpectedOutput, result)
				}

				return nil
			},
		},
	}

	RunDataDrivenTests(t, tests)
}

func TestWorkflowScenario(t *testing.T) {
	helper := NewWorkflowTestHelper(t)
	
	scenario := WorkflowScenario{
		Name:     "CompleteScenario",
		Workflow: SimpleWorkflow,
		Input:    []interface{}{"Scenario"},
		Setup: func(env *TestEnvironment) {
			env.RegisterWorkflow(SimpleWorkflow)
			env.RegisterActivity(SimpleActivity)
		},
		DuringExecution: func(env *TestEnvironment) {
			// Could send signals or queries here
		},
		VerifyResults: func(env *TestEnvironment) {
			var result string
			err := env.GetWorkflowResult(&result)
			if err != nil {
				t.Fatal(err)
			}
			
			if result != "Hello, Scenario!" {
				t.Errorf("Unexpected result: %s", result)
			}
		},
		Timeout: 10 * time.Second,
	}
	
	helper.TestWorkflowScenario(scenario)
}

func TestActivityHelper(t *testing.T) {
	helper := NewActivityTestHelper(t)
	
	helper.
		AddTestCase(&ActivityTestCase{
			Name:           "SimpleActivity",
			Function:       SimpleActivity,
			Input:          []interface{}{"Test"},
			ExpectedOutput: "Hello, Test!",
		}).
		AddTestCase(&ActivityTestCase{
			Name:           "CheckpointActivity",
			Function:       CheckpointActivity,
			Input:          []interface{}{5},
			ExpectedOutput: "Checkpoint 5 completed",
		})
	
	helper.TestAll()
}

func TestWorkflowSnapshot(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(QueryWorkflow)
	env.ExecuteWorkflow(QueryWorkflow, "initial")

	// Capture initial snapshot
	snapshot1, err := CaptureSnapshot(env)
	if err != nil {
		t.Fatal(err)
	}

	// Skip time to change state
	env.SkipTime(3 * time.Second)

	// Capture second snapshot
	snapshot2, err := CaptureSnapshot(env)
	if err != nil {
		t.Fatal(err)
	}

	// Compare snapshots
	differences := CompareSnapshots(snapshot1, snapshot2)
	if len(differences) == 0 {
		t.Error("Expected differences between snapshots")
	}

	// Should detect state change
	foundStateChange := false
	for _, diff := range differences {
		if contains(diff, "State changed") {
			foundStateChange = true
			break
		}
	}

	if !foundStateChange {
		t.Error("Expected to detect state change")
	}
}

func TestMockActivityFactoryIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	factory := NewMockActivityFactory()
	
	// Create different types of mocks
	dbMock := factory.CreateDatabaseActivity("SaveData")
	httpMock := factory.CreateHTTPActivity("CallAPI", 200, map[string]string{"status": "ok"})
	asyncMock := factory.CreateAsyncActivity("ProcessAsync", 2*time.Second)
	failingMock := factory.CreateFailingActivity("FailingActivity", errors.New("database error"), 2)
	retryableMock := factory.CreateRetryableActivity("RetryableActivity", 3)
	
	// Register mocks with environment
	env.mockActivities["SaveData"] = dbMock
	env.mockActivities["CallAPI"] = httpMock
	env.mockActivities["ProcessAsync"] = asyncMock
	env.mockActivities["FailingActivity"] = failingMock
	env.mockActivities["RetryableActivity"] = retryableMock
	
	// Test that mocks work as expected
	ctx := context.Background()
	
	// Test database mock
	result, err := dbMock.Execute(ctx, "test-data")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("Expected result from database mock")
	}
	
	// Test HTTP mock
	result, err = httpMock.Execute(ctx, "http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	httpResult := result.(map[string]interface{})
	if httpResult["statusCode"] != 200 {
		t.Error("Expected status code 200")
	}
	
	// Test failing mock
	for i := 0; i < 3; i++ {
		_, err = failingMock.Execute(ctx)
		if i <= 2 && err == nil {
			t.Error("Expected error from failing mock")
		}
		if i > 2 && err != nil {
			t.Error("Expected success after failures")
		}
	}
	
	// Reset all mocks
	factory.ResetAll()
	
	// Verify reset worked
	if dbMock.CallCount != 0 {
		t.Error("Expected call count to be reset")
	}
}

func TestLongRunningWorkflowTester(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	env.RegisterWorkflow(LongRunningWorkflow)
	env.MockActivity("CheckpointActivity").Returns("Checkpoint completed")

	tester := NewLongRunningWorkflowTester(env, 24*time.Hour)
	tester.SetCheckInterval(2 * time.Hour)
	
	checkpointCount := 0
	tester.AddPeriodicCheck(func() error {
		checkpointCount++
		// Verify workflow is still running
		if env.IsWorkflowCompleted() && checkpointCount < 10 {
			return fmt.Errorf("workflow completed too early")
		}
		return nil
	})

	env.ExecuteWorkflow(LongRunningWorkflow, 10*time.Hour)

	err := tester.RunUntilComplete()
	if err != nil {
		t.Fatal(err)
	}

	elapsed := tester.GetElapsedTime()
	if elapsed < 10*time.Hour {
		t.Errorf("Expected at least 10 hours elapsed, got %v", elapsed)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && s[:len(substr)] == substr || 
		len(s) > len(substr) && contains(s[1:], substr)
}