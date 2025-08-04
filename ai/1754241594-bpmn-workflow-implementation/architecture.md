# BPMN 2.0 Workflow Architecture

## System Overview

The BPMN 2.0 workflow system extends the existing JSON schema
validation framework to support Business Process Model and Notation
diagrams. The system focuses on JSON representation (not XML) and
provides comprehensive validation including semantic correctness
checks.

```
┌─────────────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   AI Command        │     │    CLI Tool      │     │  JSON Schemas       │
│ ai-bpmn-create.md   │────▶│ workflows bpmn   │────▶│ bpmn-*.json         │
└─────────────────────┘     └──────────────────┘     └─────────────────────┘
         │                           │                         │
         │                           ▼                         ▼
         │                  ┌──────────────────┐     ┌─────────────────────┐
         │                  │   Go Packages    │     │ Validation Engine   │
         └─────────────────▶│  internal/bpmn/  │────▶│ + Semantic Checks   │
                            └──────────────────┘     └─────────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │  Output Files    │
                            │ .json & reports  │
                            └──────────────────┘
```

## BPMN 2.0 Core Elements

### 1. Flow Objects

#### Events
- **Start Event**: Triggers process initiation
- **End Event**: Marks process completion
- **Intermediate Event**: Occurs during process execution
- **Boundary Event**: Attached to activities for exception handling

#### Activities  
- **Task**: Atomic unit of work
  - User Task: Human interaction required
  - Service Task: System/automated execution
  - Script Task: Predefined script execution
  - Send/Receive Task: Message handling
  - **Manual Task**: General human task (agent-agnostic)
  - **Business Rule Task**: Decision logic execution
- **Sub-Process**: Encapsulated process with internal flow
- **Call Activity**: References external process

#### Gateways
- **Exclusive (XOR)**: Single path selection
- **Parallel (AND)**: Fork/join concurrent paths
- **Inclusive (OR)**: Multiple path selection
- **Event-based**: Path selection based on events
- **Complex**: Custom conditions

### 2. Connecting Objects
- **Sequence Flow**: Ordered activity execution
- **Message Flow**: Inter-process communication
- **Association**: Data/annotation links

### 3. Swimlanes
- **Pool**: Process participant/entity
- **Lane**: Role/responsibility within pool

### 4. Artifacts
- **Data Object**: Information carrier
- **Data Store**: Persistent data repository
- **Group**: Visual element grouping
- **Text Annotation**: Documentation

## JSON Schema Design

### Schema Hierarchy
```
schemas/
├── bpmn-process.json       # Root process definition
├── bpmn-common.json        # Shared definitions
├── bpmn-flow-objects.json  # Events, activities, gateways
├── bpmn-connectors.json    # Flows and associations
└── bpmn-artifacts.json     # Data objects, annotations
```

### Core Schema Structure
```json
{
  "process": {
    "id": "string",
    "name": "string",
    "elements": {
      "events": [],
      "activities": [],
      "gateways": [],
      "flows": [],
      "dataObjects": []
    },
    "swimlanes": {
      "pools": [],
      "lanes": []
    },
    "agents": {
      "definitions": [],
      "assignments": []
    }
  }
}
```

### Agent and Review Model
```json
{
  "activity": {
    "id": "task_001",
    "type": "userTask",
    "name": "Review Document",
    "agent": {
      "type": "dynamic|ai|human|system",
      "assignment": {
        "strategy": "runtime|design-time",
        "preferredAgent": "ai|human",
        "constraints": ["requires-domain-knowledge", "requires-approval"]
      },
      "review": {
        "required": true,
        "reviewer": "human|ai|role",
        "strategy": "automatic|manual|conditional"
      }
    }
  }
}
```

## Component Architecture

### 1. Schema Layer (`schemas/bpmn-*.json`)
- Modular design with referenced sub-schemas
- JSON Schema Draft-07 specification
- Comprehensive validation rules
- Element relationship constraints

### 2. Go Implementation (`internal/bpmn/`)

