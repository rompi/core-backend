package i18n

import (
	"strings"
	"sync"
)

// PluralCategory represents a CLDR plural category.
type PluralCategory int

const (
	// Zero is used for zero quantity in some languages.
	Zero PluralCategory = iota
	// One is the singular form.
	One
	// Two is used for dual form in some languages.
	Two
	// Few is used for small numbers in some languages.
	Few
	// Many is used for large numbers in some languages.
	Many
	// Other is the general plural form.
	Other
)

// String returns the string representation of the plural category.
func (c PluralCategory) String() string {
	switch c {
	case Zero:
		return "zero"
	case One:
		return "one"
	case Two:
		return "two"
	case Few:
		return "few"
	case Many:
		return "many"
	default:
		return "other"
	}
}

// PluralRule is a function that determines the plural category for a given count.
type PluralRule func(n int) PluralCategory

var (
	pluralRules   = make(map[string]PluralRule)
	pluralRulesMu sync.RWMutex
)

func init() {
	// Register built-in plural rules based on CLDR
	registerBuiltinPluralRules()
}

// RegisterPluralRule registers a custom plural rule for a language.
func RegisterPluralRule(lang string, rule PluralRule) {
	pluralRulesMu.Lock()
	defer pluralRulesMu.Unlock()
	pluralRules[strings.ToLower(lang)] = rule
}

// GetPluralRule returns the plural rule for a language.
func GetPluralRule(lang string) PluralRule {
	pluralRulesMu.RLock()
	defer pluralRulesMu.RUnlock()

	lang = strings.ToLower(lang)

	// Try exact match
	if rule, ok := pluralRules[lang]; ok {
		return rule
	}

	// Try language only (strip region)
	if idx := strings.Index(lang, "-"); idx != -1 {
		baseLang := lang[:idx]
		if rule, ok := pluralRules[baseLang]; ok {
			return rule
		}
	}

	// Default to "other" only rule
	return pluralRuleOther
}

// registerBuiltinPluralRules registers plural rules for common languages.
func registerBuiltinPluralRules() {
	// English, German, Spanish, Italian, Portuguese, Dutch, Swedish, Norwegian, Danish
	// one: n = 1
	// other: everything else
	for _, lang := range []string{"en", "de", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "el", "he", "hu", "tr"} {
		pluralRules[lang] = pluralRuleOneTwoGroup1
	}

	// French: one for n = 0 or n = 1
	pluralRules["fr"] = pluralRuleFrench

	// Russian, Ukrainian, Belarusian, Serbian, Croatian, Bosnian
	pluralRules["ru"] = pluralRuleSlavic
	pluralRules["uk"] = pluralRuleSlavic
	pluralRules["be"] = pluralRuleSlavic
	pluralRules["sr"] = pluralRuleSlavic
	pluralRules["hr"] = pluralRuleSlavic
	pluralRules["bs"] = pluralRuleSlavic

	// Polish
	pluralRules["pl"] = pluralRulePolish

	// Czech, Slovak
	pluralRules["cs"] = pluralRuleCzechSlovak
	pluralRules["sk"] = pluralRuleCzechSlovak

	// Arabic
	pluralRules["ar"] = pluralRuleArabic

	// Chinese, Japanese, Korean, Vietnamese, Thai, Indonesian, Malay
	// No plural forms - always use "other"
	for _, lang := range []string{"zh", "ja", "ko", "vi", "th", "id", "ms"} {
		pluralRules[lang] = pluralRuleOther
	}

	// Welsh
	pluralRules["cy"] = pluralRuleWelsh

	// Irish
	pluralRules["ga"] = pluralRuleIrish

	// Slovenian
	pluralRules["sl"] = pluralRuleSlovenian

	// Lithuanian
	pluralRules["lt"] = pluralRuleLithuanian

	// Latvian
	pluralRules["lv"] = pluralRuleLatvian

	// Romanian
	pluralRules["ro"] = pluralRuleRomanian
}

// pluralRuleOther always returns Other.
func pluralRuleOther(n int) PluralCategory {
	return Other
}

// pluralRuleOneTwoGroup1 is for languages where n=1 is singular, everything else is plural.
func pluralRuleOneTwoGroup1(n int) PluralCategory {
	if n == 1 {
		return One
	}
	return Other
}

// pluralRuleFrench: one for n = 0 or n = 1
func pluralRuleFrench(n int) PluralCategory {
	if n == 0 || n == 1 {
		return One
	}
	return Other
}

