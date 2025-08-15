// Package workflows provides workflow versioning strategy and implementation
package workflows

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

// WorkflowVersion represents a semantic version for workflows
type WorkflowVersion struct {
	Major        int       `json:"major"`
	Minor        int       `json:"minor"`
	Patch        int       `json:"patch"`
	Prerelease   string    `json:"prerelease,omitempty"`
	BuildMeta    string    `json:"build_meta,omitempty"`
	ReleaseDate  time.Time `json:"release_date"`
	Description  string    `json:"description,omitempty"`
	Deprecated   bool      `json:"deprecated"`
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`
}

// String returns the string representation of the version
func (v WorkflowVersion) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	
	if v.Prerelease != "" {
		version += "-" + v.Prerelease
	}
	
	if v.BuildMeta != "" {
		version += "+" + v.BuildMeta
	}
	
	return version
}

// IsPrerelease returns true if this is a prerelease version
func (v WorkflowVersion) IsPrerelease() bool {
	return v.Prerelease != ""
}

// IsCompatibleWith checks if this version is compatible with another version
func (v WorkflowVersion) IsCompatibleWith(other WorkflowVersion) bool {
	// Major version must match for compatibility
	if v.Major != other.Major {
		return false
	}
	
	// Minor version can be higher (backward compatible)
	if v.Minor < other.Minor {
		return false
	}
	
	// If minor versions match, patch can be higher
	if v.Minor == other.Minor && v.Patch < other.Patch {
		return false
	}
	
	return true
}

// Compare compares this version with another version
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v WorkflowVersion) Compare(other WorkflowVersion) int {
	// Compare major version
	if v.Major < other.Major {
		return -1
	} else if v.Major > other.Major {
		return 1
	}
	
	// Compare minor version
	if v.Minor < other.Minor {
		return -1
	} else if v.Minor > other.Minor {
		return 1
	}
	
	// Compare patch version
	if v.Patch < other.Patch {
		return -1
	} else if v.Patch > other.Patch {
		return 1
	}
	
	// Compare prerelease (prerelease versions are less than normal versions)
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	} else if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	} else if v.Prerelease != "" && other.Prerelease != "" {
		return strings.Compare(v.Prerelease, other.Prerelease)
	}
	
	return 0
}

// VersionedWorkflow extends BaseWorkflow with versioning capabilities
type VersionedWorkflow interface {
	BaseWorkflow
	
	// GetWorkflowVersion returns the workflow version
	GetWorkflowVersion() WorkflowVersion
	
	// GetCompatibleVersions returns a list of compatible versions
	GetCompatibleVersions() []WorkflowVersion
	
	// SupportsVersioning returns true if the workflow supports versioning
	SupportsVersioning() bool
	
	// MigrateFromVersion migrates state from an older version
	MigrateFromVersion(ctx workflow.Context, fromVersion WorkflowVersion, state interface{}) error
}

// VersionManager manages workflow versions
type VersionManager struct {
	versions      map[string][]WorkflowVersion // workflowName -> versions
	workflows     map[string]VersionedWorkflow  // workflowName@version -> workflow
	migrationRules map[string][]MigrationRule   // workflowName -> rules
}

// MigrationRule defines how to migrate from one version to another
type MigrationRule struct {
	FromVersion WorkflowVersion                                                           `json:"from_version"`
	ToVersion   WorkflowVersion                                                           `json:"to_version"`
	Handler     func(ctx workflow.Context, fromState interface{}) (interface{}, error) `json:"-"`
	Description string                                                                    `json:"description"`
	Required    bool                                                                      `json:"required"`
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions:       make(map[string][]WorkflowVersion),
		workflows:      make(map[string]VersionedWorkflow),
		migrationRules: make(map[string][]MigrationRule),
	}
}

// RegisterVersionedWorkflow registers a versioned workflow
func (vm *VersionManager) RegisterVersionedWorkflow(workflow VersionedWorkflow) error {
	name := workflow.GetName()
	version := workflow.GetWorkflowVersion()
	
	// Add version to list
	if versions, exists := vm.versions[name]; exists {
		// Check if version already exists
		for _, v := range versions {
			if v.Compare(version) == 0 {
				return fmt.Errorf("version %s already exists for workflow %s", version.String(), name)
			}
		}
		vm.versions[name] = append(versions, version)
	} else {
		vm.versions[name] = []WorkflowVersion{version}
	}
	
	// Register workflow
	key := fmt.Sprintf("%s@%s", name, version.String())
	vm.workflows[key] = workflow
	
	return nil
}

// GetWorkflow retrieves a workflow by name and version
func (vm *VersionManager) GetWorkflow(name string, version WorkflowVersion) (VersionedWorkflow, error) {
	key := fmt.Sprintf("%s@%s", name, version.String())
	if workflow, exists := vm.workflows[key]; exists {
		return workflow, nil
	}
	return nil, fmt.Errorf("workflow %s version %s not found", name, version.String())
}

// GetLatestVersion returns the latest version of a workflow
func (vm *VersionManager) GetLatestVersion(name string) (WorkflowVersion, error) {
	versions, exists := vm.versions[name]
	if !exists || len(versions) == 0 {
		return WorkflowVersion{}, fmt.Errorf("no versions found for workflow %s", name)
	}
	
	// Sort versions to find the latest
	sortedVersions := make([]WorkflowVersion, len(versions))
	copy(sortedVersions, versions)
	sort.Slice(sortedVersions, func(i, j int) bool {
		return sortedVersions[i].Compare(sortedVersions[j]) > 0
	})
	
	// Return the latest non-prerelease version, or latest prerelease if no release exists
	for _, v := range sortedVersions {
		if !v.IsPrerelease() && !v.Deprecated {
			return v, nil
		}
	}
	
	// If no release version, return latest prerelease
	for _, v := range sortedVersions {
		if !v.Deprecated {
			return v, nil
		}
	}
	
	return sortedVersions[0], nil
}

// GetLatestWorkflow returns the latest version of a workflow
func (vm *VersionManager) GetLatestWorkflow(name string) (VersionedWorkflow, error) {
	version, err := vm.GetLatestVersion(name)
	if err != nil {
		return nil, err
	}
	
	return vm.GetWorkflow(name, version)
}

// GetCompatibleWorkflow finds a compatible workflow version
func (vm *VersionManager) GetCompatibleWorkflow(name string, requiredVersion WorkflowVersion) (VersionedWorkflow, error) {
	versions, exists := vm.versions[name]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", name)
	}
	
	// Find the best compatible version (highest version that's compatible)
	var bestVersion *WorkflowVersion
	for _, v := range versions {
		if v.IsCompatibleWith(requiredVersion) && !v.Deprecated {
			if bestVersion == nil || v.Compare(*bestVersion) > 0 {
				bestVersion = &v
			}
		}
	}
	
	if bestVersion == nil {
		return nil, fmt.Errorf("no compatible version found for workflow %s (required: %s)", 
			name, requiredVersion.String())
	}
	
	return vm.GetWorkflow(name, *bestVersion)
}

// ListVersions returns all versions of a workflow
func (vm *VersionManager) ListVersions(name string) []WorkflowVersion {
	if versions, exists := vm.versions[name]; exists {
		// Return a sorted copy
		result := make([]WorkflowVersion, len(versions))
		copy(result, versions)
		sort.Slice(result, func(i, j int) bool {
			return result[i].Compare(result[j]) > 0
		})
		return result
	}
	return []WorkflowVersion{}
}

// DeprecateVersion marks a version as deprecated
func (vm *VersionManager) DeprecateVersion(name string, version WorkflowVersion, reason string) error {
	versions := vm.versions[name]
	for i, v := range versions {
		if v.Compare(version) == 0 {
			now := time.Now()
			vm.versions[name][i].Deprecated = true
			vm.versions[name][i].DeprecatedAt = &now
			vm.versions[name][i].Description += fmt.Sprintf(" [DEPRECATED: %s]", reason)
			return nil
		}
	}
	return fmt.Errorf("version %s not found for workflow %s", version.String(), name)
}

// AddMigrationRule adds a migration rule between versions
func (vm *VersionManager) AddMigrationRule(workflowName string, rule MigrationRule) error {
	if vm.migrationRules[workflowName] == nil {
		vm.migrationRules[workflowName] = []MigrationRule{}
	}
	
	// Check if rule already exists
	for _, existing := range vm.migrationRules[workflowName] {
		if existing.FromVersion.Compare(rule.FromVersion) == 0 && 
		   existing.ToVersion.Compare(rule.ToVersion) == 0 {
			return fmt.Errorf("migration rule from %s to %s already exists", 
				rule.FromVersion.String(), rule.ToVersion.String())
		}
	}
	
	vm.migrationRules[workflowName] = append(vm.migrationRules[workflowName], rule)
	return nil
}

// GetMigrationPath finds a migration path between two versions
func (vm *VersionManager) GetMigrationPath(workflowName string, fromVersion, toVersion WorkflowVersion) ([]MigrationRule, error) {
	rules := vm.migrationRules[workflowName]
	if len(rules) == 0 {
		return nil, fmt.Errorf("no migration rules defined for workflow %s", workflowName)
	}
	
	// Simple path finding - direct migration or single-hop
	
	// Try direct migration first
	for _, rule := range rules {
		if rule.FromVersion.Compare(fromVersion) == 0 && rule.ToVersion.Compare(toVersion) == 0 {
			return []MigrationRule{rule}, nil
		}
	}
	
	// Try single-hop migration (could be extended to multi-hop with graph algorithms)
	for _, rule1 := range rules {
		if rule1.FromVersion.Compare(fromVersion) == 0 {
			for _, rule2 := range rules {
				if rule2.FromVersion.Compare(rule1.ToVersion) == 0 && rule2.ToVersion.Compare(toVersion) == 0 {
					return []MigrationRule{rule1, rule2}, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("no migration path found from %s to %s", 
		fromVersion.String(), toVersion.String())
}

// ApplyMigration applies a migration between versions
func (vm *VersionManager) ApplyMigration(ctx workflow.Context, workflowName string, 
	fromVersion, toVersion WorkflowVersion, state interface{}) (interface{}, error) {
	
	logger := workflow.GetLogger(ctx)
	logger.Info("Applying workflow migration", 
		"workflow", workflowName,
		"from", fromVersion.String(),
		"to", toVersion.String())
	
	migrationPath, err := vm.GetMigrationPath(workflowName, fromVersion, toVersion)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	
	currentState := state
	for _, rule := range migrationPath {
		logger.Info("Applying migration rule", 
			"from", rule.FromVersion.String(),
			"to", rule.ToVersion.String(),
			"description", rule.Description)
		
		if rule.Handler == nil {
			return nil, fmt.Errorf("migration handler not found for %s to %s", 
				rule.FromVersion.String(), rule.ToVersion.String())
		}
		
		newState, err := rule.Handler(ctx, currentState)
		if err != nil {
			return nil, fmt.Errorf("migration handler failed: %w", err)
		}
		
		currentState = newState
	}
	
	logger.Info("Migration completed successfully")
	return currentState, nil
}

// ParseVersion parses a version string into a WorkflowVersion
func ParseVersion(versionStr string) (WorkflowVersion, error) {
	// Regular expression for semantic versioning
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$`)
	
	matches := re.FindStringSubmatch(versionStr)
	if matches == nil {
		return WorkflowVersion{}, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	prerelease := matches[4]
	buildMeta := matches[5]
	
	return WorkflowVersion{
		Major:       major,
		Minor:       minor,
		Patch:       patch,
		Prerelease:  prerelease,
		BuildMeta:   buildMeta,
		ReleaseDate: time.Now(),
	}, nil
}

