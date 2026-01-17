package feature

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule defines targeting rules.
type Rule struct {
	// ID uniquely identifies this rule.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Clauses define conditions that must all be true.
	Clauses []Clause `json:"clauses,omitempty" yaml:"clauses,omitempty"`

	// Variation is the variant index to return when rule matches.
	Variation int `json:"variation" yaml:"variation"`

	// Percentage limits the rule to a percentage of users (0-100).
	Percentage *int `json:"percentage,omitempty" yaml:"percentage,omitempty"`

	// Rollout defines weighted variant distribution.
	Rollout *Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty"`
}

// Clause defines a condition.
type Clause struct {
	// Attribute is the context attribute to check.
	Attribute string `json:"attribute" yaml:"attribute"`

	// Operator is the comparison operator.
	Operator Operator `json:"operator" yaml:"operator"`

	// Values are the values to compare against.
	Values []interface{} `json:"values" yaml:"values"`

	// Negate inverts the clause result.
	Negate bool `json:"negate,omitempty" yaml:"negate,omitempty"`
}

// Operator for clause evaluation.
type Operator string

const (
	OpEquals      Operator = "eq"
	OpNotEquals   Operator = "neq"
	OpContains    Operator = "contains"
	OpStartsWith  Operator = "startsWith"
	OpEndsWith    Operator = "endsWith"
	OpIn          Operator = "in"
	OpNotIn       Operator = "notIn"
	OpGreaterThan Operator = "gt"
	OpLessThan    Operator = "lt"
	OpMatches     Operator = "matches"
	OpSemVer      Operator = "semver"
)

// Rollout defines percentage-based rollout.
type Rollout struct {
	// BucketBy is the attribute used for consistent bucketing.
	BucketBy string `json:"bucketBy" yaml:"bucketBy"`

	// Variations defines weighted distribution.
	Variations []WeightedVariation `json:"variations" yaml:"variations"`
}

// WeightedVariation for percentage distribution.
type WeightedVariation struct {
	// Variation is the variant index.
	Variation int `json:"variation" yaml:"variation"`

	// Weight is out of 100000 (0.001% precision).
	Weight int `json:"weight" yaml:"weight"`
}

// evaluate checks if the rule matches the given context.
func (r *Rule) evaluate(ctx *Context) bool {
	if len(r.Clauses) == 0 {
		return true
	}

	for _, clause := range r.Clauses {
		if !clause.evaluate(ctx) {
			return false
		}
	}

	return true
}

// evaluate checks if the clause matches the given context.
func (c *Clause) evaluate(ctx *Context) bool {
	attrValue := ctx.getAttribute(c.Attribute)
	result := c.matchOperator(attrValue)

	if c.Negate {
		return !result
	}
	return result
}

// matchOperator applies the operator to compare attribute value with clause values.
func (c *Clause) matchOperator(attrValue interface{}) bool {
	switch c.Operator {
	case OpEquals:
		return c.matchEquals(attrValue)
	case OpNotEquals:
		return !c.matchEquals(attrValue)
	case OpContains:
		return c.matchContains(attrValue)
	case OpStartsWith:
		return c.matchStartsWith(attrValue)
	case OpEndsWith:
		return c.matchEndsWith(attrValue)
	case OpIn:
		return c.matchIn(attrValue)
	case OpNotIn:
		return !c.matchIn(attrValue)
	case OpGreaterThan:
		return c.matchGreaterThan(attrValue)
	case OpLessThan:
		return c.matchLessThan(attrValue)
	case OpMatches:
		return c.matchRegex(attrValue)
	default:
		return false
	}
}

func (c *Clause) matchEquals(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		if fmt.Sprintf("%v", v) == attrStr {
			return true
		}
	}
	return false
}

func (c *Clause) matchContains(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	// Handle slice/array attributes (e.g., groups)
	if slice, ok := attrValue.([]string); ok {
		for _, v := range c.Values {
			valStr := fmt.Sprintf("%v", v)
			for _, s := range slice {
				if s == valStr {
					return true
				}
			}
		}
		return false
	}

	// Handle string contains
	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		if strings.Contains(attrStr, fmt.Sprintf("%v", v)) {
			return true
		}
	}
	return false
}

func (c *Clause) matchStartsWith(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		if strings.HasPrefix(attrStr, fmt.Sprintf("%v", v)) {
			return true
		}
	}
	return false
}

func (c *Clause) matchEndsWith(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		if strings.HasSuffix(attrStr, fmt.Sprintf("%v", v)) {
			return true
		}
	}
	return false
}

func (c *Clause) matchIn(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		if fmt.Sprintf("%v", v) == attrStr {
			return true
		}
	}
	return false
}

func (c *Clause) matchGreaterThan(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrFloat, ok := toFloat64(attrValue)
	if !ok {
		return false
	}

	for _, v := range c.Values {
		valFloat, ok := toFloat64(v)
		if ok && attrFloat > valFloat {
			return true
		}
	}
	return false
}

func (c *Clause) matchLessThan(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrFloat, ok := toFloat64(attrValue)
	if !ok {
		return false
	}

	for _, v := range c.Values {
		valFloat, ok := toFloat64(v)
		if ok && attrFloat < valFloat {
			return true
		}
	}
	return false
}

func (c *Clause) matchRegex(attrValue interface{}) bool {
	if attrValue == nil || len(c.Values) == 0 {
		return false
	}

	attrStr := fmt.Sprintf("%v", attrValue)
	for _, v := range c.Values {
		pattern := fmt.Sprintf("%v", v)
		matched, err := regexp.MatchString(pattern, attrStr)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// toFloat64 converts a value to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
