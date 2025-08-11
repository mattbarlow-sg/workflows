package commands

import (
	"fmt"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// BPMNValidateCommand implements the BPMN validate subcommand
type BPMNValidateCommand struct {
	*cli.BaseCommand
	verbose bool
}

// NewBPMNValidateCommand creates a new BPMN validate command
func NewBPMNValidateCommand() *BPMNValidateCommand {
	cmd := &BPMNValidateCommand{
		BaseCommand: cli.NewBaseCommand(
			"validate",
			"Validate a BPMN process file",
		),
	}

	// Define flags
	cmd.FlagSet().BoolVar(&cmd.verbose, "verbose", false, "Show detailed validation results")

	return cmd
}

// Execute runs the BPMN validate command
func (c *BPMNValidateCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("validate command requires file path")
	}

	filePath := c.Arg(0)

	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(filePath, "file path").
		ValidateFileExtension(filePath, []string{".json"}, "file type").
		Error(); err != nil {
		return err
	}

	// Create validator
	validator := &bpmn.FileValidator{}

	// Validate the file
	result, err := validator.ValidateFile(filePath)
	if err != nil {
		return errors.NewIOError("reading BPMN file", err)
	}

	// Report results
	if result.Valid {
		fmt.Printf("✓ BPMN file '%s' is valid\n", filePath)
		if c.verbose && len(result.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for i, warning := range result.Warnings {
				fmt.Printf("  %d. %s\n", i+1, warning)
			}
		}
		return nil
	}

	fmt.Printf("✗ BPMN file '%s' is invalid\n", filePath)

	if len(result.Errors) > 0 {
		fmt.Println("\nValidation errors:")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	if c.verbose && len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for i, warning := range result.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	return errors.NewValidationError("BPMN validation failed", nil)
}

// Usage prints detailed usage for the BPMN validate command
func (c *BPMNValidateCommand) Usage() {
	fmt.Println("Validate a BPMN process file")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows bpmn validate [flags] <file>")
	fmt.Println()
	fmt.Println("Flags:")
	c.FlagSet().PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows bpmn validate process.json")
	fmt.Println("  workflows bpmn validate -verbose workflow.json")
	fmt.Println()
	fmt.Println("The validator checks:")
	fmt.Println("  - JSON schema compliance")
	fmt.Println("  - Required elements (start/end events)")
	fmt.Println("  - Flow connectivity")
	fmt.Println("  - Gateway logic")
	fmt.Println("  - Element references")
}
