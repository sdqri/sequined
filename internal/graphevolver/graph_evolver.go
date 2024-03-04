package graphevolver

import "sequined/internal/hyperrenderer"

type GraphEvolver struct {
	Root *hyperrenderer.Webpage
}

func New(root *hyperrenderer.Webpage) *GraphEvolver {
	return &GraphEvolver{}
}
