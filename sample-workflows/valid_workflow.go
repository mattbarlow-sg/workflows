package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
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
