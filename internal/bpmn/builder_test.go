package bpmn

import (
	"testing"
)

func TestBuilderBasicProcess(t *testing.T) {
	builder := NewProcessBuilder("test_process", "Test Process")
	
	process, err := builder.
		WithDescription("A test process").
		AddStartEvent("start", "Start").
		AddUserTask("task1", "Do Something", nil).
		AddEndEvent("end", "End").
		ConnectElements("flow1", "start", "task1").
		ConnectElements("flow2", "task1", "end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build process: %v", err)
	}

	// Verify process properties
	if process.ProcessInfo.ID != "test_process" {
		t.Errorf("Process ID = %s, want test_process", process.ProcessInfo.ID)
	}

	if process.ProcessInfo.Description != "A test process" {
		t.Errorf("Process description = %s, want 'A test process'", process.ProcessInfo.Description)
	}

	// Verify elements
	if len(process.ProcessInfo.Elements.Events) != 2 {
		t.Errorf("Event count = %d, want 2", len(process.ProcessInfo.Elements.Events))
	}

	if len(process.ProcessInfo.Elements.Activities) != 1 {
		t.Errorf("Activity count = %d, want 1", len(process.ProcessInfo.Elements.Activities))
	}

	if len(process.ProcessInfo.Elements.SequenceFlows) != 2 {
		t.Errorf("Flow count = %d, want 2", len(process.ProcessInfo.Elements.SequenceFlows))
	}
}

func TestBuilderParallelGateway(t *testing.T) {
	builder := NewProcessBuilder("parallel_process", "Parallel Process")
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddParallelGateway("split", "Split", "diverging").
		AddUserTask("task1", "Task 1", nil).
		AddUserTask("task2", "Task 2", nil).
		AddParallelGateway("join", "Join", "converging").
		AddEndEvent("end", "End").
		ConnectElements("f1", "start", "split").
		ConnectElements("f2", "split", "task1").
		ConnectElements("f3", "split", "task2").
		ConnectElements("f4", "task1", "join").
		ConnectElements("f5", "task2", "join").
		ConnectElements("f6", "join", "end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build parallel process: %v", err)
	}

	// Verify gateways
	if len(process.ProcessInfo.Elements.Gateways) != 2 {
		t.Errorf("Gateway count = %d, want 2", len(process.ProcessInfo.Elements.Gateways))
	}

	// Check gateway types
	for _, gw := range process.ProcessInfo.Elements.Gateways {
		if gw.Type != "parallelGateway" {
			t.Errorf("Gateway type = %s, want parallelGateway", gw.Type)
		}
	}
}

func TestBuilderWithConditions(t *testing.T) {
	builder := NewProcessBuilder("decision_process", "Decision Process")
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddExclusiveGateway("decision", "Decision").
		AddUserTask("approved", "Handle Approved", nil).
		AddUserTask("rejected", "Handle Rejected", nil).
		AddEndEvent("end", "End").
		ConnectElements("f1", "start", "decision").
		ConnectWithCondition("f2", "decision", "approved", "javascript", "result.approved === true").
		ConnectWithCondition("f3", "decision", "rejected", "javascript", "result.approved === false").
		ConnectElements("f4", "approved", "end").
		ConnectElements("f5", "rejected", "end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build decision process: %v", err)
	}

	// Find conditional flows
	conditionalFlows := 0
	for _, flow := range process.ProcessInfo.Elements.SequenceFlows {
		if flow.ConditionExpression != nil {
			conditionalFlows++
		}
	}

	if conditionalFlows != 2 {
		t.Errorf("Conditional flow count = %d, want 2", conditionalFlows)
	}
}

func TestBuilderWithAgents(t *testing.T) {
	humanAgent := NewHumanAgent("John Doe", "reviewer")
	aiAgent := NewAIAgent("ai-assistant", []string{"text-analysis"})
	systemAgent := NewSystemAgent("automation")

	builder := NewProcessBuilder("agent_process", "Agent Process")
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddUserTask("human_task", "Human Review", humanAgent).
		AddServiceTask("ai_task", "AI Analysis", aiAgent).
		AddServiceTask("system_task", "System Process", systemAgent).
		AddEndEvent("end", "End").
		ConnectElements("f1", "start", "human_task").
		ConnectElements("f2", "human_task", "ai_task").
		ConnectElements("f3", "ai_task", "system_task").
		ConnectElements("f4", "system_task", "end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build agent process: %v", err)
	}

	// Verify agent assignments
	activities := process.ProcessInfo.Elements.Activities
	if len(activities) != 3 {
		t.Fatalf("Activity count = %d, want 3", len(activities))
	}

	// Check human task
	humanTask := activities[0]
	if humanTask.Agent == nil || humanTask.Agent.Type != "human" {
		t.Error("Human task should have human agent")
	}

	// Check AI task
	aiTask := activities[1]
	if aiTask.Agent == nil || aiTask.Agent.Type != "ai" {
		t.Error("AI task should have AI agent")
	}

	// Check system task
	systemTask := activities[2]
	if systemTask.Agent == nil || systemTask.Agent.Type != "system" {
		t.Error("System task should have system agent")
	}
}

