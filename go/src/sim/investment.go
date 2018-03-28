package sim

import (
	"errors"
	"time"
)

type InvestmentAccount struct {
	*Simulacrum
	name string
	ledger *Ledger
	balance float64
	rmd RMD
	canWithdraw WithdrawRule
	canDeposit DepositRule
	taxable bool
}

func NewInvestmentAccount(sim *Simulation, balance float64, name string) *InvestmentAccount {
	a := &InvestmentAccount{
		Simulacrum: NewSimulacrum(sim),
		name: name,
		ledger: NewLedger(),
		taxable: false,
	}
	a.rmd = func(date time.Time, acct Account) float64 {
		return 0.0
	}
	a.canWithdraw = func(date time.Time) bool {
		return true
	}
	a.canDeposit = func(date time.Time) bool {
		return true
	}
	a.Transaction(balance, a.StartDate(), OpenAccount)
	return a
}

func New401K(sim *Simulation, balance float64, name string) *InvestmentAccount {
	a := &InvestmentAccount{
		Simulacrum: NewSimulacrum(sim),
		name: name + " 401k",
		ledger: NewLedger(),
		taxable: true,
	}
	a.rmd = K401RMD(sim)
	a.canWithdraw = K401CanWithdraw(sim)
	a.canDeposit = K401CanDeposit(sim)
	a.Transaction(balance, a.StartDate(), OpenAccount)
	return a
}

func NewInheritedIRA(sim *Simulation, balance float64, inheritDate time.Time, name string) *InvestmentAccount {
	a := &InvestmentAccount{
		Simulacrum: NewSimulacrum(sim),
		name: name + " IRA",
		ledger: NewLedger(),
		taxable: true,
	}
	a.rmd = InheritedIRARMD(sim, inheritDate)
	a.canWithdraw = InheritedIRACanWithdraw(sim)
	a.canDeposit = InheritedIRACanDeposit(sim)
	a.Transaction(balance, a.StartDate(), OpenAccount)
	return a
}

func (a *InvestmentAccount) Name() string {
	return a.name
}

func (a *InvestmentAccount) Reconcile() {
	a.balance = a.ledger.Balance()
}

func (a *InvestmentAccount) Transaction(amount float64, date time.Time, memo string) *Transaction {
	t := NewTransaction(amount, date, memo)
	a.balance = a.ledger.Add(t)
	return t
}

func (a *InvestmentAccount) YearEndBalance(date time.Time) float64 {
	return a.ledger.FilterBefore(startOfYear(date)).Balance()
}

func (a *InvestmentAccount) Balance() float64 {
	return a.balance
}

func (a *InvestmentAccount) BalanceOn(date time.Time) float64 {
	return a.ledger.FilterBefore(date).Balance()
}

func (a *InvestmentAccount) Withdraw(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount > a.Balance() {
		amount = a.Balance()
	}
	if amount == 0.0 {
		return nil, nil
	}
	if amount < 0.0 {
		return nil, errors.New("Cannot withdraw a negative amount")
	}
	if !a.CanWithdraw(date) {
		return nil, errors.New("IRS rules prevent withdrawl")
	}
	if memo == "" {
		memo = InvestmentWithdrawl
	}
	t := a.Transaction(-1.0 * amount, date, memo)
	a.CashAccount().Deposit(amount, date, memo)
	if a.Taxable() {
		a.TaxMen().Withhold(amount, date, false)
	}
	return t, nil
}

func (a *InvestmentAccount) Deposit(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount == 0.0 {
		return nil, nil
	}
	if amount < 0 {
		return nil, errors.New("Cannot deposit a negative amount")
	}
	if !a.CanDeposit(date) {
		return nil, errors.New("IRS rules prevent deposit")
	}
	if memo == "" {
		memo = InvestmentDeposit
	}
	a.CashAccount().Withdraw(amount, date, memo)
	t := a.Transaction(amount, date, memo)
	return t, nil
}

func (a *InvestmentAccount) CanWithdraw(date time.Time) bool {
	return a.canWithdraw(date)
}

func (a *InvestmentAccount) CanDeposit(date time.Time) bool {
	return a.canDeposit(date)
}

func (a *InvestmentAccount) YTDWithdrawls(date time.Time) float64 {
	ts := a.ledger.FilterAfter(startOfYear(date)).FilterAmountMax(0.0).FilterMemoOut(MarketReturn, InterestAccrual)
	var ytd float64 = 0.0
	for _, t := range *ts {
		ytd -= t.Amount
	}
	return ytd
}

func (a *InvestmentAccount) RMD(date time.Time) float64 {
	annual := a.rmd(date, a)
	rmd := annual - a.YTDWithdrawls(date)
	if rmd < 0.0 {
		return 0.0
	}
	return rmd
}

func (a *InvestmentAccount) TakeRMD(date time.Time) (*Transaction, error) {
	return a.Withdraw(a.RMD(date), date, RequiredMinimumDistribution)
}

func (a *InvestmentAccount) AccrueInterest(date time.Time) *Transaction {
	return nil
}

func (a *InvestmentAccount) AccrueMarketReturn(date time.Time) *Transaction {
	bal := a.Balance()
	if bal <= 0.0 {
		return nil
	}
	ret := a.Sim().Portfolio.PortfolioReturn(date)
	if ret < -1.0 {
		ret = -1.0
	}
	amt := bal * ret
	return a.Transaction(amt, date, MarketReturn)
}

func (a *InvestmentAccount) Taxable() bool {
	return a.taxable
}


