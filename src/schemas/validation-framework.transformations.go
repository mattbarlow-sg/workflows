// Package schemas contains transformation contracts for the validation framework
package schemas

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TransformToValidationResult combines all check results into final validation result
func TransformToValidationResult(ctx *ValidationContext) *ValidationResult {
	result := &ValidationResult{
		Timestamp:  time.Now(),
		SourceHash: ctx.Request.Options.SkipCache.String(), // Will be replaced with actual hash
		Duration:   time.Since(ctx.StartTime),
		CacheHit:   false,
	}
	
	// Determine success based on all checks passing
	result.Success = determineSuccess(ctx)
	
	// Set check statuses
	result.CheckStatus = transformCheckStatuses(ctx)
	
	// Transform and aggregate errors
	result.Errors = transformErrors(ctx)
	
	// Filter and limit info messages
	result.InfoMessages = filterInfoMessages(ctx.InfoMessages, ctx.Request.Options.MaxInfoMessages)
	
	return result
}

// determineSuccess checks if all validations passed
func determineSuccess(ctx *ValidationContext) bool {
	// Check if any ERROR severity issues exist
	for _, err := range ctx.Errors {
		if err.Severity == ErrorSeverityError {
			return false
		}
	}
	
	// Check individual validation results
	if ctx.DeterminismResult != nil && !ctx.DeterminismResult.Passed {
		return false
	}
	if ctx.SignatureResult != nil && !ctx.SignatureResult.Passed {
		return false
	}
	if ctx.PolicyResult != nil && !ctx.PolicyResult.Passed {
		return false
	}
	
	return true
}

// transformCheckStatuses converts check results to status map
func transformCheckStatuses(ctx *ValidationContext) CheckStatusMap {
	statusMap := CheckStatusMap{
		DeterminismCheck: CheckStatusPending,
		SignatureCheck:   CheckStatusPending,
		PolicyCheck:      CheckStatusPending,
	}
	
	// Determinism check status
	if ctx.DeterminismResult != nil {
		if ctx.DeterminismResult.Passed {
			statusMap.DeterminismCheck = CheckStatusCompleted
		} else {
			statusMap.DeterminismCheck = CheckStatusFailed
		}
	} else if ctx.State == ValidationStateCollectingErrors {
		statusMap.DeterminismCheck = CheckStatusCancelled
	}
	
	// Signature check status
	if ctx.SignatureResult != nil {
		if ctx.SignatureResult.Passed {
			statusMap.SignatureCheck = CheckStatusCompleted
		} else {
			statusMap.SignatureCheck = CheckStatusFailed
		}
	} else if ctx.State == ValidationStateCollectingErrors {
		statusMap.SignatureCheck = CheckStatusCancelled
	}
	
	// Policy check status
	if ctx.PolicyResult != nil {
		if ctx.PolicyResult.Passed {
			statusMap.PolicyCheck = CheckStatusCompleted
		} else {
			statusMap.PolicyCheck = CheckStatusFailed
		}
	} else if ctx.State == ValidationStateCollectingErrors {
		statusMap.PolicyCheck = CheckStatusCancelled
	}
	
	return statusMap
}

// transformErrors aggregates and transforms all errors from checks
func transformErrors(ctx *ValidationContext) []ValidationError {
	var errors []ValidationError
	
	// Add existing context errors
	errors = append(errors, ctx.Errors...)
	
	// Transform determinism violations
	if ctx.DeterminismResult != nil {
		errors = append(errors, transformDeterminismViolations(ctx.DeterminismResult.Violations)...)
	}
	
	// Transform signature violations
	if ctx.SignatureResult != nil {
		errors = append(errors, transformSignatureViolations(ctx.SignatureResult.Violations)...)
	}
	
	// Transform policy violations
	if ctx.PolicyResult != nil {
		errors = append(errors, transformPolicyViolations(ctx.PolicyResult)...)
	}
	
	// Sort errors by severity then location
	sortErrors(errors)
	
	return errors
}

