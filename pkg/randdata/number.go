package randdata

import "math/rand"

// Integer generates a randomized integer value
func Integer(maximum int64) int64 {
	if maximum == 0 {
		maximum = 100
	}
	return rand.Int63n(maximum)
}

// Float generates a randomized float value
func Float() float64 {
	return rand.Float64()
}
