package mpc

type MPC struct {
	Version      string       `json:"version" yaml:"version"`
	PlanID       string       `json:"plan_id" yaml:"plan_id"`
	PlanName     string       `json:"plan_name" yaml:"plan_name"`
	GlobalBPMN   string       `json:"global_bpmn,omitempty" yaml:"global_bpmn,omitempty"`
	Context      Context      `json:"context" yaml:"context"`
	Architecture Architecture `json:"architecture" yaml:"architecture"`
	Tooling      Tooling      `json:"tooling" yaml:"tooling"`
	EntryNode    string       `json:"entry_node" yaml:"entry_node"`
	Nodes        []Node       `json:"nodes" yaml:"nodes"`
}

type Context struct {
	BusinessGoal              string   `json:"business_goal" yaml:"business_goal"`
	NonFunctionalRequirements []string `json:"non_functional_requirements" yaml:"non_functional_requirements"`
}

type Architecture struct {
	Overview    string   `json:"overview" yaml:"overview"`
	ADRs        []string `json:"adrs" yaml:"adrs"`
	Constraints []string `json:"constraints" yaml:"constraints"`
}

type Tooling struct {
	PrimaryLanguage    string          `json:"primary_language" yaml:"primary_language"`
	SecondaryLanguages []string        `json:"secondary_languages" yaml:"secondary_languages"`
	Frameworks         []string        `json:"frameworks" yaml:"frameworks"`
	CodingStandards    CodingStandards `json:"coding_standards" yaml:"coding_standards"`
}

type CodingStandards struct {
	Lint       string `json:"lint" yaml:"lint"`
	Formatting string `json:"formatting" yaml:"formatting"`
	Testing    string `json:"testing" yaml:"testing"`
}

type Node struct {
	ID                  string     `json:"id" yaml:"id"`
	Status              string     `json:"status" yaml:"status"`
	Materialization     float64    `json:"materialization" yaml:"materialization"`
	Description         string     `json:"description" yaml:"description"`
	DetailedDescription string     `json:"detailed_description" yaml:"detailed_description"`
	Subtasks            []Subtask  `json:"subtasks" yaml:"subtasks"`
	Outputs             []string   `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	AcceptanceCriteria  []string   `json:"acceptance_criteria" yaml:"acceptance_criteria"`
	DefinitionOfDone    string     `json:"definition_of_done" yaml:"definition_of_done"`
	RequiredKnowledge   []string   `json:"required_knowledge,omitempty" yaml:"required_knowledge,omitempty"`
	Artifacts           *Artifacts `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
	Downstream          []string   `json:"downstream" yaml:"downstream"`
}

type Subtask struct {
	Description string `json:"description" yaml:"description"`
	Completed   bool   `json:"completed" yaml:"completed"`
}

// Artifacts supports both old (simple string) and new (structured) formats
type Artifacts struct {
	BPMN string `json:"bpmn,omitempty" yaml:"bpmn,omitempty"`

	// Old format - simple strings (backward compatibility)
	Spec       string `json:"spec,omitempty" yaml:"spec,omitempty"`
	Tests      string `json:"tests,omitempty" yaml:"tests,omitempty"`
	Properties string `json:"properties,omitempty" yaml:"properties,omitempty"`

	// New format - structured (enhanced)
	PropertiesStruct *ArtifactProperties `json:"properties_struct,omitempty" yaml:"properties_struct,omitempty"`
	SchemasStruct    *ArtifactSchemas    `json:"schemas_struct,omitempty" yaml:"schemas_struct,omitempty"`
	SpecsStruct      *ArtifactSpecs      `json:"specs_struct,omitempty" yaml:"specs_struct,omitempty"`
	TestsStruct      *ArtifactTests      `json:"tests_struct,omitempty" yaml:"tests_struct,omitempty"`
}

type ArtifactProperties struct {
	Invariants      string `json:"invariants,omitempty" yaml:"invariants,omitempty"`
	States          string `json:"states,omitempty" yaml:"states,omitempty"`
	Rules           string `json:"rules,omitempty" yaml:"rules,omitempty"`
	TestSpecs       string `json:"test_specs,omitempty" yaml:"test_specs,omitempty"`
	StateProperties string `json:"state_properties,omitempty" yaml:"state_properties,omitempty"` // Legacy
	Generators      string `json:"generators,omitempty" yaml:"generators,omitempty"`
}

type ArtifactSchemas struct {
	Validation      string `json:"validation,omitempty" yaml:"validation,omitempty"`
	Transformations string `json:"transformations,omitempty" yaml:"transformations,omitempty"`
	Contracts       string `json:"contracts,omitempty" yaml:"contracts,omitempty"`
}

type ArtifactSpecs struct {
	API     string `json:"api,omitempty" yaml:"api,omitempty"`
	Models  string `json:"models,omitempty" yaml:"models,omitempty"`
	Schemas string `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

type ArtifactTests struct {
	Property      string `json:"property,omitempty" yaml:"property,omitempty"`
	Deterministic string `json:"deterministic,omitempty" yaml:"deterministic,omitempty"`
	Fuzz          string `json:"fuzz,omitempty" yaml:"fuzz,omitempty"`
	Contract      string `json:"contract,omitempty" yaml:"contract,omitempty"`
	Unit          string `json:"unit,omitempty" yaml:"unit,omitempty"`
	Integration   string `json:"integration,omitempty" yaml:"integration,omitempty"`
	E2E           string `json:"e2e,omitempty" yaml:"e2e,omitempty"`
}

const (
	StatusReady      = "Ready"
	StatusInProgress = "In Progress"
	StatusBlocked    = "Blocked"
	StatusCompleted  = "Completed"
	StatusSpecified  = "Specified"
)

func (m *MPC) GetNodeByID(id string) *Node {
	for i := range m.Nodes {
		if m.Nodes[i].ID == id {
			return &m.Nodes[i]
		}
	}
	return nil
}

func (m *MPC) GetNodeIDs() []string {
	ids := make([]string, len(m.Nodes))
	for i, node := range m.Nodes {
		ids[i] = node.ID
	}
	return ids
}

func (n *Node) GetCompletedSubtaskCount() int {
	count := 0
	for _, subtask := range n.Subtasks {
		if subtask.Completed {
			count++
		}
	}
	return count
}

func (n *Node) GetCompletionPercentage() float64 {
	if len(n.Subtasks) == 0 {
		return 0
	}
	return float64(n.GetCompletedSubtaskCount()) / float64(len(n.Subtasks)) * 100
}
