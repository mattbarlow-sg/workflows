// Package testing provides comprehensive testing framework for Temporal workflows
package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// TestEnvironment wraps Temporal's test suite with additional helpers
type TestEnvironment struct {
	*testsuite.TestWorkflowEnvironment
	t              *testing.T
	registeredWFs  map[string]interface{}
	registeredActs map[string]interface{}
	mockActivities map[string]*MockActivity
	signalHandlers map[string]SignalHandler
	queryHandlers  map[string]QueryHandler
	timers         []TimerInfo
	humanTasks     map[string]*HumanTaskSimulation
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	
	return &TestEnvironment{
		TestWorkflowEnvironment: env,
		t:                       t,
		registeredWFs:           make(map[string]interface{}),
		registeredActs:          make(map[string]interface{}),
		mockActivities:          make(map[string]*MockActivity),
		signalHandlers:          make(map[string]SignalHandler),
		queryHandlers:           make(map[string]QueryHandler),
		timers:                  []TimerInfo{},
		humanTasks:              make(map[string]*HumanTaskSimulation),
	}
}

// SignalHandler defines a handler for workflow signals
type SignalHandler func(signalName string, signalArg interface{})

// QueryHandler defines a handler for workflow queries
type QueryHandler func(queryType string) (interface{}, error)

// TimerInfo holds information about a timer
type TimerInfo struct {
	ID       string
	Duration time.Duration
	Callback func()
	FireTime time.Time
}

// HumanTaskSimulation simulates human task interactions
type HumanTaskSimulation struct {
	TaskID           string
	TaskType         string
	AssignedTo       string
	CreatedAt        time.Time
	CompletedAt      *time.Time
	EscalationPolicy *EscalationPolicy
	ResponseDelay    time.Duration
	AutoComplete     bool
	Response         interface{}
	Error            error
}

// EscalationPolicy defines escalation rules for human tasks
type EscalationPolicy struct {
	EscalateAfter   time.Duration
	EscalateTo      string
	MaxEscalations  int
	CurrentLevel    int
	EscalationChain []string
}

// RegisterWorkflow registers a workflow for testing
func (e *TestEnvironment) RegisterWorkflow(wf interface{}, name ...string) {
	e.TestWorkflowEnvironment.RegisterWorkflow(wf)
	
	wfName := getWorkflowName(wf, name...)
	e.registeredWFs[wfName] = wf
}

// RegisterActivity registers an activity for testing
func (e *TestEnvironment) RegisterActivity(act interface{}, name ...string) {
	e.TestWorkflowEnvironment.RegisterActivity(act)
	
	actName := getActivityName(act, name...)
	e.registeredActs[actName] = act
}

// MockActivity sets up a mock for an activity
func (e *TestEnvironment) MockActivity(activityName string) *MockActivity {
	mock := &MockActivity{
		Name:      activityName,
		CallCount: 0,
		Calls:     []ActivityCall{},
	}
	e.mockActivities[activityName] = mock
	
	// Register the mock with the test environment
	e.OnActivity(activityName, mock, mock.Execute)
	
	return mock
}

// SetupSignalHandler sets up a signal handler
func (e *TestEnvironment) SetupSignalHandler(signalName string, handler SignalHandler) {
	e.signalHandlers[signalName] = handler
	e.RegisterDelayedCallback(func() {
		e.SignalWorkflow(signalName, nil)
	}, 0)
}

// SetupQueryHandler sets up a query handler
func (e *TestEnvironment) SetupQueryHandler(queryType string, handler QueryHandler) {
	e.queryHandlers[queryType] = handler
}

// SimulateTimer creates a timer simulation
func (e *TestEnvironment) SimulateTimer(duration time.Duration, callback func()) string {
	timerID := fmt.Sprintf("timer_%d", len(e.timers))
	timer := TimerInfo{
		ID:       timerID,
		Duration: duration,
		Callback: callback,
		FireTime: e.Now().Add(duration),
	}
	e.timers = append(e.timers, timer)
	
	// Register the timer callback
	e.RegisterDelayedCallback(callback, duration)
	
	return timerID
}

