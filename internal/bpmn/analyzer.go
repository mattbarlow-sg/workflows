package bpmn

import (
	"fmt"
	"sort"
	"strings"
)

// AnalysisResult contains the results of graph analysis
type AnalysisResult struct {
	Reachability   ReachabilityAnalysis   `json:"reachability"`
	Deadlocks      []DeadlockInfo        `json:"deadlocks"`
	Paths          PathAnalysis          `json:"paths"`
	Metrics        ProcessMetrics        `json:"metrics"`
	AgentWorkload  AgentWorkloadAnalysis `json:"agent_workload"`
}

// ReachabilityAnalysis contains reachability information
type ReachabilityAnalysis struct {
	UnreachableElements []string          `json:"unreachable_elements"`
	DeadEndElements     []string          `json:"dead_end_elements"`
	ReachableFromStart  map[string]bool   `json:"reachable_from_start"`
	ReachesEnd          map[string]bool   `json:"reaches_end"`
}

// DeadlockInfo describes a potential deadlock
type DeadlockInfo struct {
	Type        string   `json:"type"`
	Elements    []string `json:"elements"`
	Description string   `json:"description"`
}

// PathAnalysis contains path-related information
type PathAnalysis struct {
	CriticalPath      []string           `json:"critical_path"`
	AllPaths          [][]string         `json:"all_paths"`
	LoopDetected      bool               `json:"loop_detected"`
	Loops             []Loop             `json:"loops"`
	MaxPathLength     int                `json:"max_path_length"`
	MinPathLength     int                `json:"min_path_length"`
	AveragePathLength float64            `json:"average_path_length"`
}

// Loop describes a loop in the process
type Loop struct {
	Elements []string `json:"elements"`
	Type     string   `json:"type"` // "simple", "nested", "overlapping"
}

// ProcessMetrics contains complexity and other metrics
type ProcessMetrics struct {
	Elements      ElementCount `json:"elements"`
	Complexity    int          `json:"complexity"`
	Depth         int          `json:"depth"`
	Width         int          `json:"width"`
	Connectivity  float64      `json:"connectivity"`
}

// ElementCount tracks counts of different element types
type ElementCount struct {
	Total      int `json:"total"`
	Events     int `json:"events"`
	Activities int `json:"activities"`
	Gateways   int `json:"gateways"`
	Flows      int `json:"flows"`
}

// AgentWorkloadAnalysis analyzes agent assignment distribution
type AgentWorkloadAnalysis struct {
	AgentTasks      map[string][]string `json:"agent_tasks"`
	WorkloadBalance float64             `json:"workload_balance"`
	OverloadedAgents []string           `json:"overloaded_agents"`
	UnassignedTasks []string            `json:"unassigned_tasks"`
}

// Analyzer performs graph analysis on BPMN processes
type Analyzer struct {
	process *Process
	graph   map[string][]string // adjacency list representation
	reverse map[string][]string // reverse adjacency list
}

// NewAnalyzer creates a new analyzer for a process
func NewAnalyzer(process *Process) *Analyzer {
	a := &Analyzer{
		process: process,
		graph:   make(map[string][]string),
		reverse: make(map[string][]string),
	}
	a.buildGraph()
	return a
}

// buildGraph constructs adjacency lists from the process
func (a *Analyzer) buildGraph() {
	// Add all elements to the graph
	for _, e := range a.process.ProcessInfo.Elements.Events {
		a.graph[e.ID] = []string{}
		a.reverse[e.ID] = []string{}
	}
	for _, act := range a.process.ProcessInfo.Elements.Activities {
		a.graph[act.ID] = []string{}
		a.reverse[act.ID] = []string{}
	}
	for _, g := range a.process.ProcessInfo.Elements.Gateways {
		a.graph[g.ID] = []string{}
		a.reverse[g.ID] = []string{}
	}

	// Build adjacency lists from sequence flows
	for _, flow := range a.process.ProcessInfo.Elements.SequenceFlows {
		a.graph[flow.SourceRef] = append(a.graph[flow.SourceRef], flow.TargetRef)
		a.reverse[flow.TargetRef] = append(a.reverse[flow.TargetRef], flow.SourceRef)
	}
}

// Analyze performs comprehensive graph analysis
func (a *Analyzer) Analyze() *AnalysisResult {
	return &AnalysisResult{
		Reachability:  a.analyzeReachability(),
		Deadlocks:     a.detectDeadlocks(),
		Paths:         a.analyzePaths(),
		Metrics:       a.calculateMetrics(),
		AgentWorkload: a.analyzeAgentWorkload(),
	}
}

