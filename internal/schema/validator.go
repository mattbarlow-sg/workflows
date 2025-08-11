package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func ValidateFile(schemaPath, filePath string) (*ValidationResult, error) {
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", err)
	}
	defer schemaFile.Close()

	schemaData, err := io.ReadAll(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	dataFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}
	defer dataFile.Close()

	data, err := io.ReadAll(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	return ValidateJSON(schemaData, data)
}

func ValidateJSON(schemaData, data []byte) (*ValidationResult, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	vr := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]string, 0),
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			vr.Errors = append(vr.Errors, err.String())
		}
	}

	return vr, nil
}

func ValidateObject(schema *Schema, data interface{}) (*ValidationResult, error) {
	schemaBytes, err := json.Marshal(schema.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return ValidateJSON(schemaBytes, dataBytes)
}
