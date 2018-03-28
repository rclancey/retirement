package sim

import (
	"time"
)

var actuarialTable = []float64{
	0.005313, 0.000346, 0.000221, 0.000162, 0.000131,
	0.000116, 0.000106, 0.000098, 0.000091, 0.000086,
	0.000084, 0.000087, 0.000100, 0.000124, 0.000157,
	0.000194, 0.000232, 0.000269, 0.000304, 0.000338,
	0.000373, 0.000409, 0.000442, 0.000471, 0.000497,
	0.000524, 0.000553, 0.000582, 0.000611, 0.000641,
	0.000673, 0.000710, 0.000753, 0.000805, 0.000864,
	0.000932, 0.001005, 0.001082, 0.001160, 0.001243,
	0.001336, 0.001442, 0.001562, 0.001698, 0.001849,
	0.002014, 0.002195, 0.002402, 0.002639, 0.002903,
	0.003189, 0.003488, 0.003795, 0.004105, 0.004423,
	0.004775, 0.005153, 0.005528, 0.005893, 0.006266,
	0.006688, 0.007176, 0.007724, 0.008339, 0.009034,
	0.009832, 0.010740, 0.011754, 0.012881, 0.014141,
	0.015612, 0.017275, 0.019047, 0.020909, 0.022939,
	0.025297, 0.028045, 0.031131, 0.034582, 0.038467,
	0.043008, 0.048175, 0.053772, 0.059770, 0.066367,
	0.073828, 0.082382, 0.092183, 0.103305, 0.115746,
	0.129475, 0.144443, 0.160590, 0.177853, 0.196165,
	0.214677, 0.233091, 0.251082, 0.268304, 0.284403,
	0.301467, 0.319555, 0.338728, 0.359052, 0.380595,
	0.403431, 0.427637, 0.453295, 0.480492, 0.509322,
	0.539881, 0.572274, 0.606611, 0.643007, 0.681588,
	0.722483, 0.761882, 0.799976, 0.839975, 0.881973,
	1.0,
}

var riskFactors = map[string][]float64{
	"male": []float64{
		0.001009, 0.000050, 0.000061, 0.000050, 0.000055,
		0.000046, 0.000038, 0.000031, 0.000023, 0.000014,
		0.000009, 0.000014, 0.000036, 0.000081, 0.000142,
		0.000207, 0.000273, 0.000351, 0.000443, 0.000541,
		0.000646, 0.000742, 0.000810, 0.000838, 0.000838,
		0.000825, 0.000816, 0.000809, 0.000811, 0.000818,
		0.000825, 0.000826, 0.000823, 0.000811, 0.000797,
		0.000784, 0.000777, 0.000772, 0.000771, 0.000775,
		0.000787, 0.000810, 0.000851, 0.000913, 0.000996,
		0.001095, 0.001207, 0.001334, 0.001475, 0.001630,
		0.001798, 0.001985, 0.002202, 0.002455, 0.002736,
		0.003028, 0.003327, 0.003642, 0.003970, 0.004306,
		0.004666, 0.005026, 0.005337, 0.005581, 0.005785,
		0.005994, 0.006246, 0.006541, 0.006895, 0.007307,
		0.007768, 0.008274, 0.008838, 0.009465, 0.010160,
		0.010957, 0.011837, 0.012748, 0.013674, 0.014656,
		0.015703, 0.016906, 0.018367, 0.020142, 0.022162,
		0.024320, 0.026520, 0.028703, 0.030844, 0.032953,
		0.035050, 0.037157, 0.039294, 0.041478, 0.043721,
		0.045592, 0.047018, 0.047931, 0.048274, 0.048003,
		0.047560, 0.046923, 0.046074, 0.044990, 0.043649,
		0.042025, 0.040092, 0.037821, 0.035179, 0.032133,
		0.028647, 0.024680, 0.020191, 0.015135, 0.009461,
		0.003119, 0.000000, 0.000000, 0.000000, 0.000000,
		0.0,
	},
}

type Actuary struct {
	*Simulacrum
	BirthDate time.Time
	RiskFactors []string
	DeathDate time.Time
}

func NewActuary(sim *Simulation, birthDate time.Time, factors []string) *Actuary {
	a := &Actuary{
		Simulacrum: NewSimulacrum(sim),
		BirthDate: birthDate,
		RiskFactors: factors,
	}
	startDate := a.StartDate()
	var deathDate time.Time
	for i, r := range actuarialTable {
		deathDate = birthDate.AddDate(i, 0, 0)
		if deathDate.Before(startDate) {
			continue
		}
		for _, fs := range factors {
			f, ok := riskFactors[fs]
			if ok && len(f) > i {
				r += f[i]
			}
		}
		if a.Random() <= r {
			deathDate = deathDate.Add(time.Hour * time.Duration(int64(a.Random() * 365.25 * 24.0)))
			break
		}
	}
	a.DeathDate = deathDate
	return a
}

func (a *Actuary) Age(date time.Time) float64 {
	if date.After(a.DeathDate) {
		return a.Age(a.DeathDate)
	}
	return years(date.Sub(a.BirthDate))
}

func (a *Actuary) DeathRisk(date time.Time) float64 {
	age := int(a.Age(date))
	r := actuarialTable[age]
	for _, fs := range a.RiskFactors {
		f, ok := riskFactors[fs]
		if ok && len(f) > age {
			r += f[age]
		}
	}
	return r
}

func (a *Actuary) DeathAge() float64 {
	return a.Age(a.DeathDate)
}

