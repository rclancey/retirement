package sim

import (
	"errors"
	"time"
)

type Home struct {
	*Simulacrum
	ledger *Ledger
	value float64
	mortgage *Debt
	propertyTax float64
	rent float64
}

func NewHome(sim *Simulation, value float64, mortgage *Debt, tax, rent float64) *Home {
	h := &Home{
		Simulacrum: NewSimulacrum(sim),
		ledger: NewLedger(),
		mortgage: mortgage,
		propertyTax: tax,
		rent: rent,
	}
	h.Transaction(value, h.StartDate(), OpenAccount)
	return h
}

func (h *Home) Mortgage() *Debt {
	return h.mortgage
}

func (h *Home) Reconcile() {
	h.value = h.ledger.Balance()
}

func (h *Home) Transaction(amount float64, date time.Time, memo string) *Transaction {
	t := NewTransaction(amount, date, memo)
	h.value = h.ledger.Add(t)
	return t
}

func (h *Home) YearEndBalance(date time.Time) float64 {
	return h.ledger.FilterBefore(startOfYear(date)).Balance() + h.mortgage.YearEndBalance(date)
}

func (h *Home) Balance() float64 {
	return h.value + h.mortgage.Balance()
}

func (h *Home) BalanceOn(date time.Time) float64 {
	return h.ledger.FilterBefore(date).Balance() + h.mortgage.BalanceOn(date)
}

func (h *Home) Withdraw(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount == 0.0 {
		return nil, nil
	}
	return nil, errors.New("Cannot withdraw from a home")
}

func (h *Home) Deposit(amount float64, date time.Time, memo string) (*Transaction, error) {
	if amount < 0.0 {
		return nil, errors.New("Cannot deposit a negative amount")
	}
	if amount == 0.0 {
		return nil, nil
	}
	h.CashAccount().Withdraw(amount, date, memo)
	t := h.Transaction(amount, date, memo)
	return t, nil
}

func (h *Home) CanWithdraw(date time.Time) bool {
	return false
}

func (h *Home) CanDeposit(date time.Time) bool {
	return true
}

func (h *Home) AccrueInterest(date time.Time) *Transaction {
	return nil
}

func (h *Home) AccrueMarketReturn(date time.Time) *Transaction {
	return nil
}

func (h *Home) Name() string {
	return "Home"
}

func (h *Home) Value() float64 {
	return h.value
}

func (h *Home) Depreciate(date time.Time) {
	h.Transaction(-200.0, date, "Home Depreciation")
}

func (h *Home) Maintain(date time.Time) {
	if date.Year() % 5 == 0 && date.Month() == time.June {
		prev := time.Date(date.Year() - 5, date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		l := h.ledger.FilterAfter(prev).FilterMemoIn("Home Depreciation")
		var amount float64 = 0.0
		for _, t := range *l {
			amount += t.Amount
		}
		h.Deposit(-1.0 * amount, date, "Home Maintenance")
		h.AddEvent(date, "Home Repairs", 4)
	}
}

func (h *Home) PayOff(date time.Time) *Transaction {
	t, _ := h.mortgage.Deposit(-1.0 * h.mortgage.Balance(), date, "Mortgage Payoff")
	return t
}

func (h *Home) Liquidate(date time.Time) *Transaction {
	h.PayOff(date)
	sale := h.Transaction(-1.0 * h.Balance(), date, "Home Sale")
	h.CashAccount().Transaction(-1.0 * sale.Amount, date, "Home Sale")
	commission := 0.06 * sale.Amount
	h.CashAccount().Transaction(commission, date, "Home Sale Commission")
	h.AddEvent(date, "Sell House", 8)
	return sale
}

func (h *Home) Monthly(date time.Time) {
	if h.Sim().AssistedLiving.NeedsAssistedLiving(date) {
		if h.Value() > 0.0 {
			h.Liquidate(date)
			h.AddEvent(date, "Move to Assisted Living", 7)
		}
		return
	}
	if h.value > 0.0 {
		if date.Month() == time.November {
			ptax := h.value * h.propertyTax
			h.CashAccount().Withdraw(ptax, date, "Property Tax")
			h.TaxMen().Deduct(ptax)
			h.AddEvent(date, "Property Taxes", 2)
		}
		h.Depreciate(date)
		h.Maintain(date)
		h.mortgage.Monthly(date)
	} else {
		h.CashAccount().Withdraw(h.rent, date, "Rent")
	}
}
