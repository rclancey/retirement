package sim

import (
	"math"
	"time"
)

func years(d time.Duration) float64 {
	return d.Hours() / (24.0 * 365.25)
}

func months(d time.Duration) int {
	return int(math.Round(12.0 * years(d)))
}

func days(d time.Duration) int {
	return int(math.Round(d.Hours() / 24.0))
}

func startOfYear(d time.Time) time.Time {
	return time.Date(d.Year(), time.January, 1, 0, 0, 0, 0, d.Location())
}

func startOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

func endOfMonth(d time.Time) time.Time {
	nxt := startOfMonth(d).Add(time.Hour * 24 * 32)
	return startOfMonth(nxt).Add(-1 * time.Millisecond)
}

func endOfYear(d time.Time) time.Time {
	nxt := time.Date(d.Year() + 1, time.January, 1, 0, 0, 0, 0, d.Location())
	return nxt.Add(-1 * time.Millisecond)
}

func incrementMonth(d time.Time) time.Time {
	return startOfMonth(startOfMonth(d).Add(time.Hour * 24 * 32))
}

func round2(v float64) float64 {
	return math.Round(v * 100.0) / 100.0
}

