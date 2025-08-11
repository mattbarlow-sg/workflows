// Package schemas contains the core type definitions for the Temporal validation framework
package schemas

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

// ValidationRequest represents a request to validate Temporal workflows
type ValidationRequest struct {
	// WorkflowPath is the absolute path to the workflow file or directory
	WorkflowPath string `json:"workflow_path" validate:"required,filepath"`

	// WorkflowID is the unique identifier for the workflow
	WorkflowID string `json:"workflow_id" validate:"required,min=1,max=255"`

	// Options contains optional validation settings
	Options ValidationOptions `json:"options,omitempty"`

	// Context for cancellation and timeout
	Context context.Context `json:"-"`
}

// ValidationOptions configures validation behavior
type ValidationOptions struct {
	// SkipCache bypasses cache lookup
	SkipCache bool `json:"skip_cache,omitempty"`

	// MaxInfoMessages limits info message output (default 100)
	MaxInfoMessages int `json:"max_info_messages,omitempty" validate:"min=0,max=1000"`

	// Timeout for validation (max 5 minutes)
	Timeout time.Duration `json:"timeout,omitempty" validate:"max=300s"`

	// ParallelChecks enables concurrent validation (default true)
	ParallelChecks bool `json:"parallel_checks,omitempty"`
}

// ValidationResult contains the complete validation outcome
type ValidationResult struct {
	// Success indicates if all validations passed
	Success bool `json:"success"`

	// Errors contains all validation errors found
	Errors []ValidationError `json:"errors,omitempty"`

	// InfoMessages contains contextual information
	InfoMessages []InfoMessage `json:"info_messages,omitempty"`

	// CheckStatus tracks the status of each validation check
	CheckStatus CheckStatusMap `json:"check_status"`

	// Duration is the total validation time
	Duration time.Duration `json:"duration"`

	// CacheHit indicates if result was from cache
	CacheHit bool `json:"cache_hit"`

	// Timestamp when validation was performed
	Timestamp time.Time `json:"timestamp"`

	// SourceHash is the hash of the validated source
	SourceHash string `json:"source_hash"`
}

// CheckStatusMap tracks the status of individual checks
type CheckStatusMap struct {
	// DeterminismCheck status
	DeterminismCheck CheckStatus `json:"determinism_check"`

	// SignatureCheck status
	SignatureCheck CheckStatus `json:"signature_check"`

	// PolicyCheck status
	PolicyCheck CheckStatus `json:"policy_check"`
}

// CheckStatus represents the status of a validation check
type CheckStatus string

const (
	// CheckStatusPending indicates check not started
	CheckStatusPending CheckStatus = "PENDING"

	// CheckStatusRunning indicates check in progress
	CheckStatusRunning CheckStatus = "RUNNING"

	// CheckStatusCompleted indicates check finished successfully
	CheckStatusCompleted CheckStatus = "COMPLETED"

	// CheckStatusCancelled indicates check was cancelled
	CheckStatusCancelled CheckStatus = "CANCELLED"

	// CheckStatusSkipped indicates check was skipped
	CheckStatusSkipped CheckStatus = "SKIPPED"

	// CheckStatusFailed indicates check encountered an error
	CheckStatusFailed CheckStatus = "FAILED"
)

// ValidationError represents a validation failure
type ValidationError struct {
	// Severity of the error (ERROR only for blocking issues)
	Severity ErrorSeverity `json:"severity"`

	// Category classifies the error type
	Category ErrorCategory `json:"category"`

	// Code is a unique error identifier
	Code string `json:"code" validate:"required"`

	// Message describes the error
	Message string `json:"message" validate:"required"`

	// Location indicates where the error occurred
	Location CodeLocation `json:"location,omitempty"`

	// FixInstructions provides guidance to resolve the error
	FixInstructions string `json:"fix_instructions,omitempty"`

	// Context provides additional error context
	Context map[string]interface{} `json:"context,omitempty"`
}

// ErrorSeverity indicates the impact level
type ErrorSeverity string

const (
	// ErrorSeverityError blocks deployment
	ErrorSeverityError ErrorSeverity = "ERROR"

	// ErrorSeverityInfo provides context only
	ErrorSeverityInfo ErrorSeverity = "INFO"
)

// ErrorCategory classifies error types
type ErrorCategory string

