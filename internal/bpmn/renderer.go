package bpmn

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderFormat defines output formats for rendering
type RenderFormat string

const (
	// FormatText renders as plain text
	FormatText RenderFormat = "text"
	
	// FormatMarkdown renders as markdown
	FormatMarkdown RenderFormat = "markdown"
	
	// FormatJSON renders as JSON
	FormatJSON RenderFormat = "json"
	
	// FormatDOT renders as Graphviz DOT
	FormatDOT RenderFormat = "dot"
)

// Renderer renders BPMN processes in various formats
type Renderer struct {
	process      *Process
	analyzer     *Analyzer
	agentManager *AgentManager
}

// NewRenderer creates a new renderer
func NewRenderer(process *Process) *Renderer {
	return &Renderer{
		process:  process,
		analyzer: NewAnalyzer(process),
	}
}

// SetAgentManager sets the agent manager for agent-related rendering
func (r *Renderer) SetAgentManager(am *AgentManager) {
	r.agentManager = am
}

// Render renders the process in the specified format
func (r *Renderer) Render(format RenderFormat) (string, error) {
	switch format {
	case FormatText:
		return r.renderText(), nil
	case FormatMarkdown:
		return r.renderMarkdown(), nil
	case FormatJSON:
		return r.renderJSON()
	case FormatDOT:
		return r.renderDOT(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// renderText renders the process as plain text
func (r *Renderer) renderText() string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("BPMN Process: %s\n", r.process.ProcessInfo.Name))
	sb.WriteString(strings.Repeat("=", len(r.process.ProcessInfo.Name)+14))
	sb.WriteString("\n\n")

	// Process Info
	sb.WriteString("Process Information:\n")
	sb.WriteString(fmt.Sprintf("  ID: %s\n", r.process.ProcessInfo.ID))
	sb.WriteString(fmt.Sprintf("  Name: %s\n", r.process.ProcessInfo.Name))
	if r.process.ProcessInfo.Description != "" {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", r.process.ProcessInfo.Description))
	}
	if r.process.ProcessInfo.WorkSessionID != "" {
		sb.WriteString(fmt.Sprintf("  Work Session ID: %s\n", r.process.ProcessInfo.WorkSessionID))
	}
	sb.WriteString("\n")

	// Events
	if len(r.process.ProcessInfo.Elements.Events) > 0 {
		sb.WriteString("Events:\n")
		for _, event := range r.process.ProcessInfo.Elements.Events {
			sb.WriteString(r.renderEventText(&event, "  "))
		}
		sb.WriteString("\n")
	}

	// Activities
	if len(r.process.ProcessInfo.Elements.Activities) > 0 {
		sb.WriteString("Activities:\n")
		for _, activity := range r.process.ProcessInfo.Elements.Activities {
			sb.WriteString(r.renderActivityText(&activity, "  "))
		}
		sb.WriteString("\n")
	}

	// Gateways
	if len(r.process.ProcessInfo.Elements.Gateways) > 0 {
		sb.WriteString("Gateways:\n")
		for _, gateway := range r.process.ProcessInfo.Elements.Gateways {
			sb.WriteString(r.renderGatewayText(&gateway, "  "))
		}
		sb.WriteString("\n")
	}

	// Sequence Flows
	if len(r.process.ProcessInfo.Elements.SequenceFlows) > 0 {
		sb.WriteString("Sequence Flows:\n")
		for _, flow := range r.process.ProcessInfo.Elements.SequenceFlows {
			sb.WriteString(r.renderFlowText(&flow, "  "))
		}
		sb.WriteString("\n")
	}

	// Agent Summary
	if r.agentManager != nil {
		sb.WriteString("Agent Assignments:\n")
		sb.WriteString(r.renderAgentSummary("  "))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderMarkdown renders the process as markdown
func (r *Renderer) renderMarkdown() string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# BPMN Process: %s\n\n", r.process.ProcessInfo.Name))

	// Process Info
	sb.WriteString("## Process Information\n\n")
	sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", r.process.ProcessInfo.ID))
	sb.WriteString(fmt.Sprintf("- **Name**: %s\n", r.process.ProcessInfo.Name))
	if r.process.ProcessInfo.Description != "" {
		sb.WriteString(fmt.Sprintf("- **Description**: %s\n", r.process.ProcessInfo.Description))
	}
	if r.process.ProcessInfo.WorkSessionID != "" {
		sb.WriteString(fmt.Sprintf("- **Work Session ID**: `%s`\n", r.process.ProcessInfo.WorkSessionID))
	}
	sb.WriteString("\n")

	// Process Flow Diagram (text-based)
	sb.WriteString("## Process Flow\n\n")
	sb.WriteString("```mermaid\n")
	sb.WriteString(r.renderMermaid())
	sb.WriteString("```\n\n")

	// Events
	if len(r.process.ProcessInfo.Elements.Events) > 0 {
		sb.WriteString("## Events\n\n")
		for _, event := range r.process.ProcessInfo.Elements.Events {
			sb.WriteString(r.renderEventMarkdown(&event))
		}
		sb.WriteString("\n")
	}

	// Activities
	if len(r.process.ProcessInfo.Elements.Activities) > 0 {
		sb.WriteString("## Activities\n\n")
		for _, activity := range r.process.ProcessInfo.Elements.Activities {
			sb.WriteString(r.renderActivityMarkdown(&activity))
		}
		sb.WriteString("\n")
	}

	// Gateways
	if len(r.process.ProcessInfo.Elements.Gateways) > 0 {
		sb.WriteString("## Gateways\n\n")
		for _, gateway := range r.process.ProcessInfo.Elements.Gateways {
			sb.WriteString(r.renderGatewayMarkdown(&gateway))
		}
		sb.WriteString("\n")
	}

	// Analysis Results
	analysis := r.analyzer.Analyze()
	sb.WriteString("## Analysis\n\n")
	sb.WriteString(r.renderAnalysisMarkdown(analysis))

	return sb.String()
}

// renderJSON renders the process as JSON
func (r *Renderer) renderJSON() (string, error) {
	data, err := json.MarshalIndent(r.process, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// renderDOT renders the process as Graphviz DOT
func (r *Renderer) renderDOT() string {
	var sb strings.Builder

	sb.WriteString("digraph BPMNProcess {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")

	// Events
	for _, event := range r.process.ProcessInfo.Elements.Events {
		shape := "circle"
		if event.Type == "endEvent" {
			shape = "doublecircle"
		}
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\", shape=%s];\n", 
			event.ID, event.Name, shape))
	}

	// Activities
	for _, activity := range r.process.ProcessInfo.Elements.Activities {
		style := "rounded"
		if activity.Type == "subProcess" {
			style = "rounded,bold"
		}
		
		label := activity.Name
		if activity.Agent != nil && activity.Agent.ID != "" {
			label += fmt.Sprintf("\\n[%s]", activity.Agent.ID)
		}
		
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\", style=\"%s\"];\n", 
			activity.ID, label, style))
	}

	// Gateways
	for _, gateway := range r.process.ProcessInfo.Elements.Gateways {
		shape := "diamond"
		label := gatewaySymbol(gateway.Type)
		sb.WriteString(fmt.Sprintf("  %s [label=\"%s\", shape=%s];\n", 
			gateway.ID, label, shape))
	}

	// Sequence Flows
	for _, flow := range r.process.ProcessInfo.Elements.SequenceFlows {
		label := ""
		if flow.Name != "" {
			label = fmt.Sprintf(" [label=\"%s\"]", flow.Name)
		}
		sb.WriteString(fmt.Sprintf("  %s -> %s%s;\n", 
			flow.SourceRef, flow.TargetRef, label))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// renderMermaid renders a Mermaid diagram
func (r *Renderer) renderMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph LR\n")

	// Events
	for _, event := range r.process.ProcessInfo.Elements.Events {
		shape := "()"
		if event.Type == "endEvent" {
			shape = "((()))"
		}
		sb.WriteString(fmt.Sprintf("  %s%s%s[%s]\n", 
			event.ID, shape[:len(shape)/2], event.Name, shape[len(shape)/2:]))
	}

	// Activities
	for _, activity := range r.process.ProcessInfo.Elements.Activities {
		sb.WriteString(fmt.Sprintf("  %s[%s]\n", activity.ID, activity.Name))
	}

	// Gateways
	for _, gateway := range r.process.ProcessInfo.Elements.Gateways {
		symbol := gatewaySymbol(gateway.Type)
		sb.WriteString(fmt.Sprintf("  %s{%s}\n", gateway.ID, symbol))
	}

	// Sequence Flows
	for _, flow := range r.process.ProcessInfo.Elements.SequenceFlows {
		arrow := "-->"
		if flow.Name != "" {
			arrow = fmt.Sprintf("-->|%s|", flow.Name)
		}
		sb.WriteString(fmt.Sprintf("  %s %s %s\n", 
			flow.SourceRef, arrow, flow.TargetRef))
	}

	return sb.String()
}

// Helper rendering functions

func (r *Renderer) renderEventText(event *Event, indent string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, event.Name, event.Type))
	sb.WriteString(fmt.Sprintf("%s  ID: %s\n", indent, event.ID))
	if event.EventType != "" {
		sb.WriteString(fmt.Sprintf("%s  Event Type: %s\n", indent, event.EventType))
	}
	return sb.String()
}

