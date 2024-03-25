package graphgenerator_test

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sdqri/sequined/internal/graphgenerator"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
}

type RootGenerator func() *hr.Webpage

func CreateMockSelectByProbability() (graphgenerator.SelectorFunc, chan []float64) {
	probabilitiesChan := make(chan []float64, 1)
	return func(probabilities []float64) (int, error) {
		probabilitiesChan <- probabilities
		return graphgenerator.SelectByProbability(probabilities)
	}, probabilitiesChan
}

func TestCreateHubPage(t *testing.T) {
	testCases := []struct {
		name                   string
		rootGenerator          RootGenerator
		preferentialAttachment float64
		expectedProbabilites   []float64
	}{
		{
			name: "PreferentialAttachment=1",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				root.AddChild(hr.WebpageTypeHub)
				return root
			},
			preferentialAttachment: 1,
			expectedProbabilites:   []float64{1, 0},
		},
		{
			name: "PreferentialAttachment=0",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				root.AddChild(hr.WebpageTypeHub)
				return root
			},
			preferentialAttachment: 0,
			expectedProbabilites:   []float64{0.5, 0.5},
		},
		{
			name: "PreferentialAttachment=0.5",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				root.AddChild(hr.WebpageTypeHub)
				return root
			},
			preferentialAttachment: 0.5,
			expectedProbabilites:   []float64{0.75, 0.25},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gg := graphgenerator.New(tc.rootGenerator(), tc.preferentialAttachment)
			f, probabilitesChan := CreateMockSelectByProbability()
			gg.SelectorFunc = f
			webpage, err := gg.CreateHubPage()
			assert.NoError(t, err, "Error creating hub page")
			actualProbabilities := <-probabilitesChan
			sort.Sort(sort.Float64Slice(actualProbabilities))
			sort.Sort(sort.Float64Slice(tc.expectedProbabilites))
			assert.Equal(t, tc.expectedProbabilites, actualProbabilities)
			assert.NotNil(t, webpage)
			close(probabilitesChan)
		})
	}
}

func TestCreateAuthorityPage(t *testing.T) {
	testCases := []struct {
		name                   string
		rootGenerator          RootGenerator
		preferentialAttachment float64
		expectedProbabilites   []float64
	}{
		{
			name: "PreferentialAttachment=1",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				childHub := root.AddChild(hr.WebpageTypeHub)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				return root
			},
			preferentialAttachment: 1,
			expectedProbabilites:   []float64{0.75, 0.25},
		},
		{
			name: "PreferentialAttachment=0",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				childHub := root.AddChild(hr.WebpageTypeHub)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				return root
			},
			preferentialAttachment: 0,
			expectedProbabilites:   []float64{0.5, 0.5},
		},
		{
			name: "PreferentialAttachment=0.5",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				childHub := root.AddChild(hr.WebpageTypeHub)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				childHub.AddChild(hr.WebpageTypeAuthority)
				return root
			},
			preferentialAttachment: 0.5,
			expectedProbabilites:   []float64{0.375, 0.625},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gg := graphgenerator.New(tc.rootGenerator(), tc.preferentialAttachment)
			f, probabilitesChan := CreateMockSelectByProbability()
			gg.SelectorFunc = f
			webpage, err := gg.CreateAuthorityPage()
			assert.NoError(t, err, "Error creating hub page")
			actualProbabilities := <-probabilitesChan
			sort.Sort(sort.Float64Slice(actualProbabilities))
			sort.Sort(sort.Float64Slice(tc.expectedProbabilites))
			assert.Equal(t, tc.expectedProbabilites, actualProbabilities)
			assert.NotNil(t, webpage)
			close(probabilitesChan)
		})
	}
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name                   string
		rootGenerator          RootGenerator
		preferentialAttachment float64
		maxHubCount            int
		maxAuthCount           int
		expectedError          error
	}{
		{
			name: "nomral generate",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			preferentialAttachment: 0.5,
			maxHubCount:            10,
			maxAuthCount:           10,
			expectedError:          nil,
		},
		{
			name: "exceeded max generate",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				root.AddChild(hr.WebpageTypeAuthority).AddChild(hr.WebpageTypeAuthority)
				root.AddChild(hr.WebpageTypeHub).AddChild(hr.WebpageTypeHub)
				return root
			},
			preferentialAttachment: 0.5,
			maxHubCount:            2,
			maxAuthCount:           2,
			expectedError:          graphgenerator.ErrMaxHubOrAuthCountAlreadyExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.rootGenerator()
			gg := graphgenerator.New(root, tc.preferentialAttachment)
			err := gg.Generate(tc.maxHubCount, tc.maxAuthCount)
			assert.Equal(t, err, tc.expectedError, "Unexpected error while calling gg.Generate")
			if tc.expectedError == nil {
				hubCount, authCount := 0, 0
				hr.Traverse(root, func(currentRenderer hr.HyperRenderer) bool {
					currentPage, ok := currentRenderer.(*hr.Webpage)
					if !ok {
						t.Errorf("Unable to traverse because non Webpage node")
						return true
					}

					if currentPage.Type == hr.WebpageTypeAuthority {
						authCount++
					} else if currentPage.Type == hr.WebpageTypeHub {
						hubCount++
					}
					return false
				})
				assert.Equal(t, tc.maxHubCount, hubCount)
				assert.Equal(t, tc.maxAuthCount, authCount)
			}
		})
	}
}

