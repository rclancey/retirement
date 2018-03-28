package sim

import (
	"time"
)

var iraLifeExpectancies = []float64{
	82.4, 81.6, 80.6, 79.7, 78.7, 77.7, 76.7, 75.8, 74.8, 73.8,
	72.8, 71.8, 70.8, 69.9, 68.9, 67.9, 66.9, 66.0, 65.0, 64.0,
	63.0, 62.1, 61.1, 60.1, 59.1, 58.2, 57.2, 56.2, 55.3, 54.3,
	53.3, 52.4, 51.4, 50.4, 49.4, 48.5, 47.5, 46.5, 45.6, 44.6,
	43.6, 42.7, 41.7, 40.7, 39.8, 38.8, 37.9, 37.0, 36.0, 35.1,
	34.2, 33.3, 32.3, 31.4, 30.5, 29.6, 28.7, 27.9, 27.0, 26.1,
	25.2, 24.4, 23.5, 22.7, 21.8, 21.0, 20.2, 19.4, 18.6, 17.8,
	17.0, 16.3, 15.5, 14.8, 14.1, 13.4, 12.7, 12.1, 11.4, 10.8,
	10.2,  9.7,  9.1,  8.6,  8.1,  7.6,  7.1,  6.7,  6.3,  5.9,
	 5.5,  5.2,  4.9,  4.6,  4.3,  4.1,  3.8,  3.6,  3.4,  3.1,
	 2.9,  2.7,  2.5,  2.3,  2.1,  1.9,  1.7,  1.5,  1.4,  1.2,
	 1.1,  1.0,
}

type RMD func(date time.Time, acct Account) float64
type WithdrawRule func(date time.Time) bool
type DepositRule func(date time.Time) bool

func InheritedIRARMD(sim *Simulation, inheritanceDate time.Time) RMD {
	basisDate := endOfYear(endOfYear(inheritanceDate).Add(24 * time.Hour))
	basisAge := int(sim.Actuary.Age(basisDate))
	if basisAge < 0 {
		basisAge = 0
	} else if basisAge >= len(iraLifeExpectancies) {
		basisAge = len(iraLifeExpectancies) - 1
	}
	return func(date time.Time, acct Account) float64 {
		elapsed := float64(date.Year() - basisDate.Year())
		basis := iraLifeExpectancies[basisAge] - elapsed
		yeBal := acct.YearEndBalance(date)
		bal := acct.Balance()
		if basis < 1.0 {
			if yeBal > bal {
				return bal
			}
			return yeBal
		}
		rmd := yeBal / basis
		if rmd > bal {
			return bal
		}
		return rmd
	}
}

func InheritedIRACanWithdraw(sim *Simulation) WithdrawRule {
	return func(date time.Time) bool {
		return true
	}
}

func InheritedIRACanDeposit(sim *Simulation) DepositRule {
	return func(date time.Time) bool {
		return false
	}
}

var k401LifeExpectancies = []float64{
	27.4, 26.5, 25.6, 24.7, 23.8, 22.9, 22.0, 21.2, 20.3, 19.5, 18.7, 17.9,
	17.1, 16.3, 15.5, 14.8, 14.1, 13.4, 12.7, 12.0, 11.4, 10.8, 10.2,  9.6,
	 9.1,  8.6,  8.1,  7.6,  7.1,  6.7,  6.3,  5.9,  5.5,  5.2,  4.9,  4.5,
	 4.2,  3.9,  3.7,  3.4,  3.1,  2.9,  2.6,  2.4,  2.1,  1.9,
}

func K401RMD(sim *Simulation) RMD {
	return func(date time.Time, acct Account) float64 {
		age := sim.Actuary.Age(endOfYear(date))
		if age < 70.5 {
			return 0.0
		}
		n := int(age) - 70
		if n >= len(k401LifeExpectancies) {
			n = len(k401LifeExpectancies) - 1
		}
		denom := k401LifeExpectancies[n]
		yeBal := acct.YearEndBalance(date)
		bal := acct.Balance()
		rmd := yeBal / denom
		if rmd > bal {
			return bal
		}
		return rmd
	}
}

func K401CanWithdraw(sim *Simulation) WithdrawRule {
	return func(date time.Time) bool {
		return sim.Actuary.Age(date) >= 59.5
	}
}

func K401CanDeposit(sim *Simulation) DepositRule {
	return func(date time.Time) bool {
		return sim.Actuary.Age(date) < 70.5
	}
}

