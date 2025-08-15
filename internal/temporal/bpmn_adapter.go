// Package temporal provides BPMN to Temporal workflow conversion capabilities
package temporal

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// BPMNAdapter handles conversion of BPMN processes to Temporal workflows
type BPMNAdapter struct {
	mapper    *ElementMapper
	generator *CodeGenerator
	validator ValidatorInterface
}

// NewBPMNAdapter creates a new BPMN adapter instance
func NewBPMNAdapter() *BPMNAdapter {
	return &BPMNAdapter{
		mapper:    NewElementMapper(),
		generator: NewCodeGenerator(),
		validator: NewTemporalValidator(),
	}
}

// Convert performs the main conversion from BPMN to Temporal
func (a *BPMNAdapter) Convert(ctx context.Context, process *bpmn.Process, config schemas.ConversionConfig) (*schemas.ConversionResult, error) {
	startTime := time.Now()
	
	// Initialize result
	result := &schemas.ConversionResult{
		Metadata: schemas.ConversionMetadata{
			Timestamp:        startTime,
			SourceFile:       config.SourceFile,
			ProcessID:        process.ProcessInfo.ID,
			ProcessName:      process.ProcessInfo.Name,
			ConverterVersion: "1.0.0",
			Config:           config,
			Stats:            schemas.ConversionStats{},
		},
		GeneratedFiles: []schemas.GeneratedFile{},
		Mappings:       []schemas.ElementMapping{},
		Issues:         []schemas.ConversionIssue{},
		Success:        false,
	}

	// Stage 1: Validation
	validationResult, err := a.ValidateProcess(&process.ProcessInfo)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	result.ValidationResult = *validationResult
	
	// Check if we should proceed
	if !validationResult.Valid && config.Options.StrictMode {
		result.Issues = append(result.Issues, schemas.ConversionIssue{
			Stage:    schemas.StageValidation,
			Type:     schemas.IssueUnsupported,
			Severity: schemas.SeverityCritical,
			Message:  "Process validation failed in strict mode",
			Details:  fmt.Sprintf("Found %d errors", len(validationResult.Errors)),
			Timestamp: time.Now(),
		})
		return result, nil
	}

	// Stage 2: Mapping
	mappings, err := a.mapBPMNElements(ctx, process, config)
	if err != nil {
		return nil, fmt.Errorf("mapping failed: %w", err)
	}
	
	result.Mappings = mappings
	a.updateStats(&result.Metadata.Stats, mappings)

	// Stage 3: Code Generation
	generationConfig := schemas.GenerationConfig{
		PackageName:    config.PackageName,
		OutputDir:      config.OutputDir,
		GenerateTests:  config.Options.GenerateTests,
		PreserveDocs:   config.Options.PreserveDocs,
		GenerateTODOs:  config.Options.GenerateTODOs,
		TypeNamespace:  config.Options.TypeNamespace,
	}
	
	generatedCode, err := a.GenerateCode(mappings, generationConfig)
	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}
	
	result.GeneratedFiles = generatedCode.Files
	result.Metadata.Stats.LinesOfCode = generatedCode.TotalLines
	result.Metadata.Stats.TODOsGenerated = generatedCode.TodoCount
	if config.Options.GenerateTests {
		result.Metadata.Stats.GeneratedTests = 1
	}

	// Stage 4: Verification
	if !config.Options.SkipCache {
		for _, file := range result.GeneratedFiles {
			if file.Type == schemas.FileTypeWorkflow {
				// Create temporary file for validation
				tempPath := filepath.Join("/tmp", filepath.Base(file.Path))
				if err := a.validateGeneratedCode(ctx, tempPath, file.Content); err != nil {
					result.Issues = append(result.Issues, schemas.ConversionIssue{
						Stage:    schemas.StageVerification,
						Type:     schemas.IssueNonDeterministic,
						Severity: schemas.SeverityHigh,
						Message:  "Generated code validation failed",
						Details:  err.Error(),
						Timestamp: time.Now(),
					})
				}
			}
		}
	}

	// Set success flag
	result.Success = len(result.Issues) == 0 || 
		(config.AllowPartial && !hasBlockingIssues(result.Issues))
	
	result.Metadata.Duration = time.Since(startTime)
	
	return result, nil
}

