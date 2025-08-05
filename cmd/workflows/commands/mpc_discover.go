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
	showStatus bool
	showProgress bool
}

func NewMPCDiscoverCommand() *MPCDiscoverCommand {
	cmd := &MPCDiscoverCommand{
		BaseCommand: cli.NewBaseCommand("discover", "Discover what tasks can be worked on next"),
	}
	
	// Define flags
	cmd.FlagSet().BoolVar(&cmd.showStatus, "status", false, "Show node status")
	cmd.FlagSet().BoolVar(&cmd.showProgress, "progress", false, "Show progress indicators")
	
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
	fmt.Printf("MPC Workflow: %s\n", mpcData.ProjectName)
	fmt.Printf("Plan ID: %s\n", mpcData.PlanID)
	fmt.Println()
	
	// Analyze and display workflow state
	c.analyzeWorkflow(mpcData)

	return nil
}

func (c *MPCDiscoverCommand) analyzeWorkflow(mpcData *mpc.MPC) {
	// Categorize nodes by their status and dependencies
	workableNow := []*mpc.Node{}
	blockedNodes := []*mpc.Node{}
	inProgressNodes := []*mpc.Node{}
	completedNodes := []*mpc.Node{}
	
	// Track which nodes have incomplete dependencies
	nodeBlockers := make(map[string][]string)
	
	for i := range mpcData.Nodes {
		node := &mpcData.Nodes[i]
		
		switch node.Status {
		case mpc.StatusCompleted:
			completedNodes = append(completedNodes, node)
		case mpc.StatusInProgress:
			inProgressNodes = append(inProgressNodes, node)
		case mpc.StatusBlocked:
			blockedNodes = append(blockedNodes, node)
		case mpc.StatusReady:
			// For the entry node, it's always workable if Ready
			if node.ID == mpcData.EntryNode {
				workableNow = append(workableNow, node)
			} else {
				// Check if all dependencies are completed
				canWork := true
				blockers := []string{}
				
				// Find nodes that have this node as downstream
				for j := range mpcData.Nodes {
					upstream := &mpcData.Nodes[j]
					for _, downstream := range upstream.Downstream {
						if downstream == node.ID && upstream.Status != mpc.StatusCompleted {
							canWork = false
							blockers = append(blockers, upstream.ID)
						}
					}
				}
				
				if canWork {
					workableNow = append(workableNow, node)
				} else {
					blockedNodes = append(blockedNodes, node)
					nodeBlockers[node.ID] = blockers
				}
			}
		}
	}
	
	// Display workable nodes
	if len(workableNow) == 1 {
		fmt.Println("🚀 READY TO WORK ON NOW:")
	} else if len(workableNow) > 1 {
		fmt.Println("🚀 READY TO WORK ON NOW (can be done in parallel):")
	} else {
		fmt.Println("🚀 READY TO WORK ON NOW:")
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
		fmt.Println("\n⏳ IN PROGRESS:")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range inProgressNodes {
			c.printNodeSummary(node)
		}
	}
	
	// Display blocked nodes with their blockers
	if len(blockedNodes) > 0 {
		fmt.Println("\n🔒 BLOCKED (waiting on dependencies):")
		fmt.Println(strings.Repeat("=", 60))
		for _, node := range blockedNodes {
			c.printNodeSummary(node)
			if blockers, ok := nodeBlockers[node.ID]; ok && len(blockers) > 0 {
				fmt.Printf("     ⤷ Waiting on: %s\n", strings.Join(blockers, ", "))
			}
		}
	}
	
	// Show workflow execution stages
	fmt.Println("\n📋 WORKFLOW EXECUTION STAGES:")
	fmt.Println(strings.Repeat("=", 60))
	c.showExecutionStages(mpcData)
	
	// Summary statistics
	fmt.Println("\n📊 SUMMARY:")
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
		progress := node.GetCompletionPercentage()
		fmt.Printf("     Progress: %.0f%% complete", progress)
		fmt.Printf(" | Materialization: %.1f\n", node.Materialization)
		completedCount := node.GetCompletedSubtaskCount()
		totalCount := len(node.Subtasks)
		fmt.Printf("     Subtasks: %d/%d completed\n", completedCount, totalCount)
	}
	
	if len(node.Downstream) > 0 {
		fmt.Printf("     Unlocks: %s\n", strings.Join(node.Downstream, ", "))
	}
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
			fmt.Printf("    → Unlocks: %s\n", strings.Join(unlockList, ", "))
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
	case mpc.StatusReady:
		return "○"
	case mpc.StatusInProgress:
		return "◐"
	case mpc.StatusBlocked:
		return "■"
	case mpc.StatusCompleted:
		return "●"
	default:
		return "?"
	}
}

func (c *MPCDiscoverCommand) Help() string {
	return `Discover what tasks can be worked on next

This command analyzes the MPC workflow to show:
- Which tasks are ready to work on now (can be done in parallel)
- Which tasks are in progress
- Which tasks are blocked and what they're waiting for
- Sequential and parallel workflow paths

Usage:
  workflows mpc discover [options] <file>

Options:
  --status     Show node status (deprecated, always shown)
  --progress   Show detailed progress information

Arguments:
  file         Path to the MPC workflow file (.yaml, .yml, or .json)

Display Sections:
  🚀 READY TO WORK: Tasks with all dependencies completed
  ⏳ IN PROGRESS: Tasks currently being worked on
  🔒 BLOCKED: Tasks waiting on dependencies
  📋 EXECUTION STAGES: Ordered stages showing parallel/sequential flow
  📊 SUMMARY: Overall workflow statistics

Status Icons:
  ○ Ready, ◐ In Progress, ■ Blocked, ● Completed

Examples:
  # Basic discovery
  workflows mpc discover workflow.yaml

  # Show with detailed progress
  workflows mpc discover workflow.yaml --progress`
}