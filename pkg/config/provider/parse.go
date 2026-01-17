package provider

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// jsonUnmarshal wraps json.Unmarshal.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// yamlUnmarshal wraps yaml.Unmarshal.
func yamlUnmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