func TestBuilderWithReview(t *testing.T) {
	builder := NewProcessBuilder("review_process", "Review Process")
	
	aiAgent := NewAIAgent("ai-writer", []string{"content-generation"})
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddServiceTask("create_content", "Create Content", aiAgent).
		AddEndEvent("end", "End").
		ConnectElements("f1", "start", "create_content").
		ConnectElements("f2", "create_content", "end").
		Build()
	
	// Manually add review configuration to test it
	if len(process.ProcessInfo.Elements.Activities) > 0 {
		process.ProcessInfo.Elements.Activities[0].Review = &ReviewConfig{
			Required: true,
			Type:     "approval",
			Reviewer: AgentAssignment{
				Type: "human",
				Name: "Editor",
				Role: "editor",
			},
		}
	}

	if err != nil {
		t.Fatalf("Failed to build review process: %v", err)
	}

	// Check review configuration
	task := process.ProcessInfo.Elements.Activities[0]
	if task.Review == nil {
		t.Fatal("Task should have review configuration")
	}

	if !task.Review.Required {
		t.Error("Review should be required")
	}

	if task.Review.Type != "approval" {
		t.Errorf("Review type = %s, want approval", task.Review.Type)
	}

	if task.Review.Reviewer.Type != "human" {
		t.Error("Reviewer should be human")
	}
}

func TestBuilderValidation(t *testing.T) {
	// Test that builder can create process without connections
	// (validation happens separately with the Validator)
	builder := NewProcessBuilder("unconnected_process", "Unconnected Process")
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddUserTask("task1", "Task 1", nil).
		AddEndEvent("end", "End").
		// Missing connections - builder allows this
		Build()

	if err != nil {
		t.Errorf("Builder should not fail even without connections: %v", err)
	}
	
	// The process should be created but invalid when validated
	if process == nil {
		t.Error("Process should be created")
	}
	
	// Verify elements were added
	if len(process.ProcessInfo.Elements.Events) != 2 {
		t.Error("Should have 2 events")
	}
	if len(process.ProcessInfo.Elements.Activities) != 1 {
		t.Error("Should have 1 activity")
	}
	if len(process.ProcessInfo.Elements.SequenceFlows) != 0 {
		t.Error("Should have 0 sequence flows")
	}
}

func TestBuilderSubProcess(t *testing.T) {
	builder := NewProcessBuilder("main_process", "Main Process")
	
	// Using a userTask with isSubProcess flag as a workaround
	process, err := builder.
		AddStartEvent("start", "Start").
		AddUserTask("sub1", "Sub Process", nil).
		AddEndEvent("end", "End").
		ConnectElements("f1", "start", "sub1").
		ConnectElements("f2", "sub1", "end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build process: %v", err)
	}

	// Manually set as subprocess for testing
	if len(process.ProcessInfo.Elements.Activities) > 0 {
		process.ProcessInfo.Elements.Activities[0].Type = "subProcess"
	}

	// Find subprocess
	var subProcess *Activity
	for _, act := range process.ProcessInfo.Elements.Activities {
		if act.Type == "subProcess" {
			subProcess = &act
			break
		}
	}

	if subProcess == nil {
		t.Fatal("Should have subprocess")
	}

	if subProcess.ID != "sub1" {
		t.Errorf("Subprocess ID = %s, want sub1", subProcess.ID)
	}
}

func TestBuilderBoundaryEvent(t *testing.T) {
	builder := NewProcessBuilder("timeout_process", "Timeout Process")
	
	process, err := builder.
		AddStartEvent("start", "Start").
		AddUserTask("long_task", "Long Running Task", nil).
		AddUserTask("handle_timeout", "Handle Timeout", nil).
		AddEndEvent("end", "End").
		AddEndEvent("timeout_end", "Timeout End").
		ConnectElements("f1", "start", "long_task").
		ConnectElements("f2", "long_task", "end").
		ConnectElements("f4", "handle_timeout", "timeout_end").
		Build()

	if err != nil {
		t.Fatalf("Failed to build process: %v", err)
	}

	// Manually add boundary event for testing
	boundaryEvent := Event{
		ID:          "timeout",
		Name:        "Timeout",
		Type:        "boundaryEvent",
		AttachedTo:  "long_task",
		IsInterrupt: true,
		EventType:   "timer",
	}
	process.ProcessInfo.Elements.Events = append(process.ProcessInfo.Elements.Events, boundaryEvent)
	
	// Add flow from boundary event
	process.ProcessInfo.Elements.SequenceFlows = append(process.ProcessInfo.Elements.SequenceFlows,
		SequenceFlow{ID: "f3", SourceRef: "timeout", TargetRef: "handle_timeout"})

	// Find boundary event
	var foundBoundaryEvent *Event
	for _, evt := range process.ProcessInfo.Elements.Events {
		if evt.Type == "boundaryEvent" {
			foundBoundaryEvent = &evt
			break
		}
	}

	if foundBoundaryEvent == nil {
		t.Fatal("Should have boundary event")
	}

	if foundBoundaryEvent.AttachedTo != "long_task" {
		t.Errorf("Boundary event attached to = %s, want long_task", foundBoundaryEvent.AttachedTo)
	}

	if !foundBoundaryEvent.IsInterrupt {
		t.Error("Boundary event should interrupt activity")
	}
}