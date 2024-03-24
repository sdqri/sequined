package graphgenerator_test

import (
	"os"
	"sort"
	"testing"

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
	}{
		{
			name: "case1",
			rootGenerator: func() *hr.Webpage {
				return hr.NewWebpage(hr.WebpageTypeHub)
			},
			preferentialAttachment: 0.5,
			maxHubCount:            10,
			maxAuthCount:           10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := tc.rootGenerator()
			gg := graphgenerator.New(root, tc.preferentialAttachment)
			err := gg.Generate(tc.maxHubCount, tc.maxAuthCount)
			assert.Nil(t, err, "Error while calling gg.Generate")
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
		})
	}
}
