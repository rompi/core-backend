package i18n

// Message represents a translatable message with optional plural forms.
type Message struct {
	// ID is the unique identifier for the message.
	ID string `json:"id,omitempty"`

	// Description provides context for translators.
	Description string `json:"description,omitempty"`

	// One is the singular form of the message.
	One string `json:"one,omitempty"`

	// Other is the plural form of the message (default for languages without complex pluralization).
	Other string `json:"other,omitempty"`

	// Zero is the form used when count is zero (used by some languages like Arabic).
	Zero string `json:"zero,omitempty"`

	// Two is the form used when count is two (used by some languages like Arabic, Welsh).
	Two string `json:"two,omitempty"`

	// Few is the form used for small numbers (used by Slavic languages, etc.).
	Few string `json:"few,omitempty"`

	// Many is the form used for large numbers (used by some languages like Russian, Arabic).
	Many string `json:"many,omitempty"`
}

// GetForm returns the appropriate message form for the given plural category.
func (m *Message) GetForm(category PluralCategory) string {
	switch category {
	case Zero:
		if m.Zero != "" {
			return m.Zero
		}
		return m.Other
	case One:
		if m.One != "" {
			return m.One
		}
		return m.Other
	case Two:
		if m.Two != "" {
			return m.Two
		}
		return m.Other
	case Few:
		if m.Few != "" {
			return m.Few
		}
		return m.Other
	case Many:
		if m.Many != "" {
			return m.Many
		}
		return m.Other
	default:
		return m.Other
	}
}

// HasPluralForms returns true if the message has any plural forms defined.
func (m *Message) HasPluralForms() bool {
	return m.One != "" || m.Zero != "" || m.Two != "" || m.Few != "" || m.Many != ""
}

// SimpleMessage returns the message content for a simple (non-plural) message.
// It returns Other if set, otherwise One.
func (m *Message) SimpleMessage() string {
	if m.Other != "" {
		return m.Other
	}
	return m.One
}
