package bpmn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// FileValidator provides file-based validation
type FileValidator struct{}

// ValidateFile validates a BPMN file
func (v *FileValidator) ValidateFile(filePath string) (*ValidationResult, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	
	// Parse the process
	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{Message: fmt.Sprintf("Invalid JSON: %v", err)}},
		}, nil
	}
	
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
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	
	// Parse the process
	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	
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
	return renderer.RenderDOT(), nil
}

// RenderMermaidFile renders a BPMN file to Mermaid format
func (r *FileRenderer) RenderMermaidFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}
	
	renderer := NewRenderer(process)
	return renderer.RenderMermaid(), nil
}

// RenderTextFile renders a BPMN file to text format
func (r *FileRenderer) RenderTextFile(filePath string) (string, error) {
	process, err := r.loadProcess(filePath)
	if err != nil {
		return "", err
	}
	
	renderer := NewRenderer(process)
	return renderer.RenderText(), nil
}

// loadProcess loads a process from a file
func (r *FileRenderer) loadProcess(filePath string) (*Process, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	
	var process Process
	if err := json.Unmarshal(data, &process); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	
	return &process, nil
}