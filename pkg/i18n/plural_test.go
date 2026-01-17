package i18n

import "testing"

func TestPluralRules(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		count    int
		expected PluralCategory
	}{
		// English
		{"en singular", "en", 1, One},
		{"en plural", "en", 0, Other},
		{"en plural 2", "en", 2, Other},
		{"en plural 5", "en", 5, Other},

		// French (0 and 1 are singular)
		{"fr zero", "fr", 0, One},
		{"fr singular", "fr", 1, One},
		{"fr plural", "fr", 2, Other},

		// Russian (complex)
		{"ru 1", "ru", 1, One},
		{"ru 2", "ru", 2, Few},
		{"ru 3", "ru", 3, Few},
		{"ru 4", "ru", 4, Few},
		{"ru 5", "ru", 5, Many},
		{"ru 11", "ru", 11, Many},
		{"ru 21", "ru", 21, One},
		{"ru 22", "ru", 22, Few},
		{"ru 25", "ru", 25, Many},

		// Polish
		{"pl 1", "pl", 1, One},
		{"pl 2", "pl", 2, Few},
		{"pl 5", "pl", 5, Many},
		{"pl 12", "pl", 12, Many},
		{"pl 22", "pl", 22, Few},

		// Arabic
		{"ar 0", "ar", 0, Zero},
		{"ar 1", "ar", 1, One},
		{"ar 2", "ar", 2, Two},
		{"ar 3", "ar", 3, Few},
		{"ar 10", "ar", 10, Few},
		{"ar 11", "ar", 11, Many},
		{"ar 99", "ar", 99, Many},
		{"ar 100", "ar", 100, Other},

		// Japanese (no plurals)
		{"ja any", "ja", 1, Other},
		{"ja any 5", "ja", 5, Other},

		// Chinese (no plurals)
		{"zh any", "zh", 1, Other},
		{"zh any 5", "zh", 5, Other},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := GetPluralRule(tt.lang)
			got := rule(tt.count)
			if got != tt.expected {
				t.Errorf("GetPluralRule(%q)(%d) = %v, want %v", tt.lang, tt.count, got, tt.expected)
			}
		})
	}
}

func TestPluralCategory_String(t *testing.T) {
	tests := []struct {
		category PluralCategory
		want     string
	}{
		{Zero, "zero"},
		{One, "one"},
		{Two, "two"},
		{Few, "few"},
		{Many, "many"},
		{Other, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("PluralCategory.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPluralRule_Fallback(t *testing.T) {
	// Test that unknown languages fall back to "other" rule
	rule := GetPluralRule("unknown")
	if rule(1) != Other {
		t.Error("Unknown language should default to Other for all counts")
	}
	if rule(5) != Other {
		t.Error("Unknown language should default to Other for all counts")
	}

	// Test that language-region falls back to language
	rule = GetPluralRule("en-US")
	if rule(1) != One {
		t.Error("en-US should use English rules (1 = One)")
	}
	if rule(5) != Other {
		t.Error("en-US should use English rules (5 = Other)")
	}
}

func TestRegisterPluralRule(t *testing.T) {
	// Register a custom rule for a test language
	customRule := func(n int) PluralCategory {
		if n == 42 {
			return Few
		}
		return Other
	}

	RegisterPluralRule("xx", customRule)

	rule := GetPluralRule("xx")
	if rule(42) != Few {
		t.Error("Custom rule should return Few for 42")
	}
	if rule(1) != Other {
		t.Error("Custom rule should return Other for other counts")
	}
}
