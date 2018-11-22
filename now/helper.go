package now

import "math"

// NanoToMs convert ns to ms, with .2 fraction
func NanoToMs(ns int64) float64 {
	return math.Trunc(float64(ns)/float64(1000000)*1e2+0.5) * 1e-2
}
