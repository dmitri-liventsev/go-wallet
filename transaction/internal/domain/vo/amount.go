package vo

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
)

// Amount represents a monetary value in cents, suitable for amounts up to the maximum int value in Go.
type Amount struct {
	Cents int
}

// Inverse returns a new Amount with the value inverted.
func (a Amount) Inverse() Amount {
	return NewAmount(-a.Cents)
}

// Add returns a new Amount which is the sum of the current amount and the given amount.
func (a *Amount) Add(amount Amount) Amount {
	return NewAmount(a.Cents + amount.Cents)
}

// Equal returns true if the current amount is equal to the given amount, otherwise false.
func (a *Amount) Equal(amount Amount) bool {
	return a.Cents == amount.Cents
}

// Value implements the Valuer interface and returns the amount's value as a driver.Value.
func (a Amount) Value() (driver.Value, error) {
	return a.Cents, nil
}

// Scan implements the Scanner interface and scans the value into the Amount;
func (a *Amount) Scan(value interface{}) error {
	if value == nil {
		a.Cents = 0
		return nil
	}

	switch v := value.(type) {
	case int64:
		a.Cents = int(v)
	case float64:
		a.Cents = int(v)
	case int:
		a.Cents = v
	default:
		return fmt.Errorf("failed to scan Amount: %v", value)
	}

	return nil
}

// NewAmountFromString creates a new Amount from a string representation, returning an error if the string cannot be parsed.
func NewAmountFromString(amount string) (Amount, error) {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {

		return NewAmount(0), err
	}

	return NewAmountFromFloat(amountFloat), nil
}

// NewAmountFromFloat creates a new Amount from a float64 value, rounding to the nearest cent.
func NewAmountFromFloat(amount float64) Amount {
	amountInt := int(math.Round(amount * 100))
	return NewAmount(amountInt)
}

// NewAmount returns NewAmount instance.
func NewAmount(value int) Amount {
	return Amount{Cents: value}
}
