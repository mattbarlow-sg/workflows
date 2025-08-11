package bpmn

import (
	"fmt"
)

// Validator performs semantic validation on BPMN processes
type Validator struct {
	process  *Process
	errors   []ValidationError
	warnings []ValidationError
}

// NewValidator creates a new BPMN validator
func NewValidator(process *Process) *Validator {
	return &Validator{
		process:  process,
		errors:   []ValidationError{},
		warnings: []ValidationError{},
	}
}

// Validate performs comprehensive semantic validation
func (v *Validator) Validate() *ValidationResult {
	// Reset errors and warnings
	v.errors = []ValidationError{}
	v.warnings = []ValidationError{}

	// Run validation checks
	v.validateProcessStructure()
	v.validateStartAndEndEvents()
	v.validateSequenceFlows()
	v.validateGateways()
	v.validateActivities()
	v.validateBoundaryEvents()
	v.validateAgentAssignments()
	v.validateReviewWorkflows()

	return &ValidationResult{
		Valid:         len(v.errors) == 0,
		Errors:        v.errors,
		Warnings:      v.warnings,
		SchemaValid:   true, // Assume schema validation passed if we're here
		SemanticValid: len(v.errors) == 0,
		GraphValid:    true, // Will be set by graph analyzer
	}
}

// validateProcessStructure checks basic process structure
func (v *Validator) validateProcessStructure() {
	elements := &v.process.ProcessInfo.Elements

	// Check for at least one element
	totalElements := len(elements.Events) + len(elements.Activities) +
		len(elements.Gateways) + len(elements.Artifacts)

	if totalElements == 0 {
		v.addError("process", "Process must contain at least one element", "structure.empty")
	}

	// Check for duplicate IDs
	idMap := make(map[string]string)
	v.checkDuplicateIDs(idMap)
}

// validateStartAndEndEvents checks start and end event rules
func (v *Validator) validateStartAndEndEvents() {
	elements := &v.process.ProcessInfo.Elements
	startEvents := 0
	endEvents := 0

	for _, event := range elements.Events {
		switch event.Type {
		case "startEvent":
			startEvents++
			// Start events should not have incoming flows
			if len(event.Incoming) > 0 {
				v.addError(event.ID, "Start event should not have incoming sequence flows", "event.start.incoming")
			}
			// Start events should have at least one outgoing flow
			if len(event.Outgoing) == 0 {
				v.addError(event.ID, "Start event must have at least one outgoing sequence flow", "event.start.outgoing")
			}

		case "endEvent":
			endEvents++
			// End events should not have outgoing flows
			if len(event.Outgoing) > 0 {
				v.addError(event.ID, "End event should not have outgoing sequence flows", "event.end.outgoing")
			}
			// End events should have at least one incoming flow
			if len(event.Incoming) == 0 {
				v.addError(event.ID, "End event must have at least one incoming sequence flow", "event.end.incoming")
			}
		}
	}

	// Check for required events
	if startEvents == 0 {
		v.addError("process", "Process must have at least one start event", "process.start.missing")
	}
	if endEvents == 0 {
		v.addWarning("process", "Process should have at least one end event", "process.end.missing")
	}
}

// validateSequenceFlows checks sequence flow validity
func (v *Validator) validateSequenceFlows() {
	elements := &v.process.ProcessInfo.Elements
	elementIDs := make(map[string]bool)

	// Build element ID map
	for _, e := range elements.Events {
		elementIDs[e.ID] = true
	}
	for _, a := range elements.Activities {
		elementIDs[a.ID] = true
	}
	for _, g := range elements.Gateways {
		elementIDs[g.ID] = true
	}

	// Validate each sequence flow
	for _, flow := range elements.SequenceFlows {
		// Check source exists
		if !elementIDs[flow.SourceRef] {
			v.addError(flow.ID, fmt.Sprintf("Sequence flow source '%s' does not exist", flow.SourceRef), "flow.source.invalid")
		}

		// Check target exists
		if !elementIDs[flow.TargetRef] {
			v.addError(flow.ID, fmt.Sprintf("Sequence flow target '%s' does not exist", flow.TargetRef), "flow.target.invalid")
		}

		// Check for self-loops
		if flow.SourceRef == flow.TargetRef {
			v.addWarning(flow.ID, "Sequence flow creates a self-loop", "flow.selfloop")
		}
	}
}

