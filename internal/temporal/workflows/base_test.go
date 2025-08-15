// Package workflows provides tests for base workflow functionality
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

// BaseWorkflowTestSuite provides test suite for base workflow functionality
type BaseWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up test environment before each test
func (s *BaseWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// AfterTest cleans up after each test
func (s *BaseWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestBaseWorkflowCreation tests base workflow creation
func (s *BaseWorkflowTestSuite) TestBaseWorkflowCreation() {
	metadata := WorkflowMetadata{
		Name:        "TestWorkflow",
		Version:     "1.0.0",
		Description: "Test workflow",
		Author:      "Test Author",
		CreatedAt:   time.Now(),
		Tags: map[string]string{
			"type": "test",
		},
	}

	workflow := NewBaseWorkflow(metadata)

	assert.NotNil(s.T(), workflow)
	assert.Equal(s.T(), "TestWorkflow", workflow.GetName())
	assert.Equal(s.T(), "1.0.0", workflow.GetVersion())
	assert.Equal(s.T(), metadata, workflow.GetMetadata())
	
	state := workflow.GetState()
	assert.Equal(s.T(), "initialization", state.Phase)
	assert.Equal(s.T(), 0, state.Progress)
	assert.Equal(s.T(), WorkflowStatusPending, state.Status)
	assert.NotNil(s.T(), state.Data)
}

// TestBaseWorkflowStateManagement tests state management
func (s *BaseWorkflowTestSuite) TestBaseWorkflowStateManagement() {
	metadata := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	workflow := NewBaseWorkflow(metadata)

	// Test state update
	workflow.UpdateState("processing", 50, WorkflowStatusRunning)
	state := workflow.GetState()
	
	assert.Equal(s.T(), "processing", state.Phase)
	assert.Equal(s.T(), 50, state.Progress)
	assert.Equal(s.T(), WorkflowStatusRunning, state.Status)

	// Test data operations
	workflow.SetStateData("key1", "value1")
	workflow.SetStateData("key2", 42)

	value1, exists1 := workflow.GetStateData("key1")
	assert.True(s.T(), exists1)
	assert.Equal(s.T(), "value1", value1)

	value2, exists2 := workflow.GetStateData("key2")
	assert.True(s.T(), exists2)
	assert.Equal(s.T(), 42, value2)

	_, exists3 := workflow.GetStateData("nonexistent")
	assert.False(s.T(), exists3)
}

// TestBaseWorkflowErrorHandling tests error handling
func (s *BaseWorkflowTestSuite) TestBaseWorkflowErrorHandling() {
	metadata := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	workflow := NewBaseWorkflow(metadata)

	// Test error setting
	testError := assert.AnError
	workflow.SetError(testError)

	state := workflow.GetState()
	assert.Equal(s.T(), WorkflowStatusFailed, state.Status)
	assert.Equal(s.T(), testError.Error(), state.Error)
}

// TestBaseWorkflowQueries tests query setup and handling
func (s *BaseWorkflowTestSuite) TestBaseWorkflowQueries() {
	metadata := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	baseWorkflow := NewBaseWorkflow(metadata)

	// Create a test workflow that uses base queries
	testWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		err := baseWorkflow.SetupBaseQueries(ctx)
		if err != nil {
			return nil, err
		}

		// Update state for testing
		baseWorkflow.UpdateState("running", 75, WorkflowStatusRunning)
		baseWorkflow.SetStateData("test", "value")

		// Wait a bit to allow queries to be processed
		workflow.Sleep(ctx, time.Millisecond*100)

		return "completed", nil
	}

	s.env.RegisterWorkflow(testWorkflow)
	s.env.ExecuteWorkflow(testWorkflow, nil)

	// Test queries
	s.env.RegisterDelayedCallback(func() {
		// Test getState query
		result, err := s.env.QueryWorkflow("getState")
		s.NoError(err)
		
		var state WorkflowState
		err = result.Get(&state)
		s.NoError(err)
		s.Equal("running", state.Phase)
		s.Equal(75, state.Progress)
		s.Equal(WorkflowStatusRunning, state.Status)

		// Test getMetadata query
		metadataResult, err := s.env.QueryWorkflow("getMetadata")
		s.NoError(err)
		
		var retrievedMetadata WorkflowMetadata
		err = metadataResult.Get(&retrievedMetadata)
		s.NoError(err)
		s.Equal("TestWorkflow", retrievedMetadata.Name)

		// Test getProgress query
		progressResult, err := s.env.QueryWorkflow("getProgress")
		s.NoError(err)
		
		var progress int
		err = progressResult.Get(&progress)
		s.NoError(err)
		s.Equal(75, progress)
	}, 50*time.Millisecond)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestBaseWorkflowSignals tests signal setup and handling
func (s *BaseWorkflowTestSuite) TestBaseWorkflowSignals() {
	metadata := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	baseWorkflow := NewBaseWorkflow(metadata)

	testWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		baseWorkflow.SetupBaseSignals(ctx)

		// Wait for signals
		workflow.Sleep(ctx, time.Millisecond*200)

		return baseWorkflow.GetState(), nil
	}

	s.env.RegisterWorkflow(testWorkflow)

	// Send signals during execution
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("pause", "test pause")
	}, 50*time.Millisecond)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("updateData", DataUpdate{Key: "testKey", Value: "testValue"})
	}, 100*time.Millisecond)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("resume", "test resume")
	}, 150*time.Millisecond)

	s.env.ExecuteWorkflow(testWorkflow, nil)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result WorkflowState
	s.NoError(s.env.GetWorkflowResult(&result))
	
	// Verify data was updated
	value, exists := baseWorkflow.GetStateData("testKey")
	s.True(exists)
	s.Equal("testValue", value)
}