// transformDeterminismViolations converts determinism violations to validation errors
func transformDeterminismViolations(violations []DeterminismViolation) []ValidationError {
	errors := make([]ValidationError, 0, len(violations))
	
	for _, v := range violations {
		err := ValidationError{
			Severity:        ErrorSeverityError,
			Category:        ErrorCategoryDeterminism,
			Code:            fmt.Sprintf("DETERMINISM_%s", strings.ToUpper(v.Pattern)),
			Message:         v.Description,
			Location:        v.Location,
			FixInstructions: v.Suggestion,
			Context: map[string]interface{}{
				"pattern": v.Pattern,
			},
		}
		errors = append(errors, err)
	}
	
	return errors
}

// transformSignatureViolations converts signature violations to validation errors
func transformSignatureViolations(violations []SignatureViolation) []ValidationError {
	errors := make([]ValidationError, 0, len(violations))
	
	for _, v := range violations {
		category := ErrorCategoryTypeMismatch
		if strings.Contains(v.ViolationType, "naming") || strings.Contains(v.ViolationType, "PascalCase") {
			category = ErrorCategoryNaming
		}
		
		err := ValidationError{
			Severity: ErrorSeverityError,
			Category: category,
			Code:     fmt.Sprintf("SIGNATURE_%s", strings.ToUpper(v.ViolationType)),
			Message:  fmt.Sprintf("Activity '%s': %s", v.ActivityName, v.ViolationType),
			Location: v.Location,
			Context: map[string]interface{}{
				"activity": v.ActivityName,
				"expected": v.Expected,
				"actual":   v.Actual,
			},
		}
		
		// Add fix instructions for naming violations
		if category == ErrorCategoryNaming {
			err.FixInstructions = fmt.Sprintf("Rename activity to '%s' to follow PascalCase convention", v.Expected)
		}
		
		errors = append(errors, err)
	}
	
	return errors
}

// transformPolicyViolations converts policy violations to validation errors
func transformPolicyViolations(result *PolicyValidationResult) []ValidationError {
	var errors []ValidationError
	
	// Transform timeout violations
	for _, v := range result.TimeoutViolations {
		err := ValidationError{
			Severity: ErrorSeverityError,
			Category: ErrorCategoryTimeout,
			Code:     "TIMEOUT_POLICY_VIOLATION",
			Message:  fmt.Sprintf("%s: invalid timeout configuration", v.Name),
			Location: v.Location,
			Context: map[string]interface{}{
				"name":                v.Name,
				"is_human":            v.IsHumanTask,
				"configured_timeout":  v.ConfiguredTimeout,
				"required_timeout":    v.RequiredTimeout,
			},
		}
		
		if v.IsHumanTask {
			err.FixInstructions = "Human tasks must have infinite timeout (set to 0)"
		} else {
			err.FixInstructions = fmt.Sprintf("Set timeout to %s or less", v.RequiredTimeout)
		}
		
		errors = append(errors, err)
	}
	
	// Transform retry violations
	for _, v := range result.RetryViolations {
		err := ValidationError{
			Severity: ErrorSeverityError,
			Category: ErrorCategoryRetry,
			Code:     "RETRY_POLICY_VIOLATION",
			Message:  fmt.Sprintf("Activity '%s': insufficient retry configuration", v.ActivityName),
			Location: v.Location,
			Context: map[string]interface{}{
				"activity":           v.ActivityName,
				"configured_retries": v.ConfiguredRetries,
				"minimum_required":   v.MinimumRequired,
				"error_type":         v.ErrorType,
			},
			FixInstructions: fmt.Sprintf("Set retry count to at least %d for retryable errors", v.MinimumRequired),
		}
		
		errors = append(errors, err)
	}
	
	return errors
}

// sortErrors sorts errors by severity (ERROR first) then by location
func sortErrors(errors []ValidationError) {
	sort.Slice(errors, func(i, j int) bool {
		// Sort by severity first (ERROR before INFO)
		if errors[i].Severity != errors[j].Severity {
			return errors[i].Severity == ErrorSeverityError
		}
		
		// Then by file
		if errors[i].Location.File != errors[j].Location.File {
			return errors[i].Location.File < errors[j].Location.File
		}
		
		// Then by line number
		if errors[i].Location.Line != errors[j].Location.Line {
			return errors[i].Location.Line < errors[j].Location.Line
		}
		
		// Finally by column
		return errors[i].Location.Column < errors[j].Location.Column
	})
}

