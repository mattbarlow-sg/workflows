package temporal

import (
	"context"
	"testing"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/src/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBPMNAdapter_Convert(t *testing.T) {
	adapter := NewBPMNAdapter()
	ctx := context.Background()

	// Create a sample BPMN process
	process := &bpmn.Process{
		Type:    "bpmn:process",
		Version: "2.0",
		ProcessInfo: bpmn.ProcessInfo{
			ID:          "test-process",
			Name:        "Test Process",
			Description: "A test BPMN process",
			IsExecutable: true,
			Elements: bpmn.Elements{
				Events: []bpmn.Event{
					{
						ID:        "start-event",
						Name:      "Start",
						Type:      "startEvent",
						EventType: "none",
						Outgoing:  []string{"flow1"},
					},
					{
						ID:        "end-event",
						Name:      "End",
						Type:      "endEvent",
						EventType: "none",
						Incoming:  []string{"flow3"},
					},
				},
				Activities: []bpmn.Activity{
					{
						ID:          "task1",
						Name:        "Process Data",
						Type:        "serviceTask",
						Description: "Process the input data",
						Incoming:    []string{"flow1"},
						Outgoing:    []string{"flow2"},
					},
					{
						ID:          "task2",
						Name:        "Review Results",
						Type:        "userTask",
						Description: "Human review of results",
						Incoming:    []string{"flow2"},
						Outgoing:    []string{"flow3"},
					},
				},
				SequenceFlows: []bpmn.SequenceFlow{
					{
						ID:        "flow1",
						SourceRef: "start-event",
						TargetRef: "task1",
					},
					{
						ID:        "flow2",
						SourceRef: "task1",
						TargetRef: "task2",
					},
					{
						ID:        "flow3",
						SourceRef: "task2",
						TargetRef: "end-event",
					},
				},
			},
		},
	}

	config := schemas.ConversionConfig{
		SourceFile:  "test.bpmn",
		PackageName: "testworkflow",
		OutputDir:   "/tmp/generated",
		Options: schemas.ConversionOptions{
			GenerateTests:  true,
			InlineScripts:  true,
			MaxInlineLines: 50,
			GenerateTODOs:  true,
			PreserveDocs:   true,
			StrictMode:     false,
		},
		ValidationLevel: "lenient",
		AllowPartial:    true,
	}

	t.Run("successful conversion", func(t *testing.T) {
		result, err := adapter.Convert(ctx, process, config)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		assert.NotEmpty(t, result.GeneratedFiles)
		assert.NotEmpty(t, result.Mappings)
		
		// Check that key files were generated
		hasWorkflow := false
		hasActivities := false
		hasTypes := false
		for _, file := range result.GeneratedFiles {
			switch file.Type {
			case schemas.FileTypeWorkflow:
				hasWorkflow = true
			case schemas.FileTypeActivities:
				hasActivities = true
			case schemas.FileTypeTypes:
				hasTypes = true
			}
		}
		assert.True(t, hasWorkflow, "Should generate workflow file")
		assert.True(t, hasActivities, "Should generate activities file")
		assert.True(t, hasTypes, "Should generate types file")

		// Check stats
		assert.Equal(t, 4, result.Metadata.Stats.TotalElements) // 2 events + 2 activities
		assert.Greater(t, result.Metadata.Stats.ConvertedElements, 0)
		assert.Greater(t, result.Metadata.Stats.LinesOfCode, 0)
	})

	t.Run("strict mode with unsupported elements", func(t *testing.T) {
		// Add an unsupported element
		processWithComplex := *process
		processWithComplex.ProcessInfo.Elements.Gateways = []bpmn.Gateway{
			{
				ID:   "complex-gateway",
				Name: "Complex Decision",
				Type: "complexGateway",
			},
		}

		strictConfig := config
		strictConfig.Options.StrictMode = true

		result, err := adapter.Convert(ctx, &processWithComplex, strictConfig)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.False(t, result.Success)
		assert.NotEmpty(t, result.Issues)
		assert.Greater(t, result.ValidationResult.UnsupportedElements, 0)
	})
}

func TestBPMNAdapter_ValidateProcess(t *testing.T) {
	adapter := NewBPMNAdapter()

	t.Run("valid process", func(t *testing.T) {
		process := &bpmn.ProcessInfo{
			ID:   "valid-process",
			Name: "Valid Process",
			Elements: bpmn.Elements{
				Events: []bpmn.Event{
					{
						ID:        "start",
						Type:      "startEvent",
						EventType: "none",
					},
					{
						ID:        "end",
						Type:      "endEvent",
						EventType: "none",
					},
				},
				Activities: []bpmn.Activity{
					{
						ID:   "task",
						Name: "Simple Task",
						Type: "task",
					},
				},
			},
		}

		result, err := adapter.ValidateProcess(process)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Equal(t, 3, result.SupportedElements)
		assert.Equal(t, 0, result.UnsupportedElements)
	})

	t.Run("process with unsupported elements", func(t *testing.T) {
		process := &bpmn.ProcessInfo{
			ID:   "complex-process",
			Name: "Complex Process",
			Elements: bpmn.Elements{
				Gateways: []bpmn.Gateway{
					{
						ID:   "complex",
						Name: "Complex Gateway",
						Type: "complexGateway",
					},
				},
			},
		}

		result, err := adapter.ValidateProcess(process)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
		assert.Equal(t, 0, result.SupportedElements)
		assert.Equal(t, 1, result.UnsupportedElements)
	})

	t.Run("process with warnings", func(t *testing.T) {
		process := &bpmn.ProcessInfo{
			ID:   "warning-process",
			Name: "Warning Process",
			Elements: bpmn.Elements{
				Activities: []bpmn.Activity{
					{
						ID:   "manual",
						Name: "Manual Task",
						Type: "manualTask",
					},
				},
			},
		}

		result, err := adapter.ValidateProcess(process)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEmpty(t, result.Warnings)
		assert.Equal(t, 1, result.SupportedElements)
	})
}

func TestElementMapper_MapActivity(t *testing.T) {
	mapper := NewElementMapper()

	t.Run("service task mapping", func(t *testing.T) {
		activity := bpmn.Activity{
			ID:          "service-task",
			Name:        "Process Order",
			Type:        "serviceTask",
			Description: "Process customer order",
		}

		mapping, err := mapper.MapActivity(activity)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "Process Order", mapping.Name)
		assert.Equal(t, "ProcessOrder", mapping.FunctionName)
		assert.False(t, mapping.IsHumanTask)
		assert.False(t, mapping.HasSignal)
		assert.NotNil(t, mapping.Options.RetryPolicy)
	})

	t.Run("user task mapping", func(t *testing.T) {
		activity := bpmn.Activity{
			ID:          "user-task",
			Name:        "Approve Request",
			Type:        "userTask",
			Description: "Manager approval required",
		}

		mapping, err := mapper.MapActivity(activity)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "Approve Request", mapping.Name)
		assert.Equal(t, "ApproveRequest", mapping.FunctionName)
		assert.True(t, mapping.IsHumanTask)
		assert.Equal(t, 24*time.Hour, mapping.Options.ScheduleToCloseTimeout)
	})

	t.Run("receive task mapping", func(t *testing.T) {
		activity := bpmn.Activity{
			ID:          "receive-task",
			Name:        "Wait for Payment",
			Type:        "receiveTask",
			Description: "Wait for payment confirmation",
		}

		mapping, err := mapper.MapActivity(activity)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.True(t, mapping.HasSignal)
		assert.Equal(t, "WAIT_FOR_PAYMENT", mapping.SignalName)
	})
}

