package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/schema"
)

// ListCommand implements the list command
type ListCommand struct {
	*cli.BaseCommand
}

// NewListCommand creates a new list command
func NewListCommand() *ListCommand {
	return &ListCommand{
		BaseCommand: cli.NewBaseCommand(
			"list",
			"List all available schemas",
		),
	}
}

// Execute runs the list command
func (c *ListCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}
	
	// Create schema registry
	schemasDir := filepath.Join(".", "schemas")
	registry := schema.NewRegistry(schemasDir)
	
	// Discover schemas
	if err := registry.Discover(); err != nil {
		return errors.NewConfigError("discovering schemas", err)
	}
	
	// Get list of schemas
	schemas := registry.List()
	if len(schemas) == 0 {
		fmt.Println("No schemas found in", schemasDir)
		return nil
	}
	
	// Display schemas in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTITLE\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-----\t-----------")
	
	for _, schema := range schemas {
		title := schema.Title
		if title == "" {
			title = "-"
		}
		description := schema.Description
		if description == "" {
			description = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", schema.Name, title, description)
	}
	
	w.Flush()
	return nil
}

// Usage prints detailed usage for the list command
func (c *ListCommand) Usage() {
	fmt.Println("Usage: workflows list")
	fmt.Println()
	fmt.Println(c.Description())
	fmt.Println()
	fmt.Println("This command scans the schemas directory and lists all available")
	fmt.Println("JSON schemas that can be used for validation.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows list")
}