// analyzeReachability checks element reachability
func (a *Analyzer) analyzeReachability() ReachabilityAnalysis {
	result := ReachabilityAnalysis{
		ReachableFromStart: make(map[string]bool),
		ReachesEnd:         make(map[string]bool),
	}

	// Find start and end events
	startEvents := a.findStartEvents()
	endEvents := a.findEndEvents()

	// Forward reachability from start events
	for _, start := range startEvents {
		visited := a.dfs(start, a.graph)
		for v := range visited {
			result.ReachableFromStart[v] = true
		}
	}

	// Backward reachability to end events
	for _, end := range endEvents {
		visited := a.dfs(end, a.reverse)
		for v := range visited {
			result.ReachesEnd[v] = true
		}
	}

	// Find unreachable and dead-end elements
	for id := range a.graph {
		if !result.ReachableFromStart[id] {
			result.UnreachableElements = append(result.UnreachableElements, id)
		}
		if !result.ReachesEnd[id] {
			result.DeadEndElements = append(result.DeadEndElements, id)
		}
	}

	return result
}

// detectDeadlocks identifies potential deadlocks
func (a *Analyzer) detectDeadlocks() []DeadlockInfo {
	var deadlocks []DeadlockInfo

	// Check for missing join synchronization
	for _, gateway := range a.process.ProcessInfo.Elements.Gateways {
		if gateway.Type == "parallelGateway" && gateway.GatewayDirection == "converging" {
			incoming := a.reverse[gateway.ID]
			if len(incoming) < 2 {
				deadlocks = append(deadlocks, DeadlockInfo{
					Type:        "incomplete-join",
					Elements:    []string{gateway.ID},
					Description: fmt.Sprintf("Parallel join gateway '%s' has fewer than 2 incoming flows", gateway.Name),
				})
			} else {
				// Check if any incoming path is empty (direct connection from split to join)
				// This can cause deadlock if other paths fail
				for _, source := range incoming {
					// Check if this is a direct connection from a diverging gateway
					if srcGateway := a.findGateway(source); srcGateway != nil && 
					   srcGateway.Type == "parallelGateway" && 
					   srcGateway.GatewayDirection == "diverging" {
						deadlocks = append(deadlocks, DeadlockInfo{
							Type:        "incomplete-join",
							Elements:    []string{gateway.ID, source},
							Description: fmt.Sprintf("Parallel join gateway '%s' has direct connection from split gateway '%s' without intermediate activities", gateway.ID, source),
						})
					}
				}
			}
		}
	}

	// Check for exclusive gateway loops without exit conditions
	loops := a.findLoops()
	for _, loop := range loops {
		hasExit := false
		for _, elem := range loop.Elements {
			if gateway := a.findGateway(elem); gateway != nil && gateway.Type == "exclusiveGateway" {
				// Check if any outgoing flow leads outside the loop
				for _, target := range a.graph[elem] {
					if !contains(loop.Elements, target) {
						hasExit = true
						break
					}
				}
			}
		}
		if !hasExit {
			deadlocks = append(deadlocks, DeadlockInfo{
				Type:        "infinite-loop",
				Elements:    loop.Elements,
				Description: "Loop detected without clear exit condition",
			})
		}
	}

	return deadlocks
}

// analyzePaths analyzes all paths through the process
func (a *Analyzer) analyzePaths() PathAnalysis {
	result := PathAnalysis{
		AllPaths: [][]string{},
		Loops:    a.findLoops(),
	}

	startEvents := a.findStartEvents()
	endEvents := a.findEndEvents()

	// Find all paths from start to end
	for _, start := range startEvents {
		for _, end := range endEvents {
			paths := a.findAllPaths(start, end, []string{}, make(map[string]bool))
			result.AllPaths = append(result.AllPaths, paths...)
		}
	}

	// Calculate path metrics
	if len(result.AllPaths) > 0 {
		result.MinPathLength = len(result.AllPaths[0])
		result.MaxPathLength = len(result.AllPaths[0])
		totalLength := 0

		for _, path := range result.AllPaths {
			length := len(path)
			if length < result.MinPathLength {
				result.MinPathLength = length
			}
			if length > result.MaxPathLength {
				result.MaxPathLength = length
			}
			totalLength += length
		}

		result.AveragePathLength = float64(totalLength) / float64(len(result.AllPaths))
		
		// Critical path is the longest path
		for _, path := range result.AllPaths {
			if len(path) == result.MaxPathLength {
				result.CriticalPath = path
				break
			}
		}
	}

	result.LoopDetected = len(result.Loops) > 0
	return result
}

