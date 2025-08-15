// Package temporal provides BPMN element mapping and code generation
package temporal

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// ElementMapper handles mapping of BPMN elements to Temporal constructs
type ElementMapper struct {
	nameSanitizer *NameSanitizer
}

// NewElementMapper creates a new element mapper
func NewElementMapper() *ElementMapper {
	return &ElementMapper{
		nameSanitizer: NewNameSanitizer(),
	}
}

// MapActivity maps a BPMN activity to a Temporal activity pattern
func (m *ElementMapper) MapActivity(activity bpmn.Activity) (*schemas.ActivityMapping, error) {
	mapping := &schemas.ActivityMapping{
		BPMNActivity: activity,
		Name:         activity.Name,
		FunctionName: m.nameSanitizer.ToFunctionName(activity.Name),
		InputType:    m.determineInputType(&activity),
		OutputType:   m.determineOutputType(&activity),
		Options:      m.createActivityOptions(&activity),
		IsHumanTask:  activity.Type == "userTask",
		HasSignal:    activity.Type == "receiveTask",
	}

	// Set signal name for receive tasks
	if activity.Type == "receiveTask" {
		mapping.SignalName = m.nameSanitizer.ToSignalName(activity.Name)
	}

	// Extract timeout from properties if present
	if timeout, ok := activity.Properties["timeout"].(string); ok {
		if duration, err := time.ParseDuration(timeout); err == nil {
			mapping.TimeoutDuration = duration
		}
	}

	return mapping, nil
}

// MapGateway maps a BPMN gateway to Temporal control flow
func (m *ElementMapper) MapGateway(gateway bpmn.Gateway) (*schemas.GatewayMapping, error) {
	mapping := &schemas.GatewayMapping{
		BPMNGateway: gateway,
		Type:        gateway.Type,
	}

	switch gateway.Type {
	case "exclusiveGateway":
		mapping.ControlFlow = "selector"
		mapping.Conditions = m.extractConditions(&gateway)
		mapping.DefaultPath = gateway.DefaultFlow
		
	case "parallelGateway":
		mapping.ControlFlow = "parallel"
		mapping.ParallelCount = len(gateway.Outgoing)
		
	case "inclusiveGateway":
		mapping.ControlFlow = "multiSelector"
		mapping.Conditions = m.extractConditions(&gateway)
		
	case "eventBasedGateway":
		mapping.ControlFlow = "eventSelector"
		
	default:
		return nil, fmt.Errorf("unsupported gateway type: %s", gateway.Type)
	}

	return mapping, nil
}

// MapEvent maps a BPMN event to Temporal event handler
func (m *ElementMapper) MapEvent(event bpmn.Event) (*schemas.EventMapping, error) {
	mapping := &schemas.EventMapping{
		BPMNEvent:      event,
		IsStart:        event.Type == "startEvent",
		IsEnd:          event.Type == "endEvent",
		IsBoundary:     event.Type == "boundaryEvent",
		IsInterrupting: event.IsInterrupt,
	}

	// Map event type to handler type
	switch event.EventType {
	case "timer":
		mapping.HandlerType = "timer"
		if duration, ok := event.Properties["duration"].(string); ok {
			if d, err := time.ParseDuration(duration); err == nil {
				mapping.TimerDuration = d
			}
		}
		
	case "message", "signal":
		mapping.HandlerType = "signal"
		mapping.SignalName = m.nameSanitizer.ToSignalName(event.Name)
		
	case "error":
		mapping.HandlerType = "error"
		mapping.ErrorType = m.determineErrorType(&event)
		
	default:
		mapping.HandlerType = "none"
	}

	return mapping, nil
}

// CodeGenerator generates Go code from BPMN mappings
type CodeGenerator struct {
	imports *ImportsManager
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		imports: NewImportsManager(),
	}
}

