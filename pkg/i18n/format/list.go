package format

import (
	"strings"
)

// ListStyle defines the style for list formatting.
type ListStyle int

const (
	// ListStyleAnd formats as "a, b, and c".
	ListStyleAnd ListStyle = iota
	// ListStyleOr formats as "a, b, or c".
	ListStyleOr
	// ListStyleNarrow formats as "a, b, c".
	ListStyleNarrow
)

// ListFormat holds locale-specific list formatting rules.
type ListFormat struct {
	AndConnector     string // " and " or ", and "
	OrConnector      string // " or " or ", or "
	PairAndConnector string // " and " for two items
	PairOrConnector  string // " or " for two items
	Separator        string // ", "
	OxfordComma      bool   // Use Oxford comma before final connector
}

// localeListFormats contains list formatting rules for various locales.
var localeListFormats = map[string]ListFormat{
	"en": {
		AndConnector:     "and ",
		OrConnector:      "or ",
		PairAndConnector: " and ",
		PairOrConnector:  " or ",
		Separator:        ", ",
		OxfordComma:      true,
	},
	"en-GB": {
		AndConnector:     "and ",
		OrConnector:      "or ",
		PairAndConnector: " and ",
		PairOrConnector:  " or ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"de": {
		AndConnector:     "und ",
		OrConnector:      "oder ",
		PairAndConnector: " und ",
		PairOrConnector:  " oder ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"fr": {
		AndConnector:     "et ",
		OrConnector:      "ou ",
		PairAndConnector: " et ",
		PairOrConnector:  " ou ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"es": {
		AndConnector:     "y ",
		OrConnector:      "o ",
		PairAndConnector: " y ",
		PairOrConnector:  " o ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"it": {
		AndConnector:     "e ",
		OrConnector:      "o ",
		PairAndConnector: " e ",
		PairOrConnector:  " o ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"pt": {
		AndConnector:     "e ",
		OrConnector:      "ou ",
		PairAndConnector: " e ",
		PairOrConnector:  " ou ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"nl": {
		AndConnector:     "en ",
		OrConnector:      "of ",
		PairAndConnector: " en ",
		PairOrConnector:  " of ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"ru": {
		AndConnector:     "и ",
		OrConnector:      "или ",
		PairAndConnector: " и ",
		PairOrConnector:  " или ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"ja": {
		AndConnector:     "、",
		OrConnector:      "または",
		PairAndConnector: "と",
		PairOrConnector:  "または",
		Separator:        "、",
		OxfordComma:      false,
	},
	"zh": {
		AndConnector:     "和",
		OrConnector:      "或",
		PairAndConnector: "和",
		PairOrConnector:  "或",
		Separator:        "、",
		OxfordComma:      false,
	},
	"ko": {
		AndConnector:     " 및 ",
		OrConnector:      " 또는 ",
		PairAndConnector: "와 ",
		PairOrConnector:  " 또는 ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"ar": {
		AndConnector:     "و",
		OrConnector:      "أو ",
		PairAndConnector: " و",
		PairOrConnector:  " أو ",
		Separator:        "، ",
		OxfordComma:      false,
	},
	"pl": {
		AndConnector:     "i ",
		OrConnector:      "lub ",
		PairAndConnector: " i ",
		PairOrConnector:  " lub ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"tr": {
		AndConnector:     "ve ",
		OrConnector:      "veya ",
		PairAndConnector: " ve ",
		PairOrConnector:  " veya ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"sv": {
		AndConnector:     "och ",
		OrConnector:      "eller ",
		PairAndConnector: " och ",
		PairOrConnector:  " eller ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"da": {
		AndConnector:     "og ",
		OrConnector:      "eller ",
		PairAndConnector: " og ",
		PairOrConnector:  " eller ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"no": {
		AndConnector:     "og ",
		OrConnector:      "eller ",
		PairAndConnector: " og ",
		PairOrConnector:  " eller ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"fi": {
		AndConnector:     "ja ",
		OrConnector:      "tai ",
		PairAndConnector: " ja ",
		PairOrConnector:  " tai ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"el": {
		AndConnector:     "και ",
		OrConnector:      "ή ",
		PairAndConnector: " και ",
		PairOrConnector:  " ή ",
		Separator:        ", ",
		OxfordComma:      false,
	},
	"he": {
		AndConnector:     "ו",
		OrConnector:      "או ",
		PairAndConnector: " ו",
		PairOrConnector:  " או ",
		Separator:        ", ",
		OxfordComma:      false,
	},
}

// GetListFormat returns the list format for a locale.
func GetListFormat(locale string) ListFormat {
	// Try exact match
	if fmt, ok := localeListFormats[locale]; ok {
		return fmt
	}

	// Try language only
	if idx := strings.Index(locale, "-"); idx != -1 {
		lang := locale[:idx]
		if fmt, ok := localeListFormats[lang]; ok {
			return fmt
		}
	}

	// Default to English
	return localeListFormats["en"]
}

// FormatList formats a list of items according to locale conventions.
func FormatList(locale string, items []string, style ListStyle) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}

	lf := GetListFormat(locale)

	switch style {
	case ListStyleAnd:
		return formatListWithConnector(items, lf.Separator, lf.AndConnector, lf.PairAndConnector, lf.OxfordComma)
	case ListStyleOr:
		return formatListWithConnector(items, lf.Separator, lf.OrConnector, lf.PairOrConnector, lf.OxfordComma)
	case ListStyleNarrow:
		return strings.Join(items, lf.Separator)
	default:
		return formatListWithConnector(items, lf.Separator, lf.AndConnector, lf.PairAndConnector, lf.OxfordComma)
	}
}

// formatListWithConnector formats a list with the specified connector.
func formatListWithConnector(items []string, separator, connector, pairConnector string, oxfordComma bool) string {
	if len(items) == 2 {
		return items[0] + pairConnector + items[1]
	}

	var result strings.Builder
	for i, item := range items {
		if i > 0 {
			if i == len(items)-1 {
				// Last item
				if oxfordComma {
					result.WriteString(strings.TrimSuffix(separator, " "))
					result.WriteString(" ")
					result.WriteString(connector)
				} else {
					// Without Oxford comma, just add a space before connector
					result.WriteString(" ")
					result.WriteString(connector)
				}
			} else {
				result.WriteString(separator)
			}
		}
		result.WriteString(item)
	}

	return result.String()
}
