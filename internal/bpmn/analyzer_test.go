package bpmn

import (
	"testing"
)

func TestAnalyzerBasicProcess(t *testing.T) {
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
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	// Check reachability
	if len(result.Reachability.UnreachableElements) > 0 {
		t.Errorf("Should have no unreachable elements, got %v", result.Reachability.UnreachableElements)
	}

	// Check paths
	if len(result.Paths.AllPaths) != 1 {
		t.Errorf("Should have 1 path, got %d", len(result.Paths.AllPaths))
	}

	if result.Paths.MaxPathLength != 3 {
		t.Errorf("Max path length should be 3, got %d", result.Paths.MaxPathLength)
	}

	// Check metrics
	if result.Metrics.Elements.Total != 3 {
		t.Errorf("Total elements should be 3, got %d", result.Metrics.Elements.Total)
	}
}

func TestAnalyzerUnreachableElements(t *testing.T) {
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
					{ID: "task2", Type: "userTask"}, // Unreachable
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	if len(result.Reachability.UnreachableElements) != 1 {
		t.Errorf("Should have 1 unreachable element, got %v", result.Reachability.UnreachableElements)
	}

	if result.Reachability.UnreachableElements[0] != "task2" {
		t.Errorf("Unreachable element should be task2, got %s", result.Reachability.UnreachableElements[0])
	}
}

func TestAnalyzerDeadlock(t *testing.T) {
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
				},
				Gateways: []Gateway{
					{ID: "split", Type: "parallelGateway", GatewayDirection: "diverging"},
					{ID: "join", Type: "parallelGateway", GatewayDirection: "converging"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "split"},
					{ID: "flow2", SourceRef: "split", TargetRef: "task1"},
					{ID: "flow3", SourceRef: "split", TargetRef: "join"}, // Direct path - incomplete join
					{ID: "flow4", SourceRef: "task1", TargetRef: "join"},
					{ID: "flow5", SourceRef: "join", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	if len(result.Deadlocks) == 0 {
		t.Error("Should detect deadlock")
	}

	foundIncompleteJoin := false
	for _, deadlock := range result.Deadlocks {
		if deadlock.Type == "incomplete-join" {
			foundIncompleteJoin = true
			break
		}
	}

	if !foundIncompleteJoin {
		t.Error("Should detect incomplete join deadlock")
	}
}

func TestAnalyzerParallelPaths(t *testing.T) {
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
					{ID: "task2", Type: "userTask"},
				},
				Gateways: []Gateway{
					{ID: "split", Type: "parallelGateway", GatewayDirection: "diverging"},
					{ID: "join", Type: "parallelGateway", GatewayDirection: "converging"},
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

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	if len(result.Paths.AllPaths) != 2 {
		t.Errorf("Should have 2 paths, got %d", len(result.Paths.AllPaths))
	}

	if result.Metrics.Width != 2 {
		t.Errorf("Process width should be 2, got %d", result.Metrics.Width)
	}
}

func TestAnalyzerLoop(t *testing.T) {
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
					{ID: "task2", Type: "userTask"},
				},
				Gateways: []Gateway{
					{ID: "decision", Type: "exclusiveGateway"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "decision"},
					{ID: "flow3", SourceRef: "decision", TargetRef: "task2"},
					{ID: "flow4", SourceRef: "decision", TargetRef: "task1"}, // Loop back
					{ID: "flow5", SourceRef: "task2", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	if !result.Paths.LoopDetected {
		t.Error("Should detect loop")
	}

	if len(result.Paths.Loops) == 0 {
		t.Error("Should have at least one loop")
	}
}

func TestAnalyzerComplexity(t *testing.T) {
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
					{ID: "task2", Type: "userTask"},
					{ID: "task3", Type: "userTask"},
				},
				Gateways: []Gateway{
					{ID: "gw1", Type: "exclusiveGateway"},
					{ID: "gw2", Type: "parallelGateway"},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "gw1"},
					{ID: "flow3", SourceRef: "gw1", TargetRef: "task2"},
					{ID: "flow4", SourceRef: "gw1", TargetRef: "gw2"},
					{ID: "flow5", SourceRef: "gw2", TargetRef: "task3"},
					{ID: "flow6", SourceRef: "task2", TargetRef: "end"},
					{ID: "flow7", SourceRef: "task3", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	// Complexity = V + E - 2N = 7 + 7 - 2(1) = 12
	if result.Metrics.Complexity != 12 {
		t.Errorf("Complexity should be 12, got %d", result.Metrics.Complexity)
	}
}

func TestAnalyzerAgentWorkload(t *testing.T) {
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
							Type: "human",
							ID:   "agent1",
						},
					},
					{
						ID:   "task2",
						Type: "userTask",
						Agent: &AgentAssignment{
							Type: "human",
							ID:   "agent1",
						},
					},
					{
						ID:   "task3",
						Type: "userTask",
						Agent: &AgentAssignment{
							Type: "human",
							ID:   "agent2",
						},
					},
				},
				SequenceFlows: []SequenceFlow{
					{ID: "flow1", SourceRef: "start", TargetRef: "task1"},
					{ID: "flow2", SourceRef: "task1", TargetRef: "task2"},
					{ID: "flow3", SourceRef: "task2", TargetRef: "task3"},
					{ID: "flow4", SourceRef: "task3", TargetRef: "end"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(process)
	result := analyzer.Analyze()

	if len(result.AgentWorkload.AgentTasks["agent1"]) != 2 {
		t.Errorf("Agent1 should have 2 tasks, got %d", len(result.AgentWorkload.AgentTasks["agent1"]))
	}

	if len(result.AgentWorkload.AgentTasks["agent2"]) != 1 {
		t.Errorf("Agent2 should have 1 task, got %d", len(result.AgentWorkload.AgentTasks["agent2"]))
	}

	// Workload balance score should indicate imbalance
	if result.AgentWorkload.WorkloadBalance == 0 {
		t.Error("Should have non-zero workload balance score indicating imbalance")
	}
}
