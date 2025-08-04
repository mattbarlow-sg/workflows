package bpmn

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// AgentManager manages agent assignment and review workflows
type AgentManager struct {
	agents          map[string]*Agent
	assignmentRules []AssignmentRule
	reviewWorkflows map[string]*ReviewWorkflow
	metrics         map[string]*AgentMetrics
}

// AgentMetrics tracks performance metrics for an agent
type AgentMetrics struct {
	TasksAssigned   int           `json:"tasks_assigned"`
	TasksCompleted  int           `json:"tasks_completed"`
	AverageTaskTime time.Duration `json:"average_task_time"`
	CurrentLoad     int           `json:"current_load"`
	LastAssignment  time.Time     `json:"last_assignment"`
}

// AssignmentContext provides context for agent assignment decisions
type AssignmentContext struct {
	Activity      *Activity              `json:"activity"`
	ProcessData   map[string]interface{} `json:"process_data"`
	CurrentAgents map[string]*Agent      `json:"current_agents"`
	Metrics       map[string]*AgentMetrics `json:"metrics"`
}

// AssignmentResult contains the result of an agent assignment
type AssignmentResult struct {
	AssignedAgent *Agent   `json:"assigned_agent"`
	Reason        string   `json:"reason"`
	Alternatives  []*Agent `json:"alternatives"`
	Score         float64  `json:"score"`
}

// ReviewResult contains the result of a review workflow
type ReviewResult struct {
	Approved     bool                   `json:"approved"`
	ReviewerID   string                 `json:"reviewer_id"`
	Comments     string                 `json:"comments"`
	Timestamp    time.Time              `json:"timestamp"`
	ReviewData   map[string]interface{} `json:"review_data"`
}

// NewAgentManager creates a new agent manager
func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents:          make(map[string]*Agent),
		assignmentRules: []AssignmentRule{},
		reviewWorkflows: make(map[string]*ReviewWorkflow),
		metrics:         make(map[string]*AgentMetrics),
	}
}

// RegisterAgent registers an agent with the manager
func (am *AgentManager) RegisterAgent(agent *Agent) error {
	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}
	
	am.agents[agent.ID] = agent
	am.metrics[agent.ID] = &AgentMetrics{
		LastAssignment: time.Now(),
	}
	
	return nil
}

// RegisterAssignmentRule registers a custom assignment rule
func (am *AgentManager) RegisterAssignmentRule(rule AssignmentRule) {
	am.assignmentRules = append(am.assignmentRules, rule)
}

// RegisterReviewWorkflow registers a review workflow
func (am *AgentManager) RegisterReviewWorkflow(id string, workflow *ReviewWorkflow) {
	am.reviewWorkflows[id] = workflow
}

// AssignAgent assigns an agent to an activity based on rules and context
func (am *AgentManager) AssignAgent(ctx AssignmentContext) (*AssignmentResult, error) {
	result := &AssignmentResult{
		Alternatives: []*Agent{},
	}

	// Check if activity has a specific agent assigned
	if ctx.Activity.Agent != nil && ctx.Activity.Agent.ID != "" {
		if agent, exists := am.agents[ctx.Activity.Agent.ID]; exists {
			result.AssignedAgent = agent
			result.Reason = "explicitly assigned"
			result.Score = 1.0
			return result, nil
		}
	}

	// Get eligible agents based on capabilities
	eligibleAgents := am.getEligibleAgents(ctx)
	if len(eligibleAgents) == 0 {
		return nil, fmt.Errorf("no eligible agents found for activity %s", ctx.Activity.ID)
	}

	// Apply assignment rules
	for _, rule := range am.assignmentRules {
		if rule.Applies(ctx) {
			agent, score := rule.SelectAgent(eligibleAgents, ctx)
			if agent != nil {
				result.AssignedAgent = agent
				result.Reason = rule.Name()
				result.Score = score
				result.Alternatives = am.getAlternatives(eligibleAgents, agent, 3)
				return result, nil
			}
		}
	}

	// Default assignment strategy (load balancing)
	agent := am.selectByLoadBalancing(eligibleAgents)
	result.AssignedAgent = agent
	result.Reason = "load balancing"
	result.Score = am.calculateLoadScore(agent)
	result.Alternatives = am.getAlternatives(eligibleAgents, agent, 3)

	// Update metrics
	am.updateAssignmentMetrics(agent.ID)

	return result, nil
}

