package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattbarlow-sg/workflows/internal/adr"
	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// ADRNewCommand implements the ADR new subcommand
type ADRNewCommand struct {
	*cli.BaseCommand
	flags *ADRNewFlags
}

// NewADRNewCommand creates a new ADR new command
func NewADRNewCommand() *ADRNewCommand {
	cmd := &ADRNewCommand{
		BaseCommand: cli.NewBaseCommand(
			"new",
			"Create a new ADR with all required fields",
		),
		flags: NewADRNewFlags(),
	}

	// Register all flags
	f := cmd.FlagSet()

	// Required fields
	f.StringVar(&cmd.flags.Title, "title", "", "ADR title (required)")
	f.StringVar(&cmd.flags.Problem, "problem", "", "Problem statement - what issue are we addressing? (required, min 10 chars)")
	f.StringVar(&cmd.flags.Background, "background", "", "Background context - current state and why change is needed (required)")
	f.StringVar(&cmd.flags.ChosenOption, "chosen", "", "Chosen option name (required)")
	f.StringVar(&cmd.flags.Rationale, "rationale", "", "Rationale - why was this option chosen? (required)")

	// Consequences
	f.StringVar(&cmd.flags.Positive, "positive", "", "Positive consequences (comma-separated)")
	f.StringVar(&cmd.flags.Negative, "negative", "", "Negative consequences (comma-separated)")
	f.StringVar(&cmd.flags.Neutral, "neutral", "", "Neutral consequences (comma-separated)")

	// Optional fields
	f.StringVar(&cmd.flags.Status, "status", "draft", "ADR status: draft, proposed, accepted, deprecated, superseded, rejected")
	f.StringVar(&cmd.flags.Deciders, "deciders", "", "Comma-separated list of deciders")
	f.StringVar(&cmd.flags.Consulted, "consulted", "", "Comma-separated list of people consulted")
	f.StringVar(&cmd.flags.Informed, "informed", "", "Comma-separated list of people to be informed")

	// Context
	f.StringVar(&cmd.flags.Constraints, "constraints", "", "Constraints (comma-separated)")
	f.StringVar(&cmd.flags.Assumptions, "assumptions", "", "Assumptions (comma-separated)")

	// Options
	f.StringVar(&cmd.flags.Options, "options", "", "Considered options (comma-separated names)")
	f.StringVar(&cmd.flags.OptionDescs, "option-descs", "", "Option descriptions (comma-separated, same order as options)")

	// Technical story
	f.StringVar(&cmd.flags.StoryID, "story-id", "", "Technical story/ticket ID")
	f.StringVar(&cmd.flags.StoryTitle, "story-title", "", "Technical story title")
	f.StringVar(&cmd.flags.StoryDesc, "story-desc", "", "Technical story description")

	// Decision drivers
	f.StringVar(&cmd.flags.Drivers, "drivers", "", "Decision drivers (comma-separated)")
	f.StringVar(&cmd.flags.DriverWeights, "driver-weights", "", "Driver weights 1-5 (comma-separated, same order)")

	// AI metadata
	f.StringVar(&cmd.flags.Tags, "tags", "", "Tags for categorization (comma-separated)")
	f.StringVar(&cmd.flags.Keywords, "keywords", "", "Searchable keywords (comma-separated)")

	// Output options
	f.StringVar(&cmd.flags.Output, "output", "", "Output file path (optional)")
	f.StringVar(&cmd.flags.Format, "format", "json", "Output format: json or markdown")

	return cmd
}

// Execute runs the ADR new command
func (c *ADRNewCommand) Execute(args []string) error {
	// Parse flags
	if err := c.ParseFlags(args); err != nil {
		return errors.NewUsageError(fmt.Sprintf("invalid arguments: %v", err))
	}

	// Validate flags
	if err := c.flags.Validate(); err != nil {
		return err
	}

	// Build the ADR
	builder := c.buildADR()

	// Get the ADR data
	adrData := builder.Build()

	// Output based on format
	var output string
	var err error

	if c.flags.Format == "markdown" {
		renderer := adr.NewRenderer()
		output, err = renderer.Render(adrData)
		if err != nil {
			return errors.NewIOError("rendering ADR", err)
		}
	} else {
		// JSON format
		jsonBytes, err := json.MarshalIndent(adrData, "", "  ")
		if err != nil {
			return errors.NewIOError("marshaling ADR", err)
		}
		output = string(jsonBytes)
	}

	// Write output
	if c.flags.Output == "" {
		fmt.Print(output)
	} else {
		if err := os.WriteFile(c.flags.Output, []byte(output), 0644); err != nil {
			return errors.NewIOError("writing output file", err)
		}
		fmt.Printf("✓ ADR created: %s\n", c.flags.Output)
	}

	return nil
}

