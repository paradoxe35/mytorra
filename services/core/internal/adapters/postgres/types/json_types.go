package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// CoordinatesSlice is a wrapper for []valueobjects.Coordinates that implements sql.Scanner and driver.Valuer
type CoordinatesSlice []valueobjects.Coordinates

// Scan implements sql.Scanner interface
func (cs *CoordinatesSlice) Scan(value interface{}) error {
	if value == nil {
		*cs = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, cs)
	case string:
		return json.Unmarshal([]byte(v), cs)
	default:
		return fmt.Errorf("cannot scan type %T into CoordinatesSlice", value)
	}
}

// Value implements driver.Valuer interface
func (cs CoordinatesSlice) Value() (driver.Value, error) {
	if cs == nil {
		return nil, nil
	}
	return json.Marshal(cs)
}

// JSONMap is a wrapper for map[string]interface{} that implements sql.Scanner and driver.Valuer
type JSONMap map[string]interface{}

// Scan implements sql.Scanner interface
func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, jm)
	case string:
		return json.Unmarshal([]byte(v), jm)
	default:
		return fmt.Errorf("cannot scan type %T into JSONMap", value)
	}
}

// Value implements driver.Valuer interface
func (jm JSONMap) Value() (driver.Value, error) {
	if jm == nil {
		return nil, nil
	}
	return json.Marshal(jm)
}

// JSONBoolMap is a wrapper for map[string]bool that implements sql.Scanner and driver.Valuer
type JSONBoolMap map[string]bool

// Scan implements sql.Scanner interface
func (jbm *JSONBoolMap) Scan(value interface{}) error {
	if value == nil {
		*jbm = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, jbm)
	case string:
		return json.Unmarshal([]byte(v), jbm)
	default:
		return fmt.Errorf("cannot scan type %T into JSONBoolMap", value)
	}
}

// Value implements driver.Valuer interface
func (jbm JSONBoolMap) Value() (driver.Value, error) {
	if jbm == nil {
		return nil, nil
	}
	return json.Marshal(jbm)
}

// JSONFloatMap is a wrapper for map[string]float64 that implements sql.Scanner and driver.Valuer
type JSONFloatMap map[string]float64

// Scan implements sql.Scanner interface
func (jfm *JSONFloatMap) Scan(value interface{}) error {
	if value == nil {
		*jfm = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, jfm)
	case string:
		return json.Unmarshal([]byte(v), jfm)
	default:
		return fmt.Errorf("cannot scan type %T into JSONFloatMap", value)
	}
}

// Value implements driver.Valuer interface
func (jfm JSONFloatMap) Value() (driver.Value, error) {
	if jfm == nil {
		return nil, nil
	}
	return json.Marshal(jfm)
}

// TransitionZoneSlice is a wrapper for []entities.TransitionZone that implements sql.Scanner and driver.Valuer
type TransitionZoneSlice []struct {
	AreaID           string                     `json:"area_id"`
	OverlapPolygon   []valueobjects.Coordinates `json:"overlap_polygon"`
	TransitionRadius float64                    `json:"transition_radius"`
}

// Scan implements sql.Scanner interface
func (tzs *TransitionZoneSlice) Scan(value interface{}) error {
	if value == nil {
		*tzs = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, tzs)
	case string:
		return json.Unmarshal([]byte(v), tzs)
	default:
		return fmt.Errorf("cannot scan type %T into TransitionZoneSlice", value)
	}
}

// Value implements driver.Valuer interface
func (tzs TransitionZoneSlice) Value() (driver.Value, error) {
	if tzs == nil {
		return nil, nil
	}
	return json.Marshal(tzs)
}
