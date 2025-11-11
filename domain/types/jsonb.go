// domain/types/jsonb.go
package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB is a type for handling PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for database/sql
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database/sql
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	if len(bytes) == 0 {
		*j = make(JSONB)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// MarshalJSON implements the json.Marshaler interface
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}

	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	*j = JSONB(raw)
	return nil
}

// Get retrieves a value from the JSONB map with a given key
func (j JSONB) Get(key string) interface{} {
	if j == nil {
		return nil
	}
	return j[key]
}

// GetString gets a string value from the JSONB
func (j JSONB) GetString(key string) string {
	val := j.Get(key)
	if val == nil {
		return ""
	}

	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// GetFloat gets a float64 value from the JSONB
func (j JSONB) GetFloat(key string) float64 {
	val := j.Get(key)
	if val == nil {
		return 0
	}

	// Direct type assertion if it's already float64
	if f, ok := val.(float64); ok {
		return f
	}

	// Try converting from number represented as something else
	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return 0
	}
}

// GetBool gets a boolean value from the JSONB
func (j JSONB) GetBool(key string) bool {
	val := j.Get(key)
	if val == nil {
		return false
	}

	b, ok := val.(bool)
	if !ok {
		return false
	}
	return b
}

// Set sets a value in the JSONB with the given key
func (j JSONB) Set(key string, value interface{}) {
	if j == nil {
		return
	}
	j[key] = value
}

// SafeForGorm คืนค่า map[string]interface{} ที่ปลอดภัยสำหรับ GORM
func (j JSONB) SafeForGorm() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range j {
		switch val := v.(type) {
		case time.Time:
			// แปลง time.Time เป็น string
			result[k] = val.Format("2006-01-02 15:04:05")
		default:
			result[k] = val
		}
	}
	return result
}
