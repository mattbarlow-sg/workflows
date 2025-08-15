// Package schemas provides transformation contracts for BPMN to Temporal conversion
package schemas

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
)

// Transformer defines the interface for BPMN to Temporal transformations
type Transformer interface {
	// TransformProcess transforms a complete BPMN process
	TransformProcess(process *bpmn.Process, config ConversionConfig) (*ConversionResult, error)

	// TransformActivity transforms a BPMN activity
	TransformActivity(activity bpmn.Activity) (*ActivityMapping, error)

	// TransformGateway transforms a BPMN gateway
	TransformGateway(gateway bpmn.Gateway) (*GatewayMapping, error)

	// TransformEvent transforms a BPMN event
	TransformEvent(event bpmn.Event) (*EventMapping, error)

	// TransformDataObject transforms a BPMN data object
	TransformDataObject(dataObject bpmn.DataObject) (string, error)
}

// ActivityTransformer handles activity transformations
type ActivityTransformer struct {
	Config ConversionConfig
}

// TransformUserTask transforms a user task to Temporal activity with signal
func (t *ActivityTransformer) TransformUserTask(activity bpmn.Activity) (*ActivityMapping, error) {
	mapping := &ActivityMapping{
		BPMNActivity: activity,
		Name:         SanitizeGoIdentifier(activity.Name),
		FunctionName: fmt.Sprintf("%sActivity", SanitizeGoIdentifier(activity.Name)),
		IsHumanTask:  true,
		HasSignal:    true,
		SignalName:   fmt.Sprintf("%sSignal", SanitizeGoIdentifier(activity.Name)),
	}

	// Set default timeout for human tasks
	mapping.TimeoutDuration = 24 * time.Hour

	// Configure activity options
	mapping.Options = ActivityOptions{
		TaskQueue:              "human-tasks",
		ScheduleToCloseTimeout: 48 * time.Hour,
		StartToCloseTimeout:    24 * time.Hour,
		HeartbeatTimeout:       1 * time.Minute,
		RetryPolicy: &RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}

	// Process agent assignment if present
	if activity.Agent != nil {
		mapping.Options.TaskQueue = fmt.Sprintf("%s-tasks", activity.Agent.Type)
	}

	return mapping, nil
}

// TransformServiceTask transforms a service task to standard Temporal activity
func (t *ActivityTransformer) TransformServiceTask(activity bpmn.Activity) (*ActivityMapping, error) {
	mapping := &ActivityMapping{
		BPMNActivity: activity,
		Name:         SanitizeGoIdentifier(activity.Name),
		FunctionName: fmt.Sprintf("%sActivity", SanitizeGoIdentifier(activity.Name)),
		IsHumanTask:  false,
		HasSignal:    false,
	}

	// Configure activity options
	mapping.Options = ActivityOptions{
		TaskQueue:              "default",
		ScheduleToCloseTimeout: 5 * time.Minute,
		StartToCloseTimeout:    1 * time.Minute,
		RetryPolicy: &RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    10,
		},
	}

	return mapping, nil
}

// TransformScriptTask transforms a script task (inline or activity)
func (t *ActivityTransformer) TransformScriptTask(activity bpmn.Activity) (*ActivityMapping, error) {
	if activity.Script == nil {
		return nil, fmt.Errorf("script task %s has no script", activity.ID)
	}

	// Count lines in script
	lines := strings.Count(activity.Script.Body, "\n") + 1

	// Decide whether to inline
	if t.Config.Options.InlineScripts && lines <= t.Config.Options.MaxInlineLines {
		// Return special mapping for inline script
		return &ActivityMapping{
			BPMNActivity: activity,
			Name:         SanitizeGoIdentifier(activity.Name),
			FunctionName: "inline", // Special marker for inline code
		}, nil
	}

	// Create activity for complex script
	mapping := &ActivityMapping{
		BPMNActivity: activity,
		Name:         SanitizeGoIdentifier(activity.Name),
		FunctionName: fmt.Sprintf("%sScriptActivity", SanitizeGoIdentifier(activity.Name)),
		IsHumanTask:  false,
		HasSignal:    false,
	}

	mapping.Options = ActivityOptions{
		TaskQueue:              "scripts",
		ScheduleToCloseTimeout: 1 * time.Minute,
		StartToCloseTimeout:    30 * time.Second,
		RetryPolicy: &RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
		},
	}

	return mapping, nil
}

