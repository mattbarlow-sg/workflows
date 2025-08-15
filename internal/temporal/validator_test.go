// Package temporal provides tests for the Temporal validation framework
package temporal

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// TestTemporalValidator tests the main validator
func TestTemporalValidator(t *testing.T) {
	validator := NewTemporalValidator()

	t.Run("ValidateValidWorkflow", func(t *testing.T) {
		// Create a temporary valid workflow file
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "valid_workflow.go")

		validWorkflow := `
package workflows

import (
	"context"
	"time"
	
	"go.temporal.io/sdk/workflow"
)

func ValidWorkflow(ctx workflow.Context, input string) (string, error) {
	// Use workflow.Now() instead of time.Now()
	currentTime := workflow.Now(ctx)
	
	// Use workflow timer
	err := workflow.Sleep(ctx, 5*time.Second)
	if err != nil {
		return "", err
	}
	
	// Execute activity with proper signature
	var result string
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	err = workflow.ExecuteActivity(ctx, ProcessActivity, input).Get(ctx, &result)
	if err != nil {
		return "", err
	}
	
	return result, nil
}

func ProcessActivity(ctx context.Context, input string) (string, error) {
	// Activity implementation
	return "processed: " + input, nil
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(validWorkflow), 0644); err != nil {
			t.Fatal(err)
		}

		request := schemas.ValidationRequest{
			WorkflowPath: workflowFile,
			WorkflowID:   "test-workflow",
			Options: schemas.ValidationOptions{
				ParallelChecks: true,
				Timeout:        30 * time.Second,
			},
		}

		result, err := validator.Validate(context.Background(), request)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Check for no errors in valid workflow
		if len(result.Errors) > 0 {
			t.Errorf("Expected no validation errors, got %d errors", len(result.Errors))
			for _, err := range result.Errors {
				t.Logf("Error: %s - %s", err.Code, err.Message)
			}
		}
	})

	t.Run("ValidateInvalidWorkflow", func(t *testing.T) {
		// Create a temporary invalid workflow file
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "invalid_workflow.go")

		invalidWorkflow := `
package workflows

import (
	"time"
	"math/rand"
	
	"go.temporal.io/sdk/workflow"
)

func InvalidWorkflow(ctx workflow.Context, input string) (string, error) {
	// Using time.Now() - non-deterministic!
	currentTime := time.Now()
	
	// Using native goroutine - non-deterministic!
	go func() {
		println("background task")
	}()
	
	// Using math/rand - non-deterministic!
	randomNum := rand.Intn(100)
	
	// Using native select - non-deterministic!
	ch := make(chan string)
	select {
	case msg := <-ch:
		return msg, nil
	default:
		return "", nil
	}
}

// Bad activity naming - should be PascalCase
func process_activity(ctx context.Context, input string) (string, error) {
	return input, nil
}

// Missing context parameter
func BadActivity(input string) (string, error) {
	return input, nil
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(invalidWorkflow), 0644); err != nil {
			t.Fatal(err)
		}

		request := schemas.ValidationRequest{
			WorkflowPath: workflowFile,
			WorkflowID:   "test-invalid-workflow",
			Options: schemas.ValidationOptions{
				ParallelChecks: false, // Test sequential mode
			},
		}

		result, err := validator.Validate(context.Background(), request)
		if err != nil {
			t.Errorf("Validation should not error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Should have validation errors
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for invalid workflow")
		}

		// Check for specific errors
		hasTimeNowError := false
		hasGoroutineError := false
		hasRandomError := false

		for _, err := range result.Errors {
			if strings.Contains(err.Message, "time.Now()") {
				hasTimeNowError = true
			}
			if strings.Contains(err.Message, "goroutine") {
				hasGoroutineError = true
			}
			if strings.Contains(err.Message, "random") || strings.Contains(err.Message, "Random") {
				hasRandomError = true
			}
		}

		if !hasTimeNowError {
			t.Error("Expected time.Now() determinism error")
		}
		if !hasGoroutineError {
			t.Error("Expected goroutine determinism error")
		}
		if !hasRandomError {
			t.Error("Expected random number determinism error")
		}
	})
}

