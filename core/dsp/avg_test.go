package dsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlidingMax(t *testing.T) {
	tt := []struct {
		name     string
		length   int
		values   []float64
		expected []int
	}{
		{"empty", 0, []float64{}, []int{}},
		{"ascending order", 10, []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"window 3, peaks", 3, []float64{0, 1, 2, 1, 2}, []int{0, 1, 2, 2, 4}},
	}

	for _, tc := range tt {
		max := newSlidingMax(tc.length)
		t.Run(tc.name, func(t *testing.T) {
			actual := make([]int, len(tc.expected))
			for i, v := range tc.values {
				actual[i] = max.Put(v)
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}
