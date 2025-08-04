#!/bin/bash
# Test BPMN schemas with AJV v8.17

echo "BPMN Schema Validation with AJV v8.17"
echo "====================================="
echo ""

# Check if node and ajv are available
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed"
    exit 1
fi

# Test each valid file
echo "Testing valid BPMN files:"
echo "------------------------"
for file in simple-process parallel-gateway exclusive-gateway subprocess ai-human-collab dynamic-assignment; do
    echo -n "Testing $file.json... "
    if node ajv-validate.js schemas/bpmn-process.json test-data/$file.json 2>&1; then
        echo ""  # New line after success message
    else
        echo ""  # New line after error message
    fi
done

echo ""
echo "Testing invalid BPMN file:"
echo "-------------------------"
echo -n "Testing invalid-bpmn.json... "
if node ajv-validate.js schemas/bpmn-process.json test-data/invalid-bpmn.json 2>&1; then
    echo "✗ UNEXPECTED: File passed validation but should have failed"
else
    echo "✓ EXPECTED: File failed validation as expected"
fi

echo ""
echo "Summary:"
echo "--------"
echo "Run 'node ajv-validate.js <schema> <data>' to validate individual files"
echo "All schemas are loaded automatically for reference resolution"