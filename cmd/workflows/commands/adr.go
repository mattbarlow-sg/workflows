package commands

import (
	"github.com/mattbarlow-sg/workflows/internal/cli"
)

// ADRCommand implements the ADR parent command with subcommands
type ADRCommand struct {
	*cli.SubcommandHandler
}

// NewADRCommand creates a new ADR command
func NewADRCommand() *ADRCommand {
	cmd := &ADRCommand{
		SubcommandHandler: cli.NewSubcommandHandler(
			"adr",
			"Architecture Decision Record commands",
		),
	}
	
	// Register subcommands
	cmd.Register(NewADRNewCommand())
	cmd.Register(NewADRRenderCommand())
	cmd.Register(NewADRValidateCommand())
	
	return cmd
}

// Usage prints ADR command usage with examples
func (c *ADRCommand) Usage() {
	// Call parent usage first
	c.SubcommandHandler.Usage()
	
	// Add examples
	println()
	println("Examples:")
	println("  workflows adr new --title \"Use PostgreSQL\" --status proposed")
	println("  workflows adr render my-adr.json")
	println("  workflows adr validate my-adr.json")
}