// validateGateways checks gateway-specific rules
func (v *Validator) validateGateways() {
	elements := &v.process.ProcessInfo.Elements

	for _, gateway := range elements.Gateways {
		switch gateway.Type {
		case "exclusiveGateway":
			v.validateExclusiveGateway(gateway)
		case "parallelGateway":
			v.validateParallelGateway(gateway)
		case "inclusiveGateway":
			v.validateInclusiveGateway(gateway)
		}

		// General gateway validations
		if len(gateway.Incoming) == 0 && len(gateway.Outgoing) == 0 {
			v.addError(gateway.ID, "Gateway must have at least one incoming or outgoing flow", "gateway.isolated")
		}
	}
}

// validateExclusiveGateway validates exclusive gateway rules
func (v *Validator) validateExclusiveGateway(gateway Gateway) {
	// For diverging gateways, check conditions
	if gateway.GatewayDirection == "diverging" ||
		(gateway.GatewayDirection == "" && len(gateway.Outgoing) > 1) {

		// Count flows with conditions
		conditionedFlows := 0
		for _, flowID := range gateway.Outgoing {
			flow := v.findSequenceFlow(flowID)
			if flow != nil && flow.ConditionExpression != nil {
				conditionedFlows++
			}
		}

		// At least one flow should have a condition (unless there's a default)
		if conditionedFlows == 0 && gateway.DefaultFlow == "" && len(gateway.Outgoing) > 1 {
			v.addWarning(gateway.ID, "Exclusive gateway should have conditions on outgoing flows or a default flow", "gateway.exclusive.conditions")
		}
	}
}

// validateParallelGateway validates parallel gateway rules
func (v *Validator) validateParallelGateway(gateway Gateway) {
	// Parallel gateways should not have conditions on outgoing flows
	if gateway.GatewayDirection == "diverging" ||
		(gateway.GatewayDirection == "" && len(gateway.Outgoing) > 1) {

		for _, flowID := range gateway.Outgoing {
			flow := v.findSequenceFlow(flowID)
			if flow != nil && flow.ConditionExpression != nil {
				v.addError(gateway.ID, "Parallel gateway should not have conditions on outgoing flows", "gateway.parallel.conditions")
			}
		}
	}
}

// validateInclusiveGateway validates inclusive gateway rules
func (v *Validator) validateInclusiveGateway(gateway Gateway) {
	// Similar to exclusive but all conditions can be true
	if gateway.GatewayDirection == "diverging" ||
		(gateway.GatewayDirection == "" && len(gateway.Outgoing) > 1) {

		// Should have at least one default flow for safety
		if gateway.DefaultFlow == "" && len(gateway.Outgoing) > 1 {
			v.addWarning(gateway.ID, "Inclusive gateway should have a default flow", "gateway.inclusive.default")
		}
	}
}

// validateActivities validates activity-specific rules
func (v *Validator) validateActivities() {
	elements := &v.process.ProcessInfo.Elements

	for _, activity := range elements.Activities {
		// All activities should be connected
		if len(activity.Incoming) == 0 && len(activity.Outgoing) == 0 {
			v.addError(activity.ID, "Activity must be connected to the process flow", "activity.isolated")
		}

		// Script tasks must have scripts
		if activity.Type == "scriptTask" && activity.Script == nil {
			v.addError(activity.ID, "Script task must have a script definition", "activity.script.missing")
		}

		// Check for proper agent assignment
		if activity.Type == "userTask" && activity.Agent == nil {
			v.addWarning(activity.ID, "User task should have an agent assignment", "activity.agent.missing")
		}
	}
}