func TestElementMapper_MapGateway(t *testing.T) {
	mapper := NewElementMapper()

	t.Run("exclusive gateway mapping", func(t *testing.T) {
		gateway := bpmn.Gateway{
			ID:          "exclusive",
			Name:        "Decision Point",
			Type:        "exclusiveGateway",
			DefaultFlow: "default-path",
			Outgoing:    []string{"path1", "path2", "default-path"},
		}

		mapping, err := mapper.MapGateway(gateway)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "exclusiveGateway", mapping.Type)
		assert.Equal(t, "selector", mapping.ControlFlow)
		assert.Equal(t, "default-path", mapping.DefaultPath)
		assert.NotEmpty(t, mapping.Conditions)
	})

	t.Run("parallel gateway mapping", func(t *testing.T) {
		gateway := bpmn.Gateway{
			ID:       "parallel",
			Name:     "Fork",
			Type:     "parallelGateway",
			Outgoing: []string{"branch1", "branch2", "branch3"},
		}

		mapping, err := mapper.MapGateway(gateway)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "parallelGateway", mapping.Type)
		assert.Equal(t, "parallel", mapping.ControlFlow)
		assert.Equal(t, 3, mapping.ParallelCount)
	})

	t.Run("unsupported gateway", func(t *testing.T) {
		gateway := bpmn.Gateway{
			ID:   "complex",
			Name: "Complex",
			Type: "complexGateway",
		}

		_, err := mapper.MapGateway(gateway)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported gateway type")
	})
}

