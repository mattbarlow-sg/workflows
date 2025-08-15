// Package testing provides mock implementations for Temporal testing
package testing

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.temporal.io/sdk/workflow"
)

// MockActivity represents a mocked activity
type MockActivity struct {
	Name            string
	CallCount       int
	Calls           []ActivityCall
	ReturnValues    []interface{}
	ReturnError     error
	DelayFunc       func(callIndex int) time.Duration
	ValidationFunc  func(args []interface{}) error
	SideEffectFunc  func(args []interface{}) error
	mu              sync.Mutex
}

// ActivityCall records a single activity call
type ActivityCall struct {
	Arguments  []interface{}
	CallTime   time.Time
	ReturnTime time.Time
	Error      error
}

// Execute executes the mock activity
func (m *MockActivity) Execute(ctx context.Context, args ...interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	call := ActivityCall{
		Arguments: args,
		CallTime:  time.Now(),
	}
	
	// Apply validation if configured
	if m.ValidationFunc != nil {
		if err := m.ValidationFunc(args); err != nil {
			call.Error = err
			m.Calls = append(m.Calls, call)
			return nil, err
		}
	}
	
	// Apply side effects if configured
	if m.SideEffectFunc != nil {
		if err := m.SideEffectFunc(args); err != nil {
			call.Error = err
			m.Calls = append(m.Calls, call)
			return nil, err
		}
	}
	
	// Apply delay if configured
	if m.DelayFunc != nil {
		delay := m.DelayFunc(m.CallCount)
		time.Sleep(delay)
	}
	
	m.CallCount++
	call.ReturnTime = time.Now()
	
	// Return configured values
	var result interface{}
	if len(m.ReturnValues) > 0 {
		if m.CallCount <= len(m.ReturnValues) {
			result = m.ReturnValues[m.CallCount-1]
		} else {
			result = m.ReturnValues[len(m.ReturnValues)-1] // Return last value for subsequent calls
		}
	}
	
	call.Error = m.ReturnError
	m.Calls = append(m.Calls, call)
	
	return result, m.ReturnError
}

// Returns sets the return values for the mock
func (m *MockActivity) Returns(values ...interface{}) *MockActivity {
	m.ReturnValues = values
	return m
}

// ReturnsError sets the error to return
func (m *MockActivity) ReturnsError(err error) *MockActivity {
	m.ReturnError = err
	return m
}

// WithDelay adds a delay function
func (m *MockActivity) WithDelay(delayFunc func(callIndex int) time.Duration) *MockActivity {
	m.DelayFunc = delayFunc
	return m
}

// WithValidation adds input validation
func (m *MockActivity) WithValidation(validationFunc func(args []interface{}) error) *MockActivity {
	m.ValidationFunc = validationFunc
	return m
}

// WithSideEffect adds a side effect function
func (m *MockActivity) WithSideEffect(sideEffectFunc func(args []interface{}) error) *MockActivity {
	m.SideEffectFunc = sideEffectFunc
	return m
}

// AssertCalled asserts the activity was called with specific arguments
func (m *MockActivity) AssertCalled(t TestingT, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, call := range m.Calls {
		if reflect.DeepEqual(call.Arguments, args) {
			return
		}
	}
	
	t.Errorf("Activity %s was not called with arguments %v", m.Name, args)
}

// AssertNotCalled asserts the activity was not called
func (m *MockActivity) AssertNotCalled(t TestingT) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.CallCount > 0 {
		t.Errorf("Activity %s was called %d times, expected 0", m.Name, m.CallCount)
	}
}

// AssertNumberOfCalls asserts the activity was called a specific number of times
func (m *MockActivity) AssertNumberOfCalls(t TestingT, expected int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.CallCount != expected {
		t.Errorf("Activity %s was called %d times, expected %d", m.Name, m.CallCount, expected)
	}
}

// GetCall returns a specific call by index
func (m *MockActivity) GetCall(index int) *ActivityCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if index < 0 || index >= len(m.Calls) {
		return nil
	}
	return &m.Calls[index]
}

// Reset resets the mock state
func (m *MockActivity) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.CallCount = 0
	m.Calls = []ActivityCall{}
}

// TestingT is a subset of testing.T for assertions
type TestingT interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// MockWorkflow represents a mocked workflow
type MockWorkflow struct {
	Name         string
	CallCount    int
	Calls        []WorkflowCall
	ReturnValues []interface{}
	ReturnError  error
	mu           sync.Mutex
}

