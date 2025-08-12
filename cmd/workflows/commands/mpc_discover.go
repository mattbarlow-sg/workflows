package commands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/mpc"
)

type MPCDiscoverCommand struct {
	*cli.BaseCommand
	showStatus   bool
	showProgress bool
	nextOnly     bool
}

func NewMPCDiscoverCommand() *MPCDiscoverCommand {
	cmd := &MPCDiscoverCommand{
		BaseCommand: cli.NewBaseCommand("discover", "Discover what tasks can be worked on next"),
	}

	// Define flags
	cmd.FlagSet().BoolVar(&cmd.showStatus, "status", false, "Show node status")
	cmd.FlagSet().BoolVar(&cmd.showProgress, "progress", false, "Show progress indicators")
	cmd.FlagSet().BoolVar(&cmd.nextOnly, "next-only", false, "Show only what's immediately actionable")

	return cmd
}

func (c *MPCDiscoverCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		// Check if it's a help request
		if err == flag.ErrHelp {
			fmt.Println(c.Help())
			return nil
		}
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("discover command requires file path")
	}

	inputFile := c.Arg(0)

	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(inputFile, "file path").
		ValidateFileExtension(inputFile, []string{".yaml", ".yml", ".json"}, "file type").
		Error(); err != nil {
		return err
	}

	// Load MPC from file
	mpcData, err := mpc.LoadMPCFromFile(inputFile)
	if err != nil {
		return errors.NewIOError(fmt.Sprintf("failed to load MPC file: %v", err), err)
	}

	// Display the discovery analysis
	if c.nextOnly {
		c.analyzeWorkflowNextOnly(mpcData)
	} else {
		fmt.Printf("MPC Workflow: %s\n", mpcData.PlanName)
		fmt.Printf("Plan ID: %s\n", mpcData.PlanID)
		fmt.Println()

		// Analyze and display workflow state
		c.analyzeWorkflow(mpcData)
	}

	return nil
}

func (c *MPCDiscoverCommand) analyzeWorkflowNextOnly(mpcData *mpc.MPC) {
	// Categorize nodes by their status and dependencies
	workableNow := []*mpc.Node{}
	inProgressNodes := []*mpc.Node{}
	needsBPMN := []*mpc.Node{}
	needsSpecs := []*mpc.Node{}

	for i := range mpcData.Nodes {
		node := &mpcData.Nodes[i]

		switch node.Status {
		case "In Progress":
			inProgressNodes = append(inProgressNodes, node)
		case "Ready":
			// For the entry node, it's always workable if Ready
			if node.ID == mpcData.EntryNode {
				workableNow = append(workableNow, node)
				// Check if node needs artifacts
				c.checkNodeArtifacts(node, &needsBPMN, &needsSpecs)
			} else {
				// Check if all dependencies are completed
				canWork := true

				// Find nodes that have this node as downstream
				for j := range mpcData.Nodes {
					upstream := &mpcData.Nodes[j]
					for _, downstream := range upstream.Downstream {
						if downstream == node.ID && upstream.Status != "Completed" {
							canWork = false
							break
						}
					}
					if !canWork {
						break
					}
				}

				if canWork {
					workableNow = append(workableNow, node)
					// Check if node needs artifacts
					c.checkNodeArtifacts(node, &needsBPMN, &needsSpecs)
				}
			}
		}
	}

	// Show what needs to be done right now
	hasWork := false

	fmt.Printf("MPC Workflow: %s\n", mpcData.PlanName)
	fmt.Printf("Plan ID: %s\n", mpcData.PlanID)
	fmt.Println()

	// Show artifact generation first
	if len(needsBPMN) > 0 {
		hasWork = true
		fmt.Println("ARTIFACT GENERATION NEEDED:")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\n  Nodes needing BPMN design:")
		for _, node := range needsBPMN {
			fmt.Printf("    * %s - %s\n", node.ID, node.Description)
			fmt.Printf("      Run: ./workflows ai bpmn-create --node %s\n", node.ID)
		}
		fmt.Println()
	}

	if len(needsSpecs) > 0 {
		hasWork = true
		if len(needsBPMN) == 0 {
			fmt.Println("ARTIFACT GENERATION NEEDED:")
			fmt.Println(strings.Repeat("=", 60))
		}
		fmt.Println("  Nodes needing Specs/Tests/Properties:")
		for _, node := range needsSpecs {
			missing := c.getMissingArtifacts(node)
			fmt.Printf("    * %s - %s\n", node.ID, node.Description)
			fmt.Printf("      Missing: %s\n", strings.Join(missing, ", "))
			fmt.Printf("      Run: ./workflows ai spec-generate --node %s\n", node.ID)
		}
		fmt.Println()
	}

	// Show ready work with full details
	if len(workableNow) > 0 {
		hasWork = true
		fmt.Println("READY TO WORK ON NOW:")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range workableNow {
			c.printFullNodeDetails(node)
		}
	}

	// Show in-progress work with full details
	if len(inProgressNodes) > 0 {
		hasWork = true
		fmt.Println("⏳ IN PROGRESS:")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range inProgressNodes {
			c.printFullNodeDetails(node)
		}
	}

	completedCount := 0
	for _, node := range mpcData.Nodes {
		if node.Status == "Completed" {
			completedCount++
		}
	}
	fmt.Printf("  Completed: %d\n", completedCount)
	fmt.Printf("  Total: %d\n", len(mpcData.Nodes))

	if len(mpcData.Nodes) > 0 {
		completionRate := float64(completedCount) / float64(len(mpcData.Nodes)) * 100
		fmt.Printf("  Overall completion: %.1f%%\n", completionRate)
	}

	if !hasWork {
		if completedCount == len(mpcData.Nodes) {
			fmt.Println("\nAll tasks completed!")
		} else {
			fmt.Println("\nNo tasks are currently ready to work on.")
			fmt.Println("    Check blocked dependencies with: ./workflows mpc discover <file>")
		}
	}
}

