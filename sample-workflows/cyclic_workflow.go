package workflows

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
