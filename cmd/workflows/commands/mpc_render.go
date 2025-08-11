package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/mpc"
)

type MPCRenderCommand struct {
	*cli.BaseCommand
	format string
	output string
}

func NewMPCRenderCommand() *MPCRenderCommand {
	cmd := &MPCRenderCommand{
		BaseCommand: cli.NewBaseCommand("render", "Render an MPC workflow in different formats"),
	}

	// Define flags
	cmd.FlagSet().StringVar(&cmd.format, "format", "yaml", "Output format (yaml, json, text)")
	cmd.FlagSet().StringVar(&cmd.output, "output", "", "Output file (default: stdout)")

	return cmd
}

func (c *MPCRenderCommand) Execute(args []string) error {
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
		return errors.NewUsageError("render command requires file path")
	}

	inputFile := c.Arg(0)

	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(inputFile, "file path").
		ValidateFileExtension(inputFile, []string{".yaml", ".yml", ".json"}, "file type").
		Error(); err != nil {
		return err
	}

	// Validate format
	validFormats := []string{"yaml", "json", "text"}
	if !contains(validFormats, c.format) {
		return errors.NewUsageError(fmt.Sprintf("invalid format '%s'. Valid formats: %s", c.format, strings.Join(validFormats, ", ")))
	}

	// Load MPC from file
	mpcData, err := mpc.LoadMPCFromFile(inputFile)
	if err != nil {
		return errors.NewIOError(fmt.Sprintf("failed to load MPC file: %v", err), err)
	}

	// Create renderer
	renderer := mpc.NewRenderer(mpcData)

	// Render content
	content, err := renderer.Render(c.format)
	if err != nil {
		return errors.NewIOError(fmt.Sprintf("failed to render MPC: %v", err), err)
	}

	// Output result
	if c.output != "" {
		// Write to file
		if err := os.WriteFile(c.output, []byte(content), 0644); err != nil {
			return errors.NewIOError(fmt.Sprintf("failed to write output file: %v", err), err)
		}
		fmt.Printf("MPC workflow rendered to %s\n", c.output)
	} else {
		// Write to stdout
		fmt.Print(content)
	}

	return nil
}

func (c *MPCRenderCommand) Help() string {
	return `Render an MPC workflow in different formats

This command loads an MPC workflow file and renders it in various output formats.
The YAML format preserves the original structure, JSON provides a structured 
representation, and text format provides a human-readable summary.

Usage:
  workflows mpc render [options] <file>

Options:
  --format <format>    Output format: yaml, json, text (default: yaml)
  -o, --output <file>  Output file (default: stdout)

Arguments:
  file                 Path to the MPC workflow file (.yaml, .yml, or .json)

Output Formats:
  yaml    YAML format (preserves structure)
  json    JSON format (structured data)
  text    Human-readable text summary with statistics

Examples:
  # Render to stdout in YAML format
  workflows mpc render workflow.yaml

  # Render to stdout in text format
  workflows mpc render workflow.yaml --format text

  # Render to file in JSON format
  workflows mpc render workflow.yaml --format json -o output.json

  # Convert YAML to JSON
  workflows mpc render workflow.yaml --format json -o workflow.json`
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
