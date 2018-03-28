package sim

import (
	"math"
	"time"
)

type gaussianInputs struct {
	mean float64
	stddev float64
	min float64
	max float64
}

type nextRegimeGetter func(r *Regime) *Regime

type Regime struct {
	*Simulacrum
	durationArgs *gaussianInputs
	meanReturnArgs *gaussianInputs
	volatilityArgs *gaussianInputs
	nexter nextRegimeGetter
	duration int
	meanReturn float64
	vol float64
	returns []float64
	next *Regime
}

func newRegime(sim *Simulation, durArgs, mretArgs, volArgs *gaussianInputs, nexter nextRegimeGetter) *Regime {
	r := &Regime{
		Simulacrum: NewSimulacrum(sim),
		durationArgs: durArgs,
		meanReturnArgs: mretArgs,
		volatilityArgs: volArgs,
		nexter: nexter,
	}
	r.duration = int(r.ClampedGauss(
		r.durationArgs.mean,
		r.durationArgs.stddev,
		r.durationArgs.min,
		r.durationArgs.max,
	))
	r.meanReturn = r.ClampedGauss(
		r.meanReturnArgs.mean,
		r.meanReturnArgs.stddev,
		r.meanReturnArgs.min,
		r.meanReturnArgs.max,
	)
	r.vol = r.ClampedGauss(
		r.volatilityArgs.mean,
		r.volatilityArgs.stddev,
		r.volatilityArgs.min,
		r.volatilityArgs.max,
	)
	return r
}

func (r *Regime) Duration() int {
	return r.duration
}

func (r *Regime) MeanReturn() float64 {
	return r.meanReturn
}

func (r *Regime) Volatility() float64 {
	return r.vol
}

func (r *Regime) MarketReturns() []float64 {
	if r.returns == nil || len(r.returns) == 0 {
		r.returns = make([]float64, r.duration)
		for i := 0; i < r.duration; i++ {
			r.returns[i] = r.Gauss(r.meanReturn, r.vol)
		}
	}
	return r.returns
}

func (r *Regime) Last() float64 {
	rets := r.MarketReturns()
	return rets[len(rets) - 1]
}

func (r *Regime) Next() *Regime {
	if r.next == nil {
		r.next = r.nexter(r)
	}
	return r.next
}

