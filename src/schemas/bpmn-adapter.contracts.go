// Package schemas provides validation contracts for BPMN to Temporal conversion
package schemas

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
)

// ValidationContract defines what must be validated during conversion
type ValidationContract struct {
	// Pre-conversion validations
	PreValidation PreValidationRules `json:"preValidation" validate:"required"`

	// During conversion validations
	ConversionValidation ConversionRules `json:"conversionValidation" validate:"required"`

	// Post-conversion validations
	PostValidation PostValidationRules `json:"postValidation" validate:"required"`
}

// PreValidationRules defines pre-conversion validation rules
type PreValidationRules struct {
	// Check BPMN is well-formed
	CheckBPMNSchema bool `json:"checkBPMNSchema" validate:"required"`

	// Check all references resolve
	CheckReferences bool `json:"checkReferences" validate:"required"`

	// Check for cycles
	CheckCycles bool `json:"checkCycles" validate:"required"`

	// Check supported elements
	CheckSupported bool `json:"checkSupported" validate:"required"`

	// Check data flow consistency
	CheckDataFlow bool `json:"checkDataFlow" validate:"required"`

	// Check naming conflicts
	CheckNamingConflicts bool `json:"checkNamingConflicts" validate:"required"`
}

// ConversionRules defines during-conversion validation rules
type ConversionRules struct {
	// Ensure deterministic patterns
	EnforceDeterminism bool `json:"enforceDeterminism" validate:"required"`

	// Validate type safety
	ValidateTypes bool `json:"validateTypes" validate:"required"`

	// Check naming conflicts
	CheckNamingConflicts bool `json:"checkNamingConflicts" validate:"required"`

	// Validate data flows
	ValidateDataFlows bool `json:"validateDataFlows" validate:"required"`

	// Check activity signatures
	ValidateActivitySignatures bool `json:"validateActivitySignatures" validate:"required"`

	// Validate signal consistency
	ValidateSignals bool `json:"validateSignals" validate:"required"`
}

// PostValidationRules defines post-conversion validation rules
type PostValidationRules struct {
	// Run Go compiler
	CompileGenerated bool `json:"compileGenerated" validate:"required"`

	// Run Temporal validator
	RunTemporalValidator bool `json:"runTemporalValidator" validate:"required"`

	// Run determinism checker
	RunDeterminismChecker bool `json:"runDeterminismChecker" validate:"required"`

	// Generate test coverage
	CheckTestCoverage bool `json:"checkTestCoverage"`

	// Minimum coverage percentage
	MinCoveragePercent int `json:"minCoveragePercent" validate:"min=0,max=100"`

	// Run go fmt
	FormatCode bool `json:"formatCode"`

	// Run go vet
	VetCode bool `json:"vetCode"`

	// Run staticcheck
	RunStaticCheck bool `json:"runStaticCheck"`
}

// BPMNValidator validates BPMN processes for conversion
type BPMNValidator interface {
	// ValidateProcess validates a BPMN process
	ValidateProcess(ctx context.Context, process *bpmn.Process) (*BPMNValidationResult, error)

	// CheckReferences validates all references in the process
	CheckReferences(process *bpmn.Process) []BPMNValidationError

	// CheckCycles detects cycles in the process flow
	CheckCycles(process *bpmn.Process) (bool, []string)

	// CheckSupported checks for unsupported elements
	CheckSupported(process *bpmn.Process) ([]BPMNElementRef, []BPMNElementRef)
}

// ConversionValidator validates during conversion
type ConversionValidator interface {
	// ValidateDeterminism checks for non-deterministic patterns
	ValidateDeterminism(code string) []BPMNDeterminismViolation

	// ValidateTypeMapping validates type conversions
	ValidateTypeMapping(mapping ElementMapping) error

	// ValidateNaming checks for naming conflicts
	ValidateNaming(names []string) []NamingConflict

	// ValidateDataFlow validates data flow consistency
	ValidateDataFlow(mappings []ElementMapping) error
}

// PostConversionValidator validates generated code
type PostConversionValidator interface {
	// CompileCheck verifies code compiles
	CompileCheck(files []GeneratedFile) error

	// RunTemporalValidation runs Temporal validation
	RunTemporalValidation(workflowPath string) (*ValidationResult, error)

	// CheckDeterminism verifies workflow determinism
	CheckDeterminism(workflowPath string) (bool, []string)

	// FormatCode runs go fmt
	FormatCode(files []string) error
}

