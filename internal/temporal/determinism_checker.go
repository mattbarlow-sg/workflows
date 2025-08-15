// Package temporal provides determinism checking for Temporal workflows
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

// DeterminismCheckerImpl implements determinism validation
type DeterminismCheckerImpl struct{}

// Check performs determinism validation on workflow code
func (d *DeterminismCheckerImpl) Check(ctx context.Context, workflowPath string) (*schemas.DeterminismCheckResult, error) {
	startTime := time.Now()
	violations := []schemas.DeterminismViolation{}

	// Find all workflow files
	files, err := d.findWorkflowFiles(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow files: %w", err)
	}

	// Check each file for determinism violations using line-based checking
	for _, file := range files {
		fileViolations, err := d.checkFile(ctx, file)
		if err != nil {
			// Continue checking other files even if one fails
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "parse_error",
				Description: fmt.Sprintf("Failed to parse file: %v", err),
				Location: schemas.CodeLocation{
					File: file,
					Line: 1,
				},
			})
			continue
		}
		violations = append(violations, fileViolations...)
	}

	return &schemas.DeterminismCheckResult{
		Passed:     len(violations) == 0,
		Violations: violations,
		Duration:   time.Since(startTime),
	}, nil
}

// checkFile checks a single file for determinism violations using pattern matching
func (d *DeterminismCheckerImpl) checkFile(ctx context.Context, filePath string) ([]schemas.DeterminismViolation, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	violations := []schemas.DeterminismViolation{}
	lines := strings.Split(string(content), "\n")

	// Get relative path for better error messages
	relPath := filePath
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, filePath); err == nil {
			relPath = rel
		}
	}

	// Check each line against determinism patterns
	for lineNum, line := range lines {
		// Skip comments and strings
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") {
			continue
		}

		// Check for time.Now()
		if strings.Contains(line, "time.Now()") && !strings.Contains(line, "workflow.Now()") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "time_now",
				Description: "time.Now() is non-deterministic in workflows",
				Suggestion:  "Use workflow.Now() instead",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  strings.Index(line, "time.Now()") + 1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for native goroutines
		if strings.Contains(line, "go func") || (strings.HasPrefix(trimmedLine, "go ") && !strings.Contains(line, "workflow.Go")) {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "goroutine",
				Description: "Native goroutines are non-deterministic in workflows",
				Suggestion:  "Use workflow.Go() instead",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  strings.Index(line, "go ") + 1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for native channels
		if strings.Contains(line, "make(chan ") && !strings.Contains(line, "workflow.NewChannel") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "channel",
				Description: "Native channels are non-deterministic in workflows",
				Suggestion:  "Use workflow.NewChannel() instead",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  strings.Index(line, "chan ") + 1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for select statements
		if strings.HasPrefix(trimmedLine, "select {") && !strings.Contains(line, "workflow.Selector") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "select",
				Description: "Native select is non-deterministic in workflows",
				Suggestion:  "Use workflow.NewSelector() instead",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for math/rand usage
		if strings.Contains(line, "rand.") && !strings.Contains(line, "workflow.SideEffect") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "random",
				Description: "Random number generation is non-deterministic in workflows",
				Suggestion:  "Use workflow.SideEffect() for random values",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  strings.Index(line, "rand.") + 1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for os.Getenv
		if strings.Contains(line, "os.Getenv") && !strings.Contains(line, "workflow.SideEffect") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "environment",
				Description: "Environment variables are non-deterministic in workflows",
				Suggestion:  "Pass environment values as workflow inputs or use workflow.SideEffect()",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  strings.Index(line, "os.Getenv") + 1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for file I/O operations
		if (strings.Contains(line, "os.Open") || strings.Contains(line, "ioutil.Read") ||
			strings.Contains(line, "os.Create") || strings.Contains(line, "os.Write")) &&
			!strings.Contains(line, "activity") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "file_io",
				Description: "File I/O operations are non-deterministic in workflows",
				Suggestion:  "Move file operations to activities",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}

		// Check for network operations
		if (strings.Contains(line, "http.") || strings.Contains(line, "net.")) &&
			!strings.Contains(line, "activity") {
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "network",
				Description: "Network operations are non-deterministic in workflows",
				Suggestion:  "Move network operations to activities",
				Location: schemas.CodeLocation{
					File:    relPath,
					Line:    lineNum + 1,
					Column:  1,
					Snippet: strings.TrimSpace(line),
				},
			})
		}
	}

	return violations, nil
}

