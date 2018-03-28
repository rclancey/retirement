package sim

import (
	"time"
)

type Child struct {
	*Simulacrum
	name string
	birthDate time.Time
	tuition []float64
}

func NewChild(sim *Simulation, name string, birthDate time.Time, tuition []float64) *Child {
	return &Child{
		Simulacrum: NewSimulacrum(sim),
		name: name,
		birthDate: birthDate,
		tuition: tuition,
	}
}

func (c *Child) Name() string {
	return c.name
}

func (c *Child) Age(date time.Time) float64 {
	return years(date.Sub(c.birthDate))
}

func (c *Child) GraduationDate() time.Time {
	y := 0
	if c.birthDate.Month() >= time.September {
		y = 1
	}
	start := time.Date(c.birthDate.Year() + y, time.June, 30, 0, 0, 0, 0, time.Local)
	return start.AddDate(len(c.tuition), 0, 0)
}

func (c *Child) Remaining(date time.Time) float64 {
	var balance float64 = 0.0
	age := c.Age(date)
	if age <= 0.0 {
		for _, t := range c.tuition {
			balance += t
		}
	} else {
		iage := int(age)
		if iage >= len(c.tuition) {
			return 0.0
		}
		rem := 1.0 + float64(iage) - age
		balance += c.tuition[iage] * rem
		for i := iage + 1; i < len(c.tuition); i++ {
			balance += c.tuition[i]
		}
	}
	return balance
}

func (c *Child) Monthly(date time.Time) {
	age := int(c.Age(date))
	if age < 0 {
		return
	}
	if age < len(c.tuition) {
		c.CashAccount().Withdraw(c.tuition[age] / 12.0, date, "Tuition Payment")
	}
}