// calculateMetrics calculates process complexity metrics
func (a *Analyzer) calculateMetrics() ProcessMetrics {
	metrics := ProcessMetrics{
		Elements: ElementCount{
			Events:     len(a.process.ProcessInfo.Elements.Events),
			Activities: len(a.process.ProcessInfo.Elements.Activities),
			Gateways:   len(a.process.ProcessInfo.Elements.Gateways),
			Flows:      len(a.process.ProcessInfo.Elements.SequenceFlows),
		},
	}
	
	metrics.Elements.Total = metrics.Elements.Events + 
		metrics.Elements.Activities + 
		metrics.Elements.Gateways

	// Calculate complexity for BPMN processes
	// Complexity = V + E - 2N (V=vertices, E=edges, N=connected components)
	vertices := metrics.Elements.Total
	edges := len(a.process.ProcessInfo.Elements.SequenceFlows)
	components := a.countConnectedComponents()
	metrics.Complexity = vertices + edges - 2*components

	// Calculate depth and width
	metrics.Depth = a.calculateDepth()
	metrics.Width = a.calculateWidth()

	// Calculate connectivity (edges/vertices ratio)
	if vertices > 0 {
		metrics.Connectivity = float64(edges) / float64(vertices)
	}

	return metrics
}

// analyzeAgentWorkload analyzes agent task distribution
func (a *Analyzer) analyzeAgentWorkload() AgentWorkloadAnalysis {
	result := AgentWorkloadAnalysis{
		AgentTasks: make(map[string][]string),
	}

	// Count tasks per agent
	agentCounts := make(map[string]int)
	for _, activity := range a.process.ProcessInfo.Elements.Activities {
		if activity.Agent != nil && activity.Agent.ID != "" {
			agentID := activity.Agent.ID
			result.AgentTasks[agentID] = append(result.AgentTasks[agentID], activity.ID)
			agentCounts[agentID]++
		} else {
			result.UnassignedTasks = append(result.UnassignedTasks, activity.ID)
		}
	}

	// Calculate workload balance (standard deviation)
	if len(agentCounts) > 1 {
		counts := make([]float64, 0, len(agentCounts))
		total := 0.0
		for _, count := range agentCounts {
			counts = append(counts, float64(count))
			total += float64(count)
		}
		mean := total / float64(len(counts))
		
		variance := 0.0
		for _, count := range counts {
			diff := count - mean
			variance += diff * diff
		}
		variance /= float64(len(counts))
		
		// Balance score: 1.0 = perfect balance, 0.0 = worst imbalance
		if mean > 0 {
			result.WorkloadBalance = 1.0 - (variance / (mean * mean))
		}

		// Identify overloaded agents (> 1.5x average)
		for agent, count := range agentCounts {
			if float64(count) > mean*1.5 {
				result.OverloadedAgents = append(result.OverloadedAgents, agent)
			}
		}
	}

	return result
}

// Helper functions

func (a *Analyzer) findStartEvents() []string {
	var starts []string
	for _, event := range a.process.ProcessInfo.Elements.Events {
		if event.Type == "startEvent" {
			starts = append(starts, event.ID)
		}
	}
	return starts
}

func (a *Analyzer) findEndEvents() []string {
	var ends []string
	for _, event := range a.process.ProcessInfo.Elements.Events {
		if event.Type == "endEvent" {
			ends = append(ends, event.ID)
		}
	}
	return ends
}

func (a *Analyzer) dfs(start string, graph map[string][]string) map[string]bool {
	visited := make(map[string]bool)
	stack := []string{start}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[node] {
			continue
		}
		visited[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				stack = append(stack, neighbor)
			}
		}
	}

	return visited
}

func (a *Analyzer) findLoops() []Loop {
	var loops []Loop
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for node := range a.graph {
		if !visited[node] {
			a.findLoopsDFS(node, visited, recursionStack, []string{}, &loops)
		}
	}

	return loops
}

