package sim

import (
	"math"
	"time"
)

type Job struct {
	*Simulacrum
	employed int
	unemployed int
	monthly float64
}

func NewJob(sim *Simulation, baseAnnualSalary float64) *Job {
	return &Job{
		Simulacrum: NewSimulacrum(sim),
		employed: 0,
		unemployed: 0,
		monthly: baseAnnualSalary / 12.0,
	}
}

func (j *Job) Employed(date time.Time) bool {
	if j.Retired(date) {
		j.unemployed = 0
		return false
	}
	if j.employed < 0 {
		return false
	}
	bc := j.Sim().Economy.MarketReturn(date)
	if bc < -10.0 {
		if j.employed > int(math.Round(j.Gauss(12.0, 4.0))) {
			j.employed = int(math.Min(-1.0, math.Round(j.Gauss(-4.0, 2.0))))
			j.monthly *= 0.9
			j.AddEvent(date, j.Sim().Name + " Job Loss", 7)
			start := date.AddDate(0, -1 * j.employed, 0)
			if j.Age(start) < j.Sim().RetirementAge() - 0.5 {
				j.AddEvent(start, j.Sim().Name + " New Job", 3)
				j.AddEvent(start, j.Sim().Name + " 10% Pay Cut", 5)
			} else {
				j.employed -= 12
			}
			return false
		}
		return true
	}
	if bc < -2.0 {
		if j.employed > int(math.Round(j.Gauss(24.0, 6.0))) {
			j.employed = int(math.Min(-1.0, math.Round(j.Gauss(-2.0, 2.0))))
			j.monthly *= 0.95
			j.AddEvent(date, j.Sim().Name + " Job Loss", 7)
			start := date.AddDate(0, -1 * j.employed, 0)
			if j.Age(start) < j.Sim().RetirementAge() - 0.5 {
				j.AddEvent(start, j.Sim().Name + " New Job", 3)
				j.AddEvent(start, j.Sim().Name + " 5% Pay Cut", 4)
			} else {
				j.employed -= 12
			}
			return false
		}
		return true
	}
	if bc < 2.0 {
		if j.employed > int(math.Round(j.Gauss(36.0, 6.0))) {
			j.employed = int(math.Min(-1.0, math.Round(j.Gauss(-2.0, 1.0))))
			j.AddEvent(date, j.Sim().Name + " Job Loss", 7)
			start := date.AddDate(0, -1 * j.employed, 0)
			if j.Age(start) < j.Sim().RetirementAge() - 0.5 {
				j.AddEvent(start, j.Sim().Name + " New Job", 3)
			} else {
				j.employed -= 12
			}
			return false
		}
		return true
	}
	if bc < 8.0 {
		if j.employed > int(math.Round(j.Gauss(42.0, 6.0))) {
			j.employed = int(math.Min(0.0, math.Round(j.Gauss(-2.0, 1.0))))
			j.AddEvent(date, j.Sim().Name + " Job Loss", 7)
			start := date.AddDate(0, -1 * j.employed, 0)
			if j.Age(start) < j.Sim().RetirementAge() - 0.5 {
				j.AddEvent(start, j.Sim().Name + " New Job", 3)
			} else {
				j.employed -= 12
			}
			return false
		}
		return true
	}
	if j.employed > int(math.Round(j.Gauss(42.0, 6.0))) {
		if j.Age(date) < j.Sim().RetirementAge() - 2.0 {
			j.employed = 0
			j.monthly *= 1.05
			j.AddEvent(date, j.Sim().Name + " Better Job", 3)
		}
	}
	return true
}

func (j *Job) Earn(date time.Time) float64 {
	if !j.Employed(date) {
		j.unemployed++
		return 0.0
	}
	j.unemployed = 0
	if j.employed == 0 {
		return j.monthly
	}
	if j.employed % 12 == 0 {
		j.monthly *= 1.01
	}
	return j.monthly
}

func (j *Job) Unemployment(date time.Time) float64 {
	if j.unemployed > 1 && j.unemployed <= 12 {
		return 2500.0
	}
	return 0.0
}

func (j *Job) Monthly(date time.Time) {
	amount := j.Earn(date)
	j.CashAccount().Deposit(amount, date, "Salary")
	j.TaxMen().Withhold(amount, date, true)
	unem := j.Unemployment(date)
	j.CashAccount().Deposit(unem, date, "Unemployment")
	j.TaxMen().Withhold(unem, date, false)
	j.employed++
}

