package bpmn

import (
	"testing"
)

func TestProcessCreation(t *testing.T) {
	process := &Process{
		ProcessInfo: ProcessInfo{
			ID:           "test_process",
			Name:         "Test Process",
			IsExecutable: true,
			Elements: Elements{
				Events: []Event{
					{
						ID:   "start",
						Name: "Start",
						Type: "startEvent",
					},
					{
						ID:   "end",
						Name: "End",
						Type: "endEvent",
					},
				},
				Activities: []Activity{
					{
						ID:   "task1",
						Name: "Task 1",
						Type: "userTask",
					},
				},
				SequenceFlows: []SequenceFlow{
					{
						ID:       "flow1",
						Name:     "Flow 1",
						SourceRef: "start",
						TargetRef: "task1",
					},
					{
						ID:       "flow2",
						Name:     "Flow 2",
						SourceRef: "task1",
						TargetRef: "end",
					},
				},
			},
		},
	}

	// Test element counts
	events := process.ProcessInfo.Elements.Events
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	activities := process.ProcessInfo.Elements.Activities
	if len(activities) != 1 {
		t.Errorf("Expected 1 activity, got %d", len(activities))
	}
	
	// Test finding elements
	var foundStart bool
	for _, e := range events {
		if e.ID == "start" {
			foundStart = true
			break
		}
	}
	if !foundStart {
		t.Error("Failed to find start event")
	}
}

func TestAgentTypes(t *testing.T) {
	tests := []struct {
		name     string
		agent    Agent
		wantType string
	}{
		{
			name: "Human agent",
			agent: Agent{
				ID:   "human1",
				Type: "human",
				Name: "John Doe",
			},
			wantType: "human",
		},
		{
			name: "AI agent",
			agent: Agent{
				ID:   "ai1",
				Type: "ai",
				Name: "AI Assistant",
			},
			wantType: "ai",
		},
		{
			name: "System agent",
			agent: Agent{
				ID:   "system1",
				Type: "system",
				Name: "Automation System",
			},
			wantType: "system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.agent.Type != tt.wantType {
				t.Errorf("Agent type = %v, want %v", tt.agent.Type, tt.wantType)
			}
		})
	}
}

func TestSequenceFlowWithCondition(t *testing.T) {
	flow := SequenceFlow{
		ID:        "flow1",
		Name:      "Conditional Flow",
		SourceRef: "gateway1",
		TargetRef: "task1",
		ConditionExpression: &Expression{
			Language: "javascript",
			Body:     "result.approved === true",
		},
	}

	if flow.ConditionExpression == nil {
		t.Error("Condition expression should not be nil")
	}

	if flow.ConditionExpression.Language != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", flow.ConditionExpression.Language)
	}
}

func TestGatewayTypes(t *testing.T) {
	tests := []struct {
		name        string
		gateway     Gateway
		wantType    string
		wantDirection string
	}{
		{
			name: "Exclusive gateway",
			gateway: Gateway{
				ID:       "xor1",
				Name:     "Decision Point",
				Type:     "exclusiveGateway",
				GatewayDirection: "diverging",
			},
			wantType:      "exclusiveGateway",
			wantDirection: "diverging",
		},
		{
			name: "Parallel gateway",
			gateway: Gateway{
				ID:       "and1",
				Name:     "Fork",
				Type:     "parallelGateway",
				GatewayDirection: "diverging",
			},
			wantType:      "parallelGateway",
			wantDirection: "diverging",
		},
		{
			name: "Inclusive gateway",
			gateway: Gateway{
				ID:       "or1",
				Name:     "Inclusive Decision",
				Type:     "inclusiveGateway",
				GatewayDirection: "converging",
			},
			wantType:      "inclusiveGateway",
			wantDirection: "converging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gateway.Type != tt.wantType {
				t.Errorf("Gateway type = %v, want %v", tt.gateway.Type, tt.wantType)
			}
			if tt.gateway.GatewayDirection != tt.wantDirection {
				t.Errorf("Gateway direction = %v, want %v", tt.gateway.GatewayDirection, tt.wantDirection)
			}
		})
	}
}

func TestActivityWithAgent(t *testing.T) {
	activity := Activity{
		ID:   "task1",
		Name: "Review Document",
		Type: "userTask",
		Agent: &AgentAssignment{
			Type:     "human",
			Strategy: "static",
			ID:       "reviewer1",
		},
	}

	if activity.Agent == nil {
		t.Fatal("Agent assignment should not be nil")
	}

	if activity.Agent.Type != "human" {
		t.Errorf("Expected agent type 'human', got '%s'", activity.Agent.Type)
	}

	if activity.Agent.Strategy != "static" {
		t.Errorf("Expected strategy 'static', got '%s'", activity.Agent.Strategy)
	}
}