// Package temporal provides Temporal workflow infrastructure
package temporal_test

import (
	"context"
	"fmt"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/temporal"
)

func ExampleGenerator_GenerateWorkflow() {
	// Create a new generator
	gen, err := temporal.NewGenerator(temporal.GeneratorConfig{
		TemplateDir: "internal/temporal/templates",
		OutputDir:   "pkg/workflows/generated",
	})
	if err != nil {
		panic(err)
	}

	// Define a workflow specification
	spec := temporal.WorkflowSpec{
		Name:        "OrderProcessing",
		Package:     "orders",
		Description: "Process customer orders with approval",
		InputType:   "OrderID string\nCustomerID string\nAmount float64",
		OutputType:  "OrderStatus string\nProcessedAt time.Time",
		Template:    "approval",
		Activities: []temporal.ActivitySpec{
			{
				Name:        "ValidateOrder",
				Description: "Validate order details",
				InputType:   "OrderInput",
				OutputType:  "ValidationResult",
				Timeout:     5 * time.Minute,
			},
			{
				Name:        "ProcessPayment",
				Description: "Process payment for the order",
				InputType:   "PaymentRequest",
				OutputType:  "PaymentResult",
				Timeout:     10 * time.Minute,
				RetryPolicy: temporal.RetryPolicy{
					InitialInterval:    1 * time.Second,
					BackoffCoefficient: 2.0,
					MaximumInterval:    1 * time.Minute,
					MaximumAttempts:    3,
				},
			},
		},
		HumanTasks: []temporal.HumanTaskSpec{
			{
				Name:           "ManagerApproval",
				Description:    "Manager approval for high-value orders",
				AssignedTo:     "manager@example.com",
				EscalationTime: 2 * time.Hour,
				EscalationTo:   "director@example.com",
				Priority:       "high",
				Deadline:       4 * time.Hour,
			},
		},
		Signals: []temporal.SignalSpec{
			{
				Name:        "CancelOrder",
				Description: "Cancel the order processing",
				PayloadType: "CancelReason",
			},
		},
		Queries: []temporal.QuerySpec{
			{
				Name:         "GetOrderStatus",
				Description:  "Get current order status",
				ResponseType: "OrderStatus",
			},
		},
		Options: temporal.WorkflowOptions{
			TaskQueue:                "order-processing",
			WorkflowExecutionTimeout: 24 * time.Hour,
			WorkflowRunTimeout:       12 * time.Hour,
			WorkflowTaskTimeout:      1 * time.Minute,
		},
	}

	// Generate the workflow code
	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	if err != nil {
		panic(err)
	}

	// Save the generated code
	if err := gen.SaveGeneratedCode(code); err != nil {
		panic(err)
	}

	fmt.Println("Workflow generated successfully")
	// Output: Workflow generated successfully
}

func ExampleWorkflowBuilder() {
	// Use the fluent builder API for simpler workflows
	builder := temporal.NewWorkflowBuilder("DataProcessing", "data").
		WithDescription("Process data files in batches").
		WithTemplate("scheduled").
		WithInput("BatchSize int\nSourcePath string").
		WithOutput("ProcessedCount int\nErrors []string").
		WithTaskQueue("data-processing").
		WithExecutionTimeout(6 * time.Hour).
		AddActivity("ReadBatch", "Read batch of data").
			WithInput("BatchRequest").
			WithOutput("BatchData").
			WithTimeout(30 * time.Minute).
			Done().
		AddActivity("ProcessBatch", "Process the batch").
			WithInput("BatchData").
			WithOutput("ProcessedData").
			WithTimeout(1 * time.Hour).
			WithRetryPolicy(5*time.Second, 2.0, 5*time.Minute, 5).
			Done().
		AddActivity("WriteBatch", "Write processed data").
			WithInput("ProcessedData").
			WithOutput("WriteResult").
			WithTimeout(30 * time.Minute).
			Done().
		AddSignal("PauseBatch", "Pause batch processing", "PauseRequest").
		AddSignal("ResumeBatch", "Resume batch processing", "ResumeRequest").
		AddQuery("GetProgress", "Get processing progress", "ProgressInfo")

	// Build the specification
	spec, err := builder.Build()
	if err != nil {
		panic(err)
	}

	// Generate the workflow
	gen, _ := temporal.NewGenerator(temporal.GeneratorConfig{})
	ctx := context.Background()
	code, err := builder.Generate(ctx, gen)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated workflow: %s\n", code.WorkflowName)
	// Output: Generated workflow: DataProcessing
}

func ExampleGenerateApprovalWorkflow() {
	// Use pre-configured generator for common patterns
	approvers := []string{
		"manager@example.com",
		"director@example.com",
		"cfo@example.com",
	}

	builder := temporal.GenerateApprovalWorkflow(
		"ExpenseApproval",
		"expenses",
		approvers,
	)

	// Add additional customization if needed
	builder.WithTaskQueue("expense-approvals").
		AddActivity("ValidateExpense", "Validate expense report").
			WithInput("ExpenseReport").
			WithOutput("ValidationResult").
			WithTimeout(10 * time.Minute).
			Done()

	// Generate the workflow
	gen, _ := temporal.NewGenerator(temporal.GeneratorConfig{})
	ctx := context.Background()
	code, err := builder.Generate(ctx, gen)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated approval workflow with %d approvers\n", len(approvers))
	// Output: Generated approval workflow with 3 approvers
}

func ExampleGenerateLongRunningWorkflow() {
	// Generate a long-running workflow with checkpoints
	builder := temporal.GenerateLongRunningWorkflow(
		"DataMigration",
		"migration",
		72*time.Hour, // 3 days max duration
	)

	// Generate the workflow
	gen, _ := temporal.NewGenerator(temporal.GeneratorConfig{})
	ctx := context.Background()
	code, err := builder.Generate(ctx, gen)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated long-running workflow: %s\n", code.WorkflowName)
	// Output: Generated long-running workflow: DataMigration
}

func ExampleGenerateETLWorkflow() {
	// Generate an ETL workflow
	builder := temporal.GenerateETLWorkflow("CustomerDataETL", "etl")

	// Customize the ETL workflow
	builder.WithTaskQueue("etl-jobs").
		WithExecutionTimeout(8 * time.Hour)

	// Generate the workflow
	gen, _ := temporal.NewGenerator(temporal.GeneratorConfig{})
	ctx := context.Background()
	code, err := builder.Generate(ctx, gen)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated ETL workflow: %s\n", code.WorkflowName)
	// Output: Generated ETL workflow: CustomerDataETL
}