// checkFileAST performs AST-based determinism checks
func (d *DeterminismCheckerImpl) checkFileAST(ctx context.Context, filePath string) ([]schemas.DeterminismViolation, error) {
	var file *ast.File
	var fset *token.FileSet
	file, fset, err := parseGoFile(filePath)
	if err != nil {
		return nil, err
	}

	violations := []schemas.DeterminismViolation{}

	// Get relative path
	relPath := filePath
	if cwd, err := filepath.Abs("."); err == nil {
		if rel, err := filepath.Rel(cwd, filePath); err == nil {
			relPath = rel
		}
	}

	// Walk the AST looking for problematic patterns
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.RangeStmt:
			// Check for map iteration
			if d.isMapType(node.X) {
				pos := fset.Position(node.Pos())
				violations = append(violations, schemas.DeterminismViolation{
					Pattern:     "map_iteration",
					Description: "Map iteration order is non-deterministic",
					Suggestion:  "Sort map keys before iteration or use ordered data structures",
					Location: schemas.CodeLocation{
						File:   relPath,
						Line:   pos.Line,
						Column: pos.Column,
					},
				})
			}

		case *ast.CallExpr:
			// Check for problematic function calls
			if d.isFunctionCall(node, "time", "Now") {
				pos := fset.Position(node.Pos())
				violations = append(violations, schemas.DeterminismViolation{
					Pattern:     "time_now",
					Description: "time.Now() is non-deterministic",
					Suggestion:  "Use workflow.Now() instead",
					Location: schemas.CodeLocation{
						File:   relPath,
						Line:   pos.Line,
						Column: pos.Column,
					},
				})
			}

			if d.isFunctionCall(node, "time", "After") || d.isFunctionCall(node, "time", "Tick") {
				pos := fset.Position(node.Pos())
				violations = append(violations, schemas.DeterminismViolation{
					Pattern:     "timer",
					Description: "Native timers are non-deterministic",
					Suggestion:  "Use workflow.NewTimer() instead",
					Location: schemas.CodeLocation{
						File:   relPath,
						Line:   pos.Line,
						Column: pos.Column,
					},
				})
			}

		case *ast.GoStmt:
			// Check for goroutines
			pos := fset.Position(node.Pos())
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "goroutine",
				Description: "Native goroutines are non-deterministic",
				Suggestion:  "Use workflow.Go() instead",
				Location: schemas.CodeLocation{
					File:   relPath,
					Line:   pos.Line,
					Column: pos.Column,
				},
			})

		case *ast.SelectStmt:
			// Check for select statements
			pos := fset.Position(node.Pos())
			violations = append(violations, schemas.DeterminismViolation{
				Pattern:     "select",
				Description: "Native select is non-deterministic",
				Suggestion:  "Use workflow.NewSelector() instead",
				Location: schemas.CodeLocation{
					File:   relPath,
					Line:   pos.Line,
					Column: pos.Column,
				},
			})
		}

		return true
	})

	return violations, nil
}

// isMapType checks if an expression is a map type
func (d *DeterminismCheckerImpl) isMapType(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Simple heuristic: check if identifier contains "map" or "Map"
		return strings.Contains(strings.ToLower(e.Name), "map")
	case *ast.IndexExpr:
		// Could be a map access
		return true
	default:
		return false
	}
}

// isFunctionCall checks if a call expression matches a specific package and function
func (d *DeterminismCheckerImpl) isFunctionCall(call *ast.CallExpr, pkg, fn string) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if ident, ok := fun.X.(*ast.Ident); ok {
			return ident.Name == pkg && fun.Sel.Name == fn
		}
	}
	return false
}

// findWorkflowFiles finds all Go files that contain workflow code
func (d *DeterminismCheckerImpl) findWorkflowFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// Single file
		return []string{path}, nil
	}

	// Directory - find all workflow files
	var files []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Look for Go files (excluding tests for now)
		if !info.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			// Check if it contains workflow code
			content, err := ioutil.ReadFile(p)
			if err != nil {
				return nil // Skip files we can't read
			}

			contentStr := string(content)
			// Look for workflow signatures
			if strings.Contains(contentStr, "workflow.Context") ||
				strings.Contains(contentStr, "go.temporal.io/sdk/workflow") ||
				strings.Contains(contentStr, "func Workflow") {
				files = append(files, p)
			}
		}

		return nil
	})

	return files, err
}