// GatewayTransformer handles gateway transformations
type GatewayTransformer struct {
	Config ConversionConfig
}

// TransformExclusiveGateway transforms exclusive gateway to if/else
func (t *GatewayTransformer) TransformExclusiveGateway(gateway bpmn.Gateway) (*GatewayMapping, error) {
	mapping := &GatewayMapping{
		BPMNGateway: gateway,
		Type:        "exclusive",
		ControlFlow: "if-else",
	}

	// Process outgoing flows for conditions
	for _, flowID := range gateway.Outgoing {
		// This would look up the actual flow and its condition
		condition := Condition{
			Expression: fmt.Sprintf("condition_%s", flowID),
			TargetPath: flowID,
		}
		mapping.Conditions = append(mapping.Conditions, condition)
	}

	// Set default path if specified
	if gateway.DefaultFlow != "" {
		mapping.DefaultPath = gateway.DefaultFlow
	}

	return mapping, nil
}

// TransformParallelGateway transforms parallel gateway to workflow.Go
func (t *GatewayTransformer) TransformParallelGateway(gateway bpmn.Gateway) (*GatewayMapping, error) {
	mapping := &GatewayMapping{
		BPMNGateway:   gateway,
		Type:          "parallel",
		ControlFlow:   "workflow.Go",
		ParallelCount: len(gateway.Outgoing),
	}

	return mapping, nil
}

// TransformInclusiveGateway transforms inclusive gateway to conditional workflow.Go
func (t *GatewayTransformer) TransformInclusiveGateway(gateway bpmn.Gateway) (*GatewayMapping, error) {
	mapping := &GatewayMapping{
		BPMNGateway: gateway,
		Type:        "inclusive",
		ControlFlow: "conditional-parallel",
	}

	// Process conditions for each path
	for _, flowID := range gateway.Outgoing {
		condition := Condition{
			Expression: fmt.Sprintf("condition_%s", flowID),
			TargetPath: flowID,
		}
		mapping.Conditions = append(mapping.Conditions, condition)
	}

	return mapping, nil
}

// TransformEventBasedGateway transforms event-based gateway to workflow.Selector
func (t *GatewayTransformer) TransformEventBasedGateway(gateway bpmn.Gateway) (*GatewayMapping, error) {
	mapping := &GatewayMapping{
		BPMNGateway: gateway,
		Type:        "event-based",
		ControlFlow: "workflow.Selector",
	}

	return mapping, nil
}

// EventTransformer handles event transformations
type EventTransformer struct {
	Config ConversionConfig
}

// TransformStartEvent transforms a start event
func (t *EventTransformer) TransformStartEvent(event bpmn.Event) (*EventMapping, error) {
	mapping := &EventMapping{
		BPMNEvent:   event,
		HandlerType: "workflow-start",
		IsStart:     true,
	}

	// Handle different start event types
	switch event.EventType {
	case "timer":
		mapping.HandlerType = "cron-trigger"
	case "message":
		mapping.HandlerType = "signal-trigger"
		mapping.SignalName = fmt.Sprintf("%sStartSignal", SanitizeGoIdentifier(event.Name))
	case "signal":
		mapping.HandlerType = "signal-trigger"
		mapping.SignalName = fmt.Sprintf("%sSignal", SanitizeGoIdentifier(event.Name))
	}

	return mapping, nil
}

// TransformEndEvent transforms an end event
func (t *EventTransformer) TransformEndEvent(event bpmn.Event) (*EventMapping, error) {
	mapping := &EventMapping{
		BPMNEvent:   event,
		HandlerType: "workflow-complete",
		IsEnd:       true,
	}

	// Handle different end event types
	switch event.EventType {
	case "error":
		mapping.HandlerType = "error-end"
		mapping.ErrorType = fmt.Sprintf("%sError", SanitizeGoIdentifier(event.Name))
	case "terminate":
		mapping.HandlerType = "terminate"
	case "compensation":
		mapping.HandlerType = "compensation-trigger"
	}

	return mapping, nil
}