// ValidateProcess validates a BPMN process for conversion compatibility
func (a *BPMNAdapter) ValidateProcess(process *bpmn.ProcessInfo) (*schemas.BPMNValidationResult, error) {
	result := &schemas.BPMNValidationResult{
		Valid:               true,
		Errors:              []schemas.BPMNValidationError{},
		Warnings:            []schemas.ValidationWarning{},
		SupportedElements:   0,
		UnsupportedElements: 0,
		DetectedPatterns:    []string{},
		ComplexityScore:     0,
	}

	// Count and validate elements
	elementCount := 0
	
	// Validate events
	for _, event := range process.Elements.Events {
		elementCount++
		if err := a.validateEvent(&event, result); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, schemas.BPMNValidationError{
				Code:    schemas.ErrUnsupportedElement,
				Message: fmt.Sprintf("Event %s: %v", event.ID, err),
				Element: &schemas.BPMNElementRef{
					ID:   event.ID,
					Type: "event",
					Name: event.Name,
				},
				Fatal: false,
			})
			result.UnsupportedElements++
		} else {
			result.SupportedElements++
		}
	}

	// Validate activities
	for _, activity := range process.Elements.Activities {
		elementCount++
		if err := a.validateActivity(&activity, result); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, schemas.BPMNValidationError{
				Code:    schemas.ErrUnsupportedElement,
				Message: fmt.Sprintf("Activity %s: %v", activity.ID, err),
				Element: &schemas.BPMNElementRef{
					ID:   activity.ID,
					Type: "activity",
					Name: activity.Name,
				},
				Fatal: false,
			})
			result.UnsupportedElements++
		} else {
			result.SupportedElements++
		}
	}

	// Validate gateways
	for _, gateway := range process.Elements.Gateways {
		elementCount++
		if err := a.validateGateway(&gateway, result); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, schemas.BPMNValidationError{
				Code:    schemas.ErrUnsupportedElement,
				Message: fmt.Sprintf("Gateway %s: %v", gateway.ID, err),
				Element: &schemas.BPMNElementRef{
					ID:   gateway.ID,
					Type: "gateway",
					Name: gateway.Name,
				},
				Fatal: false,
			})
			result.UnsupportedElements++
		} else {
			result.SupportedElements++
		}
	}

	// Detect patterns
	result.DetectedPatterns = a.detectPatterns(process)
	
	// Calculate complexity score
	result.ComplexityScore = a.calculateComplexity(process)

	// Check for fatal errors
	for _, err := range result.Errors {
		if err.Fatal {
			result.Valid = false
			break
		}
	}

	return result, nil
}

