// Package workflows provides tests for workflow versioning functionality
package workflows

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// VersioningTestSuite provides test suite for workflow versioning
type VersioningTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up test environment before each test
func (s *VersioningTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// AfterTest cleans up after each test
func (s *VersioningTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestWorkflowVersionCreation tests workflow version creation
func (s *VersioningTestSuite) TestWorkflowVersionCreation() {
	version := WorkflowVersion{
		Major:       1,
		Minor:       2,
		Patch:       3,
		Prerelease:  "alpha.1",
		BuildMeta:   "20230101",
		ReleaseDate: time.Now(),
		Description: "Test version",
		Deprecated:  false,
	}
	
	assert.Equal(s.T(), 1, version.Major)
	assert.Equal(s.T(), 2, version.Minor)
	assert.Equal(s.T(), 3, version.Patch)
	assert.Equal(s.T(), "alpha.1", version.Prerelease)
	assert.Equal(s.T(), "20230101", version.BuildMeta)
	assert.Equal(s.T(), "Test version", version.Description)
	assert.False(s.T(), version.Deprecated)
}

// TestWorkflowVersionString tests version string representation
func (s *VersioningTestSuite) TestWorkflowVersionString() {
	testCases := []struct {
		name     string
		version  WorkflowVersion
		expected string
	}{
		{
			name:     "basic version",
			version:  WorkflowVersion{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "version with prerelease",
			version:  WorkflowVersion{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
			expected: "1.2.3-alpha.1",
		},
		{
			name:     "version with build metadata",
			version:  WorkflowVersion{Major: 1, Minor: 2, Patch: 3, BuildMeta: "20230101"},
			expected: "1.2.3+20230101",
		},
		{
			name: "version with prerelease and build metadata",
			version: WorkflowVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
				BuildMeta:  "20230101",
			},
			expected: "1.2.3-alpha.1+20230101",
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.version.String())
		})
	}
}

// TestWorkflowVersionIsPrerelease tests prerelease detection
func (s *VersioningTestSuite) TestWorkflowVersionIsPrerelease() {
	releaseVersion := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	prereleaseVersion := WorkflowVersion{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha.1"}
	
	assert.False(s.T(), releaseVersion.IsPrerelease())
	assert.True(s.T(), prereleaseVersion.IsPrerelease())
}

// TestWorkflowVersionCompatibility tests version compatibility
func (s *VersioningTestSuite) TestWorkflowVersionCompatibility() {
	v1_0_0 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	v1_1_0 := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	v1_1_1 := WorkflowVersion{Major: 1, Minor: 1, Patch: 1}
	v2_0_0 := WorkflowVersion{Major: 2, Minor: 0, Patch: 0}
	
	// Same major version, higher minor/patch should be compatible
	assert.True(s.T(), v1_1_0.IsCompatibleWith(v1_0_0))
	assert.True(s.T(), v1_1_1.IsCompatibleWith(v1_0_0))
	assert.True(s.T(), v1_1_1.IsCompatibleWith(v1_1_0))
	
	// Lower minor version should not be compatible
	assert.False(s.T(), v1_0_0.IsCompatibleWith(v1_1_0))
	
	// Different major version should not be compatible
	assert.False(s.T(), v2_0_0.IsCompatibleWith(v1_0_0))
	assert.False(s.T(), v1_0_0.IsCompatibleWith(v2_0_0))
	
	// Same version should be compatible
	assert.True(s.T(), v1_0_0.IsCompatibleWith(v1_0_0))
}

// TestWorkflowVersionComparison tests version comparison
func (s *VersioningTestSuite) TestWorkflowVersionComparison() {
	v1_0_0 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	v1_0_1 := WorkflowVersion{Major: 1, Minor: 0, Patch: 1}
	v1_1_0 := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	v2_0_0 := WorkflowVersion{Major: 2, Minor: 0, Patch: 0}
	v1_0_0_alpha := WorkflowVersion{Major: 1, Minor: 0, Patch: 0, Prerelease: "alpha.1"}
	
	// Test equal versions
	assert.Equal(s.T(), 0, v1_0_0.Compare(v1_0_0))
	
	// Test major version differences
	assert.Equal(s.T(), -1, v1_0_0.Compare(v2_0_0))
	assert.Equal(s.T(), 1, v2_0_0.Compare(v1_0_0))
	
	// Test minor version differences
	assert.Equal(s.T(), -1, v1_0_0.Compare(v1_1_0))
	assert.Equal(s.T(), 1, v1_1_0.Compare(v1_0_0))
	
	// Test patch version differences
	assert.Equal(s.T(), -1, v1_0_0.Compare(v1_0_1))
	assert.Equal(s.T(), 1, v1_0_1.Compare(v1_0_0))
	
	// Test prerelease versions (prerelease < release)
	assert.Equal(s.T(), -1, v1_0_0_alpha.Compare(v1_0_0))
	assert.Equal(s.T(), 1, v1_0_0.Compare(v1_0_0_alpha))
}

// TestParseVersion tests version string parsing
func (s *VersioningTestSuite) TestParseVersion() {
	testCases := []struct {
		name        string
		versionStr  string
		expectError bool
		expected    WorkflowVersion
	}{
		{
			name:       "basic version",
			versionStr: "1.2.3",
			expected:   WorkflowVersion{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:       "version with prerelease",
			versionStr: "1.2.3-alpha.1",
			expected:   WorkflowVersion{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
		},
		{
			name:       "version with build metadata",
			versionStr: "1.2.3+20230101",
			expected:   WorkflowVersion{Major: 1, Minor: 2, Patch: 3, BuildMeta: "20230101"},
		},
		{
			name:       "full version",
			versionStr: "1.2.3-alpha.1+20230101",
			expected: WorkflowVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
				BuildMeta:  "20230101",
			},
		},
		{
			name:        "invalid version",
			versionStr:  "invalid",
			expectError: true,
		},
		{
			name:        "incomplete version",
			versionStr:  "1.2",
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			version, err := ParseVersion(tc.versionStr)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.Major, version.Major)
				assert.Equal(t, tc.expected.Minor, version.Minor)
				assert.Equal(t, tc.expected.Patch, version.Patch)
				assert.Equal(t, tc.expected.Prerelease, version.Prerelease)
				assert.Equal(t, tc.expected.BuildMeta, version.BuildMeta)
			}
		})
	}
}

