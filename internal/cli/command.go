package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Command represents a CLI command with its own flag set and execution logic
type Command interface {
	// Name returns the command name
	Name() string

	// Description returns a short description of the command
	Description() string

	// Usage prints detailed usage information
	Usage()

	// Execute runs the command with the given arguments
	Execute(args []string) error
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	name        string
	description string
	flagSet     *flag.FlagSet
	output      io.Writer
}

// NewBaseCommand creates a new base command
func NewBaseCommand(name, description string) *BaseCommand {
	return &BaseCommand{
		name:        name,
		description: description,
		flagSet:     flag.NewFlagSet(name, flag.ContinueOnError),
		output:      os.Stdout,
	}
}

// Name returns the command name
func (c *BaseCommand) Name() string {
	return c.name
}

// Description returns the command description
func (c *BaseCommand) Description() string {
	return c.description
}

// FlagSet returns the command's flag set for configuration
func (c *BaseCommand) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// ParseFlags parses the command arguments
func (c *BaseCommand) ParseFlags(args []string) error {
	return c.flagSet.Parse(args)
}

// Usage prints the command usage
func (c *BaseCommand) Usage() {
	fmt.Fprintf(c.output, "Usage: workflows %s [options]\n", c.name)
	fmt.Fprintf(c.output, "\n%s\n", c.description)
	if c.flagSet != nil {
		fmt.Fprintf(c.output, "\nOptions:\n")
		c.flagSet.PrintDefaults()
	}
}

// SetOutput sets the output writer for the command
func (c *BaseCommand) SetOutput(w io.Writer) {
	c.output = w
	if c.flagSet != nil {
		c.flagSet.SetOutput(w)
	}
}

// Args returns the non-flag arguments after parsing
func (c *BaseCommand) Args() []string {
	return c.flagSet.Args()
}

// NArg returns the number of non-flag arguments
func (c *BaseCommand) NArg() int {
	return c.flagSet.NArg()
}

// Arg returns the i'th non-flag argument
func (c *BaseCommand) Arg(i int) string {
	return c.flagSet.Arg(i)
}
