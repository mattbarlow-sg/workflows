package validation

import (
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty path", "", true},
		{"simple filename", "test.json", false},
		{"relative path", "data/test.json", false},
		{"path traversal attempt", "../../../etc/passwd", true},
		{"path traversal in middle", "data/../../etc/passwd", true},
		{"hidden path traversal", "data/../../../etc/passwd", true},
		{"dot file", ".gitignore", false},
		{"current directory", ".", false},
		{"nested relative", "./data/test.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSchemaName(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{"empty name", "", true},
		{"simple name", "adr", false},
		{"with dash", "adr-schema", false},
		{"with underscore", "adr_schema", false},
		{"with numbers", "adr123", false},
		{"with slash", "adr/schema", true},
		{"with backslash", "adr\\schema", true},
		{"with dots", "adr..schema", true},
		{"with space", "adr schema", true},
		{"with special char", "adr@schema", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchemaName(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSchemaName(%q) error = %v, wantErr %v", tt.schema, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		allowed    []string
		wantErr    bool
	}{
		{"json file allowed", "test.json", []string{".json"}, false},
		{"JSON uppercase", "test.JSON", []string{".json"}, false},
		{"multiple allowed", "test.yaml", []string{".json", ".yaml", ".yml"}, false},
		{"not allowed", "test.txt", []string{".json"}, true},
		{"no extension", "test", []string{".json"}, true},
		{"nested path", "data/test.json", []string{".json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.path, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileExtension(%q, %v) error = %v, wantErr %v", 
					tt.path, tt.allowed, err, tt.wantErr)
			}
		})
	}
}