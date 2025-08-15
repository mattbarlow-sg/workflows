// Package main provides a CLI tool to test the Temporal validation framework
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/temporal"
	"github.com/mattbarlow-sg/workflows/src/schemas"
)

func main() {
	var (
		workflowPath = flag.String("path", "", "Path to workflow file or directory to validate")
		workflowID   = flag.String("id", "test-workflow", "Workflow ID")
		useCache     = flag.Bool("cache", false, "Enable caching")
		parallel     = flag.Bool("parallel", true, "Run checks in parallel")
		createSample = flag.Bool("sample", false, "Create sample workflow files for testing")
		verbose      = flag.Bool("v", false, "Verbose output")
	)

	flag.Parse()

	if *createSample {
		createSampleWorkflows()
		return
	}

	if *workflowPath == "" {
		fmt.Println("Usage: validate-temporal -path <workflow-file-or-directory>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Validate a single workflow file")
		fmt.Println("  validate-temporal -path ./workflow.go")
		fmt.Println("")
		fmt.Println("  # Create sample workflows for testing")
		fmt.Println("  validate-temporal -sample")
		fmt.Println("")
		fmt.Println("  # Validate the sample workflows")
		fmt.Println("  validate-temporal -path ./sample-workflows/valid_workflow.go")
		fmt.Println("  validate-temporal -path ./sample-workflows/invalid_workflow.go")
		os.Exit(1)
	}

	// Create validator
	validator := temporal.NewTemporalValidator()

	// Create validation request
	request := schemas.ValidationRequest{
		WorkflowPath: *workflowPath,
		WorkflowID:   *workflowID,
		Options: schemas.ValidationOptions{
			ParallelChecks:  *parallel,
			Timeout:         30 * time.Second,
			MaxInfoMessages: 20,
			SkipCache:       !*useCache,
		},
	}

	fmt.Printf("Validating: %s\n", *workflowPath)
	fmt.Printf("   Workflow ID: %s\n", *workflowID)
	fmt.Printf("   Parallel checks: %v\n", *parallel)
	fmt.Printf("   Cache enabled: %v\n", *useCache)
	fmt.Println()

	// Perform validation
	ctx := context.Background()
	startTime := time.Now()

	var result *schemas.ValidationResult
	var err error

	if *useCache {
		result, err = validator.ValidateWithCache(ctx, request)
	} else {
		result, err = validator.Validate(ctx, request)
	}

	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Println(strings.Repeat("=", 60))

	if result.Success {
		fmt.Println("VALIDATION PASSED")
	} else {
		fmt.Printf("VALIDATION FAILED (%d errors)\n", len(result.Errors))
	}

	fmt.Printf("Duration: %s", result.Duration)

	if result.CacheHit {
		fmt.Println(" (from cache)")
	} else {
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 60))

	// Display check status
	fmt.Println("\nCheck Status:")
	fmt.Printf("   * Determinism: %s\n", formatStatus(result.CheckStatus.DeterminismCheck))
	fmt.Printf("   * Signatures:  %s\n", formatStatus(result.CheckStatus.SignatureCheck))
	fmt.Printf("   * Policies:    %s\n", formatStatus(result.CheckStatus.PolicyCheck))

	// Display errors
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("\n%d. [%s] %s\n", i+1, err.Category, err.Code)
			fmt.Printf("   %s\n", err.Message)

			if err.Location.File != "" {
				fmt.Printf("   %s:%d", err.Location.File, err.Location.Line)
				if err.Location.Column > 0 {
					fmt.Printf(":%d", err.Location.Column)
				}
				fmt.Println()
				if err.Location.Snippet != "" {
					fmt.Printf("   %s\n", err.Location.Snippet)
				}
			}

			if err.FixInstructions != "" {
				fmt.Printf("   Fix: %s\n", err.FixInstructions)
			}

			if *verbose && err.Context != nil {
				fmt.Printf("   Context: %v\n", err.Context)
			}
		}
	}

	// Display info messages
	if len(result.InfoMessages) > 0 && *verbose {
		fmt.Printf("\nInfo Messages (%d):\n", len(result.InfoMessages))
		for _, msg := range result.InfoMessages {
			fmt.Printf("   * [%s] %s\n", msg.Category, msg.Message)
		}
	}

	fmt.Printf("\nTotal time: %s\n", time.Since(startTime))

	if !result.Success {
		os.Exit(1)
	}
}

func formatStatus(status schemas.CheckStatus) string {
	switch status {
	case schemas.CheckStatusCompleted:
		return "COMPLETED"
	case schemas.CheckStatusFailed:
		return "FAILED"
	case schemas.CheckStatusSkipped:
		return "SKIPPED"
	case schemas.CheckStatusCancelled:
		return "CANCELLED"
	case schemas.CheckStatusRunning:
		return "RUNNING"
	default:
		return "PENDING"
	}
}

