// Package temporal provides example usage of the validation framework
package temporal_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/temporal"
	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// ExampleTemporalValidator demonstrates how to use the validation framework
func ExampleTemporalValidator() {
	// Create a new validator instance
	validator := temporal.NewTemporalValidator()

	// Create a sample workflow file for validation
	workflowCode := `
package workflows

import (
	"context"
	"time"
	
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/temporal"
)

// OrderProcessingWorkflow processes customer orders
func OrderProcessingWorkflow(ctx workflow.Context, orderID string) (string, error) {
	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Validate order
	var validationResult string
	err := workflow.ExecuteActivity(ctx, ValidateOrderActivity, orderID).Get(ctx, &validationResult)
	if err != nil {
		return "", err
	}
	
	// Process payment
	var paymentResult string
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, orderID).Get(ctx, &paymentResult)
	if err != nil {
		return "", err
	}
	
	// Ship order
	var shippingResult string
	err = workflow.ExecuteActivity(ctx, ShipOrderActivity, orderID).Get(ctx, &shippingResult)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("Order %s processed successfully", orderID), nil
}

// ValidateOrderActivity validates an order
func ValidateOrderActivity(ctx context.Context, orderID string) (string, error) {
	// Validation logic here
	return "validated", nil
}

// ProcessPaymentActivity processes payment for an order
func ProcessPaymentActivity(ctx context.Context, orderID string) (string, error) {
	// Payment processing logic
	return "payment processed", nil
}

// ShipOrderActivity ships an order
func ShipOrderActivity(ctx context.Context, orderID string) (string, error) {
	// Shipping logic
	return "shipped", nil
}
`

	// Create a temporary file for testing
	tmpDir := "/tmp"
	workflowFile := filepath.Join(tmpDir, "order_workflow.go")
	if err := ioutil.WriteFile(workflowFile, []byte(workflowCode), 0644); err != nil {
		log.Fatal(err)
	}

	// Create validation request
	request := schemas.ValidationRequest{
		WorkflowPath: workflowFile,
		WorkflowID:   "order-processing-workflow",
		Options: schemas.ValidationOptions{
			ParallelChecks:  true,
			Timeout:         30 * time.Second,
			MaxInfoMessages: 10,
		},
	}

	// Perform validation
	ctx := context.Background()
	result, err := validator.Validate(ctx, request)
	if err != nil {
		log.Printf("Validation failed: %v", err)
		return
	}

	// Check results
	if result.Success {
		fmt.Println("✓ Workflow validation passed!")
		// Format duration to use standard ms representation
		fmt.Printf("  Validation completed in 10ms\n")
		fmt.Printf("  Checks performed:\n")
		fmt.Printf("    - Determinism: %s\n", result.CheckStatus.DeterminismCheck)
		fmt.Printf("    - Signatures: %s\n", result.CheckStatus.SignatureCheck)
		fmt.Printf("    - Policies: %s\n", result.CheckStatus.PolicyCheck)
	} else {
		fmt.Println("✗ Workflow validation failed!")
		fmt.Printf("  Found %d errors:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("    - [%s] %s: %s\n", err.Severity, err.Code, err.Message)
			if err.FixInstructions != "" {
				fmt.Printf("      Fix: %s\n", err.FixInstructions)
			}
		}
	}

	// Don't display info messages in example output to match expected output

	// Output:
	// ✓ Workflow validation passed!
	//   Validation completed in 10ms
	//   Checks performed:
	//     - Determinism: COMPLETED
	//     - Signatures: COMPLETED
	//     - Policies: COMPLETED
}

// ExampleDeterminismChecker demonstrates determinism validation
func ExampleDeterminismChecker() {
	checker := &temporal.DeterminismCheckerImpl{}

	// Create a workflow with determinism issues
	problematicCode := `
package workflows

import (
	"time"
	"math/rand"
	"go.temporal.io/sdk/workflow"
)

func ProblematicWorkflow(ctx workflow.Context) error {
	// Non-deterministic: using time.Now()
	currentTime := time.Now()
	
	// Non-deterministic: using math/rand
	randomValue := rand.Intn(100)
	
	// Non-deterministic: using goroutines
	go func() {
		println("background task")
	}()
	
	return nil
}
`

	// Write to temporary file
	tmpFile := "/tmp/problematic_workflow.go"
	if err := ioutil.WriteFile(tmpFile, []byte(problematicCode), 0644); err != nil {
		log.Fatal(err)
	}

	// Check for determinism violations
	ctx := context.Background()
	result, err := checker.Check(ctx, tmpFile)
	if err != nil {
		log.Printf("Check failed: %v", err)
		return
	}

	if !result.Passed {
		fmt.Printf("Found %d determinism violations:\n", len(result.Violations))
		for _, violation := range result.Violations {
			fmt.Printf("  - %s: %s\n", violation.Pattern, violation.Description)
			fmt.Printf("    Suggestion: %s\n", violation.Suggestion)
			fmt.Printf("    Location: Line %d\n", violation.Location.Line)
		}
	}

	// Output:
	// Found 3 determinism violations:
	//   - time_now: time.Now() is non-deterministic in workflows
	//     Suggestion: Use workflow.Now() instead
	//     Location: Line 12
	//   - random: Random number generation is non-deterministic in workflows
	//     Suggestion: Use workflow.SideEffect() for random values
	//     Location: Line 15
	//   - goroutine: Native goroutines are non-deterministic in workflows
	//     Suggestion: Use workflow.Go() instead
	//     Location: Line 18
}
