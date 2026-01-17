package parser

import (
	"encoding/json"
	"fmt"
)

// JSONParser parses JSON configuration data.
type JSONParser struct{}

// NewJSONParser creates a new JSON parser.
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse parses JSON data into a configuration map.
func (p *JSONParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// Extensions returns the file extensions supported by this parser.
func (p *JSONParser) Extensions() []string {
	return []string{".json"}
}
