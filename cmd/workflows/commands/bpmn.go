package commands

import (
	"github.com/mattbarlow-sg/workflows/internal/cli"
)

// BPMNCommand implements the BPMN parent command with subcommands
type BPMNCommand struct {
	*cli.SubcommandHandler
}

// NewBPMNCommand creates a new BPMN command
func NewBPMNCommand() *BPMNCommand {
	cmd := &BPMNCommand{
		SubcommandHandler: cli.NewSubcommandHandler(
			"bpmn",
			"BPMN 2.0 workflow commands",
		),
	}

	// Register subcommands
	cmd.Register(NewBPMNValidateCommand())
	cmd.Register(NewBPMNAnalyzeCommand())
	cmd.Register(NewBPMNRenderCommand())

	return cmd
}

// Usage prints BPMN command usage with examples
func (c *BPMNCommand) Usage() {
	// Call parent usage first
	c.SubcommandHandler.Usage()

	// Add examples
	println()
	println("Examples:")
	println("  workflows bpmn validate process.json")
	println("  workflows bpmn analyze workflow.json")
	println("  workflows bpmn render -format dot process.json")
}