func (c *MPCDiscoverCommand) analyzeWorkflow(mpcData *mpc.MPC) {
	// Categorize nodes by their status and dependencies
	workableNow := []*mpc.Node{}
	blockedNodes := []*mpc.Node{}
	inProgressNodes := []*mpc.Node{}
	completedNodes := []*mpc.Node{}

	// New categories for artifact tracking
	needsBPMN := []*mpc.Node{}
	needsSpecs := []*mpc.Node{}

	// Track which nodes have incomplete dependencies
	nodeBlockers := make(map[string][]string)

	for i := range mpcData.Nodes {
		node := &mpcData.Nodes[i]

		switch node.Status {
		case "Completed":
			completedNodes = append(completedNodes, node)
		case "In Progress":
			inProgressNodes = append(inProgressNodes, node)
		case "Blocked":
			blockedNodes = append(blockedNodes, node)
		case "Ready":
			// For the entry node, it's always workable if Ready
			if node.ID == mpcData.EntryNode {
				workableNow = append(workableNow, node)
				// Check if node needs artifacts
				c.checkNodeArtifacts(node, &needsBPMN, &needsSpecs)
			} else {
				// Check if all dependencies are completed
				canWork := true
				blockers := []string{}

				// Find nodes that have this node as downstream
				for j := range mpcData.Nodes {
					upstream := &mpcData.Nodes[j]
					for _, downstream := range upstream.Downstream {
						if downstream == node.ID && upstream.Status != "Completed" {
							canWork = false
							blockers = append(blockers, upstream.ID)
						}
					}
				}

				if canWork {
					workableNow = append(workableNow, node)
					// Check if node needs artifacts
					c.checkNodeArtifacts(node, &needsBPMN, &needsSpecs)
				} else {
					blockedNodes = append(blockedNodes, node)
					nodeBlockers[node.ID] = blockers
				}
			}
		}
	}

	// Display artifact needs first
	if len(needsBPMN) > 0 || len(needsSpecs) > 0 {
		fmt.Println("ARTIFACT GENERATION NEEDED:")
		fmt.Println(strings.Repeat("=", 60))

		if len(needsBPMN) > 0 {
			fmt.Println("\n  Nodes needing BPMN design:")
			for _, node := range needsBPMN {
				fmt.Printf("    * %s - %s\n", node.ID, node.Description)
				fmt.Printf("      Run: ./workflows ai bpmn-create --node %s\n", node.ID)
			}
		}

		if len(needsSpecs) > 0 {
			fmt.Println("\n  Nodes needing Specs/Tests/Properties:")
			for _, node := range needsSpecs {
				fmt.Printf("    * %s - %s\n", node.ID, node.Description)
				missing := c.getMissingArtifacts(node)
				fmt.Printf("      Missing: %s\n", strings.Join(missing, ", "))
				fmt.Printf("      Run: ./workflows ai spec-generate --node %s\n", node.ID)
			}
		}
		fmt.Println()
	}

	// Display workable nodes
	if len(workableNow) == 1 {
		fmt.Println("READY TO WORK ON NOW:")
	} else if len(workableNow) > 1 {
		fmt.Println("READY TO WORK ON NOW (can be done in parallel):")
	} else {
		fmt.Println("READY TO WORK ON NOW:")
	}
	fmt.Println(strings.Repeat("=", 60))
	if len(workableNow) == 0 {
		fmt.Println("  No nodes are ready to work on.")
	} else {
		for _, node := range workableNow {
			c.printNodeSummary(node)
		}
	}

	// Display in-progress nodes
	if len(inProgressNodes) > 0 {
		fmt.Println("\nIN PROGRESS:")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range inProgressNodes {
			c.printNodeSummary(node)
		}
	}

	// Display blocked nodes with their blockers
	if len(blockedNodes) > 0 {
		fmt.Println("\nBLOCKED (waiting on dependencies):")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range blockedNodes {
			c.printNodeSummary(node)
			if blockers, ok := nodeBlockers[node.ID]; ok && len(blockers) > 0 {
				fmt.Printf("     Waiting on: %s\n", strings.Join(blockers, ", "))
			}
		}
	}

	// Show workflow execution stages
	fmt.Println("\nWORKFLOW EXECUTION STAGES:")
	fmt.Println(strings.Repeat("=", 60))
	c.showExecutionStages(mpcData)

	// Summary statistics
	fmt.Println("\nSUMMARY:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  Ready to work: %d\n", len(workableNow))
	fmt.Printf("  In progress: %d\n", len(inProgressNodes))
	fmt.Printf("  Blocked: %d\n", len(blockedNodes))
	fmt.Printf("  Completed: %d\n", len(completedNodes))
	fmt.Printf("  Total: %d\n", len(mpcData.Nodes))

	if len(mpcData.Nodes) > 0 {
		completionRate := float64(len(completedNodes)) / float64(len(mpcData.Nodes)) * 100
		fmt.Printf("  Overall completion: %.1f%%\n", completionRate)
	}
}

