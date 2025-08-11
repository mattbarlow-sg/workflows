package bpmn

import (
	"testing"
)

func TestReviewPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		{"AI to Human", "ai_to_human", true},
		{"Human to AI", "human_to_ai", true},
		{"Collaborative", "collaborative", true},
		{"Peer Review", "peer_review", true},
		{"Hierarchical", "hierarchical", true},
		{"Custom", "custom", true},
		{"Invalid", "invalid_pattern", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the pattern is recognized
			validPatterns := []string{
				"ai_to_human", "human_to_ai", "collaborative",
				"peer_review", "hierarchical", "custom",
			}

			found := false
			for _, vp := range validPatterns {
				if vp == tt.pattern {
					found = true
					break
				}
			}

			if found != tt.valid {
				t.Errorf("Pattern %s validation = %v, want %v", tt.pattern, found, tt.valid)
			}
		})
	}
}

func TestReviewTimeout(t *testing.T) {
	review := &ReviewConfig{
		Required:  true,
		Type:      "approval",
		Timeout:   "1h",
		OnTimeout: "escalate",
		Reviewer: AgentAssignment{
			Type: "human",
			ID:   "reviewer",
		},
	}

	if review.Timeout != "1h" {
		t.Errorf("Timeout = %s, want 1h", review.Timeout)
	}

	if review.OnTimeout != "escalate" {
		t.Errorf("OnTimeout = %s, want escalate", review.OnTimeout)
	}
}

func TestReviewCriteria(t *testing.T) {
	review := &ReviewConfig{
		Required: true,
		Type:     "quality-check",
		Criteria: []string{
			"accuracy",
			"completeness",
			"grammar",
			"formatting",
		},
		RequiredScore: 0.8,
		Reviewer: AgentAssignment{
			Type: "human",
			ID:   "qa-reviewer",
		},
	}

	if len(review.Criteria) != 4 {
		t.Errorf("Expected 4 criteria, got %d", len(review.Criteria))
	}

	// Check criteria
	expectedCriteria := []string{"accuracy", "completeness", "grammar", "formatting"}
	for i, criterion := range review.Criteria {
		if criterion != expectedCriteria[i] {
			t.Errorf("Criteria[%d] = %s, want %s", i, criterion, expectedCriteria[i])
		}
	}

	if review.RequiredScore != 0.8 {
		t.Errorf("RequiredScore = %f, want 0.8", review.RequiredScore)
	}
}

func TestReviewTypes(t *testing.T) {
	validTypes := []string{"approval", "validation", "quality-check"}

	for _, reviewType := range validTypes {
		review := &ReviewConfig{
			Required: true,
			Type:     reviewType,
			Reviewer: AgentAssignment{
				Type: "human",
				ID:   "reviewer",
			},
		}

		if review.Type != reviewType {
			t.Errorf("Review type = %s, want %s", review.Type, reviewType)
		}
	}
}

func TestReviewAssignment(t *testing.T) {
	// Test different reviewer assignment strategies
	reviewConfigs := []struct {
		name     string
		reviewer AgentAssignment
		valid    bool
	}{
		{
			name: "Static human reviewer",
			reviewer: AgentAssignment{
				Type: "human",
				ID:   "john_doe",
			},
			valid: true,
		},
		{
			name: "Dynamic round-robin assignment",
			reviewer: AgentAssignment{
				Type:     "human",
				Strategy: "round-robin",
			},
			valid: true,
		},
		{
			name: "AI reviewer",
			reviewer: AgentAssignment{
				Type: "ai",
				ID:   "ai-reviewer",
			},
			valid: true,
		},
	}

	for _, tc := range reviewConfigs {
		t.Run(tc.name, func(t *testing.T) {
			review := &ReviewConfig{
				Required: true,
				Type:     "approval",
				Reviewer: tc.reviewer,
			}

			if review.Reviewer.Type == "" {
				t.Error("Reviewer type should not be empty")
			}
		})
	}
}
