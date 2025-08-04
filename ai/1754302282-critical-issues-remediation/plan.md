# Critical Issues Remediation Plan

## Overview
This plan addresses the critical and high-impact issues identified in the codebase risk analysis. The work is organized into phases with clear priorities and parallel execution opportunities.

## Phase 1: Critical Security and Stability (Week 1)

### Task 1.1: Replace os.Exit() Pattern (Critical)
**Assignee**: AI
**Priority**: Critical
**Estimated Time**: 2 days
**Dependencies**: None
**Parallel Work**: Can be done alongside 1.2 and 1.3

**Steps**:
1. Create error types and interfaces in `internal/errors/errors.go`
2. Refactor command functions to return errors instead of calling `os.Exit()`
3. Update `main()` to handle errors and control exit codes
4. Create unit tests for command functions

**Files to modify**:
- `cmd/workflows/main.go` (8 instances)
- `cmd/workflows/adr.go` (10 instances)
- `cmd/workflows/bpmn.go` (18 instances)

### Task 1.2: Add Input Validation (Security Critical)
**Assignee**: AI
**Priority**: Critical
**Estimated Time**: 1 day
**Dependencies**: None
**Parallel Work**: Can be done alongside 1.1 and 1.3

**Steps**:
1. Create validation package `internal/validation/paths.go`
2. Implement path traversal detection
3. Add file extension validation
4. Apply validation to all file input points

**Functions to update**:
- `validateCommand()`: Validate schema name and file path
- `adrNewCommand()`: Validate output path
- `adrRenderCommand()`: Validate input and output paths
- `adrValidateCommand()`: Validate file path
- `bpmnNewCommand()`: Validate process name and output path
- `bpmnValidateCommand()`: Validate file path
- `bpmnAnalyzeCommand()`: Validate file path

### Task 1.3: Fix Hardcoded Schema Paths
**Assignee**: AI
**Priority**: High
**Estimated Time**: 1 day
**Dependencies**: None
**Parallel Work**: Can be done alongside 1.1 and 1.2

**Steps**:
1. Add schema path resolution logic
2. Support environment variable `WORKFLOWS_SCHEMA_DIR`
3. Implement schema path discovery (check multiple locations)
4. Update all hardcoded references

**Locations to fix**:
- `cmd/workflows/main.go:62`: `filepath.Join(".", "schemas")`
- `cmd/workflows/main.go:107`: `filepath.Join(".", "schemas")`
- `cmd/workflows/adr.go:452`: `filepath.Join(".", "schemas", "adr.json")`

## Phase 2: Technical Debt Cleanup (Week 1-2)

### Task 2.1: Replace Deprecated ioutil Package
**Assignee**: AI
**Priority**: Medium
**Estimated Time**: 2 hours
**Dependencies**: Can start after Phase 1 tasks
**Parallel Work**: Can be done alongside 2.2

**Steps**:
1. Replace `ioutil.ReadFile` with `os.ReadFile`
2. Replace `ioutil.WriteFile` with `os.WriteFile`
3. Update imports
4. Run tests to ensure compatibility

**Files to update**:
- `cmd/workflows/adr.go`: Lines 356, 398, 413, 463
- `cmd/workflows/bpmn.go`: Lines 104, 128

### Task 2.2: Break Down Large Functions
**Assignee**: Human
**Priority**: High
**Estimated Time**: 3 days
**Dependencies**: Task 1.1 (os.Exit removal)
**Parallel Work**: Can be done alongside 2.1

**Steps**:
1. Review `adrNewCommand()` function (300+ lines)
2. Extract sub-functions for:
   - Flag parsing and validation
   - Interactive mode handling
   - ADR generation
   - Output formatting
3. Create unit tests for each extracted function
4. Apply similar refactoring to other large functions

**Target functions**:
- `adrNewCommand()`: 300+ lines ’ 5-6 smaller functions
- `bpmnAnalyzeCommand()`: Extract analysis and output logic
- Command handler functions: Separate concerns

## Phase 3: Performance and Architecture (Week 2-3)

