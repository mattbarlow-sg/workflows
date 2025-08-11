package adr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// Renderer renders ADR data
type Renderer struct{}

// NewRenderer creates a new ADR renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render renders ADR to markdown
func (r *Renderer) Render(adr *ADR) (string, error) {
	return adr.ToMarkdown(), nil
}

// RenderFile reads an ADR file and renders it to markdown
func (r *Renderer) RenderFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	adr, err := FromJSON(data)
	if err != nil {
		// Try to provide validation error details
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return "", &ValidationError{
				Errors: []string{fmt.Sprintf("JSON syntax error at position %d: %v", syntaxErr.Offset, err)},
			}
		}

		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return "", &ValidationError{
				Errors: []string{fmt.Sprintf("Type error: expected %s for field %s, got %s", typeErr.Type, typeErr.Field, typeErr.Value)},
			}
		}

		return "", &ValidationError{
			Errors: []string{fmt.Sprintf("Invalid JSON: %v", err)},
		}
	}

	return r.Render(adr)
}

// ValidationError represents validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(e.Errors))
}
