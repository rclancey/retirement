package sim

import (
	"time"
)

type Asset interface {
	Account
	Value() float64
	Loan() Debt
	Depreciate(date time.Time)
	Maintain(amount float64, date time.Time)
	CommissionRate() float64
	PayOff(date time.Time) *Transaction
	Liquidate(date time.Time) *Transaction
}

