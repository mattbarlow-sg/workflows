package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/adr"
	"github.com/mattbarlow-sg/workflows/internal/schema"
)

func adrCommand(args []string) {
	if len(args) < 1 {
		printADRUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "new":
		adrNewCommand(subargs)
	case "render":
		adrRenderCommand(subargs)
	case "validate":
		adrValidateCommand(subargs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown ADR subcommand: %s\n", subcommand)
		printADRUsage()
		os.Exit(1)
	}
}

func printADRUsage() {
	fmt.Println("ADR (Architecture Decision Record) Commands")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  workflows adr <subcommand> [arguments]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  new                    Create a new ADR with all required fields")
	fmt.Println("  render <file>          Convert ADR JSON to Markdown")
	fmt.Println("  validate <file>        Validate an ADR against the schema")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Minimal ADR (will prompt for missing required fields)")
	fmt.Println("  workflows adr new -title \"Use React\" -problem \"Need a UI framework\" -background \"Building new app\"")
	fmt.Println()
	fmt.Println("  # Complete ADR with all required fields")
	fmt.Println("  workflows adr new -title \"Use React\" -problem \"Need a UI framework\" \\")
	fmt.Println("    -background \"Building new app\" -chosen \"React\" \\")
	fmt.Println("    -rationale \"Best ecosystem\" -positive \"Great tools\" -negative \"Learning curve\"")
	fmt.Println()
	fmt.Println("  # Render and validate")
	fmt.Println("  workflows adr render adr-001.json")
	fmt.Println("  workflows adr validate adr-001.json")
}

func adrNewCommand(args []string) {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	
	// Required fields
	title := newCmd.String("title", "", "ADR title (required)")
	problem := newCmd.String("problem", "", "Problem statement - what issue are we addressing? (required, min 10 chars)")
	background := newCmd.String("background", "", "Background context - current state and why change is needed (required)")
	chosenOption := newCmd.String("chosen", "", "Chosen option name (required)")
	rationale := newCmd.String("rationale", "", "Rationale - why was this option chosen? (required)")
	
	// Consequences (at least one positive or negative required)
	positive := newCmd.String("positive", "", "Positive consequences (comma-separated)")
	negative := newCmd.String("negative", "", "Negative consequences (comma-separated)")
	neutral := newCmd.String("neutral", "", "Neutral consequences (comma-separated)")
	
	// Optional fields
	status := newCmd.String("status", "draft", "ADR status: draft, proposed, accepted, deprecated, superseded, rejected")
	decidersStr := newCmd.String("deciders", "", "Comma-separated list of deciders")
	consulted := newCmd.String("consulted", "", "Comma-separated list of people consulted")
	informed := newCmd.String("informed", "", "Comma-separated list of people to be informed")
	
	// Context optional fields
	constraints := newCmd.String("constraints", "", "Constraints (comma-separated)")
	assumptions := newCmd.String("assumptions", "", "Assumptions (comma-separated)")
	
	// Options
	options := newCmd.String("options", "", "Considered options (comma-separated names)")
	optionDescs := newCmd.String("option-descs", "", "Option descriptions (comma-separated, same order as options)")
	
	// Technical story
	storyId := newCmd.String("story-id", "", "Technical story/ticket ID")
	storyTitle := newCmd.String("story-title", "", "Technical story title")
	storyDesc := newCmd.String("story-desc", "", "Technical story description")
	
	// Decision drivers
	drivers := newCmd.String("drivers", "", "Decision drivers (comma-separated)")
	driverWeights := newCmd.String("driver-weights", "", "Driver weights 1-5 (comma-separated, same order)")
	
	// AI metadata
	tags := newCmd.String("tags", "", "Tags for categorization (comma-separated)")
	keywords := newCmd.String("keywords", "", "Searchable keywords (comma-separated)")
	
	// Output options
	output := newCmd.String("output", "", "Output file path (optional)")
	format := newCmd.String("format", "json", "Output format: json or markdown")
	
	// Custom usage function for verbose help
	newCmd.Usage = func() {
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
		fmt.Println("Optional Context Flags:")
		fmt.Println("  -constraints string   Technical/business constraints (comma-separated)")
		fmt.Println("  -assumptions string   Assumptions made (comma-separated)")
		fmt.Println()
		fmt.Println("Optional Stakeholder Flags:")
		fmt.Println("  -deciders string      People making the decision (comma-separated)")
		fmt.Println("  -consulted string     People consulted (comma-separated)")
		fmt.Println("  -informed string      People to be informed (comma-separated)")
		fmt.Println()
		fmt.Println("Optional Decision Driver Flags:")
		fmt.Println("  -drivers string       Decision criteria (comma-separated)")
		fmt.Println("  -driver-weights string   Weights 1-5 for each driver (comma-separated)")
		fmt.Println()
		fmt.Println("Optional Option Flags:")
		fmt.Println("  -options string       Other options considered (comma-separated)")
		fmt.Println("  -option-descs string  Descriptions for each option (comma-separated)")
		fmt.Println()
		fmt.Println("Optional Technical Story Flags:")
		fmt.Println("  -story-id string      Issue/ticket ID (e.g., JIRA-123)")
		fmt.Println("  -story-title string   Story title")
		fmt.Println("  -story-desc string    Story description")
		fmt.Println()
		fmt.Println("Optional Metadata Flags:")
		fmt.Println("  -status string        Status: draft, proposed, accepted, etc. (default \"draft\")")
		fmt.Println("  -tags string          Semantic tags (comma-separated)")
		fmt.Println("  -keywords string      Search keywords (comma-separated)")
		fmt.Println()
		fmt.Println("Output Flags:")
		fmt.Println("  -format string        Output format: json or markdown (default \"json\")")
		fmt.Println("  -output string        Output file path (prints to stdout if not specified)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println()
		fmt.Println("  # Minimal valid ADR")
		fmt.Println("  workflows adr new \\")
		fmt.Println("    -title \"Use PostgreSQL for main database\" \\")
		fmt.Println("    -problem \"Need reliable relational database for user data\" \\")
		fmt.Println("    -background \"Currently using SQLite which won't scale\" \\")
		fmt.Println("    -chosen \"PostgreSQL\" \\")
		fmt.Println("    -rationale \"Proven reliability and scalability\" \\")
		fmt.Println("    -positive \"ACID compliance,JSON support,Extensions\"")
		fmt.Println()
		fmt.Println("  # Complete ADR with all details")
		fmt.Println("  workflows adr new \\")
		fmt.Println("    -title \"Adopt microservices architecture\" \\")
		fmt.Println("    -problem \"Monolith is becoming difficult to maintain and scale\" \\")
		fmt.Println("    -background \"Current monolith serves 1M users but development velocity is decreasing\" \\")
		fmt.Println("    -chosen \"Microservices\" \\")
		fmt.Println("    -rationale \"Enables independent scaling and deployment of services\" \\")
		fmt.Println("    -positive \"Independent scaling,Technology diversity,Fault isolation\" \\")
		fmt.Println("    -negative \"Increased complexity,Network latency,Distributed transactions\" \\")
		fmt.Println("    -neutral \"Requires DevOps culture shift\" \\")
		fmt.Println("    -status \"proposed\" \\")
		fmt.Println("    -deciders \"cto,lead-architect,engineering-manager\" \\")
		fmt.Println("    -consulted \"backend-team,devops-team,security-team\" \\")
		fmt.Println("    -informed \"all-engineers,product-team\" \\")
		fmt.Println("    -constraints \"Must maintain 99.9% uptime,Budget limited to $50k/month\" \\")
		fmt.Println("    -assumptions \"Team can learn Kubernetes,Cloud costs will decrease over time\" \\")
		fmt.Println("    -drivers \"Scalability,Development velocity,Operational complexity,Cost\" \\")
		fmt.Println("    -driver-weights \"5,5,3,4\" \\")
		fmt.Println("    -options \"Microservices,Modular monolith,Serverless\" \\")
		fmt.Println("    -option-descs \"Full microservices with k8s,Monolith with clear module boundaries,FaaS approach\" \\")
		fmt.Println("    -story-id \"ARCH-2024-001\" \\")
		fmt.Println("    -story-title \"Migrate to scalable architecture\" \\")
		fmt.Println("    -tags \"architecture,microservices,scalability\" \\")
		fmt.Println("    -keywords \"microservice,kubernetes,scaling,distributed\" \\")
		fmt.Println("    -output \"adr-microservices.json\"")
		fmt.Println()
	}
	
	newCmd.Parse(args)

	// Check for help flag
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		newCmd.Usage()
		os.Exit(0)
	}

	// Validate required fields
	var errors []string
	if *title == "" {
		errors = append(errors, "-title is required")
	}
	if *problem == "" {
		errors = append(errors, "-problem is required")
	} else if len(*problem) < 10 {
		errors = append(errors, "-problem must be at least 10 characters")
	}
	if *background == "" {
		errors = append(errors, "-background is required")
	}
	if *chosenOption == "" {
		errors = append(errors, "-chosen is required")
	}
	if *rationale == "" {
		errors = append(errors, "-rationale is required")
	}
	if *positive == "" && *negative == "" {
		errors = append(errors, "at least one of -positive or -negative is required")
	}

	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "Error: Missing required fields:")
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		fmt.Fprintln(os.Stderr, "\nUse 'workflows adr new --help' for usage information")
		os.Exit(1)
	}

	// Parse comma-separated values
	parseCSV := func(s string) []string {
		if s == "" {
			return []string{}
		}
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	// Create ADR
	adrObj := adr.GenerateTemplate(*title, parseCSV(*decidersStr))
	
	// Set required fields
	adrObj.Status = *status
	adrObj.Context.Problem = *problem
	adrObj.Context.Background = *background
	adrObj.Decision.ChosenOption = *chosenOption
	adrObj.Decision.Rationale = *rationale
	
	// Set consequences
	adrObj.Consequences.Positive = parseCSV(*positive)
	adrObj.Consequences.Negative = parseCSV(*negative)
	adrObj.Consequences.Neutral = parseCSV(*neutral)
	
	// Set optional context
	adrObj.Context.Constraints = parseCSV(*constraints)
	adrObj.Context.Assumptions = parseCSV(*assumptions)
	
	// Set stakeholders
	if *consulted != "" || *informed != "" {
		if adrObj.Stakeholders == nil {
			adrObj.Stakeholders = &adr.Stakeholders{}
		}
		adrObj.Stakeholders.Consulted = parseCSV(*consulted)
		adrObj.Stakeholders.Informed = parseCSV(*informed)
	}
	
	// Set technical story
	if *storyId != "" || *storyTitle != "" || *storyDesc != "" {
		adrObj.TechnicalStory = &adr.TechnicalStory{
			ID:          *storyId,
			Title:       *storyTitle,
			Description: *storyDesc,
		}
	}
	
	// Set decision drivers
	if *drivers != "" {
		driverList := parseCSV(*drivers)
		weights := parseCSV(*driverWeights)
		
		for i, driver := range driverList {
			weight := 3.0 // default weight
			if i < len(weights) {
				// Parse weight, default to 3 if invalid
				fmt.Sscanf(weights[i], "%f", &weight)
				if weight < 1 || weight > 5 {
					weight = 3
				}
			}
			adrObj.DecisionDrivers = append(adrObj.DecisionDrivers, adr.DecisionDriver{
				Driver: driver,
				Weight: weight,
			})
		}
	}
	
	// Set options
	if *options != "" {
		optionList := parseCSV(*options)
		descList := parseCSV(*optionDescs)
		
		for i, opt := range optionList {
			desc := ""
			if i < len(descList) {
				desc = descList[i]
			}
			adrObj.Options = append(adrObj.Options, adr.Option{
				Name:        opt,
				Description: desc,
			})
		}
	}
	
	// Set AI metadata
	if *tags != "" || *keywords != "" {
		adrObj.AIMetadata = &adr.AIMetadata{
			Tags:     parseCSV(*tags),
			Keywords: parseCSV(*keywords),
		}
	}

	// Generate output
	var outputContent string
	var err error

	switch *format {
	case "json":
		outputContent, err = adrObj.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
	case "markdown", "md":
		outputContent = adrObj.ToMarkdown()
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'. Use 'json' or 'markdown'\n", *format)
		os.Exit(1)
	}

	if *output != "" {
		err = ioutil.WriteFile(*output, []byte(outputContent), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ADR created: %s\n", *output)
	} else {
		fmt.Println(outputContent)
	}
}

func adrRenderCommand(args []string) {
	renderCmd := flag.NewFlagSet("render", flag.ExitOnError)
	output := renderCmd.String("output", "", "Output file path (optional)")
	
	renderCmd.Usage = func() {
		fmt.Println("Render an ADR from JSON to Markdown format")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  workflows adr render [flags] <adr-file.json>")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -output string    Output file path (prints to stdout if not specified)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Render to stdout")
		fmt.Println("  workflows adr render my-adr.json")
		fmt.Println()
		fmt.Println("  # Render to file")
		fmt.Println("  workflows adr render my-adr.json -output my-adr.md")
	}
	
	renderCmd.Parse(args)

	if renderCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: render command requires input file path")
		renderCmd.Usage()
		os.Exit(1)
	}

	inputPath := renderCmd.Arg(0)

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	adrObj, err := adr.FromJSON(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ADR JSON: %v\n", err)
		os.Exit(1)
	}

	markdown := adrObj.ToMarkdown()

	if *output != "" {
		err = ioutil.WriteFile(*output, []byte(markdown), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Markdown rendered to: %s\n", *output)
	} else {
		fmt.Println(markdown)
	}
}

