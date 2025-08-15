// Package testing provides helper utilities for testing Temporal workflows
package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/temporal"
)

// WorkflowTestHelper provides helper methods for workflow testing
type WorkflowTestHelper struct {
	t   *testing.T
	env *TestEnvironment
}

// NewWorkflowTestHelper creates a new workflow test helper
func NewWorkflowTestHelper(t *testing.T) *WorkflowTestHelper {
	return &WorkflowTestHelper{
		t:   t,
		env: NewTestEnvironment(t),
	}
}

// GetEnvironment returns the test environment
func (h *WorkflowTestHelper) GetEnvironment() *TestEnvironment {
	return h.env
}

// TestWorkflowScenario tests a complete workflow scenario
func (h *WorkflowTestHelper) TestWorkflowScenario(scenario WorkflowScenario) {
	h.t.Helper()
	
	// Setup
	if scenario.Setup != nil {
		scenario.Setup(h.env)
	}
	
	// Execute workflow
	h.env.ExecuteWorkflow(scenario.Workflow, scenario.Input...)
	
	// Perform checks during execution
	if scenario.DuringExecution != nil {
		scenario.DuringExecution(h.env)
	}
	
	// Wait for completion
	h.env.AssertWorkflowCompleted(scenario.Timeout)
	
	// Verify results
	if scenario.VerifyResults != nil {
		scenario.VerifyResults(h.env)
	}
	
	// Cleanup
	if scenario.Cleanup != nil {
		scenario.Cleanup(h.env)
	}
}

// WorkflowScenario defines a test scenario
type WorkflowScenario struct {
	Name            string
	Workflow        interface{}
	Input           []interface{}
	Setup           func(*TestEnvironment)
	DuringExecution func(*TestEnvironment)
	VerifyResults   func(*TestEnvironment)
	Cleanup         func(*TestEnvironment)
	Timeout         time.Duration
}

// SignalTestHelper helps test signal handling
type SignalTestHelper struct {
	env     *TestEnvironment
	signals []SignalTest
}

// SignalTest represents a signal test case
type SignalTest struct {
	Name      string
	SignalArg interface{}
	Delay     time.Duration
	Verify    func() error
}

// NewSignalTestHelper creates a new signal test helper
func NewSignalTestHelper(env *TestEnvironment) *SignalTestHelper {
	return &SignalTestHelper{
		env:     env,
		signals: []SignalTest{},
	}
}

// AddSignal adds a signal to test
func (s *SignalTestHelper) AddSignal(name string, arg interface{}, delay time.Duration) *SignalTestHelper {
	s.signals = append(s.signals, SignalTest{
		Name:      name,
		SignalArg: arg,
		Delay:     delay,
	})
	return s
}

// AddSignalWithVerification adds a signal with verification
func (s *SignalTestHelper) AddSignalWithVerification(name string, arg interface{}, delay time.Duration, verify func() error) *SignalTestHelper {
	s.signals = append(s.signals, SignalTest{
		Name:      name,
		SignalArg: arg,
		Delay:     delay,
		Verify:    verify,
	})
	return s
}

// Execute executes all signals in sequence
func (s *SignalTestHelper) Execute() error {
	for _, signal := range s.signals {
		// Wait for delay
		if signal.Delay > 0 {
			s.env.SkipTime(signal.Delay)
		}
		
		// Send signal
		s.env.SignalWorkflow(signal.Name, signal.SignalArg)
		
		// Verify if needed
		if signal.Verify != nil {
			if err := signal.Verify(); err != nil {
				return fmt.Errorf("signal %s verification failed: %w", signal.Name, err)
			}
		}
	}
	return nil
}

// QueryTestHelper helps test query handling
type QueryTestHelper struct {
	env     *TestEnvironment
	queries []QueryTest
}

// QueryTest represents a query test case
type QueryTest struct {
	Type     string
	Expected interface{}
	Verify   func(result interface{}) error
}

// NewQueryTestHelper creates a new query test helper
func NewQueryTestHelper(env *TestEnvironment) *QueryTestHelper {
	return &QueryTestHelper{
		env:     env,
		queries: []QueryTest{},
	}
}

// AddQuery adds a query to test
func (q *QueryTestHelper) AddQuery(queryType string, expected interface{}) *QueryTestHelper {
	q.queries = append(q.queries, QueryTest{
		Type:     queryType,
		Expected: expected,
	})
	return q
}

