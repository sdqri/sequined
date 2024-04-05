package dashboard

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/sdqri/sequined/internal/dashboard/snippetrenderer"
	hyr "github.com/sdqri/sequined/internal/hyperrenderer"
	obs "github.com/sdqri/sequined/internal/observer"
)

//go:embed templates/dashboard.html.tmpl
var dashboardTmpl string

type Dashboard struct {
	observer *obs.Observer
	root     *hyr.Webpage
}

func NewDashboard(root *hyr.Webpage, observer *obs.Observer) *Dashboard {
	return &Dashboard{
		observer: observer,
		root:     root,
	}
}

func (dashboard *Dashboard) HandleBy(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard", dashboard.HandleMainPage)
	mux.HandleFunc("/charts/freshness", dashboard.HandleFreshnessChart)
	mux.HandleFunc("/charts/age", dashboard.HandleAgeChart)
	mux.HandleFunc("/charts/tree", dashboard.HandleTreeChart)
}

func (dashboard *Dashboard) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	dashboardTemplate, err := template.New("dashboard-template").Parse(dashboardTmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = dashboardTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (dashboard *Dashboard) GetFreshnessChart(bucketDuration time.Duration, duration time.Duration, ip string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Freshness - Last " + duration.String(),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			Min:  0,
			Max:  1,
		}),
	)

	now := time.Now().UTC()
	numBuckets := int(duration / bucketDuration)
	buckets := make([]time.Time, 0, numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets = append(buckets, now.Add(-time.Duration(i*int(bucketDuration))))
	}
	slices.Reverse(buckets)

	freshnessSeries := make([]opts.LineData, numBuckets)

	for i := 0; i < numBuckets; i++ {
		freshness := dashboard.observer.GetFreshness(ip, buckets[i])
		freshnessSeries[i] = opts.LineData{Value: freshness}
	}

	xs := ConvertToHHMMSS(buckets)
	line.SetXAxis(xs).
		AddSeries("Freshness", freshnessSeries).
		SetSeriesOptions(
			charts.WithLineChartOpts(
				opts.LineChart{
					Stack: "stack",
				}),
		)

	return line
}

func (dashboard *Dashboard) HandleFreshnessChart(w http.ResponseWriter, r *http.Request) {
	bucketDurationStr := r.URL.Query().Get("bucket-duration")
	durationStr := r.URL.Query().Get("duration")
	ip := r.URL.Query().Get("ip")

	bucketDuration, err := time.ParseDuration(bucketDurationStr)
	if err != nil {
		http.Error(w, "Invalid bucketDuration", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	freshnessChart := dashboard.GetFreshnessChart(bucketDuration, duration, ip)
	snippetRenderer := snippetrenderer.NewSnippetRenderer(freshnessChart, freshnessChart.Validate)
	err = snippetRenderer.Render(w)
	if err != nil {
		http.Error(w, "Failed to render charts", http.StatusInternalServerError)
		return
	}

}

func (dashboard *Dashboard) GetAgeChart(bucketDuration time.Duration, duration time.Duration, ip string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Age - Last " + duration.String(),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
		}),
	)

	now := time.Now().UTC()
	numBuckets := int(duration / bucketDuration)
	buckets := make([]time.Time, 0, numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets = append(buckets, now.Add(-time.Duration(i*int(bucketDuration))))
	}
	slices.Reverse(buckets)

	ageSeries := make([]opts.LineData, numBuckets)

	for i := 0; i < numBuckets; i++ {
		age := dashboard.observer.GetAge(ip, buckets[i])
		ageSeries[i] = opts.LineData{Value: age.Seconds()}
	}

	xs := ConvertToHHMMSS(buckets)
	line.SetXAxis(xs).
		AddSeries("Age (seconds)", ageSeries).
		SetSeriesOptions(
			charts.WithLineChartOpts(
				opts.LineChart{
					Stack: "stack",
				}),
		)

	return line
}

func (dashboard *Dashboard) HandleAgeChart(w http.ResponseWriter, r *http.Request) {
	bucketDurationStr := r.URL.Query().Get("bucket-duration")
	durationStr := r.URL.Query().Get("duration")
	ip := r.URL.Query().Get("ip")

	bucketDuration, err := time.ParseDuration(bucketDurationStr)
	if err != nil {
		http.Error(w, "Invalid bucketDuration", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	ageChart := dashboard.GetAgeChart(bucketDuration, duration, ip)
	err = ageChart.Render(w)
	if err != nil {
		http.Error(w, "Failed to render charts", http.StatusInternalServerError)
		return
	}
}

func (dashboard *Dashboard) A() *[]opts.TreeData {
	return GetTreeData(dashboard.root)
}

func GetTreeData(root *hyr.Webpage) *[]opts.TreeData {
	Result := make([]*opts.TreeData, 0)

	var f func(*hyr.Webpage, *[]*opts.TreeData)
	f = func(node *hyr.Webpage, treeData *[]*opts.TreeData) {
		children := make([]*opts.TreeData, 0)
		for _, child := range node.Links {
			f(child, &children)
		}

		td := opts.TreeData{
			Name: node.Faker().City(),
		}
		switch node.Type {
		case hyr.WebpageTypeHub:
			td.Symbol = "circle"
			td.ItemStyle = &opts.ItemStyle{
				Color: "#4CAF50", // Green for Hub
			}
		case hyr.WebpageTypeAuthority:
			td.Symbol = "rect"
			td.ItemStyle = &opts.ItemStyle{
				Color: "#2196F3", // Blue for Authority
			}
		}
		td.Children = children

		*treeData = append(*treeData, &td)
		fmt.Println(Result)
	}

	f(root, &Result)

	fmt.Println("Result=", Result)
	return &[]opts.TreeData{*Result[0]}
}

func (dashboard *Dashboard) GetTreeChart() *charts.Tree {
	tree := charts.NewTree()
	tree.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Tree Chart",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			// Add legend data for Hub and Authority
			Data: []string{"Hub", "Authority"},
		}),
	)

	treeData := GetTreeData(dashboard.root)
	tree.AddSeries("Root", *treeData).SetSeriesOptions(
		charts.WithTreeOpts(
			opts.TreeChart{
				Layout:           "orthogonal",
				Orient:           "TB",
				InitialTreeDepth: -1,
				Leaves: &opts.TreeLeaves{
					Label: &opts.Label{Show: true, Position: "right", Color: "Black", },
				},
			},
		),
		charts.WithLabelOpts(opts.Label{Show: true, Position: "top", Color: "Black"}),
	)

	return tree
}

func (dashboard *Dashboard) HandleTreeChart(w http.ResponseWriter, r *http.Request) {
	treeChart := dashboard.GetTreeChart()
	if err := treeChart.Render(w); err != nil {
		http.Error(w, "Failed to render charts", http.StatusInternalServerError)
		return
	}
}

func ConvertToHHMMSS(times []time.Time) []string {
	formattedTimes := make([]string, len(times))
	for i, t := range times {
		formattedTimes[i] = t.Format("15:04:05")
	}
	return formattedTimes
}