func createSampleWorkflows() {
	dir := "./sample-workflows"

	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Valid workflow
	validWorkflow := `package workflows

import (
	"context"
	"fmt"
	"time"
	
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/temporal"
)

// OrderProcessingWorkflow is a valid Temporal workflow
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting order processing", "orderID", orderID)
	
	// Use workflow.Now() instead of time.Now()
	startTime := workflow.Now(ctx)
	
	// Configure activity options with proper timeout and retry
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Execute activities
	var validationResult string
	err := workflow.ExecuteActivity(ctx, ValidateOrderActivity, orderID).Get(ctx, &validationResult)
	if err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}
	
	var paymentResult string
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, orderID).Get(ctx, &paymentResult)
	if err != nil {
		return "", fmt.Errorf("payment processing failed: %w", err)
	}
	
	// Use workflow timer instead of time.Sleep
	err = workflow.Sleep(ctx, 2*time.Second)
	if err != nil {
		return "", err
	}
	
	var shippingResult string
	err = workflow.ExecuteActivity(ctx, ShipOrderActivity, orderID).Get(ctx, &shippingResult)
	if err != nil {
		return "", fmt.Errorf("shipping failed: %w", err)
	}
	
	endTime := workflow.Now(ctx)
	duration := endTime.Sub(startTime)
	
	return fmt.Sprintf("Order %s completed in %v", orderID, duration), nil
}

// ValidateOrderActivity validates an order (proper signature)
func ValidateOrderActivity(ctx context.Context, orderID string) (string, error) {
	// Simulate validation
	return fmt.Sprintf("Order %s validated", orderID), nil
}

// ProcessPaymentActivity processes payment (proper signature)
func ProcessPaymentActivity(ctx context.Context, orderID string) (string, error) {
	// Simulate payment processing
	return fmt.Sprintf("Payment processed for order %s", orderID), nil
}

// ShipOrderActivity ships the order (proper signature)
func ShipOrderActivity(ctx context.Context, orderID string) (string, error) {
	// Simulate shipping
	return fmt.Sprintf("Order %s shipped", orderID), nil
}
`

	// Invalid workflow with issues
	invalidWorkflow := `package workflows

import (
	"context"
	"math/rand"
	"os"
	"time"
	
	"go.temporal.io/sdk/workflow"
)

// BadWorkflow has multiple validation issues
func BadWorkflow(ctx workflow.Context, input string) (string, error) {
	// ISSUE: Using time.Now() instead of workflow.Now()
	currentTime := time.Now()
	println("Current time:", currentTime)
	
	// ISSUE: Using math/rand without workflow.SideEffect
	randomNum := rand.Intn(100)
	
	// ISSUE: Using native goroutine instead of workflow.Go
	go func() {
		println("Background task running")
	}()
	
	// ISSUE: Using native channel instead of workflow.Channel
	ch := make(chan string)
	
	// ISSUE: Using select instead of workflow.Selector
	select {
	case msg := <-ch:
		println(msg)
	default:
		println("No message")
	}
	
	// ISSUE: Reading environment variable (non-deterministic)
	apiKey := os.Getenv("API_KEY")
	
	// ISSUE: Map iteration without sorting (non-deterministic order)
	data := map[string]int{"a": 1, "b": 2, "c": 3}
	for key, value := range data {
		println(key, value)
	}
	
	// Activity with issues
	var result string
	err := workflow.ExecuteActivity(ctx, badActivity, input).Get(ctx, &result)
	if err != nil {
		return "", err
	}
	
	return result, nil
}

// ISSUE: Activity with bad naming (should be PascalCase)
func bad_activity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// ISSUE: Activity missing context parameter
func ProcessDataActivity(input string) (string, error) {
	return "processed", nil
}

// ISSUE: Activity missing error return
func NotifyUserActivity(ctx context.Context, userID string) string {
	return "notified"
}

// HumanApprovalWorkflow without proper escalation
func HumanApprovalWorkflow(ctx workflow.Context, requestID string) error {
	// ISSUE: Human task without escalation policy
	var approved bool
	signalChan := workflow.GetSignalChannel(ctx, "approval-signal")
	signalChan.Receive(ctx, &approved)
	
	if !approved {
		return fmt.Errorf("request %s was rejected", requestID)
	}
	
	return nil
}
`

	// Workflow with graph cycles
	cyclicWorkflow := `package workflows

import (
	"go.temporal.io/sdk/workflow"
)

// WorkflowA calls WorkflowB
func WorkflowA(ctx workflow.Context) error {
	err := workflow.ExecuteChildWorkflow(ctx, "WorkflowB", nil).Get(ctx, nil)
	return err
}

// WorkflowB calls WorkflowC
func WorkflowB(ctx workflow.Context) error {
	err := workflow.ExecuteChildWorkflow(ctx, "WorkflowC", nil).Get(ctx, nil)
	return err
}

// WorkflowC calls WorkflowA (creates a cycle!)
func WorkflowC(ctx workflow.Context) error {
	err := workflow.ExecuteChildWorkflow(ctx, "WorkflowA", nil).Get(ctx, nil)
	return err
}
`

	// Write files
	files := map[string]string{
		"valid_workflow.go":   validWorkflow,
		"invalid_workflow.go": invalidWorkflow,
		"cyclic_workflow.go":  cyclicWorkflow,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", filename, err)
		}
		fmt.Printf("Created: %s\n", path)
	}

	fmt.Println("\nSample workflows created in ./sample-workflows/")
	fmt.Println("\nTest them with:")
	fmt.Println("  go run cmd/validate-temporal/main.go -path ./sample-workflows/valid_workflow.go")
	fmt.Println("  go run cmd/validate-temporal/main.go -path ./sample-workflows/invalid_workflow.go -v")
	fmt.Println("  go run cmd/validate-temporal/main.go -path ./sample-workflows/cyclic_workflow.go")
}
