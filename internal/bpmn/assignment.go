package bpmn

import (
	"fmt"
	"strconv"
	"strings"
)

// AssignmentRule defines an interface for custom assignment rules
type AssignmentRule interface {
	// Name returns the name of the rule
	Name() string

	// Applies checks if this rule applies to the given context
	Applies(ctx AssignmentContext) bool

	// SelectAgent selects an agent based on the rule logic
	// Returns the selected agent and a confidence score (0-1)
	SelectAgent(agents []*Agent, ctx AssignmentContext) (*Agent, float64)
}

// AssignmentStrategy defines different strategies for agent assignment
type AssignmentStrategy string

const (
	// StrategyStatic assigns a specific agent
	StrategyStatic AssignmentStrategy = "static"

	// StrategyDynamic assigns based on runtime conditions
	StrategyDynamic AssignmentStrategy = "dynamic"

	// StrategyPool assigns from a pool of agents
	StrategyPool AssignmentStrategy = "pool"

	// StrategyLoadBalance assigns based on workload
	StrategyLoadBalance AssignmentStrategy = "load_balance"

	// StrategyRoundRobin assigns in rotation
	StrategyRoundRobin AssignmentStrategy = "round_robin"

	// StrategyRandom assigns randomly
	StrategyRandom AssignmentStrategy = "random"

	// StrategyCapabilityMatch assigns based on best capability match
	StrategyCapabilityMatch AssignmentStrategy = "capability_match"
)

// LoadBalancingStrategy defines how load is distributed
type LoadBalancingStrategy string

const (
	// LoadBalanceLeastTasks assigns to agent with fewest tasks
	LoadBalanceLeastTasks LoadBalancingStrategy = "least_tasks"

	// LoadBalanceLeastLoad assigns to agent with lowest workload score
	LoadBalanceLeastLoad LoadBalancingStrategy = "least_load"

	// LoadBalanceWeighted assigns based on agent weights
	LoadBalanceWeighted LoadBalancingStrategy = "weighted"

	// LoadBalanceFairShare ensures equal distribution over time
	LoadBalanceFairShare LoadBalancingStrategy = "fair_share"
)

// DynamicAssignmentCondition represents a condition for dynamic assignment
type DynamicAssignmentCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// EvaluateCondition evaluates a dynamic assignment condition
func EvaluateCondition(condition DynamicAssignmentCondition, data map[string]interface{}) bool {
	fieldValue, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "equals", "==":
		return fieldValue == condition.Value
	case "not_equals", "!=":
		return fieldValue != condition.Value
	case "greater_than", ">":
		return compareNumeric(fieldValue, condition.Value, ">")
	case "less_than", "<":
		return compareNumeric(fieldValue, condition.Value, "<")
	case "contains":
		return containsString(toString(fieldValue), toString(condition.Value))
	case "in":
		return isIn(fieldValue, condition.Value)
	default:
		return false
	}
}

// compareNumeric compares two numeric values
func compareNumeric(a, b interface{}, op string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch op {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts interface to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		// Try to parse string as number
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString converts interface to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// isIn checks if value is in a list
func isIn(value, list interface{}) bool {
	switch l := list.(type) {
	case []interface{}:
		for _, item := range l {
			if value == item {
				return true
			}
		}
	case []string:
		valStr := toString(value)
		for _, item := range l {
			if valStr == item {
				return true
			}
		}
	}
	return false
}

// contains checks if string contains substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