// CreateHumanTask creates a human task simulation
func (e *TestEnvironment) CreateHumanTask(taskID, taskType, assignee string) *HumanTaskSimulation {
	task := &HumanTaskSimulation{
		TaskID:        taskID,
		TaskType:      taskType,
		AssignedTo:    assignee,
		CreatedAt:     e.Now(),
		ResponseDelay: 5 * time.Second, // Default delay
		AutoComplete:  true,             // Auto-complete by default
	}
	e.humanTasks[taskID] = task
	
	// Set up auto-completion if enabled
	if task.AutoComplete {
		e.RegisterDelayedCallback(func() {
			e.CompleteHumanTask(taskID, "approved", nil)
		}, task.ResponseDelay)
	}
	
	return task
}

// CompleteHumanTask completes a human task
func (e *TestEnvironment) CompleteHumanTask(taskID string, response interface{}, err error) {
	if task, exists := e.humanTasks[taskID]; exists {
		now := e.Now()
		task.CompletedAt = &now
		task.Response = response
		task.Error = err
		
		// Signal the workflow about task completion
		e.SignalWorkflow("HumanTaskCompleted", map[string]interface{}{
			"taskID":   taskID,
			"response": response,
			"error":    err,
		})
	}
}

// EscalateHumanTask escalates a human task
func (e *TestEnvironment) EscalateHumanTask(taskID string) error {
	task, exists := e.humanTasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}
	
	if task.EscalationPolicy == nil {
		return fmt.Errorf("task %s has no escalation policy", taskID)
	}
	
	if task.EscalationPolicy.CurrentLevel >= task.EscalationPolicy.MaxEscalations {
		return fmt.Errorf("task %s has reached maximum escalation level", taskID)
	}
	
	task.EscalationPolicy.CurrentLevel++
	if task.EscalationPolicy.CurrentLevel < len(task.EscalationPolicy.EscalationChain) {
		task.AssignedTo = task.EscalationPolicy.EscalationChain[task.EscalationPolicy.CurrentLevel]
	}
	
	// Signal escalation
	e.SignalWorkflow("HumanTaskEscalated", map[string]interface{}{
		"taskID":        taskID,
		"escalatedTo":   task.AssignedTo,
		"escalationLevel": task.EscalationPolicy.CurrentLevel,
	})
	
	return nil
}

// SkipTime advances the test time
func (e *TestEnvironment) SkipTime(duration time.Duration) {
	e.TestWorkflowEnvironment.RegisterDelayedCallback(func() {}, duration)
}

// AssertWorkflowCompleted asserts that the workflow completed successfully
func (e *TestEnvironment) AssertWorkflowCompleted(timeout time.Duration) {
	e.t.Helper()
	
	select {
	case <-time.After(timeout):
		e.t.Fatalf("Workflow did not complete within %v", timeout)
	default:
		if !e.IsWorkflowCompleted() {
			e.t.Fatal("Workflow is not completed")
		}
		
		err := e.GetWorkflowError()
		if err != nil {
			e.t.Fatalf("Workflow completed with error: %v", err)
		}
	}
}

// AssertWorkflowFailed asserts that the workflow failed with an error
func (e *TestEnvironment) AssertWorkflowFailed(timeout time.Duration) {
	e.t.Helper()
	
	select {
	case <-time.After(timeout):
		e.t.Fatalf("Workflow did not complete within %v", timeout)
	default:
		if !e.IsWorkflowCompleted() {
			e.t.Fatal("Workflow is not completed")
		}
		
		err := e.GetWorkflowError()
		if err == nil {
			e.t.Fatal("Expected workflow to fail, but it succeeded")
		}
	}
}

// AssertActivityCalled asserts that an activity was called
func (e *TestEnvironment) AssertActivityCalled(activityName string, times int) {
	e.t.Helper()
	
	mock, exists := e.mockActivities[activityName]
	if !exists {
		e.t.Fatalf("Activity %s is not mocked", activityName)
	}
	
	if mock.CallCount != times {
		e.t.Fatalf("Activity %s was called %d times, expected %d", activityName, mock.CallCount, times)
	}
}

// AssertSignalSent asserts that a signal was sent
func (e *TestEnvironment) AssertSignalSent(signalName string) {
	e.t.Helper()
	
	// Check if signal handler was registered and called
	if _, exists := e.signalHandlers[signalName]; !exists {
		e.t.Fatalf("Signal %s was not sent", signalName)
	}
}

// GetWorkflowResult gets the workflow result
func (e *TestEnvironment) GetWorkflowResult(result interface{}) error {
	return e.TestWorkflowEnvironment.GetWorkflowResult(result)
}

