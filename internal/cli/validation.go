package cli

import (
	"fmt"

	"github.com/mattbarlow-sg/workflows/internal/errors"
	"github.com/mattbarlow-sg/workflows/internal/validation"
)

// ValidationChain provides a fluent interface for chaining validations
type ValidationChain struct {
	err error
}

// NewValidationChain creates a new validation chain
func NewValidationChain() *ValidationChain {
	return &ValidationChain{}
}

// ValidateSchemaName validates a schema name
func (v *ValidationChain) ValidateSchemaName(name string, fieldName string) *ValidationChain {
	if v.err != nil {
		return v
	}
	
	if err := validation.ValidateSchemaName(name); err != nil {
		v.err = errors.NewValidationError(fmt.Sprintf("invalid %s", fieldName), err)
	}
	return v
}

// ValidateFilePath validates a file path
func (v *ValidationChain) ValidateFilePath(path string, fieldName string) *ValidationChain {
	if v.err != nil {
		return v
	}
	
	if err := validation.ValidateFilePath(path); err != nil {
		v.err = errors.NewValidationError(fmt.Sprintf("invalid %s", fieldName), err)
	}
	return v
}

// ValidateFileExtension validates file extension
func (v *ValidationChain) ValidateFileExtension(path string, extensions []string, fieldName string) *ValidationChain {
	if v.err != nil {
		return v
	}
	
	if err := validation.ValidateFileExtension(path, extensions); err != nil {
		v.err = errors.NewValidationError(fmt.Sprintf("invalid %s", fieldName), err)
	}
	return v
}

// ValidateRequired validates that a value is not empty
func (v *ValidationChain) ValidateRequired(value string, fieldName string) *ValidationChain {
	if v.err != nil {
		return v
	}
	
	if value == "" {
		v.err = errors.NewValidationError(fmt.Sprintf("%s is required", fieldName), nil)
	}
	return v
}

// Error returns the first error encountered in the chain
func (v *ValidationChain) Error() error {
	return v.err
}

// Valid returns true if no errors were encountered
func (v *ValidationChain) Valid() bool {
	return v.err == nil
}