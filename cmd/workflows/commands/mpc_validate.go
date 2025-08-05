package commands

import (
	"flag"
	"fmt"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/config"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/mpc"
)

type MPCValidateCommand struct {
	*cli.BaseCommand
	verbose bool
}

func NewMPCValidateCommand() *MPCValidateCommand {
	cmd := &MPCValidateCommand{
		BaseCommand: cli.NewBaseCommand("validate", "Validate an MPC workflow file"),
	}
	
	// Define flags
	cmd.FlagSet().BoolVar(&cmd.verbose, "verbose", false, "Show detailed validation information")
	
	return cmd
}

func (c *MPCValidateCommand) Execute(args []string) error {
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
		return errors.NewUsageError("validate command requires file path")
	}
	
	inputFile := c.Arg(0)
	
	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(inputFile, "file path").
		ValidateFileExtension(inputFile, []string{".yaml", ".yml", ".json"}, "file type").
		Error(); err != nil {
		return err
	}
	
	// Get schema path
	cfg := config.New()
	schemaPath := cfg.GetSchemaPath("mpc")

	// Create validator
	validator := mpc.NewValidator(schemaPath, c.verbose)

	// Validate file
	result, err := validator.ValidateFile(inputFile)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("validation failed: %v", err), err)
	}

	// Print results
	fmt.Println(result.String())
	
	if c.verbose || !result.Valid || len(result.Warnings) > 0 {
		result.PrintDetails()
	}

	if !result.Valid {
		return errors.NewValidationError("MPC workflow validation failed", nil)
	}

	return nil
}

func (c *MPCValidateCommand) Help() string {
	return `Validate an MPC workflow file against the schema

This command performs both schema validation and semantic validation to ensure
the MPC workflow is well-formed and consistent.

Usage:
  workflows mpc validate [options] <file>

Options:
  --verbose    Show detailed validation information

Arguments:
  file         Path to the MPC workflow file (.yaml, .yml, or .json)

Validation includes:
  - JSON schema compliance
  - Node ID uniqueness
  - Entry node existence
  - Downstream node references
  - Circular dependency detection
  - Status and completion consistency
  - Reachability analysis

Examples:
  # Basic validation
  workflows mpc validate workflow.yaml

  # Verbose validation with detailed output
  workflows mpc validate workflow.yaml --verbose`
}