// GenerateWorkflow generates workflow Go code
func (g *CodeGenerator) GenerateWorkflow(packageName string, mappings []schemas.ElementMapping) (string, error) {
	var builder strings.Builder
	
	// Package declaration
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	
	// Imports
	g.imports.AddImport("context", "")
	g.imports.AddImport("time", "")
	g.imports.AddImport("go.temporal.io/sdk/workflow", "")
	builder.WriteString(g.imports.Generate())
	builder.WriteString("\n")
	
	// Workflow function
	builder.WriteString("// ProcessWorkflow is the main workflow function\n")
	builder.WriteString("func ProcessWorkflow(ctx workflow.Context, input WorkflowInput) (*WorkflowOutput, error) {\n")
	builder.WriteString("\tlogger := workflow.GetLogger(ctx)\n")
	builder.WriteString("\tlogger.Info(\"Workflow started\", \"input\", input)\n\n")
	
	// Generate workflow body from mappings
	for _, mapping := range mappings {
		code := g.generateMappingCode(mapping)
		if code != "" {
			builder.WriteString(code)
			builder.WriteString("\n")
		}
	}
	
	// Return statement
	builder.WriteString("\n\t// Workflow completed\n")
	builder.WriteString("\treturn &WorkflowOutput{\n")
	builder.WriteString("\t\tStatus: \"completed\",\n")
	builder.WriteString("\t\tResult: result,\n")
	builder.WriteString("\t}, nil\n")
	builder.WriteString("}\n")
	
	return builder.String(), nil
}

// GenerateActivities generates activities Go code
func (g *CodeGenerator) GenerateActivities(activities []schemas.ActivityMapping) (string, error) {
	var builder strings.Builder
	
	// Package declaration
	builder.WriteString("package workflows\n\n")
	
	// Imports
	g.imports.Reset()
	g.imports.AddImport("context", "")
	g.imports.AddImport("fmt", "")
	g.imports.AddImport("time", "")
	g.imports.AddImport("go.temporal.io/sdk/activity", "")
	builder.WriteString(g.imports.Generate())
	builder.WriteString("\n")
	
	// Generate each activity
	for _, activity := range activities {
		builder.WriteString(g.generateActivityFunction(activity))
		builder.WriteString("\n")
	}
	
	// Register activities function
	builder.WriteString("// RegisterActivities registers all workflow activities\n")
	builder.WriteString("func RegisterActivities(w worker.Worker) {\n")
	for _, activity := range activities {
		builder.WriteString(fmt.Sprintf("\tw.RegisterActivity(%s)\n", activity.FunctionName))
	}
	builder.WriteString("}\n")
	
	return builder.String(), nil
}

// GenerateTypes generates type definitions Go code
func (g *CodeGenerator) GenerateTypes(dataObjects []bpmn.DataObject, properties []bpmn.Property) (string, error) {
	var builder strings.Builder
	
	// Package declaration
	builder.WriteString("package workflows\n\n")
	
	// Imports if needed
	builder.WriteString("import (\n")
	builder.WriteString("\t\"time\"\n")
	builder.WriteString(")\n\n")
	
	// Workflow input/output types
	builder.WriteString("// WorkflowInput represents the workflow input parameters\n")
	builder.WriteString("type WorkflowInput struct {\n")
	builder.WriteString("\tProcessID   string                 `json:\"processId\"`\n")
	builder.WriteString("\tProcessName string                 `json:\"processName\"`\n")
	builder.WriteString("\tParameters  map[string]interface{} `json:\"parameters\"`\n")
	builder.WriteString("\tStartedBy   string                 `json:\"startedBy\"`\n")
	builder.WriteString("\tStartedAt   time.Time              `json:\"startedAt\"`\n")
	builder.WriteString("}\n\n")
	
	builder.WriteString("// WorkflowOutput represents the workflow output\n")
	builder.WriteString("type WorkflowOutput struct {\n")
	builder.WriteString("\tStatus      string                 `json:\"status\"`\n")
	builder.WriteString("\tResult      map[string]interface{} `json:\"result\"`\n")
	builder.WriteString("\tCompletedAt time.Time              `json:\"completedAt\"`\n")
	builder.WriteString("\tDuration    time.Duration          `json:\"duration\"`\n")
	builder.WriteString("}\n\n")
	
	// Activity input/output types
	builder.WriteString("// ActivityInput represents generic activity input\n")
	builder.WriteString("type ActivityInput struct {\n")
	builder.WriteString("\tActivityID string                 `json:\"activityId\"`\n")
	builder.WriteString("\tName       string                 `json:\"name\"`\n")
	builder.WriteString("\tData       map[string]interface{} `json:\"data\"`\n")
	builder.WriteString("}\n\n")
	
	builder.WriteString("// ActivityOutput represents generic activity output\n")
	builder.WriteString("type ActivityOutput struct {\n")
	builder.WriteString("\tSuccess bool                   `json:\"success\"`\n")
	builder.WriteString("\tResult  map[string]interface{} `json:\"result\"`\n")
	builder.WriteString("\tError   string                 `json:\"error,omitempty\"`\n")
	builder.WriteString("}\n\n")
	
	// Generate types from data objects
	for _, dataObj := range dataObjects {
		builder.WriteString(g.generateDataObjectType(dataObj))
		builder.WriteString("\n")
	}
	
	// Generate types from properties
	for _, prop := range properties {
		builder.WriteString(g.generatePropertyType(prop))
		builder.WriteString("\n")
	}
	
	return builder.String(), nil
}

