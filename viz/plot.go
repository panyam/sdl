// Package viz provides plotting and diagramming utilities.
package viz

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)


// PlotMetadata contains chart labels and title.
type PlotMetadata struct {
	XLabel string `json:"xLabel,omitempty"`
	YLabel string `json:"yLabel,omitempty"`
	Title  string `json:"title,omitempty"`
}

// MultiSeriesPlotData represents a complete dataset for a multi-series plot.
type MultiSeriesPlotData struct {
	Series   []DataSeries
	Metadata PlotMetadata
}

// PlotConfig holds styling and dimension configuration.
type PlotConfig struct {
	Width        int
	Height       int
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int
	GridColor    string
	TextColor    string
	YAxisMode    YAxisMode // Y-axis scaling mode
	Colors       []string  // Palette for multi-series plots
}

// TemplateData contains all data needed for SVG template rendering.
type TemplateData struct {
	Config      PlotConfig
	Metadata    PlotMetadata
	InnerWidth  int
	InnerHeight int
	XTicks      []XTick
	YTicks      []YTick
	GridLines   []GridLine
	SeriesPaths []SeriesPath
	LegendItems []LegendItem
}

// Helper structs for template rendering
type XTick struct{ X, Y int; Label string }
type YTick struct{ X, Y int; Label string }
type GridLine struct{ X1, Y1, X2, Y2 int }
type SeriesPath struct{ Path, Color string }
type LegendItem struct{ Name, Color string; X, Y int }

// SVG template with multi-series and legend support.
const svgTemplate = `<svg width="{{.Config.Width}}" height="{{.Config.Height}}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <style>
      .axis { font: 12px sans-serif; fill: {{.Config.TextColor}}; }
      .axis path, .axis line { fill: none; stroke: {{.Config.TextColor}}; shape-rendering: crispEdges; }
      .grid-line { stroke: {{.Config.GridColor}}; stroke-width: 0.5px; }
      .title { font: bold 16px sans-serif; text-anchor: middle; fill: {{.Config.TextColor}}; }
      .axis-label { font: 12px sans-serif; text-anchor: middle; fill: {{.Config.TextColor}}; }
      .legend { font: 12px sans-serif; fill: {{.Config.TextColor}}; }
    </style>
  </defs>

  {{if .Metadata.Title}}
  <text class="title" x="{{div .Config.Width 2}}" y="20">{{.Metadata.Title}}</text>
  {{end}}

  <g transform="translate({{.Config.MarginLeft}},{{.Config.MarginTop}})">
    <!-- Grid Lines -->
    {{range .GridLines}}<line class="grid-line" x1="{{.X1}}" x2="{{.X2}}" y1="{{.Y1}}" y2="{{.Y2}}"></line>{{end}}

    <!-- X Axis -->
    <g class="axis" transform="translate(0,{{.InnerHeight}})">
      {{range .XTicks}}<line x1="{{.X}}" x2="{{.X}}" y1="0" y2="6"></line><text x="{{.X}}" y="20" text-anchor="middle">{{.Label}}</text>{{end}}
      <path d="M0,0H{{$.InnerWidth}}"></path>
      {{if .Metadata.XLabel}}<text class="axis-label" x="{{div .InnerWidth 2}}" y="35">{{.Metadata.XLabel}}</text>{{end}}
    </g>

    <!-- Y Axis -->
    <g class="axis">
      {{range .YTicks}}<line x1="0" x2="-6" y1="{{.Y}}" y2="{{.Y}}"></line><text x="-10" y="{{add .Y 4}}" text-anchor="end">{{.Label}}</text>{{end}}
      <path d="M0,0V{{$.InnerHeight}}"></path>
      {{if .Metadata.YLabel}}<text class="axis-label" transform="rotate(-90)" x="{{neg (div .InnerHeight 2)}}" y="-50">{{.Metadata.YLabel}}</text>{{end}}
    </g>

    <!-- Data Lines (Multi-series) -->
    {{range .SeriesPaths}}
    <path fill="none" stroke="{{.Color}}" stroke-width="2px" d="{{.Path}}"></path>
    {{end}}
  </g>

  <!-- Legend -->
  <g class="legend" transform="translate({{add (add .Config.MarginLeft .InnerWidth) -100}}, {{.Config.MarginTop}})">
    {{range .LegendItems}}
    <rect x="0" y="{{.Y}}" width="12" height="12" fill="{{.Color}}"></rect>
    <text x="20" y="{{add .Y 10}}">{{.Name}}</text>
    {{end}}
  </g>
</svg>`

// SVGPlotter implements the Plotter interface to generate SVG charts.
type SVGPlotter struct {
	config   PlotConfig
	template *template.Template
}

