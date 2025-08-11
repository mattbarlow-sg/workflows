package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// Manager manages the registration and execution of commands
type Manager struct {
	commands map[string]Command
	output   io.Writer
	appName  string
	appDesc  string
}

// NewManager creates a new command manager
func NewManager(appName, appDesc string) *Manager {
	return &Manager{
		commands: make(map[string]Command),
		output:   os.Stdout,
		appName:  appName,
		appDesc:  appDesc,
	}
}

// Register adds a new command to the manager
func (m *Manager) Register(cmd Command) error {
	if cmd == nil {
		return fmt.Errorf("cannot register nil command")
	}

	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if _, exists := m.commands[name]; exists {
		return fmt.Errorf("command '%s' already registered", name)
	}

	m.commands[name] = cmd
	return nil
}

// Execute runs the appropriate command based on the arguments
func (m *Manager) Execute(args []string) error {
	if len(args) < 1 {
		m.printUsage()
		return errors.NewUsageError("no command specified")
	}

	cmdName := args[0]

	// Handle help commands
	if cmdName == "help" || cmdName == "--help" || cmdName == "-h" {
		if len(args) > 1 {
			// Help for specific command
			return m.showCommandHelp(args[1])
		}
		m.printUsage()
		return nil
	}

	cmd, exists := m.commands[cmdName]
	if !exists {
		m.printUsage()
		return errors.NewUsageError(fmt.Sprintf("unknown command: %s", cmdName))
	}

	// Execute the command with remaining arguments
	return cmd.Execute(args[1:])
}

// showCommandHelp displays help for a specific command
func (m *Manager) showCommandHelp(cmdName string) error {
	cmd, exists := m.commands[cmdName]
	if !exists {
		return errors.NewUsageError(fmt.Sprintf("unknown command: %s", cmdName))
	}

	cmd.Usage()
	return nil
}

// printUsage prints the general usage information
func (m *Manager) printUsage() {
	fmt.Fprintln(m.output, m.appName)
	if m.appDesc != "" {
		fmt.Fprintln(m.output)
		fmt.Fprintln(m.output, m.appDesc)
	}
	fmt.Fprintln(m.output)
	fmt.Fprintln(m.output, "Usage:")
	fmt.Fprintln(m.output, "  workflows <command> [arguments]")
	fmt.Fprintln(m.output)
	fmt.Fprintln(m.output, "Commands:")

	// Sort commands for consistent output
	var names []string
	for name := range m.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	// Use tabwriter for aligned output
	w := tabwriter.NewWriter(m.output, 0, 0, 2, ' ', 0)
	for _, name := range names {
		cmd := m.commands[name]
		fmt.Fprintf(w, "  %s\t%s\n", name, cmd.Description())
	}
	fmt.Fprintf(w, "  help\tShow this help message\n")
	w.Flush()

	fmt.Fprintln(m.output)
	fmt.Fprintln(m.output, "Run 'workflows help <command>' for more information on a command.")
}

// SetOutput sets the output writer for the manager
func (m *Manager) SetOutput(w io.Writer) {
	m.output = w
}