// validateBoundaryEvents validates boundary event attachments
func (v *Validator) validateBoundaryEvents() {
	elements := &v.process.ProcessInfo.Elements
	activityIDs := make(map[string]bool)

	// Build activity ID map
	for _, a := range elements.Activities {
		activityIDs[a.ID] = true
	}

	// Check boundary events
	for _, event := range elements.Events {
		if event.Type == "boundaryEvent" {
			if event.AttachedTo == "" {
				v.addError(event.ID, "Boundary event must be attached to an activity", "event.boundary.attachment")
			} else if !activityIDs[event.AttachedTo] {
				v.addError(event.ID, fmt.Sprintf("Boundary event attached to non-existent activity '%s'", event.AttachedTo), "event.boundary.invalid")
			}

			// Boundary events should have outgoing flows
			if len(event.Outgoing) == 0 {
				v.addError(event.ID, "Boundary event must have at least one outgoing flow", "event.boundary.outgoing")
			}
		}
	}
}

// validateAgentAssignments validates agent configurations
func (v *Validator) validateAgentAssignments() {
	elements := &v.process.ProcessInfo.Elements

	for _, activity := range elements.Activities {
		if activity.Agent != nil {
			// Validate assignment strategy
			if activity.Agent.Type == "unspecified" && activity.Agent.Strategy == "" {
				v.addError(activity.ID, "Unspecified agent must have an assignment strategy", "agent.strategy.missing")
			}

			// Validate assignment rules
			for i, rule := range activity.Agent.AssignmentRules {
				if rule.Condition.Body == "" {
					v.addError(activity.ID, fmt.Sprintf("Assignment rule %d has empty condition", i+1), "agent.rule.empty")
				}
			}
		}
	}
}

// validateReviewWorkflows validates review configurations
func (v *Validator) validateReviewWorkflows() {
	elements := &v.process.ProcessInfo.Elements

	for _, activity := range elements.Activities {
		if activity.Review != nil && activity.Review.Required {
			// Check reviewer assignment
			if activity.Review.Reviewer.Type == "" {
				v.addError(activity.ID, "Review configuration must specify reviewer", "review.reviewer.missing")
			}

			// Check timeout configuration
			if activity.Review.Timeout != "" && activity.Review.OnTimeout == "" {
				v.addWarning(activity.ID, "Review timeout specified without timeout action", "review.timeout.action")
			}

			// Validate review type
			if activity.Review.Type == "" {
				v.addError(activity.ID, "Review configuration must specify type", "review.type.missing")
			}
		}
	}
}

// Helper methods

func (v *Validator) checkDuplicateIDs(idMap map[string]string) {
	elements := &v.process.ProcessInfo.Elements

	// Check all element types
	for _, e := range elements.Events {
		if existing, found := idMap[e.ID]; found {
			v.addError(e.ID, fmt.Sprintf("Duplicate ID: also used by %s", existing), "id.duplicate")
		}
		idMap[e.ID] = "event"
	}

	for _, a := range elements.Activities {
		if existing, found := idMap[a.ID]; found {
			v.addError(a.ID, fmt.Sprintf("Duplicate ID: also used by %s", existing), "id.duplicate")
		}
		idMap[a.ID] = "activity"
	}

	for _, g := range elements.Gateways {
		if existing, found := idMap[g.ID]; found {
			v.addError(g.ID, fmt.Sprintf("Duplicate ID: also used by %s", existing), "id.duplicate")
		}
		idMap[g.ID] = "gateway"
	}

	for _, f := range elements.SequenceFlows {
		if existing, found := idMap[f.ID]; found {
			v.addError(f.ID, fmt.Sprintf("Duplicate ID: also used by %s", existing), "id.duplicate")
		}
		idMap[f.ID] = "flow"
	}
}

func (v *Validator) findSequenceFlow(id string) *SequenceFlow {
	for i := range v.process.ProcessInfo.Elements.SequenceFlows {
		if v.process.ProcessInfo.Elements.SequenceFlows[i].ID == id {
			return &v.process.ProcessInfo.Elements.SequenceFlows[i]
		}
	}
	return nil
}

func (v *Validator) addError(path, message, rule string) {
	v.errors = append(v.errors, ValidationError{
		Level:   "error",
		Type:    "semantic",
		Path:    path,
		Message: message,
		Rule:    rule,
	})
}

func (v *Validator) addWarning(path, message, rule string) {
	v.warnings = append(v.warnings, ValidationError{
		Level:   "warning",
		Type:    "semantic",
		Path:    path,
		Message: message,
		Rule:    rule,
	})
}
