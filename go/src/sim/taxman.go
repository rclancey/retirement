package sim

import (
	"math"
	"time"
)

type Bracket struct {
	Min float64
	Max float64
	Rate float64
}

var Brackets = map[string][]*Bracket {
	"US": []*Bracket{
		&Bracket{0.0, 9525.0, 0.1},
		&Bracket{9525.0, 38700.0, 0.15},
		&Bracket{38700.0, 93700.0, 0.25},
		&Bracket{93700.0, 195450.0, 0.28},
		&Bracket{195450.0, 424950.0, 0.33},
		&Bracket{424950.0, 426700.0, 0.35},
		&Bracket{426700.0, math.MaxFloat64, 0.396},
	},
	"SSA": []*Bracket{
		&Bracket{0.0, 128400.0, 0.062},
		&Bracket{128400.0, math.MaxFloat64, 0.0},
	},
	"MED": []*Bracket{
		&Bracket{0.0, math.MaxFloat64, 0.0145},
	},
	"CA": []*Bracket{
		&Bracket{0.0, 8223.0, 0.01},
		&Bracket{8223.0, 19495.0, 0.02},
		&Bracket{19495.0, 30769.0, 0.04},
		&Bracket{30769.0, 42711.0, 0.06},
		&Bracket{42711.0, 53980.0, 0.08},
		&Bracket{53980.0, 275738.0, 0.093},
		&Bracket{275738.0, 330884.0, 0.103},
		&Bracket{330884.0, 551473.0, 0.113},
		&Bracket{551473.0, math.MaxFloat64, 0.123},
	},
}

var StandardDeduction = map[string]float64{
	"US": 6500.0,
	"SSA": 0.0,
	"MED": 0.0,
	"CA": 4236.0,
}

type TaxMan struct {
	*Simulacrum
	state string
	earnings []float64
	withholding []float64
	deductions []float64
}

func NewTaxMan(sim *Simulation, state string) *TaxMan {
	return &TaxMan{
		Simulacrum: NewSimulacrum(sim),
		state: state,
		earnings: []float64{},
		withholding: []float64{},
		deductions: []float64{},
	}
}

func (t *TaxMan) Name() string {
	return t.state
}

func (t *TaxMan) StandardDeduction() float64 {
	return StandardDeduction[t.state]
}

func (t *TaxMan) Brackets() []*Bracket {
	return Brackets[t.state]
}

func (t *TaxMan) Withhold(earnings float64, date time.Time) float64 {
	t.earnings = append(t.earnings, earnings)
	w := t.Bracket(earnings * 12.0 - t.StandardDeduction()) / 12.0
	t.CashAccount().Withdraw(w, date, t.Name() + " Withholding")
	t.withholding = append(t.withholding, w)
	return w
}

func (t *TaxMan) Deduct(amount float64) {
	t.deductions = append(t.deductions, amount)
}

func (t *TaxMan) Tax() (total, owed float64) {
	var paid float64 = 0.0
	for _, w := range t.withholding {
		paid += w
	}
	var net float64 = 0.0
	for _, e := range t.earnings {
		net += e
	}
	for _, d := range t.deductions {
		net -= d
	}
	net -= t.StandardDeduction()
	if net < 0.0 {
		return 0.0, -1.0 * paid
	}
	total = t.Bracket(net)
	owed = total - paid
	t.earnings = []float64{}
	t.withholding = []float64{}
	t.deductions = []float64{}
	return total, owed
}

func (t *TaxMan) Bracket(net float64) float64 {
	var tax float64 = 0.0
	for _, b := range t.Brackets() {
		if net < b.Min {
			break
		}
		tax += b.Rate * (math.Min(b.Max, net) - b.Min)
	}
	return tax
}

type TaxMen struct {
	*Simulacrum
	Federal *TaxMan
	SocialSecurity *TaxMan
	Medicare *TaxMan
	State *TaxMan
}

func NewTaxMen(sim *Simulation, state string) *TaxMen {
	return &TaxMen{
		Simulacrum: NewSimulacrum(sim),
		Federal: NewTaxMan(sim, "US"),
		SocialSecurity: NewTaxMan(sim, "SSA"),
		Medicare: NewTaxMan(sim, "MED"),
		State: NewTaxMan(sim, state),
	}
}

func (t *TaxMen) Withhold(earnings float64, date time.Time, payroll bool) float64 {
	if earnings <= 0.0 {
		return 0.0
	}
	var w float64 = 0.0
	if payroll {
		w += t.SocialSecurity.Withhold(earnings, date)
		w += t.Medicare.Withhold(earnings, date)
	}
	w += t.Federal.Withhold(earnings, date)
	w += t.State.Withhold(earnings, date)
	return w
}

func (t *TaxMen) Deduct(amount float64) {
	t.Federal.Deduct(amount)
	t.State.Deduct(amount)
}

func (t *TaxMen) Annual(date time.Time) {
	stTot, stOwed := t.State.Tax()
	t.Federal.Deduct(stTot)
	_, fedOwed := t.Federal.Tax()
	owed := stOwed + fedOwed
	if owed > 0.0 {
		t.CashAccount().Withdraw(owed, date, TaxPayment)
		t.AddEvent(date, "Tax Bill", 1)
	} else {
		t.CashAccount().Deposit(-1.0 * owed, date, TaxRefund)
		t.AddEvent(date, "Tax Refund", 1)
	}
}

