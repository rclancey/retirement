package sim

import (
	"math"
	"math/rand"
	"time"
)

type Simulacrum struct {
	sim *Simulation
	rng *rand.Rand
}

func NewSimulacrum(sim *Simulation) *Simulacrum {
	rng, _ := sim.Rng()
	return &Simulacrum{
		sim: sim,
		rng: rng,
	}
}

func (s *Simulacrum) Sim() *Simulation {
	return s.sim
}

func (s *Simulacrum) Rng() *rand.Rand {
	return s.rng
}

func (s *Simulacrum) StartDate() time.Time {
	return s.Sim().StartDate()
}

func (s *Simulacrum) CashAccount() *CashAccount {
	return s.Sim().CashAccount
}

func (s *Simulacrum) TaxMen() *TaxMen {
	return s.Sim().TaxMen
}

func (s *Simulacrum) Age(date time.Time) float64 {
	return s.Sim().Actuary.Age(date)
}

func (s *Simulacrum) Retired(date time.Time) bool {
	return s.Sim().Actuary.Age(date) >= s.Sim().RetirementAge()
}

func (s *Simulacrum) AddEvent(date time.Time, value string, severity int) {
	s.Sim().Events.Add(date, value, severity)
}

func (s *Simulacrum) ClampedGauss(mean, stddev, min_val, max_val float64) float64 {
	v := s.Gauss(mean, stddev)
	if v < min_val {
		return min_val
	}
	if v > max_val {
		return max_val
	}
	return v
}

func (s *Simulacrum) Gauss(mean, stddev float64) float64 {
	if stddev == 0.0 {
		return mean
	}
	return s.Rng().NormFloat64() * stddev + mean
}

func (s *Simulacrum) Random() float64 {
	return s.Rng().Float64()
}

func (s *Simulacrum) RandomRange(min_val, max_val float64) float64 {
	return s.Random() * (max_val - min_val) + min_val
}

func (s *Simulacrum) RandTan(median, squish, slope, offset float64) float64 {
	if slope == 0.0 {
		slope = math.Pi / (math.Atan((1.0 - median) * squish) - math.Atan(-1.0 * median * squish))
	}
	if offset < 0.0 {
		offset = -1.0 * slope * (math.Atan(-1.0 * median * squish) + (math.Pi / 2.0)) / math.Pi
	}
	r := s.Random()
	return slope * (math.Atan((r - median) * squish) + math.Pi / 2.0) / math.Pi + offset
}

