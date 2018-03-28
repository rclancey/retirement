package sim

import (
	"time"
)

type Portfolio struct {
	*Simulacrum
	speculative float64
	aggressive float64
	moderate float64
	conservative float64
	lastDate *time.Time
	lastRet *float64
}

func NewPortfolio(sim *Simulation, speculative, aggressive, moderate, conservative float64) *Portfolio {
	return &Portfolio{
		Simulacrum: NewSimulacrum(sim),
		speculative: speculative,
		aggressive: aggressive,
		moderate: moderate,
		conservative: conservative,
	}
}

func (p *Portfolio) riskTarget(date time.Time) float64 {
	yearsToRetirement := p.Sim().RetirementAge() - p.Age(date)
	if yearsToRetirement <= 0.0 {
		return p.conservative
	}
	if yearsToRetirement <= 5.0 {
		return p.moderate
	}
	if yearsToRetirement <= 40.0 {
		return p.aggressive
	}
	return p.speculative
}

func (p *Portfolio) PortfolioReturn(date time.Time) float64 {
	risk := p.riskTarget(date)
	if p.lastDate != nil && p.lastRet != nil && date.Equal(*p.lastDate) {
		return *p.lastRet * (1.0 + (p.Gauss(0.0, risk * 0.05) / 12.0))
	}
	mkt := p.Sim().Economy.MarketReturn(date)
	mean := risk * mkt / 10.0
	stddev := risk * 0.25
	ret := p.Gauss(mean, stddev) / 12.0
	p.lastDate = &date
	p.lastRet = &ret
	return ret
}