// getEligibleAgents returns agents that can handle the activity
func (am *AgentManager) getEligibleAgents(ctx AssignmentContext) []*Agent {
	var eligible []*Agent

	for _, agent := range am.agents {
		if am.isAgentEligible(agent, ctx) {
			eligible = append(eligible, agent)
		}
	}

	return eligible
}

// isAgentEligible checks if an agent can handle an activity
func (am *AgentManager) isAgentEligible(agent *Agent, ctx AssignmentContext) bool {
	// Check agent type if specified
	if ctx.Activity.Agent != nil && ctx.Activity.Agent.Type != "" {
		if agent.Type != ctx.Activity.Agent.Type {
			return false
		}
	}

	// Check required capabilities
	if ctx.Activity.Agent != nil && len(ctx.Activity.Agent.Capabilities) > 0 {
		for _, required := range ctx.Activity.Agent.Capabilities {
			found := false
			for _, capability := range agent.Capabilities {
				if capability == required {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check constraints
	if !am.checkConstraints(agent, ctx) {
		return false
	}

	// Check availability
	if agent.Availability != nil && !agent.Availability.Available {
		return false
	}

	return true
}

// checkConstraints checks if agent constraints are satisfied
func (am *AgentManager) checkConstraints(agent *Agent, ctx AssignmentContext) bool {
	if agent.Constraints == nil {
		return true
	}

	// Check max concurrent tasks
	metrics := am.metrics[agent.ID]
	if agent.Constraints.MaxConcurrentTasks > 0 && 
	   metrics.CurrentLoad >= agent.Constraints.MaxConcurrentTasks {
		return false
	}

	// Check allowed process types
	if len(agent.Constraints.AllowedProcessTypes) > 0 {
		// Would need process type from context
		// For now, assume it's allowed
	}

	// Check time constraints
	if agent.Constraints.TimeConstraints != nil {
		now := time.Now()
		tc := agent.Constraints.TimeConstraints
		
		// Check business hours
		if tc.BusinessHoursOnly && !isBusinessHours(now) {
			return false
		}
		
		// Check availability windows
		if len(tc.AvailabilityWindows) > 0 {
			inWindow := false
			for _, window := range tc.AvailabilityWindows {
				if isInWindow(now, window) {
					inWindow = true
					break
				}
			}
			if !inWindow {
				return false
			}
		}
	}

	return true
}

// selectByLoadBalancing selects agent with lowest current load
func (am *AgentManager) selectByLoadBalancing(agents []*Agent) *Agent {
	if len(agents) == 0 {
		return nil
	}

	// Sort by current load
	sort.Slice(agents, func(i, j int) bool {
		loadI := am.metrics[agents[i].ID].CurrentLoad
		loadJ := am.metrics[agents[j].ID].CurrentLoad
		return loadI < loadJ
	})

	// Return agent with lowest load
	return agents[0]
}

// calculateLoadScore calculates a load balance score (0-1, higher is better)
func (am *AgentManager) calculateLoadScore(agent *Agent) float64 {
	metrics := am.metrics[agent.ID]
	if agent.Constraints == nil || agent.Constraints.MaxConcurrentTasks == 0 {
		return 1.0
	}
	
	load := float64(metrics.CurrentLoad) / float64(agent.Constraints.MaxConcurrentTasks)
	return 1.0 - load
}

// getAlternatives returns alternative agents
func (am *AgentManager) getAlternatives(eligible []*Agent, selected *Agent, count int) []*Agent {
	var alternatives []*Agent
	
	for _, agent := range eligible {
		if agent.ID != selected.ID {
			alternatives = append(alternatives, agent)
			if len(alternatives) >= count {
				break
			}
		}
	}
	
	return alternatives
}

// updateAssignmentMetrics updates agent metrics after assignment
func (am *AgentManager) updateAssignmentMetrics(agentID string) {
	metrics := am.metrics[agentID]
	metrics.TasksAssigned++
	metrics.CurrentLoad++
	metrics.LastAssignment = time.Now()
}

// CompleteTask updates metrics when a task is completed
func (am *AgentManager) CompleteTask(agentID string, duration time.Duration) {
	metrics := am.metrics[agentID]
	metrics.TasksCompleted++
	metrics.CurrentLoad--
	
	// Update average task time
	if metrics.TasksCompleted == 1 {
		metrics.AverageTaskTime = duration
	} else {
		// Moving average
		metrics.AverageTaskTime = time.Duration(
			(int64(metrics.AverageTaskTime)*(int64(metrics.TasksCompleted)-1) + int64(duration)) / 
			int64(metrics.TasksCompleted),
		)
	}
}

// ProcessReview processes a review workflow
func (am *AgentManager) ProcessReview(workflowID string, data map[string]interface{}) (*ReviewResult, error) {
	workflow, exists := am.reviewWorkflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("review workflow %s not found", workflowID)
	}

	// Select reviewer based on workflow configuration
	reviewer, err := am.selectReviewer(workflow, data)
	if err != nil {
		return nil, err
	}

	// Simulate review process (in real implementation, this would be async)
	result := &ReviewResult{
		ReviewerID: reviewer.ID,
		Timestamp:  time.Now(),
		ReviewData: data,
	}

	// Apply review rules
	approved := true
	comments := []string{}

	if workflow.Rules != nil {
		// Check approval threshold
		if workflow.Rules.ApprovalThreshold > 0 {
			if score, ok := data["score"].(float64); ok {
				if score < workflow.Rules.ApprovalThreshold {
					approved = false
					comments = append(comments, fmt.Sprintf("Score %.2f below threshold %.2f", 
						score, workflow.Rules.ApprovalThreshold))
				}
			}
		}

		// Check required fields
		for _, field := range workflow.Rules.RequiredFields {
			if _, exists := data[field]; !exists {
				approved = false
				comments = append(comments, fmt.Sprintf("Required field '%s' missing", field))
			}
		}
	}

	result.Approved = approved
	result.Comments = strings.Join(comments, "; ")

	return result, nil
}

// selectReviewer selects a reviewer for the workflow
func (am *AgentManager) selectReviewer(workflow *ReviewWorkflow, data map[string]interface{}) (*Agent, error) {
	// Get eligible reviewers
	var eligibleReviewers []*Agent
	
	for _, agent := range am.agents {
		if am.canReview(agent, workflow) {
			eligibleReviewers = append(eligibleReviewers, agent)
		}
	}

	if len(eligibleReviewers) == 0 {
		return nil, fmt.Errorf("no eligible reviewers found")
	}

	// Select based on workflow type
	switch workflow.Type {
	case "peer_review":
		// Random selection for peer review
		return eligibleReviewers[rand.Intn(len(eligibleReviewers))], nil
	case "hierarchical":
		// Select by seniority (would need seniority data)
		return eligibleReviewers[0], nil
	case "specialist":
		// Select by expertise (would need expertise matching)
		return eligibleReviewers[0], nil
	default:
		// Default to first eligible
		return eligibleReviewers[0], nil
	}
}

// canReview checks if an agent can review for a workflow
func (am *AgentManager) canReview(agent *Agent, workflow *ReviewWorkflow) bool {
	// Check if agent has review capability
	hasReviewCapability := false
	for _, cap := range agent.Capabilities {
		if cap == "review" || cap == workflow.Type {
			hasReviewCapability = true
			break
		}
	}

	return hasReviewCapability
}

// GetAgentWorkloadReport generates a workload report
func (am *AgentManager) GetAgentWorkloadReport() string {
	var report strings.Builder
	
	report.WriteString("=== Agent Workload Report ===\n\n")
	
	// Sort agents by current load
	var agents []*Agent
	for _, agent := range am.agents {
		agents = append(agents, agent)
	}
	
	sort.Slice(agents, func(i, j int) bool {
		return am.metrics[agents[i].ID].CurrentLoad > am.metrics[agents[j].ID].CurrentLoad
	})
	
	// Generate report
	for _, agent := range agents {
		metrics := am.metrics[agent.ID]
		report.WriteString(fmt.Sprintf("Agent: %s (%s)\n", agent.Name, agent.ID))
		report.WriteString(fmt.Sprintf("  Type: %s\n", agent.Type))
		report.WriteString(fmt.Sprintf("  Current Load: %d", metrics.CurrentLoad))
		
		if agent.Constraints != nil && agent.Constraints.MaxConcurrentTasks > 0 {
			report.WriteString(fmt.Sprintf("/%d", agent.Constraints.MaxConcurrentTasks))
		}
		report.WriteString("\n")
		
		report.WriteString(fmt.Sprintf("  Tasks Assigned: %d\n", metrics.TasksAssigned))
		report.WriteString(fmt.Sprintf("  Tasks Completed: %d\n", metrics.TasksCompleted))
		
		if metrics.AverageTaskTime > 0 {
			report.WriteString(fmt.Sprintf("  Average Task Time: %v\n", metrics.AverageTaskTime))
		}
		
		report.WriteString(fmt.Sprintf("  Last Assignment: %v\n", metrics.LastAssignment.Format(time.RFC3339)))
		report.WriteString("\n")
	}
	
	return report.String()
}

// Helper functions

func isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	weekday := t.Weekday()
	
	// Monday-Friday, 9 AM - 5 PM
	return weekday >= time.Monday && weekday <= time.Friday &&
		hour >= 9 && hour < 17
}

func isInWindow(t time.Time, window TimeWindow) bool {
	// Simple time window check
	// In real implementation, would parse and check window format
	return true
}

// Built-in Assignment Rules

// CapabilityMatchRule assigns based on best capability match
type CapabilityMatchRule struct{}

func (r CapabilityMatchRule) Name() string {
	return "capability_match"
}

func (r CapabilityMatchRule) Applies(ctx AssignmentContext) bool {
	return ctx.Activity.Agent != nil && len(ctx.Activity.Agent.Capabilities) > 0
}

func (r CapabilityMatchRule) SelectAgent(agents []*Agent, ctx AssignmentContext) (*Agent, float64) {
	requiredCaps := ctx.Activity.Agent.Capabilities
	var bestAgent *Agent
	bestScore := 0.0

	for _, agent := range agents {
		score := r.calculateCapabilityScore(agent, requiredCaps)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	return bestAgent, bestScore
}

func (r CapabilityMatchRule) calculateCapabilityScore(agent *Agent, required []string) float64 {
	matches := 0
	for _, req := range required {
		for _, cap := range agent.Capabilities {
			if cap == req {
				matches++
				break
			}
		}
	}
	
	if len(required) == 0 {
		return 1.0
	}
	
	return float64(matches) / float64(len(required))
}

// RoundRobinRule assigns agents in round-robin fashion
type RoundRobinRule struct {
	lastAssigned map[string]int
}

func NewRoundRobinRule() *RoundRobinRule {
	return &RoundRobinRule{
		lastAssigned: make(map[string]int),
	}
}

func (r *RoundRobinRule) Name() string {
	return "round_robin"
}

func (r *RoundRobinRule) Applies(ctx AssignmentContext) bool {
	// Apply to activities without specific assignment
	return ctx.Activity.Agent == nil || ctx.Activity.Agent.ID == ""
}

func (r *RoundRobinRule) SelectAgent(agents []*Agent, ctx AssignmentContext) (*Agent, float64) {
	if len(agents) == 0 {
		return nil, 0
	}

	processID := "default" // Would get from context
	lastIdx := r.lastAssigned[processID]
	nextIdx := (lastIdx + 1) % len(agents)
	
	r.lastAssigned[processID] = nextIdx
	return agents[nextIdx], 1.0
}

// RandomAssignmentRule assigns agents randomly
type RandomAssignmentRule struct{}

func (r RandomAssignmentRule) Name() string {
	return "random"
}

func (r RandomAssignmentRule) Applies(ctx AssignmentContext) bool {
	return true // Can always apply as fallback
}

func (r RandomAssignmentRule) SelectAgent(agents []*Agent, ctx AssignmentContext) (*Agent, float64) {
	if len(agents) == 0 {
		return nil, 0
	}
	
	idx := rand.Intn(len(agents))
	return agents[idx], 0.5 // Medium confidence score
}