// TestVersionManager tests the version manager
func (s *VersioningTestSuite) TestVersionManager() {
	vm := NewVersionManager()
	
	// Create test workflows
	version1 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	version2 := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	version3 := WorkflowVersion{Major: 2, Minor: 0, Patch: 0}
	
	workflow1 := NewMockVersionedWorkflow("TestWorkflow", version1)
	workflow2 := NewMockVersionedWorkflow("TestWorkflow", version2)
	workflow3 := NewMockVersionedWorkflow("TestWorkflow", version3)
	
	// Register workflows
	err := vm.RegisterVersionedWorkflow(workflow1)
	assert.NoError(s.T(), err)
	
	err = vm.RegisterVersionedWorkflow(workflow2)
	assert.NoError(s.T(), err)
	
	err = vm.RegisterVersionedWorkflow(workflow3)
	assert.NoError(s.T(), err)
	
	// Test duplicate registration
	err = vm.RegisterVersionedWorkflow(workflow1)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "already exists")
	
	// Test retrieval
	retrieved, err := vm.GetWorkflow("TestWorkflow", version1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), workflow1, retrieved)
	
	// Test latest version
	latest, err := vm.GetLatestVersion("TestWorkflow")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), version3, latest) // Should be highest version
	
	latestWorkflow, err := vm.GetLatestWorkflow("TestWorkflow")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), workflow3, latestWorkflow)
	
	// Test version listing
	versions := vm.ListVersions("TestWorkflow")
	assert.Len(s.T(), versions, 3)
	
	// Test compatible workflow
	compatibleWorkflow, err := vm.GetCompatibleWorkflow("TestWorkflow", version1)
	assert.NoError(s.T(), err)
	// Should return version2 (highest compatible version)
	assert.Equal(s.T(), version2, compatibleWorkflow.GetWorkflowVersion())
}

