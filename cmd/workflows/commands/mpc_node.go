package commands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/mpc"
)

type MPCNodeCommand struct {
	*cli.BaseCommand
}

func NewMPCNodeCommand() *MPCNodeCommand {
	cmd := &MPCNodeCommand{
		BaseCommand: cli.NewBaseCommand("node", "Display details about a specific node in an MPC plan"),
	}

	return cmd
}

func (c *MPCNodeCommand) Execute(args []string) error {
	if err := c.ParseFlags(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(c.Help())
			return nil
		}
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	if c.NArg() < 2 {
		c.Usage()
		return errors.NewUsageError("node command requires file path and node ID")
	}

	inputFile := c.Arg(0)
	nodeID := c.Arg(1)

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

	node := mpcPlan.GetNodeByID(nodeID)
	if node == nil {
		return errors.NewUsageError(fmt.Sprintf("node '%s' not found in plan", nodeID))
	}

	c.printNodeDetails(node)

	return nil
}

func (c *MPCNodeCommand) printNodeDetails(node *mpc.Node) {
	fmt.Printf("Node: %s\n", node.ID)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	fmt.Printf("Status: %s\n", node.Status)
	fmt.Printf("Materialization: %.2f\n", node.Materialization)
	fmt.Println()

	if node.Description != "" {
		fmt.Println("Description:")
		fmt.Printf("  %s\n", node.Description)
		fmt.Println()
	}

	if node.DetailedDescription != "" {
		fmt.Println("Detailed Description:")
		c.printWrapped(node.DetailedDescription, 2)
		fmt.Println()
	}

	if len(node.Subtasks) > 0 {
		completed := 0
		for _, subtask := range node.Subtasks {
			if subtask.Completed {
				completed++
			}
		}
		percentage := float64(completed) / float64(len(node.Subtasks)) * 100
		fmt.Printf("Subtasks: %d/%d completed (%.0f%%)\n",
			completed, len(node.Subtasks), percentage)
		for i, subtask := range node.Subtasks {
			status := "[ ]"
			if subtask.Completed {
				status = "[✓]"
			}
			fmt.Printf("  %s %d. %s\n", status, i+1, subtask.Description)
		}
		fmt.Println()
	}

	if len(node.AcceptanceCriteria) > 0 {
		fmt.Println("Acceptance Criteria:")
		for _, criteria := range node.AcceptanceCriteria {
			fmt.Printf("  • %s\n", criteria)
		}
		fmt.Println()
	}

	if node.DefinitionOfDone != "" {
		fmt.Println("Definition of Done:")
		fmt.Printf("  %s\n", node.DefinitionOfDone)
		fmt.Println()
	}

	if len(node.Outputs) > 0 {
		fmt.Println("Outputs:")
		for _, output := range node.Outputs {
			fmt.Printf("  - %s\n", output)
		}
		fmt.Println()
	}

	if len(node.RequiredKnowledge) > 0 {
		fmt.Println("Required Knowledge:")
		for _, knowledge := range node.RequiredKnowledge {
			fmt.Printf("  - %s\n", knowledge)
		}
		fmt.Println()
	}

	if len(node.Downstream) > 0 {
		fmt.Println("Downstream Nodes:")
		for _, downstream := range node.Downstream {
			fmt.Printf("  → %s\n", downstream)
		}
		fmt.Println()
	}

	if node.Artifacts != nil {
		c.printArtifacts(node.Artifacts)
	}
}

func (c *MPCNodeCommand) printArtifacts(artifacts *mpc.Artifacts) {
	fmt.Println("Artifacts:")
	fmt.Println(strings.Repeat("-", 40))

	if artifacts.BPMN != nil && *artifacts.BPMN != "" {
		fmt.Printf("  BPMN: %s\n", *artifacts.BPMN)
	}

	if artifacts.FormalSpec != nil && *artifacts.FormalSpec != "" {
		fmt.Printf("  Formal Spec: %s\n", *artifacts.FormalSpec)
	}

	if artifacts.Schemas != nil && *artifacts.Schemas != "" {
		fmt.Printf("  Schemas: %s\n", *artifacts.Schemas)
	}

	if artifacts.ModelChecking != nil && *artifacts.ModelChecking != "" {
		fmt.Printf("  Model Checking: %s\n", *artifacts.ModelChecking)
	}

	if artifacts.TestGenerators != nil && *artifacts.TestGenerators != "" {
		fmt.Printf("  Test Generators: %s\n", *artifacts.TestGenerators)
	}

	fmt.Println()
}


func (c *MPCNodeCommand) printWrapped(text string, indent int) {
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

func (c *MPCNodeCommand) Help() string {
	return `Display details about a specific node in an MPC plan

Shows comprehensive information about a node including its status, subtasks,
acceptance criteria, downstream dependencies, and artifacts.

Usage:
  workflows mpc node <file> <node-id>

Arguments:
  file         Path to the MPC plan file (.yaml, .yml, or .json)
  node-id      ID of the node to display

Examples:
  # Display node details
  workflows mpc node plan.yaml validation-framework

  # Display another node
  workflows mpc node plan.yaml testing-framework`
}
