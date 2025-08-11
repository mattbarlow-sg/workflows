// Package schemas contains validation contracts for the Temporal validation framework
package schemas

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ValidationPipeline defines the main validation interface
type ValidationPipeline interface {
	// Validate performs complete validation of Temporal workflows
	Validate(ctx context.Context, request ValidationRequest) (*ValidationResult, error)
	
	// ValidateWithCache performs validation with cache support
	ValidateWithCache(ctx context.Context, request ValidationRequest) (*ValidationResult, error)
	
	// Individual check methods
	CheckDeterminism(ctx context.Context, workflowPath string) (*DeterminismCheckResult, error)
	CheckActivitySignatures(ctx context.Context, workflowPath string) (*ActivitySignatureResult, error)
	CheckPolicies(ctx context.Context, workflowPath string) (*PolicyValidationResult, error)
}

// CacheManager defines cache operations
type CacheManager interface {
	// Get retrieves a cached validation result
	Get(key string) (*CacheEntry, error)
	
	// Set stores a validation result in cache
	Set(entry *CacheEntry) error
	
	// Invalidate removes a cache entry
	Invalidate(key string) error
	
	// InvalidateByWorkflow removes all entries for a workflow
	InvalidateByWorkflow(workflowID string) error
}

// Validator provides validation helper methods
type Validator struct{}

// ValidateRequest validates a ValidationRequest
func (v *Validator) ValidateRequest(req *ValidationRequest) error {
	// Check workflow path exists and is readable
	if err := v.validateWorkflowPath(req.WorkflowPath); err != nil {
		return fmt.Errorf("invalid workflow path: %w", err)
	}
	
	// Validate workflow ID
	if req.WorkflowID == "" {
		return errors.New("workflow ID is required")
	}
	if len(req.WorkflowID) > 255 {
		return errors.New("workflow ID exceeds maximum length of 255")
	}
	
	// Validate options
	if err := v.validateOptions(req.Options); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	
	return nil
}

// validateWorkflowPath checks if path exists and is readable
func (v *Validator) validateWorkflowPath(path string) error {
	if path == "" {
		return errors.New("workflow path is required")
	}
	
	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return errors.New("workflow path must be absolute")
	}
	
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: %s", path)
		}
		return fmt.Errorf("cannot access path: %w", err)
	}
	
	// Check if readable (basic check)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read path: %w", err)
	}
	file.Close()
	
	// If directory, check for .go files
	if info.IsDir() {
		hasGoFiles := false
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(p, ".go") {
				hasGoFiles = true
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("cannot walk directory: %w", err)
		}
		if !hasGoFiles {
			return errors.New("no .go files found in directory")
		}
	}
	
	return nil
}

// validateOptions validates ValidationOptions
func (v *Validator) validateOptions(opts ValidationOptions) error {
	// Validate max info messages
	if opts.MaxInfoMessages < 0 {
		return errors.New("max info messages cannot be negative")
	}
	if opts.MaxInfoMessages > 1000 {
		return errors.New("max info messages exceeds limit of 1000")
	}
	
	// Validate timeout
	if opts.Timeout < 0 {
		return errors.New("timeout cannot be negative")
	}
	if opts.Timeout > 5*time.Minute {
		return errors.New("timeout exceeds maximum of 5 minutes")
	}
	
	return nil
}

// ValidateActivityName checks if activity name follows PascalCase
func ValidateActivityName(name string) error {
	if name == "" {
		return errors.New("activity name cannot be empty")
	}
	
	// PascalCase regex: starts with uppercase, followed by alphanumeric
	pascalCaseRegex := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	if !pascalCaseRegex.MatchString(name) {
		// Provide specific error message
		if name[0] >= 'a' && name[0] <= 'z' {
			return fmt.Errorf("activity name '%s' must start with uppercase letter", name)
		}
		if strings.Contains(name, "_") {
			return fmt.Errorf("activity name '%s' cannot contain underscores", name)
		}
		if strings.Contains(name, "-") {
			return fmt.Errorf("activity name '%s' cannot contain hyphens", name)
		}
		return fmt.Errorf("activity name '%s' must follow PascalCase convention", name)
	}
	
	return nil
}

// ValidateTimeoutPolicy validates timeout configuration
func ValidateTimeoutPolicy(name string, timeout time.Duration, isHumanTask bool) error {
	if isHumanTask {
		// Human tasks must have infinite timeout (0)
		if timeout != 0 {
			return fmt.Errorf("human task '%s' must have infinite timeout (0), got %s", name, timeout)
		}
	} else {
		// Non-human workflows limited to 15 minutes
		if timeout > 15*time.Minute {
			return fmt.Errorf("workflow '%s' timeout %s exceeds maximum of 15 minutes", name, timeout)
		}
		if timeout <= 0 {
			return fmt.Errorf("workflow '%s' must have positive timeout, got %s", name, timeout)
		}
	}
	
	return nil
}