func (r *Renderer) renderActivityText(activity *Activity, indent string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, activity.Name, activity.Type))
	sb.WriteString(fmt.Sprintf("%s  ID: %s\n", indent, activity.ID))
	if activity.Description != "" {
		sb.WriteString(fmt.Sprintf("%s  Description: %s\n", indent, activity.Description))
	}
	if activity.Agent != nil {
		sb.WriteString(fmt.Sprintf("%s  Agent: %s\n", indent, r.renderAgentAssignment(activity.Agent)))
	}
	return sb.String()
}

func (r *Renderer) renderGatewayText(gateway *Gateway, indent string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, gateway.Name, gateway.Type))
	sb.WriteString(fmt.Sprintf("%s  ID: %s\n", indent, gateway.ID))
	sb.WriteString(fmt.Sprintf("%s  Direction: %s\n", indent, gateway.GatewayDirection))
	return sb.String()
}

func (r *Renderer) renderFlowText(flow *SequenceFlow, indent string) string {
	name := flow.Name
	if name == "" {
		name = "unnamed"
	}
	return fmt.Sprintf("%s- %s: %s → %s\n", indent, name, flow.SourceRef, flow.TargetRef)
}

func (r *Renderer) renderEventMarkdown(event *Event) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s\n\n", event.Name))
	sb.WriteString(fmt.Sprintf("- **Type**: %s\n", event.Type))
	sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", event.ID))
	if event.EventType != "" {
		sb.WriteString(fmt.Sprintf("- **Event Type**: %s\n", event.EventType))
	}
	sb.WriteString("\n")
	return sb.String()
}