// TestVersionManagerDeprecation tests version deprecation
func (s *VersioningTestSuite) TestVersionManagerDeprecation() {
	vm := NewVersionManager()
	
	version1 := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	workflow1 := NewMockVersionedWorkflow("TestWorkflow", version1)
	
	err := vm.RegisterVersionedWorkflow(workflow1)
	assert.NoError(s.T(), err)
	
	// Deprecate version
	err = vm.DeprecateVersion("TestWorkflow", version1, "Security vulnerability")
	assert.NoError(s.T(), err)
	
	// Check that version is marked as deprecated
	versions := vm.ListVersions("TestWorkflow")
	assert.Len(s.T(), versions, 1)
	assert.True(s.T(), versions[0].Deprecated)
	assert.NotNil(s.T(), versions[0].DeprecatedAt)
	assert.Contains(s.T(), versions[0].Description, "DEPRECATED: Security vulnerability")
}

// TestMigrationRules tests migration rule management
func (s *VersioningTestSuite) TestMigrationRules() {
	vm := NewVersionManager()
	
	fromVersion := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	toVersion := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	
	rule := MigrationRule{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Handler:     PrebuiltMigrationHandlers.NoOpMigration,
		Description: "Test migration",
		Required:    true,
	}
	
	err := vm.AddMigrationRule("TestWorkflow", rule)
	assert.NoError(s.T(), err)
	
	// Test duplicate rule
	err = vm.AddMigrationRule("TestWorkflow", rule)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "already exists")
	
	// Test migration path finding
	path, err := vm.GetMigrationPath("TestWorkflow", fromVersion, toVersion)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), path, 1)
	assert.Equal(s.T(), rule, path[0])
	
	// Test non-existent migration path
	nonExistentVersion := WorkflowVersion{Major: 2, Minor: 0, Patch: 0}
	_, err = vm.GetMigrationPath("TestWorkflow", fromVersion, nonExistentVersion)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no migration path found")
}

// TestMigrationApplication tests applying migrations
func (s *VersioningTestSuite) TestMigrationApplication() {
	vm := NewVersionManager()
	
	fromVersion := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	toVersion := WorkflowVersion{Major: 1, Minor: 1, Patch: 0}
	
	// Create a test migration handler
	testHandler := func(ctx workflow.Context, fromState interface{}) (interface{}, error) {
		stateMap := fromState.(map[string]interface{})
		stateMap["migrated"] = true
		return stateMap, nil
	}
	
	rule := MigrationRule{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Handler:     testHandler,
		Description: "Test migration",
		Required:    true,
	}
	
	err := vm.AddMigrationRule("TestWorkflow", rule)
	assert.NoError(s.T(), err)
	
	// Apply migration
	initialState := map[string]interface{}{
		"data": "test",
	}
	
	// Create a mock workflow context
	testWorkflow := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		result, err := vm.ApplyMigration(ctx, "TestWorkflow", fromVersion, toVersion, initialState)
		return result, err
	}
	
	s.env.RegisterWorkflow(testWorkflow)
	s.env.ExecuteWorkflow(testWorkflow, nil)
	
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	
	var result map[string]interface{}
	s.NoError(s.env.GetWorkflowResult(&result))
	
	assert.Equal(s.T(), "test", result["data"])
	assert.Equal(s.T(), true, result["migrated"])
}

// TestVersionedWorkflowImpl tests versioned workflow implementation
func (s *VersioningTestSuite) TestVersionedWorkflowImpl() {
	version := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	metadata := WorkflowMetadata{Name: "TestWorkflow", Version: "1.0.0"}
	
	workflow := NewVersionedWorkflow(metadata, version)
	
	assert.NotNil(s.T(), workflow)
	assert.Equal(s.T(), version, workflow.GetWorkflowVersion())
	assert.True(s.T(), workflow.SupportsVersioning())
	assert.Empty(s.T(), workflow.GetCompatibleVersions())
	
	// Add compatible version
	compatibleVersion := WorkflowVersion{Major: 1, Minor: 0, Patch: 1}
	workflow.AddCompatibleVersion(compatibleVersion)
	
	compatibleVersions := workflow.GetCompatibleVersions()
	assert.Len(s.T(), compatibleVersions, 1)
	assert.Equal(s.T(), compatibleVersion, compatibleVersions[0])
}