func NewSVGPlotter(config PlotConfig) *SVGPlotter {
	tmpl := template.Must(template.New("svg").Funcs(template.FuncMap{
		"div": func(a, b int) int { return a / b },
		"add": func(a, b int) int { return a + b },
		"neg": func(a int) int { return -a },
	}).Parse(svgTemplate))
	return &SVGPlotter{config: config, template: tmpl}
}

// DefaultPlotConfig returns sensible defaults.
func DefaultPlotConfig() PlotConfig {
	return PlotConfig{
		Width: 800, Height: 400, MarginTop: 40, MarginRight: 120,
		MarginBottom: 50, MarginLeft: 60, GridColor: "#e5e7eb", TextColor: "#000000",
		YAxisMode: YAxisAuto,
		Colors:    []string{"#3b82f6", "#ef4444", "#10b981", "#f97316", "#8b5cf6", "#ec4899"},
	}
}

// Generate creates an SVG string from a multi-series dataset.
func (p *SVGPlotter) Generate(series []DataSeries, title, xLabel, yLabel string) (string, error) {
	data := MultiSeriesPlotData{
		Series:   series,
		Metadata: PlotMetadata{Title: title, XLabel: xLabel, YLabel: yLabel},
	}
	if len(data.Series) == 0 {
		return p.generateEmptyChart(data.Metadata)
	}

	innerWidth := p.config.Width - p.config.MarginLeft - p.config.MarginRight
	innerHeight := p.config.Height - p.config.MarginTop - p.config.MarginBottom

	xExtent, yExtent := p.findMultiSeriesExtents(data.Series)
	finalYExtent := p.adjustValueExtent(yExtent, p.config.YAxisMode)
	xScale := p.createTimeScale(xExtent, innerWidth)
	yScale := p.createLinearScale(finalYExtent, innerHeight)

	seriesPaths := make([]SeriesPath, len(data.Series))
	legendItems := make([]LegendItem, len(data.Series))
	for i, s := range data.Series {
		color := p.config.Colors[i%len(p.config.Colors)]
		seriesPaths[i] = SeriesPath{
			Path:  p.generateLinePath(s.Points, xScale, yScale),
			Color: color,
		}
		legendItems[i] = LegendItem{Name: s.Name, Color: color, Y: i * 20}
	}

	templateData := TemplateData{
		Config: p.config, Metadata: data.Metadata, InnerWidth: innerWidth, InnerHeight: innerHeight,
		XTicks: p.generateXTicks(xScale, innerHeight), YTicks: p.generateYTicks(yScale),
		GridLines: p.generateGridLines(yScale, innerWidth, innerHeight),
		SeriesPaths: seriesPaths, LegendItems: legendItems,
	}

	var result strings.Builder
	err := p.template.Execute(&result, templateData)
	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" + result.String(), err
}

func (p *SVGPlotter) generateEmptyChart(metadata PlotMetadata) (string, error) {
	templateData := TemplateData{
		Config: p.config, Metadata: metadata,
		InnerWidth: p.config.Width - p.config.MarginLeft - p.config.MarginRight,
		InnerHeight: p.config.Height - p.config.MarginTop - p.config.MarginBottom,
	}
	var result strings.Builder
	err := p.template.Execute(&result, templateData)
	return result.String(), err
}

// --- Helper methods for SVG generation ---

type timeScale struct{ domain [2]time.Time; rangeX [2]int }
type linearScale struct{ domain [2]float64; rangeY [2]int }

