// Package schemas provides type definitions for BPMN to Temporal conversion
package schemas

import (
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
)

// ConversionConfig defines configuration for BPMN to Temporal conversion
type ConversionConfig struct {
	// Source BPMN file path
	SourceFile string `json:"sourceFile" validate:"required,file"`

	// Target package name for generated Go code
	PackageName string `json:"packageName" validate:"required,alphanum"`

	// Output directory for generated files
	OutputDir string `json:"outputDir" validate:"required"`

	// Conversion options
	Options ConversionOptions `json:"options" validate:"required"`

	// Validation level
	ValidationLevel string `json:"validationLevel" validate:"required,oneof=strict lenient"`

	// Allow partial conversion with TODOs
	AllowPartial bool `json:"allowPartial"`
}

// ConversionOptions contains options for the conversion process
type ConversionOptions struct {
	// Generate unit tests
	GenerateTests bool `json:"generateTests"`

	// Inline simple scripts
	InlineScripts bool `json:"inlineScripts"`

	// Max inline script lines
	MaxInlineLines int `json:"maxInlineLines" validate:"min=10,max=100"`

	// Generate TODOs for unsupported features
	GenerateTODOs bool `json:"generateTODOs"`

	// Preserve BPMN documentation
	PreserveDocs bool `json:"preserveDocs"`

	// Namespace for generated types
	TypeNamespace string `json:"typeNamespace"`

	// Strict mode - fail on any unsupported element
	StrictMode bool `json:"strictMode"`

	// Enable incremental migration
	IncrementalMigration bool `json:"incrementalMigration"`

	// Skip cache during validation
	SkipCache bool `json:"skipCache"`

	// Timeout for conversion process
	Timeout time.Duration `json:"timeout"`
}

// ElementMapping represents a BPMN element to Temporal mapping
type ElementMapping struct {
	// BPMN element reference
	BPMNElement BPMNElementRef `json:"bpmnElement" validate:"required"`

	// Target Temporal construct
	TemporalConstruct TemporalConstructType `json:"temporalConstruct" validate:"required"`

	// Mapping status
	Status MappingStatus `json:"status" validate:"required"`

	// Generated code location
	CodeLocation *GeneratedCodeLocation `json:"codeLocation,omitempty"`

	// Generated code snippet
	GeneratedCode string `json:"generatedCode,omitempty"`

	// Warnings/errors during mapping
	Issues []MappingIssue `json:"issues,omitempty"`
}

// BPMNElementRef references a BPMN element
type BPMNElementRef struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required,oneof=event activity gateway sequenceFlow association artifact dataObject property"`
	Name string `json:"name,omitempty"`
}

// TemporalConstructType represents types of Temporal constructs
type TemporalConstructType string

const (
	TemporalActivity      TemporalConstructType = "activity"
	TemporalWorkflow      TemporalConstructType = "workflow"
	TemporalSignal        TemporalConstructType = "signal"
	TemporalTimer         TemporalConstructType = "timer"
	TemporalSelector      TemporalConstructType = "selector"
	TemporalChildWorkflow TemporalConstructType = "childWorkflow"
	TemporalQuery         TemporalConstructType = "query"
	TemporalUpdate        TemporalConstructType = "update"
)

// MappingStatus represents the status of element mapping
type MappingStatus string

const (
	MappingPending    MappingStatus = "pending"
	MappingInProgress MappingStatus = "in_progress"
	MappingComplete   MappingStatus = "complete"
	MappingFailed     MappingStatus = "failed"
	MappingPartial    MappingStatus = "partial"
	MappingSkipped    MappingStatus = "skipped"
)

// GeneratedCodeLocation specifies where generated code is located
type GeneratedCodeLocation struct {
	File       string `json:"file"`
	LineStart  int    `json:"lineStart"`
	LineEnd    int    `json:"lineEnd"`
	Function   string `json:"function,omitempty"`
	StructName string `json:"structName,omitempty"`
}

// MappingIssue represents an issue during mapping
type MappingIssue struct {
	Type        IssueType `json:"type"`
	Severity    Severity  `json:"severity"`
	Message     string    `json:"message"`
	Suggestion  string    `json:"suggestion,omitempty"`
	ElementPath string    `json:"elementPath,omitempty"`
}

// IssueType categorizes mapping issues
type IssueType string

const (
	IssueUnsupported    IssueType = "unsupported"
	IssueNonDeterministic IssueType = "non_deterministic"
	IssueTypeMismatch   IssueType = "type_mismatch"
	IssueInvalidReference IssueType = "invalid_reference"
	IssueNameConflict   IssueType = "name_conflict"
	IssueComplexity     IssueType = "complexity"
)