// TestDefaultActivityOptions tests default activity options
func (s *BaseWorkflowTestSuite) TestDefaultActivityOptions() {
	opts := DefaultActivityOptions()
	
	assert.Equal(s.T(), 10*time.Minute, opts.StartToCloseTimeout)
	assert.Equal(s.T(), 30*time.Second, opts.HeartbeatTimeout)
	assert.NotNil(s.T(), opts.RetryPolicy)
	assert.Equal(s.T(), time.Second, opts.RetryPolicy.InitialInterval)
	assert.Equal(s.T(), float64(2.0), opts.RetryPolicy.BackoffCoefficient)
	assert.Equal(s.T(), 100*time.Second, opts.RetryPolicy.MaximumInterval)
	assert.Equal(s.T(), int32(3), opts.RetryPolicy.MaximumAttempts)
}

// TestLongRunningActivityOptions tests long-running activity options
func (s *BaseWorkflowTestSuite) TestLongRunningActivityOptions() {
	opts := LongRunningActivityOptions()
	
	assert.Equal(s.T(), 24*time.Hour, opts.StartToCloseTimeout)
	assert.Equal(s.T(), 5*time.Minute, opts.HeartbeatTimeout)
	assert.NotNil(s.T(), opts.RetryPolicy)
	assert.Equal(s.T(), 30*time.Second, opts.RetryPolicy.InitialInterval)
	assert.Equal(s.T(), float64(2.0), opts.RetryPolicy.BackoffCoefficient)
	assert.Equal(s.T(), 10*time.Minute, opts.RetryPolicy.MaximumInterval)
	assert.Equal(s.T(), int32(5), opts.RetryPolicy.MaximumAttempts)
}

// TestHumanTaskActivityOptions tests human task activity options
func (s *BaseWorkflowTestSuite) TestHumanTaskActivityOptions() {
	opts := HumanTaskActivityOptions()
	
	assert.Equal(s.T(), 7*24*time.Hour, opts.StartToCloseTimeout) // 7 days
	assert.Equal(s.T(), time.Hour, opts.HeartbeatTimeout)
	assert.NotNil(s.T(), opts.RetryPolicy)
	assert.Equal(s.T(), 10*time.Minute, opts.RetryPolicy.InitialInterval)
	assert.Equal(s.T(), float64(1.5), opts.RetryPolicy.BackoffCoefficient)
	assert.Equal(s.T(), time.Hour, opts.RetryPolicy.MaximumInterval)
	assert.Equal(s.T(), int32(3), opts.RetryPolicy.MaximumAttempts)
}