const (
	// ErrorCategoryPermission indicates permission denied
	ErrorCategoryPermission ErrorCategory = "PERMISSION"

	// ErrorCategorySyntax indicates syntax error
	ErrorCategorySyntax ErrorCategory = "SYNTAX"

	// ErrorCategoryMissingDeps indicates missing dependencies
	ErrorCategoryMissingDeps ErrorCategory = "MISSING_DEPS"

	// ErrorCategoryTypeMismatch indicates type error
	ErrorCategoryTypeMismatch ErrorCategory = "TYPE_MISMATCH"

	// ErrorCategoryInvalidConfig indicates configuration error
	ErrorCategoryInvalidConfig ErrorCategory = "INVALID_CONFIG"

	// ErrorCategoryDeterminism indicates non-deterministic code
	ErrorCategoryDeterminism ErrorCategory = "DETERMINISM"

	// ErrorCategoryNaming indicates naming convention violation
	ErrorCategoryNaming ErrorCategory = "NAMING"

	// ErrorCategoryTimeout indicates timeout policy violation
	ErrorCategoryTimeout ErrorCategory = "TIMEOUT"

	// ErrorCategoryRetry indicates retry policy violation
	ErrorCategoryRetry ErrorCategory = "RETRY"
)

// IsRetryable returns true if the error category is retryable
func (c ErrorCategory) IsRetryable() bool {
	switch c {
	case ErrorCategoryPermission, ErrorCategorySyntax,
		ErrorCategoryMissingDeps, ErrorCategoryTypeMismatch,
		ErrorCategoryInvalidConfig:
		return false
	default:
		return true
	}
}

// CodeLocation identifies a position in source code
type CodeLocation struct {
	// File path relative to workflow root
	File string `json:"file"`

	// Line number (1-indexed)
	Line int `json:"line" validate:"min=1"`

	// Column number (1-indexed)
	Column int `json:"column,omitempty" validate:"min=0"`

	// Snippet of code around the error
	Snippet string `json:"snippet,omitempty"`
}

// InfoMessage provides contextual information
type InfoMessage struct {
	// Category classifies the info message
	Category InfoCategory `json:"category"`

	// Message content
	Message string `json:"message" validate:"required"`

	// Timestamp when message was generated
	Timestamp time.Time `json:"timestamp"`

	// Data contains structured information
	Data map[string]interface{} `json:"data,omitempty"`
}

// InfoCategory classifies info messages
type InfoCategory string

const (
	// InfoCategoryMetrics provides workflow metrics
	InfoCategoryMetrics InfoCategory = "METRICS"

	// InfoCategoryProgress indicates validation progress
	InfoCategoryProgress InfoCategory = "PROGRESS"

	// InfoCategoryCache provides cache information
	InfoCategoryCache InfoCategory = "CACHE"

	// InfoCategoryPerformance shows performance metrics
	InfoCategoryPerformance InfoCategory = "PERFORMANCE"

	// InfoCategoryConfig shows configuration details
	InfoCategoryConfig InfoCategory = "CONFIG"
)

// DeterminismCheckResult contains determinism validation results
type DeterminismCheckResult struct {
	// Passed indicates if check succeeded
	Passed bool `json:"passed"`

	// Violations lists all determinism violations found
	Violations []DeterminismViolation `json:"violations,omitempty"`

	// Duration of the check
	Duration time.Duration `json:"duration"`
}

// DeterminismViolation represents a non-deterministic pattern
type DeterminismViolation struct {
	// Pattern that was violated
	Pattern string `json:"pattern" validate:"required"`

	// Location in source code
	Location CodeLocation `json:"location"`

	// Description of the violation
	Description string `json:"description" validate:"required"`

	// Suggestion for fixing
	Suggestion string `json:"suggestion,omitempty"`
}

// ActivitySignatureResult contains activity signature validation results
type ActivitySignatureResult struct {
	// Passed indicates if check succeeded
	Passed bool `json:"passed"`

	// Activities found and validated
	Activities []ActivityValidation `json:"activities,omitempty"`

	// Violations lists naming or signature issues
	Violations []SignatureViolation `json:"violations,omitempty"`

	// Duration of the check
	Duration time.Duration `json:"duration"`
}

// ActivityValidation represents validation of a single activity
type ActivityValidation struct {
	// Name of the activity
	Name string `json:"name" validate:"required"`

	// Valid indicates if activity passes all checks
	Valid bool `json:"valid"`

	// Issues found with this activity
	Issues []string `json:"issues,omitempty"`
}

