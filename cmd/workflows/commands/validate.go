package commands

import (
	"fmt"
	"os"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/config"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/schema"
)

// ValidateCommand implements the validate command
type ValidateCommand struct {
	*cli.BaseCommand
}

// NewValidateCommand creates a new validate command
func NewValidateCommand() *ValidateCommand {
	cmd := &ValidateCommand{
		BaseCommand: cli.NewBaseCommand(
			"validate",
			"Validate a file against a schema",
		),
	}
	return cmd
}

// Execute runs the validate command
func (c *ValidateCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}
	
	// Check required arguments
	if c.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: workflows validate <schema> <file>")
		return errors.NewUsageError("validate command requires schema name and file path")
	}
	
	schemaName := c.Arg(0)
	filePath := c.Arg(1)
	
	// Validate inputs using the validation chain
	if err := cli.NewValidationChain().
		ValidateSchemaName(schemaName, "schema name").
		ValidateFilePath(filePath, "file path").
		ValidateFileExtension(filePath, []string{".json"}, "file type").
		Error(); err != nil {
		return err
	}
	
	// Load configuration and registry
	cfg := config.New()
	registry := schema.NewRegistry(cfg.SchemaDir)
	
	if err := registry.Discover(); err != nil {
		return errors.NewConfigError("discovering schemas", err)
	}
	
	// Get the schema
	schemaObj, found := registry.Get(schemaName)
	if !found {
		fmt.Fprintln(os.Stderr, "Use 'workflows list' to see available schemas")
		return errors.NewValidationError(fmt.Sprintf("schema '%s' not found", schemaName), nil)
	}
	
	// Validate the file
	result, err := schema.ValidateFile(schemaObj.Path, filePath)
	if err != nil {
		return errors.NewIOError("validating file", err)
	}
	
	// Report results
	if result.Valid {
		fmt.Printf("✓ File '%s' is valid according to schema '%s'\n", filePath, schemaName)
		return nil
	}
	
	fmt.Printf("✗ File '%s' is invalid according to schema '%s'\n", filePath, schemaName)
	fmt.Println("\nValidation errors:")
	for i, err := range result.Errors {
		fmt.Printf("  %d. %s\n", i+1, err)
	}
	return errors.NewValidationError("file validation failed", nil)
}

// Usage prints detailed usage for the validate command
func (c *ValidateCommand) Usage() {
	fmt.Println("Usage: workflows validate <schema> <file>")
	fmt.Println()
	fmt.Println(c.Description())
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  schema    Name of the schema to validate against")
	fmt.Println("  file      Path to the JSON file to validate")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows validate config-schema config.json")
	fmt.Println("  workflows validate adr-schema docs/adr/0001-example.json")
}