// Severity levels for issues
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// ConversionResult represents the output of BPMN to Temporal conversion
type ConversionResult struct {
	// Conversion metadata
	Metadata ConversionMetadata `json:"metadata" validate:"required"`

	// Generated files
	GeneratedFiles []GeneratedFile `json:"generatedFiles" validate:"required,min=1"`

	// Element mappings
	Mappings []ElementMapping `json:"mappings" validate:"required"`

	// Validation results
	ValidationResult BPMNValidationResult `json:"validationResult" validate:"required"`

	// Conversion issues
	Issues []ConversionIssue `json:"issues,omitempty"`

	// Success indicator
	Success bool `json:"success"`
}

// ConversionMetadata contains metadata about the conversion
type ConversionMetadata struct {
	// Conversion timestamp
	Timestamp time.Time `json:"timestamp" validate:"required"`

	// Source BPMN file
	SourceFile string `json:"sourceFile" validate:"required"`

	// BPMN process ID
	ProcessID string `json:"processId" validate:"required"`

	// Process name
	ProcessName string `json:"processName,omitempty"`

	// Converter version
	ConverterVersion string `json:"converterVersion" validate:"required"`

	// Configuration used
	Config ConversionConfig `json:"config"`

	// Statistics
	Stats ConversionStats `json:"stats" validate:"required"`

	// Duration of conversion
	Duration time.Duration `json:"duration"`
}

// ConversionStats provides statistics about the conversion
type ConversionStats struct {
	TotalElements       int `json:"totalElements"`
	ConvertedElements   int `json:"convertedElements"`
	PartialElements     int `json:"partialElements"`
	FailedElements      int `json:"failedElements"`
	SkippedElements     int `json:"skippedElements"`
	GeneratedWorkflows  int `json:"generatedWorkflows"`
	GeneratedActivities int `json:"generatedActivities"`
	GeneratedTypes      int `json:"generatedTypes"`
	GeneratedTests      int `json:"generatedTests"`
	TODOsGenerated      int `json:"todosGenerated"`
	LinesOfCode         int `json:"linesOfCode"`
}

// GeneratedFile represents a generated Go source file
type GeneratedFile struct {
	// File path relative to output directory
	Path string `json:"path" validate:"required"`

	// File content
	Content string `json:"content" validate:"required"`

	// File type
	Type FileType `json:"type" validate:"required"`

	// Line count
	Lines int `json:"lines"`

	// Whether file is executable (has main function)
	Executable bool `json:"executable"`
}

// FileType categorizes generated files
type FileType string

const (
	FileTypeWorkflow   FileType = "workflow"
	FileTypeActivities FileType = "activities"
	FileTypeTypes      FileType = "types"
	FileTypeTests      FileType = "tests"
	FileTypeMapper     FileType = "mapper"
	FileTypeConfig     FileType = "config"
)

// BPMNValidationResult contains BPMN validation results
type BPMNValidationResult struct {
	// Whether process is valid for conversion
	Valid bool `json:"valid"`

	// Validation errors found
	Errors []BPMNValidationError `json:"errors,omitempty"`

	// Validation warnings
	Warnings []ValidationWarning `json:"warnings,omitempty"`

	// Count of supported elements
	SupportedElements int `json:"supportedElements"`

	// Count of unsupported elements
	UnsupportedElements int `json:"unsupportedElements"`

	// Detected patterns
	DetectedPatterns []string `json:"detectedPatterns,omitempty"`

	// Complexity score
	ComplexityScore int `json:"complexityScore"`
}

