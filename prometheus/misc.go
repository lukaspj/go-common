package prometheus

import (
	"time"
)

func timeToFloat64(v time.Time) float64 {
	return float64(v.Unix())
}
