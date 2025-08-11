package bpmn

import (
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition DynamicAssignmentCondition
		data      map[string]interface{}
		want      bool
	}{
		{
			name: "Simple equals",
			condition: DynamicAssignmentCondition{
				Field:    "status",
				Operator: "equals",
				Value:    "active",
			},
			data: map[string]interface{}{
				"status": "active",
			},
			want: true,
		},
		{
			name: "Not equals",
			condition: DynamicAssignmentCondition{
				Field:    "status",
				Operator: "not_equals",
				Value:    "inactive",
			},
			data: map[string]interface{}{
				"status": "active",
			},
			want: true,
		},
		{
			name: "Greater than numeric",
			condition: DynamicAssignmentCondition{
				Field:    "priority",
				Operator: "greater_than",
				Value:    5,
			},
			data: map[string]interface{}{
				"priority": 10,
			},
			want: true,
		},
		{
			name: "Less than numeric",
			condition: DynamicAssignmentCondition{
				Field:    "count",
				Operator: "less_than",
				Value:    100,
			},
			data: map[string]interface{}{
				"count": 50,
			},
			want: true,
		},
		{
			name: "Contains string",
			condition: DynamicAssignmentCondition{
				Field:    "description",
				Operator: "contains",
				Value:    "urgent",
			},
			data: map[string]interface{}{
				"description": "This is an urgent request",
			},
			want: true,
		},
		{
			name: "In list",
			condition: DynamicAssignmentCondition{
				Field:    "category",
				Operator: "in",
				Value:    []string{"A", "B", "C"},
			},
			data: map[string]interface{}{
				"category": "B",
			},
			want: true,
		},
		{
			name: "Field not found",
			condition: DynamicAssignmentCondition{
				Field:    "missing",
				Operator: "equals",
				Value:    "value",
			},
			data: map[string]interface{}{
				"other": "data",
			},
			want: false,
		},
		{
			name: "Invalid operator",
			condition: DynamicAssignmentCondition{
				Field:    "field",
				Operator: "invalid_op",
				Value:    "value",
			},
			data: map[string]interface{}{
				"field": "value",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateCondition(tt.condition, tt.data)
			if got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name string
		a    interface{}
		b    interface{}
		op   string
		want bool
	}{
		{"int greater", 10, 5, ">", true},
		{"float less", 3.14, 5.0, "<", true},
		{"equal floats", 5.0, 5.0, ">=", true},
		{"string numbers", "10", "5", ">", true},
		{"mixed types", 10, "5", ">", true},
		{"invalid string", "abc", 5, ">", false},
		{"invalid op", 10, 5, "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareNumeric(tt.a, tt.b, tt.op)
			if got != tt.want {
				t.Errorf("compareNumeric(%v, %v, %s) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.input)
			if got != tt.want {
				t.Errorf("toString(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsIn(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		list  interface{}
		want  bool
	}{
		{"string in slice", "b", []string{"a", "b", "c"}, true},
		{"string not in slice", "d", []string{"a", "b", "c"}, false},
		{"int in interface slice", 2, []interface{}{1, 2, 3}, true},
		{"string in interface slice", "b", []interface{}{"a", "b", "c"}, true},
		{"not a slice", "a", "abc", false},
		{"nil list", "a", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIn(tt.value, tt.list)
			if got != tt.want {
				t.Errorf("isIn(%v, %v) = %v, want %v", tt.value, tt.list, got, tt.want)
			}
		})
	}
}
