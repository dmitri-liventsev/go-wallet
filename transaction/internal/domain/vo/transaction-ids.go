package vo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type TransactionIds []string

func (p *TransactionIds) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (p TransactionIds) Value() (driver.Value, error) {
	return json.Marshal(p)
}