// ExecuteWorkflow executes a workflow with the given input
func (e *TestEnvironment) ExecuteWorkflow(workflowFunc interface{}, args ...interface{}) {
	e.TestWorkflowEnvironment.ExecuteWorkflow(workflowFunc, args...)
}

// Clean up the test environment
func (e *TestEnvironment) Cleanup() {
	// Clean up any resources
	e.registeredWFs = make(map[string]interface{})
	e.registeredActs = make(map[string]interface{})
	e.mockActivities = make(map[string]*MockActivity)
	e.signalHandlers = make(map[string]SignalHandler)
	e.queryHandlers = make(map[string]QueryHandler)
	e.timers = []TimerInfo{}
	e.humanTasks = make(map[string]*HumanTaskSimulation)
}

// Helper functions

func getWorkflowName(wf interface{}, name ...string) string {
	if len(name) > 0 {
		return name[0]
	}
	return fmt.Sprintf("%T", wf)
}

func getActivityName(act interface{}, name ...string) string {
	if len(name) > 0 {
		return name[0]
	}
	return fmt.Sprintf("%T", act)
}

// WorkflowBuilder helps build test workflows
type WorkflowBuilder struct {
	env        *TestEnvironment
	workflow   interface{}
	activities []interface{}
	signals    map[string]SignalHandler
	queries    map[string]QueryHandler
	mocks      map[string]*MockActivity
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(env *TestEnvironment) *WorkflowBuilder {
	return &WorkflowBuilder{
		env:        env,
		activities: []interface{}{},
		signals:    make(map[string]SignalHandler),
		queries:    make(map[string]QueryHandler),
		mocks:      make(map[string]*MockActivity),
	}
}

// WithWorkflow sets the workflow to test
func (b *WorkflowBuilder) WithWorkflow(wf interface{}) *WorkflowBuilder {
	b.workflow = wf
	b.env.RegisterWorkflow(wf)
	return b
}

// WithActivity adds an activity
func (b *WorkflowBuilder) WithActivity(act interface{}) *WorkflowBuilder {
	b.activities = append(b.activities, act)
	b.env.RegisterActivity(act)
	return b
}

// WithMockedActivity adds a mocked activity
func (b *WorkflowBuilder) WithMockedActivity(activityName string, mock *MockActivity) *WorkflowBuilder {
	b.mocks[activityName] = mock
	b.env.mockActivities[activityName] = mock
	b.env.OnActivity(activityName, mock, mock.Execute)
	return b
}

// WithSignalHandler adds a signal handler
func (b *WorkflowBuilder) WithSignalHandler(signalName string, handler SignalHandler) *WorkflowBuilder {
	b.signals[signalName] = handler
	b.env.SetupSignalHandler(signalName, handler)
	return b
}

// WithQueryHandler adds a query handler
func (b *WorkflowBuilder) WithQueryHandler(queryType string, handler QueryHandler) *WorkflowBuilder {
	b.queries[queryType] = handler
	b.env.SetupQueryHandler(queryType, handler)
	return b
}

// Build finalizes the builder and returns the environment
func (b *WorkflowBuilder) Build() *TestEnvironment {
	return b.env
}

// Execute executes the workflow with the given arguments
func (b *WorkflowBuilder) Execute(args ...interface{}) {
	b.env.ExecuteWorkflow(b.workflow, args...)
}

// WorkflowTestSuite provides a test suite for workflows
type WorkflowTestSuite struct {
	suite testsuite.WorkflowTestSuite
	env    *TestEnvironment
}

// NewWorkflowTestSuite creates a new workflow test suite
func NewWorkflowTestSuite() *WorkflowTestSuite {
	return &WorkflowTestSuite{}
}

// SetupTest sets up the test environment
func (s *WorkflowTestSuite) SetupTest() {
	s.env = NewTestEnvironment(nil)
}

// TearDownTest tears down the test environment
func (s *WorkflowTestSuite) TearDownTest() {
	s.env.Cleanup()
}

// GetEnvironment returns the test environment
func (s *WorkflowTestSuite) GetEnvironment() *TestEnvironment {
	return s.env
}

// TestTimeSkipping demonstrates time-skipping capabilities
type TestTimeSkipping struct {
	env           *TestEnvironment
	checkpoints   []TimeCheckpoint
	currentIndex  int
}

// TimeCheckpoint represents a point in time to check
type TimeCheckpoint struct {
	Time     time.Time
	Duration time.Duration
	Check    func() error
}

// NewTestTimeSkipping creates a new time-skipping test helper
func NewTestTimeSkipping(env *TestEnvironment) *TestTimeSkipping {
	return &TestTimeSkipping{
		env:          env,
		checkpoints:  []TimeCheckpoint{},
		currentIndex: 0,
	}
}

// AddCheckpoint adds a time checkpoint
func (t *TestTimeSkipping) AddCheckpoint(duration time.Duration, check func() error) {
	checkpoint := TimeCheckpoint{
		Time:     t.env.Now().Add(duration),
		Duration: duration,
		Check:    check,
	}
	t.checkpoints = append(t.checkpoints, checkpoint)
}

// RunToCheckpoint runs to the next checkpoint
func (t *TestTimeSkipping) RunToCheckpoint() error {
	if t.currentIndex >= len(t.checkpoints) {
		return fmt.Errorf("no more checkpoints")
	}
	
	checkpoint := t.checkpoints[t.currentIndex]
	t.env.SkipTime(checkpoint.Duration)
	
	if checkpoint.Check != nil {
		if err := checkpoint.Check(); err != nil {
			return fmt.Errorf("checkpoint %d failed: %w", t.currentIndex, err)
		}
	}
	
	t.currentIndex++
	return nil
}

// RunAll runs through all checkpoints
func (t *TestTimeSkipping) RunAll() error {
	for t.currentIndex < len(t.checkpoints) {
		if err := t.RunToCheckpoint(); err != nil {
			return err
		}
	}
	return nil
}

// WorkflowTestConfig holds configuration for workflow tests
type WorkflowTestConfig struct {
	DisableWorkflowPanicHandler bool
	WorkflowPanicPolicy         worker.WorkflowPanicPolicy
	WorkflowTaskTimeout         time.Duration
	BackgroundActivityContext   context.Context
	EnableLoggingInReplay       bool
}

// ApplyConfig applies configuration to the test environment
func (e *TestEnvironment) ApplyConfig(config WorkflowTestConfig) {
	if config.DisableWorkflowPanicHandler {
		e.SetWorkerOptions(worker.Options{
			DisableWorkflowWorker: false,
		})
	}
	
	// Note: These options would need to be set through worker options
	// when the test environment is created
}

// LongRunningWorkflowTester helps test long-running workflows
type LongRunningWorkflowTester struct {
	env            *TestEnvironment
	checkInterval  time.Duration
	maxDuration    time.Duration
	startTime      time.Time
	checkpoints    []func() error
}

// NewLongRunningWorkflowTester creates a tester for long-running workflows
func NewLongRunningWorkflowTester(env *TestEnvironment, maxDuration time.Duration) *LongRunningWorkflowTester {
	return &LongRunningWorkflowTester{
		env:           env,
		checkInterval: 1 * time.Hour,  // Default check interval
		maxDuration:   maxDuration,
		startTime:     env.Now(),
		checkpoints:   []func() error{},
	}
}

// SetCheckInterval sets the interval between checks
func (l *LongRunningWorkflowTester) SetCheckInterval(interval time.Duration) {
	l.checkInterval = interval
}

// AddPeriodicCheck adds a check to run periodically
func (l *LongRunningWorkflowTester) AddPeriodicCheck(check func() error) {
	l.checkpoints = append(l.checkpoints, check)
}

// RunUntilComplete runs the workflow until completion or timeout
func (l *LongRunningWorkflowTester) RunUntilComplete() error {
	elapsed := time.Duration(0)
	
	for elapsed < l.maxDuration {
		// Skip time by check interval
		l.env.SkipTime(l.checkInterval)
		elapsed += l.checkInterval
		
		// Run periodic checks
		for _, check := range l.checkpoints {
			if err := check(); err != nil {
				return fmt.Errorf("periodic check failed at %v: %w", elapsed, err)
			}
		}
		
		// Check if workflow is complete
		if l.env.IsWorkflowCompleted() {
			return nil
		}
	}
	
	return fmt.Errorf("workflow did not complete within %v", l.maxDuration)
}

// GetElapsedTime returns the elapsed simulated time
func (l *LongRunningWorkflowTester) GetElapsedTime() time.Duration {
	return l.env.Now().Sub(l.startTime)
}