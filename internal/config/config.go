package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	SchemaDir string
}

// New creates a new configuration instance with defaults and environment overrides
func New() *Config {
	cfg := &Config{
		SchemaDir: getSchemaDir(),
	}
	return cfg
}

// getSchemaDir determines the schema directory path with the following precedence:
// 1. WORKFLOWS_SCHEMA_DIR environment variable
// 2. ./schemas (relative to current directory)
// 3. schemas directory relative to executable location
// 4. /etc/workflows/schemas (system location)
func getSchemaDir() string {
	// Check environment variable first
	if envDir := os.Getenv("WORKFLOWS_SCHEMA_DIR"); envDir != "" {
		if absPath, err := filepath.Abs(envDir); err == nil {
			return absPath
		}
		return envDir
	}

	// Try common locations in order
	locations := []string{
		filepath.Join(".", "schemas"),                    // Current directory
		getExecutableRelativePath("schemas"),             // Next to executable
		getExecutableRelativePath("../schemas"),          // Parent of executable
		"/etc/workflows/schemas",                         // System location
		filepath.Join(os.Getenv("HOME"), ".workflows", "schemas"), // User home
	}

	// Return the first location that exists
	for _, loc := range locations {
		if loc != "" && dirExists(loc) {
			if absPath, err := filepath.Abs(loc); err == nil {
				return absPath
			}
			return loc
		}
	}

	// Default to ./schemas even if it doesn't exist yet
	return filepath.Join(".", "schemas")
}

// getExecutableRelativePath returns a path relative to the executable location
func getExecutableRelativePath(relPath string) string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	
	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		realPath = exePath
	}
	
	exeDir := filepath.Dir(realPath)
	return filepath.Join(exeDir, relPath)
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetSchemaPath returns the full path to a schema file
func (c *Config) GetSchemaPath(schemaName string) string {
	// Add .json extension if not present
	if filepath.Ext(schemaName) == "" {
		schemaName = schemaName + ".json"
	}
	return filepath.Join(c.SchemaDir, schemaName)
}