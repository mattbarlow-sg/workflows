// Package testing demonstrates usage of the testing framework
package testing

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// sampleWorkflow demonstrates a simple workflow for testing
func sampleWorkflow(ctx context.Context, name string) (string, error) {
	// This would normally be a workflow.Context in real Temporal code
	return fmt.Sprintf("Hello, %s!", name), nil
}

// sampleActivity demonstrates a simple activity
func sampleActivity(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Processed: %s", input), nil
}

// TestBasicMocking demonstrates basic mocking capabilities
func TestBasicMocking(t *testing.T) {
	// Create a mock activity
	mock := &MockActivity{
		Name:         "TestActivity",
		ReturnValues: []interface{}{"mocked result"},
	}
	
	// Execute the mock
	result, err := mock.Execute(context.Background(), "test input")
	if err != nil {
		t.Fatal(err)
	}
	
	if result != "mocked result" {
		t.Errorf("Expected 'mocked result', got %v", result)
	}
	
	// Verify call count
	if mock.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mock.CallCount)
	}
}

// TestMockWithValidation demonstrates validation in mocks
func TestMockWithValidation(t *testing.T) {
	mock := &MockActivity{
		Name: "ValidatedActivity",
		ReturnValues: []interface{}{"success"},
	}
	
	// Add validation
	mock.WithValidation(func(args []interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("no arguments provided")
		}
		if args[0] == "invalid" {
			return fmt.Errorf("invalid input")
		}
		return nil
	})
	
	// Test valid input
	result, err := mock.Execute(context.Background(), "valid")
	if err != nil {
		t.Fatal(err)
	}
	if result != "success" {
		t.Errorf("Expected success, got %v", result)
	}
	
	// Test invalid input
	_, err = mock.Execute(context.Background(), "invalid")
	if err == nil {
		t.Error("Expected validation error")
	}
}

// TestRetryBehaviorMock demonstrates retry simulation
func TestRetryBehaviorMock(t *testing.T) {
	factory := NewMockActivityFactory()
	
	// Create a mock that succeeds after 3 attempts
	mock := factory.CreateRetryableActivity("RetryActivity", 3)
	
	// First two calls should fail
	_, err := mock.Execute(context.Background())
	if err == nil {
		t.Error("Expected first call to fail")
	}
	
	_, err = mock.Execute(context.Background())
	if err == nil {
		t.Error("Expected second call to fail")
	}
	
	// Third call should succeed
	result, err := mock.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected third call to succeed, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got %v", result)
	}
}

// TestMockActivityFactory demonstrates the factory pattern
func TestMockActivityFactory(t *testing.T) {
	factory := NewMockActivityFactory()
	
	// Create different types of mocks
	dbMock := factory.CreateDatabaseActivity("SaveToDatabase")
	httpMock := factory.CreateHTTPActivity("CallExternalAPI", 200, map[string]string{"status": "ok"})
	asyncMock := factory.CreateAsyncActivity("AsyncProcess", 100*time.Millisecond)
	
	// Test database mock
	result, err := dbMock.Execute(context.Background(), "data")
	if err != nil {
		t.Fatal(err)
	}
	dbResult := result.(map[string]interface{})
	if dbResult["status"] != "success" {
		t.Error("Expected success status")
	}
	
	// Test HTTP mock
	result, err = httpMock.Execute(context.Background(), "http://api.example.com")
	if err != nil {
		t.Fatal(err)
	}
	httpResult := result.(map[string]interface{})
	if httpResult["statusCode"] != 200 {
		t.Error("Expected status code 200")
	}
	
	// Test async mock
	start := time.Now()
	result, err = asyncMock.Execute(context.Background())
	duration := time.Since(start)
	
	if err != nil {
		t.Fatal(err)
	}
	if duration < 100*time.Millisecond {
		t.Error("Expected async mock to simulate delay")
	}
	
	// Reset all mocks
	factory.ResetAll()
	if dbMock.CallCount != 0 {
		t.Error("Expected mocks to be reset")
	}
}

