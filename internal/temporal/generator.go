// Package temporal provides Temporal workflow infrastructure
package temporal

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// Generator generates Temporal workflow code from templates
type Generator struct {
	templates      map[string]*template.Template
	validator      *TemporalValidator
	templateDir    string
	outputDir      string
}

// GeneratorConfig configures the workflow generator
type GeneratorConfig struct {
	TemplateDir string
	OutputDir   string
	Validator   *TemporalValidator
}

// NewGenerator creates a new workflow generator
func NewGenerator(config GeneratorConfig) (*Generator, error) {
	if config.TemplateDir == "" {
		config.TemplateDir = "internal/temporal/templates"
	}
	if config.OutputDir == "" {
		config.OutputDir = "pkg/workflows/generated"
	}
	if config.Validator == nil {
		config.Validator = NewTemporalValidator()
	}

	g := &Generator{
		templates:   make(map[string]*template.Template),
		validator:   config.Validator,
		templateDir: config.TemplateDir,
		outputDir:   config.OutputDir,
	}

	// Load templates
	if err := g.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return g, nil
}

// WorkflowSpec defines the specification for generating a workflow
type WorkflowSpec struct {
	Name           string
	Package        string
	Description    string
	InputType      string
	OutputType     string
	Activities     []ActivitySpec
	Signals        []SignalSpec
	Queries        []QuerySpec
	HumanTasks     []HumanTaskSpec
	Options        WorkflowOptions
	ChildWorkflows []ChildWorkflowSpec
	Retries        RetrySpec
	Timeouts       TimeoutSpec
	Template       string // Template name to use
}

// ActivitySpec defines an activity specification
type ActivitySpec struct {
	Name        string
	Description string
	InputType   string
	OutputType  string
	Timeout     time.Duration
	RetryPolicy RetryPolicy
	IsHumanTask bool
}

// SignalSpec defines a signal specification
type SignalSpec struct {
	Name        string
	Description string
	PayloadType string
}

// QuerySpec defines a query specification
type QuerySpec struct {
	Name         string
	Description  string
	ResponseType string
}

// HumanTaskSpec defines a human task specification
type HumanTaskSpec struct {
	Name           string
	Description    string
	AssignedTo     string
	EscalationTime time.Duration
	EscalationTo   string
	Priority       string
	Deadline       time.Duration
}

// ChildWorkflowSpec defines a child workflow specification
type ChildWorkflowSpec struct {
	Name       string
	InputType  string
	OutputType string
	TaskQueue  string
}

// WorkflowOptions defines workflow execution options
type WorkflowOptions struct {
	TaskQueue           string
	WorkflowExecutionTimeout time.Duration
	WorkflowRunTimeout      time.Duration
	WorkflowTaskTimeout     time.Duration
	RetryPolicy         RetryPolicy
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float32
	MaximumInterval    time.Duration
	MaximumAttempts    int32
}

// RetrySpec defines workflow-level retry configuration
type RetrySpec struct {
	MaxAttempts        int
	BackoffCoefficient float32
}

// TimeoutSpec defines workflow timeout configuration
type TimeoutSpec struct {
	WorkflowExecution time.Duration
	WorkflowRun       time.Duration
	WorkflowTask      time.Duration
}

// GenerateWorkflow generates a workflow from a specification
func (g *Generator) GenerateWorkflow(ctx context.Context, spec WorkflowSpec) (*GeneratedCode, error) {
	// Default to basic template if not specified
	if spec.Template == "" {
		spec.Template = "basic"
	}

	// Get the appropriate template
	tmpl, exists := g.templates[spec.Template]
	if !exists {
		return nil, fmt.Errorf("template %s not found", spec.Template)
	}

	// Prepare template data
	data := g.prepareTemplateData(spec)

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code with error for debugging
		return &GeneratedCode{
			WorkflowCode: buf.String(),
			Package:      spec.Package,
			Errors:       []string{fmt.Sprintf("formatting error: %v", err)},
		}, nil
	}

	// Generate activity code if needed
	activityCode := ""
	if len(spec.Activities) > 0 {
		activityCode, err = g.generateActivities(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to generate activities: %w", err)
		}
	}

	// Generate test code
	testCode, err := g.generateTests(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests: %w", err)
	}

	// Create generated code structure
	result := &GeneratedCode{
		WorkflowCode:  string(formatted),
		ActivityCode:  activityCode,
		TestCode:      testCode,
		Package:       spec.Package,
		WorkflowName:  spec.Name,
		FileName:      g.getFileName(spec.Name),
		TestFileName:  g.getTestFileName(spec.Name),
	}

	// Validate generated code
	if err := g.validateGeneratedCode(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("validation error: %v", err))
	}

	return result, nil
}

// GeneratedCode represents generated workflow code
type GeneratedCode struct {
	WorkflowCode  string
	ActivityCode  string
	TestCode      string
	Package       string
	WorkflowName  string
	FileName      string
	TestFileName  string
	Errors        []string
	Warnings      []string
}

// prepareTemplateData prepares data for template execution
func (g *Generator) prepareTemplateData(spec WorkflowSpec) map[string]interface{} {
	return map[string]interface{}{
		"Package":        spec.Package,
		"WorkflowName":   spec.Name,
		"Description":    spec.Description,
		"InputType":      spec.InputType,
		"OutputType":     spec.OutputType,
		"Activities":     spec.Activities,
		"Signals":        spec.Signals,
		"Queries":        spec.Queries,
		"HumanTasks":     spec.HumanTasks,
		"Options":        spec.Options,
		"ChildWorkflows": spec.ChildWorkflows,
		"Retries":        spec.Retries,
		"Timeouts":       spec.Timeouts,
		"Timestamp":      time.Now().Format(time.RFC3339),
		"Imports":        g.generateImports(spec),
	}
}

