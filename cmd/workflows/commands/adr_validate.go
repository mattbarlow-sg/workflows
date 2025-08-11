package commands

import (
	"fmt"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/config"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/schema"
)

// ADRValidateCommand implements the ADR validate subcommand
type ADRValidateCommand struct {
	*cli.BaseCommand
}

// NewADRValidateCommand creates a new ADR validate command
func NewADRValidateCommand() *ADRValidateCommand {
	return &ADRValidateCommand{
		BaseCommand: cli.NewBaseCommand(
			"validate",
			"Validate an ADR against the schema",
		),
	}
}

// Execute runs the ADR validate command
func (c *ADRValidateCommand) Execute(args []string) error {
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

	// Get schema path
	cfg := config.New()
	schemaPath := cfg.GetSchemaPath("adr")

	// Validate the file
	result, err := schema.ValidateFile(schemaPath, filePath)
	if err != nil {
		return errors.NewIOError("validating file", err)
	}

	// Report results
	if result.Valid {
		fmt.Printf("✓ ADR file '%s' is valid\n", filePath)
		return nil
	}

	fmt.Printf("✗ ADR file '%s' is invalid\n", filePath)
	fmt.Println("\nValidation errors:")
	for i, err := range result.Errors {
		fmt.Printf("  %d. %s\n", i+1, err)
	}
	return errors.NewValidationError("ADR validation failed", nil)
}

// Usage prints detailed usage for the ADR validate command
func (c *ADRValidateCommand) Usage() {
	fmt.Println("Validate an ADR file against the JSON schema")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows adr validate <adr-file.json>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows adr validate my-adr.json")
	fmt.Println()
	fmt.Println("The validator will check:")
	fmt.Println("  - All required fields are present")
	fmt.Println("  - Field values meet constraints (length, format, enum values)")
	fmt.Println("  - JSON structure matches the schema")
}
