package commands

import (
	"fmt"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// BPMNAnalyzeCommand implements the BPMN analyze subcommand
type BPMNAnalyzeCommand struct {
	*cli.BaseCommand
}

// NewBPMNAnalyzeCommand creates a new BPMN analyze command
func NewBPMNAnalyzeCommand() *BPMNAnalyzeCommand {
	return &BPMNAnalyzeCommand{
		BaseCommand: cli.NewBaseCommand(
			"analyze",
			"Analyze a BPMN process",
		),
	}
}

// Execute runs the BPMN analyze command
func (c *BPMNAnalyzeCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("analyze command requires file path")
	}

	filePath := c.Arg(0)

	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(filePath, "file path").
		ValidateFileExtension(filePath, []string{".json"}, "file type").
		Error(); err != nil {
		return err
	}

	// Create analyzer
	analyzer := &bpmn.FileAnalyzer{}

	// Analyze the file
	result, err := analyzer.AnalyzeFile(filePath)
	if err != nil {
		return errors.NewIOError("analyzing BPMN file", err)
	}

	// Display analysis results
	fmt.Printf("BPMN Process Analysis for: %s\n", filePath)
	fmt.Println(strings.Repeat("=", 50))

	// Process metrics
	metrics := result.Metrics
	fmt.Printf("\nProcess Metrics:\n")
	fmt.Printf("  Complexity: %d\n", metrics.Complexity)
	fmt.Printf("  Depth: %d\n", metrics.Depth)
	fmt.Printf("  Width: %d\n", metrics.Width)
	fmt.Printf("  Connectivity: %.2f\n", metrics.Connectivity)

	// Element breakdown
	fmt.Printf("\nElement Breakdown:\n")
	fmt.Printf("  Total Elements: %d\n", metrics.Elements.Total)
	fmt.Printf("  Events: %d\n", metrics.Elements.Events)
	fmt.Printf("  Activities: %d\n", metrics.Elements.Activities)
	fmt.Printf("  Gateways: %d\n", metrics.Elements.Gateways)
	fmt.Printf("  Flows: %d\n", metrics.Elements.Flows)

	// Path analysis
	fmt.Printf("\nPath Analysis:\n")
	fmt.Printf("  Critical Path Length: %d\n", len(result.Paths.CriticalPath))
	fmt.Printf("  Min Path Length: %d\n", result.Paths.MinPathLength)
	fmt.Printf("  Max Path Length: %d\n", result.Paths.MaxPathLength)
	fmt.Printf("  Average Path Length: %.1f\n", result.Paths.AveragePathLength)
	if result.Paths.LoopDetected {
		fmt.Printf("  Loops Detected: %d\n", len(result.Paths.Loops))
	}

	// Reachability issues
	if len(result.Reachability.UnreachableElements) > 0 {
		fmt.Printf("\nUnreachable Elements:\n")
		for _, elem := range result.Reachability.UnreachableElements {
			fmt.Printf("  - %s\n", elem)
		}
	}

	if len(result.Reachability.DeadEndElements) > 0 {
		fmt.Printf("\nDead End Elements:\n")
		for _, elem := range result.Reachability.DeadEndElements {
			fmt.Printf("  - %s\n", elem)
		}
	}

	// Deadlocks
	if len(result.Deadlocks) > 0 {
		fmt.Printf("\nPotential Deadlocks:\n")
		for i, deadlock := range result.Deadlocks {
			fmt.Printf("  %d. %s: %s\n", i+1, deadlock.Type, deadlock.Description)
		}
	}

	// Agent workload
	if len(result.AgentWorkload.AgentTasks) > 0 {
		fmt.Printf("\nAgent Workload:\n")
		for agent, tasks := range result.AgentWorkload.AgentTasks {
			fmt.Printf("  %s: %d tasks\n", agent, len(tasks))
		}
		if len(result.AgentWorkload.UnassignedTasks) > 0 {
			fmt.Printf("  Unassigned: %d tasks\n", len(result.AgentWorkload.UnassignedTasks))
		}
		fmt.Printf("  Workload Balance: %.2f\n", result.AgentWorkload.WorkloadBalance)
		if len(result.AgentWorkload.OverloadedAgents) > 0 {
			fmt.Printf("  Overloaded Agents: %v\n", result.AgentWorkload.OverloadedAgents)
		}
	}

	return nil
}

// Usage prints detailed usage for the BPMN analyze command
func (c *BPMNAnalyzeCommand) Usage() {
	fmt.Println("Analyze a BPMN process")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows bpmn analyze <file>")
	fmt.Println()
	fmt.Println("The analyzer provides:")
	fmt.Println("  - Process metrics and complexity analysis")
	fmt.Println("  - Element counts and breakdown")
	fmt.Println("  - Agent workload distribution")
	fmt.Println("  - Potential issues and recommendations")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows bpmn analyze process.json")
	fmt.Println("  workflows bpmn analyze complex-workflow.json")
}