func adrValidateCommand(args []string) {
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	
	validateCmd.Usage = func() {
		fmt.Println("Validate an ADR file against the JSON schema")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  workflows adr validate <adr-file.json>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  workflows adr validate my-adr.json")
		fmt.Println()
		fmt.Println("The validator will check:")
		fmt.Println("  - All required fields are present")
		fmt.Println("  - Field values meet constraints (length, format, enum values)")
		fmt.Println("  - JSON structure matches the schema")
	}
	
	validateCmd.Parse(args)

	if validateCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: validate command requires file path")
		validateCmd.Usage()
		os.Exit(1)
	}

	filePath := validateCmd.Arg(0)
	
	schemaPath := filepath.Join(".", "schemas", "adr.json")
	
	result, err := schema.ValidateFile(schemaPath, filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Printf("✓ ADR file '%s' is valid\n", filePath)
		
		data, _ := ioutil.ReadFile(filePath)
		var metadata struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &metadata); err == nil {
			fmt.Printf("  ID: %s\n", metadata.ID)
			fmt.Printf("  Title: %s\n", metadata.Title)
			fmt.Printf("  Status: %s\n", metadata.Status)
		}
	} else {
		fmt.Printf("✗ ADR file '%s' is invalid\n", filePath)
		fmt.Println("\nValidation errors:")
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		os.Exit(1)
	}
}