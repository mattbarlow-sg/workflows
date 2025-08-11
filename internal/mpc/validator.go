package mpc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ValidationResult struct {
	Valid      bool
	Errors     []ValidationError
	Warnings   []ValidationWarning
	GlobalBPMN string // Path to global BPMN if configured
}

type ValidationError struct {
	Path    string
	Message string
}

type ValidationWarning struct {
	Path    string
	Message string
}

type Validator struct {
	schemaPath string
	verbose    bool
}

func NewValidator(schemaPath string, verbose bool) *Validator {
	return &Validator{
		schemaPath: schemaPath,
		verbose:    verbose,
	}
}

func (v *Validator) ValidateFile(filePath string) (*ValidationResult, error) {
	// Read and parse file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var mpc MPC
	if err := yaml.Unmarshal(data, &mpc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Auto-detect enriched format and adjust schema path
	actualSchemaPath := v.schemaPath
	if v.isEnrichedFormat(&mpc) {
		// Use enriched schema if enriched format is detected
		actualSchemaPath = strings.Replace(v.schemaPath, "mpc.json", "mpc-enriched.json", 1)
		if v.verbose {
			fmt.Println("Detected enriched MPC format, using mpc-enriched.json schema")
		}
	}

	// Convert to JSON for schema validation
	jsonData, err := json.Marshal(mpc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Perform schema validation
	result := &ValidationResult{
		Valid:      true,
		Errors:     []ValidationError{},
		Warnings:   []ValidationWarning{},
		GlobalBPMN: mpc.GlobalBPMN,
	}

	schemaResult, err := v.validateSchemaWithPath(jsonData, actualSchemaPath)
	if err != nil {
		return nil, err
	}

	if !schemaResult.Valid() {
		result.Valid = false
		for _, err := range schemaResult.Errors() {
			result.Errors = append(result.Errors, ValidationError{
				Path:    err.Field(),
				Message: err.Description(),
			})
		}
	}

	// Perform semantic validation
	semanticErrors, semanticWarnings := v.validateSemantics(&mpc)
	result.Errors = append(result.Errors, semanticErrors...)
	result.Warnings = append(result.Warnings, semanticWarnings...)
	
	if len(semanticErrors) > 0 {
		result.Valid = false
	}

	return result, nil
}

func (v *Validator) validateSchema(jsonData []byte) (*gojsonschema.Result, error) {
	return v.validateSchemaWithPath(jsonData, v.schemaPath)
}

func (v *Validator) validateSchemaWithPath(jsonData []byte, schemaPath string) (*gojsonschema.Result, error) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	
	return gojsonschema.Validate(schemaLoader, documentLoader)
}

// isEnrichedFormat detects if the MPC uses the enriched artifact format
func (v *Validator) isEnrichedFormat(mpc *MPC) bool {
	// Check if any node has enriched artifacts (nested properties or schemas)
	for _, node := range mpc.Nodes {
		if node.Artifacts != nil {
			// Check for enriched properties format
			if node.Artifacts.PropertiesStruct != nil {
				return true
			}
			// Check for enriched schemas format
			if node.Artifacts.SchemasStruct != nil {
				return true
			}
			// Check for enriched specs format
			if node.Artifacts.SpecsStruct != nil {
				return true
			}
			// Check for enriched tests format
			if node.Artifacts.TestsStruct != nil {
				return true
			}
		}
	}
	return false
}

func (v *Validator) validateSemantics(mpc *MPC) ([]ValidationError, []ValidationWarning) {
	errors := []ValidationError{}
	warnings := []ValidationWarning{}

	// Validate global BPMN if specified
	if mpc.GlobalBPMN != "" {
		if !strings.HasSuffix(mpc.GlobalBPMN, ".json") && !strings.HasSuffix(mpc.GlobalBPMN, ".bpmn") {
			warnings = append(warnings, ValidationWarning{
				Path:    "global_bpmn",
				Message: fmt.Sprintf("Global BPMN file should have .json or .bpmn extension, got %s", mpc.GlobalBPMN),
			})
		}
	}

	// Build node ID map
	nodeMap := make(map[string]*Node)
	for i := range mpc.Nodes {
		node := &mpc.Nodes[i]
		if _, exists := nodeMap[node.ID]; exists {
			errors = append(errors, ValidationError{
				Path:    fmt.Sprintf("nodes[%d].id", i),
				Message: fmt.Sprintf("duplicate node ID: %s", node.ID),
			})
		}
		nodeMap[node.ID] = node
	}

	// Validate entry node exists
	if _, exists := nodeMap[mpc.EntryNode]; !exists {
		errors = append(errors, ValidationError{
			Path:    "entry_node",
			Message: fmt.Sprintf("entry node '%s' not found in nodes", mpc.EntryNode),
		})
	}

	// Track upstream dependencies for each node
	upstreamDeps := make(map[string][]string)
	
	// Validate each node
	for i, node := range mpc.Nodes {
		nodePath := fmt.Sprintf("nodes[%d]", i)

		// Validate downstream references and track upstream
		for j, downstreamID := range node.Downstream {
			if _, exists := nodeMap[downstreamID]; !exists {
				errors = append(errors, ValidationError{
					Path:    fmt.Sprintf("%s.downstream[%d]", nodePath, j),
					Message: fmt.Sprintf("downstream node '%s' not found", downstreamID),
				})
			} else {
				upstreamDeps[downstreamID] = append(upstreamDeps[downstreamID], node.ID)
			}
		}

		// Validate status
		if !isValidStatus(node.Status) {
			errors = append(errors, ValidationError{
				Path:    fmt.Sprintf("%s.status", nodePath),
				Message: fmt.Sprintf("invalid status '%s'", node.Status),
			})
		}

		// Validate materialization score
		if node.Materialization < 0 || node.Materialization > 1 {
			errors = append(errors, ValidationError{
				Path:    fmt.Sprintf("%s.materialization", nodePath),
				Message: fmt.Sprintf("materialization must be between 0.0 and 1.0, got %f", node.Materialization),
			})
		}

		// Validate artifacts if present
		if node.Artifacts != nil {
			if err := v.validateArtifacts(node.Artifacts, nodePath); err != nil {
				warnings = append(warnings, ValidationWarning{
					Path:    fmt.Sprintf("%s.artifacts", nodePath),
					Message: err.Error(),
				})
			}
		}

		// Check for circular dependencies
		if hasCircularDependency(node.ID, nodeMap, make(map[string]bool)) {
			errors = append(errors, ValidationError{
				Path:    nodePath,
				Message: fmt.Sprintf("circular dependency detected starting from node '%s'", node.ID),
			})
		}

		// Validate subtask completion consistency with status
		completedCount := node.GetCompletedSubtaskCount()
		totalSubtasks := len(node.Subtasks)
		
		if node.Status == StatusCompleted && completedCount < totalSubtasks {
			warnings = append(warnings, ValidationWarning{
				Path:    fmt.Sprintf("%s.status", nodePath),
				Message: fmt.Sprintf("node marked as 'Completed' but only %d/%d subtasks are completed", completedCount, totalSubtasks),
			})
		}
		
		if node.Status == StatusReady && completedCount > 0 {
			warnings = append(warnings, ValidationWarning{
				Path:    fmt.Sprintf("%s.status", nodePath),
				Message: fmt.Sprintf("node marked as 'Ready' but %d subtasks are already completed", completedCount),
			})
		}
	}

	// Check for unreachable nodes
	reachableNodes := findReachableNodes(mpc.EntryNode, nodeMap)
	for id := range nodeMap {
		if !reachableNodes[id] {
			warnings = append(warnings, ValidationWarning{
				Path:    "nodes",
				Message: fmt.Sprintf("node '%s' is not reachable from entry node '%s'", id, mpc.EntryNode),
			})
		}
	}

	// Check for inconsistent dependency structure
	for nodeID, upstreams := range upstreamDeps {
		if len(upstreams) > 1 {
			// Check if all upstreams lead to the same path
			// This is a warning because some workflows might intentionally have multiple paths
			warnings = append(warnings, ValidationWarning{
				Path:    "nodes",
				Message: fmt.Sprintf("node '%s' has multiple upstream dependencies: %s. This may create ambiguous execution order.", nodeID, strings.Join(upstreams, ", ")),
			})
		}
	}

	return errors, warnings
}

func (v *Validator) validateArtifacts(artifacts *Artifacts, nodePath string) error {
	// Check if at least one artifact is specified
	hasArtifact := false
	
	// BPMN
	if artifacts.BPMN != "" {
		hasArtifact = true
		if !strings.HasSuffix(artifacts.BPMN, ".json") && !strings.HasSuffix(artifacts.BPMN, ".bpmn") {
			return fmt.Errorf("BPMN file should have .json or .bpmn extension")
		}
	}
	
	// Old format validation
	if artifacts.Spec != "" {
		hasArtifact = true
		if !strings.HasSuffix(artifacts.Spec, ".yaml") && !strings.HasSuffix(artifacts.Spec, ".yml") && !strings.HasSuffix(artifacts.Spec, ".json") {
			return fmt.Errorf("spec file should have .yaml, .yml, or .json extension")
		}
	}
	if artifacts.Tests != "" {
		hasArtifact = true
	}
	if artifacts.Properties != "" {
		hasArtifact = true
		if !strings.HasSuffix(artifacts.Properties, ".json") {
			return fmt.Errorf("properties file should have .json extension")
		}
	}
	
	// New format validation
	if artifacts.PropertiesStruct != nil {
		hasArtifact = true
		if artifacts.PropertiesStruct.Invariants != "" && !strings.HasSuffix(artifacts.PropertiesStruct.Invariants, ".json") {
			return fmt.Errorf("invariants file should have .json extension")
		}
		if artifacts.PropertiesStruct.StateProperties != "" && !strings.HasSuffix(artifacts.PropertiesStruct.StateProperties, ".json") {
			return fmt.Errorf("state properties file should have .json extension")
		}
		if artifacts.PropertiesStruct.Generators != "" && !strings.HasSuffix(artifacts.PropertiesStruct.Generators, ".json") && !strings.HasSuffix(artifacts.PropertiesStruct.Generators, ".ts") {
			return fmt.Errorf("generators file should have .json or .ts extension")
		}
	}
	
	if artifacts.SpecsStruct != nil {
		hasArtifact = true
		if artifacts.SpecsStruct.API != "" && !strings.HasSuffix(artifacts.SpecsStruct.API, ".yaml") && !strings.HasSuffix(artifacts.SpecsStruct.API, ".yml") {
			return fmt.Errorf("API spec file should have .yaml or .yml extension")
		}
		if artifacts.SpecsStruct.Models != "" && !strings.HasSuffix(artifacts.SpecsStruct.Models, ".tla") && !strings.HasSuffix(artifacts.SpecsStruct.Models, ".als") {
			return fmt.Errorf("models file should have .tla or .als extension")
		}
		if artifacts.SpecsStruct.Schemas != "" && !strings.HasSuffix(artifacts.SpecsStruct.Schemas, ".json") {
			return fmt.Errorf("schemas file should have .json extension")
		}
	}
	
	if artifacts.TestsStruct != nil {
		hasArtifact = true
		// Test paths can have wildcards, so we don't validate extensions
	}

	if !hasArtifact {
		return fmt.Errorf("artifacts defined but no artifact paths specified")
	}

	return nil
}

func isValidStatus(status string) bool {
	validStatuses := []string{StatusReady, StatusInProgress, StatusBlocked, StatusCompleted, StatusSpecified}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}


func hasCircularDependency(nodeID string, nodeMap map[string]*Node, visited map[string]bool) bool {
	if visited[nodeID] {
		return true
	}
	
	visited[nodeID] = true
	defer delete(visited, nodeID)

	node := nodeMap[nodeID]
	if node == nil {
		return false
	}

	for _, downstream := range node.Downstream {
		if hasCircularDependency(downstream, nodeMap, visited) {
			return true
		}
	}

	return false
}

func findReachableNodes(startNode string, nodeMap map[string]*Node) map[string]bool {
	reachable := make(map[string]bool)
	queue := []string{startNode}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if reachable[current] {
			continue
		}

		reachable[current] = true
		
		if node := nodeMap[current]; node != nil {
			queue = append(queue, node.Downstream...)
		}
	}

	return reachable
}

func (r *ValidationResult) String() string {
	if r.Valid {
		msg := "Validation passed"
		if len(r.Warnings) > 0 {
			msg += fmt.Sprintf(" with %d warning(s)", len(r.Warnings))
		}
		return msg
	}
	return fmt.Sprintf("Validation failed with %d error(s)", len(r.Errors))
}

func (r *ValidationResult) PrintDetails() {
	// Always show global BPMN status
	fmt.Println("\nGlobal BPMN Status:")
	if r.GlobalBPMN != "" {
		fmt.Printf("  Global BPMN Configured: %s\n", r.GlobalBPMN)
	} else {
		fmt.Printf("  Global BPMN Configured: skipped\n")
	}

	if len(r.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range r.Errors {
			if err.Path != "" {
				fmt.Printf("  - %s: %s\n", err.Path, err.Message)
			} else {
				fmt.Printf("  - %s\n", err.Message)
			}
		}
	}

	if len(r.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warn := range r.Warnings {
			if warn.Path != "" {
				fmt.Printf("  - %s: %s\n", warn.Path, warn.Message)
			} else {
				fmt.Printf("  - %s\n", warn.Message)
			}
		}
	}
}