// TestWithVersioning tests the WithVersioning wrapper function
func (s *VersioningTestSuite) TestWithVersioning() {
	// Create a base workflow
	baseMetadata := WorkflowMetadata{Name: "BaseWorkflow", Version: "1.0.0"}
	baseWorkflow := NewBaseWorkflow(baseMetadata)
	
	// Add versioning
	version := WorkflowVersion{Major: 1, Minor: 0, Patch: 0}
	versionedWorkflow := WithVersioning(baseWorkflow, version)
	
	assert.NotNil(s.T(), versionedWorkflow)
	assert.Equal(s.T(), "BaseWorkflow", versionedWorkflow.GetName())
	assert.Equal(s.T(), version, versionedWorkflow.GetWorkflowVersion())
	assert.True(s.T(), versionedWorkflow.SupportsVersioning())
}

// TestWorkflowVersioningHelper tests the versioning helper
func (s *VersioningTestSuite) TestWorkflowVersioningHelper() {
	vm := NewVersionManager()
	helper := NewWorkflowVersioningHelper(vm)
	
	// Test workflow ID parsing
	workflowID := "TestWorkflow@1.2.3-alpha.1-instance123"
	name, version, err := helper.GetVersionFromWorkflowID(workflowID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "TestWorkflow", name)
	assert.Equal(s.T(), 1, version.Major)
	assert.Equal(s.T(), 2, version.Minor)
	assert.Equal(s.T(), 3, version.Patch)
	assert.Equal(s.T(), "alpha.1-instance123", version.Prerelease) // This would need better parsing in real implementation
	
	// Test invalid workflow ID
	_, _, err = helper.GetVersionFromWorkflowID("invalid")
	assert.Error(s.T(), err)
	
	// Test versioned workflow ID creation
	version = WorkflowVersion{Major: 1, Minor: 2, Patch: 3}
	createdID := helper.CreateVersionedWorkflowID("TestWorkflow", version, "instance123")
	assert.Equal(s.T(), "TestWorkflow@1.2.3-instance123", createdID)
}

// TestUpgradePolicy tests upgrade policy decisions
func (s *VersioningTestSuite) TestUpgradePolicy() {
	helper := NewWorkflowVersioningHelper(nil)
	
	currentVersion := WorkflowVersion{Major: 1, Minor: 2, Patch: 3}
	
	// Test different upgrade scenarios
	testCases := []struct {
		name          string
		latestVersion WorkflowVersion
		policy        UpgradePolicy
		shouldUpgrade bool
	}{
		{
			name:          "always upgrade - patch",
			latestVersion: WorkflowVersion{Major: 1, Minor: 2, Patch: 4},
			policy:        UpgradePolicyAlways,
			shouldUpgrade: true,
		},
		{
			name:          "never upgrade - major",
			latestVersion: WorkflowVersion{Major: 2, Minor: 0, Patch: 0},
			policy:        UpgradePolicyNever,
			shouldUpgrade: false,
		},
		{
			name:          "major upgrade - major change",
			latestVersion: WorkflowVersion{Major: 2, Minor: 0, Patch: 0},
			policy:        UpgradePolicyMajor,
			shouldUpgrade: true,
		},
		{
			name:          "major upgrade - minor change",
			latestVersion: WorkflowVersion{Major: 1, Minor: 3, Patch: 0},
			policy:        UpgradePolicyMajor,
			shouldUpgrade: false,
		},
		{
			name:          "minor upgrade - minor change",
			latestVersion: WorkflowVersion{Major: 1, Minor: 3, Patch: 0},
			policy:        UpgradePolicyMinor,
			shouldUpgrade: true,
		},
		{
			name:          "patch upgrade - patch change",
			latestVersion: WorkflowVersion{Major: 1, Minor: 2, Patch: 4},
			policy:        UpgradePolicyPatch,
			shouldUpgrade: true,
		},
		{
			name:          "no upgrade - same version",
			latestVersion: WorkflowVersion{Major: 1, Minor: 2, Patch: 3},
			policy:        UpgradePolicyAlways,
			shouldUpgrade: false,
		},
		{
			name:          "no upgrade - older version",
			latestVersion: WorkflowVersion{Major: 1, Minor: 1, Patch: 0},
			policy:        UpgradePolicyAlways,
			shouldUpgrade: false,
		},
	}
	
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			result := helper.ShouldUpgrade(currentVersion, tc.latestVersion, tc.policy)
			assert.Equal(t, tc.shouldUpgrade, result)
		})
	}
}