// TestWorkflowBuilder tests the workflow builder pattern
func (s *BaseWorkflowTestSuite) TestWorkflowBuilder() {
	builder := NewWorkflowBuilder("TestWorkflow", "1.0.0")
	
	// Build workflow with various configurations
	builder.WithDescription("Test workflow description").
		WithAuthor("Test Author").
		WithTag("type", "test").
		WithTag("category", "unit-test").
		WithActivity("testActivity", func() {}, DefaultActivityOptions(), "Test activity").
		WithSignal("testSignal", func() {}, "Test signal").
		WithQuery("testQuery", func() {}, "Test query")

	metadata, activities, signals, queries, options := builder.Build()

	// Verify metadata
	assert.Equal(s.T(), "TestWorkflow", metadata.Name)
	assert.Equal(s.T(), "1.0.0", metadata.Version)
	assert.Equal(s.T(), "Test workflow description", metadata.Description)
	assert.Equal(s.T(), "Test Author", metadata.Author)
	assert.Equal(s.T(), "test", metadata.Tags["type"])
	assert.Equal(s.T(), "unit-test", metadata.Tags["category"])

	// Verify activities
	assert.Len(s.T(), activities, 1)
	assert.Equal(s.T(), "testActivity", activities[0].Name)
	assert.Equal(s.T(), "Test activity", activities[0].Description)

	// Verify signals
	assert.Len(s.T(), signals, 1)
	assert.Equal(s.T(), "testSignal", signals[0].Name)
	assert.Equal(s.T(), "Test signal", signals[0].Description)

	// Verify queries
	assert.Len(s.T(), queries, 1)
	assert.Equal(s.T(), "testQuery", queries[0].Name)
	assert.Equal(s.T(), "Test query", queries[0].Description)

	// Options should be initialized (even if empty)
	assert.NotNil(s.T(), options)
}

// TestWorkflowRegistry tests the workflow registry
func (s *BaseWorkflowTestSuite) TestWorkflowRegistry() {
	registry := NewWorkflowRegistry()

	// Create test workflows
	metadata1 := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	workflow1 := NewBaseWorkflow(metadata1)

	metadata2 := WorkflowMetadata{Name: "TestWorkflow", Version: "2.0.0"}
	workflow2 := NewBaseWorkflow(metadata2)

	metadata3 := WorkflowMetadata{Name: "AnotherWorkflow", Version: "1.0.0"}
	workflow3 := NewBaseWorkflow(metadata3)

	// Register workflows
	err := registry.Register(workflow1)
	assert.NoError(s.T(), err)

	err = registry.Register(workflow2)
	assert.NoError(s.T(), err)

	err = registry.Register(workflow3)
	assert.NoError(s.T(), err)

	// Test retrieval
	retrieved1, exists := registry.Get("TestWorkflow", "1.0.0")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), workflow1, retrieved1)

	retrieved2, exists := registry.Get("TestWorkflow", "2.0.0")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), workflow2, retrieved2)

	_, exists = registry.Get("NonexistentWorkflow", "1.0.0")
	assert.False(s.T(), exists)

	// Test latest version
	latest, exists := registry.GetLatest("TestWorkflow")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), workflow2, latest) // Should be version 2.0.0

	// Test version listing
	versions := registry.ListVersions("TestWorkflow")
	assert.Len(s.T(), versions, 2)

	// Test workflow listing
	workflows := registry.ListWorkflows()
	assert.Len(s.T(), workflows, 2)
	assert.Contains(s.T(), workflows, "TestWorkflow")
	assert.Contains(s.T(), workflows, "AnotherWorkflow")
}

// TestWorkflowStatus tests workflow status enum
func (s *BaseWorkflowTestSuite) TestWorkflowStatus() {
	statuses := []WorkflowStatus{
		WorkflowStatusPending,
		WorkflowStatusRunning,
		WorkflowStatusCompleted,
		WorkflowStatusFailed,
		WorkflowStatusCancelled,
		WorkflowStatusPaused,
		WorkflowStatusEscalated,
	}

	expectedValues := []string{
		"pending",
		"running",
		"completed",
		"failed",
		"cancelled",
		"paused",
		"escalated",
	}

	for i, status := range statuses {
		assert.Equal(s.T(), expectedValues[i], string(status))
	}
}

