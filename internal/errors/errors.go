package errors

import (
	"errors"
	"fmt"
)

type ErrorType int

const (
	ErrorTypeUsage ErrorType = iota + 1
	ErrorTypeValidation
	ErrorTypeIO
	ErrorTypeConfig
	ErrorTypeInternal
)

type CLIError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

func (e *CLIError) ExitCode() int {
	switch e.Type {
	case ErrorTypeUsage:
		return 1
	case ErrorTypeValidation:
		return 2
	case ErrorTypeIO:
		return 3
	case ErrorTypeConfig:
		return 4
	case ErrorTypeInternal:
		return 5
	default:
		return 1
	}
}

func NewUsageError(message string) *CLIError {
	return &CLIError{
		Type:    ErrorTypeUsage,
		Message: message,
	}
}

func NewValidationError(message string, err error) *CLIError {
	return &CLIError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

func NewIOError(message string, err error) *CLIError {
	return &CLIError{
		Type:    ErrorTypeIO,
		Message: message,
		Err:     err,
	}
}

func NewConfigError(message string, err error) *CLIError {
	return &CLIError{
		Type:    ErrorTypeConfig,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *CLIError {
	return &CLIError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func IsCLIError(err error) (*CLIError, bool) {
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr, true
	}
	return nil, false
}