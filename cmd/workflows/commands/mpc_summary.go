package commands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/mpc"
)

type MPCSummaryCommand struct {
	*cli.BaseCommand
	verbose bool
}

func NewMPCSummaryCommand() *MPCSummaryCommand {
	cmd := &MPCSummaryCommand{
		BaseCommand: cli.NewBaseCommand("summary", "Display a summary of an MPC plan"),
	}

	cmd.FlagSet().BoolVar(&cmd.verbose, "verbose", false, "Show additional details")

	return cmd
}

func (c *MPCSummaryCommand) Execute(args []string) error {
	if err := c.ParseFlags(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(c.Help())
			return nil
		}
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("summary command requires file path")
	}

	inputFile := c.Arg(0)

	if err := cli.NewValidationChain().
		ValidateFilePath(inputFile, "file path").
		ValidateFileExtension(inputFile, []string{".yaml", ".yml", ".json"}, "file type").
		Error(); err != nil {
		return err
	}

	mpcPlan, err := mpc.LoadMPCFromFile(inputFile)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("failed to load MPC file: %v", err), err)
	}

	c.printSummary(mpcPlan)

	return nil
}

func (c *MPCSummaryCommand) printSummary(m *mpc.MPC) {
	// Basic Info
	fmt.Println("PLAN INFORMATION")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Name: %s\n", m.PlanName)
	fmt.Printf("ID: %s\n", m.PlanID)
	fmt.Printf("Version: %s\n", m.Version)
	fmt.Printf("Entry Node: %s\n", m.EntryNode)
	fmt.Printf("Total Nodes: %d\n", len(m.Nodes))

	// Global BPMN Status
	if m.GlobalBPMN != "" {
		fmt.Printf("Global BPMN Configured: %s\n", m.GlobalBPMN)
	} else {
		fmt.Printf("Global BPMN Configured: skipped\n")
	}
	fmt.Println()

	// Context
	fmt.Println("CONTEXT")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Business Goal:")
	c.printWrapped(m.Context.BusinessGoal, 2)
	fmt.Println()

	if len(m.Context.NonFunctionalRequirements) > 0 {
		fmt.Println("Non-Functional Requirements:")
		for _, req := range m.Context.NonFunctionalRequirements {
			fmt.Printf("  • %s\n", req)
		}
		fmt.Println()
	}

	// Architecture
	fmt.Println("ARCHITECTURE")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Overview:")
	c.printWrapped(m.Architecture.Overview, 2)
	fmt.Println()

	if len(m.Architecture.ADRs) > 0 {
		fmt.Println("ADRs:")
		for _, adr := range m.Architecture.ADRs {
			fmt.Printf("  • %s\n", adr)
		}
		fmt.Println()
	}

	if len(m.Architecture.Constraints) > 0 {
		fmt.Println("Constraints:")
		for _, constraint := range m.Architecture.Constraints {
			fmt.Printf("  • %s\n", constraint)
		}
		fmt.Println()
	}

	// Tooling
	fmt.Println("TOOLING")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Primary Language: %s\n", m.Tooling.PrimaryLanguage)

	if len(m.Tooling.SecondaryLanguages) > 0 {
		fmt.Printf("Secondary Languages: %s\n", strings.Join(m.Tooling.SecondaryLanguages, ", "))
	}

	if len(m.Tooling.Frameworks) > 0 {
		fmt.Println("Frameworks:")
		for _, framework := range m.Tooling.Frameworks {
			fmt.Printf("  • %s\n", framework)
		}
	}

	if m.Tooling.CodingStandards.Lint != "" || m.Tooling.CodingStandards.Formatting != "" || m.Tooling.CodingStandards.Testing != "" {
		fmt.Println("Coding Standards:")
		if m.Tooling.CodingStandards.Lint != "" {
			fmt.Printf("  Linting: %s\n", m.Tooling.CodingStandards.Lint)
		}
		if m.Tooling.CodingStandards.Formatting != "" {
			fmt.Printf("  Formatting: %s\n", m.Tooling.CodingStandards.Formatting)
		}
		if m.Tooling.CodingStandards.Testing != "" {
			fmt.Printf("  Testing: %s\n", m.Tooling.CodingStandards.Testing)
		}
	}
	fmt.Println()

	// Progress Summary
	statusCounts := make(map[string]int)
	totalSubtasks := 0
	completedSubtasks := 0

	for _, node := range m.Nodes {
		statusCounts[node.Status]++
		totalSubtasks += len(node.Subtasks)
		completedSubtasks += node.GetCompletedSubtaskCount()
	}

	fmt.Println("PROGRESS SUMMARY")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Node Status:")
	for status, count := range statusCounts {
		fmt.Printf("  %s: %d\n", status, count)
	}

	if totalSubtasks > 0 {
		fmt.Printf("\nSubtasks: %d/%d completed (%.1f%%)\n", completedSubtasks, totalSubtasks,
			float64(completedSubtasks)/float64(totalSubtasks)*100)
	}
	fmt.Println()

	// Nodes List
	fmt.Println("NODES")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-30s %-15s %-12s %s\n", "ID", "Status", "Material.", "Progress")
	fmt.Println(strings.Repeat("-", 80))

	for _, node := range m.Nodes {
		progress := ""
		if len(node.Subtasks) > 0 {
			progress = fmt.Sprintf("%d/%d (%.0f%%)",
				node.GetCompletedSubtaskCount(),
				len(node.Subtasks),
				node.GetCompletionPercentage())
		} else {
			progress = "N/A"
		}

		fmt.Printf("%-30s %-15s %11.2f  %s\n",
			truncateString(node.ID, 30),
			node.Status,
			node.Materialization,
			progress)

		if c.verbose && node.Description != "" {
			fmt.Printf("  └─ %s\n", node.Description)
		}
	}
}

func (c *MPCSummaryCommand) printWrapped(text string, indent int) {
	maxWidth := 78 - indent
	words := strings.Fields(text)
	line := ""
	indentStr := strings.Repeat(" ", indent)

	for _, word := range words {
		if len(line)+len(word)+1 > maxWidth {
			fmt.Printf("%s%s\n", indentStr, line)
			line = word
		} else {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		}
	}
	if line != "" {
		fmt.Printf("%s%s\n", indentStr, line)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (c *MPCSummaryCommand) Help() string {
	return `Display a summary of an MPC plan

Shows an overview of the plan including node statuses, progress, and materialization scores.

Usage:
  workflows mpc summary [options] <file>

Options:
  --verbose    Show additional details including descriptions

Arguments:
  file         Path to the MPC plan file (.yaml, .yml, or .json)

Examples:
  # Basic summary
  workflows mpc summary plan.yaml

  # Verbose summary with descriptions
  workflows mpc summary plan.yaml --verbose`
}