func (c *MPCDiscoverCommand) printNodeSummary(node *mpc.Node) {
	fmt.Printf("  %s %s\n", c.getStatusIcon(node.Status), node.ID)
	fmt.Printf("     Description: %s\n", node.Description)

	if c.showProgress {
		completedCount := 0
		for _, subtask := range node.Subtasks {
			if subtask.Completed {
				completedCount++
			}
		}
		progress := float64(0)
		if len(node.Subtasks) > 0 {
			progress = float64(completedCount) / float64(len(node.Subtasks)) * 100
		}
		fmt.Printf("     Progress: %.0f%% complete", progress)
		fmt.Printf(" | Materialization: %.1f\n", node.Materialization)
		totalCount := len(node.Subtasks)
		fmt.Printf("     Subtasks: %d/%d completed\n", completedCount, totalCount)
	}

	if len(node.Downstream) > 0 {
		fmt.Printf("     Unlocks: %s\n", strings.Join(node.Downstream, ", "))
	}
}

func (c *MPCDiscoverCommand) printFullNodeDetails(node *mpc.Node) {
	fmt.Printf("  %s %s\n", c.getStatusIcon(node.Status), node.ID)
	fmt.Printf("     Status: %s\n", node.Status)
	fmt.Printf("     Materialization: %.1f\n", node.Materialization)
	fmt.Printf("     Description: %s\n", node.Description)

	if node.DetailedDescription != "" {
		fmt.Printf("     Detailed Description:\n")
		lines := strings.Split(strings.TrimSpace(node.DetailedDescription), "\n")
		for _, line := range lines {
			fmt.Printf("       %s\n", line)
		}
	}

	if len(node.Subtasks) > 0 {
		fmt.Printf("     Subtasks:\n")
		for _, subtask := range node.Subtasks {
			if subtask.Completed {
				fmt.Printf("       [x] %s\n", subtask.Description)
			} else {
				fmt.Printf("       - %s\n", subtask.Description)
			}
		}
	}

	if len(node.Outputs) > 0 {
		fmt.Printf("     Outputs: %s\n", strings.Join(node.Outputs, ", "))
	}

	if len(node.AcceptanceCriteria) > 0 {
		fmt.Printf("     Acceptance Criteria:\n")
		for _, criteria := range node.AcceptanceCriteria {
			fmt.Printf("       - %s\n", criteria)
		}
	}

	if node.DefinitionOfDone != "" {
		fmt.Printf("     Definition of Done: %s\n", node.DefinitionOfDone)
	}

	if len(node.RequiredKnowledge) > 0 {
		fmt.Printf("     Required Knowledge: %s\n", strings.Join(node.RequiredKnowledge, ", "))
	}

	if node.Artifacts != nil {
		fmt.Printf("     Artifacts:\n")
		if node.Artifacts.BPMN != nil && *node.Artifacts.BPMN != "" {
			fmt.Printf("       BPMN: %s\n", *node.Artifacts.BPMN)
		}
		if node.Artifacts.FormalSpec != nil && *node.Artifacts.FormalSpec != "" {
			fmt.Printf("       Formal Spec: %s\n", *node.Artifacts.FormalSpec)
		}
		if node.Artifacts.Schemas != nil && *node.Artifacts.Schemas != "" {
			fmt.Printf("       Schemas: %s\n", *node.Artifacts.Schemas)
		}
		if node.Artifacts.ModelChecking != nil && *node.Artifacts.ModelChecking != "" {
			fmt.Printf("       Model Checking: %s\n", *node.Artifacts.ModelChecking)
		}
		if node.Artifacts.TestGenerators != nil && *node.Artifacts.TestGenerators != "" {
			fmt.Printf("       Test Generators: %s\n", *node.Artifacts.TestGenerators)
		}
	}

	if len(node.Downstream) > 0 {
		fmt.Printf("     Downstream: %s\n", strings.Join(node.Downstream, ", "))
	}

	fmt.Println()
}