func (p *SVGPlotter) findMultiSeriesExtents(series []DataSeries) (x [2]time.Time, y [2]float64) {
	y = [2]float64{math.Inf(1), math.Inf(-1)}
	xSet := false
	for _, s := range series {
		if len(s.Points) == 0 { continue }
		if !xSet {
			x[0] = time.Unix(0, s.Points[0].X*int64(time.Millisecond)); x[1] = x[0]; xSet = true
		}
		for _, pt := range s.Points {
			t := time.Unix(0, pt.X*int64(time.Millisecond))
			if t.Before(x[0]) { x[0] = t }
			if t.After(x[1]) { x[1] = t }
			if pt.Y < y[0] { y[0] = pt.Y }
			if pt.Y > y[1] { y[1] = pt.Y }
		}
	}
	if y[0] > y[1] { return x, [2]float64{0, 100} }
	return x, y
}
func (p *SVGPlotter) createTimeScale(extent [2]time.Time, width int) timeScale {
	return timeScale{domain: extent, rangeX: [2]int{0, width}}
}
func (p *SVGPlotter) createLinearScale(extent [2]float64, height int) linearScale {
	return linearScale{domain: extent, rangeY: [2]int{height, 0}}
}
func (ts timeScale) scale(t time.Time) int {
	d := ts.domain[1].Sub(ts.domain[0]); if d == 0 { return ts.rangeX[0] }
	r := float64(t.Sub(ts.domain[0])) / float64(d)
	return ts.rangeX[0] + int(r*float64(ts.rangeX[1]-ts.rangeX[0]))
}
func (ls linearScale) scale(v float64) int {
	d := ls.domain[1] - ls.domain[0]; if d == 0 { return ls.rangeY[0] }
	r := (v - ls.domain[0]) / d
	return ls.rangeY[0] + int(r*float64(ls.rangeY[1]-ls.rangeY[0]))
}
func (p *SVGPlotter) generateLinePath(data []DataPoint, xs timeScale, ys linearScale) string {
	if len(data) < 2 { return "" }
	var b strings.Builder; b.WriteString("M")
	for i, pt := range data {
		x, y := xs.scale(time.Unix(0, pt.X*int64(time.Millisecond))), ys.scale(pt.Y)
		if i == 0 { b.WriteString(fmt.Sprintf("%d,%d", x, y)) } else { b.WriteString(fmt.Sprintf(" L%d,%d", x, y)) }
	}
	return b.String()
}
func (p *SVGPlotter) generateXTicks(xs timeScale, h int) []XTick {
	var ticks []XTick; numTicks := 8
	dur := xs.domain[1].Sub(xs.domain[0]); if dur <= 0 { return ticks }
	interval := dur / time.Duration(numTicks-1)
	for i := 0; i < numTicks; i++ {
		t := xs.domain[0].Add(time.Duration(i) * interval)
		ticks = append(ticks, XTick{X: xs.scale(t), Label: t.Format("15:04:05")})
	}
	return ticks
}
func (p *SVGPlotter) generateYTicks(ys linearScale) []YTick {
	var ticks []YTick; numTicks := 6
	valTicks := p.generateValueTicks(ys.domain[0], ys.domain[1], numTicks)
	prec := p.calculateOptimalPrecision(valTicks)
	for _, tick := range valTicks {
		ticks = append(ticks, YTick{Y: ys.scale(tick), Label: p.formatValue(tick, prec)})
	}
	return ticks
}
func (p *SVGPlotter) generateGridLines(ys linearScale, w, h int) []GridLine {
	var lines []GridLine
	for _, tick := range p.generateValueTicks(ys.domain[0], ys.domain[1], 6) {
		y := ys.scale(tick); lines = append(lines, GridLine{0, y, w, y})
	}
	return lines
}

// --- Value formatting and scaling helpers ---
type YAxisMode int
const ( YAxisAuto YAxisMode = iota; YAxisZeroBased; YAxisTight; YAxisNice )
func (p *SVGPlotter) adjustValueExtent(extent [2]float64, mode YAxisMode) [2]float64 {
	min, max := extent[0], extent[1]
	if min == max { if min == 0 { return [2]float64{-1, 1} }; padding := math.Abs(min) * 0.1; return [2]float64{min - padding, max + padding} }
	if mode == YAxisZeroBased { if min > 0 { min = 0 }; if max < 0 { max = 0 } }
	padding := (max - min) * 0.05
	return [2]float64{min - padding, max + padding}
}
func (p *SVGPlotter) generateValueTicks(min, max float64, maxTicks int) []float64 {
	if min >= max { return []float64{min} }; range_ := max - min
	rawStep := range_ / float64(maxTicks-1); magnitude := math.Pow(10, math.Floor(math.Log10(rawStep)))
	var step float64
	if normalizedStep := rawStep / magnitude; normalizedStep <= 1 { step = magnitude } else if normalizedStep <= 2 { step = 2 * magnitude } else if normalizedStep <= 5 { step = 5 * magnitude } else { step = 10 * magnitude }
	start := math.Floor(min/step) * step; var ticks []float64
	for tick := start; tick <= max+step/2; tick += step { if tick >= min-step/2 { ticks = append(ticks, tick) } }
	return ticks
}
func (p *SVGPlotter) calculateOptimalPrecision(values []float64) int {
	if len(values) <= 1 { return 1 }; minDiff := math.Inf(1)
	for i := 1; i < len(values); i++ { diff := math.Abs(values[i] - values[i-1]); if diff > 0 && diff < minDiff { minDiff = diff } }
	if minDiff > 0 && !math.IsInf(minDiff, 0) { precision := int(math.Max(0, -math.Floor(math.Log10(minDiff)))) + 1; if precision > 8 { return 8 }; return precision }
	return 2
}
func (p *SVGPlotter) formatValue(value float64, precision int) string {
	formatted := fmt.Sprintf("%."+fmt.Sprintf("%d", precision)+"f", value)
	if strings.Contains(formatted, ".") { formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".") }
	if formatted == "" || formatted == "-" { return "0" }
	return formatted
}
