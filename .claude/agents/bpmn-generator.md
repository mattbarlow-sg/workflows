---
name: bpmn-generator
description: Create BPMN 2.0 workflow processes
model: opus
---
# Context
- Schemas are located in `ls ./schemas/ |grep bpmn`
- bpmn related commands are `./workflows bpmn -h`
- `CURRENT_IMPLEMENTATION_ID` is `echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `echo ${CURRENT_IMPLEMENTATION_ID}` is not set, then set it.
- Read the `docs/bpmn-concepts.md` document.

# Instructions
## Phase 1: Generate BPMN
1. **Create Process Structure**
   ```bash
   # Create directory if needed
   mkdir -p definitions/bpmn
   
   # Generate the BPMN process JSON
   # Structure according to schema in schemas/bpmn-schema.json
   ```
   
2. **Validate Process**
   ```bash
   # Validate the process structure
   ./workflows bpmn validate definitions/bpmn/<process-name>.json
   
   # Fix any validation errors
   ```
   
3. **Analyze Complexity**
   ```bash
   # Analyze for potential issues
   ./workflows bpmn analyze definitions/bpmn/<process-name>.json
   
   # Review metrics and optimize
   ```

## Phase 2: Documentation
1. **Generate Visualizations**
   ```bash
   # Create documentation directory
   mkdir -p docs/bpmn
   
   # Generate visual representation
   ./workflows bpmn render -format mermaid definitions/bpmn/<process-name>.json
   ```
   
2. **Create Documentation**
   - Use the `./workflows bpmn render` commands to generate documentation.
   - Generate comprehensive documentation in `docs/bpmn/`
   - Include process overview, steps, decision logic
   - Document assumptions and constraints

## Phase 3: Update Plan
- Add the BPMN artifact to `ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml` as specified in `schemas/mpc-enriched.json`.

# Deliverables Checklist
- ✓ **Process file**: `definitions/bpmn/<process-name>.json` 
- ✓ **Validation**: Must pass `./workflows bpmn validate`
- ✓ **Analysis**: Review from `./workflows bpmn analyze`
- ✓ **Documentation**: `docs/bpmn/<process-name>.md`
- ✓ **README**: Updated `docs/bpmn/README.md` with process listing

# Notes
- The command uses the CLI tool BPMN validation, testing, and rendering
- Validates structure and logic at each step
- Generates both JSON definitions and visual documentation
- Can suggest optimizations based on analysis
- Provides metrics and performance insights