// VersionedWorkflowImpl provides a base implementation for versioned workflows
type VersionedWorkflowImpl struct {
	*BaseWorkflowImpl
	version            WorkflowVersion
	compatibleVersions []WorkflowVersion
	supportsVersioning bool
}

// NewVersionedWorkflow creates a new versioned workflow implementation
func NewVersionedWorkflow(metadata WorkflowMetadata, version WorkflowVersion) *VersionedWorkflowImpl {
	return &VersionedWorkflowImpl{
		BaseWorkflowImpl:   NewBaseWorkflow(metadata),
		version:            version,
		compatibleVersions: []WorkflowVersion{},
		supportsVersioning: true,
	}
}

// GetWorkflowVersion returns the workflow version
func (vw *VersionedWorkflowImpl) GetWorkflowVersion() WorkflowVersion {
	return vw.version
}

// GetCompatibleVersions returns compatible versions
func (vw *VersionedWorkflowImpl) GetCompatibleVersions() []WorkflowVersion {
	return vw.compatibleVersions
}

// SupportsVersioning returns true if the workflow supports versioning
func (vw *VersionedWorkflowImpl) SupportsVersioning() bool {
	return vw.supportsVersioning
}

// AddCompatibleVersion adds a compatible version
func (vw *VersionedWorkflowImpl) AddCompatibleVersion(version WorkflowVersion) {
	vw.compatibleVersions = append(vw.compatibleVersions, version)
}