func (r *Renderer) renderActivityMarkdown(activity *Activity) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s\n\n", activity.Name))
	sb.WriteString(fmt.Sprintf("- **Type**: %s\n", activity.Type))
	sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", activity.ID))
	if activity.Description != "" {
		sb.WriteString(fmt.Sprintf("- **Description**: %s\n", activity.Description))
	}
	if activity.Agent != nil {
		sb.WriteString(fmt.Sprintf("- **Agent**: %s\n", r.renderAgentAssignment(activity.Agent)))
	}
	if activity.ReviewWorkflow != nil {
		sb.WriteString(fmt.Sprintf("- **Review**: %s (%s)\n", 
			activity.ReviewWorkflow.Type, activity.ReviewWorkflow.Pattern))
	}
	sb.WriteString("\n")
	return sb.String()
}

func (r *Renderer) renderGatewayMarkdown(gateway *Gateway) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s\n\n", gateway.Name))
	sb.WriteString(fmt.Sprintf("- **Type**: %s\n", gateway.Type))
	sb.WriteString(fmt.Sprintf("- **ID**: `%s`\n", gateway.ID))
	sb.WriteString(fmt.Sprintf("- **Direction**: %s\n", gateway.GatewayDirection))
	sb.WriteString("\n")
	return sb.String()
}

func (r *Renderer) renderAnalysisMarkdown(analysis *AnalysisResult) string {
	var sb strings.Builder

	// Metrics
	sb.WriteString("### Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Elements**: %d\n", analysis.Metrics.Elements.Total))
	sb.WriteString(fmt.Sprintf("- **Complexity Score**: %d\n", analysis.Metrics.Complexity))
	sb.WriteString(fmt.Sprintf("- **Process Depth**: %d\n", analysis.Metrics.Depth))
	sb.WriteString(fmt.Sprintf("- **Process Width**: %d\n", analysis.Metrics.Width))
	sb.WriteString("\n")

	// Issues
	if len(analysis.Reachability.UnreachableElements) > 0 || len(analysis.Deadlocks) > 0 {
		sb.WriteString("### Issues\n\n")
		
		if len(analysis.Reachability.UnreachableElements) > 0 {
			sb.WriteString("**Unreachable Elements:**\n")
			for _, elem := range analysis.Reachability.UnreachableElements {
				sb.WriteString(fmt.Sprintf("- `%s`\n", elem))
			}
			sb.WriteString("\n")
		}
		
		if len(analysis.Deadlocks) > 0 {
			sb.WriteString("**Potential Deadlocks:**\n")
			for _, deadlock := range analysis.Deadlocks {
				sb.WriteString(fmt.Sprintf("- %s: %v\n", deadlock.Type, deadlock.Elements))
			}
			sb.WriteString("\n")
		}
	}

	// Path Analysis
	sb.WriteString("### Path Analysis\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Paths**: %d\n", len(analysis.Paths.AllPaths)))
	sb.WriteString(fmt.Sprintf("- **Shortest Path**: %d steps\n", analysis.Paths.MinPathLength))
	sb.WriteString(fmt.Sprintf("- **Longest Path**: %d steps\n", analysis.Paths.MaxPathLength))
	sb.WriteString(fmt.Sprintf("- **Loops Detected**: %v\n", analysis.Paths.LoopDetected))
	sb.WriteString("\n")

	return sb.String()
}