// GenerateCode generates Go code from element mappings
func (a *BPMNAdapter) GenerateCode(mappings []schemas.ElementMapping, config schemas.GenerationConfig) (*schemas.GeneratedCode, error) {
	result := &schemas.GeneratedCode{
		Files:      []schemas.GeneratedFile{},
		TotalLines: 0,
		TodoCount:  0,
	}

	// Separate mappings by type
	var activities []schemas.ActivityMapping
	workflowMappings := []schemas.ElementMapping{}

	// For now, create dummy activity mappings from element types
	// In production, these would be extracted from the actual mappings
	for _, mapping := range mappings {
		if mapping.BPMNElement.Type == "activity" {
			activities = append(activities, schemas.ActivityMapping{
				Name:         mapping.BPMNElement.Name,
				FunctionName: sanitizeName(mapping.BPMNElement.Name),
				InputType:    "ActivityInput",
				OutputType:   "ActivityOutput",
				Options: schemas.ActivityOptions{
					TaskQueue:              "default",
					ScheduleToCloseTimeout: 10 * time.Minute,
					StartToCloseTimeout:    5 * time.Minute,
				},
			})
		}
		workflowMappings = append(workflowMappings, mapping)
	}

	// Generate workflow file
	workflowCode, err := a.generator.GenerateWorkflow(config.PackageName, workflowMappings)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow: %w", err)
	}
	
	workflowFile := schemas.GeneratedFile{
		Path:       fmt.Sprintf("%s/workflow.go", config.OutputDir),
		Content:    workflowCode,
		Type:       schemas.FileTypeWorkflow,
		Lines:      countLines(workflowCode),
		Executable: false,
	}
	result.Files = append(result.Files, workflowFile)
	result.TotalLines += workflowFile.Lines

	// Generate activities file
	if len(activities) > 0 {
		activitiesCode, err := a.generator.GenerateActivities(activities)
		if err != nil {
			return nil, fmt.Errorf("failed to generate activities: %w", err)
		}
		
		activitiesFile := schemas.GeneratedFile{
			Path:    fmt.Sprintf("%s/activities.go", config.OutputDir),
			Content: activitiesCode,
			Type:    schemas.FileTypeActivities,
			Lines:   countLines(activitiesCode),
		}
		result.Files = append(result.Files, activitiesFile)
		result.TotalLines += activitiesFile.Lines
	}

	// Generate types file
	typesCode, err := a.generator.GenerateTypes(nil, nil) // Extract from mappings
	if err != nil {
		return nil, fmt.Errorf("failed to generate types: %w", err)
	}
	
	typesFile := schemas.GeneratedFile{
		Path:    fmt.Sprintf("%s/types.go", config.OutputDir),
		Content: typesCode,
		Type:    schemas.FileTypeTypes,
		Lines:   countLines(typesCode),
	}
	result.Files = append(result.Files, typesFile)
	result.TotalLines += typesFile.Lines

	// Generate tests if requested
	if config.GenerateTests {
		testCode, err := a.generator.GenerateTests(config.PackageName, mappings)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tests: %w", err)
		}
		
		testFile := schemas.GeneratedFile{
			Path:    fmt.Sprintf("%s/workflow_test.go", config.OutputDir),
			Content: testCode,
			Type:    schemas.FileTypeTests,
			Lines:   countLines(testCode),
		}
		result.Files = append(result.Files, testFile)
		result.TotalLines += testFile.Lines
	}

	// Count TODOs
	for _, file := range result.Files {
		result.TodoCount += countTODOs(file.Content)
	}

	return result, nil
}

// Private helper methods

func (a *BPMNAdapter) mapBPMNElements(ctx context.Context, process *bpmn.Process, config schemas.ConversionConfig) ([]schemas.ElementMapping, error) {
	var mappings []schemas.ElementMapping
	
	// Map events
	for _, event := range process.ProcessInfo.Elements.Events {
		_, err := a.mapper.MapEvent(event)
		if err != nil {
			if config.Options.StrictMode {
				return nil, err
			}
			// Create partial mapping with TODO
		}
		
		mappings = append(mappings, schemas.ElementMapping{
			BPMNElement: schemas.BPMNElementRef{
				ID:   event.ID,
				Type: "event",
				Name: event.Name,
			},
			TemporalConstruct: a.determineEventConstruct(&event),
			Status:           schemas.MappingComplete,
		})
	}
	
	// Map activities
	for _, activity := range process.ProcessInfo.Elements.Activities {
		_, err := a.mapper.MapActivity(activity)
		if err != nil {
			if config.Options.StrictMode {
				return nil, err
			}
			// Create partial mapping with TODO
		}
		
		mappings = append(mappings, schemas.ElementMapping{
			BPMNElement: schemas.BPMNElementRef{
				ID:   activity.ID,
				Type: "activity",
				Name: activity.Name,
			},
			TemporalConstruct: schemas.TemporalActivity,
			Status:           schemas.MappingComplete,
		})
	}
	
	// Map gateways
	for _, gateway := range process.ProcessInfo.Elements.Gateways {
		_, err := a.mapper.MapGateway(gateway)
		if err != nil {
			if config.Options.StrictMode {
				return nil, err
			}
			// Create partial mapping with TODO
		}
		
		mappings = append(mappings, schemas.ElementMapping{
			BPMNElement: schemas.BPMNElementRef{
				ID:   gateway.ID,
				Type: "gateway",
				Name: gateway.Name,
			},
			TemporalConstruct: schemas.TemporalSelector,
			Status:           schemas.MappingComplete,
		})
	}
	
	return mappings, nil
}

