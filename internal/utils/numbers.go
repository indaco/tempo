package utils

import (
	"fmt"
	"math"
)

func Int64ToInt(i64 int64) (int, error) {
	if i64 > math.MaxInt || i64 < math.MinInt {
		return 0, fmt.Errorf("int64 value %d overflows int", i64)
	}
	return int(i64), nil
}
