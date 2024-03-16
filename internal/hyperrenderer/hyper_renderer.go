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

type Visited map[HyperRenderer]struct{}

func Traverse(root HyperRenderer, f func(HyperRenderer)) Visited {
	visited := make(Visited)
	traverse(root, visited, f)
	return visited
}

func traverse(root HyperRenderer, visited Visited, f func(HyperRenderer)) {
	if _, ok := visited[root]; ok {
		return
	}

	visited[root] = struct{}{}
	f(root)

	links := root.GetLinks()
	for _, link := range links {
		traverse(link, visited, f)
	}
}

func CreateMap(root HyperRenderer) map[string]HyperRenderer {
	nodeMap := make(map[string]HyperRenderer)
	// CreateMapHelper(root, visited, func)
	Traverse(root, func(hr HyperRenderer) {
		nodeMap[hr.GetPath()] = hr
	})
	return nodeMap
}

func FindHyperRendererByPath(root HyperRenderer, path string) HyperRenderer {
	target := root
	Traverse(root, func(hr HyperRenderer) {
		if path == hr.GetPath() {
			target = hr
			return
		}
	})
	return target
}

// func findHyperRendererByID(root HyperRenderer, targetID string, visited Visited) HyperRenderer {
// 	if root.GetID() == targetID {
// 		return root
// 	}
//
// 	visited[root.GetID()] = true
// 	links := root.GetLinks()
// 	for _, link := range links {
// 		if !visited[link.GetID()] {
// 			if node := findHyperRendererByID(link, targetID, visited); node != nil {
// 				return node
// 			}
// 		}
// 	}
// 	return nil
// }
