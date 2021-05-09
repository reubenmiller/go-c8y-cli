package randdata

import (
	"math"
	"math/rand"
)

// Integer generates a randomized integer value. Return a random value from min to max-1
func Integer(max int64, min int64) int64 {
	if max < min {
		max, min = min, max
	}

	if max-1-min <= 0 {
		max = max + 1
	}
	return min + rand.Int63n(max-1-min+1)
}

// Float generates a randomized float value
func Float(max float64, min float64, precision int) float64 {
	if precision < 1 {
		precision = 1
	}
	factor := math.Pow10(precision)
	val := min + rand.Float64()*(max-min)
	return math.Round(val*factor) / factor
}
