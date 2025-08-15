// Package temporal provides validation rules and implementations for Temporal workflows
package temporal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// WorkflowGraphValidatorImpl validates workflow graph structure
type WorkflowGraphValidatorImpl struct {
	mu    sync.Mutex
	graph map[string][]string // workflow -> called workflows
}

// Validate checks the workflow graph for issues
func (w *WorkflowGraphValidatorImpl) Validate(ctx context.Context, workflowPath string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Build workflow graph
	if err := w.buildGraph(workflowPath); err != nil {
		return fmt.Errorf("failed to build workflow graph: %w", err)
	}

	// Check for cycles
	cycles, err := w.DetectCycles(workflowPath)
	if err != nil {
		return err
	}

	if len(cycles) > 0 {
		return fmt.Errorf("detected workflow cycles: %v", cycles)
	}

	// Check connectivity
	connected, err := w.CheckConnectivity(workflowPath)
	if err != nil {
		return err
	}

	if !connected {
		return fmt.Errorf("workflow graph has disconnected components")
	}

	return nil
}

// DetectCycles finds cycles in the workflow graph
func (w *WorkflowGraphValidatorImpl) DetectCycles(workflowPath string) ([]string, error) {
	if w.graph == nil {
		if err := w.buildGraph(workflowPath); err != nil {
			return nil, err
		}
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycles []string

	for node := range w.graph {
		if !visited[node] {
			if w.detectCycleDFS(node, visited, recStack, &cycles) {
				// Found cycle
			}
		}
	}

	return cycles, nil
}

// detectCycleDFS performs depth-first search to detect cycles
func (w *WorkflowGraphValidatorImpl) detectCycleDFS(node string, visited, recStack map[string]bool, cycles *[]string) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range w.graph[node] {
		if !visited[neighbor] {
			if w.detectCycleDFS(neighbor, visited, recStack, cycles) {
				*cycles = append(*cycles, fmt.Sprintf("%s -> %s", node, neighbor))
				return true
			}
		} else if recStack[neighbor] {
			// Found a cycle
			*cycles = append(*cycles, fmt.Sprintf("%s -> %s (cycle)", node, neighbor))
			return true
		}
	}

	recStack[node] = false
	return false
}

// CheckConnectivity verifies all workflows are reachable
func (w *WorkflowGraphValidatorImpl) CheckConnectivity(workflowPath string) (bool, error) {
	if w.graph == nil {
		if err := w.buildGraph(workflowPath); err != nil {
			return false, err
		}
	}

	if len(w.graph) == 0 {
		return true, nil // Empty graph is connected
	}

	// Collect all nodes (both keys and values in the graph)
	allNodes := make(map[string]bool)
	for node, neighbors := range w.graph {
		allNodes[node] = true
		for _, neighbor := range neighbors {
			allNodes[neighbor] = true
		}
	}

	// Find all components using DFS
	visited := make(map[string]bool)
	components := 0

	for node := range allNodes {
		if !visited[node] {
			components++
			w.dfsVisit(node, visited)
		}
	}

	// Graph is connected if there's only one component
	return components == 1, nil
}

// dfsVisit performs depth-first traversal
func (w *WorkflowGraphValidatorImpl) dfsVisit(node string, visited map[string]bool) {
	visited[node] = true
	for _, neighbor := range w.graph[node] {
		if !visited[neighbor] {
			w.dfsVisit(neighbor, visited)
		}
	}
}

// buildGraph constructs the workflow dependency graph
func (w *WorkflowGraphValidatorImpl) buildGraph(workflowPath string) error {
	w.graph = make(map[string][]string)

	files, err := findWorkflowFiles(workflowPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := w.analyzeFile(file); err != nil {
			continue // Skip files with errors
		}
	}

	return nil
}

// analyzeFile extracts workflow dependencies from a file
func (w *WorkflowGraphValidatorImpl) analyzeFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Simple pattern matching for workflow calls
	// Look for ExecuteChildWorkflow, ExecuteWorkflow, etc.
	patterns := []string{
		`ExecuteChildWorkflow\s*\(\s*[^,]+,\s*"([^"]+)"`,
		`ExecuteWorkflow\s*\(\s*[^,]+,\s*"([^"]+)"`,
		`StartChildWorkflow\s*\(\s*[^,]+,\s*"([^"]+)"`,
	}

	contentStr := string(content)
	currentWorkflow := w.extractWorkflowName(contentStr)

	if currentWorkflow != "" {
		if _, exists := w.graph[currentWorkflow]; !exists {
			w.graph[currentWorkflow] = []string{}
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(contentStr, -1)
			for _, match := range matches {
				if len(match) > 1 {
					childWorkflow := match[1]
					w.graph[currentWorkflow] = append(w.graph[currentWorkflow], childWorkflow)
				}
			}
		}
	}

	return nil
}

