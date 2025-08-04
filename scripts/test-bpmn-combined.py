#!/usr/bin/env python3
"""
Combine BPMN schemas into a single file for validation testing
"""

import json
from pathlib import Path

def combine_schemas():
    """Combine all BPMN schemas into a single schema file"""
    
    # Load all BPMN schemas
    schemas = {}
    schema_files = [
        "schemas/bpmn-common.json",
        "schemas/bpmn-flow-objects.json",
        "schemas/bpmn-connectors.json",
        "schemas/bpmn-artifacts.json",
        "schemas/bpmn-agents.json",
        "schemas/bpmn-process.json"
    ]
    
    for filepath in schema_files:
        with open(filepath, 'r') as f:
            schema = json.load(f)
            name = Path(filepath).stem
            schemas[name] = schema
    
    # Start with the process schema as base
    combined = schemas['bpmn-process'].copy()
    
    # Merge all definitions into a single definitions object
    all_definitions = {}
    
    for name, schema in schemas.items():
        if 'definitions' in schema:
            # Prefix definitions with schema name to avoid conflicts
            for def_name, definition in schema['definitions'].items():
                if name == 'bpmn-process':
                    # Keep process definitions without prefix
                    all_definitions[def_name] = definition
                else:
                    # Add prefix for other schemas
                    prefix = name.replace('bpmn-', '') + '_'
                    all_definitions[prefix + def_name] = definition
    
    combined['definitions'] = all_definitions
    
    # Update all $ref to use local definitions
    def update_refs(obj, schema_name):
        if isinstance(obj, dict):
            new_obj = {}
            for key, value in obj.items():
                if key == "$ref" and isinstance(value, str):
                    # Convert external ref to local ref
                    if value.startswith("bpmn-"):
                        parts = value.split("#/definitions/")
                        if len(parts) == 2:
                            source_schema = parts[0].replace(".json", "")
                            def_name = parts[1]
                            if source_schema == "bpmn-process":
                                new_obj[key] = f"#/definitions/{def_name}"
                            else:
                                prefix = source_schema.replace('bpmn-', '') + '_'
                                new_obj[key] = f"#/definitions/{prefix}{def_name}"
                        else:
                            new_obj[key] = value
                    else:
                        new_obj[key] = value
                else:
                    new_obj[key] = update_refs(value, schema_name)
            return new_obj
        elif isinstance(obj, list):
            return [update_refs(item, schema_name) for item in obj]
        else:
            return obj
    
    # Update all references in the combined schema
    combined = update_refs(combined, 'bpmn-process')
    
    # Remove the $id to avoid conflicts
    if '$id' in combined:
        del combined['$id']
    
    # Save the combined schema
    with open('schemas/bpmn-combined.json', 'w') as f:
        json.dump(combined, f, indent=2)
    
    print("Created combined schema: schemas/bpmn-combined.json")
    
    # Test validation with the combined schema
    print("\nTesting validation with combined schema...")
    test_files = [
        "test-data/simple-process.json",
        "test-data/parallel-gateway.json",
        "test-data/exclusive-gateway.json"
    ]
    
    for test_file in test_files:
        print(f"\nValidating {test_file}...")
        import subprocess
        result = subprocess.run(
            ['ajv', 'validate', '-s', 'schemas/bpmn-combined.json', '-d', test_file],
            capture_output=True,
            text=True
        )
        if result.returncode == 0:
            print(f"✓ {test_file} is valid")
        else:
            print(f"✗ {test_file} validation failed:")
            print(result.stderr)

if __name__ == "__main__":
    combine_schemas()