// WorkflowCall records a single workflow call
type WorkflowCall struct {
	Arguments  []interface{}
	CallTime   time.Time
	ReturnTime time.Time
	Error      error
}

// Execute executes the mock workflow
func (m *MockWorkflow) Execute(ctx workflow.Context, args ...interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	call := WorkflowCall{
		Arguments: args,
		CallTime:  workflow.Now(ctx),
	}
	
	m.CallCount++
	call.ReturnTime = workflow.Now(ctx)
	
	// Return configured values
	var result interface{}
	if len(m.ReturnValues) > 0 {
		if m.CallCount <= len(m.ReturnValues) {
			result = m.ReturnValues[m.CallCount-1]
		} else {
			result = m.ReturnValues[len(m.ReturnValues)-1]
		}
	}
	
	call.Error = m.ReturnError
	m.Calls = append(m.Calls, call)
	
	return result, m.ReturnError
}

// Returns sets the return values for the mock workflow
func (m *MockWorkflow) Returns(values ...interface{}) *MockWorkflow {
	m.ReturnValues = values
	return m
}

// ReturnsError sets the error to return
func (m *MockWorkflow) ReturnsError(err error) *MockWorkflow {
	m.ReturnError = err
	return m
}

// MockSignal represents a mocked signal
type MockSignal struct {
	Name       string
	Received   []SignalData
	mu         sync.Mutex
}

// SignalData represents data from a signal
type SignalData struct {
	Data      interface{}
	Timestamp time.Time
}

// Send sends a signal
func (m *MockSignal) Send(data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Received = append(m.Received, SignalData{
		Data:      data,
		Timestamp: time.Now(),
	})
}

// AssertReceived asserts a signal was received
func (m *MockSignal) AssertReceived(t TestingT, expected interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, signal := range m.Received {
		if reflect.DeepEqual(signal.Data, expected) {
			return
		}
	}
	
	t.Errorf("Signal %s did not receive expected data: %v", m.Name, expected)
}

// MockQuery represents a mocked query
type MockQuery struct {
	Name         string
	Handler      func() (interface{}, error)
	CallCount    int
	mu           sync.Mutex
}

// Handle handles a query
func (m *MockQuery) Handle() (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.CallCount++
	
	if m.Handler != nil {
		return m.Handler()
	}
	
	return nil, nil
}

// SetHandler sets the query handler
func (m *MockQuery) SetHandler(handler func() (interface{}, error)) {
	m.Handler = handler
}

// MockActivityFactory creates mock activities with common patterns
type MockActivityFactory struct {
	mocks map[string]*MockActivity
	mu    sync.Mutex
}

// NewMockActivityFactory creates a new mock activity factory
func NewMockActivityFactory() *MockActivityFactory {
	return &MockActivityFactory{
		mocks: make(map[string]*MockActivity),
	}
}

// CreateDatabaseActivity creates a mock database activity
func (f *MockActivityFactory) CreateDatabaseActivity(name string) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	mock := &MockActivity{
		Name: name,
		ReturnValues: []interface{}{
			map[string]interface{}{
				"id":        "123",
				"status":    "success",
				"timestamp": time.Now(),
			},
		},
	}
	
	f.mocks[name] = mock
	return mock
}

// CreateHTTPActivity creates a mock HTTP activity
func (f *MockActivityFactory) CreateHTTPActivity(name string, statusCode int, response interface{}) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	mock := &MockActivity{
		Name: name,
		ReturnValues: []interface{}{
			map[string]interface{}{
				"statusCode": statusCode,
				"body":       response,
				"headers":    map[string]string{"Content-Type": "application/json"},
			},
		},
	}
	
	// Add network delay simulation
	mock.WithDelay(func(callIndex int) time.Duration {
		return time.Duration(50+callIndex*10) * time.Millisecond
	})
	
	f.mocks[name] = mock
	return mock
}

// CreateAsyncActivity creates a mock for async activities
func (f *MockActivityFactory) CreateAsyncActivity(name string, processingTime time.Duration) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	mock := &MockActivity{
		Name: name,
		ReturnValues: []interface{}{
			map[string]interface{}{
				"jobId":      fmt.Sprintf("job_%d", time.Now().Unix()),
				"status":     "completed",
				"processedAt": time.Now().Add(processingTime),
			},
		},
	}
	
	mock.WithDelay(func(callIndex int) time.Duration {
		return processingTime
	})
	
	f.mocks[name] = mock
	return mock
}