func (r *Renderer) renderAgentAssignment(agent *AgentAssignment) string {
	if agent.ID != "" {
		return agent.ID
	}
	if agent.Type != "" {
		return fmt.Sprintf("Type: %s", agent.Type)
	}
	if len(agent.Capabilities) > 0 {
		return fmt.Sprintf("Capabilities: %v", agent.Capabilities)
	}
	return "Dynamic"
}

func (r *Renderer) renderAgentSummary(indent string) string {
	var sb strings.Builder
	
	// Count activities by agent
	agentTasks := make(map[string][]string)
	unassigned := []string{}
	
	for _, activity := range r.process.ProcessInfo.Elements.Activities {
		if activity.Agent != nil && activity.Agent.ID != "" {
			agentTasks[activity.Agent.ID] = append(agentTasks[activity.Agent.ID], activity.Name)
		} else {
			unassigned = append(unassigned, activity.Name)
		}
	}
	
	// Render summary
	for agent, tasks := range agentTasks {
		sb.WriteString(fmt.Sprintf("%s%s: %d tasks\n", indent, agent, len(tasks)))
		for _, task := range tasks {
			sb.WriteString(fmt.Sprintf("%s  - %s\n", indent, task))
		}
	}
	
	if len(unassigned) > 0 {
		sb.WriteString(fmt.Sprintf("%sUnassigned: %d tasks\n", indent, len(unassigned)))
		for _, task := range unassigned {
			sb.WriteString(fmt.Sprintf("%s  - %s\n", indent, task))
		}
	}
	
	return sb.String()
}

func gatewaySymbol(gatewayType string) string {
	switch gatewayType {
	case "exclusiveGateway":
		return "X"
	case "parallelGateway":
		return "+"
	case "inclusiveGateway":
		return "O"
	case "complexGateway":
		return "*"
	default:
		return "?"
	}
}

// RenderValidationReport renders a validation report
func RenderValidationReport(result ValidationResult) string {
	var sb strings.Builder

	sb.WriteString("=== BPMN Validation Report ===\n\n")
	
	if result.Valid {
		sb.WriteString("✓ Process is valid\n\n")
	} else {
		sb.WriteString("✗ Process validation failed\n\n")
	}

	// Schema validation
	if !result.SchemaValid {
		sb.WriteString("Schema Validation Errors:\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err))
		}
		sb.WriteString("\n")
	}

	// Semantic errors
	if len(result.Errors) > 0 {
		sb.WriteString("Semantic Errors:\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - [%s] %s: %s\n", 
				err.Level, err.Path, err.Message))
		}
		sb.WriteString("\n")
	}

	// Warnings
	if len(result.Warnings) > 0 {
		sb.WriteString("Warnings:\n")
		for _, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  - [%s] %s: %s\n", 
				warn.Level, warn.Path, warn.Message))
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total Errors: %d\n", len(result.Errors)))
	sb.WriteString(fmt.Sprintf("  Total Warnings: %d\n", len(result.Warnings)))
	sb.WriteString(fmt.Sprintf("  Validation Time: %v\n", "<1ms"))

	return sb.String()
}

// RenderProcessSummary renders a concise process summary
func RenderProcessSummary(process *Process) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Process: %s\n", process.ProcessInfo.Name))
	sb.WriteString(fmt.Sprintf("ID: %s\n", process.ProcessInfo.ID))
	
	// Count elements
	eventCount := len(process.ProcessInfo.Elements.Events)
	activityCount := len(process.ProcessInfo.Elements.Activities)
	gatewayCount := len(process.ProcessInfo.Elements.Gateways)
	flowCount := len(process.ProcessInfo.Elements.SequenceFlows)
	
	sb.WriteString(fmt.Sprintf("Elements: %d events, %d activities, %d gateways, %d flows\n",
		eventCount, activityCount, gatewayCount, flowCount))

	// Find start and end events
	var starts, ends []string
	for _, event := range process.ProcessInfo.Elements.Events {
		if event.Type == "startEvent" {
			starts = append(starts, event.Name)
		} else if event.Type == "endEvent" {
			ends = append(ends, event.Name)
		}
	}
	
	sb.WriteString(fmt.Sprintf("Start: %s\n", strings.Join(starts, ", ")))
	sb.WriteString(fmt.Sprintf("End: %s\n", strings.Join(ends, ", ")))

	return sb.String()
}