func (c *MPCDiscoverCommand) showExecutionStages(mpcData *mpc.MPC) {
	// Build execution stages based on dependencies
	stages := c.buildExecutionStages(mpcData)

	if len(stages) == 0 {
		fmt.Println("  No execution stages found.")
		return
	}

	for i, stage := range stages {
		fmt.Printf("\n  Stage %d:", i+1)
		if len(stage) > 1 {
			fmt.Printf(" (parallel execution)\n")
		} else {
			fmt.Printf("\n")
		}

		for _, nodeID := range stage {
			node := mpcData.GetNodeByID(nodeID)
			if node != nil {
				fmt.Printf("    %s %s - %s\n", c.getStatusIcon(node.Status), nodeID, node.Description)
			}
		}

		// Show what this stage unlocks
		unlocks := make(map[string]bool)
		for _, nodeID := range stage {
			node := mpcData.GetNodeByID(nodeID)
			if node != nil {
				for _, downstream := range node.Downstream {
					unlocks[downstream] = true
				}
			}
		}
		if len(unlocks) > 0 {
			unlockList := []string{}
			for id := range unlocks {
				unlockList = append(unlockList, id)
			}
			fmt.Printf("    Unlocks: %s\n", strings.Join(unlockList, ", "))
		}
	}
}

func (c *MPCDiscoverCommand) buildExecutionStages(mpcData *mpc.MPC) [][]string {
	// Track which nodes have been assigned to stages
	assigned := make(map[string]bool)
	stages := [][]string{}

	// Track dependencies
	dependencies := make(map[string][]string)
	for _, node := range mpcData.Nodes {
		for _, downstream := range node.Downstream {
			dependencies[downstream] = append(dependencies[downstream], node.ID)
		}
	}

	// Build stages
	for len(assigned) < len(mpcData.Nodes) {
		currentStage := []string{}

		for i := range mpcData.Nodes {
			node := &mpcData.Nodes[i]

			// Skip if already assigned
			if assigned[node.ID] {
				continue
			}

			// Check if all dependencies are satisfied
			canAdd := true
			if deps, hasDeps := dependencies[node.ID]; hasDeps {
				for _, dep := range deps {
					if !assigned[dep] {
						canAdd = false
						break
					}
				}
			} else if node.ID != mpcData.EntryNode {
				// If no dependencies and not entry node, check if it's reachable
				canAdd = false
			}

			if canAdd {
				currentStage = append(currentStage, node.ID)
			}
		}

		// Add current stage
		if len(currentStage) > 0 {
			for _, nodeID := range currentStage {
				assigned[nodeID] = true
			}
			stages = append(stages, currentStage)
		} else {
			// Prevent infinite loop if there are cycles or unreachable nodes
			break
		}
	}

	return stages
}