// TransformTimerEvent transforms a timer event
func (t *EventTransformer) TransformTimerEvent(event bpmn.Event) (*EventMapping, error) {
	mapping := &EventMapping{
		BPMNEvent:   event,
		HandlerType: "timer",
	}

	// Parse timer duration from properties
	if duration, ok := event.Properties["duration"].(string); ok {
		if d, err := time.ParseDuration(duration); err == nil {
			mapping.TimerDuration = d
		}
	}

	// Default timer duration if not specified
	if mapping.TimerDuration == 0 {
		mapping.TimerDuration = 1 * time.Hour
	}

	return mapping, nil
}

// TransformBoundaryEvent transforms a boundary event
func (t *EventTransformer) TransformBoundaryEvent(event bpmn.Event) (*EventMapping, error) {
	mapping := &EventMapping{
		BPMNEvent:      event,
		HandlerType:    "boundary",
		IsBoundary:     true,
		IsInterrupting: event.IsInterrupt,
	}

	// Handle different boundary event types
	switch event.EventType {
	case "timer":
		mapping.HandlerType = "timeout"
		if duration, ok := event.Properties["timeout"].(string); ok {
			if d, err := time.ParseDuration(duration); err == nil {
				mapping.TimerDuration = d
			}
		}
	case "error":
		mapping.HandlerType = "error-boundary"
		mapping.ErrorType = fmt.Sprintf("%sError", SanitizeGoIdentifier(event.Name))
	case "signal":
		mapping.HandlerType = "signal-boundary"
		mapping.SignalName = fmt.Sprintf("%sSignal", SanitizeGoIdentifier(event.Name))
	}

	return mapping, nil
}

// DataTransformer handles data transformations
type DataTransformer struct {
	Config ConversionConfig
}

// TransformDataObject transforms a BPMN data object to Go struct
func (t *DataTransformer) TransformDataObject(dataObject bpmn.DataObject) (string, error) {
	structName := SanitizeGoIdentifier(dataObject.Name)
	
	// Build struct definition
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// %s represents %s\n", structName, dataObject.Name))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	
	// Add ID field
	sb.WriteString(fmt.Sprintf("\tID string `json:\"id\"`\n"))
	
	// Add state field if present
	if dataObject.State != "" {
		sb.WriteString(fmt.Sprintf("\tState string `json:\"state\"`\n"))
	}
	
	// Add collection field if needed
	if dataObject.IsCollection {
		sb.WriteString(fmt.Sprintf("\tItems []interface{} `json:\"items\"`\n"))
	}
	
	sb.WriteString("}\n")
	
	return sb.String(), nil
}

// TransformProperty transforms a BPMN property to Go field
func (t *DataTransformer) TransformProperty(property bpmn.Property) (string, error) {
	fieldName := SanitizeGoIdentifier(property.Name)
	fieldType := "interface{}" // Default type
	
	// Try to infer type from ItemSubject
	if property.ItemSubject != "" {
		switch strings.ToLower(property.ItemSubject) {
		case "string":
			fieldType = "string"
		case "int", "integer":
			fieldType = "int"
		case "bool", "boolean":
			fieldType = "bool"
		case "float", "double":
			fieldType = "float64"
		case "time", "datetime":
			fieldType = "time.Time"
		}
	}
	
	jsonTag := strings.ToLower(property.Name)
	return fmt.Sprintf("\t%s %s `json:\"%s\"`", fieldName, fieldType, jsonTag), nil
}

// CodeGenerationTransformer generates Go code from mappings
type CodeGenerationTransformer struct {
	Config ConversionConfig
}

// GenerateWorkflowCode generates workflow Go code
func (t *CodeGenerationTransformer) GenerateWorkflowCode(processName string, mappings []ElementMapping) (string, error) {
	var sb strings.Builder
	
	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", t.Config.PackageName))
	
	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"time\"\n")
	sb.WriteString("\t\"go.temporal.io/sdk/workflow\"\n")
	sb.WriteString(")\n\n")
	
	// Workflow function
	workflowName := SanitizeGoIdentifier(processName) + "Workflow"
	sb.WriteString(fmt.Sprintf("// %s implements the %s BPMN process\n", workflowName, processName))
	sb.WriteString(fmt.Sprintf("func %s(ctx workflow.Context, input WorkflowInput) (WorkflowOutput, error) {\n", workflowName))
	
	// Generate workflow body based on mappings
	sb.WriteString("\tlogger := workflow.GetLogger(ctx)\n")
	sb.WriteString("\tlogger.Info(\"Starting workflow\", \"processName\", \"" + processName + "\")\n\n")
	
	// Add TODO for unsupported elements if configured
	if t.Config.Options.GenerateTODOs {
		for _, mapping := range mappings {
			if mapping.Status == MappingSkipped || mapping.Status == MappingFailed {
				sb.WriteString(fmt.Sprintf("\t// TODO: Implement %s (%s) - %s\n",
					mapping.BPMNElement.Name,
					mapping.BPMNElement.Type,
					"Unsupported element type"))
			}
		}
	}
	
	sb.WriteString("\n\t// Workflow implementation\n")
	sb.WriteString("\tvar result WorkflowOutput\n")
	sb.WriteString("\n\treturn result, nil\n")
	sb.WriteString("}\n")
	
	return sb.String(), nil
}

