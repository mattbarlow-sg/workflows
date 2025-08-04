package bpmn

import (
	"encoding/json"
)

// Process represents the root BPMN process structure
type Process struct {
	Type        string      `json:"$type" validate:"required,eq=bpmn:process"`
	Version     string      `json:"version" validate:"required,eq=2.0"`
	ProcessInfo ProcessInfo `json:"process" validate:"required"`
}

// ProcessInfo contains the main process information
type ProcessInfo struct {
	ID           string       `json:"id" validate:"required"`
	Name         string       `json:"name" validate:"required"`
	Description  string       `json:"description,omitempty"`
	IsExecutable bool         `json:"isExecutable"`
	Elements     Elements     `json:"elements" validate:"required"`
	DataObjects  []DataObject `json:"dataObjects,omitempty"`
	Properties   []Property   `json:"properties,omitempty"`
}

// Elements contains all BPMN elements in the process
type Elements struct {
	Events        []Event        `json:"events,omitempty"`
	Activities    []Activity     `json:"activities,omitempty"`
	Gateways      []Gateway      `json:"gateways,omitempty"`
	SequenceFlows []SequenceFlow `json:"sequenceFlows,omitempty"`
	Associations  []Association  `json:"associations,omitempty"`
	Artifacts     []Artifact     `json:"artifacts,omitempty"`
}

// Event represents a BPMN event
type Event struct {
	ID          string              `json:"id" validate:"required"`
	Name        string              `json:"name,omitempty"`
	Type        string              `json:"type" validate:"required,oneof=startEvent endEvent intermediateThrowEvent intermediateCatchEvent boundaryEvent"`
	EventType   string              `json:"eventType,omitempty" validate:"omitempty,oneof=none message timer error signal compensation cancel conditional link terminate"`
	AttachedTo  string              `json:"attachedToRef,omitempty"`
	Incoming    []string            `json:"incoming,omitempty"`
	Outgoing    []string            `json:"outgoing,omitempty"`
	IsInterrupt bool                `json:"isInterrupting,omitempty"`
	Properties  map[string]any      `json:"properties,omitempty"`
}

// Activity represents a BPMN activity
type Activity struct {
	ID                 string              `json:"id" validate:"required"`
	Name               string              `json:"name" validate:"required"`
	Type               string              `json:"type" validate:"required,oneof=task userTask serviceTask scriptTask sendTask receiveTask manualTask businessRuleTask callActivity subProcess"`
	Description        string              `json:"description,omitempty"`
	Agent              *AgentAssignment    `json:"agent,omitempty"`
	Review             *ReviewConfig       `json:"review,omitempty"`
	ReviewWorkflow     *ReviewWorkflow     `json:"reviewWorkflow,omitempty"`
	Incoming           []string            `json:"incoming,omitempty"`
	Outgoing           []string            `json:"outgoing,omitempty"`
	IsSequential       bool                `json:"isSequential,omitempty"`
	CompletionQuantity int                 `json:"completionQuantity,omitempty"`
	StartQuantity      int                 `json:"startQuantity,omitempty"`
	Script             *Script             `json:"script,omitempty"`
	IOSpecification    *IOSpecification    `json:"ioSpecification,omitempty"`
	Properties         map[string]any      `json:"properties,omitempty"`
	BoundaryEvents     []string            `json:"boundaryEventRefs,omitempty"`
}

// Gateway represents a BPMN gateway
type Gateway struct {
	ID               string          `json:"id" validate:"required"`
	Name             string          `json:"name,omitempty"`
	Type             string          `json:"type" validate:"required,oneof=exclusiveGateway parallelGateway inclusiveGateway eventBasedGateway complexGateway"`
	GatewayDirection string          `json:"gatewayDirection,omitempty" validate:"omitempty,oneof=converging diverging mixed"`
	DefaultFlow      string          `json:"defaultFlow,omitempty"`
	Incoming         []string        `json:"incoming,omitempty"`
	Outgoing         []string        `json:"outgoing,omitempty"`
	Properties       map[string]any  `json:"properties,omitempty"`
}