// GenerateTests generates test code for the workflow
func (g *CodeGenerator) GenerateTests(packageName string, mappings []schemas.ElementMapping) (string, error) {
	var builder strings.Builder
	
	// Package declaration
	builder.WriteString(fmt.Sprintf("package %s_test\n\n", packageName))
	
	// Imports
	builder.WriteString("import (\n")
	builder.WriteString("\t\"context\"\n")
	builder.WriteString("\t\"testing\"\n")
	builder.WriteString("\t\"time\"\n\n")
	builder.WriteString("\t\"github.com/stretchr/testify/assert\"\n")
	builder.WriteString("\t\"github.com/stretchr/testify/require\"\n")
	builder.WriteString("\t\"go.temporal.io/sdk/testsuite\"\n")
	builder.WriteString(fmt.Sprintf("\t. \"%s\"\n", packageName))
	builder.WriteString(")\n\n")
	
	// Test suite
	builder.WriteString("type WorkflowTestSuite struct {\n")
	builder.WriteString("\ttestsuite.WorkflowTestSuite\n")
	builder.WriteString("}\n\n")
	
	// Main workflow test
	builder.WriteString("func (s *WorkflowTestSuite) TestProcessWorkflow() {\n")
	builder.WriteString("\tenv := s.NewTestWorkflowEnvironment()\n\n")
	
	// Register activities
	builder.WriteString("\t// Register activities\n")
	for _, mapping := range mappings {
		if mapping.TemporalConstruct == schemas.TemporalActivity {
			builder.WriteString(fmt.Sprintf("\tenv.RegisterActivity(%s)\n", "ActivityName"))
		}
	}
	builder.WriteString("\n")
	
	// Test input
	builder.WriteString("\t// Prepare test input\n")
	builder.WriteString("\tinput := WorkflowInput{\n")
	builder.WriteString("\t\tProcessID:   \"test-process\",\n")
	builder.WriteString("\t\tProcessName: \"Test Process\",\n")
	builder.WriteString("\t\tStartedBy:   \"test-user\",\n")
	builder.WriteString("\t\tStartedAt:   time.Now(),\n")
	builder.WriteString("\t}\n\n")
	
	// Execute workflow
	builder.WriteString("\t// Execute workflow\n")
	builder.WriteString("\tenv.ExecuteWorkflow(ProcessWorkflow, input)\n\n")
	
	// Assertions
	builder.WriteString("\t// Check results\n")
	builder.WriteString("\trequire.True(s.T(), env.IsWorkflowCompleted())\n")
	builder.WriteString("\trequire.NoError(s.T(), env.GetWorkflowError())\n\n")
	
	builder.WriteString("\tvar output WorkflowOutput\n")
	builder.WriteString("\terr := env.GetWorkflowResult(&output)\n")
	builder.WriteString("\trequire.NoError(s.T(), err)\n")
	builder.WriteString("\tassert.Equal(s.T(), \"completed\", output.Status)\n")
	builder.WriteString("}\n\n")
	
	// Test runner
	builder.WriteString("func TestWorkflowTestSuite(t *testing.T) {\n")
	builder.WriteString("\tsuite.Run(t, new(WorkflowTestSuite))\n")
	builder.WriteString("}\n")
	
	return builder.String(), nil
}

// Private helper methods

func (m *ElementMapper) determineInputType(activity *bpmn.Activity) string {
	if activity.IOSpecification != nil && len(activity.IOSpecification.DataInputs) > 0 {
		// Extract from IO specification
		return "ActivityInput"
	}
	return "map[string]interface{}"
}

func (m *ElementMapper) determineOutputType(activity *bpmn.Activity) string {
	if activity.IOSpecification != nil && len(activity.IOSpecification.DataOutputs) > 0 {
		// Extract from IO specification
		return "ActivityOutput"
	}
	return "map[string]interface{}"
}

