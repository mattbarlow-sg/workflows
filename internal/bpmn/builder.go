package bpmn

import (
	"fmt"
)

// ProcessBuilder provides a fluent interface for building BPMN processes
type ProcessBuilder struct {
	process *Process
	errors  []error
}

// NewProcessBuilder creates a new process builder
func NewProcessBuilder(id, name string) *ProcessBuilder {
	return &ProcessBuilder{
		process: &Process{
			Type:    "bpmn:process",
			Version: "2.0",
			ProcessInfo: ProcessInfo{
				ID:           id,
				Name:         name,
				IsExecutable: true,
				Elements:     Elements{},
			},
		},
		errors: []error{},
	}
}

// WithDescription adds a description to the process
func (pb *ProcessBuilder) WithDescription(desc string) *ProcessBuilder {
	pb.process.ProcessInfo.Description = desc
	return pb
}

// WithWorkSessionID sets the work session ID for the process
func (pb *ProcessBuilder) WithWorkSessionID(workSessionID string) *ProcessBuilder {
	pb.process.ProcessInfo.WorkSessionID = workSessionID
	return pb
}

// AddStartEvent adds a start event to the process
func (pb *ProcessBuilder) AddStartEvent(id, name string) *ProcessBuilder {
	event := Event{
		ID:        id,
		Name:      name,
		Type:      "startEvent",
		EventType: "none",
		Outgoing:  []string{},
	}
	pb.process.ProcessInfo.Elements.Events = append(pb.process.ProcessInfo.Elements.Events, event)
	return pb
}

// AddEndEvent adds an end event to the process
func (pb *ProcessBuilder) AddEndEvent(id, name string) *ProcessBuilder {
	event := Event{
		ID:        id,
		Name:      name,
		Type:      "endEvent",
		EventType: "none",
		Incoming:  []string{},
	}
	pb.process.ProcessInfo.Elements.Events = append(pb.process.ProcessInfo.Elements.Events, event)
	return pb
}

// AddUserTask adds a user task with agent assignment
func (pb *ProcessBuilder) AddUserTask(id, name string, agent *AgentAssignment) *ProcessBuilder {
	task := Activity{
		ID:       id,
		Name:     name,
		Type:     "userTask",
		Agent:    agent,
		Incoming: []string{},
		Outgoing: []string{},
	}
	pb.process.ProcessInfo.Elements.Activities = append(pb.process.ProcessInfo.Elements.Activities, task)
	return pb
}

// AddServiceTask adds a service task
func (pb *ProcessBuilder) AddServiceTask(id, name string, agent *AgentAssignment) *ProcessBuilder {
	task := Activity{
		ID:       id,
		Name:     name,
		Type:     "serviceTask",
		Agent:    agent,
		Incoming: []string{},
		Outgoing: []string{},
	}
	pb.process.ProcessInfo.Elements.Activities = append(pb.process.ProcessInfo.Elements.Activities, task)
	return pb
}

// AddScriptTask adds a script task
func (pb *ProcessBuilder) AddScriptTask(id, name, language, script string) *ProcessBuilder {
	task := Activity{
		ID:   id,
		Name: name,
		Type: "scriptTask",
		Script: &Script{
			Language: language,
			Body:     script,
		},
		Incoming: []string{},
		Outgoing: []string{},
	}
	pb.process.ProcessInfo.Elements.Activities = append(pb.process.ProcessInfo.Elements.Activities, task)
	return pb
}

// AddExclusiveGateway adds an exclusive gateway
func (pb *ProcessBuilder) AddExclusiveGateway(id, name string) *ProcessBuilder {
	gateway := Gateway{
		ID:       id,
		Name:     name,
		Type:     "exclusiveGateway",
		Incoming: []string{},
		Outgoing: []string{},
	}
	pb.process.ProcessInfo.Elements.Gateways = append(pb.process.ProcessInfo.Elements.Gateways, gateway)
	return pb
}

// AddParallelGateway adds a parallel gateway
func (pb *ProcessBuilder) AddParallelGateway(id, name string, direction string) *ProcessBuilder {
	gateway := Gateway{
		ID:               id,
		Name:             name,
		Type:             "parallelGateway",
		GatewayDirection: direction,
		Incoming:         []string{},
		Outgoing:         []string{},
	}
	pb.process.ProcessInfo.Elements.Gateways = append(pb.process.ProcessInfo.Elements.Gateways, gateway)
	return pb
}

// ConnectElements adds a sequence flow between two elements
func (pb *ProcessBuilder) ConnectElements(id, sourceID, targetID string) *ProcessBuilder {
	flow := SequenceFlow{
		ID:        id,
		SourceRef: sourceID,
		TargetRef: targetID,
	}

	// Update source element's outgoing
	if err := pb.addOutgoing(sourceID, id); err != nil {
		pb.errors = append(pb.errors, err)
	}

	// Update target element's incoming
	if err := pb.addIncoming(targetID, id); err != nil {
		pb.errors = append(pb.errors, err)
	}

	pb.process.ProcessInfo.Elements.SequenceFlows = append(pb.process.ProcessInfo.Elements.SequenceFlows, flow)
	return pb
}