// Execute provides a default implementation - should be overridden
func (vw *VersionedWorkflowImpl) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	// Default implementation - should be overridden by concrete workflows
	return nil, fmt.Errorf("Execute method not implemented for versioned workflow %s", vw.GetName())
}

// MigrateFromVersion provides default migration behavior (no-op)
func (vw *VersionedWorkflowImpl) MigrateFromVersion(ctx workflow.Context, fromVersion WorkflowVersion, state interface{}) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Default migration (no-op)", 
		"from", fromVersion.String(), 
		"to", vw.version.String())
	
	// Default implementation does nothing
	// Subclasses should override this method to provide actual migration logic
	return nil
}

// WithVersioning adds versioning capabilities to a workflow
func WithVersioning(workflow BaseWorkflow, version WorkflowVersion) VersionedWorkflow {
	// Check if already a versioned workflow
	if vw, ok := workflow.(VersionedWorkflow); ok {
		return vw
	}
	
	// Create a new versioned wrapper
	metadata := WorkflowMetadata{
		Name:        workflow.GetName(),
		Version:     version.String(),
		Description: fmt.Sprintf("Versioned wrapper for %s", workflow.GetName()),
		CreatedAt:   time.Now(),
	}
	
	versionedImpl := NewVersionedWorkflow(metadata, version)
	
	// Create an adapter that wraps the original workflow
	adapter := &versionedWorkflowAdapter{
		VersionedWorkflowImpl: versionedImpl,
		wrapped:               workflow,
	}
	
	return adapter
}