// TestDeterminismChecker tests the determinism checker
func TestDeterminismChecker(t *testing.T) {
	checker := &DeterminismCheckerImpl{}

	t.Run("DetectTimeNow", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "workflow.go")

		code := `
package workflows

import "time"

func MyWorkflow() {
	now := time.Now()
	println(now)
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := checker.Check(context.Background(), workflowFile)
		if err != nil {
			t.Fatal(err)
		}

		if result.Passed {
			t.Error("Expected determinism check to fail")
		}

		if len(result.Violations) == 0 {
			t.Error("Expected violations")
		}

		found := false
		for _, v := range result.Violations {
			if v.Pattern == "time_now" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected time_now violation")
		}
	})

	t.Run("DetectGoroutines", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "workflow.go")

		code := `
package workflows

func MyWorkflow() {
	go func() {
		println("async")
	}()
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := checker.Check(context.Background(), workflowFile)
		if err != nil {
			t.Fatal(err)
		}

		if result.Passed {
			t.Error("Expected determinism check to fail")
		}

		found := false
		for _, v := range result.Violations {
			if v.Pattern == "goroutine" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected goroutine violation")
		}
	})
}

// TestActivitySignatureValidator tests activity signature validation
func TestActivitySignatureValidator(t *testing.T) {
	validator := &ActivitySignatureValidatorImpl{}

	t.Run("ValidActivity", func(t *testing.T) {
		tmpDir := t.TempDir()
		activityFile := filepath.Join(tmpDir, "activities.go")

		code := `
package activities

import "context"

func ProcessOrderActivity(ctx context.Context, orderID string) (string, error) {
	return "processed: " + orderID, nil
}

func SendEmailActivity(ctx context.Context, to string, subject string) error {
	return nil
}
`
		if err := ioutil.WriteFile(activityFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := validator.Validate(context.Background(), activityFile)
		if err != nil {
			t.Fatal(err)
		}

		if !result.Passed {
			t.Error("Expected validation to pass")
		}

		if len(result.Violations) > 0 {
			t.Errorf("Expected no violations, got %d", len(result.Violations))
		}
	})

	t.Run("InvalidActivityNaming", func(t *testing.T) {
		tmpDir := t.TempDir()
		activityFile := filepath.Join(tmpDir, "activities.go")

		code := `
package activities

import "context"

// Wrong naming - snake_case
func process_order_activity(ctx context.Context, orderID string) (string, error) {
	return "processed", nil
}

// Wrong naming - starts with lowercase
func sendEmailActivity(ctx context.Context) error {
	return nil
}
`
		if err := ioutil.WriteFile(activityFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := validator.Validate(context.Background(), activityFile)
		if err != nil {
			t.Fatal(err)
		}

		if result.Passed {
			t.Error("Expected validation to fail")
		}

		if len(result.Violations) == 0 {
			t.Error("Expected naming violations")
		}
	})

	t.Run("MissingContext", func(t *testing.T) {
		tmpDir := t.TempDir()
		activityFile := filepath.Join(tmpDir, "activities.go")

		code := `
package activities

// Missing context.Context parameter
func ProcessActivity(orderID string) (string, error) {
	return "processed", nil
}
`
		if err := ioutil.WriteFile(activityFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := validator.Validate(context.Background(), activityFile)
		if err != nil {
			t.Fatal(err)
		}

		if result.Passed {
			t.Error("Expected validation to fail")
		}

		found := false
		for _, v := range result.Violations {
			if v.ViolationType == "missing_context" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected missing_context violation")
		}
	})
}

// TestWorkflowGraphValidator tests the workflow graph validator
func TestWorkflowGraphValidator(t *testing.T) {
	validator := &WorkflowGraphValidatorImpl{}

	t.Run("DetectCycle", func(t *testing.T) {
		// Manually build a graph with cycle
		validator.graph = map[string][]string{
			"WorkflowA": {"WorkflowB"},
			"WorkflowB": {"WorkflowC"},
			"WorkflowC": {"WorkflowA"}, // Cycle!
		}

		cycles, err := validator.DetectCycles("")
		if err != nil {
			t.Fatal(err)
		}

		if len(cycles) == 0 {
			t.Error("Expected to detect cycle")
		}
	})

	t.Run("NoCycle", func(t *testing.T) {
		// Build a DAG (no cycles)
		validator.graph = map[string][]string{
			"WorkflowA": {"WorkflowB", "WorkflowC"},
			"WorkflowB": {"WorkflowD"},
			"WorkflowC": {"WorkflowD"},
			"WorkflowD": {},
		}

		cycles, err := validator.DetectCycles("")
		if err != nil {
			t.Fatal(err)
		}

		if len(cycles) > 0 {
			t.Errorf("Expected no cycles, got %v", cycles)
		}
	})

	t.Run("CheckConnectivity", func(t *testing.T) {
		// Connected graph
		validator.graph = map[string][]string{
			"WorkflowA": {"WorkflowB"},
			"WorkflowB": {"WorkflowC"},
			"WorkflowC": {},
		}

		connected, err := validator.CheckConnectivity("")
		if err != nil {
			t.Fatal(err)
		}

		if !connected {
			t.Error("Expected graph to be connected")
		}

		// Disconnected graph
		validator.graph = map[string][]string{
			"WorkflowA": {"WorkflowB"},
			"WorkflowC": {"WorkflowD"}, // Separate component
		}

		connected, err = validator.CheckConnectivity("")
		if err != nil {
			t.Fatal(err)
		}

		if connected {
			t.Error("Expected graph to be disconnected")
		}
	})
}

// TestPolicyValidator tests policy validation
func TestPolicyValidator(t *testing.T) {
	validator := &PolicyValidatorImpl{}

	t.Run("ValidPolicies", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "workflow.go")

		code := `
package workflows

import (
	"time"
	"go.temporal.io/sdk/workflow"
)

func MyWorkflow(ctx workflow.Context) error {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	return nil
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := validator.Validate(context.Background(), workflowFile)
		if err != nil {
			t.Fatal(err)
		}

		if !result.Passed {
			t.Error("Expected validation to pass")
		}
	})

	t.Run("InsufficientRetries", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "workflow.go")

		code := `
package workflows

import (
	"time"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/temporal"
)

func MyWorkflow(ctx workflow.Context) error {
	// Activity with insufficient retry attempts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	return nil
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := validator.Validate(context.Background(), workflowFile)
		if err != nil {
			t.Fatal(err)
		}

		// Should detect insufficient retry attempts
		if len(result.RetryViolations) == 0 {
			t.Error("Expected retry violations")
		}
	})
}

// TestHumanTaskValidator tests human task validation
func TestHumanTaskValidator(t *testing.T) {
	validator := &HumanTaskValidatorImpl{}

	t.Run("ValidateEscalation", func(t *testing.T) {
		validConfig := map[string]interface{}{
			"escalation_after": 1 * time.Hour,
			"escalate_to":      "manager",
			"max_escalations":  3,
		}

		err := validator.ValidateEscalation(validConfig)
		if err != nil {
			t.Errorf("Expected valid escalation config, got error: %v", err)
		}

		// Test invalid config
		invalidConfig := map[string]interface{}{
			"escalation_after": 10 * time.Minute, // Too short
			"escalate_to":      "manager",
			"max_escalations":  10, // Too many
		}

		err = validator.ValidateEscalation(invalidConfig)
		if err == nil {
			t.Error("Expected error for invalid escalation config")
		}
	})

	t.Run("ValidateAssignment", func(t *testing.T) {
		validConfig := map[string]interface{}{
			"assignment_strategy": "round_robin",
			"assignee_pool":       []string{"user1", "user2"},
		}

		err := validator.ValidateAssignment(validConfig)
		if err != nil {
			t.Errorf("Expected valid assignment config, got error: %v", err)
		}

		// Test invalid strategy
		invalidConfig := map[string]interface{}{
			"assignment_strategy": "invalid_strategy",
			"assignee_pool":       []string{"user1"},
		}

		err = validator.ValidateAssignment(invalidConfig)
		if err == nil {
			t.Error("Expected error for invalid assignment strategy")
		}
	})

	t.Run("HumanTaskWithoutEscalation", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowFile := filepath.Join(tmpDir, "workflow.go")

		code := `
package workflows

func HumanTaskWorkflow() {
	// Human task without escalation policy
	task := HumanTask{
		Name: "ApprovalTask",
	}
}
`
		if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		err := validator.Validate(context.Background(), workflowFile)
		if err == nil {
			t.Error("Expected error for human task without escalation")
		}
	})
}

// TestMemoryCache tests the cache implementation
func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()

	t.Run("SetAndGet", func(t *testing.T) {
		entry := &schemas.CacheEntry{
			Key:        "test-key",
			WorkflowID: "test-workflow",
			SourceHash: "abc123",
			Result: schemas.ValidationResult{
				Success: true,
			},
			CreatedAt:   time.Now(),
			AccessedAt:  time.Now(),
			AccessCount: 0,
		}

		err := cache.Set(entry)
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := cache.Get("test-key")
		if err != nil {
			t.Fatal(err)
		}

		if retrieved.Key != "test-key" {
			t.Errorf("Expected key 'test-key', got '%s'", retrieved.Key)
		}

		if retrieved.AccessCount != 1 {
			t.Errorf("Expected access count 1, got %d", retrieved.AccessCount)
		}
	})

	t.Run("CacheMiss", func(t *testing.T) {
		_, err := cache.Get("non-existent")
		if err == nil {
			t.Error("Expected error for cache miss")
		}
	})

	t.Run("Invalidate", func(t *testing.T) {
		entry := &schemas.CacheEntry{
			Key:        "test-key-2",
			WorkflowID: "test-workflow",
			SourceHash: "def456",
			Result:     schemas.ValidationResult{},
		}

		cache.Set(entry)

		err := cache.Invalidate("test-key-2")
		if err != nil {
			t.Fatal(err)
		}

		_, err = cache.Get("test-key-2")
		if err == nil {
			t.Error("Expected error after invalidation")
		}
	})

	t.Run("InvalidateByWorkflow", func(t *testing.T) {
		// Add multiple entries for same workflow
		cache.Set(&schemas.CacheEntry{
			Key:        "workflow1:hash1",
			WorkflowID: "workflow1",
		})
		cache.Set(&schemas.CacheEntry{
			Key:        "workflow1:hash2",
			WorkflowID: "workflow1",
		})
		cache.Set(&schemas.CacheEntry{
			Key:        "workflow2:hash1",
			WorkflowID: "workflow2",
		})

		err := cache.InvalidateByWorkflow("workflow1")
		if err != nil {
			t.Fatal(err)
		}

		// workflow1 entries should be gone
		_, err1 := cache.Get("workflow1:hash1")
		_, err2 := cache.Get("workflow1:hash2")
		if err1 == nil || err2 == nil {
			t.Error("Expected workflow1 entries to be invalidated")
		}

		// workflow2 entry should still exist
		_, err3 := cache.Get("workflow2:hash1")
		if err3 != nil {
			t.Error("workflow2 entry should not be invalidated")
		}
	})
}

// Benchmark tests
func BenchmarkValidator(b *testing.B) {
	validator := NewTemporalValidator()

	// Create test workflow
	tmpDir := b.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.go")

	code := `
package workflows

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

func BenchWorkflow(ctx workflow.Context) error {
	return nil
}

func BenchActivity(ctx context.Context) error {
	return nil
}
`
	if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	request := schemas.ValidationRequest{
		WorkflowPath: workflowFile,
		WorkflowID:   "bench-workflow",
		Options: schemas.ValidationOptions{
			ParallelChecks: true,
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(context.Background(), request)
	}
}

func BenchmarkDeterminismChecker(b *testing.B) {
	checker := &DeterminismCheckerImpl{}

	tmpDir := b.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.go")

	// Large workflow file for benchmarking
	code := strings.Repeat(`
func Workflow() {
	x := 1
	y := 2
	z := x + y
}
`, 100)

	if err := ioutil.WriteFile(workflowFile, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = checker.Check(context.Background(), workflowFile)
	}
}