// SequenceFlow represents a connection between elements
type SequenceFlow struct {
	ID                 string              `json:"id" validate:"required"`
	Name               string              `json:"name,omitempty"`
	SourceRef          string              `json:"sourceRef" validate:"required"`
	TargetRef          string              `json:"targetRef" validate:"required"`
	ConditionExpression *Expression        `json:"conditionExpression,omitempty"`
	IsDefault          bool                `json:"isDefault,omitempty"`
	Properties         map[string]any      `json:"properties,omitempty"`
}

// Association represents a non-flow connection
type Association struct {
	ID               string              `json:"id" validate:"required"`
	SourceRef        string              `json:"sourceRef" validate:"required"`
	TargetRef        string              `json:"targetRef" validate:"required"`
	AssociationDirection string          `json:"associationDirection,omitempty" validate:"omitempty,oneof=none one both"`
}

// Artifact represents documentation elements
type Artifact struct {
	ID          string              `json:"id" validate:"required"`
	Type        string              `json:"type" validate:"required,oneof=textAnnotation group"`
	Text        string              `json:"text,omitempty"`
	GroupCategory string            `json:"groupCategory,omitempty"`
}

// DataObject represents data elements
type DataObject struct {
	ID          string              `json:"id" validate:"required"`
	Name        string              `json:"name" validate:"required"`
	ItemSubject string              `json:"itemSubject,omitempty"`
	IsCollection bool               `json:"isCollection,omitempty"`
	State       string              `json:"state,omitempty"`
}

// Property represents process properties
type Property struct {
	ID          string              `json:"id" validate:"required"`
	Name        string              `json:"name" validate:"required"`
	ItemSubject string              `json:"itemSubject,omitempty"`
}

// AgentAssignment represents an agent assignment
type AgentAssignment struct {
	Type           string          `json:"type" validate:"required,oneof=human ai system unspecified"`
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	Role           string          `json:"role,omitempty"`
	Capabilities   []string        `json:"capabilities,omitempty"`
	Strategy       string          `json:"strategy,omitempty" validate:"omitempty,oneof=random round-robin least-loaded capability-based priority-based"`
	Priority       int             `json:"priority,omitempty"`
	AssignmentRules []DynamicAssignmentRule `json:"assignmentRules,omitempty"`
}

// DynamicAssignmentRule defines dynamic assignment rules in types
type DynamicAssignmentRule struct {
	Condition   Expression      `json:"condition" validate:"required"`
	TargetAgent AgentAssignment `json:"targetAgent" validate:"required"`
}

// ReviewConfig defines review requirements
type ReviewConfig struct {
	Required      bool            `json:"required"`
	Type          string          `json:"type" validate:"required,oneof=approval validation quality-check"`
	Reviewer      AgentAssignment `json:"reviewer" validate:"required"`
	Criteria      []string        `json:"criteria,omitempty"`
	RequiredScore float64         `json:"requiredScore,omitempty"`
	Timeout       string          `json:"timeout,omitempty"`
	OnTimeout     string          `json:"onTimeout,omitempty" validate:"omitempty,oneof=approve reject escalate"`
}

// ReviewWorkflow defines a review workflow
type ReviewWorkflow struct {
	ID              string           `json:"id" validate:"required"`
	Name            string           `json:"name" validate:"required"`
	Description     string           `json:"description,omitempty"`
	Type            string           `json:"type" validate:"required"`
	Pattern         string           `json:"pattern" validate:"required,oneof=ai-human human-ai collaborative peer-review hierarchical custom"`
	ReviewerID      string           `json:"reviewerId,omitempty"`
	Rules           *ReviewRules     `json:"rules,omitempty"`
}

// ReviewRules defines rules for review workflows
type ReviewRules struct {
	ApprovalThreshold float64                `json:"approvalThreshold,omitempty"`
	RequiredFields    []string               `json:"requiredFields,omitempty"`
	CustomRules       map[string]interface{} `json:"customRules,omitempty"`
}

// Script represents script task content
type Script struct {
	Language string `json:"language" validate:"required"`
	Body     string `json:"body" validate:"required"`
}

// Expression represents conditional expressions
type Expression struct {
	Language string `json:"language" validate:"required"`
	Body     string `json:"body" validate:"required"`
}

// IOSpecification defines input/output parameters
type IOSpecification struct {
	DataInputs  []DataInput  `json:"dataInputs,omitempty"`
	DataOutputs []DataOutput `json:"dataOutputs,omitempty"`
}