// pluralRuleSlavic: Russian, Ukrainian, Belarusian, Serbian, Croatian, Bosnian
// one: n % 10 = 1 and n % 100 != 11
// few: n % 10 in 2..4 and n % 100 not in 12..14
// many: n % 10 = 0 or n % 10 in 5..9 or n % 100 in 11..14
// other: everything else (mainly fractions)
func pluralRuleSlavic(n int) PluralCategory {
	mod10 := n % 10
	mod100 := n % 100

	if mod10 == 1 && mod100 != 11 {
		return One
	}
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return Few
	}
	if mod10 == 0 || (mod10 >= 5 && mod10 <= 9) || (mod100 >= 11 && mod100 <= 14) {
		return Many
	}
	return Other
}

// pluralRulePolish
// one: n = 1
// few: n % 10 in 2..4 and n % 100 not in 12..14
// many: n != 1 and n % 10 in 0..1 or n % 10 in 5..9 or n % 100 in 12..14
func pluralRulePolish(n int) PluralCategory {
	if n == 1 {
		return One
	}

	mod10 := n % 10
	mod100 := n % 100

	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return Few
	}

	if (mod10 == 0 || mod10 == 1 || (mod10 >= 5 && mod10 <= 9)) || (mod100 >= 12 && mod100 <= 14) {
		return Many
	}

	return Other
}

// pluralRuleCzechSlovak
// one: n = 1
// few: n in 2..4
// other: everything else
func pluralRuleCzechSlovak(n int) PluralCategory {
	if n == 1 {
		return One
	}
	if n >= 2 && n <= 4 {
		return Few
	}
	return Other
}

// pluralRuleArabic
// zero: n = 0
// one: n = 1
// two: n = 2
// few: n % 100 in 3..10
// many: n % 100 in 11..99
// other: everything else
func pluralRuleArabic(n int) PluralCategory {
	if n == 0 {
		return Zero
	}
	if n == 1 {
		return One
	}
	if n == 2 {
		return Two
	}

	mod100 := n % 100
	if mod100 >= 3 && mod100 <= 10 {
		return Few
	}
	if mod100 >= 11 && mod100 <= 99 {
		return Many
	}
	return Other
}

// pluralRuleWelsh
// zero: n = 0
// one: n = 1
// two: n = 2
// few: n = 3
// many: n = 6
// other: everything else
func pluralRuleWelsh(n int) PluralCategory {
	switch n {
	case 0:
		return Zero
	case 1:
		return One
	case 2:
		return Two
	case 3:
		return Few
	case 6:
		return Many
	default:
		return Other
	}
}

// pluralRuleIrish
// one: n = 1
// two: n = 2
// few: n in 3..6
// many: n in 7..10
// other: everything else
func pluralRuleIrish(n int) PluralCategory {
	if n == 1 {
		return One
	}
	if n == 2 {
		return Two
	}
	if n >= 3 && n <= 6 {
		return Few
	}
	if n >= 7 && n <= 10 {
		return Many
	}
	return Other
}

// pluralRuleSlovenian
// one: n % 100 = 1
// two: n % 100 = 2
// few: n % 100 in 3..4
// other: everything else
func pluralRuleSlovenian(n int) PluralCategory {
	mod100 := n % 100
	if mod100 == 1 {
		return One
	}
	if mod100 == 2 {
		return Two
	}
	if mod100 == 3 || mod100 == 4 {
		return Few
	}
	return Other
}

// pluralRuleLithuanian
// one: n % 10 = 1 and n % 100 not in 11..19
// few: n % 10 in 2..9 and n % 100 not in 11..19
// other: everything else
func pluralRuleLithuanian(n int) PluralCategory {
	mod10 := n % 10
	mod100 := n % 100

	if mod10 == 1 && (mod100 < 11 || mod100 > 19) {
		return One
	}
	if mod10 >= 2 && mod10 <= 9 && (mod100 < 11 || mod100 > 19) {
		return Few
	}
	return Other
}

// pluralRuleLatvian
// zero: n = 0
// one: n % 10 = 1 and n % 100 != 11
// other: everything else
func pluralRuleLatvian(n int) PluralCategory {
	if n == 0 {
		return Zero
	}
	if n%10 == 1 && n%100 != 11 {
		return One
	}
	return Other
}

// pluralRuleRomanian
// one: n = 1
// few: n = 0 or n != 1 and n % 100 in 1..19
// other: everything else
func pluralRuleRomanian(n int) PluralCategory {
	if n == 1 {
		return One
	}
	mod100 := n % 100
	if n == 0 || (mod100 >= 1 && mod100 <= 19) {
		return Few
	}
	return Other
}