func (m *ElementMapper) createActivityOptions(activity *bpmn.Activity) schemas.ActivityOptions {
	options := schemas.ActivityOptions{
		TaskQueue:              "default",
		ScheduleToCloseTimeout: 10 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
	}
	
	// Adjust for human tasks
	if activity.Type == "userTask" {
		options.ScheduleToCloseTimeout = 24 * time.Hour
		options.StartToCloseTimeout = 24 * time.Hour
		options.HeartbeatTimeout = 1 * time.Minute
	}
	
	// Add retry policy for service tasks
	if activity.Type == "serviceTask" {
		options.RetryPolicy = &schemas.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		}
	}
	
	return options
}

func (m *ElementMapper) extractConditions(gateway *bpmn.Gateway) []schemas.Condition {
	var conditions []schemas.Condition
	
	// In a real implementation, this would extract conditions from
	// the sequence flows connected to the gateway
	for _, outgoing := range gateway.Outgoing {
		conditions = append(conditions, schemas.Condition{
			Expression: fmt.Sprintf("condition_%s", outgoing),
			TargetPath: outgoing,
		})
	}
	
	return conditions
}

func (m *ElementMapper) determineErrorType(event *bpmn.Event) string {
	if errorCode, ok := event.Properties["errorCode"].(string); ok {
		return errorCode
	}
	return "GenericError"
}

func (g *CodeGenerator) generateMappingCode(mapping schemas.ElementMapping) string {
	var builder strings.Builder
	
	switch mapping.TemporalConstruct {
	case schemas.TemporalActivity:
		builder.WriteString(fmt.Sprintf("\t// Execute activity: %s\n", mapping.BPMNElement.Name))
		builder.WriteString("\tactivityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{\n")
		builder.WriteString("\t\tScheduleToCloseTimeout: 10 * time.Minute,\n")
		builder.WriteString("\t})\n")
		builder.WriteString(fmt.Sprintf("\tvar %sResult ActivityOutput\n", sanitizeName(mapping.BPMNElement.ID)))
		builder.WriteString(fmt.Sprintf("\terr := workflow.ExecuteActivity(activityCtx, %sActivity, input).Get(ctx, &%sResult)\n", 
			sanitizeName(mapping.BPMNElement.Name), sanitizeName(mapping.BPMNElement.ID)))
		builder.WriteString("\tif err != nil {\n")
		builder.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"activity %s failed: %%w\", err)\n", mapping.BPMNElement.Name))
		builder.WriteString("\t}\n")
		
	case schemas.TemporalSelector:
		builder.WriteString(fmt.Sprintf("\t// Gateway: %s\n", mapping.BPMNElement.Name))
		builder.WriteString("\tselector := workflow.NewSelector(ctx)\n")
		builder.WriteString("\t// TODO: Add selector branches\n")
		
	case schemas.TemporalTimer:
		builder.WriteString(fmt.Sprintf("\t// Timer: %s\n", mapping.BPMNElement.Name))
		builder.WriteString("\tworkflow.Sleep(ctx, 1*time.Minute) // TODO: Configure duration\n")
		
	case schemas.TemporalSignal:
		builder.WriteString(fmt.Sprintf("\t// Signal: %s\n", mapping.BPMNElement.Name))
		builder.WriteString("\tvar signalData interface{}\n")
		builder.WriteString(fmt.Sprintf("\tworkflow.GetSignalChannel(ctx, \"%s\").Receive(ctx, &signalData)\n", 
			sanitizeName(mapping.BPMNElement.Name)))
	}
	
	return builder.String()
}

