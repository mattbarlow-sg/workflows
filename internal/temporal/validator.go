// Package temporal provides validation framework for Temporal workflows
package temporal

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// ValidatorInterface defines the main validation contract
type ValidatorInterface interface {
	// Validate performs comprehensive validation of Temporal workflows
	Validate(ctx context.Context, request schemas.ValidationRequest) (*schemas.ValidationResult, error)

	// ValidateWithCache performs validation with cache support
	ValidateWithCache(ctx context.Context, request schemas.ValidationRequest) (*schemas.ValidationResult, error)
}

// TemporalValidator implements the ValidatorInterface
type TemporalValidator struct {
	cache              CacheManager
	determinismChecker DeterminismChecker
	signatureValidator ActivitySignatureValidator
	policyValidator    PolicyValidator
	graphValidator     WorkflowGraphValidator
	humanTaskValidator HumanTaskValidator
}

// NewTemporalValidator creates a new validator instance
func NewTemporalValidator() *TemporalValidator {
	return &TemporalValidator{
		cache:              NewMemoryCache(),
		determinismChecker: &DeterminismCheckerImpl{},
		signatureValidator: &ActivitySignatureValidatorImpl{},
		policyValidator:    &PolicyValidatorImpl{},
		graphValidator:     &WorkflowGraphValidatorImpl{},
		humanTaskValidator: &HumanTaskValidatorImpl{},
	}
}

// Validate performs comprehensive validation without cache
func (v *TemporalValidator) Validate(ctx context.Context, request schemas.ValidationRequest) (*schemas.ValidationResult, error) {
	return v.performValidation(ctx, request, false)
}

// ValidateWithCache performs validation with cache support
func (v *TemporalValidator) ValidateWithCache(ctx context.Context, request schemas.ValidationRequest) (*schemas.ValidationResult, error) {
	return v.performValidation(ctx, request, true)
}

// performValidation executes the validation pipeline
func (v *TemporalValidator) performValidation(ctx context.Context, request schemas.ValidationRequest, useCache bool) (*schemas.ValidationResult, error) {
	startTime := time.Now()

	// Validate request
	validator := &schemas.Validator{}
	if err := validator.ValidateRequest(&request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Compute source hash for caching
	sourceContent, err := v.readWorkflowSource(request.WorkflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow source: %w", err)
	}
	sourceHash := schemas.ComputeSourceHash(sourceContent)

	// Check cache if enabled
	if useCache && !request.Options.SkipCache {
		cacheKey := schemas.CacheKey(request.WorkflowID, sourceHash)
		if entry, err := v.cache.Get(cacheKey); err == nil && entry != nil {
			// Update cache hit info and return
			result := entry.Result
			result.CacheHit = true
			result.Duration = time.Since(startTime)
			return &result, nil
		}
	}

	// Create validation context
	validationCtx := &ValidationContext{
		State:        schemas.ValidationStateValidating,
		Request:      request,
		StartTime:    startTime,
		Errors:       []schemas.ValidationError{},
		InfoMessages: []schemas.InfoMessage{},
	}

	// Apply timeout if specified
	if request.Options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Options.Timeout)
		defer cancel()
		validationCtx.CancelFunc = cancel
	}

	// Run validation checks
	checkStatus := schemas.CheckStatusMap{
		DeterminismCheck: schemas.CheckStatusPending,
		SignatureCheck:   schemas.CheckStatusPending,
		PolicyCheck:      schemas.CheckStatusPending,
	}

	if request.Options.ParallelChecks {
		// Run checks in parallel
		v.runParallelChecks(ctx, validationCtx, &checkStatus)
	} else {
		// Run checks sequentially
		v.runSequentialChecks(ctx, validationCtx, &checkStatus)
	}

	// Collect results
	validationCtx.State = schemas.ValidationStateCollectingErrors
	errors := v.collectErrors(validationCtx)
	infoMessages := v.collectInfoMessages(validationCtx, request.Options.MaxInfoMessages)

	// Build final result
	result := &schemas.ValidationResult{
		Success:      len(errors) == 0,
		Errors:       errors,
		InfoMessages: infoMessages,
		CheckStatus:  checkStatus,
		Duration:     time.Since(startTime),
		CacheHit:     false,
		Timestamp:    time.Now(),
		SourceHash:   sourceHash,
	}

	// Cache result if successful and caching enabled
	if useCache && result.Success {
		cacheEntry := &schemas.CacheEntry{
			Key:         schemas.CacheKey(request.WorkflowID, sourceHash),
			WorkflowID:  request.WorkflowID,
			SourceHash:  sourceHash,
			Result:      *result,
			CreatedAt:   time.Now(),
			AccessedAt:  time.Now(),
			AccessCount: 0,
		}
		_ = v.cache.Set(cacheEntry) // Ignore cache errors
	}

	validationCtx.State = schemas.ValidationStateCompleted
	return result, nil
}

