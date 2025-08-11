package mpc

// MPC represents the Model Predictive Control workflow definition
// Generated from schemas/mpc.json
type MPC struct {
	Version    string       `json:"version" yaml:"version"`                           // Version of the MPC workflow schema
	PlanID     string       `json:"plan_id" yaml:"plan_id"`                           // Unique identifier for the plan
	PlanName   string       `json:"plan_name" yaml:"plan_name"`                       // Name of the plan (plan_id without timestamp)
	GlobalBPMN string       `json:"global_bpmn,omitempty" yaml:"global_bpmn,omitempty"` // Path to global BPMN process diagram for the entire plan (optional)
	Context    Context      `json:"context" yaml:"context"`                           // Business context
	Architecture Architecture `json:"architecture" yaml:"architecture"`                 // Architecture details
	Tooling    Tooling      `json:"tooling" yaml:"tooling"`                           // Tooling configuration
	EntryNode  string       `json:"entry_node" yaml:"entry_node"`                     // ID of the entry node
	Nodes      []Node       `json:"nodes" yaml:"nodes"`                               // MPC workflow nodes
}

// Context represents the business context of the plan
type Context struct {
	BusinessGoal              string   `json:"business_goal" yaml:"business_goal"`                               // The business goal this project aims to achieve
	NonFunctionalRequirements []string `json:"non_functional_requirements" yaml:"non_functional_requirements"`   // List of non-functional requirements
}

// Architecture represents the architecture details
type Architecture struct {
	Overview    string   `json:"overview" yaml:"overview"`       // High-level architecture overview
	ADRs        []string `json:"adrs" yaml:"adrs"`               // Architecture Decision Records
	Constraints []string `json:"constraints" yaml:"constraints"` // Technical and business constraints
}

// Tooling represents the tooling configuration
type Tooling struct {
	PrimaryLanguage    string          `json:"primary_language" yaml:"primary_language"`       // Primary programming language
	SecondaryLanguages []string        `json:"secondary_languages" yaml:"secondary_languages"` // Additional languages used
	Frameworks         []string        `json:"frameworks" yaml:"frameworks"`                   // Frameworks and libraries used
	CodingStandards    CodingStandards `json:"coding_standards" yaml:"coding_standards"`       // Coding standards configuration
}

// CodingStandards represents the coding standards configuration
type CodingStandards struct {
	Lint       string `json:"lint" yaml:"lint"`             // Linting configuration
	Formatting string `json:"formatting" yaml:"formatting"` // Code formatting standards
	Testing    string `json:"testing" yaml:"testing"`       // Testing requirements
}

// Node represents an MPC workflow node
type Node struct {
	ID                  string     `json:"id" yaml:"id"`                                           // Unique node identifier
	Status              string     `json:"status" yaml:"status"`                                   // Current status of the node
	Materialization     float64    `json:"materialization" yaml:"materialization"`                 // Materialization score (0.0 to 1.0)
	Description         string     `json:"description" yaml:"description"`                         // Brief description of the node
	DetailedDescription string     `json:"detailed_description" yaml:"detailed_description"`       // Detailed description of the node
	Subtasks            []Subtask  `json:"subtasks" yaml:"subtasks"`                               // List of subtasks
	Outputs             []string   `json:"outputs,omitempty" yaml:"outputs,omitempty"`             // Outputs produced by this node
	AcceptanceCriteria  []string   `json:"acceptance_criteria" yaml:"acceptance_criteria"`         // Criteria that must be met for acceptance
	DefinitionOfDone    string     `json:"definition_of_done" yaml:"definition_of_done"`           // Definition of when this node is complete
	RequiredKnowledge   []string   `json:"required_knowledge,omitempty" yaml:"required_knowledge,omitempty"` // Knowledge required to complete this node
	Artifacts           *Artifacts `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`         // Associated artifacts (null indicates reviewed but not required)
	Downstream          []string   `json:"downstream" yaml:"downstream"`                           // IDs of downstream nodes
}

// Subtask represents a subtask within a node
type Subtask struct {
	Description string `json:"description" yaml:"description"` // Description of the subtask
	Completed   bool   `json:"completed" yaml:"completed"`     // Whether the subtask is completed
}

// Artifacts represents the artifacts associated with a node
type Artifacts struct {
	BPMN          *string `json:"bpmn,omitempty" yaml:"bpmn,omitempty"`                       // Path to BPMN process diagram (null if reviewed but not required)
	FormalSpec    *string `json:"formal_spec,omitempty" yaml:"formal_spec,omitempty"`         // Path to formal specification with invariants and properties
	Schemas       *string `json:"schemas,omitempty" yaml:"schemas,omitempty"`                 // Path to data schemas and validation contracts
	ModelChecking *string `json:"model_checking,omitempty" yaml:"model_checking,omitempty"`   // Path to TLA+/Alloy model checking specifications
	TestGenerators *string `json:"test_generators,omitempty" yaml:"test_generators,omitempty"` // Path to property-based test generators
}