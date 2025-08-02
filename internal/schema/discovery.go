package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Schema struct {
	Name        string
	Path        string
	Title       string
	Description string
	Content     map[string]interface{}
}

type Registry struct {
	schemasDir string
	schemas    map[string]*Schema
}

func NewRegistry(schemasDir string) *Registry {
	return &Registry{
		schemasDir: schemasDir,
		schemas:    make(map[string]*Schema),
	}
}

func (r *Registry) Discover() error {
	r.schemas = make(map[string]*Schema)

	err := filepath.Walk(r.schemasDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		schema, err := r.loadSchema(path)
		if err != nil {
			return fmt.Errorf("failed to load schema %s: %w", path, err)
		}

		r.schemas[schema.Name] = schema
		return nil
	})

	return err
}

func (r *Registry) loadSchema(path string) (*Schema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(path), ".json")
	
	schema := &Schema{
		Name:    name,
		Path:    path,
		Content: content,
	}

	if title, ok := content["title"].(string); ok {
		schema.Title = title
	}

	if desc, ok := content["description"].(string); ok {
		schema.Description = desc
	}

	return schema, nil
}

func (r *Registry) List() []*Schema {
	schemas := make([]*Schema, 0, len(r.schemas))
	for _, schema := range r.schemas {
		schemas = append(schemas, schema)
	}
	return schemas
}

func (r *Registry) Get(name string) (*Schema, bool) {
	schema, ok := r.schemas[name]
	return schema, ok
}