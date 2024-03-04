package hyperrenderer

import (
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

var _ HyperRenderer = &Webpage{}

type WebpageType string

const (
	WebpageTypeAuthority = "authority"
	WebpageTypeHub       = "hub"
)

type Webpage struct {
	ID     uint64
	Path   string
	Parent *Webpage
	Links  []*Webpage
	Type   WebpageType

	PathGenerator PathGeneratorfunc
	AuthorityTmpl *template.Template
	HubTmpl       *template.Template
}

type PathGeneratorfunc func(parent *Webpage) string

type WebpageOption func(*Webpage)

func NewWebpage(
	webpageType WebpageType,
	opts ...WebpageOption,
) *Webpage {
	id := rand.Uint64()

	fnMaps := template.FuncMap{"Split": strings.Split}

	defaultAuthorityTmpl, err := template.New("default_authority.tmpl").
		Funcs(fnMaps).
		ParseFiles("internal/templates/default_authority.tmpl")
	if err != nil {
		panic(err)
	}
	defaultHubTmpl, err := template.New("default_hub.tmpl").
		Funcs(fnMaps).
		ParseFiles("internal/templates/default_hub.tmpl")
	if err != nil {
		panic(err)
	}

	webpage := Webpage{
		ID:    id,
		Links: make([]*Webpage, 0),
		Type:  webpageType,

		PathGenerator: defaultPageGenerator,
		AuthorityTmpl: defaultAuthorityTmpl,
		HubTmpl:       defaultHubTmpl,
	}

	for _, opt := range opts {
		opt(&webpage)
	}
	return &webpage
}

func (wp *Webpage) AddWebPage(webpageType WebpageType) *Webpage {
	webpage := *wp
	id := rand.Uint64()
	webpage.ID = id
	webpage.Links = make([]*Webpage, 0)
	webpage.Type = webpageType
	return &webpage
}

func (wp *Webpage) GetID() string {
	return fmt.Sprintf("%d", wp.ID)
}

func (wp *Webpage) GetPath() string {
	return wp.PathGenerator(wp)
}

func (wp *Webpage) Render(writer io.Writer) error {
	data := struct {
		Node *Webpage
	}{
		Node: wp,
	}
	if wp.Type == WebpageTypeAuthority {
		return wp.AuthorityTmpl.Execute(writer, data)
	}
	return wp.HubTmpl.Execute(writer, data)
}

func (wp *Webpage) GetLinks() []HyperRenderer {
	links := make([]HyperRenderer, len(wp.Links))
	for i, link := range wp.Links {
		links[i] = link
	}
	return links
}

func (wp *Webpage) AddLink(page *Webpage) {
	page.Parent = wp
	wp.Links = append(wp.Links, page)
}

func (wp *Webpage) Faker() *gofakeit.Faker {
	return gofakeit.New(wp.ID)
}

func defaultPageGenerator(webpage *Webpage) string {
	if webpage.Parent == nil {
		return "/"
	}
	result, err := url.JoinPath(webpage.Parent.GetPath(), fmt.Sprintf("%d", webpage.ID))
	if err != nil {
		panic(err)
	}
	return result
}

func (wp *Webpage) Draw(filename string, format graphviz.Format) error {
	// Create a new graph
	g := graphviz.New()

	// Create a new directed graph
	graph, err := g.Graph()
	if err != nil {
		return err
	}

	// Create a map to store nodes
	nodes := make(map[string]*cgraph.Node)
	visited := Traverse(wp, func(hr HyperRenderer) {})

	// Create nodes
	for renderer := range visited {
		webpage, ok := renderer.(*Webpage)
		if !ok {
			return fmt.Errorf("unable to type assert renderer to webpage")
		}
		// Create node for the webpage
		node, err := graph.CreateNode(renderer.GetID())
		if err != nil {
			return err
		}

		node.SetStyle(cgraph.FilledNodeStyle)
		if webpage.Type == WebpageTypeHub {
			node.SetFillColor("#99D19C")
		} else if webpage.Type == WebpageTypeAuthority {
			node.SetFillColor("#ADE1E5")
		}

		node.SetLabel(
			fmt.Sprintf(
				"Title: %s\nType: %s\nPath: %s",
				webpage.Faker().City(),
				string(webpage.Type),
				webpage.GetPath(),
			),
		)
		nodes[renderer.GetID()] = node
	}

	// Add edges
	for node := range visited {
		parentNode, ok := nodes[node.GetID()]
		if !ok {
			panic("unabe to get parentNode id")
		}

		for _, link := range node.GetLinks() {
			child, ok := nodes[link.GetID()]
			if !ok {
				panic("unable to get child id")
			}
			_, err := graph.CreateEdge("", parentNode, child)
			if err != nil {
				return err
			}
		}
	}

	// Render the graph to DOT format
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	err = g.Render(graph, format, file)
	if err != nil {
		return err
	}

	return nil
}

func WithAuthorityTemplate(tmpl *template.Template) WebpageOption {
	return func(w *Webpage) {
		w.AuthorityTmpl = tmpl
	}
}

func WithHubTemplate(tmpl *template.Template) WebpageOption {
	return func(w *Webpage) {
		w.HubTmpl = tmpl
	}
}
