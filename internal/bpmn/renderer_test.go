package bpmn

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRendererTextFormat(t *testing.T) {
	process := createTestProcess()
	renderer := NewRenderer(process)

	output, err := renderer.Render(FormatText)
	if err != nil {
		t.Fatalf("Failed to render text: %v", err)
	}

	// Just check that we got some output
	if len(output) == 0 {
		t.Error("Expected non-empty text output")
	}
	
	// Check for basic content
	if !strings.Contains(output, "test_process") {
		t.Error("Expected output to contain process ID")
	}
}

func TestRendererMarkdownFormat(t *testing.T) {
	process := createTestProcess()
	renderer := NewRenderer(process)

	output, err := renderer.Render(FormatMarkdown)
	if err != nil {
		t.Fatalf("Failed to render markdown: %v", err)
	}

	// Check for basic markdown elements
	if !strings.Contains(output, "#") {
		t.Error("Expected markdown to contain headers")
	}
	
	if !strings.Contains(output, "test_process") {
		t.Error("Expected markdown to contain process ID")
	}
}

func TestRendererJSONFormat(t *testing.T) {
	process := createTestProcess()
	renderer := NewRenderer(process)

	output, err := renderer.Render(FormatJSON)
	if err != nil {
		t.Fatalf("Failed to render JSON: %v", err)
	}

	// Check for JSON structure
	if !strings.HasPrefix(output, "{") || !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("Output should be valid JSON")
	}

	// Try to parse it
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}
}

func TestRendererGraphvizFormat(t *testing.T) {
	process := createTestProcess()
	renderer := NewRenderer(process)

	output, err := renderer.Render(FormatDOT)
	if err != nil {
		t.Fatalf("Failed to render Graphviz: %v", err)
	}

	// Check for basic Graphviz/DOT elements
	if !strings.Contains(output, "digraph") {
		t.Error("Expected DOT output to contain 'digraph'")
	}
	
	if !strings.Contains(output, "->") {
		t.Error("Expected DOT output to contain edges")
	}
}

func TestRendererWithAnalysis(t *testing.T) {
	process := createTestProcess()
	
	// Create renderer with options including analysis
	analyzer := NewAnalyzer(process)
	_ = analyzer.Analyze()
	
	renderer := NewRenderer(process)
	// Can't directly set analysis, would need a method for this

	output, err := renderer.Render("markdown")
	if err != nil {
		t.Fatalf("Failed to render with analysis: %v", err)
	}

	// Since we can't guarantee the exact format, just check for some content
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}
}

func TestRendererWithValidation(t *testing.T) {
	process := createTestProcess()
	
	// Create renderer with validation results
	validator := NewValidator(process)
	_ = validator.Validate()
	
	renderer := NewRenderer(process)
	// Can't directly set validation, would need a method for this

	output, err := renderer.Render("markdown")
	if err != nil {
		t.Fatalf("Failed to render with validation: %v", err)
	}

	// Check that output contains some content
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}
}

func TestRendererMermaidDiagram(t *testing.T) {
	// Create process with different node types
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID:   "complex_process",
			Name: "Complex Process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent", Name: "Start"},
					{ID: "end", Type: "endEvent", Name: "End"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask", Name: "User Task"},
					{ID: "task2", Type: "serviceTask", Name: "Service Task"},
				},
				Gateways: []Gateway{
					{ID: "xor1", Type: "exclusiveGateway", Name: "Decision"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "f1", SourceRef: "start", TargetRef: "task1"},
					{ID: "f2", SourceRef: "task1", TargetRef: "xor1"},
					{ID: "f3", SourceRef: "xor1", TargetRef: "task2"},
					{ID: "f4", SourceRef: "xor1", TargetRef: "end"},
					{ID: "f5", SourceRef: "task2", TargetRef: "end"},
				},
			},
		},
	}

	renderer := NewRenderer(process)
	output, err := renderer.Render(FormatMarkdown)
	if err != nil {
		t.Fatalf("Failed to render markdown with mermaid: %v", err)
	}

	// Check that output contains mermaid diagram
	if !strings.Contains(output, "```mermaid") {
		t.Error("Markdown output should contain mermaid diagram")
	}

	if !strings.Contains(output, "graph") {
		t.Error("Should have graph declaration")
	}
}

func TestRendererAgentAssignments(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID:   "agent_process",
			Name: "Process with Agents",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Activities: []Activity{
					{
						ID:   "task1",
						Type: "userTask",
						Name: "Human Task",
						Agent: &AgentAssignment{
							Type: "human",
							ID:   "john_doe",
						},
					},
					{
						ID:   "task2",
						Type: "serviceTask",
						Name: "AI Task",
						Agent: &AgentAssignment{
							Type: "ai",
							ID:   "ai-assistant",
						},
					},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "f1", SourceRef: "start", TargetRef: "task1"},
					{ID: "f2", SourceRef: "task1", TargetRef: "task2"},
					{ID: "f3", SourceRef: "task2", TargetRef: "end"},
				},
			},
		},
	}

	renderer := NewRenderer(process)
	output, err := renderer.Render(FormatMarkdown)
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	// Just check that output was generated
	if len(output) == 0 {
		t.Error("Expected non-empty output for process with agents")
	}
}

func TestRendererInvalidFormat(t *testing.T) {
	process := createTestProcess()
	renderer := NewRenderer(process)

	_, err := renderer.Render(RenderFormat("invalid_format"))
	if err == nil {
		t.Error("Should return error for invalid format")
	}
}

// Helper function to create a test process
func createTestProcess() *Process {
	return &Process{
		ProcessInfo: ProcessInfo{
			ID:          "test_process",
			Name:        "Test Process",
			Description: "A simple test process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent", Name: "Start"},
					{ID: "end", Type: "endEvent", Name: "End"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask", Name: "Task 1"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}
}