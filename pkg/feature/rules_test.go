package feature

import (
	"testing"
)

func TestClause_Equals(t *testing.T) {
	tests := []struct {
		name      string
		clause    Clause
		ctx       *Context
		want      bool
	}{
		{
			name: "equals string match",
			clause: Clause{
				Attribute: "country",
				Operator:  OpEquals,
				Values:    []interface{}{"US"},
			},
			ctx:  NewContext("user-1").WithCountry("US"),
			want: true,
		},
		{
			name: "equals string no match",
			clause: Clause{
				Attribute: "country",
				Operator:  OpEquals,
				Values:    []interface{}{"US"},
			},
			ctx:  NewContext("user-1").WithCountry("UK"),
			want: false,
		},
		{
			name: "equals multiple values match",
			clause: Clause{
				Attribute: "country",
				Operator:  OpEquals,
				Values:    []interface{}{"US", "UK", "CA"},
			},
			ctx:  NewContext("user-1").WithCountry("UK"),
			want: true,
		},
		{
			name: "equals nil value",
			clause: Clause{
				Attribute: "country",
				Operator:  OpEquals,
				Values:    []interface{}{"US"},
			},
			ctx:  NewContext("user-1"), // country not set
			want: false,
		},
		{
			name: "equals custom attribute match",
			clause: Clause{
				Attribute: "plan",
				Operator:  OpEquals,
				Values:    []interface{}{"pro"},
			},
			ctx:  NewContext("user-1").WithAttribute("plan", "pro"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.clause.evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_NotEquals(t *testing.T) {
	clause := Clause{
		Attribute: "plan",
		Operator:  OpNotEquals,
		Values:    []interface{}{"free"},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"not equals match", "pro", true},
		{"not equals no match", "free", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithAttribute("plan", tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_Contains(t *testing.T) {
	tests := []struct {
		name   string
		clause Clause
		ctx    *Context
		want   bool
	}{
		{
			name: "contains string match",
			clause: Clause{
				Attribute: "email",
				Operator:  OpContains,
				Values:    []interface{}{"@example.com"},
			},
			ctx:  NewContext("user-1").WithEmail("user@example.com"),
			want: true,
		},
		{
			name: "contains string no match",
			clause: Clause{
				Attribute: "email",
				Operator:  OpContains,
				Values:    []interface{}{"@example.com"},
			},
			ctx:  NewContext("user-1").WithEmail("user@other.com"),
			want: false,
		},
		{
			name: "contains slice match",
			clause: Clause{
				Attribute: "groups",
				Operator:  OpContains,
				Values:    []interface{}{"beta-testers"},
			},
			ctx:  NewContext("user-1").WithGroups([]string{"users", "beta-testers", "admins"}),
			want: true,
		},
		{
			name: "contains slice no match",
			clause: Clause{
				Attribute: "groups",
				Operator:  OpContains,
				Values:    []interface{}{"beta-testers"},
			},
			ctx:  NewContext("user-1").WithGroups([]string{"users", "admins"}),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.clause.evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_StartsWith(t *testing.T) {
	clause := Clause{
		Attribute: "email",
		Operator:  OpStartsWith,
		Values:    []interface{}{"admin"},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"startsWith match", "admin@example.com", true},
		{"startsWith no match", "user@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithEmail(tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_EndsWith(t *testing.T) {
	clause := Clause{
		Attribute: "email",
		Operator:  OpEndsWith,
		Values:    []interface{}{"@company.com"},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"endsWith match", "user@company.com", true},
		{"endsWith no match", "user@other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithEmail(tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_In(t *testing.T) {
	clause := Clause{
		Attribute: "plan",
		Operator:  OpIn,
		Values:    []interface{}{"pro", "enterprise"},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"in match first", "pro", true},
		{"in match second", "enterprise", true},
		{"in no match", "free", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithAttribute("plan", tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_NotIn(t *testing.T) {
	clause := Clause{
		Attribute: "plan",
		Operator:  OpNotIn,
		Values:    []interface{}{"free", "trial"},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"notIn match", "pro", true},
		{"notIn no match free", "free", false},
		{"notIn no match trial", "trial", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithAttribute("plan", tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_GreaterThan(t *testing.T) {
	clause := Clause{
		Attribute: "age",
		Operator:  OpGreaterThan,
		Values:    []interface{}{18},
	}

	tests := []struct {
		name      string
		attrValue interface{}
		want      bool
	}{
		{"gt match", 21, true},
		{"gt no match equal", 18, false},
		{"gt no match less", 16, false},
		{"gt with float", 20.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithAttribute("age", tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_LessThan(t *testing.T) {
	clause := Clause{
		Attribute: "score",
		Operator:  OpLessThan,
		Values:    []interface{}{100},
	}

	tests := []struct {
		name      string
		attrValue interface{}
		want      bool
	}{
		{"lt match", 50, true},
		{"lt no match equal", 100, false},
		{"lt no match greater", 150, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithAttribute("score", tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_Matches(t *testing.T) {
	clause := Clause{
		Attribute: "email",
		Operator:  OpMatches,
		Values:    []interface{}{`^[a-z]+@example\.com$`},
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"regex match", "user@example.com", true},
		{"regex no match", "User@example.com", false},
		{"regex no match domain", "user@other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithEmail(tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_Negate(t *testing.T) {
	clause := Clause{
		Attribute: "country",
		Operator:  OpEquals,
		Values:    []interface{}{"US"},
		Negate:    true,
	}

	tests := []struct {
		name      string
		attrValue string
		want      bool
	}{
		{"negated equals - was true, now false", "US", false},
		{"negated equals - was false, now true", "UK", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext("user-1").WithCountry(tt.attrValue)
			got := clause.evaluate(ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_Evaluate(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
		ctx  *Context
		want bool
	}{
		{
			name: "empty clauses always match",
			rule: Rule{
				ID:        "rule-1",
				Variation: 1,
			},
			ctx:  NewContext("user-1"),
			want: true,
		},
		{
			name: "single clause match",
			rule: Rule{
				ID: "rule-1",
				Clauses: []Clause{
					{
						Attribute: "plan",
						Operator:  OpEquals,
						Values:    []interface{}{"pro"},
					},
				},
				Variation: 1,
			},
			ctx:  NewContext("user-1").WithAttribute("plan", "pro"),
			want: true,
		},
		{
			name: "all clauses must match - all pass",
			rule: Rule{
				ID: "rule-1",
				Clauses: []Clause{
					{
						Attribute: "plan",
						Operator:  OpEquals,
						Values:    []interface{}{"pro"},
					},
					{
						Attribute: "country",
						Operator:  OpEquals,
						Values:    []interface{}{"US"},
					},
				},
				Variation: 1,
			},
			ctx:  NewContext("user-1").WithAttribute("plan", "pro").WithCountry("US"),
			want: true,
		},
		{
			name: "all clauses must match - one fails",
			rule: Rule{
				ID: "rule-1",
				Clauses: []Clause{
					{
						Attribute: "plan",
						Operator:  OpEquals,
						Values:    []interface{}{"pro"},
					},
					{
						Attribute: "country",
						Operator:  OpEquals,
						Values:    []interface{}{"US"},
					},
				},
				Variation: 1,
			},
			ctx:  NewContext("user-1").WithAttribute("plan", "pro").WithCountry("UK"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		want   float64
		wantOk bool
	}{
		{"int", 42, 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"float32", float32(42.5), 42.5, true},
		{"float64", float64(42.5), 42.5, true},
		{"string", "42", 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.value)
			if ok != tt.wantOk {
				t.Errorf("toFloat64() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("toFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}
