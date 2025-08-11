package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFilePath validates a file path for security issues
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for path traversal attempts
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	// Check for absolute paths that might escape the working directory
	if filepath.IsAbs(path) && !IsPathWithinWorkingDir(path) {
		return fmt.Errorf("absolute path outside working directory: %s", path)
	}

	return nil
}

// ValidateSchemaName validates a schema name to prevent injection
func ValidateSchemaName(name string) error {
	if name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	// Only allow alphanumeric, dash, underscore
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_') {
			return fmt.Errorf("invalid character in schema name: %c", r)
		}
	}

	// Prevent directory traversal in schema names
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("schema name cannot contain path separators")
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("schema name cannot contain '..'")
	}

	return nil
}

// ValidateFileExtension validates that a file has an allowed extension
func ValidateFileExtension(path string, allowedExtensions []string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}

	return fmt.Errorf("file extension %s not allowed, must be one of: %v", ext, allowedExtensions)
}

// IsPathWithinWorkingDir checks if an absolute path is within the current working directory
func IsPathWithinWorkingDir(absPath string) bool {
	wd, err := filepath.Abs(".")
	if err != nil {
		return false
	}

	// Clean and compare paths
	cleanWd := filepath.Clean(wd)
	cleanPath := filepath.Clean(absPath)

	// Check if the path starts with the working directory
	rel, err := filepath.Rel(cleanWd, cleanPath)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's outside the working directory
	return !strings.HasPrefix(rel, "..")
}

// ValidateOutputPath validates a path for writing output files
func ValidateOutputPath(path string) error {
	if err := ValidateFilePath(path); err != nil {
		return err
	}

	// Check if the directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		// This is just validation, we don't create directories here
		// The actual file writing code should handle directory creation
	}

	return nil
}