// runParallelChecks executes validation checks concurrently
func (v *TemporalValidator) runParallelChecks(ctx context.Context, valCtx *ValidationContext, status *schemas.CheckStatusMap) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Determinism check
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		status.DeterminismCheck = schemas.CheckStatusRunning
		mu.Unlock()

		result, err := v.determinismChecker.Check(ctx, valCtx.Request.WorkflowPath)
		mu.Lock()
		if err != nil {
			status.DeterminismCheck = schemas.CheckStatusFailed
			valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryDeterminism,
				Code:     "DETERMINISM_CHECK_FAILED",
				Message:  fmt.Sprintf("Determinism check failed: %v", err),
			})
		} else {
			valCtx.DeterminismResult = result
			status.DeterminismCheck = schemas.CheckStatusCompleted
		}
		mu.Unlock()
	}()

	// Activity signature check
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		status.SignatureCheck = schemas.CheckStatusRunning
		mu.Unlock()

		result, err := v.signatureValidator.Validate(ctx, valCtx.Request.WorkflowPath)
		mu.Lock()
		if err != nil {
			status.SignatureCheck = schemas.CheckStatusFailed
			valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryTypeMismatch,
				Code:     "SIGNATURE_CHECK_FAILED",
				Message:  fmt.Sprintf("Signature validation failed: %v", err),
			})
		} else {
			valCtx.SignatureResult = result
			status.SignatureCheck = schemas.CheckStatusCompleted
		}
		mu.Unlock()
	}()

	// Policy check
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		status.PolicyCheck = schemas.CheckStatusRunning
		mu.Unlock()

		result, err := v.policyValidator.Validate(ctx, valCtx.Request.WorkflowPath)
		mu.Lock()
		if err != nil {
			status.PolicyCheck = schemas.CheckStatusFailed
			valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryInvalidConfig,
				Code:     "POLICY_CHECK_FAILED",
				Message:  fmt.Sprintf("Policy validation failed: %v", err),
			})
		} else {
			valCtx.PolicyResult = result
			status.PolicyCheck = schemas.CheckStatusCompleted
		}
		mu.Unlock()
	}()

	wg.Wait()
}