// generateImports generates required imports for the workflow
func (g *Generator) generateImports(spec WorkflowSpec) []string {
	imports := []string{
		"context",
		"fmt",
		"time",
		"go.temporal.io/sdk/workflow",
		"go.temporal.io/sdk/activity",
	}

	// Add temporal types if needed
	if len(spec.Signals) > 0 || len(spec.Queries) > 0 {
		imports = append(imports, "go.temporal.io/sdk/temporal")
	}

	// Add log package if human tasks exist
	if len(spec.HumanTasks) > 0 {
		imports = append(imports, "go.temporal.io/sdk/log")
	}

	return imports
}

// generateActivities generates activity code
func (g *Generator) generateActivities(spec WorkflowSpec) (string, error) {
	tmpl, exists := g.templates["activity"]
	if !exists {
		return "", fmt.Errorf("activity template not found")
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Package":    spec.Package,
		"Activities": spec.Activities,
		"Timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute activity template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String(), nil // Return unformatted
	}

	return string(formatted), nil
}

// generateTests generates test code for the workflow
func (g *Generator) generateTests(spec WorkflowSpec) (string, error) {
	tmpl, exists := g.templates["test"]
	if !exists {
		return "", fmt.Errorf("test template not found")
	}

	var buf bytes.Buffer
	data := g.prepareTemplateData(spec)

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute test template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String(), nil // Return unformatted
	}

	return string(formatted), nil
}

// validateGeneratedCode validates the generated workflow code
func (g *Generator) validateGeneratedCode(ctx context.Context, code *GeneratedCode) error {
	// Create a temporary file for validation
	tmpFile := filepath.Join("/tmp", code.FileName)
	if err := ioutil.WriteFile(tmpFile, []byte(code.WorkflowCode), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Run validation
	request := schemas.ValidationRequest{
		WorkflowID:   code.WorkflowName,
		WorkflowPath: tmpFile,
		Options: schemas.ValidationOptions{
			ParallelChecks: true,
			Timeout:        30 * time.Second,
		},
	}

	result, err := g.validator.Validate(ctx, request)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Success {
		var errors []string
		for _, e := range result.Errors {
			errors = append(errors, e.Message)
		}
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// loadTemplates loads all templates from the template directory
func (g *Generator) loadTemplates() error {
	// Define template functions
	funcMap := template.FuncMap{
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title,
		"capitalize": capitalize,
		"duration":   formatDuration,
		"indent":     indent,
	}

	// Template definitions - embedded for reliability
	templates := map[string]string{
		"basic":           basicWorkflowTemplate,
		"approval":        approvalWorkflowTemplate,
		"scheduled":       scheduledWorkflowTemplate,
		"human_task":      humanTaskWorkflowTemplate,
		"long_running":    longRunningWorkflowTemplate,
		"activity":        activityTemplate,
		"test":            testTemplate,
		"signal_handler":  signalHandlerTemplate,
		"query_handler":   queryHandlerTemplate,
		"child_workflow":  childWorkflowTemplate,
	}

	// Parse templates
	for name, content := range templates {
		tmpl, err := template.New(name).Funcs(funcMap).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		g.templates[name] = tmpl
	}

	return nil
}

// getFileName generates a file name for the workflow
func (g *Generator) getFileName(workflowName string) string {
	name := toSnakeCase(workflowName)
	return fmt.Sprintf("%s_workflow.go", name)
}

// getTestFileName generates a test file name for the workflow
func (g *Generator) getTestFileName(workflowName string) string {
	name := toSnakeCase(workflowName)
	return fmt.Sprintf("%s_workflow_test.go", name)
}

// SaveGeneratedCode saves the generated code to files
func (g *Generator) SaveGeneratedCode(code *GeneratedCode) error {
	// Ensure output directory exists
	outputPath := filepath.Join(g.outputDir, code.Package)
	if err := ensureDir(outputPath); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save workflow code
	workflowPath := filepath.Join(outputPath, code.FileName)
	if err := ioutil.WriteFile(workflowPath, []byte(code.WorkflowCode), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	// Save activity code if present
	if code.ActivityCode != "" {
		activityPath := filepath.Join(outputPath, "activities.go")
		if err := ioutil.WriteFile(activityPath, []byte(code.ActivityCode), 0644); err != nil {
			return fmt.Errorf("failed to write activity file: %w", err)
		}
	}

	// Save test code
	testPath := filepath.Join(outputPath, code.TestFileName)
	if err := ioutil.WriteFile(testPath, []byte(code.TestCode), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	return nil
}

// GenerateFromBPMN generates a workflow from a BPMN definition
func (g *Generator) GenerateFromBPMN(ctx context.Context, bpmnPath string, options BPMNConversionOptions) (*GeneratedCode, error) {
	// This would integrate with the BPMN adapter
	// For now, return a placeholder
	return nil, fmt.Errorf("BPMN generation not yet implemented")
}

// BPMNConversionOptions defines options for BPMN to workflow conversion
type BPMNConversionOptions struct {
	Package    string
	OutputDir  string
	TaskQueue  string
	Validate   bool
}

// Helper functions

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("time.Duration(%d)", d.Nanoseconds())
}

func indent(spaces int, text string) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func ensureDir(path string) error {
	// Implementation would create directory if it doesn't exist
	// For now, return nil
	return nil
}