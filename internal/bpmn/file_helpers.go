package bpmn

import (
	"encoding/json"
	"fmt"
	"os"
)

// buildFlowConnections populates incoming/outgoing arrays from sequence flows
func buildFlowConnections(process *Process) {
	if process == nil || process.ProcessInfo.Elements.SequenceFlows == nil {
		return
	}

	// Clear existing connections
	for i := range process.ProcessInfo.Elements.Events {
		process.ProcessInfo.Elements.Events[i].Incoming = []string{}
		process.ProcessInfo.Elements.Events[i].Outgoing = []string{}
	}
	for i := range process.ProcessInfo.Elements.Activities {
		process.ProcessInfo.Elements.Activities[i].Incoming = []string{}
		process.ProcessInfo.Elements.Activities[i].Outgoing = []string{}
	}
	for i := range process.ProcessInfo.Elements.Gateways {
		process.ProcessInfo.Elements.Gateways[i].Incoming = []string{}
		process.ProcessInfo.Elements.Gateways[i].Outgoing = []string{}
	}

	// Build connections from sequence flows
	for _, flow := range process.ProcessInfo.Elements.SequenceFlows {
		// Add outgoing to source
		addOutgoingToElement(process, flow.SourceRef, flow.ID)
		// Add incoming to target
		addIncomingToElement(process, flow.TargetRef, flow.ID)
	}
}

func addOutgoingToElement(process *Process, elementID, flowID string) {
	// Check events
	for i := range process.ProcessInfo.Elements.Events {
		if process.ProcessInfo.Elements.Events[i].ID == elementID {
			process.ProcessInfo.Elements.Events[i].Outgoing = append(
				process.ProcessInfo.Elements.Events[i].Outgoing, flowID)
			return
		}
	}

	// Check activities
	for i := range process.ProcessInfo.Elements.Activities {
		if process.ProcessInfo.Elements.Activities[i].ID == elementID {
			process.ProcessInfo.Elements.Activities[i].Outgoing = append(
				process.ProcessInfo.Elements.Activities[i].Outgoing, flowID)
			return
		}
	}

	// Check gateways
	for i := range process.ProcessInfo.Elements.Gateways {
		if process.ProcessInfo.Elements.Gateways[i].ID == elementID {
			process.ProcessInfo.Elements.Gateways[i].Outgoing = append(
				process.ProcessInfo.Elements.Gateways[i].Outgoing, flowID)
			return
		}
	}
}

func addIncomingToElement(process *Process, elementID, flowID string) {
	// Check events
	for i := range process.ProcessInfo.Elements.Events {
		if process.ProcessInfo.Elements.Events[i].ID == elementID {
			process.ProcessInfo.Elements.Events[i].Incoming = append(
				process.ProcessInfo.Elements.Events[i].Incoming, flowID)
			return
		}
	}

	// Check activities
	for i := range process.ProcessInfo.Elements.Activities {
		if process.ProcessInfo.Elements.Activities[i].ID == elementID {
			process.ProcessInfo.Elements.Activities[i].Incoming = append(
				process.ProcessInfo.Elements.Activities[i].Incoming, flowID)
			return
		}
	}

	// Check gateways
	for i := range process.ProcessInfo.Elements.Gateways {
		if process.ProcessInfo.Elements.Gateways[i].ID == elementID {
			process.ProcessInfo.Elements.Gateways[i].Incoming = append(
				process.ProcessInfo.Elements.Gateways[i].Incoming, flowID)
			return
		}
	}
}

// FileValidator provides file-based validation
type FileValidator struct{}

// ValidateFile validates a BPMN file
func (v *FileValidator) ValidateFile(filePath string) (*ValidationResult, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse the process
	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Message: fmt.Sprintf("Invalid JSON: %v", err)}},
		}, nil
	}

	// Build incoming/outgoing connections from sequence flows
	buildFlowConnections(&process)

	// Create validator with the process
	validator := NewValidator(&process)

	// Validate
	return validator.Validate(), nil
}

// FileAnalyzer provides file-based analysis
type FileAnalyzer struct{}

// AnalyzeFile analyzes a BPMN file
func (a *FileAnalyzer) AnalyzeFile(filePath string) (*AnalysisResult, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse the process
	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// Build incoming/outgoing connections from sequence flows
	buildFlowConnections(&process)

	// Create analyzer with the process
	analyzer := NewAnalyzer(&process)

	// Analyze
	return analyzer.Analyze(), nil
}

// FileRenderer provides file-based rendering
type FileRenderer struct{}

// RenderDotFile renders a BPMN file to DOT format
func (r *FileRenderer) RenderDotFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(process)
	return renderer.Render(FormatDOT)
}

// RenderMermaidFile renders a BPMN file to Mermaid format
func (r *FileRenderer) RenderMermaidFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(process)
	// Mermaid is rendered as DOT for now
	return renderer.Render(FormatDOT)
}

// RenderTextFile renders a BPMN file to text format
func (r *FileRenderer) RenderTextFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(process)
	return renderer.Render(FormatText)
}

// RenderMarkdownFile renders a BPMN file to markdown format
func (r *FileRenderer) RenderMarkdownFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(process)
	return renderer.Render(FormatMarkdown)
}

// loadProcess loads a process from a file
func (r *FileRenderer) loadProcess(filePath string) (*Process, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// Build incoming/outgoing connections from sequence flows
	buildFlowConnections(&process)

	return &process, nil
}
