package feature

import "time"

// FlagType defines the flag value type.
type FlagType string

const (
	FlagTypeBool   FlagType = "bool"
	FlagTypeString FlagType = "string"
	FlagTypeInt    FlagType = "int"
	FlagTypeFloat  FlagType = "float"
	FlagTypeJSON   FlagType = "json"
)

// Flag defines a feature flag.
type Flag struct {
	// Key is the unique identifier for the flag.
	Key string `json:"key" yaml:"key"`

	// Name is the human-readable name.
	Name string `json:"name" yaml:"name"`

	// Description provides additional context about the flag.
	Description string `json:"description" yaml:"description"`

	// Type defines the value type (bool, string, int, float, json).
	Type FlagType `json:"type" yaml:"type"`

	// DefaultValue is returned when no rules match.
	DefaultValue interface{} `json:"default" yaml:"default"`

	// Enabled controls whether the flag is active.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Rules define targeting conditions.
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`

	// Variants define possible flag values.
	Variants []Variant `json:"variants,omitempty" yaml:"variants,omitempty"`

	// Prerequisites define flags that must match before evaluation.
	Prerequisites []Prerequisite `json:"prerequisites,omitempty" yaml:"prerequisites,omitempty"`

	// CreatedAt is when the flag was created.
	CreatedAt time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`

	// UpdatedAt is when the flag was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// Variant defines a flag variation.
type Variant struct {
	// Name identifies the variant.
	Name string `json:"name" yaml:"name"`

	// Value is the variant's value.
	Value interface{} `json:"value" yaml:"value"`

	// Weight is used for percentage rollouts (out of 100000 for 0.001% precision).
	Weight int `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// Prerequisite defines a flag dependency.
type Prerequisite struct {
	// Key is the prerequisite flag's key.
	Key string `json:"key" yaml:"key"`

	// Variation is the required variation index.
	Variation int `json:"variation" yaml:"variation"`
}

// Evaluation contains flag evaluation result.
type Evaluation struct {
	// Key is the flag key that was evaluated.
	Key string

	// Value is the evaluated result.
	Value interface{}

	// VariationIdx is the index of the variant returned.
	VariationIdx int

	// Reason explains why this value was returned.
	Reason EvaluationReason

	// RuleID is the ID of the matched rule (if any).
	RuleID string

	// InExperiment indicates if this evaluation is part of an experiment.
	InExperiment bool

	// PrerequisiteFailed indicates if a prerequisite check failed.
	PrerequisiteFailed bool
}

// EvaluationReason explains why a value was returned.
type EvaluationReason string

const (
	// ReasonFallthrough indicates the default value was used.
	ReasonFallthrough EvaluationReason = "FALLTHROUGH"

	// ReasonTargetMatch indicates a direct target match.
	ReasonTargetMatch EvaluationReason = "TARGET_MATCH"

	// ReasonRuleMatch indicates a rule matched.
	ReasonRuleMatch EvaluationReason = "RULE_MATCH"

	// ReasonPrerequisite indicates a prerequisite failed.
	ReasonPrerequisite EvaluationReason = "PREREQUISITE_FAILED"

	// ReasonOff indicates the flag is disabled.
	ReasonOff EvaluationReason = "OFF"

	// ReasonError indicates an error occurred.
	ReasonError EvaluationReason = "ERROR"
)

// getVariantValue returns the value at the given variant index, or default if out of range.
func (f *Flag) getVariantValue(idx int) interface{} {
	if idx >= 0 && idx < len(f.Variants) {
		return f.Variants[idx].Value
	}
	return f.DefaultValue
}