// runSequentialChecks executes validation checks one by one
func (v *TemporalValidator) runSequentialChecks(ctx context.Context, valCtx *ValidationContext, status *schemas.CheckStatusMap) {
	// Determinism check
	status.DeterminismCheck = schemas.CheckStatusRunning
	result, err := v.determinismChecker.Check(ctx, valCtx.Request.WorkflowPath)
	if err != nil {
		status.DeterminismCheck = schemas.CheckStatusFailed
		valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
			Severity: schemas.ErrorSeverityError,
			Category: schemas.ErrorCategoryDeterminism,
			Code:     "DETERMINISM_CHECK_FAILED",
			Message:  fmt.Sprintf("Determinism check failed: %v", err),
		})
	} else {
		valCtx.DeterminismResult = result
		status.DeterminismCheck = schemas.CheckStatusCompleted
	}

	// Activity signature check
	status.SignatureCheck = schemas.CheckStatusRunning
	sigResult, err := v.signatureValidator.Validate(ctx, valCtx.Request.WorkflowPath)
	if err != nil {
		status.SignatureCheck = schemas.CheckStatusFailed
		valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
			Severity: schemas.ErrorSeverityError,
			Category: schemas.ErrorCategoryTypeMismatch,
			Code:     "SIGNATURE_CHECK_FAILED",
			Message:  fmt.Sprintf("Signature validation failed: %v", err),
		})
	} else {
		valCtx.SignatureResult = sigResult
		status.SignatureCheck = schemas.CheckStatusCompleted
	}

	// Policy check
	status.PolicyCheck = schemas.CheckStatusRunning
	policyResult, err := v.policyValidator.Validate(ctx, valCtx.Request.WorkflowPath)
	if err != nil {
		status.PolicyCheck = schemas.CheckStatusFailed
		valCtx.Errors = append(valCtx.Errors, schemas.ValidationError{
			Severity: schemas.ErrorSeverityError,
			Category: schemas.ErrorCategoryInvalidConfig,
			Code:     "POLICY_CHECK_FAILED",
			Message:  fmt.Sprintf("Policy validation failed: %v", err),
		})
	} else {
		valCtx.PolicyResult = policyResult
		status.PolicyCheck = schemas.CheckStatusCompleted
	}
}

// collectErrors aggregates all validation errors
func (v *TemporalValidator) collectErrors(valCtx *ValidationContext) []schemas.ValidationError {
	errors := append([]schemas.ValidationError{}, valCtx.Errors...)

	// Collect determinism violations
	if valCtx.DeterminismResult != nil && !valCtx.DeterminismResult.Passed {
		for _, violation := range valCtx.DeterminismResult.Violations {
			errors = append(errors, schemas.ValidationError{
				Severity:        schemas.ErrorSeverityError,
				Category:        schemas.ErrorCategoryDeterminism,
				Code:            fmt.Sprintf("DETERMINISM_%s", strings.ToUpper(violation.Pattern)),
				Message:         violation.Description,
				Location:        violation.Location,
				FixInstructions: violation.Suggestion,
			})
		}
	}

	// Collect signature violations
	if valCtx.SignatureResult != nil && !valCtx.SignatureResult.Passed {
		for _, violation := range valCtx.SignatureResult.Violations {
			errors = append(errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryTypeMismatch,
				Code:     fmt.Sprintf("SIGNATURE_%s", strings.ToUpper(violation.ViolationType)),
				Message:  fmt.Sprintf("Activity '%s': %s", violation.ActivityName, violation.ViolationType),
				Location: violation.Location,
				Context: map[string]interface{}{
					"expected": violation.Expected,
					"actual":   violation.Actual,
				},
			})
		}
	}

	// Collect policy violations
	if valCtx.PolicyResult != nil && !valCtx.PolicyResult.Passed {
		for _, violation := range valCtx.PolicyResult.TimeoutViolations {
			errors = append(errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryTimeout,
				Code:     "TIMEOUT_POLICY_VIOLATION",
				Message:  fmt.Sprintf("Workflow '%s': timeout policy violation", violation.Name),
				Location: violation.Location,
				Context: map[string]interface{}{
					"configured":    violation.ConfiguredTimeout,
					"required":      violation.RequiredTimeout,
					"is_human_task": violation.IsHumanTask,
				},
			})
		}

		for _, violation := range valCtx.PolicyResult.RetryViolations {
			errors = append(errors, schemas.ValidationError{
				Severity: schemas.ErrorSeverityError,
				Category: schemas.ErrorCategoryRetry,
				Code:     "RETRY_POLICY_VIOLATION",
				Message:  fmt.Sprintf("Activity '%s': retry policy violation", violation.ActivityName),
				Location: violation.Location,
				Context: map[string]interface{}{
					"configured": violation.ConfiguredRetries,
					"minimum":    violation.MinimumRequired,
					"error_type": violation.ErrorType,
				},
			})
		}
	}

	return errors
}

