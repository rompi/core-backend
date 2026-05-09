package redis

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// Serializer defines the interface for value serialization.
type Serializer interface {
	Serialize(value interface{}) ([]byte, error)
	Deserialize(data []byte, dest interface{}) error
}

// NewSerializer creates a new serializer based on the type.
func NewSerializer(serializerType SerializerType) (Serializer, error) {
	switch serializerType {
	case SerializerJSON, "":
		return &JSONSerializer{}, nil
	case SerializerGob:
		return &GobSerializer{}, nil
	case SerializerMsgPack:
		return &MsgPackSerializer{}, nil
	default:
		return nil, fmt.Errorf("unknown serializer type: %s", serializerType)
	}
}

// JSONSerializer serializes values using JSON.
type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(value interface{}) ([]byte, error) {
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}
	if b, ok := value.([]byte); ok {
		return b, nil
	}
	return json.Marshal(value)
}

func (s *JSONSerializer) Deserialize(data []byte, dest interface{}) error {
	if strPtr, ok := dest.(*string); ok {
		*strPtr = string(data)
		return nil
	}
	return json.Unmarshal(data, dest)
}

// GobSerializer serializes values using gob encoding.
type GobSerializer struct{}

func (s *GobSerializer) Serialize(value interface{}) ([]byte, error) {
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *GobSerializer) Deserialize(data []byte, dest interface{}) error {
	if strPtr, ok := dest.(*string); ok {
		*strPtr = string(data)
		return nil
	}
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(dest)
}

// MsgPackSerializer uses JSON as fallback.
type MsgPackSerializer struct{}

func (s *MsgPackSerializer) Serialize(value interface{}) ([]byte, error) {
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}
	return json.Marshal(value)
}

func (s *MsgPackSerializer) Deserialize(data []byte, dest interface{}) error {
	if strPtr, ok := dest.(*string); ok {
		*strPtr = string(data)
		return nil
	}
	return json.Unmarshal(data, dest)
}