func (a *Analyzer) findLoopsDFS(node string, visited, stack map[string]bool, path []string, loops *[]Loop) {
	visited[node] = true
	stack[node] = true
	path = append(path, node)

	for _, neighbor := range a.graph[node] {
		if !visited[neighbor] {
			a.findLoopsDFS(neighbor, visited, stack, path, loops)
		} else if stack[neighbor] {
			// Found a loop
			loopStart := -1
			for i, n := range path {
				if n == neighbor {
					loopStart = i
					break
				}
			}
			if loopStart >= 0 {
				loop := Loop{
					Elements: append([]string{}, path[loopStart:]...),
					Type:     "simple",
				}
				*loops = append(*loops, loop)
			}
		}
	}

	stack[node] = false
}

func (a *Analyzer) findAllPaths(start, end string, path []string, visited map[string]bool) [][]string {
	path = append(path, start)
	
	if start == end {
		return [][]string{append([]string{}, path...)}
	}

	visited[start] = true
	defer func() { visited[start] = false }()

	var allPaths [][]string
	for _, next := range a.graph[start] {
		if !visited[next] {
			paths := a.findAllPaths(next, end, path, visited)
			allPaths = append(allPaths, paths...)
		}
	}

	return allPaths
}

func (a *Analyzer) countConnectedComponents() int {
	// For BPMN processes, we typically have a single connected component
	// as all elements should be reachable from start to end.
	// We'll check if all nodes are in one connected component by
	// treating the graph as undirected for this calculation.
	
	visited := make(map[string]bool)
	components := 0

	// Build undirected graph for connected component analysis
	undirected := make(map[string][]string)
	for node := range a.graph {
		undirected[node] = []string{}
	}
	
	// Add edges in both directions
	for source, targets := range a.graph {
		for _, target := range targets {
			undirected[source] = append(undirected[source], target)
			undirected[target] = append(undirected[target], source)
		}
	}

	// Get all nodes in sorted order for deterministic results
	nodes := make([]string, 0, len(undirected))
	for node := range undirected {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if !visited[node] {
			components++
			componentNodes := a.dfs(node, undirected)
			for n := range componentNodes {
				visited[n] = true
			}
		}
	}

	return components
}

