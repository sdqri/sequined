package hyperrenderer_test

import (
	"io"
	"os"
	"testing"

	"github.com/sdqri/sequined/internal/hyperrenderer"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
}

var _ hr.HyperRenderer = &MockHyperRenderer{}

type MockHyperRenderer struct {
	ID    string
	Path  string
	Links []*MockHyperRenderer
}

func (m *MockHyperRenderer) GetID() string {
	return m.ID
}

func (m *MockHyperRenderer) GetPath() string {
	return m.Path
}

func (m *MockHyperRenderer) Render(wr io.Writer) error {
	return nil
}

func (m *MockHyperRenderer) GetLinks() []hr.HyperRenderer {
	links := make([]hr.HyperRenderer, len(m.Links))
	for i, link := range m.Links {
		links[i] = link
	}
	return links
}

func TestTraverse(t *testing.T) {
	testCases := []struct {
		Name           string
		Root           hr.HyperRenderer
		ExpectedVisits int
	}{
		{
			Name: "Single Node",
			Root: &MockHyperRenderer{
				ID:   "root",
				Path: "/root",
				Links: []*MockHyperRenderer{
					{
						ID:    "child",
						Path:  "/root/child",
						Links: nil,
					},
				},
			},
			ExpectedVisits: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			visitedMap := hr.Traverse(tc.Root, hr.NoOpVisit)
			assert.Len(t, visitedMap, tc.ExpectedVisits, "unexpected number of visited nodes")
		})
	}
}

func TestCreatePathMap(t *testing.T) {
	root := &MockHyperRenderer{
		ID:   "root",
		Path: "/root",
		Links: []*MockHyperRenderer{
			{
				ID:    "child1",
				Path:  "/root/child1",
				Links: nil,
			},
			{
				ID:   "child2",
				Path: "/root/child2",
				Links: []*MockHyperRenderer{
					{
						ID:    "grandchild1",
						Path:  "/root/child2/grandchild1",
						Links: nil,
					},
				},
			},
		},
	}

	pathMap := hyperrenderer.CreatePathMap(root)

	expectedMap := map[string]hyperrenderer.HyperRenderer{
		"/root":                    root,
		"/root/child1":             root.Links[0],
		"/root/child2":             root.Links[1],
		"/root/child2/grandchild1": root.Links[1].GetLinks()[0],
	}

	assert.Equal(t, len(expectedMap), len(pathMap), "length of pathMap does not match expected length")

	for path, expectedNode := range expectedMap {
		actualNode, ok := pathMap[path]
		assert.True(t, ok, "expected path %q missing in pathMap", path)
		assert.Equal(t, expectedNode, actualNode, "unexpected node for path %q", path)
	}
}

func TestFindHyperRendererByPath(t *testing.T) {
	root := &MockHyperRenderer{
		ID:   "root",
		Path: "/root",
		Links: []*MockHyperRenderer{
			{
				ID:    "child1",
				Path:  "/root/child1",
				Links: nil,
			},
			{
				ID:    "child2",
				Path:  "/root/child2",
				Links: nil,
			},
		},
	}

	found := hr.FindHyperRendererByPath(root, "/root/child1")
	assert.NotNil(t, found)
	assert.Equal(t, "child1", found.GetID())
}

func TestFindHyperRendererByID(t *testing.T) {
	root := &MockHyperRenderer{
		ID:   "root",
		Path: "/root",
		Links: []*MockHyperRenderer{
			{
				ID:    "child1",
				Path:  "/root/child1",
				Links: nil,
			},
			{
				ID:    "child2",
				Path:  "/root/child2",
				Links: nil,
			},
		},
	}

	found := hr.FindHyperRendererByID(root, "child2")
	assert.NotNil(t, found)
	assert.Equal(t, "child2", found.GetID())
}
