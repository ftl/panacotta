package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBRange_Width(t *testing.T) {
	tt := []struct {
		from     DB
		to       DB
		expected DB
	}{
		{10, -180, 190},
		{-180, 10, 190},
		{-180, 0, 180},
		{0, 30, 30},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := DBRange{tc.from, tc.to}.Width()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDBRange_ToFrct(t *testing.T) {
	tt := []struct {
		from     DB
		to       DB
		value    DB
		expected Frct
	}{
		{-80, 20, -90, -0.1},
		{-80, 20, -80, 0.0},
		{-80, 20, -60, 0.2},
		{-80, 20, 0, 0.8},
		{-80, 20, 10, 0.9},
		{-80, 20, 30, 1.1},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			actual := ToDBFrct(tc.value, DBRange{tc.from, tc.to})
			assert.Equal(t, tc.expected, actual)
		})
	}
}
