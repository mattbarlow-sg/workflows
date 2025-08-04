package commands

import (
	"fmt"
	"os"

	"github.com/mattbarlow-sg/workflows/internal/adr"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// ADRRenderCommand implements the ADR render subcommand
type ADRRenderCommand struct {
	*cli.BaseCommand
	output string
}

// NewADRRenderCommand creates a new ADR render command
func NewADRRenderCommand() *ADRRenderCommand {
	cmd := &ADRRenderCommand{
		BaseCommand: cli.NewBaseCommand(
			"render",
			"Convert ADR JSON to Markdown",
		),
	}
	
	// Define flags
	cmd.FlagSet().StringVar(&cmd.output, "output", "", "Output file path (optional)")
	
	return cmd
}

// Execute runs the ADR render command
func (c *ADRRenderCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}
	
	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("render command requires input file path")
	}
	
	inputPath := c.Arg(0)
	
	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(inputPath, "input file path").
		ValidateFileExtension(inputPath, []string{".json"}, "input file type").
		Error(); err != nil {
		return err
	}
	
	// If output is specified, validate it
	if c.output != "" {
		if err := cli.NewValidationChain().
			ValidateFileExtension(c.output, []string{".md"}, "output file type").
			Error(); err != nil {
			return err
		}
	}
	
	// Render the ADR
	renderer := adr.NewRenderer()
	markdown, err := renderer.RenderFile(inputPath)
	if err != nil {
		if validationErr, ok := err.(*adr.ValidationError); ok {
			fmt.Fprintln(os.Stderr, "Validation errors:")
			for i, e := range validationErr.Errors {
				fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, e)
			}
			return errors.NewValidationError("ADR validation failed", nil)
		}
		return errors.NewIOError("rendering ADR", err)
	}
	
	// Output the result
	if c.output == "" {
		fmt.Print(markdown)
	} else {
		if err := os.WriteFile(c.output, []byte(markdown), 0644); err != nil {
			return errors.NewIOError("writing output file", err)
		}
		fmt.Printf("✓ ADR rendered to %s\n", c.output)
	}
	
	return nil
}

// Usage prints detailed usage for the ADR render command
func (c *ADRRenderCommand) Usage() {
	fmt.Println("Render an ADR from JSON to Markdown format")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows adr render [flags] <adr-file.json>")
	fmt.Println()
	fmt.Println("Flags:")
	c.FlagSet().PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Render to stdout")
	fmt.Println("  workflows adr render my-adr.json")
	fmt.Println()
	fmt.Println("  # Render to file")
	fmt.Println("  workflows adr render my-adr.json -output my-adr.md")
}