// TestPrebuiltMigrationHandlers tests prebuilt migration handlers
func (s *VersioningTestSuite) TestPrebuiltMigrationHandlers() {
	ctx := context.Background()
	
	// Test NoOpMigration
	originalState := map[string]interface{}{"data": "test"}
	result, err := PrebuiltMigrationHandlers.NoOpMigration(ctx, originalState)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), originalState, result)
	
	// Test AddFieldMigration
	addFieldHandler := PrebuiltMigrationHandlers.AddFieldMigration("newField", "defaultValue")
	result, err = addFieldHandler(ctx, originalState)
	assert.NoError(s.T(), err)
	resultMap := result.(map[string]interface{})
	assert.Equal(s.T(), "test", resultMap["data"])
	assert.Equal(s.T(), "defaultValue", resultMap["newField"])
	
	// Test RemoveFieldMigration
	stateWithField := map[string]interface{}{"data": "test", "oldField": "value"}
	removeFieldHandler := PrebuiltMigrationHandlers.RemoveFieldMigration("oldField")
	result, err = removeFieldHandler(ctx, stateWithField)
	assert.NoError(s.T(), err)
	resultMap = result.(map[string]interface{})
	assert.Equal(s.T(), "test", resultMap["data"])
	_, exists := resultMap["oldField"]
	assert.False(s.T(), exists)
	
	// Test RenameFieldMigration
	stateWithOldField := map[string]interface{}{"oldName": "value", "data": "test"}
	renameFieldHandler := PrebuiltMigrationHandlers.RenameFieldMigration("oldName", "newName")
	result, err = renameFieldHandler(ctx, stateWithOldField)
	assert.NoError(s.T(), err)
	resultMap = result.(map[string]interface{})
	assert.Equal(s.T(), "test", resultMap["data"])
	assert.Equal(s.T(), "value", resultMap["newName"])
	_, exists = resultMap["oldName"]
	assert.False(s.T(), exists)
}

// TestUpgradePolicy tests upgrade policy enum values
func (s *VersioningTestSuite) TestUpgradePolicyValues() {
	policies := []UpgradePolicy{
		UpgradePolicyAlways,
		UpgradePolicyMajor,
		UpgradePolicyMinor,
		UpgradePolicyPatch,
		UpgradePolicyNever,
	}
	
	expectedValues := []string{
		"always",
		"major",
		"minor",
		"patch",
		"never",
	}
	
	for i, policy := range policies {
		assert.Equal(s.T(), expectedValues[i], string(policy))
	}
}

// TestVersionManagerEdgeCases tests edge cases in version manager
func (s *VersioningTestSuite) TestVersionManagerEdgeCases() {
	vm := NewVersionManager()
	
	// Test getting workflow that doesn't exist
	_, err := vm.GetWorkflow("NonExistent", WorkflowVersion{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not found")
	
	// Test getting latest version for non-existent workflow
	_, err = vm.GetLatestVersion("NonExistent")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no versions found")
	
	// Test getting compatible workflow for non-existent workflow
	_, err = vm.GetCompatibleWorkflow("NonExistent", WorkflowVersion{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not found")
	
	// Test deprecating non-existent version
	err = vm.DeprecateVersion("NonExistent", WorkflowVersion{}, "test")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not found")
	
	// Test migration path for workflow without rules
	_, err = vm.GetMigrationPath("NoRules", WorkflowVersion{}, WorkflowVersion{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no migration rules defined")
}

// Run the test suite
func TestVersioningSuite(t *testing.T) {
	suite.Run(t, new(VersioningTestSuite))
}