// GenerateActivityCode generates activity Go code
func (t *CodeGenerationTransformer) GenerateActivityCode(activities []ActivityMapping) (string, error) {
	var sb strings.Builder
	
	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", t.Config.PackageName))
	
	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"time\"\n")
	sb.WriteString("\t\"go.temporal.io/sdk/activity\"\n")
	sb.WriteString(")\n\n")
	
	// Generate each activity
	for _, activity := range activities {
		if activity.FunctionName == "inline" {
			continue // Skip inline scripts
		}
		
		sb.WriteString(fmt.Sprintf("// %s implements the %s activity\n", activity.FunctionName, activity.Name))
		sb.WriteString(fmt.Sprintf("func %s(ctx context.Context, input %s) (%s, error) {\n",
			activity.FunctionName,
			activity.InputType,
			activity.OutputType))
		
		sb.WriteString("\tlogger := activity.GetLogger(ctx)\n")
		sb.WriteString(fmt.Sprintf("\tlogger.Info(\"Executing activity\", \"name\", \"%s\")\n\n", activity.Name))
		
		if activity.IsHumanTask {
			sb.WriteString("\t// Human task - wait for signal\n")
			sb.WriteString("\t// TODO: Implement human task logic\n")
		}
		
		sb.WriteString("\n\t// TODO: Implement activity logic\n")
		sb.WriteString(fmt.Sprintf("\tvar result %s\n", activity.OutputType))
		sb.WriteString("\treturn result, nil\n")
		sb.WriteString("}\n\n")
	}
	
	return sb.String(), nil
}

// GenerateTypeCode generates type definitions Go code
func (t *CodeGenerationTransformer) GenerateTypeCode(dataObjects []bpmn.DataObject, properties []bpmn.Property) (string, error) {
	var sb strings.Builder
	
	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", t.Config.PackageName))
	
	// Imports if needed
	needsTime := false
	for _, prop := range properties {
		if strings.Contains(strings.ToLower(prop.ItemSubject), "time") {
			needsTime = true
			break
		}
	}
	
	if needsTime {
		sb.WriteString("import \"time\"\n\n")
	}
	
	// Workflow input/output types
	sb.WriteString("// WorkflowInput represents the input to the workflow\n")
	sb.WriteString("type WorkflowInput struct {\n")
	for _, prop := range properties {
		if field, err := (&DataTransformer{Config: t.Config}).TransformProperty(prop); err == nil {
			sb.WriteString(field + "\n")
		}
	}
	sb.WriteString("}\n\n")
	
	sb.WriteString("// WorkflowOutput represents the output from the workflow\n")
	sb.WriteString("type WorkflowOutput struct {\n")
	sb.WriteString("\tSuccess bool `json:\"success\"`\n")
	sb.WriteString("\tMessage string `json:\"message\"`\n")
	sb.WriteString("\tData interface{} `json:\"data,omitempty\"`\n")
	sb.WriteString("}\n\n")
	
	// Data object types
	transformer := &DataTransformer{Config: t.Config}
	for _, dataObj := range dataObjects {
		if structDef, err := transformer.TransformDataObject(dataObj); err == nil {
			sb.WriteString(structDef)
			sb.WriteString("\n")
		}
	}
	
	// Activity input/output types (simplified)
	sb.WriteString("// ActivityInput represents generic activity input\n")
	sb.WriteString("type ActivityInput map[string]interface{}\n\n")
	
	sb.WriteString("// ActivityOutput represents generic activity output\n")
	sb.WriteString("type ActivityOutput map[string]interface{}\n")
	
	return sb.String(), nil
}