// Package temporal provides activity signature validation for Temporal workflows
package temporal

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattbarlow-sg/workflows/src/schemas"
)

// ActivitySignatureValidatorImpl implements activity signature validation
type ActivitySignatureValidatorImpl struct{}

// Validate checks all activity signatures in the workflow
func (a *ActivitySignatureValidatorImpl) Validate(ctx context.Context, workflowPath string) (*schemas.ActivitySignatureResult, error) {
	startTime := time.Now()
	activities := []schemas.ActivityValidation{}
	violations := []schemas.SignatureViolation{}

	// Find all activity files
	files, err := a.findActivityFiles(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find activity files: %w", err)
	}

	// Validate each file
	for _, file := range files {
		fileActivities, fileViolations, err := a.validateFile(ctx, file)
		if err != nil {
			// Record error as violation
			violations = append(violations, schemas.SignatureViolation{
				ActivityName:  filepath.Base(file),
				ViolationType: "parse_error",
				Expected:      "valid Go syntax",
				Actual:        fmt.Sprintf("parse error: %v", err),
				Location: schemas.CodeLocation{
					File: file,
					Line: 1,
				},
			})
			continue
		}

		activities = append(activities, fileActivities...)
		violations = append(violations, fileViolations...)
	}

	return &schemas.ActivitySignatureResult{
		Passed:     len(violations) == 0,
		Activities: activities,
		Violations: violations,
		Duration:   time.Since(startTime),
	}, nil
}

// validateFile validates all activities in a single file
func (a *ActivitySignatureValidatorImpl) validateFile(ctx context.Context, filePath string) ([]schemas.ActivityValidation, []schemas.SignatureViolation, error) {
	file, fset, err := parseGoFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	activities := []schemas.ActivityValidation{}
	violations := []schemas.SignatureViolation{}

	// Get relative path
	relPath := filePath
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, filePath); err == nil {
			relPath = rel
		}
	}

	// Find all function declarations
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if this looks like an activity function
		if !a.isActivityFunction(funcDecl) {
			return true
		}

		activityName := funcDecl.Name.Name
		issues := []string{}
		pos := fset.Position(funcDecl.Pos())

		// Validate naming convention (PascalCase)
		if err := a.validateActivityName(activityName); err != nil {
			violations = append(violations, schemas.SignatureViolation{
				ActivityName:  activityName,
				ViolationType: "naming_convention",
				Expected:      "PascalCase (e.g., ProcessOrder)",
				Actual:        activityName,
				Location: schemas.CodeLocation{
					File:   relPath,
					Line:   pos.Line,
					Column: pos.Column,
				},
			})
			issues = append(issues, err.Error())
		}

		// Validate function signature
		signatureIssues := a.validateSignature(funcDecl, fset)
		for _, issue := range signatureIssues {
			violations = append(violations, issue)
			issues = append(issues, fmt.Sprintf("%s: %s", issue.ViolationType, issue.Expected))
		}

		// Validate return types
		returnIssues := a.validateReturnTypes(funcDecl, fset)
		for _, issue := range returnIssues {
			violations = append(violations, issue)
			issues = append(issues, fmt.Sprintf("%s: %s", issue.ViolationType, issue.Expected))
		}

		activities = append(activities, schemas.ActivityValidation{
			Name:   activityName,
			Valid:  len(issues) == 0,
			Issues: issues,
		})

		return true
	})

	return activities, violations, nil
}

// isActivityFunction determines if a function is likely an activity
func (a *ActivitySignatureValidatorImpl) isActivityFunction(funcDecl *ast.FuncDecl) bool {
	// Check function name patterns
	name := funcDecl.Name.Name

	// Skip workflow functions
	if strings.Contains(name, "Workflow") {
		return false
	}

	// Common activity prefixes/suffixes
	activityPatterns := []string{
		"Activity",
		"Process",
		"Send",
		"Fetch",
		"Update",
		"Create",
		"Delete",
		"Get",
		"Set",
		"Validate",
		"Execute",
		"Handle",
	}

	for _, pattern := range activityPatterns {
		if strings.HasPrefix(name, pattern) || strings.HasSuffix(name, pattern) {
			return true
		}
	}

	// Check if it has context.Context as first parameter
	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		firstParam := funcDecl.Type.Params.List[0]
		if a.isContextType(firstParam.Type) {
			return true
		}
	}

	// Check if it returns error
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			if a.isErrorType(result.Type) {
				return true
			}
		}
	}

	return false
}

// validateActivityName checks if the activity name follows conventions
func (a *ActivitySignatureValidatorImpl) validateActivityName(name string) error {
	if name == "" {
		return fmt.Errorf("activity name cannot be empty")
	}

	// Check for PascalCase
	if name[0] < 'A' || name[0] > 'Z' {
		return fmt.Errorf("activity name must start with uppercase letter")
	}

	// Check for invalid characters
	for i, ch := range name {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			if ch == '_' {
				return fmt.Errorf("activity name should use PascalCase instead of underscores")
			}
			return fmt.Errorf("activity name contains invalid character at position %d", i)
		}
	}

	// Check for common anti-patterns
	if strings.Contains(name, "__") {
		return fmt.Errorf("activity name should not contain double underscores")
	}

	if strings.HasPrefix(name, "Do") && len(name) > 2 {
		// "Do" prefix is often redundant
		return fmt.Errorf("consider removing 'Do' prefix from activity name")
	}

	return nil
}