func TestStartGraphEvolution(t *testing.T) {
	testCases := []struct {
		name                       string
		rootGenerator              RootGenerator
		preferentialAttachment     float64
		maxHubCount                int
		maxAuthCount               int
		expectedError              error
		expectedCountUpdateMessage int
		waitFor                    time.Duration
	}{
		{
			name: "nomral generate",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			preferentialAttachment:     0.5,
			maxHubCount:                10,
			maxAuthCount:               5,
			expectedError:              nil,
			expectedCountUpdateMessage: 9 + 5,
			waitFor:                    (9 + 5) * 2 * (time.Hour / 1_000_000),
		},
		{
			name: "big generate",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			preferentialAttachment:     0.5,
			maxHubCount:                100,
			maxAuthCount:               5000,
			expectedError:              nil,
			expectedCountUpdateMessage: 100 + 5000,
			waitFor:                    (100 + 5000) * 2 * (time.Hour / 1_000_000),
		},
		{
			name: "exceeded max generate",
			rootGenerator: func() *hr.Webpage {
				root := hr.NewWebpage(hr.WebpageTypeHub)
				root.AddChild(hr.WebpageTypeAuthority).AddChild(hr.WebpageTypeAuthority)
				root.AddChild(hr.WebpageTypeHub).AddChild(hr.WebpageTypeHub)
				return root
			},
			preferentialAttachment: 0.5,
			maxHubCount:            2,
			maxAuthCount:           2,
			expectedError:          graphgenerator.ErrMaxHubOrAuthCountAlreadyExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.rootGenerator()
			gg := graphgenerator.New(root, tc.preferentialAttachment)
			updateChan, errChan, err := gg.StartGraphEvolution(
				tc.maxHubCount, tc.maxAuthCount,
				1_000_000, 1_000_000,
			)
			assert.Equal(t, err, tc.expectedError, "Unexpected error while calling gg.Generate")
			if tc.expectedError == nil {
				countUpdateMessage := 0
			outerLoop:
				for {
					select {
					case <-updateChan:
						countUpdateMessage++
						if countUpdateMessage == tc.expectedCountUpdateMessage {
							assert.Equal(t, countUpdateMessage, tc.expectedCountUpdateMessage)
							break outerLoop
						}
					case err, ok := <-errChan:
						if ok {
							assert.FailNow(t, fmt.Sprintf("Received an unexpected error from errChan, err:%s", err.Error()))
							break outerLoop
						}
					case <-time.After(tc.waitFor):
						assert.Error(t, errors.Errorf("Expected number of updateMessage not reached in wait time"))
						break outerLoop
					}
				}
				hubCount, authCount := 0, 0
				hr.Traverse(root, func(currentRenderer hr.HyperRenderer) bool {
					currentPage, ok := currentRenderer.(*hr.Webpage)
					if !ok {
						t.Errorf("Unable to traverse because non Webpage node")
						return true
					}

					if currentPage.Type == hr.WebpageTypeAuthority {
						authCount++
					} else if currentPage.Type == hr.WebpageTypeHub {
						hubCount++
					}
					return false
				})
				assert.Equal(t, tc.maxHubCount, hubCount)
				assert.Equal(t, tc.maxAuthCount, authCount)

			}
		})
	}
}
