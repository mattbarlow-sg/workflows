#!/usr/bin/env node
/**
 * AJV v8.17 validation wrapper for BPMN schemas
 * Usage: node ajv-validate.js <schema-file> <data-file>
 */

const fs = require("fs");
const path = require("path");

// Load AJV from global installation
const globalModulePath = "/usr/lib/node_modules";
const Ajv = require(path.join(globalModulePath, "ajv"));

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length < 2) {
  console.error("Usage: node ajv-validate.js <schema-file> <data-file>");
  process.exit(1);
}

const schemaFile = args[0];
const dataFile = args[1];

// Create AJV instance with draft-07 support
const ajv = new Ajv({
  strict: false,
  loadSchema: loadSchema,
  addUsedSchema: true,
  inlineRefs: false
});

// Try to load format validators if available
try {
  const addFormats = require(path.join(globalModulePath, "ajv-formats"));
  addFormats(ajv);
} catch (e) {
  // ajv-formats not installed, formats will be ignored
}

// Custom schema loader for local file references
async function loadSchema(uri) {
  // Handle file: URIs
  if (uri.startsWith("file:")) {
    const filename = uri.replace("file:", "");
    const filepath = path.join(path.dirname(schemaFile), filename);
    try {
      const content = fs.readFileSync(filepath, "utf8");
      return JSON.parse(content);
    } catch (err) {
      throw new Error(`Failed to load schema ${uri}: ${err.message}`);
    }
  }
  
  // Handle https URLs by converting to local files
  if (uri.startsWith("https://workflows.example.com/schemas/")) {
    const filename = uri.replace("https://workflows.example.com/schemas/", "");
    const filepath = path.join("schemas", filename);
    try {
      const content = fs.readFileSync(filepath, "utf8");
      return JSON.parse(content);
    } catch (err) {
      throw new Error(`Failed to load schema ${uri}: ${err.message}`);
    }
  }
  
  throw new Error(`Unsupported schema URI: ${uri}`);
}

// Main validation function
async function validate() {
  try {
    // Load main schema
    const schemaContent = fs.readFileSync(schemaFile, "utf8");
    const schema = JSON.parse(schemaContent);
    
    // Load all BPMN schemas for references
    const schemaDir = "schemas";
    const bpmnSchemas = fs.readdirSync(schemaDir)
      .filter(f => f.startsWith("bpmn-") && f.endsWith(".json"))
      .map(f => {
        const content = fs.readFileSync(path.join(schemaDir, f), "utf8");
        return JSON.parse(content);
      });
    
    // Add all schemas to AJV except the main one
    for (const bpmnSchema of bpmnSchemas) {
      if (bpmnSchema.$id && bpmnSchema.$id !== schema.$id) {
        try {
          ajv.addSchema(bpmnSchema);
        } catch (err) {
          // Ignore if schema already added
          if (!err.message.includes("already exists")) {
            throw err;
          }
        }
      }
    }
    
    // Compile the main schema
    const validate = await ajv.compileAsync(schema);
    
    // Load and validate data
    const dataContent = fs.readFileSync(dataFile, "utf8");
    const data = JSON.parse(dataContent);
    
    const valid = validate(data);
    
    if (valid) {
      console.log(`✓ ${dataFile} is valid according to ${schemaFile}`);
      process.exit(0);
    } else {
      console.log(`✗ ${dataFile} is invalid according to ${schemaFile}`);
      console.log("\nValidation errors:");
      validate.errors.forEach((err, i) => {
        const path = err.instancePath || "/";
        const message = err.message;
        console.log(`  ${i + 1}. ${path}: ${message}`);
        if (err.params) {
          console.log(`     Details:`, err.params);
        }
      });
      process.exit(1);
    }
    
  } catch (err) {
    console.error(`Error: ${err.message}`);
    if (err.stack) {
      console.error(err.stack);
    }
    process.exit(1);
  }
}

// Run validation
validate();