func (c *MPCDiscoverCommand) getStatusIcon(status string) string {
	switch status {
	case "Ready":
		return "○"
	case "In Progress":
		return "◐"
	case "Blocked":
		return "■"
	case "Completed":
		return "●"
	default:
		return "?"
	}
}

func (c *MPCDiscoverCommand) checkNodeArtifacts(node *mpc.Node, needsBPMN, needsSpecs *[]*mpc.Node) {
	if node.Artifacts == nil {
		*needsBPMN = append(*needsBPMN, node)
		return
	}

	// Check for BPMN first
	if node.Artifacts.BPMN == nil || *node.Artifacts.BPMN == "" {
		*needsBPMN = append(*needsBPMN, node)
		return
	}

	// Check for formal spec and schemas
	hasSpec := node.Artifacts.FormalSpec != nil && *node.Artifacts.FormalSpec != ""
	hasSchemas := node.Artifacts.Schemas != nil && *node.Artifacts.Schemas != ""
	hasTestGenerators := node.Artifacts.TestGenerators != nil && *node.Artifacts.TestGenerators != ""

	if !hasSpec || !hasSchemas || !hasTestGenerators {
		*needsSpecs = append(*needsSpecs, node)
	}
}

func (c *MPCDiscoverCommand) getMissingArtifacts(node *mpc.Node) []string {
	missing := []string{}

	if node.Artifacts == nil {
		return []string{"formal_spec", "schemas", "test_generators"}
	}

	hasSpec := node.Artifacts.FormalSpec != nil && *node.Artifacts.FormalSpec != ""
	hasSchemas := node.Artifacts.Schemas != nil && *node.Artifacts.Schemas != ""
	hasTestGenerators := node.Artifacts.TestGenerators != nil && *node.Artifacts.TestGenerators != ""

	if !hasSpec {
		missing = append(missing, "formal_spec")
	}
	if !hasSchemas {
		missing = append(missing, "schemas")
	}
	if !hasTestGenerators {
		missing = append(missing, "test_generators")
	}

	return missing
}

func (c *MPCDiscoverCommand) Help() string {
	return `Discover what tasks can be worked on next

This command analyzes the MPC workflow to show:
- Which tasks need artifact generation (BPMN, specs, tests)
- Which tasks are ready to work on now (can be done in parallel)
- Which tasks are in progress
- Which tasks are blocked and what they're waiting for
- Sequential and parallel workflow paths

Usage:
  workflows mpc discover [options] <file>

Options:
  --next-only  Show only what's immediately actionable (concise output)
  --progress   Show detailed progress information (verbose mode only)
  --status     Show node status (deprecated, always shown)

Arguments:
  file         Path to the MPC workflow file (.yaml, .yml, or .json)

Output Modes:
  Default (verbose): Shows all workflow information including blocked tasks,
                     execution stages, and summary statistics
  
  --next-only:       Shows only immediate next actions:
                     - BPMN designs to generate
                     - Specs/tests to create
                     - In-progress tasks to continue
                     - Ready tasks to implement

Display Sections (verbose mode):
  ARTIFACT GENERATION: Tasks needing BPMN or spec generation
  READY TO WORK: Tasks with all dependencies completed
  ⏳ IN PROGRESS: Tasks currently being worked on
  BLOCKED: Tasks waiting on dependencies
  EXECUTION STAGES: Ordered stages showing parallel/sequential flow
  SUMMARY: Overall workflow statistics

Status Icons:
  ○ Ready, ◐ In Progress, ■ Blocked, ● Completed

Examples:
  # Show only next actions (concise)
  workflows mpc discover workflow.yaml --next-only

  # Full discovery view (verbose)
  workflows mpc discover workflow.yaml

  # Show with detailed progress
  workflows mpc discover workflow.yaml --progress`
}
