package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mattbarlow-sg/workflows/internal/schema"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "list":
		listCommand(args)
	case "validate":
		validateCommand(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Schema Validation CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                   List all available schemas")
	fmt.Println("  validate <schema> <file>  Validate a file against a schema")
	fmt.Println("  help                   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  workflows list")
	fmt.Println("  workflows validate config-schema config.json")
}

func listCommand(args []string) {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Parse(args)

	schemasDir := filepath.Join(".", "schemas")
	registry := schema.NewRegistry(schemasDir)

	if err := registry.Discover(); err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering schemas: %v\n", err)
		os.Exit(1)
	}

	schemas := registry.List()
	if len(schemas) == 0 {
		fmt.Println("No schemas found in", schemasDir)
		return
	}

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
}

func validateCommand(args []string) {
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	validateCmd.Parse(args)

	if validateCmd.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Error: validate command requires schema name and file path")
		fmt.Fprintln(os.Stderr, "Usage: workflows validate <schema> <file>")
		os.Exit(1)
	}

	schemaName := validateCmd.Arg(0)
	filePath := validateCmd.Arg(1)

	schemasDir := filepath.Join(".", "schemas")
	registry := schema.NewRegistry(schemasDir)

	if err := registry.Discover(); err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering schemas: %v\n", err)
		os.Exit(1)
	}

	schemaObj, found := registry.Get(schemaName)
	if !found {
		fmt.Fprintf(os.Stderr, "Error: schema '%s' not found\n", schemaName)
		fmt.Fprintln(os.Stderr, "Use 'workflows list' to see available schemas")
		os.Exit(1)
	}

	result, err := schema.ValidateFile(schemaObj.Path, filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Printf("✓ File '%s' is valid according to schema '%s'\n", filePath, schemaName)
	} else {
		fmt.Printf("✗ File '%s' is invalid according to schema '%s'\n", filePath, schemaName)
		fmt.Println("\nValidation errors:")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		os.Exit(1)
	}
}