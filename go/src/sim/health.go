package sim

import (
	"math"
	"time"
)

type AssistedLiving struct {
	*Simulacrum
	basicRate float64
	terminalRate float64
}

func NewAssistedLiving(sim *Simulation, basicRate, terminalRate float64) *AssistedLiving {
	return &AssistedLiving{
		Simulacrum: NewSimulacrum(sim),
		basicRate: basicRate,
		terminalRate: terminalRate,
	}
}

func (l *AssistedLiving) NeedsAssistedLiving(date time.Time) bool {
	death := l.Sim().Actuary.DeathAge()
	if death < 82.0 {
		return false
	}
	age := l.Age(date)
	return age > 95.0 || age > death - 2.0
}

func (l *AssistedLiving) Terminal(date time.Time) bool {
	death := l.Sim().Actuary.DeathAge()
	if death < 85.0 {
		return false
	}
	return l.Age(date) > death - 1.0
}

func (l *AssistedLiving) Monthly(date time.Time) {
	if l.NeedsAssistedLiving(date) {
		amt := l.basicRate
		if l.Terminal(date) {
			amt = l.terminalRate
		}
		l.CashAccount().Withdraw(amt, date, "Assisted Living")
	}
}

type HealthCare struct {
	*Simulacrum
	premium float64
	outOfPocket float64
	inflator float64
}

func NewHealthCare(sim *Simulation, premium, outOfPocket, inflator float64) *HealthCare {
	return &HealthCare{
		Simulacrum: NewSimulacrum(sim),
		premium: premium,
		outOfPocket: outOfPocket,
		inflator: inflator,
	}
}

func (h *HealthCare) Inflator(date time.Time) float64 {
	y := years(date.Sub(h.StartDate()))
	return math.Pow(1.0 + h.inflator, y)
}

func (h *HealthCare) Premium(date time.Time) float64 {
	return h.premium * h.Inflator(date)
}

func (h *HealthCare) OutOfPocket(date time.Time) float64 {
	r := h.Sim().Actuary.DeathRisk(date) / 12.0
	if h.Random() <= r {
		h.AddEvent(date, "Health Issue", 8)
		return h.outOfPocket * h.Inflator(date)
	}
	return 0.0
}

func (h *HealthCare) Monthly(date time.Time) {
	h.CashAccount().Withdraw(h.Premium(date), date, "Health Insurance")
	h.CashAccount().Withdraw(h.OutOfPocket(date), date, "Health Care")
}