// ValidateRetryPolicy validates retry configuration
func ValidateRetryPolicy(activityName string, retryCount int, errorType ErrorCategory) error {
	// Check if error type is retryable
	if !errorType.IsRetryable() {
		if retryCount > 0 {
			return fmt.Errorf("activity '%s': %s errors should not be retried", activityName, errorType)
		}
		return nil
	}
	
	// Minimum retry count for retryable errors
	if retryCount < 3 {
		return fmt.Errorf("activity '%s': retry count %d is below minimum of 3", activityName, retryCount)
	}
	
	return nil
}

// DeterminismPatterns defines patterns to check for non-deterministic code
var DeterminismPatterns = []struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
	Suggestion  string
}{
	{
		Name:        "time_now",
		Pattern:     regexp.MustCompile(`\btime\.Now\(\)`),
		Description: "time.Now() is non-deterministic",
		Suggestion:  "Use workflow.Now() instead",
	},
	{
		Name:        "goroutine",
		Pattern:     regexp.MustCompile(`\bgo\s+func\s*\(`),
		Description: "Native goroutines are non-deterministic",
		Suggestion:  "Use workflow.Go() instead",
	},
	{
		Name:        "channel",
		Pattern:     regexp.MustCompile(`\bchan\s+`),
		Description: "Native channels are non-deterministic",
		Suggestion:  "Use workflow.Channel instead",
	},
	{
		Name:        "select",
		Pattern:     regexp.MustCompile(`\bselect\s*{`),
		Description: "Native select is non-deterministic",
		Suggestion:  "Use workflow.Selector instead",
	},
	{
		Name:        "random",
		Pattern:     regexp.MustCompile(`\brand\.\w+\(`),
		Description: "Random number generation is non-deterministic",
		Suggestion:  "Use deterministic random with workflow.SideEffect",
	},
	{
		Name:        "map_range",
		Pattern:     regexp.MustCompile(`\bfor\s+\w+\s*,?\s*\w*\s*:=\s*range\s+\w+\s*{`),
		Description: "Map iteration order is non-deterministic",
		Suggestion:  "Sort map keys before iteration",
	},
}

// CheckDeterminismPatterns checks source code for non-deterministic patterns
func CheckDeterminismPatterns(source string) []DeterminismViolation {
	var violations []DeterminismViolation
	
	lines := strings.Split(source, "\n")
	for lineNum, line := range lines {
		for _, pattern := range DeterminismPatterns {
			if pattern.Pattern.MatchString(line) {
				violations = append(violations, DeterminismViolation{
					Pattern:     pattern.Name,
					Description: pattern.Description,
					Suggestion:  pattern.Suggestion,
					Location: CodeLocation{
						Line:    lineNum + 1,
						Snippet: strings.TrimSpace(line),
					},
				})
			}
		}
	}
	
	return violations
}

// ValidateActivitySignature checks if activity has valid signature
func ValidateActivitySignature(signature string) error {
	// Activity should return (result, error) tuple
	if !strings.Contains(signature, "error") {
		return errors.New("activity must return error as last value")
	}
	
	// Check for context.Context as first parameter
	if strings.Contains(signature, "(") && !strings.Contains(signature, "context.Context") {
		// It's OK if activity has no parameters
		if !strings.HasPrefix(signature, "()") {
			return errors.New("activity must have context.Context as first parameter when parameters present")
		}
	}
	
	return nil
}

// FastFailController manages fast-fail behavior
type FastFailController struct {
	cancelFunc context.CancelFunc
	cancelled  bool
}

// NewFastFailController creates a new fast-fail controller
func NewFastFailController(cancelFunc context.CancelFunc) *FastFailController {
	return &FastFailController{
		cancelFunc: cancelFunc,
		cancelled:  false,
	}
}

// TriggerFastFail cancels all parallel operations
func (f *FastFailController) TriggerFastFail() {
	if !f.cancelled && f.cancelFunc != nil {
		f.cancelFunc()
		f.cancelled = true
	}
}

// IsCancelled returns true if fast-fail was triggered
func (f *FastFailController) IsCancelled() bool {
	return f.cancelled
}

// ParallelCheckRunner runs validation checks in parallel with fast-fail
func ParallelCheckRunner(ctx context.Context, checks map[string]func(context.Context) error) map[string]error {
	results := make(map[string]error)
	resultChan := make(chan struct {
		name string
		err  error
	}, len(checks))
	
	// Create cancellable context for fast-fail
	checkCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Run checks in parallel
	for name, check := range checks {
		go func(n string, fn func(context.Context) error) {
			err := fn(checkCtx)
			resultChan <- struct {
				name string
				err  error
			}{n, err}
		}(name, check)
	}
	
	// Collect results with fast-fail
	for i := 0; i < len(checks); i++ {
		result := <-resultChan
		results[result.name] = result.err
		
		// Trigger fast-fail on first error
		if result.err != nil {
			cancel() // Cancel remaining checks
		}
	}
	
	return results
}

// TimeoutEnforcer ensures validation completes within time limit
func TimeoutEnforcer(timeout time.Duration, fn func() error) error {
	if timeout <= 0 {
		timeout = 5 * time.Minute // Default max timeout
	}
	
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("validation timeout exceeded: %s", timeout)
	}
}