func (a *BPMNAdapter) validateEvent(event *bpmn.Event, result *schemas.BPMNValidationResult) error {
	// Check supported event types
	switch event.Type {
	case "startEvent", "endEvent", "intermediateThrowEvent", "intermediateCatchEvent":
		// Supported
	case "boundaryEvent":
		if event.AttachedTo == "" {
			return fmt.Errorf("boundary event must have attachedTo reference")
		}
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
	
	// Check event sub-types
	switch event.EventType {
	case "none", "message", "timer", "signal", "error":
		// Supported
	case "compensation", "cancel", "conditional", "link", "terminate":
		result.Warnings = append(result.Warnings, schemas.ValidationWarning{
			Code:    schemas.WarnPartialSupport,
			Message: fmt.Sprintf("Event type '%s' has partial support", event.EventType),
			Element: &schemas.BPMNElementRef{ID: event.ID, Type: "event", Name: event.Name},
			Impact:  "May require manual intervention",
		})
	default:
		if event.EventType != "" {
			return fmt.Errorf("unsupported event type: %s", event.EventType)
		}
	}
	
	return nil
}

func (a *BPMNAdapter) validateActivity(activity *bpmn.Activity, result *schemas.BPMNValidationResult) error {
	// Check supported activity types
	switch activity.Type {
	case "task", "serviceTask", "userTask", "scriptTask":
		// Fully supported
	case "sendTask", "receiveTask":
		// Supported via signals
	case "manualTask", "businessRuleTask":
		result.Warnings = append(result.Warnings, schemas.ValidationWarning{
			Code:    schemas.WarnPartialSupport,
			Message: fmt.Sprintf("Activity type '%s' requires custom implementation", activity.Type),
			Element: &schemas.BPMNElementRef{ID: activity.ID, Type: "activity", Name: activity.Name},
			Impact:  "Will generate TODO for manual implementation",
		})
	case "callActivity", "subProcess":
		// Supported as child workflows
	default:
		return fmt.Errorf("unsupported activity type: %s", activity.Type)
	}
	
	// Check for complex features
	if activity.IsSequential && activity.CompletionQuantity > 1 {
		result.Warnings = append(result.Warnings, schemas.ValidationWarning{
			Code:    schemas.WarnComplexity,
			Message: "Multi-instance activities require special handling",
			Element: &schemas.BPMNElementRef{ID: activity.ID, Type: "activity", Name: activity.Name},
		})
	}
	
	return nil
}

func (a *BPMNAdapter) validateGateway(gateway *bpmn.Gateway, result *schemas.BPMNValidationResult) error {
	// Check supported gateway types
	switch gateway.Type {
	case "exclusiveGateway", "parallelGateway":
		// Fully supported
	case "inclusiveGateway":
		result.Warnings = append(result.Warnings, schemas.ValidationWarning{
			Code:    schemas.WarnPartialSupport,
			Message: "Inclusive gateway requires careful condition handling",
			Element: &schemas.BPMNElementRef{ID: gateway.ID, Type: "gateway", Name: gateway.Name},
		})
	case "eventBasedGateway":
		// Supported via selectors
	case "complexGateway":
		return fmt.Errorf("complex gateways are not supported")
	default:
		return fmt.Errorf("unsupported gateway type: %s", gateway.Type)
	}
	
	return nil
}

func (a *BPMNAdapter) detectPatterns(process *bpmn.ProcessInfo) []string {
	patterns := []string{}
	
	// Check for human task pattern
	hasHumanTasks := false
	for _, activity := range process.Elements.Activities {
		if activity.Type == "userTask" {
			hasHumanTasks = true
			break
		}
	}
	if hasHumanTasks {
		patterns = append(patterns, "human-task-workflow")
	}
	
	// Check for parallel execution pattern
	hasParallel := false
	for _, gateway := range process.Elements.Gateways {
		if gateway.Type == "parallelGateway" {
			hasParallel = true
			break
		}
	}
	if hasParallel {
		patterns = append(patterns, "parallel-execution")
	}
	
	// Check for event-driven pattern
	hasEvents := false
	for _, event := range process.Elements.Events {
		if event.EventType == "message" || event.EventType == "signal" {
			hasEvents = true
			break
		}
	}
	if hasEvents {
		patterns = append(patterns, "event-driven")
	}
	
	// Check for subprocess pattern
	hasSubprocess := false
	for _, activity := range process.Elements.Activities {
		if activity.Type == "subProcess" || activity.Type == "callActivity" {
			hasSubprocess = true
			break
		}
	}
	if hasSubprocess {
		patterns = append(patterns, "hierarchical-workflow")
	}
	
	return patterns
}

func (a *BPMNAdapter) calculateComplexity(process *bpmn.ProcessInfo) int {
	score := 0
	
	// Base complexity from element counts
	score += len(process.Elements.Activities) * 2
	score += len(process.Elements.Gateways) * 3
	score += len(process.Elements.Events)
	score += len(process.Elements.SequenceFlows)
	
	// Additional complexity for specific patterns
	for _, gateway := range process.Elements.Gateways {
		if gateway.Type == "complexGateway" {
			score += 5
		} else if gateway.Type == "inclusiveGateway" {
			score += 3
		}
	}
	
	for _, activity := range process.Elements.Activities {
		if activity.Type == "subProcess" {
			score += 5
		}
		if activity.IsSequential && activity.CompletionQuantity > 1 {
			score += 3
		}
	}
	
	return score
}

func (a *BPMNAdapter) determineEventConstruct(event *bpmn.Event) schemas.TemporalConstructType {
	switch event.EventType {
	case "timer":
		return schemas.TemporalTimer
	case "message", "signal":
		return schemas.TemporalSignal
	default:
		return schemas.TemporalWorkflow
	}
}

func (a *BPMNAdapter) updateStats(stats *schemas.ConversionStats, mappings []schemas.ElementMapping) {
	stats.TotalElements = len(mappings)
	
	for _, mapping := range mappings {
		switch mapping.Status {
		case schemas.MappingComplete:
			stats.ConvertedElements++
		case schemas.MappingPartial:
			stats.PartialElements++
		case schemas.MappingFailed:
			stats.FailedElements++
		case schemas.MappingSkipped:
			stats.SkippedElements++
		}
		
		switch mapping.TemporalConstruct {
		case schemas.TemporalWorkflow, schemas.TemporalChildWorkflow:
			stats.GeneratedWorkflows++
		case schemas.TemporalActivity:
			stats.GeneratedActivities++
		}
	}
}

func (a *BPMNAdapter) validateGeneratedCode(ctx context.Context, path string, content string) error {
	// Write to temporary file
	if err := writeFile(path, content); err != nil {
		return err
	}
	
	// Run validation
	request := schemas.ValidationRequest{
		WorkflowID:   "generated",
		WorkflowPath: path,
		Options: schemas.ValidationOptions{
			Timeout: 10 * time.Second,
		},
	}
	
	result, err := a.validator.Validate(ctx, request)
	if err != nil {
		return err
	}
	
	if !result.Success {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}
	
	return nil
}

func (a *BPMNAdapter) extractActivityMapping(mapping schemas.ElementMapping) *schemas.ActivityMapping {
	// Implementation would extract activity-specific mapping data
	return nil
}

func (a *BPMNAdapter) extractGatewayMapping(mapping schemas.ElementMapping) *schemas.GatewayMapping {
	// Implementation would extract gateway-specific mapping data
	return nil
}

func (a *BPMNAdapter) extractEventMapping(mapping schemas.ElementMapping) *schemas.EventMapping {
	// Implementation would extract event-specific mapping data
	return nil
}

func hasBlockingIssues(issues []schemas.ConversionIssue) bool {
	for _, issue := range issues {
		if issue.Severity == schemas.SeverityCritical {
			return true
		}
	}
	return false
}

func countLines(content string) int {
	lines := 1
	for _, ch := range content {
		if ch == '\n' {
			lines++
		}
	}
	return lines
}

func countTODOs(content string) int {
	count := 0
	lines := splitLines(content)
	for _, line := range lines {
		if containsTODO(line) {
			count++
		}
	}
	return count
}

func splitLines(content string) []string {
	// Simple line splitting implementation
	var lines []string
	current := ""
	for _, ch := range content {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func containsTODO(line string) bool {
	return len(line) > 0 && (containsString(line, "TODO") || containsString(line, "FIXME"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func writeFile(path, content string) error {
	// Implementation would write content to file
	return nil
}