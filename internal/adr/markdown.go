package adr

import (
	"fmt"
	"strings"
)

func (a *ADR) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", a.Title))

	sb.WriteString("## Metadata\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")
	sb.WriteString(fmt.Sprintf("| ID | %s |\n", a.ID))
	sb.WriteString(fmt.Sprintf("| Status | %s |\n", a.Status))
	sb.WriteString(fmt.Sprintf("| Date | %s |\n", a.Date))

	if a.Stakeholders != nil {
		if len(a.Stakeholders.Deciders) > 0 {
			sb.WriteString(fmt.Sprintf("| Deciders | %s |\n", strings.Join(a.Stakeholders.Deciders, ", ")))
		}
		if len(a.Stakeholders.Consulted) > 0 {
			sb.WriteString(fmt.Sprintf("| Consulted | %s |\n", strings.Join(a.Stakeholders.Consulted, ", ")))
		}
		if len(a.Stakeholders.Informed) > 0 {
			sb.WriteString(fmt.Sprintf("| Informed | %s |\n", strings.Join(a.Stakeholders.Informed, ", ")))
		}
	}

	if a.TechnicalStory != nil && a.TechnicalStory.ID != "" {
		sb.WriteString(fmt.Sprintf("| Technical Story | %s |\n", a.TechnicalStory.ID))
	}

	if a.AIMetadata != nil && len(a.AIMetadata.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("| Tags | %s |\n", strings.Join(a.AIMetadata.Tags, ", ")))
	}

	if a.WorkSessionID != "" {
		sb.WriteString(fmt.Sprintf("| Work Session ID | %s |\n", a.WorkSessionID))
	}

	sb.WriteString("\n## Context and Problem Statement\n\n")
	sb.WriteString(fmt.Sprintf("**Problem:** %s\n\n", a.Context.Problem))
	sb.WriteString(fmt.Sprintf("**Background:** %s\n\n", a.Context.Background))

	if len(a.Context.Constraints) > 0 {
		sb.WriteString("**Constraints:**\n")
		for _, constraint := range a.Context.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		sb.WriteString("\n")
	}

	if len(a.Context.Assumptions) > 0 {
		sb.WriteString("**Assumptions:**\n")
		for _, assumption := range a.Context.Assumptions {
			sb.WriteString(fmt.Sprintf("- %s\n", assumption))
		}
		sb.WriteString("\n")
	}

	if len(a.DecisionDrivers) > 0 {
		sb.WriteString("## Decision Drivers\n\n")
		for _, driver := range a.DecisionDrivers {
			sb.WriteString(fmt.Sprintf("- **%s** (weight: %.1f)", driver.Driver, driver.Weight))
			if driver.Description != "" {
				sb.WriteString(fmt.Sprintf(" - %s", driver.Description))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Considered Options\n\n")
	for i, option := range a.Options {
		sb.WriteString(fmt.Sprintf("### Option %d: %s\n\n", i+1, option.Name))
		sb.WriteString(option.Description)
		sb.WriteString("\n\n")

		if len(option.Pros) > 0 {
			sb.WriteString("**Pros:**\n")
			for _, pro := range option.Pros {
				sb.WriteString(fmt.Sprintf("- %s\n", pro))
			}
			sb.WriteString("\n")
		}

		if len(option.Cons) > 0 {
			sb.WriteString("**Cons:**\n")
			for _, con := range option.Cons {
				sb.WriteString(fmt.Sprintf("- %s\n", con))
			}
			sb.WriteString("\n")
		}

		if option.Score != nil {
			sb.WriteString(fmt.Sprintf("**Score:** %.2f\n\n", *option.Score))
		}
	}

	sb.WriteString("## Decision Outcome\n\n")
	sb.WriteString(fmt.Sprintf("**Chosen option:** %s\n\n", a.Decision.ChosenOption))
	sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", a.Decision.Rationale))

	if a.Decision.Implementation != "" {
		sb.WriteString(fmt.Sprintf("**Implementation approach:** %s\n\n", a.Decision.Implementation))
	}

	if a.Validation != nil {
		sb.WriteString("## Validation\n\n")
		if a.Validation.Method != "" {
			sb.WriteString(fmt.Sprintf("**Method:** %s\n\n", a.Validation.Method))
		}
		if len(a.Validation.SuccessCriteria) > 0 {
			sb.WriteString("**Success Criteria:**\n")
			for _, criteria := range a.Validation.SuccessCriteria {
				sb.WriteString(fmt.Sprintf("- %s\n", criteria))
			}
			sb.WriteString("\n")
		}
		if len(a.Validation.Metrics) > 0 {
			sb.WriteString("**Metrics:**\n")
			for _, metric := range a.Validation.Metrics {
				sb.WriteString(fmt.Sprintf("- %s: %s", metric.Metric, metric.Target))
				if metric.Timeframe != "" {
					sb.WriteString(fmt.Sprintf(" (timeframe: %s)", metric.Timeframe))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Consequences\n\n")

	if len(a.Consequences.Positive) > 0 {
		sb.WriteString("### Positive\n\n")
		for _, consequence := range a.Consequences.Positive {
			sb.WriteString(fmt.Sprintf("- %s\n", consequence))
		}
		sb.WriteString("\n")
	}

	if len(a.Consequences.Negative) > 0 {
		sb.WriteString("### Negative\n\n")
		for _, consequence := range a.Consequences.Negative {
			sb.WriteString(fmt.Sprintf("- %s\n", consequence))
		}
		sb.WriteString("\n")
	}

	if len(a.Consequences.Neutral) > 0 {
		sb.WriteString("### Neutral\n\n")
		for _, consequence := range a.Consequences.Neutral {
			sb.WriteString(fmt.Sprintf("- %s\n", consequence))
		}
		sb.WriteString("\n")
	}

	if a.Compliance != nil && (len(a.Compliance.Standards) > 0 || len(a.Compliance.Regulations) > 0) {
		sb.WriteString("## Compliance\n\n")
		if len(a.Compliance.Standards) > 0 {
			sb.WriteString("**Standards:** " + strings.Join(a.Compliance.Standards, ", ") + "\n\n")
		}
		if len(a.Compliance.Regulations) > 0 {
			sb.WriteString("**Regulations:** " + strings.Join(a.Compliance.Regulations, ", ") + "\n\n")
		}
	}

	if a.Notes != nil && *a.Notes != "" {
		sb.WriteString("## Notes\n\n")
		sb.WriteString(*a.Notes)
		sb.WriteString("\n\n")
	}

	if len(a.Links) > 0 {
		sb.WriteString("## Links\n\n")
		for _, link := range a.Links {
			linkType := ""
			if link.Type != "" {
				linkType = fmt.Sprintf(" (%s)", link.Type)
			}
			sb.WriteString(fmt.Sprintf("- [%s](%s)%s\n", link.Title, link.URL, linkType))
		}
		sb.WriteString("\n")
	}

	if a.AIMetadata != nil && len(a.AIMetadata.Dependencies) > 0 {
		sb.WriteString("## Related ADRs\n\n")
		sb.WriteString("```mermaid\ngraph TD\n")

		for _, dep := range a.AIMetadata.Dependencies {
			arrow := "-->"
			label := ""
			switch dep.Relationship {
			case "supersedes":
				label = "|supersedes|"
			case "superseded-by":
				label = "|superseded by|"
			case "depends-on":
				label = "|depends on|"
			case "conflicts-with":
				arrow = "-.->|conflicts|"
			case "relates-to":
				arrow = "-.->|relates|"
			}

			if dep.Relationship == "superseded-by" {
				sb.WriteString(fmt.Sprintf("    %s %s %s %s\n", dep.ADRID, arrow, label, a.ID))
			} else {
				sb.WriteString(fmt.Sprintf("    %s %s %s %s\n", a.ID, arrow, label, dep.ADRID))
			}
		}

		sb.WriteString("```\n")
	}

	return sb.String()
}
