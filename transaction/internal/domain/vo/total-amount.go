package vo

import (
	"database/sql/driver"
	"fmt"
)

// TotalAmount represents a monetary value in cents, suitable for large amounts such as total balances,
// using int64 for extended range.
type TotalAmount struct {
	Cents int64
}

// String returns a string representation of the TotalAmount, formatted as a decimal with two decimal places.
func (t TotalAmount) String() string {
	return fmt.Sprintf("%.2f", float64(t.Cents)/100)
}

// HumanReadable Returns formatted amount
func (t TotalAmount) HumanReadable() string {
	return fmt.Sprintf("%.2f", float64(t.Cents)/100)
}

// Value implements the Valuer interface and returns the total amount's value as a driver.Value.
func (t TotalAmount) Value() (driver.Value, error) {
	return t.Cents, nil
}

// Scan implements the Scanner interface and scans the value into the total amount.
func (t *TotalAmount) Scan(value interface{}) error {
	if value == nil {
		t.Cents = 0
		return nil
	}
	intValue, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to scan TotalAmount: %v", value)
	}
	t.Cents = intValue
	return nil
}

// NewTotalAmount returns TotalAmount instance.
func NewTotalAmount(value int64) TotalAmount {
	return TotalAmount{Cents: value}
}
