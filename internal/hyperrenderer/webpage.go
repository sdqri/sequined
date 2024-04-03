package hyperrenderer

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/url"
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

type PathGeneratorfunc func(parent *Webpage) string

type Webpage struct {
	ID         uint64
	Path       string
	PathPrefix string
	Parent     *Webpage
	Links      []*Webpage
	Type       WebpageType

	PathGenerator PathGeneratorfunc
	AuthorityTmpl *template.Template
	HubTmpl       *template.Template
	CustomTmpl    *template.Template
}

type WebpageOption func(*Webpage)

//go:embed templates/*.tmpl
var templateFS embed.FS

func NewWebpage(
	webpageType WebpageType,
	opts ...WebpageOption,
) *Webpage {
	id := rand.Uint64()

	fnMaps := template.FuncMap{"Split": strings.Split}

	defaultAuthorityTmpl, err := template.New("default_authority.html.tmpl").
		Funcs(fnMaps).
		ParseFS(templateFS, "templates/default_authority.html.tmpl")
	if err != nil {
		panic(err)
	}
	defaultHubTmpl, err := template.New("default_hub.html.tmpl").
		Funcs(fnMaps).
		ParseFS(templateFS, "templates/default_hub.html.tmpl")
	if err != nil {
		panic(err)
	}

	webpage := Webpage{
		ID:    id,
		Links: make([]*Webpage, 0),
		Type:  webpageType,

		PathGenerator: defaultPathGenerator,
		AuthorityTmpl: defaultAuthorityTmpl,
		HubTmpl:       defaultHubTmpl,
	}

	for _, opt := range opts {
		opt(&webpage)
	}
	return &webpage
}

// Clone creates a new Webpage instance based on the current one with a specified type, a unique ID, and no links.
func (wp *Webpage) Clone(webpageType WebpageType) *Webpage {
	webpage := *wp
	// assign a unique id
	id := rand.Uint64()
	webpage.ID = id

	// initializing links & type
	webpage.Links = make([]*Webpage, 0)
	webpage.Type = webpageType
	return &webpage
}

func (wp *Webpage) AddChild(webpageType WebpageType, opts ...WebpageOption) *Webpage {
	newPage := wp.Clone(webpageType)
	wp.AddLink(newPage)

	for _, opt := range opts {
		opt(newPage)
	}
	return newPage
}

func (wp *Webpage) GetID() string {
	return fmt.Sprintf("%d", wp.ID)
}

func (wp *Webpage) GetPath() string {
	if wp.PathGenerator != nil {
		return wp.PathGenerator(wp)
	}
	return fmt.Sprintf("%s/%s", wp.PathPrefix, wp.Path)
}

func (wp *Webpage) Render(writer io.Writer) error {
	data := struct {
		Node *Webpage
	}{
		Node: wp,
	}
	if wp.CustomTmpl != nil {
		return wp.CustomTmpl.Execute(writer, data)
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

func (wp *Webpage) CountLinksByType(t WebpageType) int {
	i := 0
	for _, link := range wp.Links {
		if link.Type == t {
			i++
		}
	}
	return i
}

func (wp *Webpage) Draw(filename string, format graphviz.Format) error {
	g := graphviz.New()

	graph, err := g.Graph()
	if err != nil {
		return err
	}

	nodes := make(map[string]*cgraph.Node)
	visited := Traverse(wp, NoOpVisit)

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

func defaultPathGenerator(webpage *Webpage) string {
	if webpage.Parent == nil {
		if webpage.PathPrefix != "" {
			return webpage.PathPrefix
		}
		return "/"
	}
	result, err := url.JoinPath(webpage.Parent.GetPath(), fmt.Sprintf("%s", webpage.GetID()))
	if err != nil {
		panic(err)
	}
	return result
}

func CityPathGenerator(webpage *Webpage) string {
	if webpage.Parent == nil {
		return "/"
	}

	result, err := url.JoinPath(
		webpage.Parent.GetPath(),
		strings.ReplaceAll(strings.ToLower(webpage.Faker().City()), " ", "-"),
	)
	if err != nil {
		panic(err)
	}
	return result
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

func WithCustomTemplate(tmpl *template.Template) WebpageOption {
	return func(w *Webpage) {
		w.CustomTmpl = tmpl
	}
}

func WithPathGenerator(f func(*Webpage) string) WebpageOption {
	return func(w *Webpage) {
		w.PathGenerator = f
	}
}

func WithPathPrefix(prefix string) WebpageOption {
	return func(w *Webpage) {
		w.PathPrefix = strings.TrimSuffix(prefix, "/")
	}
}