// versionedWorkflowAdapter adapts a BaseWorkflow to be versioned
type versionedWorkflowAdapter struct {
	*VersionedWorkflowImpl
	wrapped BaseWorkflow
}

// Execute delegates to the wrapped workflow
func (vwa *versionedWorkflowAdapter) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	return vwa.wrapped.Execute(ctx, input)
}

// Validate delegates to the wrapped workflow
func (vwa *versionedWorkflowAdapter) Validate(input interface{}) error {
	return vwa.wrapped.Validate(input)
}

// GetName returns the wrapped workflow's name
func (vwa *versionedWorkflowAdapter) GetName() string {
	return vwa.wrapped.GetName()
}

// WorkflowVersioningHelper provides utility functions for workflow versioning
type WorkflowVersioningHelper struct {
	versionManager *VersionManager
}

// NewWorkflowVersioningHelper creates a new versioning helper
func NewWorkflowVersioningHelper(vm *VersionManager) *WorkflowVersioningHelper {
	return &WorkflowVersioningHelper{
		versionManager: vm,
	}
}

// GetVersionFromWorkflowID extracts version from workflow ID
func (wvh *WorkflowVersioningHelper) GetVersionFromWorkflowID(workflowID string) (string, WorkflowVersion, error) {
	parts := strings.Split(workflowID, "@")
	if len(parts) != 2 {
		return "", WorkflowVersion{}, fmt.Errorf("invalid workflow ID format: %s", workflowID)
	}
	
	workflowName := parts[0]
	versionStr := parts[1]
	
	version, err := ParseVersion(versionStr)
	if err != nil {
		return "", WorkflowVersion{}, fmt.Errorf("invalid version in workflow ID: %w", err)
	}
	
	return workflowName, version, nil
}

// CreateVersionedWorkflowID creates a versioned workflow ID
func (wvh *WorkflowVersioningHelper) CreateVersionedWorkflowID(workflowName string, version WorkflowVersion, instanceID string) string {
	return fmt.Sprintf("%s@%s-%s", workflowName, version.String(), instanceID)
}

// ShouldUpgrade determines if a workflow should be upgraded to a newer version
func (wvh *WorkflowVersioningHelper) ShouldUpgrade(currentVersion, latestVersion WorkflowVersion, policy UpgradePolicy) bool {
	if currentVersion.Compare(latestVersion) >= 0 {
		return false // Already at or above latest version
	}
	
	switch policy {
	case UpgradePolicyAlways:
		return true
	case UpgradePolicyMajor:
		return latestVersion.Major > currentVersion.Major
	case UpgradePolicyMinor:
		return latestVersion.Major > currentVersion.Major || latestVersion.Minor > currentVersion.Minor
	case UpgradePolicyPatch:
		return latestVersion.Major > currentVersion.Major || 
			latestVersion.Minor > currentVersion.Minor || 
			latestVersion.Patch > currentVersion.Patch
	case UpgradePolicyNever:
		return false
	default:
		return false
	}
}