#### Core Packages
- **types.go**: BPMN element struct definitions
- **builder.go**: Process construction helpers
- **validator.go**: Semantic validation engine
- **analyzer.go**: Process analysis tools

#### Validation Levels
1. **Schema Validation**: Structure and data types
2. **Semantic Validation**: BPMN rule compliance
3. **Graph Analysis**: Connectivity and reachability
4. **Business Logic**: Custom rule application

### 3. CLI Extension (`cmd/workflows/bpmn.go`)

#### Commands
- `bpmn new`: Interactive process builder
- `bpmn validate`: Multi-level validation
- `bpmn analyze`: Process metrics and issues
- `bpmn convert`: Format transformations

#### Features
- Comprehensive flag support
- Detailed error reporting
- Analysis reports
- Multiple output formats

### 4. AI Workflow (`.claude/commands/ai-bpmn-create.md`)

#### Process Flow
1. **Requirements Gathering**
   - Process goals and participants
   - Key activities and decisions
   - Data requirements

2. **Process Design**
   - Activity identification
   - Flow logic definition
   - Gateway placement
   - Exception handling

3. **Validation & Refinement**
   - Semantic correctness
   - Completeness checks
   - Optimization suggestions

4. **Output Generation**
   - JSON process definition
   - Validation report
   - Process documentation

## Agent Assignment and Review Workflows

### Agent Types and Capabilities

#### 1. Agent Definitions
- **AI Agent**: Automated task execution with defined capabilities
- **Human Agent**: Manual task execution requiring human judgment
- **System Agent**: Pure automated/programmatic execution
- **Dynamic Agent**: Runtime determination based on context

#### 2. Assignment Strategies
- **Design-time Assignment**: Agent specified during process design
- **Runtime Assignment**: Agent determined by rules/conditions
- **Capability-based Assignment**: Match task requirements to agent capabilities
- **Load-balanced Assignment**: Distribute work across available agents

#### 3. Review Workflows
```json
{
  "reviewPattern": {
    "ai-human": {
      "description": "AI performs, human reviews",
      "steps": [
        {"agent": "ai", "action": "execute"},
        {"agent": "human", "action": "review"},
        {"agent": "human", "action": "approve|reject|revise"}
      ]
    },
    "human-ai": {
      "description": "Human performs, AI validates",
      "steps": [
        {"agent": "human", "action": "execute"},
        {"agent": "ai", "action": "validate"},
        {"agent": "ai", "action": "flag-issues|pass"}
      ]
    },
    "collaborative": {
      "description": "Iterative collaboration",
      "steps": [
        {"agent": "dynamic", "action": "draft"},
        {"agent": "dynamic", "action": "review"},
        {"agent": "dynamic", "action": "finalize"}
      ]
    }
  }
}
```

### Agent Assignment Schema Extension
```json
{
  "agentDefinition": {
    "id": "agent_001",
    "type": "ai|human|system",
    "capabilities": ["text-analysis", "code-generation", "decision-making"],
    "constraints": ["max-complexity", "domain-restrictions"],
    "availability": "always|scheduled|on-demand"
  },
  
  "taskAssignment": {
    "taskId": "task_001",
    "assignmentStrategy": {
      "type": "static|dynamic|rule-based",
      "rules": [
        {
          "condition": "task.complexity > threshold",
          "assignTo": "human"
        },
        {
          "condition": "task.type == 'code-review'",
          "assignTo": "ai",
          "withReview": "human"
        }
      ]
    }
  }
}
```

### Review Process Integration
1. **Automatic Review Triggers**
   - AI task completion → Human review
   - Confidence threshold checks
   - Error/exception conditions
   
2. **Review Actions**
   - Approve: Continue to next task
   - Reject: Return to performer
   - Revise: Modify and resubmit
   - Escalate: Route to supervisor

3. **Review Tracking**
   - Review history and audit trail
   - Performance metrics
   - Learning feedback loop

## Validation Beyond Schema

### 1. Structural Validation
- All start events have outgoing flows
- All end events have incoming flows
- No orphaned elements
- Proper gateway pairing (split/join)

