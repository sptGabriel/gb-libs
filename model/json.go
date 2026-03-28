package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON json.RawMessage

func (JSON) GormDataType() string {
	return "jsonb"
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			bytes = make([]byte, len(v))
			copy(bytes, v)
		}
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	result := json.RawMessage(bytes)
	*j = JSON(result)
	return nil
}
