package sim

import (
	"errors"
	"math"
	"time"
)

type Debt struct {
	*Simulacrum
	name string
	ledger *Ledger
	balance float64
	InterestRate float64
	DueDate time.Time
	TaxDeductible bool
}

func NewDebt(sim *Simulation, principal, interest float64, dueDate time.Time, name string) *Debt {
	d := &Debt{
		Simulacrum: NewSimulacrum(sim),
		name: name,
		ledger: NewLedger(),
		InterestRate: interest,
		DueDate: dueDate,
		TaxDeductible: false,
	}
	d.Transaction(-1.0 * principal, d.StartDate(), OpenAccount)
	return d
}

func NewMortgage(sim *Simulation, principal, interest float64, dueDate time.Time, name string) *Debt {
	d := NewDebt(sim, principal, interest, dueDate, name + " Mortgage")
	d.TaxDeductible = true
	return d
}

func (d *Debt) Name() string {
	return d.name
}

func (d *Debt) Reconcile() {
	d.balance = d.ledger.Balance()
}

func (d *Debt) Transaction(amount float64, date time.Time, memo string) *Transaction {
	t := NewTransaction(amount, date, memo)
	d.balance = d.ledger.Add(t)
	return t
}

func (d *Debt) YearEndBalance(date time.Time) float64 {
	return d.ledger.FilterBefore(startOfYear(date)).Balance()
}

func (d *Debt) Balance() float64 {
	return d.balance
}

func (d *Debt) BalanceOn(date time.Time) float64 {
	return d.ledger.FilterBefore(date).Balance()
}

func (d *Debt) Deposit(amount float64, date time.Time, memo string) (*Transaction, error) {
	if d.Balance() >= 0 {
		return nil, nil
	}
	if amount > -1.0 * d.Balance() {
		amount = -1.0 * d.Balance()
	}
	d.CashAccount().Withdraw(amount, date, memo)
	t := d.Transaction(amount, date, memo)
	return t, nil
}

func (d *Debt) Withdraw(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount == 0.0 {
		return nil, nil
	}
	return nil, errors.New("Cannot withdraw from a debt account")
}

func (d *Debt) CanWithdraw(date time.Time) bool {
	return false
}

func (d *Debt) CanDeposit(date time.Time) bool {
	return d.balance < 0.0
}

func (d *Debt) AccrueInterest(date time.Time) *Transaction {
	delta := d.Balance() * d.MonthlyInterestRate()
	t := d.Transaction(delta, date, InterestAccrual)
	if d.TaxDeductible {
		d.TaxMen().Deduct(delta)
	}
	return t
}

func (d *Debt) AccrueMarketReturn(date time.Time) *Transaction {
	return nil
}

func (d *Debt) MonthlyInterestRate() float64 {
	return d.InterestRate / 12.0
}

func (d *Debt) MonthlyInterestMultiplier() float64 {
	return 1.0 + d.MonthlyInterestRate()
}

func (d *Debt) MonthsUntilDue(date time.Time) int {
	return months(d.DueDate.Sub(date))
}

func (d *Debt) MinimumPayment(date time.Time) float64 {
	if d.Balance() >= 0.0 {
		return 0.0
	}
	months := d.MonthsUntilDue(date)
	if months <= 0 {
		return -1.0 * d.Balance()
	}
	var isum float64 = 0.0
	for k := 0; k < months; k++ {
		isum += math.Pow(d.MonthlyInterestMultiplier(), float64(k))
	}
	fullInterest := math.Pow(d.MonthlyInterestMultiplier(), float64(months)) / isum
	return -1.0 * d.Balance() * fullInterest
}

func (d *Debt) Monthly(date time.Time) {
	d.AccrueInterest(date)
	amount := d.MinimumPayment(date)
	d.Deposit(amount, date, DebtPayment)
	if amount > 0.0 && d.Balance() == 0.0 {
		d.AddEvent(date, "Payoff " + d.Name(), 6)
	}
}