// CreateFailingActivity creates a mock that fails
func (f *MockActivityFactory) CreateFailingActivity(name string, err error, failAfter int) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	mock := &MockActivity{
		Name: name,
	}
	
	callCount := 0
	mock.WithSideEffect(func(args []interface{}) error {
		callCount++
		if callCount > failAfter {
			return err
		}
		return nil
	})
	
	f.mocks[name] = mock
	return mock
}

// CreateRetryableActivity creates a mock that succeeds after retries
func (f *MockActivityFactory) CreateRetryableActivity(name string, succeedAfter int) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	mock := &MockActivity{
		Name: name,
		ReturnValues: []interface{}{
			"success",
		},
	}
	
	callCount := 0
	mock.WithSideEffect(func(args []interface{}) error {
		callCount++
		if callCount < succeedAfter {
			return fmt.Errorf("temporary failure %d", callCount)
		}
		return nil
	})
	
	f.mocks[name] = mock
	return mock
}

// GetMock retrieves a mock by name
func (f *MockActivityFactory) GetMock(name string) *MockActivity {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	return f.mocks[name]
}

// ResetAll resets all mocks
func (f *MockActivityFactory) ResetAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for _, mock := range f.mocks {
		mock.Reset()
	}
}

// MockContext provides a mock activity context
type MockContext struct {
	context.Context
	heartbeats  []interface{}
	recordedVal interface{}
	workflowID  string
	runID       string
	activityID  string
	attempt     int32
}

// NewMockContext creates a new mock activity context
func NewMockContext() *MockContext {
	return &MockContext{
		Context:     context.Background(),
		workflowID:  "test-workflow-id",
		runID:       "test-run-id",
		activityID:  "test-activity",
		attempt:     1,
		heartbeats:  []interface{}{},
	}
}

// GetWorkflowID returns the workflow ID
func (m *MockContext) GetWorkflowID() string {
	return m.workflowID
}

// GetRunID returns the run ID
func (m *MockContext) GetRunID() string {
	return m.runID
}

// GetActivityID returns the activity ID
func (m *MockContext) GetActivityID() string {
	return m.activityID
}

// GetAttempt returns the attempt number
func (m *MockContext) GetAttempt() int32 {
	return m.attempt
}

// RecordHeartbeat records a heartbeat
func (m *MockContext) RecordHeartbeat(details ...interface{}) {
	m.heartbeats = append(m.heartbeats, details...)
}

// GetHeartbeatDetails gets heartbeat details
func (m *MockContext) GetHeartbeatDetails() []interface{} {
	return m.heartbeats
}

// HasHeartbeatDetails checks if heartbeat details exist
func (m *MockContext) HasHeartbeatDetails() bool {
	return len(m.heartbeats) > 0
}

// GetRecordedValue gets the recorded value
func (m *MockContext) GetRecordedValue() interface{} {
	return m.recordedVal
}

// SetRecordedValue sets the recorded value
func (m *MockContext) SetRecordedValue(val interface{}) {
	m.recordedVal = val
}

// MockRetryPolicy provides configurable retry behavior for testing
type MockRetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
	RetryFilter        func(err error) bool
}

// NewMockRetryPolicy creates a new mock retry policy
func NewMockRetryPolicy() *MockRetryPolicy {
	return &MockRetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    100 * time.Second,
		MaximumAttempts:    3,
	}
}

// WithRetryFilter adds a retry filter
func (p *MockRetryPolicy) WithRetryFilter(filter func(err error) bool) *MockRetryPolicy {
	p.RetryFilter = filter
	return p
}

// ShouldRetry determines if an error should be retried
func (p *MockRetryPolicy) ShouldRetry(err error, attempt int32) bool {
	if attempt >= p.MaximumAttempts {
		return false
	}
	
	if p.RetryFilter != nil {
		return p.RetryFilter(err)
	}
	
	return true
}

// GetNextRetryDelay calculates the next retry delay
func (p *MockRetryPolicy) GetNextRetryDelay(attempt int32) time.Duration {
	delay := p.InitialInterval
	
	for i := int32(1); i < attempt; i++ {
		delay = time.Duration(float64(delay) * p.BackoffCoefficient)
		if delay > p.MaximumInterval {
			delay = p.MaximumInterval
			break
		}
	}
	
	return delay
}