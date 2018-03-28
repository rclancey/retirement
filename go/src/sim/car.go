package sim

import (
	"errors"
	"math"
	"time"
)

type Car struct {
	*Simulacrum
	ledger *Ledger
	value float64
	purchasePrice float64
	purchaseDate time.Time
	loan *Debt
}

func NewCar(sim *Simulation, purchasePrice float64, purchaseDate time.Time, loan *Debt) *Car {
	if loan == nil {
		loan = NewDebt(sim, 0.0, 0.0, time.Date(2200, time.January, 1, 0, 0, 0, 0, time.Local), "Imaginary Car Loan")
	}
	c := &Car{
		Simulacrum: NewSimulacrum(sim),
		ledger: NewLedger(),
		purchasePrice: purchasePrice,
		purchaseDate: purchaseDate,
		loan: loan,
	}
	c.Transaction(purchasePrice, purchaseDate, OpenAccount)
	c.Depreciate(c.StartDate())
	return c
}

func (c *Car) Loan() *Debt {
	return c.loan
}

func (c *Car) Reconcile() {
	c.value = c.ledger.Balance()
}

func (c *Car) Transaction(amount float64, date time.Time, memo string) *Transaction {
	t := NewTransaction(amount, date, memo)
	c.value = c.ledger.Add(t)
	return t
}

func (c *Car) YearEndBalance(date time.Time) float64 {
	return c.ledger.FilterBefore(startOfYear(date)).Balance() + c.loan.YearEndBalance(date)
}

func (c *Car) Balance() float64 {
	return c.value + c.loan.Balance()
}

func (c *Car) BalanceOn(date time.Time) float64 {
	return c.ledger.FilterBefore(date).Balance() + c.loan.BalanceOn(date)
}

func (c *Car) Withdraw(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount == 0.0 {
		return nil, nil
	}
	return nil, errors.New("Cannot withdraw from a car")
}

func (c *Car) Deposit(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount == 0.0 {
		return nil, nil
	}
	return nil, errors.New("Cannot deposit into a car")
}

func (c *Car) CanWithdraw(date time.Time) bool {
	return false
}

func (c *Car) CanDeposit(date time.Time) bool {
	return false
}

func (c *Car) AccrueInterest(date time.Time) *Transaction {
	return nil
}

func (c *Car) AccrueMarketReturn(date time.Time) *Transaction {
	return nil
}

func (c *Car) Name() string {
	return "Car"
}

func (c *Car) Value() float64 {
	return c.value
}

func (c *Car) Age(date time.Time) float64 {
	return years(date.Sub(c.purchaseDate))
}

func (c *Car) Depreciate(date time.Time) {
	if c.Value() <= 0.0 {
		return
	}
	age := c.Age(date) * 12.0
	value := c.purchasePrice * 0.9 * (math.Atan(-0.03 * (age - 80.0)) + (math.Pi / 2.0)) / math.Pi
	delta := value - c.Value()
	if -1.0 * delta > c.Value() {
		delta = -1.0 * c.Value()
	}
	c.Transaction(delta, date, "Car " + Depreciation)
}

func (c *Car) Maintain(date time.Time) {
	if c.value > 0.0 && c.Age(date) > 5.0 {
		c.CashAccount().Withdraw(c.purchasePrice * 0.005, date, "Car " + Maintenance)
	}
}

func (c *Car) PayOff(date time.Time) *Transaction {
	t, _ := c.loan.Deposit(-1.0 * c.loan.Balance(), date, "Car Payoff")
	return t
}

func (c *Car) Liquidate(date time.Time) *Transaction {
	c.PayOff(date)
	sale := c.Transaction(-1.0 * c.Balance(), date, "Car Sale")
	c.CashAccount().Transaction(-1.0 * sale.Amount, date, "Car Sale")
	return sale
}

func (c *Car) Replace(date time.Time) {
	cash := c.CashAccount().Balance()
	cost := c.purchasePrice
	var finance float64
	if cash > cost {
		finance = 0.5
	} else if cash > cost / 2.0 {
		finance = 0.75
	} else {
		finance = 0.95
	}
	c.Liquidate(date)
	loan := NewDebt(c.Sim(), cost * finance, 0.05, date.AddDate(5, 0, 0), c.Name() + " Loan")
	c.loan = loan
	c.CashAccount().Withdraw(cost * (1.0 - finance), date, "Car Purchase")
	c.Transaction(cost, date, "Car Purchase")
	c.purchaseDate = date
	c.Depreciate(date)
	c.AddEvent(date, c.Sim().Name + " New Car", 4)
}

func (c *Car) Monthly(date time.Time) {
	c.Depreciate(date)
	if c.loan != nil {
		c.loan.Monthly(date)
	}
	if c.Simulacrum.Age(date) > 85 {
		if c.value > 0.0 {
			c.Liquidate(date)
			c.AddEvent(date, c.Sim().Name + " Quit Driving", 7)
		}
	} else {
		if c.Age(date) > c.Gauss(10.0, 2.0) {
			if c.Simulacrum.Age(date) < 80 {
				c.Replace(date)
			} else {
				if c.value > 0.0 {
					c.Liquidate(date)
					c.AddEvent(date, c.Sim().Name + " Quit Driving", 7)
				}
			}
		} else {
			c.Maintain(date)
		}
	}
}

