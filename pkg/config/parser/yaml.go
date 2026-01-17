package parser

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAMLParser parses YAML configuration data.
type YAMLParser struct{}

// NewYAMLParser creates a new YAML parser.
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse parses YAML data into a configuration map.
func (p *YAMLParser) Parse(data []byte) (map[string]any, error) {
	var result map[string]any

	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return result, nil
}

// Extensions returns the file extensions supported by this parser.
func (p *YAMLParser) Extensions() []string {
	return []string{".yaml", ".yml"}
}
