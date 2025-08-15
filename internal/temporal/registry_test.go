package temporal

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Sample workflow and activity for testing
func sampleWorkflow(ctx workflow.Context, input string) (string, error) {
	return "result: " + input, nil
}

func sampleActivity(ctx context.Context, input string) (string, error) {
	return "activity: " + input, nil
}

func TestRegistryCreation(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.workflows)
	assert.NotNil(t, registry.activities)
	assert.Empty(t, registry.workflows)
	assert.Empty(t, registry.activities)
}

func TestRegisterWorkflow(t *testing.T) {
	registry := NewRegistry()

	// Register workflow with default options
	err := registry.RegisterWorkflow(sampleWorkflow)
	assert.NoError(t, err)

	// Check registration
	wf, exists := registry.GetWorkflow(getWorkflowName(sampleWorkflow))
	assert.True(t, exists)
	assert.NotNil(t, wf.Function)
	assert.Contains(t, wf.TaskQueues, "default")

	// Try to register same workflow again
	err = registry.RegisterWorkflow(sampleWorkflow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterWorkflowWithOptions(t *testing.T) {
	registry := NewRegistry()

	// Register workflow with custom options
	err := registry.RegisterWorkflow(
		sampleWorkflow,
		WithWorkflowName("CustomWorkflow"),
		WithWorkflowTaskQueues("queue1", "queue2"),
		WithWorkflowDescription("Test workflow"),
		WithWorkflowOptions(workflow.RegisterOptions{
			Name: "CustomWorkflow",
		}),
	)
	assert.NoError(t, err)

	// Check registration
	wf, exists := registry.GetWorkflow("CustomWorkflow")
	assert.True(t, exists)
	assert.Equal(t, "CustomWorkflow", wf.Name)
	assert.Equal(t, []string{"queue1", "queue2"}, wf.TaskQueues)
	assert.Equal(t, "Test workflow", wf.Description)
	assert.Equal(t, "CustomWorkflow", wf.Options.Name)
}

func TestRegisterActivity(t *testing.T) {
	registry := NewRegistry()

	// Register activity with default options
	err := registry.RegisterActivity(sampleActivity)
	assert.NoError(t, err)

	// Check registration
	act, exists := registry.GetActivity(getActivityName(sampleActivity))
	assert.True(t, exists)
	assert.NotNil(t, act.Function)
	assert.Contains(t, act.TaskQueues, "default")

	// Try to register same activity again
	err = registry.RegisterActivity(sampleActivity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterActivityWithOptions(t *testing.T) {
	registry := NewRegistry()

	// Register activity with custom options
	err := registry.RegisterActivity(
		sampleActivity,
		WithActivityName("CustomActivity"),
		WithActivityTaskQueues("queue1", "queue2"),
		WithActivityDescription("Test activity"),
		WithActivityOptions(activity.RegisterOptions{
			Name: "CustomActivity",
		}),
	)
	assert.NoError(t, err)

	// Check registration
	act, exists := registry.GetActivity("CustomActivity")
	assert.True(t, exists)
	assert.Equal(t, "CustomActivity", act.Name)
	assert.Equal(t, []string{"queue1", "queue2"}, act.TaskQueues)
	assert.Equal(t, "Test activity", act.Description)
	assert.Equal(t, "CustomActivity", act.Options.Name)
}

func TestListWorkflows(t *testing.T) {
	registry := NewRegistry()

	// Register multiple workflows
	workflow1 := func(ctx workflow.Context) error { return nil }
	workflow2 := func(ctx workflow.Context) error { return nil }

	registry.RegisterWorkflow(workflow1, WithWorkflowName("workflow1"))
	registry.RegisterWorkflow(workflow2, WithWorkflowName("workflow2"))

	// List workflows
	workflows := registry.ListWorkflows()
	assert.Len(t, workflows, 2)

	names := make([]string, 0)
	for _, wf := range workflows {
		names = append(names, wf.Name)
	}
	assert.Contains(t, names, "workflow1")
	assert.Contains(t, names, "workflow2")
}

func TestListActivities(t *testing.T) {
	registry := NewRegistry()

	// Register multiple activities
	activity1 := func(ctx context.Context) error { return nil }
	activity2 := func(ctx context.Context) error { return nil }

	registry.RegisterActivity(activity1, WithActivityName("activity1"))
	registry.RegisterActivity(activity2, WithActivityName("activity2"))

	// List activities
	activities := registry.ListActivities()
	assert.Len(t, activities, 2)

	names := make([]string, 0)
	for _, act := range activities {
		names = append(names, act.Name)
	}
	assert.Contains(t, names, "activity1")
	assert.Contains(t, names, "activity2")
}

func TestGetWorkflowsForTaskQueue(t *testing.T) {
	registry := NewRegistry()

	// Register workflows for different task queues
	workflow1 := func(ctx workflow.Context) error { return nil }
	workflow2 := func(ctx workflow.Context) error { return nil }
	workflow3 := func(ctx workflow.Context) error { return nil }

	registry.RegisterWorkflow(workflow1, WithWorkflowName("wf1"), WithWorkflowTaskQueues("queue1"))
	registry.RegisterWorkflow(workflow2, WithWorkflowName("wf2"), WithWorkflowTaskQueues("queue1", "queue2"))
	registry.RegisterWorkflow(workflow3, WithWorkflowName("wf3"), WithWorkflowTaskQueues("queue2"))

	// Get workflows for queue1
	wfs := registry.GetWorkflowsForTaskQueue("queue1")
	assert.Len(t, wfs, 2)
	assert.Contains(t, wfs, "wf1")
	assert.Contains(t, wfs, "wf2")

	// Get workflows for queue2
	wfs = registry.GetWorkflowsForTaskQueue("queue2")
	assert.Len(t, wfs, 2)
	assert.Contains(t, wfs, "wf2")
	assert.Contains(t, wfs, "wf3")

	// Get workflows for non-existent queue
	wfs = registry.GetWorkflowsForTaskQueue("queue3")
	assert.Empty(t, wfs)
}

func TestGetActivitiesForTaskQueue(t *testing.T) {
	registry := NewRegistry()

	// Register activities for different task queues
	activity1 := func(ctx context.Context) error { return nil }
	activity2 := func(ctx context.Context) error { return nil }
	activity3 := func(ctx context.Context) error { return nil }

	registry.RegisterActivity(activity1, WithActivityName("act1"), WithActivityTaskQueues("queue1"))
	registry.RegisterActivity(activity2, WithActivityName("act2"), WithActivityTaskQueues("queue1", "queue2"))
	registry.RegisterActivity(activity3, WithActivityName("act3"), WithActivityTaskQueues("queue2"))

	// Get activities for queue1
	acts := registry.GetActivitiesForTaskQueue("queue1")
	assert.Len(t, acts, 2)
	assert.Contains(t, acts, "act1")
	assert.Contains(t, acts, "act2")

	// Get activities for queue2
	acts = registry.GetActivitiesForTaskQueue("queue2")
	assert.Len(t, acts, 2)
	assert.Contains(t, acts, "act2")
	assert.Contains(t, acts, "act3")
}

func TestUnregisterWorkflow(t *testing.T) {
	registry := NewRegistry()

	// Register and then unregister
	workflow1 := func(ctx workflow.Context) error { return nil }
	registry.RegisterWorkflow(workflow1, WithWorkflowName("workflow1"))

	err := registry.UnregisterWorkflow("workflow1")
	assert.NoError(t, err)

	// Check it's gone
	_, exists := registry.GetWorkflow("workflow1")
	assert.False(t, exists)

	// Try to unregister non-existent workflow
	err = registry.UnregisterWorkflow("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUnregisterActivity(t *testing.T) {
	registry := NewRegistry()

	// Register and then unregister
	activity1 := func(ctx context.Context) error { return nil }
	registry.RegisterActivity(activity1, WithActivityName("activity1"))

	err := registry.UnregisterActivity("activity1")
	assert.NoError(t, err)

	// Check it's gone
	_, exists := registry.GetActivity("activity1")
	assert.False(t, exists)

	// Try to unregister non-existent activity
	err = registry.UnregisterActivity("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClearRegistry(t *testing.T) {
	registry := NewRegistry()

	// Register some items
	workflow1 := func(ctx workflow.Context) error { return nil }
	activity1 := func(ctx context.Context) error { return nil }

	registry.RegisterWorkflow(workflow1, WithWorkflowName("workflow1"))
	registry.RegisterActivity(activity1, WithActivityName("activity1"))

	// Clear registry
	registry.Clear()

	// Check everything is gone
	assert.Empty(t, registry.ListWorkflows())
	assert.Empty(t, registry.ListActivities())
}

func TestRegistryBuilder(t *testing.T) {
	builder := NewRegistryBuilder()

	workflow1 := func(ctx workflow.Context) error { return nil }
	workflow2 := func(ctx workflow.Context) error { return nil }
	activity1 := func(ctx context.Context) error { return nil }

	// Build registry with multiple items
	builder.
		AddWorkflow(workflow1, WithWorkflowName("wf1")).
		AddWorkflow(workflow2, WithWorkflowName("wf2")).
		AddActivity(activity1, WithActivityName("act1"))

	registry, err := builder.Build()
	assert.NoError(t, err)
	assert.NotNil(t, registry)

	// Check items were registered
	_, exists := registry.GetWorkflow("wf1")
	assert.True(t, exists)
	_, exists = registry.GetWorkflow("wf2")
	assert.True(t, exists)
	_, exists = registry.GetActivity("act1")
	assert.True(t, exists)
}

func TestRegistryBuilderWithErrors(t *testing.T) {
	builder := NewRegistryBuilder()

	workflow1 := func(ctx workflow.Context) error { return nil }

	// Try to register same workflow twice
	builder.
		AddWorkflow(workflow1, WithWorkflowName("wf1")).
		AddWorkflow(workflow1, WithWorkflowName("wf1")) // Duplicate

	registry, err := builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registry build errors")
	assert.Nil(t, registry)
}

func TestRegistrySnapshot(t *testing.T) {
	registry := NewRegistry()

	// Register items
	workflow1 := func(ctx workflow.Context, input string) (string, error) { return input, nil }
	activity1 := func(ctx context.Context, input int) (int, error) { return input, nil }

	registry.RegisterWorkflow(workflow1, 
		WithWorkflowName("wf1"),
		WithWorkflowTaskQueues("queue1"),
		WithWorkflowDescription("Test workflow"))
	registry.RegisterActivity(activity1,
		WithActivityName("act1"),
		WithActivityTaskQueues("queue2"),
		WithActivityDescription("Test activity"))

	// Get snapshot
	snapshot := registry.GetSnapshot()
	assert.NotNil(t, snapshot)
	assert.Len(t, snapshot.Workflows, 1)
	assert.Len(t, snapshot.Activities, 1)

	// Check workflow info
	wfInfo := snapshot.Workflows[0]
	assert.Equal(t, "wf1", wfInfo.Name)
	assert.Equal(t, []string{"queue1"}, wfInfo.TaskQueues)
	assert.Equal(t, "Test workflow", wfInfo.Description)
	assert.NotEmpty(t, wfInfo.InputType)
	assert.NotEmpty(t, wfInfo.OutputType)

	// Check activity info
	actInfo := snapshot.Activities[0]
	assert.Equal(t, "act1", actInfo.Name)
	assert.Equal(t, []string{"queue2"}, actInfo.TaskQueues)
	assert.Equal(t, "Test activity", actInfo.Description)
	assert.NotEmpty(t, actInfo.InputType)
	assert.NotEmpty(t, actInfo.OutputType)
}

func TestDynamicRegistry(t *testing.T) {
	changeCount := 0
	var lastSnapshot RegistrySnapshot

	// Create dynamic registry with onChange callback
	dynamicRegistry := NewDynamicRegistry(func(snapshot RegistrySnapshot) {
		changeCount++
		lastSnapshot = snapshot
	})

	workflow1 := func(ctx workflow.Context) error { return nil }
	activity1 := func(ctx context.Context) error { return nil }

	// Register workflow - should trigger onChange
	err := dynamicRegistry.RegisterWorkflow(workflow1, WithWorkflowName("wf1"))
	assert.NoError(t, err)
	assert.Equal(t, 1, changeCount)
	assert.Len(t, lastSnapshot.Workflows, 1)

	// Register activity - should trigger onChange
	err = dynamicRegistry.RegisterActivity(activity1, WithActivityName("act1"))
	assert.NoError(t, err)
	assert.Equal(t, 2, changeCount)
	assert.Len(t, lastSnapshot.Activities, 1)

	// Unregister workflow - should trigger onChange
	err = dynamicRegistry.UnregisterWorkflow("wf1")
	assert.NoError(t, err)
	assert.Equal(t, 3, changeCount)
	assert.Len(t, lastSnapshot.Workflows, 0)

	// Unregister activity - should trigger onChange
	err = dynamicRegistry.UnregisterActivity("act1")
	assert.NoError(t, err)
	assert.Equal(t, 4, changeCount)
	assert.Len(t, lastSnapshot.Activities, 0)
}

func TestTypeValidation(t *testing.T) {
	registry := NewRegistry()

	// Test non-function registration
	notAFunction := "not a function"
	err := registry.RegisterWorkflow(notAFunction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a function")

	err = registry.RegisterActivity(notAFunction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a function")
}

func TestTypeExtraction(t *testing.T) {
	registry := NewRegistry()

	// Workflow with input and output types
	workflowWithTypes := func(ctx workflow.Context, input string) (int, error) {
		return len(input), nil
	}

	err := registry.RegisterWorkflow(workflowWithTypes, WithWorkflowName("typed_workflow"))
	require.NoError(t, err)

	wf, exists := registry.GetWorkflow("typed_workflow")
	require.True(t, exists)

	// Check extracted types
	assert.NotNil(t, wf.InputType)
	assert.Equal(t, reflect.TypeOf(""), wf.InputType)
	assert.NotNil(t, wf.OutputType)
	assert.Equal(t, reflect.TypeOf(0), wf.OutputType)

	// Activity with input and output types
	activityWithTypes := func(ctx context.Context, input int) (string, error) {
		return fmt.Sprintf("%d", input), nil
	}

	err = registry.RegisterActivity(activityWithTypes, WithActivityName("typed_activity"))
	require.NoError(t, err)

	act, exists := registry.GetActivity("typed_activity")
	require.True(t, exists)

	// Check extracted types
	assert.NotNil(t, act.InputType)
	assert.Equal(t, reflect.TypeOf(0), act.InputType)
	assert.NotNil(t, act.OutputType)
	assert.Equal(t, reflect.TypeOf(""), act.OutputType)
}

func TestAutoRegistry(t *testing.T) {
	registry := NewRegistry()
	autoRegistry := NewAutoRegistry(registry)
	assert.NotNil(t, autoRegistry)

	// Test package registration (mock implementation)
	err := autoRegistry.RegisterPackage("github.com/example/workflows", "test-queue")
	// This will return nil as the ScanPackage is a stub
	assert.NoError(t, err)
}

func TestTypeScanner(t *testing.T) {
	scanner := NewTypeScanner()
	assert.NotNil(t, scanner)

	// Test package scanning (mock implementation)
	workflows, activities, err := scanner.ScanPackage("github.com/example/workflows")
	assert.NoError(t, err)
	assert.Empty(t, workflows)
	assert.Empty(t, activities)
}