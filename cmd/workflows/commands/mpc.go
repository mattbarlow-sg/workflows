package commands

import (
	"github.com/mattbarlow-sg/workflows/internal/cli"
)

type MPCCommand struct {
	*cli.SubcommandHandler
}

func NewMPCCommand() *MPCCommand {
	cmd := &MPCCommand{
		SubcommandHandler: cli.NewSubcommandHandler("mpc", "Manage MPC (Model Predictive Control) workflows"),
	}

	// Register subcommands
	cmd.Register(NewMPCValidateCommand())
	cmd.Register(NewMPCRenderCommand())
	cmd.Register(NewMPCDiscoverCommand())

	return cmd
}

func (c *MPCCommand) Help() string {
	return `MPC Workflow Management

MPC (Model Predictive Control) workflows define complex, multi-node project plans
with dependencies, subtasks, and progress tracking.

Usage:
  workflows mpc <subcommand> [options]

Available subcommands:
  validate    Validate an MPC workflow file
  render      Render an MPC workflow in different formats
  discover    Discover what tasks can be worked on next

Examples:
  # Validate an MPC workflow
  workflows mpc validate workflow.yaml

  # Render an MPC workflow as text
  workflows mpc render workflow.yaml --format text

  # Render an MPC workflow as YAML
  workflows mpc render workflow.yaml --format yaml -o output.yaml

Use "workflows mpc <subcommand> --help" for more information about a subcommand.`
}