// DataInput represents input parameters
type DataInput struct {
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	ItemSubject string `json:"itemSubject,omitempty"`
}

// DataOutput represents output parameters
type DataOutput struct {
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	ItemSubject string `json:"itemSubject,omitempty"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid      bool              `json:"valid"`
	Errors     []ValidationError `json:"errors,omitempty"`
	Warnings   []ValidationError `json:"warnings,omitempty"`
	SchemaValid bool             `json:"schemaValid"`
	SemanticValid bool           `json:"semanticValid"`
	GraphValid bool              `json:"graphValid"`
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	Level    string `json:"level"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Rule     string `json:"rule,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
}

// ProcessAnalysisMetrics contains analysis metrics
type ProcessAnalysisMetrics struct {
	ElementCount     int              `json:"elementCount"`
	CyclomaticComplexity int         `json:"cyclomaticComplexity"`
	Depth            int              `json:"depth"`
	ParallelGateways int              `json:"parallelGateways"`
	DecisionPoints   int              `json:"decisionPoints"`
	ReachableElements []string        `json:"reachableElements"`
	UnreachableElements []string     `json:"unreachableElements"`
	DeadlockRisks    []DeadlockRisk   `json:"deadlockRisks,omitempty"`
}

// DeadlockRisk represents potential deadlock points
type DeadlockRisk struct {
	Elements []string `json:"elements"`
	Type     string   `json:"type"`
	Risk     string   `json:"risk"`
}

// Agent defines an actual agent that can be assigned
type Agent struct {
	ID               string             `json:"id" validate:"required"`
	Name             string             `json:"name" validate:"required"`
	Type             string             `json:"type" validate:"required,oneof=human system ai service external"`
	Description      string             `json:"description,omitempty"`
	Capabilities     []string           `json:"capabilities,omitempty"`
	Availability     *AgentAvailability `json:"availability,omitempty"`
	Constraints      *AgentConstraints  `json:"constraints,omitempty"`
}

// AgentAvailability defines when an agent is available
type AgentAvailability struct {
	Available bool   `json:"available"`
	Until     string `json:"until,omitempty"`
}

// AgentConstraints defines constraints for an agent
type AgentConstraints struct {
	MaxConcurrentTasks  int              `json:"maxConcurrentTasks,omitempty"`
	AllowedProcessTypes []string         `json:"allowedProcessTypes,omitempty"`
	TimeConstraints     *TimeConstraints `json:"timeConstraints,omitempty"`
}

// TimeConstraints defines time-based constraints
type TimeConstraints struct {
	BusinessHoursOnly   bool         `json:"businessHoursOnly,omitempty"`
	AvailabilityWindows []TimeWindow `json:"availabilityWindows,omitempty"`
}

// TimeWindow represents a time window
type TimeWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Unmarshal helpers for JSON parsing
func (p *Process) UnmarshalJSON(data []byte) error {
	type Alias Process
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(data, &aux)
}

// Validate performs basic structural validation
func (p *Process) Validate() error {
	// Basic validation will be implemented in validator.go
	return nil
}

// GetElement retrieves an element by ID
func (p *Process) GetElement(id string) interface{} {
	// Check events
	for _, e := range p.ProcessInfo.Elements.Events {
		if e.ID == id {
			return e
		}
	}
	
	// Check activities
	for _, a := range p.ProcessInfo.Elements.Activities {
		if a.ID == id {
			return a
		}
	}
	
	// Check gateways
	for _, g := range p.ProcessInfo.Elements.Gateways {
		if g.ID == id {
			return g
		}
	}
	
	// Check artifacts
	for _, art := range p.ProcessInfo.Elements.Artifacts {
		if art.ID == id {
			return art
		}
	}
	
	return nil
}

// GetAllElementIDs returns all element IDs in the process
func (p *Process) GetAllElementIDs() []string {
	var ids []string
	
	for _, e := range p.ProcessInfo.Elements.Events {
		ids = append(ids, e.ID)
	}
	for _, a := range p.ProcessInfo.Elements.Activities {
		ids = append(ids, a.ID)
	}
	for _, g := range p.ProcessInfo.Elements.Gateways {
		ids = append(ids, g.ID)
	}
	for _, art := range p.ProcessInfo.Elements.Artifacts {
		ids = append(ids, art.ID)
	}
	
	return ids
}