func NewRecovery(sim *Simulation) *Regime {
	dur := &gaussianInputs{mean: 12.0, stddev: 6.0, min: 6.0, max: 18.0}
	mret := &gaussianInputs{mean: 12.0, stddev: 3.0, min: 5.0, max: math.MaxFloat64}
	vol := &gaussianInputs{mean: 3.0, stddev: 0.0, min: 3.0, max: 3.0}
	nexter := func(r *Regime) *Regime {
		return NewExpansion(r.Sim())
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

func NewExpansion(sim *Simulation) *Regime {
	dur := &gaussianInputs{mean: 72.0, stddev: 24.0, min: 48.0, max: 96.0}
	mret := &gaussianInputs{mean: 10.0, stddev: 0.0, min: 10.0, max: 10.0}
	vol := &gaussianInputs{mean: 2.0, stddev: 0.0, min: 2.0, max: 2.0}
	nexter := func(r *Regime) *Regime {
		if r.Duration() >= 30 {
			if r.RandTan(0.3, 10.0, 0.0, -1.0) > 0.5 {
				return NewBubble(r.Sim())
			}
		}
		return NewRecession(r.Sim())
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

func NewBubble(sim *Simulation) *Regime {
	dur := &gaussianInputs{mean: 12.0, stddev: 6.0, min: 2.0, max: math.MaxFloat64}
	mret := &gaussianInputs{mean: 15.0, stddev: 0.0, min: 15.0, max: 15.0}
	vol := &gaussianInputs{mean: 0.5, stddev: 0.0, min: 0.5, max: 0.5}
	nexter := func(r *Regime) *Regime {
		if r.RandTan(0.5, 10.0, 0.0, -1.0) > 0.75 {
			return NewDepression(r.Sim())
		}
		return NewRecession(r.Sim())
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

func NewRecession(sim *Simulation) *Regime {
	dur := &gaussianInputs{mean: 12.0, stddev: 6.0, min: 6.0, max: 24.0}
	mret := &gaussianInputs{mean: -5.0, stddev: 0.0, min: -5.0, max: -5.0}
	vol := &gaussianInputs{mean: 2.0, stddev: 0.0, min: 2.0, max: 2.0}
	nexter := func(r *Regime) *Regime {
		if r.Random() > 0.8 {
			return NewStagnation(r.Sim(), -1)
		}
		return NewRecovery(r.Sim())
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

func NewDepression(sim *Simulation) *Regime {
	dur := &gaussianInputs{mean: 12.0, stddev: 3.0, min: 6.0, max: 24.0}
	mret := &gaussianInputs{mean: -30.0, stddev: 5.0, min: -40.0, max: -20.0}
	vol := &gaussianInputs{mean: 5.0, stddev: 0.0, min: 5.0, max: 5.0}
	nexter := func(r *Regime) *Regime {
		sdur := int(r.ClampedGauss(48.0, 12.0, 36.0, 60.0))
		return NewStagnation(r.Sim(), sdur)
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

func NewStagnation(sim *Simulation, duration int) *Regime {
	var dur *gaussianInputs
	if duration <= 0 {
		dur = &gaussianInputs{mean: 24.0, stddev: 6.0, min: 12.0, max: 36.0}
	} else {
		fdur := float64(duration)
		dur = &gaussianInputs{mean: fdur, stddev: 0.0, min: fdur, max: fdur}
	}
	mret := &gaussianInputs{mean: 2.0, stddev: 0.0, min: 2.0, max: 2.0}
	vol := &gaussianInputs{mean: 1.0, stddev: 0.0, min: 1.0, max: 1.0}
	nexter := func(r *Regime) *Regime {
		return NewRecovery(r.Sim())
	}
	return newRegime(sim, dur, mret, vol, nexter)
}

type Economy struct {
	*Simulacrum
	root *Regime
	returns []float64
}

func NewEconomy(sim *Simulation) *Economy {
	return &Economy{
		Simulacrum: NewSimulacrum(sim),
		root: NewRecession(sim),
	}
}

func (e *Economy) MarketReturns() []float64 {
	if e.returns == nil {
		rets := make([]float64, 1200)
		r := e.root
		i := 0
		for i < 1200 {
			rrets := r.MarketReturns()
			for j := 0; j < len(rrets); j++ {
				if i + j >= 1200 {
					break
				}
				rets[i+j] = rrets[j]
			}
			i += len(rrets)
			r = r.Next()
		}
		e.returns = rets
	}
	return e.returns
}

func (e *Economy) MarketReturn(date time.Time) float64 {
	d := months(date.Sub(e.StartDate()))
	if d < 0 {
		return 0.0
	}
	mrets := e.MarketReturns()
	if d >= len(mrets) {
		return 0.0
	}
	return mrets[d]
}

func (e *Economy) Inflation(date time.Time) float64 {
	d := months(date.Sub(e.StartDate()))
	if d < 0 {
		return 0.0
	}
	if d < 6 {
		return 2.0
	}
	mrets := e.MarketReturns()
	if d >= len(mrets) {
		return 2.0
	}
	var s float64 = 0.0
	for i := d - 6; i < d; i++ {
		s += mrets[i] / 6.0
	}
	if s >= 10.0 {
		return 3.0
	}
	if s >= 7.5 {
		return 2.5
	}
	if s >= 4.0 {
		return 2.0
	}
	if s >= 0.0 {
		return 1.0
	}
	if s >= -4.0 {
		return 0.5
	}
	return 0.0
}