func (g *CodeGenerator) generateActivityFunction(activity schemas.ActivityMapping) string {
	var builder strings.Builder
	
	// Function comment
	builder.WriteString(fmt.Sprintf("// %s implements the %s activity\n", activity.FunctionName, activity.Name))
	
	// Handle human tasks specially
	if activity.IsHumanTask {
		builder.WriteString("// This is a human task - requires manual completion\n")
		builder.WriteString(fmt.Sprintf("func %s(ctx context.Context, input ActivityInput) (*ActivityOutput, error) {\n", activity.FunctionName))
		builder.WriteString("\tlogger := activity.GetLogger(ctx)\n")
		builder.WriteString(fmt.Sprintf("\tlogger.Info(\"Human task started\", \"task\", \"%s\")\n\n", activity.Name))
		builder.WriteString("\t// TODO: Implement human task logic\n")
		builder.WriteString("\t// This might involve:\n")
		builder.WriteString("\t// - Creating a task in an external system\n")
		builder.WriteString("\t// - Waiting for human completion\n")
		builder.WriteString("\t// - Retrieving results\n\n")
		builder.WriteString("\treturn &ActivityOutput{\n")
		builder.WriteString("\t\tSuccess: true,\n")
		builder.WriteString("\t\tResult:  map[string]interface{}{\"completed\": true},\n")
		builder.WriteString("\t}, nil\n")
		builder.WriteString("}\n")
	} else {
		// Regular activity
		builder.WriteString(fmt.Sprintf("func %s(ctx context.Context, input %s) (*%s, error) {\n", 
			activity.FunctionName, activity.InputType, activity.OutputType))
		builder.WriteString("\tlogger := activity.GetLogger(ctx)\n")
		builder.WriteString(fmt.Sprintf("\tlogger.Info(\"Activity started\", \"activity\", \"%s\")\n\n", activity.Name))
		builder.WriteString("\t// TODO: Implement activity logic\n\n")
		builder.WriteString("\treturn &ActivityOutput{\n")
		builder.WriteString("\t\tSuccess: true,\n")
		builder.WriteString("\t\tResult:  map[string]interface{}{},\n")
		builder.WriteString("\t}, nil\n")
		builder.WriteString("}\n")
	}
	
	return builder.String()
}

func (g *CodeGenerator) generateDataObjectType(dataObj bpmn.DataObject) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("// %s represents %s\n", sanitizeName(dataObj.Name), dataObj.Name))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", sanitizeName(dataObj.Name)))
	
	// Default fields for data objects
	builder.WriteString("\tID    string `json:\"id\"`\n")
	builder.WriteString("\tName  string `json:\"name\"`\n")
	builder.WriteString("\tValue interface{} `json:\"value\"`\n")
	if dataObj.State != "" {
		builder.WriteString(fmt.Sprintf("\tState string `json:\"state\"` // %s\n", dataObj.State))
	}
	if dataObj.IsCollection {
		builder.WriteString("\tIsCollection bool `json:\"isCollection\"`\n")
		builder.WriteString("\tItems []interface{} `json:\"items,omitempty\"`\n")
	}
	
	builder.WriteString("}\n")
	return builder.String()
}

func (g *CodeGenerator) generatePropertyType(prop bpmn.Property) string {
	// Similar to generateDataObjectType but for properties
	return ""
}

// NameSanitizer handles name conversions
type NameSanitizer struct{}

func NewNameSanitizer() *NameSanitizer {
	return &NameSanitizer{}
}

func (s *NameSanitizer) ToFunctionName(name string) string {
	return toExportedName(sanitizeName(name))
}

func (s *NameSanitizer) ToSignalName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(sanitizeName(name), " ", "_"))
}

// ImportsManager manages Go imports
type ImportsManager struct {
	imports map[string]string
}

func NewImportsManager() *ImportsManager {
	return &ImportsManager{
		imports: make(map[string]string),
	}
}

func (m *ImportsManager) AddImport(path, alias string) {
	m.imports[path] = alias
}

func (m *ImportsManager) Generate() string {
	if len(m.imports) == 0 {
		return ""
	}
	
	var builder strings.Builder
	builder.WriteString("import (\n")
	for path, alias := range m.imports {
		if alias != "" {
			builder.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, path))
		} else {
			builder.WriteString(fmt.Sprintf("\t\"%s\"\n", path))
		}
	}
	builder.WriteString(")\n")
	return builder.String()
}

func (m *ImportsManager) Reset() {
	m.imports = make(map[string]string)
}

// Helper functions

func sanitizeName(name string) string {
	// Remove special characters and spaces
	result := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
		   (ch >= '0' && ch <= '9') || ch == '_' {
			result += string(ch)
		} else if ch == ' ' || ch == '-' {
			result += "_"
		}
	}
	return result
}

func toExportedName(name string) string {
	sanitized := sanitizeName(name)
	if len(sanitized) == 0 {
		return "Field"
	}
	// Capitalize first letter
	return strings.ToUpper(sanitized[:1]) + sanitized[1:]
}

func inferGoType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int64:
		return "int64"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case map[string]interface{}:
		return "map[string]interface{}"
	case []interface{}:
		return "[]interface{}"
	default:
		return "interface{}"
	}
}