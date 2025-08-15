# Temporal Testing Framework

A comprehensive testing framework for Temporal workflows that provides mocking, time-skipping, and test helpers for unit, integration, and end-to-end testing.

## Features

- **Test Environment**: Wraps Temporal's test suite with additional helpers
- **Activity Mocking**: Mock activities with configurable behavior
- **Time Skipping**: Test long-running workflows without waiting
- **Signal & Query Testing**: Helpers for testing workflow signals and queries
- **Human Task Simulation**: Simulate human interactions and escalations
- **Retry Testing**: Test retry behavior and failure scenarios
- **Test Builders**: Fluent API for building test scenarios

## Quick Start

### Basic Workflow Test

```go
func TestSimpleWorkflow(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(MyWorkflow)
    env.RegisterActivity(MyActivity)

    env.ExecuteWorkflow(MyWorkflow, "input")
    
    env.AssertWorkflowCompleted(5 * time.Second)
    
    var result string
    err := env.GetWorkflowResult(&result)
    if err != nil {
        t.Fatal(err)
    }
}
```

### Mocking Activities

```go
func TestWithMockedActivity(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(MyWorkflow)
    
    // Mock the activity
    mock := env.MockActivity("ProcessData")
    mock.Returns("mocked result")
    
    env.ExecuteWorkflow(MyWorkflow, "input")
    env.AssertActivityCalled("ProcessData", 1)
}
```

### Time Skipping for Long-Running Workflows

```go
func TestLongRunningWorkflow(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(LongRunningWorkflow)
    
    timeSkipper := NewTestTimeSkipping(env)
    
    // Add checkpoints every hour
    for i := 0; i < 24; i++ {
        timeSkipper.AddCheckpoint(1*time.Hour, func() error {
            // Verify workflow state
            return nil
        })
    }
    
    env.ExecuteWorkflow(LongRunningWorkflow, 24*time.Hour)
    
    // Run through all checkpoints
    if err := timeSkipper.RunAll(); err != nil {
        t.Fatal(err)
    }
}
```

## Component Documentation

### Test Environment

The `TestEnvironment` is the core component that wraps Temporal's test suite:

```go
env := NewTestEnvironment(t)
```

Key methods:
- `RegisterWorkflow(wf)` - Register a workflow for testing
- `RegisterActivity(act)` - Register an activity for testing
- `ExecuteWorkflow(wf, args...)` - Execute a workflow
- `AssertWorkflowCompleted(timeout)` - Assert workflow completed
- `AssertWorkflowFailed(timeout)` - Assert workflow failed
- `SkipTime(duration)` - Skip time in the test environment

### Mock Activities

Create sophisticated activity mocks:

```go
mock := &MockActivity{
    Name: "MyActivity",
    ReturnValues: []interface{}{"result1", "result2"},
}

// Add validation
mock.WithValidation(func(args []interface{}) error {
    if len(args) == 0 {
        return errors.New("no arguments")
    }
    return nil
})

// Add delays
mock.WithDelay(func(callIndex int) time.Duration {
    return time.Duration(callIndex) * time.Second
})

// Add side effects
mock.WithSideEffect(func(args []interface{}) error {
    // Perform side effect
    return nil
})
```

### Mock Factory

Use the factory to create common mock patterns:

```go
factory := NewMockActivityFactory()

// Database activity
dbMock := factory.CreateDatabaseActivity("SaveData")

// HTTP activity
httpMock := factory.CreateHTTPActivity("CallAPI", 200, responseBody)

// Async activity with delay
asyncMock := factory.CreateAsyncActivity("ProcessAsync", 2*time.Second)

// Activity that fails then succeeds
retryMock := factory.CreateRetryableActivity("RetryActivity", 3)

// Activity that always fails
failMock := factory.CreateFailingActivity("FailActivity", errors.New("error"), 0)
```

### Signal Testing

Test signal handling in workflows:

```go
helper := NewSignalTestHelper(env)

helper.
    AddSignal("signal1", data1, 1*time.Second).
    AddSignal("signal2", data2, 2*time.Second).
    AddSignalWithVerification("signal3", data3, 3*time.Second, func() error {
        // Verify workflow state after signal
        return nil
    })

err := helper.Execute()
```

### Query Testing

Test query handling:

```go
helper := NewQueryTestHelper(env)

helper.
    AddQuery("state", expectedState).
    AddQueryWithVerification("complex", func(result interface{}) error {
        // Custom verification
        return nil
    })

err := helper.Execute()
```

### Human Task Simulation

Simulate human task interactions:

```go
task := env.CreateHumanTask("task-123", "approval", "user@example.com")
task.ResponseDelay = 5 * time.Second
task.AutoComplete = true
task.Response = "approved"

// With escalation
task.EscalationPolicy = &EscalationPolicy{
    EscalateAfter:  2 * time.Second,
    EscalateTo:     "manager@example.com",
    MaxEscalations: 3,
}

// Manually escalate
env.EscalateHumanTask("task-123")

// Complete task
env.CompleteHumanTask("task-123", "approved", nil)
```

### Workflow Builder

Use the builder pattern for complex test setups:

```go
builder := NewWorkflowBuilder(env)

builder.
    WithWorkflow(MyWorkflow).
    WithActivity(Activity1).
    WithMockedActivity("Activity2", mock2).
    WithSignalHandler("signal", signalHandler).
    WithQueryHandler("query", queryHandler).
    Execute("workflow", "input")
```

### Data-Driven Testing

Run data-driven tests:

