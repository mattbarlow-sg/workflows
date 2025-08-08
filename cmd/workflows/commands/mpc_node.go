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
	showArtifacts bool
}

func NewMPCNodeCommand() *MPCNodeCommand {
	cmd := &MPCNodeCommand{
		BaseCommand: cli.NewBaseCommand("node", "Display details about a specific node in an MPC plan"),
	}
	
	cmd.FlagSet().BoolVar(&cmd.showArtifacts, "artifacts", false, "Show artifact details")
	
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
		completed := node.GetCompletedSubtaskCount()
		fmt.Printf("Subtasks: %d/%d completed (%.0f%%)\n", 
			completed, len(node.Subtasks), node.GetCompletionPercentage())
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
	
	if c.showArtifacts && node.Artifacts != nil {
		c.printArtifacts(node.Artifacts)
	}
}

func (c *MPCNodeCommand) printArtifacts(artifacts *mpc.Artifacts) {
	fmt.Println("Artifacts:")
	fmt.Println(strings.Repeat("-", 40))
	
	if artifacts.BPMN != "" {
		fmt.Printf("  BPMN: %s\n", artifacts.BPMN)
	}
	
	if artifacts.Spec != "" {
		fmt.Printf("  Spec: %s\n", artifacts.Spec)
	}
	
	if artifacts.Tests != "" {
		fmt.Printf("  Tests: %s\n", artifacts.Tests)
	}
	
	if artifacts.Properties != "" {
		fmt.Printf("  Properties: %s\n", artifacts.Properties)
	}
	
	if artifacts.PropertiesStruct != nil {
		fmt.Println("  Properties (Structured):")
		c.printArtifactProperties(artifacts.PropertiesStruct)
	}
	
	if artifacts.SchemasStruct != nil {
		fmt.Println("  Schemas (Structured):")
		c.printArtifactSchemas(artifacts.SchemasStruct)
	}
	
	if artifacts.SpecsStruct != nil {
		fmt.Println("  Specs (Structured):")
		c.printArtifactSpecs(artifacts.SpecsStruct)
	}
	
	if artifacts.TestsStruct != nil {
		fmt.Println("  Tests (Structured):")
		c.printArtifactTests(artifacts.TestsStruct)
	}
	
	fmt.Println()
}

func (c *MPCNodeCommand) printArtifactProperties(props *mpc.ArtifactProperties) {
	if props.Invariants != "" {
		fmt.Printf("    Invariants: %s\n", props.Invariants)
	}
	if props.States != "" {
		fmt.Printf("    States: %s\n", props.States)
	}
	if props.Rules != "" {
		fmt.Printf("    Rules: %s\n", props.Rules)
	}
	if props.TestSpecs != "" {
		fmt.Printf("    Test Specs: %s\n", props.TestSpecs)
	}
	if props.StateProperties != "" {
		fmt.Printf("    State Properties: %s\n", props.StateProperties)
	}
	if props.Generators != "" {
		fmt.Printf("    Generators: %s\n", props.Generators)
	}
}

func (c *MPCNodeCommand) printArtifactSchemas(schemas *mpc.ArtifactSchemas) {
	if schemas.Validation != "" {
		fmt.Printf("    Validation: %s\n", schemas.Validation)
	}
	if schemas.Transformations != "" {
		fmt.Printf("    Transformations: %s\n", schemas.Transformations)
	}
	if schemas.Contracts != "" {
		fmt.Printf("    Contracts: %s\n", schemas.Contracts)
	}
}

func (c *MPCNodeCommand) printArtifactSpecs(specs *mpc.ArtifactSpecs) {
	if specs.API != "" {
		fmt.Printf("    API: %s\n", specs.API)
	}
	if specs.Models != "" {
		fmt.Printf("    Models: %s\n", specs.Models)
	}
	if specs.Schemas != "" {
		fmt.Printf("    Schemas: %s\n", specs.Schemas)
	}
}

func (c *MPCNodeCommand) printArtifactTests(tests *mpc.ArtifactTests) {
	if tests.Property != "" {
		fmt.Printf("    Property: %s\n", tests.Property)
	}
	if tests.Deterministic != "" {
		fmt.Printf("    Deterministic: %s\n", tests.Deterministic)
	}
	if tests.Fuzz != "" {
		fmt.Printf("    Fuzz: %s\n", tests.Fuzz)
	}
	if tests.Contract != "" {
		fmt.Printf("    Contract: %s\n", tests.Contract)
	}
	if tests.Unit != "" {
		fmt.Printf("    Unit: %s\n", tests.Unit)
	}
	if tests.Integration != "" {
		fmt.Printf("    Integration: %s\n", tests.Integration)
	}
	if tests.E2E != "" {
		fmt.Printf("    E2E: %s\n", tests.E2E)
	}
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
acceptance criteria, downstream dependencies, and optionally artifacts.

Usage:
  workflows mpc node [options] <file> <node-id>

Options:
  --artifacts    Show artifact details if present

Arguments:
  file         Path to the MPC plan file (.yaml, .yml, or .json)
  node-id      ID of the node to display

Examples:
  # Basic node details
  workflows mpc node plan.yaml validation-framework

  # Node details with artifacts
  workflows mpc node plan.yaml validation-framework --artifacts`
}