func TestElementMapper_MapEvent(t *testing.T) {
	mapper := NewElementMapper()

	t.Run("start event mapping", func(t *testing.T) {
		event := bpmn.Event{
			ID:        "start",
			Name:      "Process Start",
			Type:      "startEvent",
			EventType: "none",
		}

		mapping, err := mapper.MapEvent(event)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.True(t, mapping.IsStart)
		assert.False(t, mapping.IsEnd)
		assert.Equal(t, "none", mapping.HandlerType)
	})

	t.Run("timer event mapping", func(t *testing.T) {
		event := bpmn.Event{
			ID:        "timer",
			Name:      "Wait Timer",
			Type:      "intermediateCatchEvent",
			EventType: "timer",
			Properties: map[string]any{
				"duration": "5m",
			},
		}

		mapping, err := mapper.MapEvent(event)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "timer", mapping.HandlerType)
		assert.Equal(t, 5*time.Minute, mapping.TimerDuration)
	})

	t.Run("signal event mapping", func(t *testing.T) {
		event := bpmn.Event{
			ID:        "signal",
			Name:      "Order Received",
			Type:      "intermediateCatchEvent",
			EventType: "signal",
		}

		mapping, err := mapper.MapEvent(event)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.Equal(t, "signal", mapping.HandlerType)
		assert.Equal(t, "ORDER_RECEIVED", mapping.SignalName)
	})

	t.Run("boundary event mapping", func(t *testing.T) {
		event := bpmn.Event{
			ID:          "boundary",
			Name:        "Timeout",
			Type:        "boundaryEvent",
			EventType:   "timer",
			AttachedTo:  "task1",
			IsInterrupt: true,
		}

		mapping, err := mapper.MapEvent(event)
		require.NoError(t, err)
		require.NotNil(t, mapping)

		assert.True(t, mapping.IsBoundary)
		assert.True(t, mapping.IsInterrupting)
		assert.Equal(t, "timer", mapping.HandlerType)
	})
}

func TestCodeGenerator_GenerateWorkflow(t *testing.T) {
	generator := NewCodeGenerator()

	mappings := []schemas.ElementMapping{
		{
			BPMNElement: schemas.BPMNElementRef{
				ID:   "task1",
				Type: "activity",
				Name: "Process Data",
			},
			TemporalConstruct: schemas.TemporalActivity,
			Status:           schemas.MappingComplete,
		},
		{
			BPMNElement: schemas.BPMNElementRef{
				ID:   "gateway1",
				Type: "gateway",
				Name: "Decision",
			},
			TemporalConstruct: schemas.TemporalSelector,
			Status:           schemas.MappingComplete,
		},
	}

	code, err := generator.GenerateWorkflow("testpkg", mappings)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Check that generated code contains expected elements
	assert.Contains(t, code, "package testpkg")
	assert.Contains(t, code, "import")
	assert.Contains(t, code, "workflow.Context")
	assert.Contains(t, code, "ProcessWorkflow")
	assert.Contains(t, code, "WorkflowInput")
	assert.Contains(t, code, "WorkflowOutput")
	assert.Contains(t, code, "Execute activity: Process Data")
	assert.Contains(t, code, "Gateway: Decision")
}

