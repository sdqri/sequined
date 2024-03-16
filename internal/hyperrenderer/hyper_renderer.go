package hyperrenderer

import (
	"io"
)

type HyperRenderer interface {
	GetID() string
	GetPath() string
	Render(wr io.Writer) error
	GetLinks() []HyperRenderer
}

func GetPaths(node HyperRenderer) []string {
	paths := make([]string, 0)
	for _, page := range node.GetLinks() {
		paths = append(paths, page.GetPath())
	}
	return paths
}

type VisitedMap map[HyperRenderer]struct{}

// boolean value indicates whether the traversal should be terminated or not.
type VisitFunc func(HyperRenderer) bool

func Traverse(root HyperRenderer, f VisitFunc) VisitedMap {
	visited := make(VisitedMap)
	traverse(root, visited, f)
	return visited
}

func traverse(root HyperRenderer, visited VisitedMap, f VisitFunc) {
	if _, ok := visited[root]; ok {
		return
	}

	visited[root] = struct{}{}
	if shouldBreak := f(root); shouldBreak {
		return
	}

	links := root.GetLinks()
	for _, link := range links {
		traverse(link, visited, f)
	}
}

func NoOpVisit(hr HyperRenderer) bool {
	return false
}

func CreatePathMap(root HyperRenderer) map[string]HyperRenderer {
	nodeMap := make(map[string]HyperRenderer)
	Traverse(root, func(hr HyperRenderer) bool {
		nodeMap[hr.GetPath()] = hr
		return false
	})
	return nodeMap
}

func FindHyperRendererByPath(root HyperRenderer, path string) HyperRenderer {
	found := root
	Traverse(root, func(hr HyperRenderer) bool {
		if path == hr.GetPath() {
			found = hr
			return true
		}
		return false
	})
	return found
}

func FindHyperRendererByID(root HyperRenderer, targetID string) HyperRenderer {
	found := root
	Traverse(root, func(hr HyperRenderer) bool {
		if hr.GetID() == targetID {
			found = hr
			return true
		}
		return false
	})

	return found
}
