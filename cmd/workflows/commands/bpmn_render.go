package commands

import (
	"fmt"
	"os"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// BPMNRenderCommand implements the BPMN render subcommand
type BPMNRenderCommand struct {
	*cli.BaseCommand
	format string
	output string
}

// NewBPMNRenderCommand creates a new BPMN render command
func NewBPMNRenderCommand() *BPMNRenderCommand {
	cmd := &BPMNRenderCommand{
		BaseCommand: cli.NewBaseCommand(
			"render",
			"Render BPMN process to various formats",
		),
	}

	// Define flags
	cmd.FlagSet().StringVar(&cmd.format, "format", "markdown", "Output format: markdown, dot, mermaid, text")
	cmd.FlagSet().StringVar(&cmd.output, "output", "", "Output file path (optional)")

	return cmd
}

// Execute runs the BPMN render command
func (c *BPMNRenderCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("render command requires file path")
	}

	filePath := c.Arg(0)

	// Validate inputs
	if err := cli.NewValidationChain().
		ValidateFilePath(filePath, "file path").
		ValidateFileExtension(filePath, []string{".json"}, "file type").
		Error(); err != nil {
		return err
	}

	// Validate format
	validFormats := []string{"markdown", "dot", "mermaid", "text"}
	valid := false
	for _, f := range validFormats {
		if c.format == f {
			valid = true
			break
		}
	}
	if !valid {
		return errors.NewValidationError(fmt.Sprintf("invalid format '%s', must be one of: %v", c.format, validFormats), nil)
	}

	// Create renderer
	renderer := &bpmn.FileRenderer{}

	// Render the file
	var output string
	var err error

	switch c.format {
	case "markdown":
		output, err = renderer.RenderMarkdownFile(filePath)
	case "dot":
		output, err = renderer.RenderDotFile(filePath)
	case "mermaid":
		output, err = renderer.RenderMermaidFile(filePath)
	case "text":
		output, err = renderer.RenderTextFile(filePath)
	}

	if err != nil {
		return errors.NewIOError(fmt.Sprintf("rendering BPMN as %s", c.format), err)
	}

	// Output the result
	if c.output == "" {
		fmt.Print(output)
	} else {
		if err := os.WriteFile(c.output, []byte(output), 0644); err != nil {
			return errors.NewIOError("writing output file", err)
		}
		fmt.Printf("✓ BPMN rendered to %s\n", c.output)
	}

	return nil
}

// Usage prints detailed usage for the BPMN render command
func (c *BPMNRenderCommand) Usage() {
	fmt.Println("Render BPMN process to various formats")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows bpmn render [flags] <file>")
	fmt.Println()
	fmt.Println("Flags:")
	c.FlagSet().PrintDefaults()
	fmt.Println()
	fmt.Println("Output Formats:")
	fmt.Println("  dot      - GraphViz DOT format for visualization")
	fmt.Println("  mermaid  - Mermaid diagram format")
	fmt.Println("  text     - Simple text representation")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Render to stdout")
	fmt.Println("  workflows bpmn render process.json")
	fmt.Println()
	fmt.Println("  # Render as Mermaid diagram")
	fmt.Println("  workflows bpmn render -format mermaid process.json")
	fmt.Println()
	fmt.Println("  # Render to file")
	fmt.Println("  workflows bpmn render -format dot -output process.dot process.json")
}
