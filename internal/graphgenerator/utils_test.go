package graphgenerator_test

import (
	"testing"

	"github.com/sdqri/sequined/internal/graphgenerator"
	"github.com/stretchr/testify/assert"
)

func TestSelectByProbability(t *testing.T) {
	testCases := []struct {
		name          string
		probabilities []float64
	}{
		{
			name:          "Empty Probabilities",
			probabilities: []float64{},
		},
		{
			name:          "Zero Sum Probabilities",
			probabilities: []float64{0, 0, 0},
		},
		{
			name:          "Equal Probabilities",
			probabilities: []float64{0.25, 0.25, 0.25, 0.25},
		},
		{
			name:          "Unequal Probabilities",
			probabilities: []float64{0.1, 0.3, 0.6},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			selectedIdx, err := graphgenerator.SelectByProbability(test.probabilities)
			if test.probabilities == nil {
				assert.ErrorIs(t, err, graphgenerator.ErrNoProbabilities, "Unexpected error for empty probabilities")
			} else if len(test.probabilities) == 0 || sum(test.probabilities) == 0 {
				assert.ErrorIs(t, err, graphgenerator.ErrZeroSumProbabilities, "Unexpected error for zero sum probabilities")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.GreaterOrEqualf(t, selectedIdx, 0, "Selected index should be greater than or equal to -1")
				assert.LessOrEqualf(t, selectedIdx, len(test.probabilities)-1, "Selected index should be less than or equal to the maximum index")
			}
		})
	}
}

func sum(nums []float64) float64 {
	total := 0.0
	for _, num := range nums {
		total += num
	}
	return total
}
