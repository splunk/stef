package tests

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/stretchr/testify/require"
)

type BarOutput struct {
	t       testing.TB
	title   string
	file    *os.File
	results map[string]map[string]float64
	enabled bool
}

func chartsEnabled() bool {
	return os.Getenv("UPDATE_BENCH_HTML") != ""
}

func (c *BarOutput) BeginChart(title string, t testing.TB) {
	c.t = t
	c.title = title
	c.results = map[string]map[string]float64{}
}

func (c *BarOutput) EndChart(unit string, globalopts ...charts.GlobalOpts) {
	if !c.enabled {
		return
	}

	var xAxis []string
	for label := range c.results {
		xAxis = append(xAxis, label)
	}
	slices.Sort(xAxis)

	itemsBySeries := map[string][]opts.BarData{}
	for _, label := range xAxis {
		values := c.results[label]
		for series, value := range values {
			itemsBySeries[series] = append(
				itemsBySeries[series], opts.BarData{Value: value, Label: &opts.Label{Show: opts.Bool(true)}},
			)
		}
	}

	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	globalopts = append(
		globalopts, charts.WithTitleOpts(
			opts.Title{
				Title: c.title,
			},
		),
		charts.WithAnimation(false),
		charts.WithYAxisOpts(
			opts.YAxis{
				Name: unit,
			},
		),
	)

	bar.SetGlobalOptions(globalopts...)

	// Put data into instance
	bars := bar.SetXAxis(xAxis)

	var seriesNames []string
	for seriesName := range itemsBySeries {
		seriesNames = append(seriesNames, seriesName)
	}
	slices.Sort(seriesNames)
	for _, seriesName := range seriesNames {
		bars = bars.AddSeries(
			seriesName,
			itemsBySeries[seriesName],
			charts.WithBarChartOpts(opts.BarChart{Stack: "stack"}),
		)
	}

	chartSnippet := bar.RenderSnippet()

	tmpl := "{{.Element}} {{.Script}}"
	t := template.New("snippet")
	t, err := t.Parse(tmpl)
	if err != nil {
		panic(err)
	}

	data := struct {
		Element template.HTML
		Script  template.HTML
		Option  template.HTML
	}{
		Element: template.HTML(chartSnippet.Element),
		Script:  template.HTML(chartSnippet.Script),
		Option:  template.HTML(chartSnippet.Option),
	}

	err = t.Execute(c.file, data)
	require.NoError(c.t, err)
}

func roundFloat(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

func (c *BarOutput) Record(b *testing.B, encoding string, series string, val float64) {
	if b != nil {
		b.ReportMetric(val, "ns/point")
	}
	c.results[encoding] = map[string]float64{series: val}
}

func (c *BarOutput) RecordStacked(b *testing.B, encoding string, series string, val float64) {
	if b != nil {
		b.ReportMetric(val, "ns/point")
	}
	if c.results[encoding] == nil {
		c.results[encoding] = map[string]float64{}
	}
	c.results[encoding][series] = math.Round(val)
}

func (c *BarOutput) Begin() {
	if !chartsEnabled() {
		return
	}
	c.enabled = true

	output, err := os.Create("../docs/benchmarks.html")
	if err != nil {
		panic(err)
	}
	c.file = output

	_, err = c.file.WriteString(
		`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Awesome go-echarts</title>
    <script src="https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js"></script>
</head>

<body>
`,
	)
	if err != nil {
		panic(err)
	}
}

func (c *BarOutput) End() {
	if !c.enabled {
		return
	}

	_, err := c.file.WriteString(`</body></html>`)
	if err != nil {
		panic(err)
	}

	err = c.file.Close()
	if err != nil {
		panic(err)
	}
}

func (c *BarOutput) BeginSection(s string) {
	if !c.enabled {
		return
	}

	_, err := c.file.WriteString(fmt.Sprintf("<div/><h2>%s</h2>\n", s))
	if err != nil {
		panic(err)
	}
}
