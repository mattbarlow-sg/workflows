package cli

import (
	"fmt"
	"sort"

	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// SubcommandHandler manages subcommands for a parent command
type SubcommandHandler struct {
	*BaseCommand
	subcommands map[string]Command
}

// NewSubcommandHandler creates a new subcommand handler
func NewSubcommandHandler(name, description string) *SubcommandHandler {
	return &SubcommandHandler{
		BaseCommand: NewBaseCommand(name, description),
		subcommands: make(map[string]Command),
	}
}

// Register adds a subcommand
func (h *SubcommandHandler) Register(cmd Command) error {
	if cmd == nil {
		return fmt.Errorf("cannot register nil subcommand")
	}
	
	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("subcommand name cannot be empty")
	}
	
	if _, exists := h.subcommands[name]; exists {
		return fmt.Errorf("subcommand '%s' already registered", name)
	}
	
	h.subcommands[name] = cmd
	return nil
}

// Execute runs the appropriate subcommand
func (h *SubcommandHandler) Execute(args []string) error {
	if len(args) < 1 {
		h.Usage()
		return errors.NewUsageError(fmt.Sprintf("no %s subcommand specified", h.name))
	}
	
	subcommand := args[0]
	cmd, exists := h.subcommands[subcommand]
	if !exists {
		h.Usage()
		return errors.NewUsageError(fmt.Sprintf("unknown %s subcommand: %s", h.name, subcommand))
	}
	
	return cmd.Execute(args[1:])
}

// Usage prints the subcommand usage
func (h *SubcommandHandler) Usage() {
	fmt.Printf("%s Commands\n", h.name)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  workflows %s <subcommand> [arguments]\n", h.name)
	fmt.Println()
	fmt.Println("Subcommands:")
	
	// Sort subcommands for consistent output
	var names []string
	for name := range h.subcommands {
		names = append(names, name)
	}
	sort.Strings(names)
	
	// Display subcommands with descriptions
	for _, name := range names {
		cmd := h.subcommands[name]
		fmt.Printf("  %-15s %s\n", name, cmd.Description())
	}
}