// TestMockContext demonstrates mock context usage
func TestMockContext(t *testing.T) {
	ctx := NewMockContext()
	
	// Test basic properties
	if ctx.GetWorkflowID() != "test-workflow-id" {
		t.Error("Unexpected workflow ID")
	}
	
	if ctx.GetRunID() != "test-run-id" {
		t.Error("Unexpected run ID")
	}
	
	if ctx.GetActivityID() != "test-activity" {
		t.Error("Unexpected activity ID")
	}
	
	if ctx.GetAttempt() != 1 {
		t.Error("Expected attempt 1")
	}
	
	// Test heartbeat recording
	ctx.RecordHeartbeat("progress", 50)
	heartbeats := ctx.GetHeartbeatDetails()
	if len(heartbeats) != 2 {
		t.Error("Expected 2 heartbeat details")
	}
	
	// Test recorded value
	ctx.SetRecordedValue("test-value")
	if ctx.GetRecordedValue() != "test-value" {
		t.Error("Unexpected recorded value")
	}
}

// TestMockSignal demonstrates signal mocking
func TestMockSignal(t *testing.T) {
	signal := &MockSignal{
		Name: "TestSignal",
	}
	
	// Send signals
	signal.Send("data1")
	signal.Send("data2")
	
	// Verify received
	if len(signal.Received) != 2 {
		t.Errorf("Expected 2 signals, got %d", len(signal.Received))
	}
	
	// Test assertion helper
	signal.AssertReceived(mockTestingT{t}, "data1")
}

// TestMockQuery demonstrates query mocking
func TestMockQuery(t *testing.T) {
	query := &MockQuery{
		Name: "TestQuery",
	}
	
	// Set handler
	query.SetHandler(func() (interface{}, error) {
		return "query result", nil
	})
	
	// Handle query
	result, err := query.Handle()
	if err != nil {
		t.Fatal(err)
	}
	
	if result != "query result" {
		t.Errorf("Expected 'query result', got %v", result)
	}
	
	if query.CallCount != 1 {
		t.Error("Expected call count 1")
	}
}

// TestMockWorkflow demonstrates workflow mocking
func TestMockWorkflow(t *testing.T) {
	mock := &MockWorkflow{
		Name:         "TestWorkflow",
		ReturnValues: []interface{}{"workflow result"},
	}
	
	// Execute mock workflow (using nil for workflow.Context in this test)
	result, err := mock.Execute(nil, "input1", "input2")
	if err != nil {
		t.Fatal(err)
	}
	
	if result != "workflow result" {
		t.Errorf("Expected 'workflow result', got %v", result)
	}
	
	// Verify call was recorded
	if mock.CallCount != 1 {
		t.Error("Expected 1 call")
	}
	
	if len(mock.Calls) != 1 {
		t.Error("Expected 1 recorded call")
	}
	
	// Test with error
	mock.ReturnsError(fmt.Errorf("workflow error"))
	_, err = mock.Execute(nil, "input")
	if err == nil {
		t.Error("Expected error")
	}
}

// TestMockRetryPolicy demonstrates retry policy testing
func TestMockRetryPolicy(t *testing.T) {
	policy := NewMockRetryPolicy()
	
	// Test default behavior
	if !policy.ShouldRetry(fmt.Errorf("error"), 1) {
		t.Error("Expected retry for attempt 1")
	}
	
	if !policy.ShouldRetry(fmt.Errorf("error"), 2) {
		t.Error("Expected retry for attempt 2")
	}
	
	if policy.ShouldRetry(fmt.Errorf("error"), 3) {
		t.Error("Should not retry after max attempts")
	}
	
	// Test with custom filter
	policy.WithRetryFilter(func(err error) bool {
		// Only retry "retryable" errors
		return err.Error() == "retryable"
	})
	
	if !policy.ShouldRetry(fmt.Errorf("retryable"), 1) {
		t.Error("Expected retry for retryable error")
	}
	
	if policy.ShouldRetry(fmt.Errorf("non-retryable"), 1) {
		t.Error("Should not retry non-retryable error")
	}
	
	// Test backoff calculation
	delay := policy.GetNextRetryDelay(1)
	if delay != time.Second {
		t.Errorf("Expected 1s delay, got %v", delay)
	}
	
	delay = policy.GetNextRetryDelay(2)
	if delay != 2*time.Second {
		t.Errorf("Expected 2s delay, got %v", delay)
	}
	
	delay = policy.GetNextRetryDelay(3)
	if delay != 4*time.Second {
		t.Errorf("Expected 4s delay, got %v", delay)
	}
}

// mockTestingT implements TestingT for testing
type mockTestingT struct {
	*testing.T
}

func (m mockTestingT) Helper() {
	// No-op for testing
}