// BPMNValidationError represents a validation error for BPMN conversion
type BPMNValidationError struct {
	Code      ErrorCode `json:"code" validate:"required"`
	Message   string    `json:"message" validate:"required"`
	Element   *BPMNElementRef `json:"element,omitempty"`
	Suggestion string   `json:"suggestion,omitempty"`
	Fatal     bool      `json:"fatal"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    WarningCode `json:"code"`
	Message string      `json:"message"`
	Element *BPMNElementRef `json:"element,omitempty"`
	Impact  string      `json:"impact,omitempty"`
}

// ErrorCode categorizes errors
type ErrorCode string

const (
	ErrUnsupportedElement ErrorCode = "UNSUPPORTED_ELEMENT"
	ErrInvalidReference   ErrorCode = "INVALID_REFERENCE"
	ErrNonDeterministic   ErrorCode = "NON_DETERMINISTIC"
	ErrTypeMismatch       ErrorCode = "TYPE_MISMATCH"
	ErrCyclicDependency   ErrorCode = "CYCLIC_DEPENDENCY"
	ErrValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrNameConflict       ErrorCode = "NAME_CONFLICT"
	ErrComplexityLimit    ErrorCode = "COMPLEXITY_LIMIT"
	ErrInvalidStructure   ErrorCode = "INVALID_STRUCTURE"
)

// WarningCode categorizes warnings
type WarningCode string

const (
	WarnPartialSupport   WarningCode = "PARTIAL_SUPPORT"
	WarnPerformance      WarningCode = "PERFORMANCE"
	WarnDeprecated       WarningCode = "DEPRECATED"
	WarnManualIntervention WarningCode = "MANUAL_INTERVENTION"
	WarnNamingConvention WarningCode = "NAMING_CONVENTION"
	WarnComplexity       WarningCode = "COMPLEXITY"
)

// ConversionIssue represents an issue during conversion
type ConversionIssue struct {
	Stage       ConversionStage `json:"stage"`
	Type        IssueType       `json:"type"`
	Severity    Severity        `json:"severity"`
	Message     string          `json:"message"`
	Details     string          `json:"details,omitempty"`
	Element     *BPMNElementRef `json:"element,omitempty"`
	Suggestion  string          `json:"suggestion,omitempty"`
	Recoverable bool            `json:"recoverable"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ConversionStage represents stages of conversion
type ConversionStage string

const (
	StageLoading     ConversionStage = "loading"
	StageValidation  ConversionStage = "validation"
	StageAnalysis    ConversionStage = "analysis"
	StageMapping     ConversionStage = "mapping"
	StageGeneration  ConversionStage = "generation"
	StageVerification ConversionStage = "verification"
	StageFinalization ConversionStage = "finalization"
)

// ActivityMapping represents a mapped BPMN activity
type ActivityMapping struct {
	BPMNActivity bpmn.Activity `json:"bpmnActivity"`
	Name         string        `json:"name"`
	FunctionName string        `json:"functionName"`
	InputType    string        `json:"inputType"`
	OutputType   string        `json:"outputType"`
	Options      ActivityOptions `json:"options"`
	IsHumanTask  bool          `json:"isHumanTask"`
	HasSignal    bool          `json:"hasSignal"`
	SignalName   string        `json:"signalName,omitempty"`
	TimeoutDuration time.Duration `json:"timeoutDuration,omitempty"`
}

// ActivityOptions contains Temporal activity options
type ActivityOptions struct {
	TaskQueue            string        `json:"taskQueue"`
	ScheduleToCloseTimeout time.Duration `json:"scheduleToCloseTimeout"`
	StartToCloseTimeout    time.Duration `json:"startToCloseTimeout"`
	HeartbeatTimeout       time.Duration `json:"heartbeatTimeout,omitempty"`
	RetryPolicy           *RetryPolicy  `json:"retryPolicy,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	InitialInterval    time.Duration `json:"initialInterval"`
	BackoffCoefficient float64       `json:"backoffCoefficient"`
	MaximumInterval    time.Duration `json:"maximumInterval"`
	MaximumAttempts    int32         `json:"maximumAttempts"`
	NonRetryableErrors []string      `json:"nonRetryableErrors,omitempty"`
}

// GatewayMapping represents a mapped BPMN gateway
type GatewayMapping struct {
	BPMNGateway  bpmn.Gateway `json:"bpmnGateway"`
	Type         string       `json:"type"`
	ControlFlow  string       `json:"controlFlow"`
	Conditions   []Condition  `json:"conditions,omitempty"`
	DefaultPath  string       `json:"defaultPath,omitempty"`
	ParallelCount int         `json:"parallelCount,omitempty"`
}

// Condition represents a gateway condition
type Condition struct {
	Expression string `json:"expression"`
	TargetPath string `json:"targetPath"`
	Variables  []string `json:"variables,omitempty"`
}

// EventMapping represents a mapped BPMN event
type EventMapping struct {
	BPMNEvent   bpmn.Event `json:"bpmnEvent"`
	HandlerType string     `json:"handlerType"`
	SignalName  string     `json:"signalName,omitempty"`
	TimerDuration time.Duration `json:"timerDuration,omitempty"`
	ErrorType   string     `json:"errorType,omitempty"`
	IsStart     bool       `json:"isStart"`
	IsEnd       bool       `json:"isEnd"`
	IsBoundary  bool       `json:"isBoundary"`
	IsInterrupting bool    `json:"isInterrupting"`
}

// GenerationConfig defines code generation configuration
type GenerationConfig struct {
	PackageName    string `json:"packageName"`
	OutputDir      string `json:"outputDir"`
	GenerateTests  bool   `json:"generateTests"`
	PreserveDocs   bool   `json:"preserveDocs"`
	GenerateTODOs  bool   `json:"generateTODOs"`
	TypeNamespace  string `json:"typeNamespace"`
}

// GeneratedCode represents the result of code generation
type GeneratedCode struct {
	Files      []GeneratedFile `json:"files"`
	TotalLines int             `json:"totalLines"`
	TodoCount  int             `json:"todoCount"`
	Metadata   CodeMetadata    `json:"metadata"`
}

// CodeMetadata contains metadata about generated code
type CodeMetadata struct {
	GeneratedAt   time.Time `json:"generatedAt"`
	ConverterVersion string `json:"converterVersion"`
	SourceProcess string    `json:"sourceProcess"`
}