// AddQueryWithVerification adds a query with custom verification
func (q *QueryTestHelper) AddQueryWithVerification(queryType string, verify func(result interface{}) error) *QueryTestHelper {
	q.queries = append(q.queries, QueryTest{
		Type:   queryType,
		Verify: verify,
	})
	return q
}

// Execute executes all queries and verifies results
func (q *QueryTestHelper) Execute() error {
	for _, query := range q.queries {
		result, err := q.env.QueryWorkflow(query.Type)
		if err != nil {
			return fmt.Errorf("query %s failed: %w", query.Type, err)
		}
		
		if query.Verify != nil {
			if err := query.Verify(result); err != nil {
				return fmt.Errorf("query %s verification failed: %w", query.Type, err)
			}
		} else if query.Expected != nil {
			if !reflect.DeepEqual(result, query.Expected) {
				return fmt.Errorf("query %s: expected %v, got %v", query.Type, query.Expected, result)
			}
		}
	}
	return nil
}

// HumanTaskTestHelper helps test human task interactions
type HumanTaskTestHelper struct {
	env   *TestEnvironment
	tasks map[string]*HumanTaskTest
}

// HumanTaskTest represents a human task test
type HumanTaskTest struct {
	TaskID          string
	TaskType        string
	Assignee        string
	ResponseDelay   time.Duration
	Response        interface{}
	ShouldEscalate  bool
	EscalationDelay time.Duration
	EscalateTo      string
}

// NewHumanTaskTestHelper creates a new human task test helper
func NewHumanTaskTestHelper(env *TestEnvironment) *HumanTaskTestHelper {
	return &HumanTaskTestHelper{
		env:   env,
		tasks: make(map[string]*HumanTaskTest),
	}
}

// AddTask adds a human task test
func (h *HumanTaskTestHelper) AddTask(task *HumanTaskTest) *HumanTaskTestHelper {
	h.tasks[task.TaskID] = task
	return h
}

// SimulateTask simulates a human task
func (h *HumanTaskTestHelper) SimulateTask(taskID string) error {
	task, exists := h.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}
	
	// Create the task
	humanTask := h.env.CreateHumanTask(task.TaskID, task.TaskType, task.Assignee)
	humanTask.ResponseDelay = task.ResponseDelay
	humanTask.AutoComplete = !task.ShouldEscalate
	
	if task.ShouldEscalate {
		// Set up escalation
		humanTask.EscalationPolicy = &EscalationPolicy{
			EscalateAfter:  task.EscalationDelay,
			EscalateTo:     task.EscalateTo,
			MaxEscalations: 3,
		}
		
		// Schedule escalation
		h.env.RegisterDelayedCallback(func() {
			h.env.EscalateHumanTask(taskID)
		}, task.EscalationDelay)
		
		// Schedule completion after escalation
		h.env.RegisterDelayedCallback(func() {
			h.env.CompleteHumanTask(taskID, task.Response, nil)
		}, task.EscalationDelay+task.ResponseDelay)
	}
	
	return nil
}

// SimulateAllTasks simulates all configured tasks
func (h *HumanTaskTestHelper) SimulateAllTasks() error {
	for taskID := range h.tasks {
		if err := h.SimulateTask(taskID); err != nil {
			return err
		}
	}
	return nil
}

// ActivityTestHelper helps test activities
type ActivityTestHelper struct {
	t          *testing.T
	activities map[string]*ActivityTestCase
}

// ActivityTestCase represents an activity test case
type ActivityTestCase struct {
	Name           string
	Function       interface{}
	Input          []interface{}
	ExpectedOutput interface{}
	ExpectedError  error
	Context        context.Context
	Timeout        time.Duration
}

// NewActivityTestHelper creates a new activity test helper
func NewActivityTestHelper(t *testing.T) *ActivityTestHelper {
	return &ActivityTestHelper{
		t:          t,
		activities: make(map[string]*ActivityTestCase),
	}
}

// AddTestCase adds an activity test case
func (a *ActivityTestHelper) AddTestCase(testCase *ActivityTestCase) *ActivityTestHelper {
	a.activities[testCase.Name] = testCase
	return a
}