// UpgradePolicy defines when to upgrade workflows
type UpgradePolicy string

const (
	UpgradePolicyAlways UpgradePolicy = "always"  // Upgrade to any newer version
	UpgradePolicyMajor  UpgradePolicy = "major"   // Upgrade only for major version changes
	UpgradePolicyMinor  UpgradePolicy = "minor"   // Upgrade for major or minor version changes
	UpgradePolicyPatch  UpgradePolicy = "patch"   // Upgrade for any version change
	UpgradePolicyNever  UpgradePolicy = "never"   // Never upgrade
)

// WorkflowVersionInfo contains version information for a workflow instance
type WorkflowVersionInfo struct {
	WorkflowName    string          `json:"workflow_name"`
	CurrentVersion  WorkflowVersion `json:"current_version"`
	LatestVersion   WorkflowVersion `json:"latest_version"`
	CanUpgrade      bool            `json:"can_upgrade"`
	UpgradePath     []WorkflowVersion `json:"upgrade_path,omitempty"`
	RequiresManualMigration bool    `json:"requires_manual_migration"`
}

// GetWorkflowVersionInfo returns version information for a workflow
func (wvh *WorkflowVersioningHelper) GetWorkflowVersionInfo(workflowName string, currentVersion WorkflowVersion) (*WorkflowVersionInfo, error) {
	latestVersion, err := wvh.versionManager.GetLatestVersion(workflowName)
	if err != nil {
		return nil, err
	}
	
	info := &WorkflowVersionInfo{
		WorkflowName:   workflowName,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		CanUpgrade:     currentVersion.Compare(latestVersion) < 0,
	}
	
	// Check if migration path exists
	if info.CanUpgrade {
		migrationPath, err := wvh.versionManager.GetMigrationPath(workflowName, currentVersion, latestVersion)
		if err != nil {
			info.RequiresManualMigration = true
		} else {
			info.RequiresManualMigration = false
			// Extract version path from migration rules
			for _, rule := range migrationPath {
				info.UpgradePath = append(info.UpgradePath, rule.ToVersion)
			}
		}
	}
	
	return info, nil
}

// PrebuiltMigrationHandlers provides common migration handlers
var PrebuiltMigrationHandlers = struct {
	// NoOpMigration does nothing - for compatible versions
	NoOpMigration func(ctx workflow.Context, fromState interface{}) (interface{}, error)
	
	// AddFieldMigration adds a new field to the state
	AddFieldMigration func(fieldName string, defaultValue interface{}) func(ctx workflow.Context, fromState interface{}) (interface{}, error)
	
	// RemoveFieldMigration removes a field from the state
	RemoveFieldMigration func(fieldName string) func(ctx workflow.Context, fromState interface{}) (interface{}, error)
	
	// RenameFieldMigration renames a field in the state
	RenameFieldMigration func(oldName, newName string) func(ctx workflow.Context, fromState interface{}) (interface{}, error)
}{
	NoOpMigration: func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
		return fromState, nil
	},
	
	AddFieldMigration: func(fieldName string, defaultValue interface{}) func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
		return func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
			// This is a simplified implementation
			// In practice, you'd use reflection or specific type handling
			if stateMap, ok := fromState.(map[string]interface{}); ok {
				stateMap[fieldName] = defaultValue
				return stateMap, nil
			}
			return fromState, nil
		}
	},
	
	RemoveFieldMigration: func(fieldName string) func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
		return func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
			if stateMap, ok := fromState.(map[string]interface{}); ok {
				delete(stateMap, fieldName)
				return stateMap, nil
			}
			return fromState, nil
		}
	},
	
	RenameFieldMigration: func(oldName, newName string) func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
		return func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
			if stateMap, ok := fromState.(map[string]interface{}); ok {
				if value, exists := stateMap[oldName]; exists {
					stateMap[newName] = value
					delete(stateMap, oldName)
				}
				return stateMap, nil
			}
			return fromState, nil
		}
	},
}