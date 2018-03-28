package sim

import (
	"time"
	"github.com/satori/go.uuid"
)

const (
	OpenAccount = "Open Account"
	InterestPayment = "Interest Payment"
	InterestAccrual = "Interest Accrual"
	MarketReturn = "Market Return"
	CloseAccount = "Close Account"
	TaxPayment = "Tax Payment"
	TaxRefund = "Tax Refund"
	DebtPayment = "Debt Payment"
	CashWithdrawl = "Cash Withdrawl"
	CashDeposit = "Cash Deposit"
	Depreciation = "Depreciation"
	Maintenance = "Maintenance"
	RequiredMinimumDistribution = "Required Minimum Distribution"
	InvestmentDeposit = "Investment Deposit"
	InvestmentWithdrawl = "Investment Withdrawl"
)

type Transaction struct {
	Id uuid.UUID `json:"-"`
	Amount float64 `json:"amount"`
	Date time.Time `json:"date"`
	Memo string `json:"memo"`
}

func NewTransaction(amount float64, date time.Time, memo string) *Transaction {
	return &Transaction{
		Id: uuid.Must(uuid.NewV4()),
		Amount: amount,
		Date: date,
		Memo: memo,
	}
}

type Ledger []*Transaction

func NewLedger() *Ledger {
	ol := Ledger([]*Transaction{})
	return &ol
}

func (l *Ledger) FilterAfter(start time.Time) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		if t.Date.Equal(start) || t.Date.After(start) {
			out = append(out, t)
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) FilterBefore(end time.Time) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		if t.Date.Before(end) {
			out = append(out, t)
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) FilterMemoIn(good ...string) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		for _, m := range good {
			if t.Memo == m {
				out = append(out, t)
				break
			}
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) FilterMemoOut(bad ...string) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		ok := true
		for _, m := range bad {
			if t.Memo == m {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, t)
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) FilterAmountMin(min float64) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		if t.Amount >= min {
			out = append(out, t)
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) FilterAmountMax(max float64) *Ledger {
	out := []*Transaction{}
	for _, t := range *l {
		if t.Amount <= max {
			out = append(out, t)
		}
	}
	ol := Ledger(out)
	return &ol
}

func (l *Ledger) Add(t *Transaction) float64 {
	if t.Amount != 0.0 {
		*l = append(*l, t)
	}
	return l.Balance()
}

func (l *Ledger) Balance() float64 {
	var sum float64 = 0.0
	for _, t := range *l {
		sum += t.Amount
	}
	return sum
}