// filterInfoMessages limits and categorizes info messages
func filterInfoMessages(messages []InfoMessage, maxMessages int) []InfoMessage {
	if maxMessages <= 0 {
		maxMessages = 100 // Default limit
	}
	
	// Sort by category then timestamp
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].Category != messages[j].Category {
			return messages[i].Category < messages[j].Category
		}
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})
	
	// Limit to max messages
	if len(messages) > maxMessages {
		messages = messages[:maxMessages]
	}
	
	return messages
}

// FormatValidationError formats an error for output
func FormatValidationError(err ValidationError) string {
	var parts []string
	
	// Add location if available
	if err.Location.File != "" {
		location := fmt.Sprintf("%s:%d", err.Location.File, err.Location.Line)
		if err.Location.Column > 0 {
			location = fmt.Sprintf("%s:%d", location, err.Location.Column)
		}
		parts = append(parts, location)
	}
	
	// Add severity and code
	parts = append(parts, fmt.Sprintf("[%s] %s", err.Severity, err.Code))
	
	// Add message
	parts = append(parts, err.Message)
	
	// Add fix instructions if available
	if err.FixInstructions != "" {
		parts = append(parts, fmt.Sprintf("\n  Fix: %s", err.FixInstructions))
	}
	
	// Add code snippet if available
	if err.Location.Snippet != "" {
		parts = append(parts, fmt.Sprintf("\n  %s", err.Location.Snippet))
	}
	
	return strings.Join(parts, ": ")
}

// FormatInfoMessage formats an info message for output
func FormatInfoMessage(msg InfoMessage) string {
	return fmt.Sprintf("[%s] %s", msg.Category, msg.Message)
}

// GenerateValidationReport creates a formatted report from validation result
func GenerateValidationReport(result *ValidationResult) string {
	var report strings.Builder
	
	// Header
	if result.Success {
		report.WriteString("✅ All Temporal workflow validations passed\n")
	} else {
		report.WriteString("❌ Temporal workflow validation failed\n")
	}
	
	// Summary
	report.WriteString(fmt.Sprintf("\nValidation Summary:\n"))
	report.WriteString(fmt.Sprintf("  Duration: %s\n", result.Duration))
	report.WriteString(fmt.Sprintf("  Cache Hit: %v\n", result.CacheHit))
	report.WriteString(fmt.Sprintf("  Errors: %d\n", len(result.Errors)))
	report.WriteString(fmt.Sprintf("  Info Messages: %d\n", len(result.InfoMessages)))
	
	// Check statuses
	report.WriteString("\nCheck Status:\n")
	report.WriteString(fmt.Sprintf("  Determinism Check: %s\n", result.CheckStatus.DeterminismCheck))
	report.WriteString(fmt.Sprintf("  Signature Check: %s\n", result.CheckStatus.SignatureCheck))
	report.WriteString(fmt.Sprintf("  Policy Check: %s\n", result.CheckStatus.PolicyCheck))
	
	// Errors
	if len(result.Errors) > 0 {
		report.WriteString("\nValidation Errors:\n")
		for i, err := range result.Errors {
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, FormatValidationError(err)))
		}
	}
	
	// Info messages
	if len(result.InfoMessages) > 0 {
		report.WriteString("\nInfo Messages:\n")
		for _, msg := range result.InfoMessages {
			report.WriteString(fmt.Sprintf("  %s\n", FormatInfoMessage(msg)))
		}
	}
	
	return report.String()
}

// TransformCacheEntry creates a cache entry from validation result
func TransformCacheEntry(workflowID string, sourceHash string, result *ValidationResult) *CacheEntry {
	now := time.Now()
	return &CacheEntry{
		Key:         CacheKey(workflowID, sourceHash),
		WorkflowID:  workflowID,
		SourceHash:  sourceHash,
		Result:      *result,
		CreatedAt:   now,
		AccessedAt:  now,
		AccessCount: 0,
	}
}