### 2. Semantic Validation
- Exclusive gateways have conditions
- Parallel gateways properly synchronized
- Message flows cross pool boundaries
- Data associations properly typed

### 3. Graph Analysis
- **Reachability**: All elements accessible from start
- **Deadlock Detection**: No circular waits
- **Liveness**: Process can complete
- **Soundness**: Proper termination

### 4. Complexity Metrics
- Cyclomatic complexity
- Depth of nesting
- Number of paths
- Cognitive complexity score

## Integration Points

### 1. Existing Workflow System
- Extends current schema registry
- Reuses validation framework
- Compatible CLI structure
- Shared configuration

### 2. External Tools
- Export to BPMN-compliant JSON
- Import from simplified formats
- Integration hooks for engines
- Visualization support

### 3. Extensibility
- Custom validation rules
- Process templates
- Domain-specific elements
- Plugin architecture

## Key Design Decisions

### 1. JSON-First Approach
- **Rationale**: Better web integration, simpler parsing
- **Trade-off**: Not standard BPMN XML format
- **Mitigation**: Conversion utilities provided

### 2. Modular Schema Design
- **Rationale**: Maintainability and reusability
- **Benefit**: Easier extension and validation
- **Implementation**: JSON Schema $ref usage

### 3. Multi-Level Validation
- **Rationale**: Catch different error types
- **Levels**: Schema → Semantic → Graph → Business
- **Benefit**: Comprehensive quality assurance

### 4. AI-Assisted Design
- **Rationale**: Lower barrier to entry
- **Approach**: Conversational process building
- **Benefit**: Best practices enforcement

### 5. Dynamic Agent Assignment
- **Rationale**: Flexibility in task execution
- **Approach**: Runtime assignment with review workflows
- **Benefit**: Optimal agent utilization and quality control

## Known Issues and Solutions

### Schema Reference Resolution
The current schemas use `$ref` with relative paths (e.g., `"$ref": "bpmn-common.json#/definitions/id"`). 
The existing validator tries to resolve these as URLs, which fails. 

**Solution**: In Task 2.2 (Semantic Validator), we'll implement a custom resolver that:
1. Loads all BPMN schemas into memory
2. Resolves references locally without network calls
3. Provides proper error messages for missing references

## Technical Implementation

### Data Structures
```go
type Process struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Elements    ProcessElements     `json:"elements"`
    Swimlanes   Swimlanes          `json:"swimlanes"`
    DataObjects []DataObject       `json:"dataObjects"`
    Agents      AgentConfig        `json:"agents"`
}

type ProcessElements struct {
    Events     []Event    `json:"events"`
    Activities []Activity `json:"activities"`
    Gateways   []Gateway  `json:"gateways"`
    Flows      []Flow     `json:"flows"`
}

type Activity struct {
    ID           string          `json:"id"`
    Type         string          `json:"type"`
    Name         string          `json:"name"`
    Agent        AgentAssignment `json:"agent"`
    ReviewConfig *ReviewConfig   `json:"review,omitempty"`
}

type AgentAssignment struct {
    Type         string   `json:"type"` // "ai", "human", "system", "dynamic"
    Strategy     string   `json:"strategy"`
    Constraints  []string `json:"constraints"`
    PreferredAgent string `json:"preferredAgent,omitempty"`
}

type ReviewConfig struct {
    Required bool   `json:"required"`
    Reviewer string `json:"reviewer"`
    Strategy string `json:"strategy"`
}
```

### Validation Engine
```go
type ValidationResult struct {
    SchemaErrors    []Error
    SemanticErrors  []Error
    GraphErrors     []Error
    Warnings        []Warning
    Metrics         ProcessMetrics
}
```

## Security Considerations
- Input sanitization for all elements
- Path traversal prevention in file operations
- Schema injection protection
- Safe template rendering

## Performance Characteristics
- O(n) schema validation
- O(n²) worst-case graph analysis
- Streaming support for large processes
- Caching for repeated validations

## Future Enhancements
1. Visual process designer integration
2. BPMN 2.0 XML import/export
3. Process simulation capabilities
4. Advanced pattern detection
5. Team collaboration features