// buildADR constructs the ADR from flags
func (c *ADRNewCommand) buildADR() *adr.Builder {
	builder := adr.NewBuilder()

	// Set basic fields
	builder.
		SetTitle(c.flags.Title).
		SetStatus(c.flags.Status).
		SetDate(time.Now().Format("2006-01-02"))

	// Set context
	builder.
		SetProblemStatement(c.flags.Problem).
		SetBackgroundContext(c.flags.Background)

	// Add constraints and assumptions
	if c.flags.Constraints != "" {
		for _, constraint := range splitAndTrim(c.flags.Constraints) {
			builder.AddConstraint(constraint)
		}
	}

	if c.flags.Assumptions != "" {
		for _, assumption := range splitAndTrim(c.flags.Assumptions) {
			builder.AddAssumption(assumption)
		}
	}

	// Set decision
	builder.
		SetDecision(c.flags.ChosenOption).
		SetRationale(c.flags.Rationale)

	// Add consequences
	if c.flags.Positive != "" {
		for _, consequence := range splitAndTrim(c.flags.Positive) {
			builder.AddPositiveConsequence(consequence)
		}
	}

	if c.flags.Negative != "" {
		for _, consequence := range splitAndTrim(c.flags.Negative) {
			builder.AddNegativeConsequence(consequence)
		}
	}

	if c.flags.Neutral != "" {
		for _, consequence := range splitAndTrim(c.flags.Neutral) {
			builder.AddNeutralConsequence(consequence)
		}
	}

	// Add stakeholders
	if c.flags.Deciders != "" {
		for _, decider := range splitAndTrim(c.flags.Deciders) {
			builder.AddDecider(decider)
		}
	}

	if c.flags.Consulted != "" {
		for _, person := range splitAndTrim(c.flags.Consulted) {
			builder.AddConsulted(person)
		}
	}

	if c.flags.Informed != "" {
		for _, person := range splitAndTrim(c.flags.Informed) {
			builder.AddInformed(person)
		}
	}

	// Add options
	if c.flags.Options != "" {
		options := splitAndTrim(c.flags.Options)
		descriptions := splitAndTrim(c.flags.OptionDescs)

		for i, option := range options {
			desc := ""
			if i < len(descriptions) {
				desc = descriptions[i]
			}
			builder.AddOption(option, desc)
		}
	}

	// Add technical story
	if c.flags.StoryID != "" || c.flags.StoryTitle != "" {
		builder.SetTechnicalStory(c.flags.StoryID, c.flags.StoryTitle, c.flags.StoryDesc)
	}

	// Add decision drivers
	if c.flags.Drivers != "" {
		drivers := splitAndTrim(c.flags.Drivers)
		weights := splitAndTrim(c.flags.DriverWeights)

		for i, driver := range drivers {
			weight := 3 // default weight
			if i < len(weights) {
				// Parse weight, ignore errors and use default
				fmt.Sscanf(weights[i], "%d", &weight)
			}
			builder.AddDecisionDriver(driver, weight)
		}
	}

	// Add metadata
	if c.flags.Tags != "" {
		for _, tag := range splitAndTrim(c.flags.Tags) {
			builder.AddTag(tag)
		}
	}

	if c.flags.Keywords != "" {
		for _, keyword := range splitAndTrim(c.flags.Keywords) {
			builder.AddKeyword(keyword)
		}
	}

	return builder
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// Usage prints detailed usage for the ADR new command
func (c *ADRNewCommand) Usage() {
	fmt.Println("Create a new Architecture Decision Record (ADR)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows adr new [flags]")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  -title string         ADR title describing the decision")
	fmt.Println("  -problem string       Problem statement (min 10 characters)")
	fmt.Println("  -background string    Background context explaining current state")
	fmt.Println("  -chosen string        Name of the chosen option")
	fmt.Println("  -rationale string     Why this option was chosen")
	fmt.Println()
	fmt.Println("At least one consequence is required:")
	fmt.Println("  -positive string      Positive consequences (comma-separated)")
	fmt.Println("  -negative string      Negative consequences (comma-separated)")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  Run 'workflows adr new --help' to see all available flags")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Minimal ADR")
	fmt.Println("  workflows adr new -title \"Use PostgreSQL\" -problem \"Need a database\" \\")
	fmt.Println("    -background \"Building new app\" -chosen \"PostgreSQL\" \\")
	fmt.Println("    -rationale \"Best fit for our needs\" -positive \"Reliable, feature-rich\"")
	fmt.Println()
	fmt.Println("  # Complete ADR with all details")
	fmt.Println("  workflows adr new -title \"Adopt Microservices Architecture\" \\")
	fmt.Println("    -problem \"Monolith is becoming hard to maintain and scale\" \\")
	fmt.Println("    -background \"Current monolithic application serving 1M users\" \\")
	fmt.Println("    -chosen \"Microservices\" -rationale \"Better scalability and team autonomy\" \\")
	fmt.Println("    -positive \"Independent deployment, Technology flexibility\" \\")
	fmt.Println("    -negative \"Increased complexity, Network latency\" \\")
	fmt.Println("    -options \"Monolith, SOA, Microservices\" \\")
	fmt.Println("    -deciders \"Tech Lead, CTO\" -status proposed -output adr-001.json")
}