// TestActivity tests a single activity
func (a *ActivityTestHelper) TestActivity(name string) {
	a.t.Helper()
	
	testCase, exists := a.activities[name]
	if !exists {
		a.t.Fatalf("Test case %s not found", name)
	}
	
	ctx := testCase.Context
	if ctx == nil {
		ctx = context.Background()
	}
	
	if testCase.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, testCase.Timeout)
		defer cancel()
	}
	
	// Use reflection to call the activity
	fn := reflect.ValueOf(testCase.Function)
	fnType := fn.Type()
	
	// Prepare arguments
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, input := range testCase.Input {
		args = append(args, reflect.ValueOf(input))
	}
	
	// Call the activity
	results := fn.Call(args)
	
	// Check results
	if fnType.NumOut() == 2 {
		// Function returns (result, error)
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		
		if testCase.ExpectedError != nil {
			if err == nil {
				a.t.Errorf("Expected error %v, got nil", testCase.ExpectedError)
			} else if err.Error() != testCase.ExpectedError.Error() {
				a.t.Errorf("Expected error %v, got %v", testCase.ExpectedError, err)
			}
		} else if err != nil {
			a.t.Errorf("Unexpected error: %v", err)
		}
		
		if testCase.ExpectedOutput != nil && err == nil {
			result := results[0].Interface()
			if !reflect.DeepEqual(result, testCase.ExpectedOutput) {
				a.t.Errorf("Expected output %v, got %v", testCase.ExpectedOutput, result)
			}
		}
	}
}

// TestAll tests all registered activities
func (a *ActivityTestHelper) TestAll() {
	for name := range a.activities {
		a.t.Run(name, func(t *testing.T) {
			a.TestActivity(name)
		})
	}
}

// ErrorMatcher helps match errors in tests
type ErrorMatcher struct {
	patterns []string
}

// NewErrorMatcher creates a new error matcher
func NewErrorMatcher(patterns ...string) *ErrorMatcher {
	return &ErrorMatcher{
		patterns: patterns,
	}
}

// Matches checks if an error matches the patterns
func (m *ErrorMatcher) Matches(err error) bool {
	if err == nil {
		return len(m.patterns) == 0
	}
	
	errStr := err.Error()
	for _, pattern := range m.patterns {
		if !strings.Contains(errStr, pattern) {
			return false
		}
	}
	
	return true
}

// WorkflowStateAssertion helps assert workflow state
type WorkflowStateAssertion struct {
	env *TestEnvironment
}

// NewWorkflowStateAssertion creates a new state assertion helper
func NewWorkflowStateAssertion(env *TestEnvironment) *WorkflowStateAssertion {
	return &WorkflowStateAssertion{env: env}
}

// AssertState asserts the workflow is in a specific state
func (a *WorkflowStateAssertion) AssertState(expected string) error {
	result, err := a.env.QueryWorkflow("state")
	if err != nil {
		return fmt.Errorf("failed to query state: %w", err)
	}
	
	// Result could be an encoded value or a string
	state := fmt.Sprintf("%v", result)
	
	if state != expected {
		return fmt.Errorf("expected state %s, got %s", expected, state)
	}
	
	return nil
}

// AssertCompleted asserts the workflow completed successfully
func (a *WorkflowStateAssertion) AssertCompleted() error {
	if !a.env.IsWorkflowCompleted() {
		return fmt.Errorf("workflow is not completed")
	}
	
	if err := a.env.GetWorkflowError(); err != nil {
		return fmt.Errorf("workflow completed with error: %w", err)
	}
	
	return nil
}

// AssertFailed asserts the workflow failed
func (a *WorkflowStateAssertion) AssertFailed() error {
	if !a.env.IsWorkflowCompleted() {
		return fmt.Errorf("workflow is not completed")
	}
	
	if err := a.env.GetWorkflowError(); err == nil {
		return fmt.Errorf("expected workflow to fail, but it succeeded")
	}
	
	return nil
}

// DataDrivenTest represents a data-driven test
type DataDrivenTest struct {
	Name     string
	TestData []TestData
	TestFunc func(data TestData) error
}

// TestData represents test data for a single test case
type TestData struct {
	Name           string
	Input          interface{}
	ExpectedOutput interface{}
	ExpectedError  error
	Context        map[string]interface{}
}