### Task 3.1: Optimize Graph Analysis Algorithms
**Assignee**: AI
**Priority**: Medium
**Estimated Time**: 2 days
**Dependencies**: None
**Parallel Work**: Can be done alongside 3.2 and 3.3

**Steps**:
1. Analyze current algorithms in `internal/bpmn/analyzer.go`
2. Implement single-pass algorithms where possible
3. Add memory pre-allocation with size hints
4. Implement caching for repeated analyses
5. Add benchmarks to measure improvements

**Specific optimizations**:
- Combine multiple DFS traversals into single pass
- Pre-allocate maps and slices with estimated sizes
- Cache graph representations

### Task 3.2: Add Timeout Protection
**Assignee**: AI
**Priority**: Medium
**Estimated Time**: 1 day
**Dependencies**: None
**Parallel Work**: Can be done alongside 3.1 and 3.3

**Steps**:
1. Add context support to analyzer functions
2. Implement timeout configuration
3. Add graceful cancellation handling
4. Update command handlers to pass context

**Functions to update**:
- `Analyzer.Analyze()`: Accept context parameter
- Graph traversal functions: Check context cancellation
- Command handlers: Create context with timeout

### Task 3.3: Implement Configuration Management
**Assignee**: Human
**Priority**: Medium
**Estimated Time**: 3 days
**Dependencies**: Task 1.3 (schema path fixes)
**Parallel Work**: Can be done alongside 3.1 and 3.2

**Steps**:
1. Design configuration structure
2. Create `internal/config/config.go`
3. Support multiple configuration sources:
   - Environment variables
   - Configuration file
   - Command-line flags
4. Replace hardcoded values with configuration
5. Document configuration options

**Configuration items**:
- Schema directory path
- Default output formats
- Analysis timeouts
- Resource limits

## Phase 4: Testing and Documentation (Week 3)

### Task 4.1: Comprehensive Test Suite
**Assignee**: AI
**Priority**: High
**Estimated Time**: 3 days
**Dependencies**: Phase 1 and 2 completion

**Steps**:
1. Create unit tests for all command functions
2. Add integration tests for CLI commands
3. Implement error case testing
4. Add security test cases (path traversal, etc.)
5. Achieve >80% code coverage

### Task 4.2: Performance Benchmarks
**Assignee**: Human
**Priority**: Medium
**Estimated Time**: 1 day
**Dependencies**: Task 3.1 completion

**Steps**:
1. Create benchmark tests for graph analysis
2. Add memory usage benchmarks
3. Document performance characteristics
4. Set up continuous performance monitoring

### Task 4.3: Update Documentation
**Assignee**: Human
**Priority**: Medium
**Estimated Time**: 1 day
**Dependencies**: All implementation tasks

**Steps**:
1. Update README with configuration options
2. Document security considerations
3. Add deployment guide
4. Create troubleshooting guide

## Success Criteria

### Phase 1 Complete When:
- [ ] All `os.Exit()` calls removed
- [ ] Input validation implemented
- [ ] Schema paths configurable
- [ ] Basic unit tests pass

### Phase 2 Complete When:
- [ ] No deprecated API usage
- [ ] Large functions refactored
- [ ] Code maintainability improved

### Phase 3 Complete When:
- [ ] Graph analysis optimized
- [ ] Timeout protection added
- [ ] Configuration system implemented

### Phase 4 Complete When:
- [ ] >80% test coverage achieved
- [ ] Performance benchmarks established
- [ ] Documentation updated

## Risk Mitigation

### Rollback Plan
- Keep original code in feature branches
- Implement changes incrementally
- Test each phase thoroughly before proceeding

### Compatibility Concerns
- Maintain CLI interface compatibility
- Version configuration format
- Document breaking changes

## Timeline Summary

**Week 1**: 
- Phase 1: Critical security and stability fixes
- Start Phase 2: Technical debt cleanup

**Week 2**:
- Complete Phase 2
- Phase 3: Performance and architecture improvements

**Week 3**:
- Complete Phase 3
- Phase 4: Testing and documentation

**Total Duration**: 3 weeks with parallel execution

## Next Steps

1. Human reviews and approves plan
2. Create feature branches for each phase
3. Begin Phase 1 implementation
4. Set up daily progress reviews