// extractWorkflowName finds the workflow name from source
func (w *WorkflowGraphValidatorImpl) extractWorkflowName(content string) string {
	// Look for workflow function declarations
	re := regexp.MustCompile(`func\s+(\w+)\s*\(\s*[^)]*workflow\.Context`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// PolicyValidatorImpl validates timeout and retry policies
type PolicyValidatorImpl struct{}

// Validate checks timeout and retry policies
func (p *PolicyValidatorImpl) Validate(ctx context.Context, workflowPath string) (*schemas.PolicyValidationResult, error) {
	startTime := time.Now()
	timeoutViolations := []schemas.TimeoutViolation{}
	retryViolations := []schemas.RetryViolation{}

	files, err := findWorkflowFiles(workflowPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileTimeouts, fileRetries := p.validateFile(file)
		timeoutViolations = append(timeoutViolations, fileTimeouts...)
		retryViolations = append(retryViolations, fileRetries...)
	}

	return &schemas.PolicyValidationResult{
		Passed:            len(timeoutViolations) == 0 && len(retryViolations) == 0,
		TimeoutViolations: timeoutViolations,
		RetryViolations:   retryViolations,
		Duration:          time.Since(startTime),
	}, nil
}

// validateFile checks policies in a single file
func (p *PolicyValidatorImpl) validateFile(filePath string) ([]schemas.TimeoutViolation, []schemas.RetryViolation) {
	timeoutViolations := []schemas.TimeoutViolation{}
	retryViolations := []schemas.RetryViolation{}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return timeoutViolations, retryViolations
	}

	relPath := filePath
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, filePath); err == nil {
			relPath = rel
		}
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Check for timeout configurations
	timeoutPatterns := []struct {
		pattern *regexp.Regexp
		extract func([]string) (string, int)
	}{
		{
			pattern: regexp.MustCompile(`WorkflowExecutionTimeout:\s*(\d+)\s*\*\s*time\.(\w+)`),
			extract: func(matches []string) (string, int) {
				if len(matches) > 2 {
					// Simple conversion (this would need proper parsing in production)
					return "workflow", 15 // Default check against 15 minutes
				}
				return "", 0
			},
		},
		{
			pattern: regexp.MustCompile(`ActivityOptions.*StartToCloseTimeout:\s*(\d+)\s*\*\s*time\.(\w+)`),
			extract: func(matches []string) (string, int) {
				if len(matches) > 2 {
					return "activity", 5 // Default check against 5 minutes
				}
				return "", 0
			},
		},
	}

	for lineNum, line := range lines {
		// Check timeout patterns
		for _, tp := range timeoutPatterns {
			if matches := tp.pattern.FindStringSubmatch(line); len(matches) > 0 {
				name, timeout := tp.extract(matches)
				if name != "" && timeout > 15 {
					timeoutViolations = append(timeoutViolations, schemas.TimeoutViolation{
						Name:              name,
						IsHumanTask:       p.isHumanTask(contentStr, name),
						ConfiguredTimeout: timeout,
						RequiredTimeout:   "15 minutes max",
						Location: schemas.CodeLocation{
							File:    relPath,
							Line:    lineNum + 1,
							Snippet: strings.TrimSpace(line),
						},
					})
				}
			}
		}

		// Check retry patterns
		retryPattern := regexp.MustCompile(`MaximumAttempts:\s*(\d+)`)
		if matches := retryPattern.FindStringSubmatch(line); len(matches) > 1 {
			// Extract retry count
			retryCount := 0
			fmt.Sscanf(matches[1], "%d", &retryCount)

			if retryCount < 3 {
				activityName := p.findNearestActivity(lines, lineNum)
				retryViolations = append(retryViolations, schemas.RetryViolation{
					ActivityName:      activityName,
					ConfiguredRetries: retryCount,
					MinimumRequired:   3,
					Location: schemas.CodeLocation{
						File:    relPath,
						Line:    lineNum + 1,
						Snippet: strings.TrimSpace(line),
					},
				})
			}
		}
	}

	return timeoutViolations, retryViolations
}

