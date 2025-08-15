package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/bpmn"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/temporal"
	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// BPMNMigrateCommand handles BPMN to Temporal workflow migration
type BPMNMigrateCommand struct {
	*cli.BaseCommand
	output      string
	packageName string
	tests       bool
	strict      bool
	dryRun      bool
	verbose     bool
	timeout     time.Duration
}

// NewBPMNMigrateCommand creates a new BPMN migrate command
func NewBPMNMigrateCommand() *BPMNMigrateCommand {
	cmd := &BPMNMigrateCommand{
		BaseCommand: cli.NewBaseCommand("migrate", "Migrate BPMN workflow to Temporal"),
	}

	// Define flags
	cmd.FlagSet().StringVar(&cmd.output, "output", "./generated", "Output directory for generated files")
	cmd.FlagSet().StringVar(&cmd.output, "o", "./generated", "Output directory for generated files (shorthand)")
	cmd.FlagSet().StringVar(&cmd.packageName, "package", "workflows", "Go package name")
	cmd.FlagSet().StringVar(&cmd.packageName, "p", "workflows", "Go package name (shorthand)")
	cmd.FlagSet().BoolVar(&cmd.tests, "tests", false, "Generate unit tests")
	cmd.FlagSet().BoolVar(&cmd.tests, "t", false, "Generate unit tests (shorthand)")
	cmd.FlagSet().BoolVar(&cmd.strict, "strict", false, "Strict mode - fail on any unsupported element")
	cmd.FlagSet().BoolVar(&cmd.strict, "s", false, "Strict mode - fail on any unsupported element (shorthand)")
	cmd.FlagSet().BoolVar(&cmd.dryRun, "dry-run", false, "Validate without generating files")
	cmd.FlagSet().BoolVar(&cmd.verbose, "verbose", false, "Verbose output")
	cmd.FlagSet().BoolVar(&cmd.verbose, "v", false, "Verbose output (shorthand)")
	cmd.FlagSet().DurationVar(&cmd.timeout, "timeout", 5*time.Minute, "Timeout for conversion process")

	return cmd
}

// Execute runs the BPMN migration command
func (c *BPMNMigrateCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Println(c.Help())
			return nil
		}
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Check required arguments
	if c.NArg() < 1 {
		c.Usage()
		return errors.NewUsageError("migrate command requires BPMN file path")
	}

	bpmnFile := c.Arg(0)

	// Validate input file exists
	if _, err := os.Stat(bpmnFile); os.IsNotExist(err) {
		return errors.NewUsageError(fmt.Sprintf("BPMN file not found: %s", bpmnFile))
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Load BPMN file
	if c.verbose {
		fmt.Printf("Loading BPMN file: %s\n", bpmnFile)
	}

	process, err := c.loadBPMNFile(bpmnFile)
	if err != nil {
		return fmt.Errorf("failed to load BPMN file: %w", err)
	}

	if c.verbose {
		fmt.Printf("Loaded process: %s (%s)\n", process.ProcessInfo.Name, process.ProcessInfo.ID)
	}

	// Create adapter
	adapter := temporal.NewBPMNAdapter()

	// Perform dry run if requested
	if c.dryRun {
		return c.performDryRun(adapter, process)
	}

	// Create conversion configuration
	config := schemas.ConversionConfig{
		SourceFile:  bpmnFile,
		PackageName: c.packageName,
		OutputDir:   c.output,
		Options: schemas.ConversionOptions{
			GenerateTests:  c.tests,
			InlineScripts:  true,
			MaxInlineLines: 50,
			GenerateTODOs:  !c.strict,
			PreserveDocs:   true,
			StrictMode:     c.strict,
			Timeout:        c.timeout,
		},
		ValidationLevel: c.getValidationLevel(),
		AllowPartial:    !c.strict,
	}

	// Perform conversion
	if c.verbose {
		fmt.Println("Starting conversion...")
	}

	result, err := adapter.Convert(ctx, process, config)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Display results
	c.displayResults(result)

	// Write files if successful
	if result.Success {
		if err := c.writeGeneratedFiles(result); err != nil {
			return fmt.Errorf("failed to write generated files: %w", err)
		}

		if c.verbose {
			fmt.Printf("\nGenerated files written to: %s\n", c.output)
		}
	} else {
		fmt.Println("\nConversion failed. No files were generated.")
		return errors.NewUsageError("conversion failed due to errors")
	}

	return nil
}

// Help returns the help text for the command
func (c *BPMNMigrateCommand) Help() string {
	return `Usage: workflows bpmn migrate [options] <bpmn-file>

Migrate a BPMN workflow definition to Temporal workflow code.

Options:
  -o, --output <dir>    Output directory for generated files (default: ./generated)
  -p, --package <name>  Go package name (default: workflows)
  -t, --tests           Generate unit tests
  -s, --strict          Strict mode - fail on any unsupported element
  --dry-run             Validate without generating files
  -v, --verbose         Verbose output
  --timeout <duration>  Timeout for conversion process (default: 5m)

Examples:
  # Basic migration
  workflows bpmn migrate process.bpmn

  # With custom output and package
  workflows bpmn migrate -o ./temporal -p myworkflows process.bpmn

  # Generate with tests
  workflows bpmn migrate --tests process.bpmn

  # Validate only (dry run)
  workflows bpmn migrate --dry-run process.bpmn

  # Strict mode (fail on unsupported elements)
  workflows bpmn migrate --strict process.bpmn
`
}