// ConnectWithCondition adds a conditional sequence flow
func (pb *ProcessBuilder) ConnectWithCondition(id, sourceID, targetID, language, condition string) *ProcessBuilder {
	flow := SequenceFlow{
		ID:        id,
		SourceRef: sourceID,
		TargetRef: targetID,
		ConditionExpression: &Expression{
			Language: language,
			Body:     condition,
		},
	}

	// Update source element's outgoing
	if err := pb.addOutgoing(sourceID, id); err != nil {
		pb.errors = append(pb.errors, err)
	}

	// Update target element's incoming
	if err := pb.addIncoming(targetID, id); err != nil {
		pb.errors = append(pb.errors, err)
	}

	pb.process.ProcessInfo.Elements.SequenceFlows = append(pb.process.ProcessInfo.Elements.SequenceFlows, flow)
	return pb
}

// AddDataObject adds a data object to the process
func (pb *ProcessBuilder) AddDataObject(id, name string) *ProcessBuilder {
	dataObj := DataObject{
		ID:   id,
		Name: name,
	}
	pb.process.ProcessInfo.DataObjects = append(pb.process.ProcessInfo.DataObjects, dataObj)
	return pb
}

// AddTextAnnotation adds a text annotation
func (pb *ProcessBuilder) AddTextAnnotation(id, text string) *ProcessBuilder {
	artifact := Artifact{
		ID:   id,
		Type: "textAnnotation",
		Text: text,
	}
	pb.process.ProcessInfo.Elements.Artifacts = append(pb.process.ProcessInfo.Elements.Artifacts, artifact)
	return pb
}

// WithReview adds review configuration to the last added activity
func (pb *ProcessBuilder) WithReview(reviewType string, reviewer *AgentAssignment) *ProcessBuilder {
	activities := pb.process.ProcessInfo.Elements.Activities
	if len(activities) == 0 {
		pb.errors = append(pb.errors, fmt.Errorf("no activity to add review to"))
		return pb
	}

	lastActivity := &pb.process.ProcessInfo.Elements.Activities[len(activities)-1]
	lastActivity.Review = &ReviewConfig{
		Required: true,
		Type:     reviewType,
		Reviewer: *reviewer,
	}

	return pb
}

// Build returns the constructed process and any errors
func (pb *ProcessBuilder) Build() (*Process, error) {
	if len(pb.errors) > 0 {
		return nil, fmt.Errorf("builder errors: %v", pb.errors)
	}
	return pb.process, nil
}

// Helper methods

func (pb *ProcessBuilder) addOutgoing(elementID, flowID string) error {
	// Check events
	for i := range pb.process.ProcessInfo.Elements.Events {
		if pb.process.ProcessInfo.Elements.Events[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Events[i].Outgoing = append(
				pb.process.ProcessInfo.Elements.Events[i].Outgoing, flowID)
			return nil
		}
	}

	// Check activities
	for i := range pb.process.ProcessInfo.Elements.Activities {
		if pb.process.ProcessInfo.Elements.Activities[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Activities[i].Outgoing = append(
				pb.process.ProcessInfo.Elements.Activities[i].Outgoing, flowID)
			return nil
		}
	}

	// Check gateways
	for i := range pb.process.ProcessInfo.Elements.Gateways {
		if pb.process.ProcessInfo.Elements.Gateways[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Gateways[i].Outgoing = append(
				pb.process.ProcessInfo.Elements.Gateways[i].Outgoing, flowID)
			return nil
		}
	}

	return fmt.Errorf("element %s not found", elementID)
}

func (pb *ProcessBuilder) addIncoming(elementID, flowID string) error {
	// Check events
	for i := range pb.process.ProcessInfo.Elements.Events {
		if pb.process.ProcessInfo.Elements.Events[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Events[i].Incoming = append(
				pb.process.ProcessInfo.Elements.Events[i].Incoming, flowID)
			return nil
		}
	}

	// Check activities
	for i := range pb.process.ProcessInfo.Elements.Activities {
		if pb.process.ProcessInfo.Elements.Activities[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Activities[i].Incoming = append(
				pb.process.ProcessInfo.Elements.Activities[i].Incoming, flowID)
			return nil
		}
	}

	// Check gateways
	for i := range pb.process.ProcessInfo.Elements.Gateways {
		if pb.process.ProcessInfo.Elements.Gateways[i].ID == elementID {
			pb.process.ProcessInfo.Elements.Gateways[i].Incoming = append(
				pb.process.ProcessInfo.Elements.Gateways[i].Incoming, flowID)
			return nil
		}
	}

	return fmt.Errorf("element %s not found", elementID)
}

// Quick builder functions for common patterns

// NewHumanAgent creates a new human agent
func NewHumanAgent(name, role string) *AgentAssignment {
	return &AgentAssignment{
		Type: "human",
		Name: name,
		Role: role,
	}
}

// NewAIAgent creates a new AI agent
func NewAIAgent(name string, capabilities []string) *AgentAssignment {
	return &AgentAssignment{
		Type:         "ai",
		Name:         name,
		Capabilities: capabilities,
	}
}

// NewSystemAgent creates a new system agent
func NewSystemAgent(id string) *AgentAssignment {
	return &AgentAssignment{
		Type: "system",
		ID:   id,
	}
}

// NewUnspecifiedAgent creates an unspecified agent for runtime assignment
func NewUnspecifiedAgent(strategy string) *AgentAssignment {
	return &AgentAssignment{
		Type:     "unspecified",
		Strategy: strategy,
	}
}