// SignatureViolation represents an activity signature issue
type SignatureViolation struct {
	// ActivityName that has the issue
	ActivityName string `json:"activity_name" validate:"required"`

	// ViolationType describes the issue
	ViolationType string `json:"violation_type" validate:"required"`

	// Expected signature or naming
	Expected string `json:"expected,omitempty"`

	// Actual signature or naming
	Actual string `json:"actual,omitempty"`

	// Location in source code
	Location CodeLocation `json:"location"`
}

// PolicyValidationResult contains timeout and retry policy validation results
type PolicyValidationResult struct {
	// Passed indicates if check succeeded
	Passed bool `json:"passed"`

	// TimeoutViolations lists timeout policy issues
	TimeoutViolations []TimeoutViolation `json:"timeout_violations,omitempty"`

	// RetryViolations lists retry policy issues
	RetryViolations []RetryViolation `json:"retry_violations,omitempty"`

	// Duration of the check
	Duration time.Duration `json:"duration"`
}

// TimeoutViolation represents a timeout policy issue
type TimeoutViolation struct {
	// WorkflowOrActivity name
	Name string `json:"name" validate:"required"`

	// IsHumanTask indicates if this is a human task
	IsHumanTask bool `json:"is_human_task"`

	// ConfiguredTimeout in seconds
	ConfiguredTimeout int `json:"configured_timeout"`

	// RequiredTimeout based on rules
	RequiredTimeout string `json:"required_timeout"`

	// Location in source code
	Location CodeLocation `json:"location"`
}

// RetryViolation represents a retry policy issue
type RetryViolation struct {
	// ActivityName with retry issue
	ActivityName string `json:"activity_name" validate:"required"`

	// ConfiguredRetries count
	ConfiguredRetries int `json:"configured_retries"`

	// MinimumRequired retries
	MinimumRequired int `json:"minimum_required"`

	// ErrorType that should/shouldn't be retried
	ErrorType string `json:"error_type,omitempty"`

	// Location in source code
	Location CodeLocation `json:"location"`
}

// CacheEntry represents a cached validation result
type CacheEntry struct {
	// Key is the composite cache key
	Key string `json:"key" validate:"required"`

	// WorkflowID that was validated
	WorkflowID string `json:"workflow_id" validate:"required"`

	// SourceHash of the validated source
	SourceHash string `json:"source_hash" validate:"required"`

	// Result of the validation
	Result ValidationResult `json:"result"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at"`

	// AccessedAt timestamp (updated on cache hit)
	AccessedAt time.Time `json:"accessed_at"`

	// AccessCount tracks cache hits
	AccessCount int `json:"access_count"`
}

// CacheKey generates a composite cache key
func CacheKey(workflowID string, sourceHash string) string {
	return fmt.Sprintf("%s:%s", workflowID, sourceHash)
}

// ComputeSourceHash calculates SHA256 hash of source content
func ComputeSourceHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// ValidationState represents the state machine state
type ValidationState string

const (
	// ValidationStateIdle waiting for request
	ValidationStateIdle ValidationState = "IDLE"

	// ValidationStateValidating executing checks
	ValidationStateValidating ValidationState = "VALIDATING"

	// ValidationStateCollectingErrors gathering errors
	ValidationStateCollectingErrors ValidationState = "COLLECTING_ERRORS"

	// ValidationStateReporting generating report
	ValidationStateReporting ValidationState = "REPORTING"

	// ValidationStateCompleted finished
	ValidationStateCompleted ValidationState = "COMPLETED"

	// ValidationStateCached using cached result
	ValidationStateCached ValidationState = "CACHED"
)

// ValidationContext maintains validation state
type ValidationContext struct {
	// State current state machine state
	State ValidationState `json:"state"`

	// Request being processed
	Request ValidationRequest `json:"request"`

	// StartTime of validation
	StartTime time.Time `json:"start_time"`

	// CancelFunc for cancellation
	CancelFunc context.CancelFunc `json:"-"`

	// Results from each check
	DeterminismResult *DeterminismCheckResult  `json:"determinism_result,omitempty"`
	SignatureResult   *ActivitySignatureResult `json:"signature_result,omitempty"`
	PolicyResult      *PolicyValidationResult  `json:"policy_result,omitempty"`

	// Errors collected during validation
	Errors []ValidationError `json:"errors,omitempty"`

	// InfoMessages collected
	InfoMessages []InfoMessage `json:"info_messages,omitempty"`
}
