package graphgenerator

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrNoProbabilities      = errors.New("no probabilities provided")
	ErrZeroSumProbabilities = errors.New("sum of probabilities is zero")
	ErrFailedToSelectIndex  = errors.New("failed to select index")
)

func SelectByProbability(probabilities []float64) (int, error) {
	if len(probabilities) == 0 {
		return -1, ErrNoProbabilities
	}

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Calculate the sum of probabilities
	var sum float64 = 0
	for _, prob := range probabilities {
		sum += prob
	}

	if sum == 0 {
		return -1, ErrZeroSumProbabilities
	}

	// Generate a random number between 0 and the sum of probabilities
	r := rng.Float64() * sum

	var cumulativeProb float64 = 0
	for i, prob := range probabilities {
		cumulativeProb += prob
		if r < cumulativeProb {
			return i, nil
		}
	}

	return -1, ErrFailedToSelectIndex
}
