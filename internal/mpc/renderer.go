package mpc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Renderer struct {
	mpc *MPC
}

func NewRenderer(mpc *MPC) *Renderer {
	return &Renderer{
		mpc: mpc,
	}
}

func LoadMPCFromFile(filePath string) (*MPC, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try YAML first
	var mpc MPC
	if err := yaml.Unmarshal(data, &mpc); err == nil {
		return &mpc, nil
	}

	// Try JSON
	if err := json.Unmarshal(data, &mpc); err == nil {
		return &mpc, nil
	}

	return nil, fmt.Errorf("failed to parse file as YAML or JSON")
}

func (r *Renderer) Render(format string) (string, error) {
	switch format {
	case "yaml":
		return r.renderYAML()
	case "json":
		return r.renderJSON()
	case "text":
		return r.renderText()
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *Renderer) renderYAML() (string, error) {
	data, err := yaml.Marshal(r.mpc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	return string(data), nil
}

func (r *Renderer) renderJSON() (string, error) {
	data, err := json.MarshalIndent(r.mpc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

func (r *Renderer) renderText() (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("MPC Workflow: %s\n", r.mpc.PlanName))
	sb.WriteString(fmt.Sprintf("Plan ID: %s\n", r.mpc.PlanID))
	sb.WriteString(fmt.Sprintf("Version: %s\n", r.mpc.Version))
	sb.WriteString(strings.Repeat("=", 80) + "\n\n")

	// Context
	sb.WriteString("CONTEXT\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Business Goal: %s\n", r.mpc.Context.BusinessGoal))
	sb.WriteString("\nNon-Functional Requirements:\n")
	for _, req := range r.mpc.Context.NonFunctionalRequirements {
		sb.WriteString(fmt.Sprintf("  • %s\n", req))
	}

	// Architecture
	sb.WriteString("\n\nARCHITECTURE\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Overview: %s\n", r.mpc.Architecture.Overview))

	if len(r.mpc.Architecture.ADRs) > 0 {
		sb.WriteString("\nADRs:\n")
		for _, adr := range r.mpc.Architecture.ADRs {
			sb.WriteString(fmt.Sprintf("  • %s\n", adr))
		}
	}

	if len(r.mpc.Architecture.Constraints) > 0 {
		sb.WriteString("\nConstraints:\n")
		for _, constraint := range r.mpc.Architecture.Constraints {
			sb.WriteString(fmt.Sprintf("  • %s\n", constraint))
		}
	}

	// Tooling
	sb.WriteString("\n\nTOOLING\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Primary Language: %s\n", r.mpc.Tooling.PrimaryLanguage))

	if len(r.mpc.Tooling.SecondaryLanguages) > 0 {
		sb.WriteString(fmt.Sprintf("Secondary Languages: %s\n", strings.Join(r.mpc.Tooling.SecondaryLanguages, ", ")))
	}

	if len(r.mpc.Tooling.Frameworks) > 0 {
		sb.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(r.mpc.Tooling.Frameworks, ", ")))
	}

	sb.WriteString("\nCoding Standards:\n")
	sb.WriteString(fmt.Sprintf("  • Lint: %s\n", r.mpc.Tooling.CodingStandards.Lint))
	sb.WriteString(fmt.Sprintf("  • Formatting: %s\n", r.mpc.Tooling.CodingStandards.Formatting))
	sb.WriteString(fmt.Sprintf("  • Testing: %s\n", r.mpc.Tooling.CodingStandards.Testing))

	// Workflow
	sb.WriteString("\n\nWORKFLOW\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Entry Node: %s\n", r.mpc.EntryNode))

	// Nodes
	sb.WriteString("\n\nNODES\n")
	sb.WriteString(strings.Repeat("=", 80) + "\n")

	for i, node := range r.mpc.Nodes {
		if i > 0 {
			sb.WriteString("\n" + strings.Repeat("-", 80) + "\n")
		}

		// Node header
		sb.WriteString(fmt.Sprintf("\nNode: %s\n", node.ID))
		sb.WriteString(fmt.Sprintf("Status: %s | Materialization: %.1f\n", node.Status, node.Materialization))
		sb.WriteString(fmt.Sprintf("Description: %s\n", node.Description))

		// Detailed description
		if node.DetailedDescription != "" {
			sb.WriteString("\nDetailed Description:\n")
			lines := strings.Split(strings.TrimSpace(node.DetailedDescription), "\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}

		// Subtasks
		completedCount := 0
		for _, subtask := range node.Subtasks {
			if subtask.Completed {
				completedCount++
			}
		}
		sb.WriteString(fmt.Sprintf("\nSubtasks (%d/%d completed):\n", completedCount, len(node.Subtasks)))
		for j, subtask := range node.Subtasks {
			status := "[ ]"
			if subtask.Completed {
				status = "[✓]"
			}
			sb.WriteString(fmt.Sprintf("  %d. %s %s\n", j+1, status, subtask.Description))
		}

		// Outputs
		if len(node.Outputs) > 0 {
			sb.WriteString("\nOutputs:\n")
			for _, output := range node.Outputs {
				sb.WriteString(fmt.Sprintf("  • %s\n", output))
			}
		}

		// Acceptance Criteria
		sb.WriteString("\nAcceptance Criteria:\n")
		for _, criteria := range node.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("  • %s\n", criteria))
		}

		// Definition of Done
		sb.WriteString(fmt.Sprintf("\nDefinition of Done: %s\n", node.DefinitionOfDone))

		// Required Knowledge
		if len(node.RequiredKnowledge) > 0 {
			sb.WriteString("\nRequired Knowledge:\n")
			for _, knowledge := range node.RequiredKnowledge {
				sb.WriteString(fmt.Sprintf("  • %s\n", knowledge))
			}
		}

		// Artifacts
		if node.Artifacts != nil {
			sb.WriteString("\nArtifacts:\n")
			if node.Artifacts.BPMN != nil && *node.Artifacts.BPMN != "" {
				sb.WriteString(fmt.Sprintf("  • BPMN: %s\n", *node.Artifacts.BPMN))
			}
			if node.Artifacts.FormalSpec != nil && *node.Artifacts.FormalSpec != "" {
				sb.WriteString(fmt.Sprintf("  • Formal Spec: %s\n", *node.Artifacts.FormalSpec))
			}
			if node.Artifacts.Schemas != nil && *node.Artifacts.Schemas != "" {
				sb.WriteString(fmt.Sprintf("  • Schemas: %s\n", *node.Artifacts.Schemas))
			}
			if node.Artifacts.ModelChecking != nil && *node.Artifacts.ModelChecking != "" {
				sb.WriteString(fmt.Sprintf("  • Model Checking: %s\n", *node.Artifacts.ModelChecking))
			}
			if node.Artifacts.TestGenerators != nil && *node.Artifacts.TestGenerators != "" {
				sb.WriteString(fmt.Sprintf("  • Test Generators: %s\n", *node.Artifacts.TestGenerators))
			}
		}

		// Downstream
		if len(node.Downstream) > 0 {
			sb.WriteString(fmt.Sprintf("\nDownstream: %s\n", strings.Join(node.Downstream, ", ")))
		} else {
			sb.WriteString("\nDownstream: (none)\n")
		}
	}

	// Summary statistics
	sb.WriteString("\n\n" + strings.Repeat("=", 80) + "\n")
	sb.WriteString("SUMMARY\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	// Count nodes by status
	statusCounts := make(map[string]int)
	totalSubtasks := 0
	completedSubtasks := 0

	for _, node := range r.mpc.Nodes {
		statusCounts[node.Status]++
		totalSubtasks += len(node.Subtasks)
		for _, subtask := range node.Subtasks {
			if subtask.Completed {
				completedSubtasks++
			}
		}
	}

	sb.WriteString(fmt.Sprintf("Total Nodes: %d\n", len(r.mpc.Nodes)))
	for status, count := range statusCounts {
		sb.WriteString(fmt.Sprintf("  • %s: %d\n", status, count))
	}

	sb.WriteString(fmt.Sprintf("\nTotal Subtasks: %d\n", totalSubtasks))
	sb.WriteString(fmt.Sprintf("Completed Subtasks: %d\n", completedSubtasks))
	if totalSubtasks > 0 {
		completionPercentage := float64(completedSubtasks) / float64(totalSubtasks) * 100
		sb.WriteString(fmt.Sprintf("Overall Completion: %.1f%%\n", completionPercentage))
	}

	return sb.String(), nil
}

func (r *Renderer) WriteToFile(filePath, format string) error {
	content, err := r.Render(format)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}