// RunDataDrivenTests runs data-driven tests
func RunDataDrivenTests(t *testing.T, tests []DataDrivenTest) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			for _, data := range test.TestData {
				t.Run(data.Name, func(t *testing.T) {
					if err := test.TestFunc(data); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}

// WorkflowSnapshot captures workflow state for comparison
type WorkflowSnapshot struct {
	State      string
	Variables  map[string]interface{}
	Activities []string
	Signals    []string
	Timers     []string
	Timestamp  time.Time
}

// CaptureSnapshot captures the current workflow state
func CaptureSnapshot(env *TestEnvironment) (*WorkflowSnapshot, error) {
	snapshot := &WorkflowSnapshot{
		Variables:  make(map[string]interface{}),
		Activities: []string{},
		Signals:    []string{},
		Timers:     []string{},
		Timestamp:  env.Now(),
	}
	
	// Query workflow state
	if state, err := env.QueryWorkflow("state"); err == nil {
		snapshot.State = fmt.Sprintf("%v", state)
	}
	
	// Capture activities
	for name := range env.mockActivities {
		snapshot.Activities = append(snapshot.Activities, name)
	}
	
	// Capture signals
	for name := range env.signalHandlers {
		snapshot.Signals = append(snapshot.Signals, name)
	}
	
	// Capture timers
	for _, timer := range env.timers {
		snapshot.Timers = append(snapshot.Timers, timer.ID)
	}
	
	return snapshot, nil
}

// CompareSnapshots compares two workflow snapshots
func CompareSnapshots(before, after *WorkflowSnapshot) []string {
	differences := []string{}
	
	if before.State != after.State {
		differences = append(differences, fmt.Sprintf("State changed from %s to %s", before.State, after.State))
	}
	
	// Compare activities
	beforeActivities := make(map[string]bool)
	for _, act := range before.Activities {
		beforeActivities[act] = true
	}
	
	for _, act := range after.Activities {
		if !beforeActivities[act] {
			differences = append(differences, fmt.Sprintf("New activity: %s", act))
		}
	}
	
	// Compare signals
	beforeSignals := make(map[string]bool)
	for _, sig := range before.Signals {
		beforeSignals[sig] = true
	}
	
	for _, sig := range after.Signals {
		if !beforeSignals[sig] {
			differences = append(differences, fmt.Sprintf("New signal: %s", sig))
		}
	}
	
	return differences
}

// RetryTestHelper helps test retry behavior
type RetryTestHelper struct {
	env            *TestEnvironment
	activityName   string
	failureCount   int
	successAfter   int
	errorToReturn  error
}

// NewRetryTestHelper creates a new retry test helper
func NewRetryTestHelper(env *TestEnvironment, activityName string) *RetryTestHelper {
	return &RetryTestHelper{
		env:          env,
		activityName: activityName,
		failureCount: 0,
		successAfter: 3,
	}
}

// SetSuccessAfter sets after how many attempts the activity should succeed
func (r *RetryTestHelper) SetSuccessAfter(attempts int) *RetryTestHelper {
	r.successAfter = attempts
	return r
}

// SetError sets the error to return on failure
func (r *RetryTestHelper) SetError(err error) *RetryTestHelper {
	r.errorToReturn = err
	return r
}

// Setup sets up the retry mock
func (r *RetryTestHelper) Setup() *MockActivity {
	mock := r.env.MockActivity(r.activityName)
	
	mock.WithSideEffect(func(args []interface{}) error {
		r.failureCount++
		if r.failureCount < r.successAfter {
			if r.errorToReturn != nil {
				return r.errorToReturn
			}
			return temporal.NewApplicationError("temporary failure", "TEMPORARY", nil)
		}
		return nil
	})
	
	mock.Returns("success")
	
	return mock
}

// AssertRetriedExpectedTimes asserts the activity was retried the expected number of times
func (r *RetryTestHelper) AssertRetriedExpectedTimes(t *testing.T) {
	mock := r.env.mockActivities[r.activityName]
	if mock == nil {
		t.Fatal("Activity not mocked")
	}
	
	if mock.CallCount != r.successAfter {
		t.Errorf("Expected %d retries, got %d", r.successAfter, mock.CallCount)
	}
}

// JSONHelper provides JSON comparison utilities
type JSONHelper struct{}

// NewJSONHelper creates a new JSON helper
func NewJSONHelper() *JSONHelper {
	return &JSONHelper{}
}

// CompareJSON compares two JSON strings
func (j *JSONHelper) CompareJSON(expected, actual string) error {
	var expectedObj, actualObj interface{}
	
	if err := json.Unmarshal([]byte(expected), &expectedObj); err != nil {
		return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
	}
	
	if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
		return fmt.Errorf("failed to unmarshal actual JSON: %w", err)
	}
	
	if !reflect.DeepEqual(expectedObj, actualObj) {
		return fmt.Errorf("JSON objects are not equal")
	}
	
	return nil
}

// MarshalAndCompare marshals an object and compares with expected JSON
func (j *JSONHelper) MarshalAndCompare(obj interface{}, expected string) error {
	actual, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}
	
	return j.CompareJSON(expected, string(actual))
}