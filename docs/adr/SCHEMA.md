# ADR Schema Documentation

This document describes the JSON schema for Architecture Decision Records (ADRs).

## Schema Overview

The ADR schema is based on MADR 4.0.0 (Markdown Architectural Decision Records) with extensions for AI-friendly metadata and enhanced searchability.

## Field Reference

### Core Fields (Required)

#### `id` (string, required)
- **Pattern**: `^ADR-[0-9]{4}$`
- **Example**: `"ADR-0001"`
- **Description**: Unique identifier for the ADR

#### `title` (string, required)
- **Length**: 5-100 characters
- **Example**: `"Use React for frontend framework"`
- **Description**: Brief, descriptive title of the decision

#### `status` (string, required)
- **Enum**: `draft`, `proposed`, `accepted`, `deprecated`, `superseded`, `rejected`
- **Example**: `"accepted"`
- **Description**: Current status of the ADR

#### `date` (string, required)
- **Format**: ISO 8601 date (`YYYY-MM-DD`)
- **Example**: `"2024-03-15"`
- **Description**: Date when the decision was made

#### `context` (object, required)
- **Required properties**:
  - `problem` (string, min 10 chars): The problem being addressed
  - `background` (string): Current state and why change is needed
- **Optional properties**:
  - `constraints` (array of strings): Technical or business constraints
  - `assumptions` (array of strings): Assumptions made

#### `decision` (object, required)
- **Required properties**:
  - `chosenOption` (string): Name of the chosen option
  - `rationale` (string): Why this option was chosen
- **Optional properties**:
  - `implementation` (string): Implementation approach

#### `consequences` (object, required)
- **Properties**:
  - `positive` (array of strings): Positive consequences
  - `negative` (array of strings): Negative consequences
  - `neutral` (array of strings, optional): Neutral consequences

### Stakeholder Fields (Optional)

#### `stakeholders` (object)
- **Properties**:
  - `deciders` (array of strings): People making the decision
  - `consulted` (array of strings): People consulted
  - `informed` (array of strings): People to be informed

### Decision Analysis Fields (Optional)

#### `decisionDrivers` (array of objects)
Each driver object contains:
- `driver` (string, required): Name of the criterion
- `weight` (number, required): Importance (1-5, where 5 is most important)
- `description` (string, optional): Explanation of the driver

#### `options` (array of objects)
Each option object contains:
- `name` (string, required): Option name
- `description` (string, required): Option description
- `pros` (array of strings): Advantages
- `cons` (array of strings): Disadvantages
- `score` (number, 0-100): Weighted score based on decision drivers

### Technical Context (Optional)

#### `technicalStory` (object)
- **Properties**:
  - `id` (string): Story/ticket identifier (e.g., "JIRA-123")
  - `title` (string): Story title
  - `description` (string): Detailed description
  - `acceptanceCriteria` (array of strings): Success criteria
  - `link` (string, uri format): External link

### Validation and Metrics (Optional)

#### `validation` (object)
- **Properties**:
  - `method` (string): How to validate the decision
  - `successCriteria` (array of strings): What success looks like
  - `metrics` (array of objects): Measurable outcomes
    - `metric` (string, required): What to measure
    - `target` (string, required): Target value
    - `timeframe` (string): When to measure

### Compliance (Optional)

#### `compliance` (object)
- **Properties**:
  - `standards` (array of strings): Industry standards
  - `regulations` (array of strings): Legal/regulatory requirements

### AI Metadata (Optional)

#### `aiMetadata` (object)
Extensions for AI processing and searchability:
- **Properties**:
  - `tags` (array of strings): Semantic categorization
  - `keywords` (array of strings): Search terms
  - `impactScores` (object):
    - `technical` (number, 1-10): Technical impact
    - `business` (number, 1-10): Business impact
    - `risk` (number, 1-10): Risk level
  - `dependencies` (array of objects): Related ADRs
    - `adrId` (string, pattern `^ADR-[0-9]{4}$`): ADR identifier
    - `relationship` (enum): Type of relationship
      - `depends-on`
      - `supersedes`
      - `superseded-by`
      - `relates-to`
      - `conflicts-with`
  - `estimatedCost` (object):
    - `development` (string): Development cost/effort
    - `maintenance` (string): Ongoing cost/effort

### Additional Fields

#### `notes` (string, optional)
Implementation notes or additional context

#### `links` (array of objects, optional)
External references:
- `title` (string, required): Link description
- `url` (string, required, uri format): URL
- `type` (enum): Link type
  - `documentation`
  - `example`
  - `article`
  - `tool`
  - `standard`

## Validation Rules

1. **ID Format**: Must match pattern `ADR-XXXX` where X is a digit
2. **Title Length**: Between 5 and 100 characters
3. **Problem Statement**: Minimum 10 characters
4. **Status Values**: Must be one of the defined enum values
5. **Date Format**: Must be valid ISO 8601 date
6. **Decision Drivers Weight**: Between 1 and 5
7. **Impact Scores**: Between 1 and 10
8. **URIs**: Must be valid URI format
9. **Option Scores**: Between 0 and 100

## Example ADR

```json
{
  "id": "ADR-0001",
  "title": "Use React for frontend framework",
  "status": "accepted",
  "date": "2024-03-15",
  "context": {
    "problem": "Need to select a frontend framework for new web application",
    "background": "Building a new customer-facing application from scratch",
    "constraints": ["Must support IE11", "Team has limited frontend experience"],
    "assumptions": ["Project timeline allows for learning curve"]
  },
  "decisionDrivers": [
    {
      "driver": "Developer Experience",
      "weight": 5,
      "description": "How easy it is to develop with"
    },
    {
      "driver": "Performance",
      "weight": 3,
      "description": "Runtime performance"
    }
  ],
  "options": [
    {
      "name": "React",
      "description": "Facebook's component library",
      "pros": ["Large ecosystem", "Good documentation"],
      "cons": ["Just a library, not a framework"],
      "score": 85
    },
    {
      "name": "Angular",
      "description": "Google's full framework",
      "pros": ["Complete solution", "TypeScript native"],
      "cons": ["Steep learning curve", "Heavy"],
      "score": 65
    }
  ],
  "decision": {
    "chosenOption": "React",
    "rationale": "Best balance of features and learning curve",
    "implementation": "Use Create React App for initial setup"
  },
  "consequences": {
    "positive": ["Fast development", "Easy to hire developers"],
    "negative": ["Need to choose additional libraries"],
    "neutral": ["Requires build toolchain"]
  },
  "stakeholders": {
    "deciders": ["tech-lead", "frontend-architect"],
    "consulted": ["frontend-team", "ux-team"],
    "informed": ["backend-team", "product-manager"]
  },
  "aiMetadata": {
    "tags": ["frontend", "framework", "react"],
    "keywords": ["react", "spa", "frontend", "javascript"],
    "impactScores": {
      "technical": 8,
      "business": 6,
      "risk": 3
    }
  }
}
```