// Package temporal provides Temporal workflow infrastructure
package temporal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config GeneratorConfig
		want   *Generator
	}{
		{
			name: "default config",
			config: GeneratorConfig{},
			want: &Generator{
				templateDir: "internal/temporal/templates",
				outputDir:   "pkg/workflows/generated",
			},
		},
		{
			name: "custom config",
			config: GeneratorConfig{
				TemplateDir: "/custom/templates",
				OutputDir:   "/custom/output",
			},
			want: &Generator{
				templateDir: "/custom/templates",
				outputDir:   "/custom/output",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.config)
			require.NoError(t, err)
			assert.NotNil(t, gen)
			assert.Equal(t, tt.want.templateDir, gen.templateDir)
			assert.Equal(t, tt.want.outputDir, gen.outputDir)
			assert.NotNil(t, gen.validator)
			assert.NotEmpty(t, gen.templates)
		})
	}
}

func TestGenerateBasicWorkflow(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "TestWorkflow",
		Package:     "testpkg",
		Description: "A test workflow",
		InputType:   "string `json:\"input\"`",
		OutputType:  "string `json:\"output\"`",
		Template:    "basic",
		Activities: []ActivitySpec{
			{
				Name:        "ProcessData",
				Description: "Process input data",
				InputType:   "string",
				OutputType:  "string",
				Timeout:     10 * time.Minute,
			},
		},
		Options: WorkflowOptions{
			TaskQueue:                "test-queue",
			WorkflowExecutionTimeout: 1 * time.Hour,
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check generated workflow code
	assert.Contains(t, code.WorkflowCode, "package testpkg")
	assert.Contains(t, code.WorkflowCode, "type TestWorkflowWorkflow struct")
	assert.Contains(t, code.WorkflowCode, "func (w *TestWorkflowWorkflow) Execute")
	assert.Contains(t, code.WorkflowCode, "ProcessDataActivity")
	assert.Contains(t, code.WorkflowCode, "workflow.ExecuteActivity")

	// Check generated activity code
	assert.Contains(t, code.ActivityCode, "func ProcessDataActivity")

	// Check generated test code
	assert.Contains(t, code.TestCode, "type TestWorkflowTestSuite struct")
	assert.Contains(t, code.TestCode, "func TestTestWorkflowSuite(t *testing.T)")
}

func TestGenerateApprovalWorkflow(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "ApprovalProcess",
		Package:     "approval",
		Description: "An approval workflow with human tasks",
		InputType:   "RequestID string",
		OutputType:  "ApprovalStatus string",
		Template:    "approval",
		HumanTasks: []HumanTaskSpec{
			{
				Name:           "ManagerApproval",
				Description:    "Manager approval required",
				AssignedTo:     "manager@example.com",
				EscalationTime: 24 * time.Hour,
				EscalationTo:   "director@example.com",
				Priority:       "high",
				Deadline:       48 * time.Hour,
			},
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check approval-specific code
	assert.Contains(t, code.WorkflowCode, "ApprovalStatus")
	assert.Contains(t, code.WorkflowCode, "ApprovalRecord")
	assert.Contains(t, code.WorkflowCode, "approval signal")
	assert.Contains(t, code.WorkflowCode, "CreateHumanTaskActivity")
	assert.Contains(t, code.WorkflowCode, "EscalateTaskActivity")
}

func TestGenerateScheduledWorkflow(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "DailyReport",
		Package:     "reporting",
		Description: "A scheduled workflow that runs daily",
		InputType:   "Schedule string",
		OutputType:  "ReportSummary string",
		Template:    "scheduled",
		Activities: []ActivitySpec{
			{
				Name:        "GenerateReport",
				Description: "Generate daily report",
				InputType:   "ReportRequest",
				OutputType:  "Report",
				Timeout:     30 * time.Minute,
			},
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check scheduled workflow code
	assert.Contains(t, code.WorkflowCode, "scheduled workflow")
	assert.Contains(t, code.WorkflowCode, "workflow.Sleep")
	assert.Contains(t, code.WorkflowCode, "ExecutionCount")
	assert.Contains(t, code.WorkflowCode, "calculateNextCronTime")
}

func TestGenerateHumanTaskWorkflow(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "DocumentReview",
		Package:     "review",
		Description: "A human task workflow for document review",
		InputType:   "DocumentID string",
		OutputType:  "ReviewResult string",
		Template:    "human_task",
		HumanTasks: []HumanTaskSpec{
			{
				Name:           "LegalReview",
				Description:    "Legal team review",
				AssignedTo:     "legal-team",
				EscalationTime: 4 * time.Hour,
				EscalationTo:   "legal-manager",
				Priority:       "high",
				Deadline:       8 * time.Hour,
			},
			{
				Name:           "ComplianceReview",
				Description:    "Compliance team review",
				AssignedTo:     "compliance-team",
				EscalationTime: 6 * time.Hour,
				EscalationTo:   "compliance-manager",
				Priority:       "medium",
				Deadline:       12 * time.Hour,
			},
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check human task specific code
	assert.Contains(t, code.WorkflowCode, "HumanTask")
	assert.Contains(t, code.WorkflowCode, "TaskStatus")
	assert.Contains(t, code.WorkflowCode, "Priority")
	assert.Contains(t, code.WorkflowCode, "taskCompleteChan")
	assert.Contains(t, code.WorkflowCode, "taskReassignChan")
	assert.Contains(t, code.WorkflowCode, "LegalReview")
	assert.Contains(t, code.WorkflowCode, "ComplianceReview")
}

func TestGenerateLongRunningWorkflow(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "DataMigration",
		Package:     "migration",
		Description: "A long-running data migration workflow",
		InputType:   "MigrationConfig string",
		OutputType:  "MigrationResult string",
		Template:    "long_running",
		Activities: []ActivitySpec{
			{
				Name:        "Initialize",
				Description: "Initialize migration",
				InputType:   "Config",
				OutputType:  "InitResult",
				Timeout:     1 * time.Hour,
			},
		},
		Timeouts: TimeoutSpec{
			WorkflowExecution: 24 * time.Hour,
			WorkflowRun:       12 * time.Hour,
			WorkflowTask:      10 * time.Minute,
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check long-running workflow code
	assert.Contains(t, code.WorkflowCode, "Checkpoint")
	assert.Contains(t, code.WorkflowCode, "WorkflowState")
	assert.Contains(t, code.WorkflowCode, "pause signal")
	assert.Contains(t, code.WorkflowCode, "resume signal")
	assert.Contains(t, code.WorkflowCode, "HeartbeatTimeout")
	assert.Contains(t, code.WorkflowCode, "phases")
}

func TestGenerateWithSignalsAndQueries(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "InteractiveWorkflow",
		Package:     "interactive",
		Description: "A workflow with signals and queries",
		InputType:   "Config string",
		OutputType:  "Result string",
		Template:    "basic",
		Signals: []SignalSpec{
			{
				Name:        "UpdateConfig",
				Description: "Update workflow configuration",
				PayloadType: "ConfigUpdate",
			},
			{
				Name:        "Abort",
				Description: "Abort workflow execution",
				PayloadType: "AbortReason",
			},
		},
		Queries: []QuerySpec{
			{
				Name:         "GetStatus",
				Description:  "Get current workflow status",
				ResponseType: "WorkflowStatus",
			},
			{
				Name:         "GetProgress",
				Description:  "Get workflow progress",
				ResponseType: "ProgressInfo",
			},
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check signal handlers
	assert.Contains(t, code.WorkflowCode, "UpdateConfig signal")
	assert.Contains(t, code.WorkflowCode, "Abort signal")
	assert.Contains(t, code.WorkflowCode, "workflow.GetSignalChannel")

	// Check query handlers
	assert.Contains(t, code.WorkflowCode, "GetStatus")
	assert.Contains(t, code.WorkflowCode, "GetProgress")
	assert.Contains(t, code.WorkflowCode, "workflow.SetQueryHandler")
}

func TestGenerateWithChildWorkflows(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "ParentWorkflow",
		Package:     "parent",
		Description: "A workflow that spawns child workflows",
		InputType:   "ParentInput string",
		OutputType:  "ParentOutput string",
		Template:    "basic",
		ChildWorkflows: []ChildWorkflowSpec{
			{
				Name:       "ChildProcess",
				InputType:  "ChildInput",
				OutputType: "ChildOutput",
				TaskQueue:  "child-queue",
			},
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check child workflow code
	assert.Contains(t, code.WorkflowCode, "workflow.ExecuteChildWorkflow")
	assert.Contains(t, code.WorkflowCode, "ChildWorkflowOptions")
}

func TestGenerateWithRetryPolicy(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	spec := WorkflowSpec{
		Name:        "RetryWorkflow",
		Package:     "retry",
		Description: "A workflow with retry policies",
		InputType:   "Input string",
		OutputType:  "Output string",
		Template:    "basic",
		Activities: []ActivitySpec{
			{
				Name:        "UnreliableActivity",
				Description: "An activity that might fail",
				InputType:   "Request",
				OutputType:  "Response",
				Timeout:     5 * time.Minute,
				RetryPolicy: RetryPolicy{
					InitialInterval:    1 * time.Second,
					BackoffCoefficient: 2.0,
					MaximumInterval:    1 * time.Minute,
					MaximumAttempts:    5,
				},
			},
		},
		Retries: RetrySpec{
			MaxAttempts:        3,
			BackoffCoefficient: 1.5,
		},
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, code)

	// Check retry policy code
	assert.Contains(t, code.WorkflowCode, "RetryPolicy")
	assert.Contains(t, code.WorkflowCode, "InitialInterval")
	assert.Contains(t, code.WorkflowCode, "BackoffCoefficient")
	assert.Contains(t, code.WorkflowCode, "MaximumAttempts")
}

func TestValidateGeneratedCode(t *testing.T) {
	gen, err := NewGenerator(GeneratorConfig{})
	require.NoError(t, err)

	// Generate a simple valid workflow
	spec := WorkflowSpec{
		Name:        "SimpleWorkflow",
		Package:     "simple",
		Description: "A simple workflow for validation",
		InputType:   "string",
		OutputType:  "string",
		Template:    "basic",
	}

	ctx := context.Background()
	code, err := gen.GenerateWorkflow(ctx, spec)
	require.NoError(t, err)

	// The generated code should be valid Go code
	assert.NotEmpty(t, code.WorkflowCode)
	assert.Empty(t, code.Errors) // No validation errors expected for basic workflow
}

func TestTemplateHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{
			name:     "capitalize",
			function: capitalize,
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "capitalize empty",
			function: capitalize,
			input:    "",
			expected: "",
		},
		{
			name:     "toSnakeCase",
			function: toSnakeCase,
			input:    "HelloWorld",
			expected: "hello_world",
		},
		{
			name:     "toSnakeCase simple",
			function: toSnakeCase,
			input:    "simple",
			expected: "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	duration := 5 * time.Minute
	result := formatDuration(duration)
	assert.Contains(t, result, "time.Duration")
	assert.Contains(t, result, "300000000000") // 5 minutes in nanoseconds
}

func TestIndent(t *testing.T) {
	text := "line1\nline2\nline3"
	result := indent(4, text)
	lines := strings.Split(result, "\n")
	assert.Equal(t, "    line1", lines[0])
	assert.Equal(t, "    line2", lines[1])
	assert.Equal(t, "    line3", lines[2])
}

func TestGetFileName(t *testing.T) {
	gen, _ := NewGenerator(GeneratorConfig{})

	tests := []struct {
		workflowName string
		expected     string
	}{
		{"SimpleWorkflow", "simple_workflow_workflow.go"},
		{"MyComplexWorkflow", "my_complex_workflow_workflow.go"},
		{"workflow", "workflow_workflow.go"},
	}

	for _, tt := range tests {
		t.Run(tt.workflowName, func(t *testing.T) {
			result := gen.getFileName(tt.workflowName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTestFileName(t *testing.T) {
	gen, _ := NewGenerator(GeneratorConfig{})

	tests := []struct {
		workflowName string
		expected     string
	}{
		{"SimpleWorkflow", "simple_workflow_workflow_test.go"},
		{"MyComplexWorkflow", "my_complex_workflow_workflow_test.go"},
		{"workflow", "workflow_workflow_test.go"},
	}

	for _, tt := range tests {
		t.Run(tt.workflowName, func(t *testing.T) {
			result := gen.getTestFileName(tt.workflowName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateImports(t *testing.T) {
	gen, _ := NewGenerator(GeneratorConfig{})

	tests := []struct {
		name     string
		spec     WorkflowSpec
		expected []string
	}{
		{
			name: "basic imports",
			spec: WorkflowSpec{},
			expected: []string{
				"context",
				"fmt",
				"time",
				"go.temporal.io/sdk/workflow",
				"go.temporal.io/sdk/activity",
			},
		},
		{
			name: "with signals and queries",
			spec: WorkflowSpec{
				Signals: []SignalSpec{{Name: "test"}},
				Queries: []QuerySpec{{Name: "test"}},
			},
			expected: []string{
				"context",
				"fmt",
				"time",
				"go.temporal.io/sdk/workflow",
				"go.temporal.io/sdk/activity",
				"go.temporal.io/sdk/temporal",
			},
		},
		{
			name: "with human tasks",
			spec: WorkflowSpec{
				HumanTasks: []HumanTaskSpec{{Name: "test"}},
			},
			expected: []string{
				"context",
				"fmt",
				"time",
				"go.temporal.io/sdk/workflow",
				"go.temporal.io/sdk/activity",
				"go.temporal.io/sdk/log",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := gen.generateImports(tt.spec)
			for _, exp := range tt.expected {
				assert.Contains(t, imports, exp)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerateBasicWorkflow(b *testing.B) {
	gen, _ := NewGenerator(GeneratorConfig{})
	spec := WorkflowSpec{
		Name:        "BenchWorkflow",
		Package:     "bench",
		Description: "Benchmark workflow",
		InputType:   "string",
		OutputType:  "string",
		Template:    "basic",
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateWorkflow(ctx, spec)
	}
}

func BenchmarkGenerateComplexWorkflow(b *testing.B) {
	gen, _ := NewGenerator(GeneratorConfig{})
	spec := WorkflowSpec{
		Name:        "ComplexBenchWorkflow",
		Package:     "bench",
		Description: "Complex benchmark workflow",
		InputType:   "ComplexInput",
		OutputType:  "ComplexOutput",
		Template:    "human_task",
		Activities: []ActivitySpec{
			{Name: "Activity1", InputType: "Input1", OutputType: "Output1"},
			{Name: "Activity2", InputType: "Input2", OutputType: "Output2"},
			{Name: "Activity3", InputType: "Input3", OutputType: "Output3"},
		},
		Signals: []SignalSpec{
			{Name: "Signal1", PayloadType: "Payload1"},
			{Name: "Signal2", PayloadType: "Payload2"},
		},
		Queries: []QuerySpec{
			{Name: "Query1", ResponseType: "Response1"},
			{Name: "Query2", ResponseType: "Response2"},
		},
		HumanTasks: []HumanTaskSpec{
			{Name: "Task1", AssignedTo: "user1"},
			{Name: "Task2", AssignedTo: "user2"},
		},
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateWorkflow(ctx, spec)
	}
}