// BPMNDeterminismViolation represents a determinism violation in BPMN conversion
type BPMNDeterminismViolation struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Pattern    string `json:"pattern"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// NamingConflict represents a naming conflict
type NamingConflict struct {
	Original   string `json:"original"`
	Sanitized  string `json:"sanitized"`
	Conflicts  []string `json:"conflicts"`
	Resolution string `json:"resolution"`
}

// ValidationConstraints defines validation constraints
type ValidationConstraints struct {
	// Maximum workflow history size (events)
	MaxHistoryEvents int `json:"maxHistoryEvents"`

	// Maximum workflow history size (bytes)
	MaxHistoryBytes int64 `json:"maxHistoryBytes"`

	// Maximum workflow duration
	MaxWorkflowDuration string `json:"maxWorkflowDuration"`

	// Maximum activity duration
	MaxActivityDuration string `json:"maxActivityDuration"`

	// Maximum number of activities
	MaxActivities int `json:"maxActivities"`

	// Maximum parallel activities
	MaxParallelActivities int `json:"maxParallelActivities"`

	// Maximum child workflows
	MaxChildWorkflows int `json:"maxChildWorkflows"`

	// Maximum signals
	MaxSignals int `json:"maxSignals"`

	// Maximum timer duration
	MaxTimerDuration string `json:"maxTimerDuration"`

	// Maximum retry attempts
	MaxRetryAttempts int32 `json:"maxRetryAttempts"`
}

// DefaultValidationConstraints returns default Temporal constraints
func DefaultValidationConstraints() ValidationConstraints {
	return ValidationConstraints{
		MaxHistoryEvents:      50000,
		MaxHistoryBytes:       50 * 1024 * 1024, // 50MB
		MaxWorkflowDuration:   "720h",           // 30 days
		MaxActivityDuration:   "300s",           // 5 minutes default
		MaxActivities:         1000,
		MaxParallelActivities: 100,
		MaxChildWorkflows:     100,
		MaxSignals:            100,
		MaxTimerDuration:      "720h", // 30 days
		MaxRetryAttempts:      100,
	}
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Severity    Severity `json:"severity"`
	Enabled     bool   `json:"enabled"`
	Check       func(interface{}) error `json:"-"`
}

// ValidationRuleSet contains all validation rules
type ValidationRuleSet struct {
	PreConversionRules  []ValidationRule `json:"preConversionRules"`
	ConversionRules     []ValidationRule `json:"conversionRules"`
	PostConversionRules []ValidationRule `json:"postConversionRules"`
}

// CreateDefaultRuleSet creates the default validation rule set
func CreateDefaultRuleSet() *ValidationRuleSet {
	return &ValidationRuleSet{
		PreConversionRules: []ValidationRule{
			{
				ID:          "PRE-001",
				Name:        "Valid BPMN Schema",
				Description: "BPMN must conform to BPMN 2.0 schema",
				Category:    "schema",
				Severity:    SeverityCritical,
				Enabled:     true,
			},
			{
				ID:          "PRE-002",
				Name:        "Reference Integrity",
				Description: "All references must resolve to existing elements",
				Category:    "integrity",
				Severity:    SeverityCritical,
				Enabled:     true,
			},
			{
				ID:          "PRE-003",
				Name:        "No Infinite Loops",
				Description: "Process must not contain infinite loops without exit",
				Category:    "structure",
				Severity:    SeverityHigh,
				Enabled:     true,
			},
		},
		ConversionRules: []ValidationRule{
			{
				ID:          "CONV-001",
				Name:        "Deterministic Code",
				Description: "Generated code must be deterministic",
				Category:    "determinism",
				Severity:    SeverityCritical,
				Enabled:     true,
			},
			{
				ID:          "CONV-002",
				Name:        "Type Safety",
				Description: "Type conversions must be safe",
				Category:    "types",
				Severity:    SeverityHigh,
				Enabled:     true,
			},
			{
				ID:          "CONV-003",
				Name:        "Valid Go Identifiers",
				Description: "Generated names must be valid Go identifiers",
				Category:    "naming",
				Severity:    SeverityHigh,
				Enabled:     true,
			},
		},
		PostConversionRules: []ValidationRule{
			{
				ID:          "POST-001",
				Name:        "Compilation Success",
				Description: "Generated code must compile",
				Category:    "compilation",
				Severity:    SeverityCritical,
				Enabled:     true,
			},
			{
				ID:          "POST-002",
				Name:        "Temporal Validation",
				Description: "Code must pass Temporal validation",
				Category:    "temporal",
				Severity:    SeverityCritical,
				Enabled:     true,
			},
			{
				ID:          "POST-003",
				Name:        "Code Formatting",
				Description: "Code must be properly formatted",
				Category:    "style",
				Severity:    SeverityLow,
				Enabled:     true,
			},
		},
	}
}

// ValidateGoIdentifier validates that a string is a valid Go identifier
func ValidateGoIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Check if it's a Go keyword
	if isGoKeyword(name) {
		return fmt.Errorf("'%s' is a Go keyword", name)
	}

	// Go identifier regex
	validIdentifier := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validIdentifier.MatchString(name) {
		return fmt.Errorf("'%s' is not a valid Go identifier", name)
	}

	return nil
}

// SanitizeGoIdentifier converts a string to a valid Go identifier
func SanitizeGoIdentifier(name string) string {
	if name == "" {
		return "Unknown"
	}

	// Replace invalid characters with underscores
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "_")

	// Ensure it doesn't start with a number
	if regexp.MustCompile(`^[0-9]`).MatchString(sanitized) {
		sanitized = "ID" + sanitized
	}

	// Handle Go keywords
	if isGoKeyword(sanitized) {
		sanitized = sanitized + "_"
	}

	// Convert to PascalCase if snake_case
	if strings.Contains(sanitized, "_") {
		parts := strings.Split(sanitized, "_")
		for i, part := range parts {
			if part != "" {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
		sanitized = strings.Join(parts, "")
	}

	return sanitized
}

// isGoKeyword checks if a string is a Go keyword
func isGoKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "default": true, "func": true, "interface": true,
		"select": true, "case": true, "defer": true, "go": true,
		"map": true, "struct": true, "chan": true, "else": true,
		"goto": true, "package": true, "switch": true, "const": true,
		"fallthrough": true, "if": true, "range": true, "type": true,
		"continue": true, "for": true, "import": true, "return": true,
		"var": true,
	}
	return keywords[s]
}

// ValidateTemporalLimits validates against Temporal constraints
func ValidateTemporalLimits(stats ConversionStats, constraints ValidationConstraints) []BPMNValidationError {
	var errors []BPMNValidationError

	if stats.GeneratedActivities > constraints.MaxActivities {
		errors = append(errors, BPMNValidationError{
			Code:    ErrComplexityLimit,
			Message: fmt.Sprintf("Too many activities: %d > %d", stats.GeneratedActivities, constraints.MaxActivities),
			Fatal:   true,
		})
	}

	// Add more constraint checks as needed

	return errors
}

// ValidateSemanticPreservation validates that BPMN semantics are preserved
func ValidateSemanticPreservation(original *bpmn.Process, mappings []ElementMapping) error {
	// Count original elements
	originalCount := countBPMNElements(original)

	// Count mapped elements
	mappedCount := 0
	for _, mapping := range mappings {
		if mapping.Status == MappingComplete || mapping.Status == MappingPartial {
			mappedCount++
		}
	}

	// Check if all critical elements are mapped
	if mappedCount < originalCount {
		unmapped := originalCount - mappedCount
		if unmapped > 0 {
			// Check if unmapped elements are critical
			for _, mapping := range mappings {
				if mapping.Status == MappingFailed {
					for _, issue := range mapping.Issues {
						if issue.Severity == SeverityCritical {
							return fmt.Errorf("critical element %s failed to map: %s",
								mapping.BPMNElement.ID, issue.Message)
						}
					}
				}
			}
		}
	}

	return nil
}

// countBPMNElements counts total elements in a BPMN process
func countBPMNElements(process *bpmn.Process) int {
	count := 0
	if process.ProcessInfo.Elements.Events != nil {
		count += len(process.ProcessInfo.Elements.Events)
	}
	if process.ProcessInfo.Elements.Activities != nil {
		count += len(process.ProcessInfo.Elements.Activities)
	}
	if process.ProcessInfo.Elements.Gateways != nil {
		count += len(process.ProcessInfo.Elements.Gateways)
	}
	if process.ProcessInfo.Elements.SequenceFlows != nil {
		count += len(process.ProcessInfo.Elements.SequenceFlows)
	}
	return count
}