```go
tests := []DataDrivenTest{
    {
        Name: "TestScenarios",
        TestData: []TestData{
            {Name: "Scenario1", Input: "A", ExpectedOutput: "B"},
            {Name: "Scenario2", Input: "C", ExpectedOutput: "D"},
        },
        TestFunc: func(data TestData) error {
            // Run test with data
            return nil
        },
    },
}

RunDataDrivenTests(t, tests)
```

### Retry Testing

Test retry behavior:

```go
helper := NewRetryTestHelper(env, "RetryActivity")
helper.SetSuccessAfter(3)
helper.SetError(temporaryError)

mock := helper.Setup()

env.ExecuteWorkflow(WorkflowWithRetry)

helper.AssertRetriedExpectedTimes(t)
```

### Long-Running Workflow Testing

Test workflows that run for extended periods:

```go
tester := NewLongRunningWorkflowTester(env, 30*24*time.Hour)
tester.SetCheckInterval(24 * time.Hour)

tester.AddPeriodicCheck(func() error {
    // Verify workflow state
    return nil
})

env.ExecuteWorkflow(MonthLongWorkflow)

err := tester.RunUntilComplete()
elapsed := tester.GetElapsedTime()
```

## Test Patterns

### Pattern 1: Testing Workflow Logic

```go
func TestWorkflowLogic(t *testing.T) {
    helper := NewWorkflowTestHelper(t)
    
    scenario := WorkflowScenario{
        Name:     "CompleteScenario",
        Workflow: MyWorkflow,
        Input:    []interface{}{"input"},
        Setup: func(env *TestEnvironment) {
            // Setup mocks and dependencies
        },
        DuringExecution: func(env *TestEnvironment) {
            // Send signals or queries during execution
        },
        VerifyResults: func(env *TestEnvironment) {
            // Verify final state
        },
        Timeout: 10 * time.Second,
    }
    
    helper.TestWorkflowScenario(scenario)
}
```

### Pattern 2: Testing Error Handling

```go
func TestErrorHandling(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(MyWorkflow)
    
    // Mock activity to return error
    mock := env.MockActivity("ProcessData")
    mock.ReturnsError(errors.New("processing failed"))
    
    env.ExecuteWorkflow(MyWorkflow, "input")
    
    // Verify workflow handles error correctly
    env.AssertWorkflowFailed(5 * time.Second)
    
    err := env.GetWorkflowError()
    if !strings.Contains(err.Error(), "processing failed") {
        t.Error("Expected specific error")
    }
}
```

### Pattern 3: Testing State Transitions

```go
func TestStateTransitions(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(StateMachineWorkflow)
    env.ExecuteWorkflow(StateMachineWorkflow)
    
    assertion := NewWorkflowStateAssertion(env)
    
    // Initial state
    assertion.AssertState("pending")
    
    // Send signal to transition
    env.SignalWorkflow("start", nil)
    env.SkipTime(1 * time.Second)
    
    assertion.AssertState("running")
    
    // Complete
    env.SignalWorkflow("complete", nil)
    env.SkipTime(1 * time.Second)
    
    assertion.AssertState("completed")
    assertion.AssertCompleted()
}
```

### Pattern 4: Testing Compensation

```go
func TestCompensation(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    env.RegisterWorkflow(SagaWorkflow)
    
    // Mock activities
    activity1 := env.MockActivity("Activity1")
    activity1.Returns("success")
    
    activity2 := env.MockActivity("Activity2")
    activity2.ReturnsError(errors.New("failed"))
    
    compensate1 := env.MockActivity("Compensate1")
    compensate1.Returns("compensated")
    
    env.ExecuteWorkflow(SagaWorkflow)
    
    env.AssertWorkflowCompleted(10 * time.Second)
    
    // Verify compensation was called
    env.AssertActivityCalled("Compensate1", 1)
}
```

## Best Practices

1. **Always clean up**: Use `defer env.Cleanup()` to ensure resources are freed
2. **Use time skipping**: Don't wait for real time in tests
3. **Mock external dependencies**: Mock all activities that call external services
4. **Test error paths**: Always test both success and failure scenarios
5. **Use assertions**: Use the provided assertion helpers for clearer test failures
6. **Isolate tests**: Each test should be independent and not rely on others
7. **Use descriptive names**: Name your mocks and test scenarios clearly

## Advanced Features

### Workflow Snapshots

Capture and compare workflow states:

```go
snapshot1, _ := CaptureSnapshot(env)

// Perform operations
env.SignalWorkflow("signal", data)
env.SkipTime(1 * time.Second)

snapshot2, _ := CaptureSnapshot(env)

differences := CompareSnapshots(snapshot1, snapshot2)
// Analyze differences
```

### Custom Mock Behaviors

Create complex mock behaviors:

```go
callCount := 0
mock.WithSideEffect(func(args []interface{}) error {
    callCount++
    if callCount < 3 {
        return errors.New("temporary failure")
    }
    return nil
})
```

### Parallel Test Execution

The framework supports parallel test execution:

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    
    env := NewTestEnvironment(t)
    defer env.Cleanup()
    
    // Test implementation
}
```

## Troubleshooting

### Common Issues

1. **Test times out**: Ensure you're using time skipping for long-running workflows
2. **Mock not called**: Verify the activity name matches exactly
3. **Signal not received**: Check that signal is sent after workflow starts
4. **Query fails**: Ensure query handler is registered in the workflow

### Debug Tips

- Enable test environment logging for detailed execution traces
- Use `t.Logf()` to log important checkpoints
- Capture snapshots at key points to understand state changes
- Use the assertion helpers for clearer error messages

## Examples

See the `integration_test.go` and `example_usage_test.go` files for complete examples of all features.