package commands

import (
	"fmt"
	"strings"

	"github.com/mattbarlow-sg/workflows/internal/cli"
	"github.com/mattbarlow-sg/workflows/internal/errors"
)

// ADRNewFlags holds all the flags for the ADR new command
type ADRNewFlags struct {
	// Required fields
	Title        string
	Problem      string
	Background   string
	ChosenOption string
	Rationale    string

	// Consequences (at least one required)
	Positive string
	Negative string
	Neutral  string

	// Optional fields
	Status    string
	Deciders  string
	Consulted string
	Informed  string

	// Context optional fields
	Constraints string
	Assumptions string

	// Options
	Options     string
	OptionDescs string

	// Technical story
	StoryID    string
	StoryTitle string
	StoryDesc  string

	// Decision drivers
	Drivers       string
	DriverWeights string

	// AI metadata
	Tags     string
	Keywords string

	// Output options
	Output string
	Format string
}

// NewADRNewFlags creates flags with defaults
func NewADRNewFlags() *ADRNewFlags {
	return &ADRNewFlags{
		Status: "draft",
		Format: "json",
	}
}

// Validate checks if required fields are present
func (f *ADRNewFlags) Validate() error {
	chain := cli.NewValidationChain()

	// Required fields
	chain.ValidateRequired(f.Title, "title")
	chain.ValidateRequired(f.Problem, "problem")
	chain.ValidateRequired(f.Background, "background")
	chain.ValidateRequired(f.ChosenOption, "chosen option")
	chain.ValidateRequired(f.Rationale, "rationale")

	// At least one consequence required
	if f.Positive == "" && f.Negative == "" && f.Neutral == "" {
		chain.ValidateRequired("", "at least one consequence (positive, negative, or neutral)")
	}

	// Problem must be at least 10 characters
	if len(f.Problem) < 10 && f.Problem != "" {
		return errors.NewValidationError("problem statement must be at least 10 characters", nil)
	}

	// Validate status if provided
	if f.Status != "" {
		validStatuses := []string{"draft", "proposed", "accepted", "deprecated", "superseded", "rejected"}
		valid := false
		for _, s := range validStatuses {
			if f.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			return errors.NewValidationError(fmt.Sprintf("invalid status '%s', must be one of: %v", f.Status, validStatuses), nil)
		}
	}

	// If driver weights provided, must match number of drivers
	if f.DriverWeights != "" && f.Drivers != "" {
		drivers := strings.Split(f.Drivers, ",")
		weights := strings.Split(f.DriverWeights, ",")
		if len(drivers) != len(weights) {
			return errors.NewValidationError(fmt.Sprintf("number of driver weights (%d) must match number of drivers (%d)", len(weights), len(drivers)), nil)
		}
	}

	return chain.Error()
}