func (a *Analyzer) calculateDepth() int {
	startEvents := a.findStartEvents()
	maxDepth := 0

	for _, start := range startEvents {
		depth := a.calculateMaxDepthFrom(start, make(map[string]bool))
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

func (a *Analyzer) calculateMaxDepthFrom(node string, visited map[string]bool) int {
	if visited[node] {
		return 0
	}
	visited[node] = true
	defer func() { visited[node] = false }()

	maxChildDepth := 0
	for _, child := range a.graph[node] {
		depth := a.calculateMaxDepthFrom(child, visited)
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}

	return maxChildDepth + 1
}

func (a *Analyzer) calculateWidth() int {
	// Width is the maximum number of parallel activities
	levels := make(map[int][]string)
	visited := make(map[string]bool)
	
	startEvents := a.findStartEvents()
	for _, start := range startEvents {
		a.assignLevels(start, 0, visited, levels)
	}

	maxWidth := 0
	for _, nodes := range levels {
		if len(nodes) > maxWidth {
			maxWidth = len(nodes)
		}
	}

	return maxWidth
}

func (a *Analyzer) assignLevels(node string, level int, visited map[string]bool, levels map[int][]string) {
	if visited[node] {
		return
	}
	visited[node] = true
	
	levels[level] = append(levels[level], node)
	
	for _, child := range a.graph[node] {
		a.assignLevels(child, level+1, visited, levels)
	}
}

func (a *Analyzer) findGateway(id string) *Gateway {
	for _, g := range a.process.ProcessInfo.Elements.Gateways {
		if g.ID == id {
			return &g
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FormatAnalysisReport creates a human-readable analysis report
func FormatAnalysisReport(result *AnalysisResult) string {
	var report strings.Builder

	report.WriteString("=== BPMN Process Analysis Report ===\n\n")

	// Metrics
	report.WriteString("Process Metrics:\n")
	report.WriteString(fmt.Sprintf("  Total Elements: %d\n", result.Metrics.Elements.Total))
	report.WriteString(fmt.Sprintf("  - Events: %d\n", result.Metrics.Elements.Events))
	report.WriteString(fmt.Sprintf("  - Activities: %d\n", result.Metrics.Elements.Activities))
	report.WriteString(fmt.Sprintf("  - Gateways: %d\n", result.Metrics.Elements.Gateways))
	report.WriteString(fmt.Sprintf("  - Sequence Flows: %d\n", result.Metrics.Elements.Flows))
	report.WriteString(fmt.Sprintf("  Complexity Score: %d\n", result.Metrics.Complexity))
	report.WriteString(fmt.Sprintf("  Process Depth: %d\n", result.Metrics.Depth))
	report.WriteString(fmt.Sprintf("  Process Width: %d\n", result.Metrics.Width))
	report.WriteString(fmt.Sprintf("  Connectivity: %.2f\n\n", result.Metrics.Connectivity))

	// Reachability
	report.WriteString("Reachability Analysis:\n")
	if len(result.Reachability.UnreachableElements) > 0 {
		report.WriteString("  ⚠️  Unreachable Elements:\n")
		for _, elem := range result.Reachability.UnreachableElements {
			report.WriteString(fmt.Sprintf("    - %s\n", elem))
		}
	} else {
		report.WriteString("  ✓ All elements are reachable from start\n")
	}
	
	if len(result.Reachability.DeadEndElements) > 0 {
		report.WriteString("  ⚠️  Dead-end Elements:\n")
		for _, elem := range result.Reachability.DeadEndElements {
			report.WriteString(fmt.Sprintf("    - %s\n", elem))
		}
	} else {
		report.WriteString("  ✓ All elements can reach an end event\n")
	}
	report.WriteString("\n")

	// Deadlocks
	report.WriteString("Deadlock Detection:\n")
	if len(result.Deadlocks) > 0 {
		report.WriteString("  ⚠️  Potential Deadlocks Found:\n")
		for _, deadlock := range result.Deadlocks {
			report.WriteString(fmt.Sprintf("    - Type: %s\n", deadlock.Type))
			report.WriteString(fmt.Sprintf("      Elements: %v\n", deadlock.Elements))
			report.WriteString(fmt.Sprintf("      Description: %s\n", deadlock.Description))
		}
	} else {
		report.WriteString("  ✓ No deadlocks detected\n")
	}
	report.WriteString("\n")

	// Path Analysis
	report.WriteString("Path Analysis:\n")
	report.WriteString(fmt.Sprintf("  Total Paths: %d\n", len(result.Paths.AllPaths)))
	if len(result.Paths.AllPaths) > 0 {
		report.WriteString(fmt.Sprintf("  Shortest Path Length: %d\n", result.Paths.MinPathLength))
		report.WriteString(fmt.Sprintf("  Longest Path Length: %d\n", result.Paths.MaxPathLength))
		report.WriteString(fmt.Sprintf("  Average Path Length: %.2f\n", result.Paths.AveragePathLength))
		if len(result.Paths.CriticalPath) > 0 {
			report.WriteString(fmt.Sprintf("  Critical Path: %v\n", result.Paths.CriticalPath))
		}
	}
	
	if result.Paths.LoopDetected {
		report.WriteString(fmt.Sprintf("  ⚠️  Loops Detected: %d\n", len(result.Paths.Loops)))
		for i, loop := range result.Paths.Loops {
			report.WriteString(fmt.Sprintf("    Loop %d: %v\n", i+1, loop.Elements))
		}
	} else {
		report.WriteString("  ✓ No loops detected\n")
	}
	report.WriteString("\n")

	// Agent Workload
	report.WriteString("Agent Workload Analysis:\n")
	if len(result.AgentWorkload.AgentTasks) > 0 {
		report.WriteString("  Task Distribution:\n")
		for agent, tasks := range result.AgentWorkload.AgentTasks {
			report.WriteString(fmt.Sprintf("    - %s: %d tasks\n", agent, len(tasks)))
		}
		report.WriteString(fmt.Sprintf("  Workload Balance Score: %.2f\n", result.AgentWorkload.WorkloadBalance))
		
		if len(result.AgentWorkload.OverloadedAgents) > 0 {
			report.WriteString("  ⚠️  Overloaded Agents:\n")
			for _, agent := range result.AgentWorkload.OverloadedAgents {
				report.WriteString(fmt.Sprintf("    - %s\n", agent))
			}
		}
	}
	
	if len(result.AgentWorkload.UnassignedTasks) > 0 {
		report.WriteString(fmt.Sprintf("  ⚠️  Unassigned Tasks: %d\n", len(result.AgentWorkload.UnassignedTasks)))
	}

	return report.String()
}