// loadBPMNFile loads and parses a BPMN file
func (c *BPMNMigrateCommand) loadBPMNFile(path string) (*bpmn.Process, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var process bpmn.Process
	if err := json.Unmarshal(data, &process); err != nil {
		return nil, fmt.Errorf("failed to parse BPMN JSON: %w", err)
	}

	// Validate basic structure
	if process.Type != "bpmn:process" {
		return nil, fmt.Errorf("invalid BPMN file: expected type 'bpmn:process', got '%s'", process.Type)
	}

	if process.Version != "2.0" {
		return nil, fmt.Errorf("unsupported BPMN version: %s (only 2.0 is supported)", process.Version)
	}

	return &process, nil
}

// performDryRun performs validation without generating files
func (c *BPMNMigrateCommand) performDryRun(adapter *temporal.BPMNAdapter, process *bpmn.Process) error {
	fmt.Println("Performing dry run validation...")

	result, err := adapter.ValidateProcess(&process.ProcessInfo)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display validation results
	fmt.Printf("\nValidation Results:\n")
	fmt.Printf("  Valid: %v\n", result.Valid)
	fmt.Printf("  Supported elements: %d\n", result.SupportedElements)
	fmt.Printf("  Unsupported elements: %d\n", result.UnsupportedElements)

	if len(result.DetectedPatterns) > 0 {
		fmt.Printf("\nDetected patterns:\n")
		for _, pattern := range result.DetectedPatterns {
			fmt.Printf("  - %s\n", pattern)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  [%s] %s\n", err.Code, err.Message)
			if err.Element != nil {
				fmt.Printf("    Element: %s (%s)\n", err.Element.ID, err.Element.Type)
			}
			if err.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", err.Suggestion)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warn := range result.Warnings {
			fmt.Printf("  [%s] %s\n", warn.Code, warn.Message)
			if warn.Element != nil {
				fmt.Printf("    Element: %s (%s)\n", warn.Element.ID, warn.Element.Type)
			}
			if warn.Impact != "" {
				fmt.Printf("    Impact: %s\n", warn.Impact)
			}
		}
	}

	fmt.Printf("\nComplexity score: %d\n", result.ComplexityScore)

	if !result.Valid && c.strict {
		return errors.NewUsageError("validation failed in strict mode")
	}

	return nil
}

// displayResults displays conversion results
func (c *BPMNMigrateCommand) displayResults(result *schemas.ConversionResult) {
	fmt.Printf("\nConversion Results:\n")
	fmt.Printf("  Success: %v\n", result.Success)
	fmt.Printf("  Duration: %v\n", result.Metadata.Duration)

	// Display statistics
	stats := result.Metadata.Stats
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Total elements: %d\n", stats.TotalElements)
	fmt.Printf("  Converted: %d\n", stats.ConvertedElements)
	fmt.Printf("  Partial: %d\n", stats.PartialElements)
	fmt.Printf("  Failed: %d\n", stats.FailedElements)
	fmt.Printf("  Skipped: %d\n", stats.SkippedElements)
	fmt.Printf("\nGenerated:\n")
	fmt.Printf("  Workflows: %d\n", stats.GeneratedWorkflows)
	fmt.Printf("  Activities: %d\n", stats.GeneratedActivities)
	fmt.Printf("  Types: %d\n", stats.GeneratedTypes)
	if c.tests {
		fmt.Printf("  Tests: %d\n", stats.GeneratedTests)
	}
	fmt.Printf("  Lines of code: %d\n", stats.LinesOfCode)
	if stats.TODOsGenerated > 0 {
		fmt.Printf("  TODOs: %d\n", stats.TODOsGenerated)
	}

	// Display files
	if len(result.GeneratedFiles) > 0 {
		fmt.Printf("\nGenerated files:\n")
		for _, file := range result.GeneratedFiles {
			fmt.Printf("  %s (%d lines)\n", file.Path, file.Lines)
		}
	}

	// Display issues
	if len(result.Issues) > 0 {
		fmt.Printf("\nIssues:\n")
		for _, issue := range result.Issues {
			severityStr := c.severityString(issue.Severity)
			fmt.Printf("  [%s] %s: %s\n", severityStr, issue.Stage, issue.Message)
			if issue.Details != "" && c.verbose {
				fmt.Printf("    Details: %s\n", issue.Details)
			}
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
	}

	// Display validation errors
	if !result.ValidationResult.Valid {
		if len(result.ValidationResult.Errors) > 0 {
			fmt.Printf("\nValidation errors:\n")
			for _, err := range result.ValidationResult.Errors {
				fmt.Printf("  [%s] %s\n", err.Code, err.Message)
			}
		}
	}
}

// writeGeneratedFiles writes generated files to disk
func (c *BPMNMigrateCommand) writeGeneratedFiles(result *schemas.ConversionResult) error {
	// Create output directory
	if err := os.MkdirAll(c.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write each file
	for _, file := range result.GeneratedFiles {
		// Ensure the file path is relative to output directory
		fullPath := file.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(c.output, file.Path)
		}

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := ioutil.WriteFile(fullPath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		if c.verbose {
			fmt.Printf("  Written: %s\n", fullPath)
		}
	}

	// Write metadata file
	metadataPath := filepath.Join(c.output, "conversion-metadata.json")
	metadataJSON, err := json.MarshalIndent(result.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := ioutil.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// getValidationLevel returns the validation level based on strict mode
func (c *BPMNMigrateCommand) getValidationLevel() string {
	if c.strict {
		return "strict"
	}
	return "lenient"
}

// severityString converts severity to display string
func (c *BPMNMigrateCommand) severityString(severity schemas.Severity) string {
	switch severity {
	case schemas.SeverityCritical:
		return "CRITICAL"
	case schemas.SeverityHigh:
		return "HIGH"
	case schemas.SeverityMedium:
		return "MEDIUM"
	case schemas.SeverityLow:
		return "LOW"
	case schemas.SeverityInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}