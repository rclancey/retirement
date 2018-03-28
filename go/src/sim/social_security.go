package sim

import (
	"time"
)

type SocialSecurity struct {
	*Simulacrum
	age float64
	payout float64
}

func NewSocialSecurity(sim *Simulation, age, payout float64) *SocialSecurity {
	return &SocialSecurity{
		Simulacrum: NewSimulacrum(sim),
		age: age,
		payout: payout,
	}
}

func (s *SocialSecurity) BenefitsDate() time.Time {
	yf := s.age
	yi := int(yf)
	d := int(365.25 * (yf - float64(yi)))
	return s.Sim().Actuary.BirthDate.AddDate(yi, 0, d)
}

func (s *SocialSecurity) Earn(date time.Time) float64 {
	if s.Age(date) < s.age {
		return 0.0
	}
	return s.payout
}

func (s *SocialSecurity) Monthly(date time.Time) {
	amount := s.Earn(date)
	s.CashAccount().Deposit(amount, date, "Social Security")
	s.TaxMen().Withhold(amount, date, false)
}