func TestCodeGenerator_GenerateActivities(t *testing.T) {
	generator := NewCodeGenerator()

	activities := []schemas.ActivityMapping{
		{
			Name:         "Process Order",
			FunctionName: "ProcessOrder",
			InputType:    "ActivityInput",
			OutputType:   "ActivityOutput",
			IsHumanTask:  false,
		},
		{
			Name:         "Approve Request",
			FunctionName: "ApproveRequest",
			InputType:    "ActivityInput",
			OutputType:   "ActivityOutput",
			IsHumanTask:  true,
		},
	}

	code, err := generator.GenerateActivities(activities)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Check that generated code contains expected elements
	assert.Contains(t, code, "package workflows")
	assert.Contains(t, code, "func ProcessOrder")
	assert.Contains(t, code, "func ApproveRequest")
	assert.Contains(t, code, "Human task started")
	assert.Contains(t, code, "RegisterActivities")
	assert.Contains(t, code, "activity.GetLogger")
}

func TestCodeGenerator_GenerateTypes(t *testing.T) {
	generator := NewCodeGenerator()

	dataObjects := []bpmn.DataObject{
		{
			ID:           "order-data",
			Name:         "OrderData",
			IsCollection: true,
			State:        "initialized",
		},
	}

	code, err := generator.GenerateTypes(dataObjects, nil)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Check that generated code contains expected elements
	assert.Contains(t, code, "package workflows")
	assert.Contains(t, code, "type WorkflowInput struct")
	assert.Contains(t, code, "type WorkflowOutput struct")
	assert.Contains(t, code, "type ActivityInput struct")
	assert.Contains(t, code, "type ActivityOutput struct")
	assert.Contains(t, code, "OrderData")
}

func TestNameSanitizer(t *testing.T) {
	sanitizer := NewNameSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		method   string
	}{
		{
			name:     "simple name to function",
			input:    "Process Order",
			expected: "ProcessOrder",
			method:   "function",
		},
		{
			name:     "with special characters",
			input:    "Process-Order-123",
			expected: "Process_Order_123",
			method:   "function",
		},
		{
			name:     "to signal name",
			input:    "Order Received",
			expected: "ORDER_RECEIVED",
			method:   "signal",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "Field",
			method:   "function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.method == "function" {
				result = sanitizer.ToFunctionName(tt.input)
			} else if tt.method == "signal" {
				result = sanitizer.ToSignalName(tt.input)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImportsManager(t *testing.T) {
	manager := NewImportsManager()

	t.Run("generate imports", func(t *testing.T) {
		manager.AddImport("fmt", "")
		manager.AddImport("github.com/stretchr/testify/assert", "")
		manager.AddImport("context", "")

		imports := manager.Generate()
		assert.Contains(t, imports, "import (")
		assert.Contains(t, imports, `"fmt"`)
		assert.Contains(t, imports, `"context"`)
		assert.Contains(t, imports, `"github.com/stretchr/testify/assert"`)
	})

	t.Run("generate with alias", func(t *testing.T) {
		manager.Reset()
		manager.AddImport("go.temporal.io/sdk/workflow", "wf")

		imports := manager.Generate()
		assert.Contains(t, imports, `wf "go.temporal.io/sdk/workflow"`)
	})

	t.Run("empty imports", func(t *testing.T) {
		manager.Reset()
		imports := manager.Generate()
		assert.Empty(t, imports)
	})
}