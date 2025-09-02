package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONBValue represents a JSONB database column that can store any JSON data
type JSONBValue json.RawMessage

// NewJSONBValue creates a new JSONBValue from any interface
func NewJSONBValue(v interface{}) JSONBValue {
	data, _ := json.Marshal(v)
	return JSONBValue(data)
}

// Value implements the driver.Valuer interface for database storage
func (j JSONBValue) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSONBValue) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = nil
			return nil
		}
		*j = JSONBValue(v)
	case string:
		if v == "" {
			*j = nil
			return nil
		}
		*j = JSONBValue([]byte(v))
	default:
		return errors.New("cannot scan unknown type into JSONBValue")
	}
	return nil
}

// Unmarshal unmarshals the JSON data into the provided interface
func (j JSONBValue) Unmarshal(v interface{}) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(j), v)
}

// MarshalJSON implements json.Marshaler
func (j JSONBValue) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONBValue) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		*j = nil
		return nil
	}
	*j = JSONBValue(data)
	return nil
}
