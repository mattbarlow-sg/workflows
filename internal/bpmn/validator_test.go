package bpmn

import (
	"strings"
	"testing"
)

// Helper function to update element connections based on sequence flows
func updateProcessConnections(p *Process) {
	// Clear existing connections
	for i := range p.ProcessInfo.Elements.Events {
		p.ProcessInfo.Elements.Events[i].Incoming = []string{}
		p.ProcessInfo.Elements.Events[i].Outgoing = []string{}
	}
	for i := range p.ProcessInfo.Elements.Activities {
		p.ProcessInfo.Elements.Activities[i].Incoming = []string{}
		p.ProcessInfo.Elements.Activities[i].Outgoing = []string{}
	}
	for i := range p.ProcessInfo.Elements.Gateways {
		p.ProcessInfo.Elements.Gateways[i].Incoming = []string{}
		p.ProcessInfo.Elements.Gateways[i].Outgoing = []string{}
	}

	// Update connections based on flows
	for _, flow := range p.ProcessInfo.Elements.SequenceFlows {
		// Update source element's outgoing
		for i := range p.ProcessInfo.Elements.Events {
			if p.ProcessInfo.Elements.Events[i].ID == flow.SourceRef {
				p.ProcessInfo.Elements.Events[i].Outgoing = append(p.ProcessInfo.Elements.Events[i].Outgoing, flow.ID)
			}
			if p.ProcessInfo.Elements.Events[i].ID == flow.TargetRef {
				p.ProcessInfo.Elements.Events[i].Incoming = append(p.ProcessInfo.Elements.Events[i].Incoming, flow.ID)
			}
		}
		for i := range p.ProcessInfo.Elements.Activities {
			if p.ProcessInfo.Elements.Activities[i].ID == flow.SourceRef {
				p.ProcessInfo.Elements.Activities[i].Outgoing = append(p.ProcessInfo.Elements.Activities[i].Outgoing, flow.ID)
			}
			if p.ProcessInfo.Elements.Activities[i].ID == flow.TargetRef {
				p.ProcessInfo.Elements.Activities[i].Incoming = append(p.ProcessInfo.Elements.Activities[i].Incoming, flow.ID)
			}
		}
		for i := range p.ProcessInfo.Elements.Gateways {
			if p.ProcessInfo.Elements.Gateways[i].ID == flow.SourceRef {
				p.ProcessInfo.Elements.Gateways[i].Outgoing = append(p.ProcessInfo.Elements.Gateways[i].Outgoing, flow.ID)
			}
			if p.ProcessInfo.Elements.Gateways[i].ID == flow.TargetRef {
				p.ProcessInfo.Elements.Gateways[i].Incoming = append(p.ProcessInfo.Elements.Gateways[i].Incoming, flow.ID)
			}
		}
	}
}

func TestValidatorBasicProcess(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID:           "test_process",
			Name:         "Test Process",
			IsExecutable: true,
			Elements: Elements{
				Events: []Event{
					{ID: "start", Name: "Start", Type: "startEvent"},
					{ID: "end", Name: "End", Type: "endEvent"},
				},
				Activities: []Activity{
					{ID: "task1", Name: "Task 1", Type: "userTask"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Basic process should be valid, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}

func TestValidatorMissingStartEvent(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "end", Type: "endEvent"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask"},
				},
			},
		},
	}

	validator := NewValidator(process)
	result := validator.Validate()

	if result.Valid {
		t.Error("Process without start event should be invalid")
	}

	foundError := false
	for _, err := range result.Errors {
		if err.Rule == "process.start.missing" || err.Message == "Process must have at least one start event" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Should have error for missing start event")
		for _, err := range result.Errors {
			t.Logf("Error rule: %s, message: %s", err.Rule, err.Message)
		}
	}
}

func TestValidatorUnconnectedElements(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask"},
					{ID: "task2", Type: "userTask"}, // Unconnected
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if result.Valid {
		t.Error("Process with unconnected elements should be invalid")
	}

	foundError := false
	for _, err := range result.Errors {
		if err.Message == "Activity must be connected to the process flow" && strings.Contains(err.Path, "task2") {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Should have error for unconnected task2")
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}

func TestValidatorExclusiveGateway(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Gateways: []Gateway{
					{ID: "xor1", Type: "exclusiveGateway", GatewayDirection: "diverging"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask"},
					{ID: "task2", Type: "userTask"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "xor1"},
					{ID: "flow2", SourceRef: "xor1", TargetRef: "task1"},
					{ID: "flow3", SourceRef: "xor1", TargetRef: "task2"},
					{ID: "flow4", SourceRef: "task1", TargetRef: "end"},
					{ID: "flow5", SourceRef: "task2", TargetRef: "end"},
				},
			},
		},
	}

	validator := NewValidator(process)
	result := validator.Validate()

	// Should have warning about missing conditions
	if len(result.Warnings) == 0 {
		t.Error("Should have warnings about missing conditions on exclusive gateway")
	}
}

func TestValidatorParallelGateway(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Gateways: []Gateway{
					{ID: "split", Type: "parallelGateway", GatewayDirection: "diverging"},
					{ID: "join", Type: "parallelGateway", GatewayDirection: "converging"},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask"},
					{ID: "task2", Type: "userTask"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "split"},
					{ID: "flow2", SourceRef: "split", TargetRef: "task1"},
					{ID: "flow3", SourceRef: "split", TargetRef: "task2"},
					{ID: "flow4", SourceRef: "task1", TargetRef: "join"},
					{ID: "flow5", SourceRef: "task2", TargetRef: "join"},
					{ID: "flow6", SourceRef: "join", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Valid parallel gateway process should pass validation, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}

func TestValidatorBoundaryEvent(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
					{
						ID:          "timeout",
						Type:        "boundaryEvent",
						AttachedTo:  "task1",
						IsInterrupt: true,
						EventType:   "timer",
					},
				},
				Activities: []Activity{
					{ID: "task1", Type: "userTask"},
					{ID: "handleTimeout", Type: "userTask"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
					{ID: "flow3", SourceRef: "timeout", TargetRef: "handleTimeout"},
					{ID: "flow4", SourceRef: "handleTimeout", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Process with boundary event should be valid, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}

func TestValidatorAgentAssignment(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Activities: []Activity{
					{
						ID:   "task1",
						Type: "userTask",
						Agent: &AgentAssignment{
							Type:     "human",
							Strategy: "dynamic",
							AssignmentRules: []DynamicAssignmentRule{
								{
									Condition: Expression{
										Language: "javascript",
										Body:     "task.type == 'review'",
									},
									TargetAgent: AgentAssignment{
										Type: "human",
									},
								},
							},
						},
					},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Process with agent assignment should be valid, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}

func TestValidatorReviewWorkflow(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID: "test_process",
			Elements: Elements{
				Events: []Event{
					{ID: "start", Type: "startEvent"},
					{ID: "end", Type: "endEvent"},
				},
				Activities: []Activity{
					{
						ID:   "task1",
						Type: "userTask",
						Agent: &AgentAssignment{
							Type: "ai",
						},
						Review: &ReviewConfig{
							Required: true,
							Type:     "approval",
							Reviewer: AgentAssignment{
								Type: "human",
							},
						},
					},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	// Update connections
	updateProcessConnections(process)

	validator := NewValidator(process)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Process with review workflow should be valid, got %d errors", len(result.Errors))
		for _, err := range result.Errors {
			t.Logf("Error: %s - %s", err.Path, err.Message)
		}
	}
}
