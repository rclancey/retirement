package sim

import (
	"errors"
	"time"
)

type Account interface {
	//Transactions(start, end *time.Time) *Ledger
	Name() string
	Transaction(amount float64, date time.Time, memo string) *Transaction
	YearEndBalance(time.Time) float64
	Balance() float64
	Withdraw(amount float64, date time.Time, memo string) (*Transaction, error)
	Deposit(amount float64, date time.Time, memo string) (*Transaction, error)
	CanWithdraw(date time.Time) bool
	CanDeposit(date time.Time) bool
	AccrueInterest(date time.Time) *Transaction
	AccrueMarketReturn(date time.Time) *Transaction
}

type CashAccount struct {
	*Simulacrum
	ledger *Ledger
	balance float64
}

func NewCashAccount(sim *Simulation, balance float64) *CashAccount {
	a := &CashAccount{
		Simulacrum: NewSimulacrum(sim),
		ledger: NewLedger(),
		balance: 0.0,
	}
	a.Transaction(balance, a.StartDate(), OpenAccount);
	return a
}

func (a *CashAccount) Name() string {
	return "Cash"
}

func (a *CashAccount) Reconcile() {
	a.balance = a.ledger.Balance()
}

func (a *CashAccount) Transaction(amount float64, date time.Time, memo string) *Transaction {
	t := NewTransaction(amount, date, memo)
	a.balance = a.ledger.Add(t)
	return t
}

func (a *CashAccount) Transactions(start, end *time.Time) *Ledger {
	l := a.ledger
	if start != nil {
		l = l.FilterAfter(*start)
	}
	if end != nil {
		l = l.FilterBefore(*end)
	}
	return l
}

func (a *CashAccount) YearEndBalance(date time.Time) float64 {
	return a.ledger.FilterBefore(startOfYear(date)).Balance()
}

func (a *CashAccount) Balance() float64 {
	return a.balance
}

func (a *CashAccount) BalanceOn(date time.Time) float64 {
	return a.ledger.FilterBefore(date).Balance()
}

func (a *CashAccount) Withdraw(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount < 0.0 {
		return nil, errors.New("Cannot withdraw a negative amount")
	}
	if amount == 0.0 {
		return nil, nil
	}
	if memo == "" {
		memo = CashWithdrawl
	}
	t := a.Transaction(-1.0 * amount, date, memo)
	return t, nil
}

func (a *CashAccount) Deposit(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount < 0.0 {
		return nil, errors.New("Cannot deposit a negative amount")
	}
	if amount == 0.0 {
		return nil, nil
	}
	if memo == "" {
		memo = CashDeposit
	}
	t := a.Transaction(amount, date, memo)
	return t, nil
}

func (a *CashAccount) CanWithdraw(date time.Time) bool {
	return true
}

func (a *CashAccount) CanDeposit(date time.Time) bool {
	return true
}

func (a *CashAccount) AccrueInterest(date time.Time) *Transaction {
	return nil
}

func (a *CashAccount) AccrueMarketReturn(date time.Time) *Transaction {
	inflation := a.Sim().Economy.Inflation(date)
	loss := -1.0 * a.Balance() * (inflation / 1200.0)
	return a.Transaction(loss, date, "Inflation")
}

