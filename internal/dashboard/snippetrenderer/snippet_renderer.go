package snippetrenderer

import (
	"bytes"
	"io"
	"regexp"

	_ "embed"

	render "github.com/go-echarts/go-echarts/v2/render"
)

var pat = regexp.MustCompile(`(__f__")|("__f__)|(__f__)`)

//go:embed base.html.tmpl
var BaseTmpl string

type SnippetRenderer struct {
	c      interface{}
	before []func()
}

func NewSnippetRenderer(c interface{}, before ...func()) render.Renderer {
	return &SnippetRenderer{c: c, before: before}
}

func (r *SnippetRenderer) Render(w io.Writer) error {

	content := r.RenderContent()
	_, err := w.Write(content)
	return err
}

func (r *SnippetRenderer) RenderContent() []byte {
	for _, fn := range r.before {
		fn()
	}

	contents := []string{BaseTmpl}
	tpl := render.MustTemplate(render.ModChart, contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, render.ModChart, r.c); err != nil {
		panic(err)
	}

	return pat.ReplaceAll(buf.Bytes(), []byte(""))
}
