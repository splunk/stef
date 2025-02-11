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
	results map[string]float64
}

func (c *BarOutput) BeginChart(title string, t testing.TB) {
	c.t = t
	c.title = title
	c.results = map[string]float64{}
}

func (c *BarOutput) EndChart(unit string, seriesName string, globalopts ...charts.GlobalOpts) {
	var xAxis []string
	for label := range c.results {
		xAxis = append(xAxis, label)
	}
	slices.Sort(xAxis)

	var items []opts.BarData
	for _, label := range xAxis {
		value := c.results[label]
		items = append(items, opts.BarData{Value: value})
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

	bar.SetXAxis(xAxis).AddSeries(seriesName, items)

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

func (c *BarOutput) Record(b *testing.B, encoding string, val float64) {
	if b != nil {
		b.ReportMetric(val, "ns/point")
	}
	c.results[encoding] = math.Round(val)
}

func (c *BarOutput) Begin() {
	output, err := os.Create("results/benchmarks.html")
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
	_, err := c.file.WriteString(fmt.Sprintf("<div/><h2>%s</h2>\n", s))
	if err != nil {
		panic(err)
	}
}
