package adr

import (
	"time"
)

// Builder helps construct ADR structs
type Builder struct {
	adr *ADR
}

// NewBuilder creates a new ADR builder
func NewBuilder() *Builder {
	return &Builder{
		adr: &ADR{
			Context: Context{
				Constraints: []string{},
				Assumptions: []string{},
			},
			Options: []Option{},
			Decision: Decision{},
			Consequences: Consequences{
				Positive: []string{},
				Negative: []string{},
				Neutral:  []string{},
			},
			DecisionDrivers: []DecisionDriver{},
			Links: []Link{},
		},
	}
}

// Build returns the constructed ADR
func (b *Builder) Build() *ADR {
	// Generate ID if not set
	if b.adr.ID == "" {
		b.adr.ID = GenerateID()
	}
	
	// Set date if not set
	if b.adr.Date == "" {
		b.adr.Date = time.Now().Format("2006-01-02")
	}
	
	return b.adr
}

// SetTitle sets the ADR title
func (b *Builder) SetTitle(title string) *Builder {
	b.adr.Title = title
	return b
}

// SetStatus sets the ADR status
func (b *Builder) SetStatus(status string) *Builder {
	b.adr.Status = status
	return b
}

// SetDate sets the ADR date
func (b *Builder) SetDate(date string) *Builder {
	b.adr.Date = date
	return b
}

// SetProblemStatement sets the problem statement
func (b *Builder) SetProblemStatement(problem string) *Builder {
	b.adr.Context.Problem = problem
	return b
}

// SetBackgroundContext sets the background context
func (b *Builder) SetBackgroundContext(background string) *Builder {
	b.adr.Context.Background = background
	return b
}

// AddConstraint adds a constraint
func (b *Builder) AddConstraint(constraint string) *Builder {
	b.adr.Context.Constraints = append(b.adr.Context.Constraints, constraint)
	return b
}

// AddAssumption adds an assumption
func (b *Builder) AddAssumption(assumption string) *Builder {
	b.adr.Context.Assumptions = append(b.adr.Context.Assumptions, assumption)
	return b
}

// SetDecision sets the chosen option
func (b *Builder) SetDecision(chosenOption string) *Builder {
	b.adr.Decision.ChosenOption = chosenOption
	return b
}

// SetRationale sets the rationale
func (b *Builder) SetRationale(rationale string) *Builder {
	b.adr.Decision.Rationale = rationale
	return b
}

// AddPositiveConsequence adds a positive consequence
func (b *Builder) AddPositiveConsequence(consequence string) *Builder {
	b.adr.Consequences.Positive = append(b.adr.Consequences.Positive, consequence)
	return b
}

// AddNegativeConsequence adds a negative consequence
func (b *Builder) AddNegativeConsequence(consequence string) *Builder {
	b.adr.Consequences.Negative = append(b.adr.Consequences.Negative, consequence)
	return b
}

// AddNeutralConsequence adds a neutral consequence
func (b *Builder) AddNeutralConsequence(consequence string) *Builder {
	b.adr.Consequences.Neutral = append(b.adr.Consequences.Neutral, consequence)
	return b
}

// AddDecider adds a decider
func (b *Builder) AddDecider(decider string) *Builder {
	if b.adr.Stakeholders == nil {
		b.adr.Stakeholders = &Stakeholders{}
	}
	b.adr.Stakeholders.Deciders = append(b.adr.Stakeholders.Deciders, decider)
	return b
}

// AddConsulted adds a consulted person
func (b *Builder) AddConsulted(person string) *Builder {
	if b.adr.Stakeholders == nil {
		b.adr.Stakeholders = &Stakeholders{}
	}
	b.adr.Stakeholders.Consulted = append(b.adr.Stakeholders.Consulted, person)
	return b
}

// AddInformed adds an informed person
func (b *Builder) AddInformed(person string) *Builder {
	if b.adr.Stakeholders == nil {
		b.adr.Stakeholders = &Stakeholders{}
	}
	b.adr.Stakeholders.Informed = append(b.adr.Stakeholders.Informed, person)
	return b
}

// AddOption adds an option with description
func (b *Builder) AddOption(name, description string) *Builder {
	b.adr.Options = append(b.adr.Options, Option{
		Name:        name,
		Description: description,
	})
	return b
}

// SetTechnicalStory sets the technical story
func (b *Builder) SetTechnicalStory(id, title, description string) *Builder {
	b.adr.TechnicalStory = &TechnicalStory{
		ID:          id,
		Title:       title,
		Description: description,
	}
	return b
}

// AddDecisionDriver adds a decision driver
func (b *Builder) AddDecisionDriver(driver string, weight int) *Builder {
	b.adr.DecisionDrivers = append(b.adr.DecisionDrivers, DecisionDriver{
		Driver: driver,
		Weight: float64(weight),
	})
	return b
}

// AddTag adds a tag
func (b *Builder) AddTag(tag string) *Builder {
	if b.adr.AIMetadata == nil {
		b.adr.AIMetadata = &AIMetadata{}
	}
	b.adr.AIMetadata.Tags = append(b.adr.AIMetadata.Tags, tag)
	return b
}

// AddKeyword adds a keyword
func (b *Builder) AddKeyword(keyword string) *Builder {
	if b.adr.AIMetadata == nil {
		b.adr.AIMetadata = &AIMetadata{}
	}
	b.adr.AIMetadata.Keywords = append(b.adr.AIMetadata.Keywords, keyword)
	return b
}

// SetWorkSessionID sets the work session ID
func (b *Builder) SetWorkSessionID(workSessionID string) *Builder {
	b.adr.WorkSessionID = workSessionID
	return b
}