// collectInfoMessages aggregates informational messages
func (v *TemporalValidator) collectInfoMessages(valCtx *ValidationContext, maxMessages int) []schemas.InfoMessage {
	messages := append([]schemas.InfoMessage{}, valCtx.InfoMessages...)

	// Add performance metrics
	if valCtx.DeterminismResult != nil {
		messages = append(messages, schemas.InfoMessage{
			Category:  schemas.InfoCategoryPerformance,
			Message:   fmt.Sprintf("Determinism check completed in %s", valCtx.DeterminismResult.Duration),
			Timestamp: time.Now(),
		})
	}

	if valCtx.SignatureResult != nil {
		messages = append(messages, schemas.InfoMessage{
			Category:  schemas.InfoCategoryMetrics,
			Message:   fmt.Sprintf("Validated %d activities", len(valCtx.SignatureResult.Activities)),
			Timestamp: time.Now(),
		})
	}

	// Limit messages if specified
	if maxMessages > 0 && len(messages) > maxMessages {
		messages = messages[:maxMessages]
	}

	return messages
}

// readWorkflowSource reads the workflow source code
func (v *TemporalValidator) readWorkflowSource(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ValidationContext maintains validation state
type ValidationContext struct {
	State      schemas.ValidationState
	Request    schemas.ValidationRequest
	StartTime  time.Time
	CancelFunc context.CancelFunc

	// Results from each check
	DeterminismResult *schemas.DeterminismCheckResult
	SignatureResult   *schemas.ActivitySignatureResult
	PolicyResult      *schemas.PolicyValidationResult

	// Collected errors and messages
	Errors       []schemas.ValidationError
	InfoMessages []schemas.InfoMessage

	mu sync.Mutex
}

// DeterminismChecker interface for determinism validation
type DeterminismChecker interface {
	Check(ctx context.Context, workflowPath string) (*schemas.DeterminismCheckResult, error)
}

// ActivitySignatureValidator interface for activity signature validation
type ActivitySignatureValidator interface {
	Validate(ctx context.Context, workflowPath string) (*schemas.ActivitySignatureResult, error)
}

// PolicyValidator interface for policy validation
type PolicyValidator interface {
	Validate(ctx context.Context, workflowPath string) (*schemas.PolicyValidationResult, error)
}

// WorkflowGraphValidator interface for graph validation
type WorkflowGraphValidator interface {
	Validate(ctx context.Context, workflowPath string) error
	DetectCycles(workflowPath string) ([]string, error)
	CheckConnectivity(workflowPath string) (bool, error)
}

// HumanTaskValidator interface for human task validation
type HumanTaskValidator interface {
	Validate(ctx context.Context, workflowPath string) error
	ValidateEscalation(taskConfig map[string]interface{}) error
	ValidateAssignment(taskConfig map[string]interface{}) error
}

// CacheManager interface for cache operations
type CacheManager interface {
	Get(key string) (*schemas.CacheEntry, error)
	Set(entry *schemas.CacheEntry) error
	Invalidate(key string) error
	InvalidateByWorkflow(workflowID string) error
}

// parseGoFile parses a Go source file and returns the AST
func parseGoFile(path string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file %s: %w", path, err)
	}
	return file, fset, nil
}

// findWorkflowFiles finds all Go files containing Temporal workflows
func findWorkflowFiles(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and test directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		// Look for Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			// Check if file contains workflow definitions
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			// Simple heuristic: look for workflow-related imports or functions
			contentStr := string(content)
			if strings.Contains(contentStr, "go.temporal.io/sdk/workflow") ||
				strings.Contains(contentStr, "workflow.Context") ||
				strings.Contains(contentStr, "RegisterWorkflow") {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
