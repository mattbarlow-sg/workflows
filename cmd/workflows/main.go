package main

import (
	"fmt"
	"os"

	"github.com/mattbarlow-sg/workflows/cmd/workflows/commands"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

func main() {
	if err := run(); err != nil {
		if cliErr, ok := errors.IsCLIError(err); ok {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cliErr)
			os.Exit(cliErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func run() error {
	// Create command manager
	manager := cli.NewManager(
		"Schema Validation CLI",
		"A tool for managing and validating JSON schemas, Architecture Decision Records (ADRs), BPMN workflows, and MPC (Model Predictive Control) workflows.",
	)

	// Register commands
	if err := manager.Register(commands.NewListCommand()); err != nil {
		return err
	}

	if err := manager.Register(commands.NewValidateCommand()); err != nil {
		return err
	}

	if err := manager.Register(commands.NewADRCommand()); err != nil {
		return err
	}

	if err := manager.Register(commands.NewBPMNCommand()); err != nil {
		return err
	}

	if err := manager.Register(commands.NewMPCCommand()); err != nil {
		return err
	}

	// Execute the command
	return manager.Execute(os.Args[1:])
}