// TestPriority tests priority enum
func (s *BaseWorkflowTestSuite) TestPriority() {
	priorities := []Priority{
		PriorityLow,
		PriorityMedium,
		PriorityHigh,
		PriorityCritical,
	}

	expectedValues := []string{
		"low",
		"medium",
		"high",
		"critical",
	}

	for i, priority := range priorities {
		assert.Equal(s.T(), expectedValues[i], string(priority))
	}
}

// TestDataUpdate tests data update signal payload
func (s *BaseWorkflowTestSuite) TestDataUpdate() {
	update := DataUpdate{
		Key:   "testKey",
		Value: "testValue",
	}

	assert.Equal(s.T(), "testKey", update.Key)
	assert.Equal(s.T(), "testValue", update.Value)

	// Test with different value types
	update.Value = 42
	assert.Equal(s.T(), 42, update.Value)

	update.Value = true
	assert.Equal(s.T(), true, update.Value)
}

// TestWorkflowEvent tests workflow event structure
func (s *BaseWorkflowTestSuite) TestWorkflowEvent() {
	timestamp := time.Now()
	event := WorkflowEvent{
		Type:      "test_event",
		Source:    "test_source",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: timestamp,
	}

	assert.Equal(s.T(), "test_event", event.Type)
	assert.Equal(s.T(), "test_source", event.Source)
	assert.Equal(s.T(), timestamp, event.Timestamp)
	assert.Equal(s.T(), "value", event.Data["key"])
}

// MockVersionedWorkflow implements VersionedWorkflow for testing
type MockVersionedWorkflow struct {
	mock.Mock
	*BaseWorkflowImpl
	version WorkflowVersion
}

// NewMockVersionedWorkflow creates a new mock versioned workflow
func NewMockVersionedWorkflow(name string, version WorkflowVersion) *MockVersionedWorkflow {
	metadata := WorkflowMetadata{
		Name:    name,
		Version: version.String(),
	}
	
	return &MockVersionedWorkflow{
		BaseWorkflowImpl: NewBaseWorkflow(metadata),
		version:         version,
	}
}

// Execute implements the Execute method
func (m *MockVersionedWorkflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

// GetWorkflowVersion implements the GetWorkflowVersion method
func (m *MockVersionedWorkflow) GetWorkflowVersion() WorkflowVersion {
	return m.version
}

// GetCompatibleVersions implements the GetCompatibleVersions method
func (m *MockVersionedWorkflow) GetCompatibleVersions() []WorkflowVersion {
	args := m.Called()
	return args.Get(0).([]WorkflowVersion)
}

// SupportsVersioning implements the SupportsVersioning method
func (m *MockVersionedWorkflow) SupportsVersioning() bool {
	args := m.Called()
	return args.Bool(0)
}

// MigrateFromVersion implements the MigrateFromVersion method
func (m *MockVersionedWorkflow) MigrateFromVersion(ctx workflow.Context, fromVersion WorkflowVersion, state interface{}) error {
	args := m.Called(ctx, fromVersion, state)
	return args.Error(0)
}

// TestComposableWorkflowInterface tests the composable workflow interface
func (s *BaseWorkflowTestSuite) TestComposableWorkflowInterface() {
	// This test ensures that the interface is properly defined
	// In a real implementation, you would test the composition logic
	
	var composable ComposableWorkflow
	assert.Nil(s.T(), composable) // Interface should be nil initially
	
	// Test that our base workflow can be extended to implement ComposableWorkflow
	// This would require actual implementation in a real scenario
}

// TestEventDrivenWorkflowInterface tests the event-driven workflow interface
func (s *BaseWorkflowTestSuite) TestEventDrivenWorkflowInterface() {
	// This test ensures that the interface is properly defined
	var eventDriven EventDrivenWorkflow
	assert.Nil(s.T(), eventDriven) // Interface should be nil initially
	
	// In a real implementation, you would test event handling logic
}

// TestStateMachineWorkflowInterface tests the state machine workflow interface
func (s *BaseWorkflowTestSuite) TestStateMachineWorkflowInterface() {
	// This test ensures that the interface is properly defined
	var stateMachine StateMachineWorkflow
	assert.Nil(s.T(), stateMachine) // Interface should be nil initially
	
	// In a real implementation, you would test state transitions
}

// Run the test suite
func TestBaseWorkflowSuite(t *testing.T) {
	suite.Run(t, new(BaseWorkflowTestSuite))
}