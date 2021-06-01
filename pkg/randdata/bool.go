package randdata

import "math/rand"

// Bool generates a randomized boolean value
func Bool() bool {
	return rand.Float32() > 0.5
}
