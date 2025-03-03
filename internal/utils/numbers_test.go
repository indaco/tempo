package utils

import (
	"math"
	"testing"
)

func TestSafeInt64ToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		want    int
		wantErr bool
	}{
		{"Valid Positive", 100, 100, false},
		{"Valid Negative", -100, -100, false},
		{"Zero", 0, 0, false},
		{"MaxInt", int64(math.MaxInt), int(math.MaxInt), false},
		{"MinInt", int64(math.MinInt), int(math.MinInt), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Int64ToInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeInt64ToInt(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("SafeInt64ToInt(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