// isHumanTask checks if a workflow/activity is a human task
func (p *PolicyValidatorImpl) isHumanTask(content, name string) bool {
	humanTaskPatterns := []string{
		"HumanTask",
		"Approval",
		"Review",
		"UserInput",
		"ManualStep",
		"WaitForSignal",
	}

	for _, pattern := range humanTaskPatterns {
		if strings.Contains(name, pattern) || strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// findNearestActivity finds the closest activity name to a line
func (p *PolicyValidatorImpl) findNearestActivity(lines []string, lineNum int) string {
	// Search backwards for activity definition
	for i := lineNum; i >= 0 && i > lineNum-20; i-- {
		if strings.Contains(lines[i], "ExecuteActivity") {
			// Extract activity name
			re := regexp.MustCompile(`ExecuteActivity\s*\([^,]+,\s*"?(\w+)"?`)
			if matches := re.FindStringSubmatch(lines[i]); len(matches) > 1 {
				return matches[1]
			}
		}
	}
	return "unknown_activity"
}

// HumanTaskValidatorImpl validates human task configurations
type HumanTaskValidatorImpl struct{}

// Validate checks human task configurations
func (h *HumanTaskValidatorImpl) Validate(ctx context.Context, workflowPath string) error {
	// Check if it's a single file or directory
	info, err := os.Stat(workflowPath)
	if err != nil {
		return err
	}

	var files []string
	if !info.IsDir() {
		// Single file - validate directly
		files = []string{workflowPath}
	} else {
		// Directory - find workflow files
		files, err = findWorkflowFiles(workflowPath)
		if err != nil {
			return err
		}
	}

	for _, file := range files {
		if err := h.validateFile(file); err != nil {
			return err
		}
	}

	return nil
}

// validateFile checks human tasks in a single file
func (h *HumanTaskValidatorImpl) validateFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Check for human task patterns
	if h.hasHumanTasks(contentStr) {
		// Validate escalation rules
		if !h.hasEscalation(contentStr) {
			return fmt.Errorf("human tasks found in %s but no escalation policy defined", filePath)
		}

		// Validate assignment rules
		if !h.hasAssignment(contentStr) {
			return fmt.Errorf("human tasks found in %s but no assignment rules defined", filePath)
		}

		// Validate timeout handling
		if !h.hasTimeoutHandling(contentStr) {
			return fmt.Errorf("human tasks found in %s but no timeout handling defined", filePath)
		}
	}

	return nil
}

// hasHumanTasks checks if file contains human tasks
func (h *HumanTaskValidatorImpl) hasHumanTasks(content string) bool {
	patterns := []string{
		"HumanTask",
		"WaitForSignal",
		"UserApproval",
		"ManualTask",
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// hasEscalation checks for escalation configuration
func (h *HumanTaskValidatorImpl) hasEscalation(content string) bool {
	patterns := []string{
		"Escalation",
		"EscalateAfter",
		"EscalationPolicy",
		"workflow.NewTimer", // Timer-based escalation
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// hasAssignment checks for task assignment rules
func (h *HumanTaskValidatorImpl) hasAssignment(content string) bool {
	patterns := []string{
		"AssignTo",
		"TaskAssignment",
		"Assignee",
		"TaskOwner",
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// hasTimeoutHandling checks for timeout handling
func (h *HumanTaskValidatorImpl) hasTimeoutHandling(content string) bool {
	patterns := []string{
		"workflow.AwaitWithTimeout",
		"workflow.NewTimer",
		"TimeoutHandler",
		"OnTimeout",
	}

	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// ValidateEscalation validates escalation configuration
func (h *HumanTaskValidatorImpl) ValidateEscalation(taskConfig map[string]interface{}) error {
	// Check for required escalation fields
	requiredFields := []string{"escalation_after", "escalate_to", "max_escalations"}

	for _, field := range requiredFields {
		if _, exists := taskConfig[field]; !exists {
			return fmt.Errorf("missing required escalation field: %s", field)
		}
	}

	// Validate escalation time
	if escalateAfter, ok := taskConfig["escalation_after"].(time.Duration); ok {
		if escalateAfter < 30*time.Minute {
			return fmt.Errorf("escalation time too short: minimum 30 minutes")
		}
		if escalateAfter > 7*24*time.Hour {
			return fmt.Errorf("escalation time too long: maximum 7 days")
		}
	}

	// Validate max escalations
	if maxEscalations, ok := taskConfig["max_escalations"].(int); ok {
		if maxEscalations < 1 || maxEscalations > 5 {
			return fmt.Errorf("max_escalations must be between 1 and 5")
		}
	}

	return nil
}

// ValidateAssignment validates task assignment configuration
func (h *HumanTaskValidatorImpl) ValidateAssignment(taskConfig map[string]interface{}) error {
	// Check for assignment strategy
	if _, exists := taskConfig["assignment_strategy"]; !exists {
		return fmt.Errorf("missing assignment_strategy")
	}

	strategy, ok := taskConfig["assignment_strategy"].(string)
	if !ok {
		return fmt.Errorf("invalid assignment_strategy type")
	}

	// Validate strategy
	validStrategies := []string{"round_robin", "least_loaded", "priority", "manual"}
	valid := false
	for _, s := range validStrategies {
		if s == strategy {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid assignment_strategy: %s", strategy)
	}

	// Check for assignee pool
	if _, exists := taskConfig["assignee_pool"]; !exists {
		return fmt.Errorf("missing assignee_pool")
	}

	return nil
}

// MemoryCache implements a simple in-memory cache
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*schemas.CacheEntry
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*schemas.CacheEntry),
	}
}

// Get retrieves a cache entry
func (c *MemoryCache) Get(key string) (*schemas.CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, fmt.Errorf("cache miss: %s", key)
	}

	// Update access time
	entry.AccessedAt = time.Now()
	entry.AccessCount++

	return entry, nil
}

// Set stores a cache entry
func (c *MemoryCache) Set(entry *schemas.CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[entry.Key] = entry
	return nil
}

// Invalidate removes a cache entry
func (c *MemoryCache) Invalidate(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

// InvalidateByWorkflow removes all entries for a workflow
func (c *MemoryCache) InvalidateByWorkflow(workflowID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.entries {
		if strings.HasPrefix(key, workflowID+":") {
			delete(c.entries, key)
		}
	}

	return nil
}