// validateSignature validates the activity function signature
func (a *ActivitySignatureValidatorImpl) validateSignature(funcDecl *ast.FuncDecl, fset *token.FileSet) []schemas.SignatureViolation {
	violations := []schemas.SignatureViolation{}
	activityName := funcDecl.Name.Name
	pos := fset.Position(funcDecl.Pos())
	relPath := pos.Filename

	// Get relative path
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, pos.Filename); err == nil {
			relPath = rel
		}
	}

	// Check parameters
	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		// First parameter should be context.Context
		firstParam := funcDecl.Type.Params.List[0]
		if !a.isContextType(firstParam.Type) {
			violations = append(violations, schemas.SignatureViolation{
				ActivityName:  activityName,
				ViolationType: "missing_context",
				Expected:      "context.Context as first parameter",
				Actual:        a.typeToString(firstParam.Type),
				Location: schemas.CodeLocation{
					File:   relPath,
					Line:   pos.Line,
					Column: pos.Column,
				},
			})
		}

		// Check for too many parameters (best practice)
		if len(funcDecl.Type.Params.List) > 3 {
			violations = append(violations, schemas.SignatureViolation{
				ActivityName:  activityName,
				ViolationType: "too_many_parameters",
				Expected:      "maximum 3 parameters (use struct for complex inputs)",
				Actual:        fmt.Sprintf("%d parameters", len(funcDecl.Type.Params.List)),
				Location: schemas.CodeLocation{
					File:   relPath,
					Line:   pos.Line,
					Column: pos.Column,
				},
			})
		}
	}

	return violations
}

// validateReturnTypes validates the activity return types
func (a *ActivitySignatureValidatorImpl) validateReturnTypes(funcDecl *ast.FuncDecl, fset *token.FileSet) []schemas.SignatureViolation {
	violations := []schemas.SignatureViolation{}
	activityName := funcDecl.Name.Name
	pos := fset.Position(funcDecl.Pos())
	relPath := pos.Filename

	// Get relative path
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, pos.Filename); err == nil {
			relPath = rel
		}
	}

	// Check return values
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		violations = append(violations, schemas.SignatureViolation{
			ActivityName:  activityName,
			ViolationType: "missing_error_return",
			Expected:      "error as return value",
			Actual:        "no return values",
			Location: schemas.CodeLocation{
				File:   relPath,
				Line:   pos.Line,
				Column: pos.Column,
			},
		})
		return violations
	}

	// Last return value should be error
	lastResult := funcDecl.Type.Results.List[len(funcDecl.Type.Results.List)-1]
	if !a.isErrorType(lastResult.Type) {
		violations = append(violations, schemas.SignatureViolation{
			ActivityName:  activityName,
			ViolationType: "missing_error_return",
			Expected:      "error as last return value",
			Actual:        a.typeToString(lastResult.Type),
			Location: schemas.CodeLocation{
				File:   relPath,
				Line:   pos.Line,
				Column: pos.Column,
			},
		})
	}

	// Check for too many return values (best practice)
	if len(funcDecl.Type.Results.List) > 2 {
		violations = append(violations, schemas.SignatureViolation{
			ActivityName:  activityName,
			ViolationType: "too_many_returns",
			Expected:      "maximum 2 return values (result, error)",
			Actual:        fmt.Sprintf("%d return values", len(funcDecl.Type.Results.List)),
			Location: schemas.CodeLocation{
				File:   relPath,
				Line:   pos.Line,
				Column: pos.Column,
			},
		})
	}

	return violations
}

// isContextType checks if a type is context.Context
func (a *ActivitySignatureValidatorImpl) isContextType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name == "context" && t.Sel.Name == "Context"
		}
	case *ast.Ident:
		// Could be an alias
		return t.Name == "Context"
	}
	return false
}

// isErrorType checks if a type is error
func (a *ActivitySignatureValidatorImpl) isErrorType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "error"
	}
	return false
}

// typeToString converts an AST type to string representation
func (a *ActivitySignatureValidatorImpl) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", ident.Name, t.Sel.Name)
		}
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", a.typeToString(t.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", a.typeToString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", a.typeToString(t.Key), a.typeToString(t.Value))
	}
	return "unknown"
}

// findActivityFiles finds all Go files that contain activity definitions
func (a *ActivitySignatureValidatorImpl) findActivityFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// Single file
		return []string{path}, nil
	}

	// Directory - find all activity files
	var files []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Look for Go files
		if !info.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			// Check if it contains activity code
			content, err := ioutil.ReadFile(p)
			if err != nil {
				return nil // Skip files we can't read
			}

			contentStr := string(content)
			// Look for activity signatures
			if strings.Contains(contentStr, "context.Context") ||
				strings.Contains(contentStr, "Activity") ||
				strings.Contains(contentStr, "RegisterActivity") ||
				strings.Contains(contentStr, "go.temporal.io/sdk/activity") {
				files = append(files, p)
			}
		}

		return nil
	})

	return files, err
}
