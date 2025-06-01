package commands

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

// DataPoint represents a single plot point
type DataPoint struct {
	X int64   `json:"x"` // Unix timestamp in milliseconds
	Y float64 `json:"y"` // Value (e.g., latency in ms)
}

// PlotMetadata contains chart labels and title
type PlotMetadata struct {
	XLabel string `json:"xLabel,omitempty"`
	YLabel string `json:"yLabel,omitempty"`
	Title  string `json:"title,omitempty"`
}

// PlotData represents the complete dataset for plotting
type PlotData struct {
	Data     []DataPoint  `json:"data"`
	Metadata PlotMetadata `json:"metadata"`
}

// PlotConfig holds styling and dimension configuration
type PlotConfig struct {
	Width        int
	Height       int
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int
	LineColor    string
	DotColor     string
	GridColor    string
	TextColor    string
	YAxisMode    YAxisMode // New field for Y-axis scaling
}

// TemplateData contains all data needed for SVG template
type TemplateData struct {
	Config      PlotConfig
	Metadata    PlotMetadata
	InnerWidth  int
	InnerHeight int
	LinePath    string
	Dots        []Dot
	XTicks      []XTick
	YTicks      []YTick
	GridLines   []GridLine
}

type Dot struct {
	X int
	Y int
}

type XTick struct {
	X     int
	Label string
}

type YTick struct {
	Y     int
	Label string
}

type GridLine struct {
	X1, Y1, X2, Y2 int
}

// SVG template
const svgTemplate = `<svg width="{{.Config.Width}}" height="{{.Config.Height}}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <style>
      .axis { font: 12px sans-serif; fill: {{.Config.TextColor}}; }
      .axis path, .axis line { fill: none; stroke: {{.Config.TextColor}}; shape-rendering: crispEdges; }
      .grid-line { stroke: {{.Config.GridColor}}; stroke-width: 0.5px; }
      .line { fill: none; stroke: {{.Config.LineColor}}; stroke-width: 2px; }
      .dot { fill: {{.Config.DotColor}}; }
      .title { font: bold 16px sans-serif; text-anchor: middle; fill: {{.Config.TextColor}}; }
      .axis-label { font: 12px sans-serif; text-anchor: middle; fill: {{.Config.TextColor}}; }
    </style>
  </defs>

  {{if .Metadata.Title}}
  <text class="title" x="{{div .Config.Width 2}}" y="20">{{.Metadata.Title}}</text>
  {{end}}

  <g transform="translate({{.Config.MarginLeft}},{{.Config.MarginTop}})">
    <!-- Grid Lines -->
    {{range .GridLines}}
    <line class="grid-line" x1="{{.X1}}" x2="{{.X2}}" y1="{{.Y1}}" y2="{{.Y2}}"></line>
    {{end}}

    <!-- X Axis -->
    <g class="axis" transform="translate(0,{{.InnerHeight}})">
      {{range .XTicks}}
      <line x1="{{.X}}" x2="{{.X}}" y1="0" y2="6"></line>
      <text x="{{.X}}" y="20" text-anchor="middle">{{.Label}}</text>
      {{end}}
      <path d="M0,0H{{$.InnerWidth}}"></path>
      {{if .Metadata.XLabel}}
      <text class="axis-label" x="{{div .InnerWidth 2}}" y="35">{{.Metadata.XLabel}}</text>
      {{end}}
    </g>

    <!-- Y Axis -->
    <g class="axis">
      {{range .YTicks}}
      <line x1="0" x2="-6" y1="{{.Y}}" y2="{{.Y}}"></line>
      <text x="-10" y="{{add .Y 4}}" text-anchor="end">{{.Label}}</text>
      {{end}}
      <path d="M0,0V{{$.InnerHeight}}"></path>
      {{if .Metadata.YLabel}}
      <text class="axis-label" transform="rotate(-90)" x="{{neg (div .InnerHeight 2)}}" y="-40">{{.Metadata.YLabel}}</text>
      {{end}}
    </g>

    <!-- Data Line -->
    {{if .LinePath}}
    <path class="line" d="{{.LinePath}}"></path>
    {{end}}

    <!-- Data Points -->
    {{range .Dots}}
    <circle class="dot" cx="{{.X}}" cy="{{.Y}}" r="3"></circle>
    {{end}}
  </g>
</svg>`

// DefaultPlotConfig returns sensible defaults
func DefaultPlotConfig() PlotConfig {
	return PlotConfig{
		Width:        800,
		Height:       400,
		MarginTop:    20,
		MarginRight:  30,
		MarginBottom: 40,
		MarginLeft:   60,
		LineColor:    "#3b82f6",
		DotColor:     "#3b82f6",
		GridColor:    "#e5e7eb",
		TextColor:    "#000000",
		YAxisMode:    YAxisAuto,
	}
}

// SVGGenerator generates SVG plots using templates
type SVGGenerator struct {
	config   PlotConfig
	template *template.Template
}

// NewSVGGenerator creates a new SVG generator with the given config
func NewSVGGenerator(config PlotConfig) *SVGGenerator {
	// Create template with helper functions
	tmpl := template.Must(template.New("svg").Funcs(template.FuncMap{
		"div": func(a, b int) int { return a / b },
		"add": func(a, b int) int { return a + b },
		"neg": func(a int) int { return -a },
	}).Parse(svgTemplate))

	return &SVGGenerator{
		config:   config,
		template: tmpl,
	}
}

// GenerateSVG creates an SVG string from plot data
func (sg *SVGGenerator) GenerateSVG(data PlotData) (string, error) {
	if len(data.Data) == 0 {
		return sg.generateEmptyChart(data.Metadata)
	}

	// Calculate inner dimensions
	innerWidth := sg.config.Width - sg.config.MarginLeft - sg.config.MarginRight
	innerHeight := sg.config.Height - sg.config.MarginTop - sg.config.MarginBottom

	// Find data extents and create scales
	xExtent := sg.findTimeExtent(data.Data)
	yExtent := sg.findValueExtentWithMode(data.Data, sg.config.YAxisMode)
	xScale := sg.createTimeScale(xExtent, innerWidth)
	yScale := sg.createLinearScale(yExtent, innerHeight)

	// Prepare template data
	templateData := TemplateData{
		Config:      sg.config,
		Metadata:    data.Metadata,
		InnerWidth:  innerWidth,
		InnerHeight: innerHeight,
		LinePath:    sg.generateLinePath(data.Data, xScale, yScale),
		Dots:        sg.generateDots(data.Data, xScale, yScale),
		XTicks:      sg.generateXTicks(xScale, innerHeight),
		YTicks:      sg.generateYTicks(yScale),
		GridLines:   sg.generateGridLines(xScale, yScale, innerWidth, innerHeight),
	}

	// Execute template
	var result strings.Builder
	err := sg.template.Execute(&result, templateData)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	preamble := `<?xml version="1.0" encoding="UTF-8"?>`
	return preamble + result.String(), nil
}

func (sg *SVGGenerator) generateEmptyChart(metadata PlotMetadata) (string, error) {
	templateData := TemplateData{
		Config:      sg.config,
		Metadata:    metadata,
		InnerWidth:  sg.config.Width - sg.config.MarginLeft - sg.config.MarginRight,
		InnerHeight: sg.config.Height - sg.config.MarginTop - sg.config.MarginBottom,
	}

	var result strings.Builder
	err := sg.template.Execute(&result, templateData)
	return result.String(), err
}

// Scale types and methods (same as before)
type TimeScale struct {
	domain [2]time.Time
	rangeX [2]int
}

type LinearScale struct {
	domain [2]float64
	rangeY [2]int
}

func (sg *SVGGenerator) findTimeExtent(data []DataPoint) [2]time.Time {
	if len(data) == 0 {
		now := time.Now()
		return [2]time.Time{now.Add(-time.Hour), now}
	}

	min := time.Unix(0, data[0].X*int64(time.Millisecond))
	max := min

	for _, point := range data {
		t := time.Unix(0, point.X*int64(time.Millisecond))
		if t.Before(min) {
			min = t
		}
		if t.After(max) {
			max = t
		}
	}

	return [2]time.Time{min, max}
}

// YAxisMode defines different Y-axis scaling strategies
type YAxisMode int

const (
	YAxisAuto      YAxisMode = iota // Auto-scale based on data with padding
	YAxisZeroBased                  // Always start from 0
	YAxisTight                      // Tight bounds with minimal padding
	YAxisNice                       // Nice round numbers
)

func (sg *SVGGenerator) findValueExtent(data []DataPoint) [2]float64 {
	return sg.findValueExtentWithMode(data, YAxisAuto)
}

func (sg *SVGGenerator) findValueExtentWithMode(data []DataPoint, mode YAxisMode) [2]float64 {
	if len(data) == 0 {
		return [2]float64{0, 100}
	}

	// Find actual data min/max
	min := data[0].Y
	max := min

	for _, point := range data {
		if point.Y < min {
			min = point.Y
		}
		if point.Y > max {
			max = point.Y
		}
	}

	// Handle case where all values are the same
	if min == max {
		if min == 0 {
			return [2]float64{-1, 1}
		}
		padding := math.Abs(min) * 0.1
		return [2]float64{min - padding, max + padding}
	}

	switch mode {
	case YAxisZeroBased:
		return sg.zeroBasedExtent(min, max)
	case YAxisTight:
		return sg.tightExtent(min, max)
	case YAxisNice:
		return sg.niceExtent(min, max)
	default: // YAxisAuto
		return sg.autoExtent(min, max)
	}
}

// Auto mode: intelligent choice based on data characteristics
func (sg *SVGGenerator) autoExtent(min, max float64) [2]float64 {
	range_ := max - min

	// If data is close to zero, use zero-based
	if min >= 0 && min <= range_*0.2 {
		return sg.zeroBasedExtent(min, max)
	}

	// If range is small relative to values, use tight
	if range_ < math.Abs(min)*0.3 {
		return sg.tightExtent(min, max)
	}

	// Otherwise use nice round numbers
	return sg.niceExtent(min, max)
}

// Zero-based: always include 0
func (sg *SVGGenerator) zeroBasedExtent(min, max float64) [2]float64 {
	if min > 0 {
		min = 0
	}
	if max < 0 {
		max = 0
	}

	// Add small padding
	range_ := max - min
	padding := range_ * 0.05

	return [2]float64{min - padding, max + padding}
}

// Tight: minimal padding, focus on data range
func (sg *SVGGenerator) tightExtent(min, max float64) [2]float64 {
	range_ := max - min
	padding := range_ * 0.02 // Very small padding

	return [2]float64{min - padding, max + padding}
}

// Nice: round to pleasant numbers
func (sg *SVGGenerator) niceExtent(min, max float64) [2]float64 {
	range_ := max - min

	// Calculate nice step size
	magnitude := math.Pow(10, math.Floor(math.Log10(range_)))
	normalizedRange := range_ / magnitude

	var niceRange float64
	if normalizedRange <= 1 {
		niceRange = magnitude
	} else if normalizedRange <= 2 {
		niceRange = 2 * magnitude
	} else if normalizedRange <= 5 {
		niceRange = 5 * magnitude
	} else {
		niceRange = 10 * magnitude
	}

	// Expand range to nice bounds
	center := (min + max) / 2
	niceMin := center - niceRange/2
	niceMax := center + niceRange/2

	// Adjust to nice round numbers
	stepSize := niceRange / 10
	niceMin = math.Floor(niceMin/stepSize) * stepSize
	niceMax = math.Ceil(niceMax/stepSize) * stepSize

	return [2]float64{niceMin, niceMax}
}

func (sg *SVGGenerator) createTimeScale(extent [2]time.Time, width int) TimeScale {
	return TimeScale{
		domain: extent,
		rangeX: [2]int{0, width},
	}
}

func (sg *SVGGenerator) createLinearScale(extent [2]float64, height int) LinearScale {
	return LinearScale{
		domain: extent,
		rangeY: [2]int{height, 0}, // Inverted for SVG coordinates
	}
}

func (ts TimeScale) scale(t time.Time) int {
	duration := ts.domain[1].Sub(ts.domain[0])
	if duration == 0 {
		return ts.rangeX[0]
	}

	elapsed := t.Sub(ts.domain[0])
	ratio := float64(elapsed) / float64(duration)
	return ts.rangeX[0] + int(ratio*float64(ts.rangeX[1]-ts.rangeX[0]))
}

func (ls LinearScale) scale(value float64) int {
	if ls.domain[1] == ls.domain[0] {
		return ls.rangeY[0]
	}

	ratio := (value - ls.domain[0]) / (ls.domain[1] - ls.domain[0])
	return ls.rangeY[0] + int(ratio*float64(ls.rangeY[1]-ls.rangeY[0]))
}

// Template data generation methods
func (sg *SVGGenerator) generateLinePath(data []DataPoint, xScale TimeScale, yScale LinearScale) string {
	if len(data) < 2 {
		return ""
	}

	var path strings.Builder
	path.WriteString("M")

	for i, point := range data {
		t := time.Unix(0, point.X*int64(time.Millisecond))
		x := xScale.scale(t)
		y := yScale.scale(point.Y)

		if i == 0 {
			path.WriteString(fmt.Sprintf("%d,%d", x, y))
		} else {
			path.WriteString(fmt.Sprintf("L%d,%d", x, y))
		}
	}

	return path.String()
}

func (sg *SVGGenerator) generateDots(data []DataPoint, xScale TimeScale, yScale LinearScale) []Dot {
	var dots []Dot

	for _, point := range data {
		t := time.Unix(0, point.X*int64(time.Millisecond))
		x := xScale.scale(t)
		y := yScale.scale(point.Y)
		dots = append(dots, Dot{X: x, Y: y})
	}

	return dots
}

func (sg *SVGGenerator) generateXTicks(xScale TimeScale, height int) []XTick {
	var ticks []XTick

	duration := xScale.domain[1].Sub(xScale.domain[0])
	timeTicks := sg.generateTimeTicks(xScale.domain[0], duration, 8)

	for _, tick := range timeTicks {
		x := xScale.scale(tick)
		label := tick.Format("15:04")
		ticks = append(ticks, XTick{X: x, Label: label})
	}

	return ticks
}

func (sg *SVGGenerator) generateYTicks(yScale LinearScale) []YTick {
	var ticks []YTick

	valueTicks := sg.generateValueTicks(yScale.domain[0], yScale.domain[1], 6)
	precision := sg.calculateOptimalPrecision(valueTicks)

	for _, tick := range valueTicks {
		y := yScale.scale(tick)
		label := sg.formatValue(tick, precision)
		ticks = append(ticks, YTick{Y: y, Label: label})
	}

	return ticks
}

// calculateOptimalPrecision determines how many decimal places to show
func (sg *SVGGenerator) calculateOptimalPrecision(values []float64) int {
	if len(values) <= 1 {
		return 1
	}

	// Find the smallest non-zero difference between consecutive values
	minDiff := math.Inf(1)
	for i := 1; i < len(values); i++ {
		diff := math.Abs(values[i] - values[i-1])
		if diff > 0 && diff < minDiff {
			minDiff = diff
		}
	}

	// Also consider the magnitude of the values themselves
	maxVal := 0.0
	for _, val := range values {
		if math.Abs(val) > maxVal {
			maxVal = math.Abs(val)
		}
	}

	// Calculate precision needed for the smallest difference
	var precisionForDiff int
	if minDiff == math.Inf(1) || minDiff == 0 {
		precisionForDiff = 1
	} else {
		// How many decimal places needed to represent the difference?
		precisionForDiff = int(math.Max(0, -math.Floor(math.Log10(minDiff)))) + 1
	}

	// Calculate precision needed for the largest value
	var precisionForValue int
	if maxVal == 0 {
		precisionForValue = 1
	} else if maxVal >= 1 {
		// For values >= 1, we usually don't need many decimal places
		precisionForValue = 2
	} else {
		// For values < 1, we need enough precision to show significant digits
		precisionForValue = int(math.Max(0, -math.Floor(math.Log10(maxVal)))) + 2
	}

	// Use the maximum of both, but cap at reasonable limits
	precision := int(math.Max(float64(precisionForDiff), float64(precisionForValue)))

	// Apply reasonable bounds
	if precision < 0 {
		precision = 0
	} else if precision > 8 {
		precision = 8 // Cap at 8 decimal places for readability
	}

	return precision
}

// formatValue formats a number with the specified precision, removing trailing zeros
func (sg *SVGGenerator) formatValue(value float64, precision int) string {
	// Handle special cases
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return fmt.Sprintf("%.1f", value)
	}

	// Format with specified precision
	formatted := fmt.Sprintf("%."+fmt.Sprintf("%d", precision)+"f", value)

	// Remove trailing zeros and unnecessary decimal point
	if strings.Contains(formatted, ".") {
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
	}

	// Handle case where we end up with empty string (shouldn't happen, but be safe)
	if formatted == "" || formatted == "-" {
		formatted = "0"
	}

	return formatted
}

func (sg *SVGGenerator) generateGridLines(xScale TimeScale, yScale LinearScale, width, height int) []GridLine {
	var lines []GridLine

	// Vertical grid lines
	duration := xScale.domain[1].Sub(xScale.domain[0])
	timeTicks := sg.generateTimeTicks(xScale.domain[0], duration, 8)

	for _, tick := range timeTicks {
		x := xScale.scale(tick)
		lines = append(lines, GridLine{X1: x, Y1: 0, X2: x, Y2: height})
	}

	// Horizontal grid lines
	valueTicks := sg.generateValueTicks(yScale.domain[0], yScale.domain[1], 6)
	for _, tick := range valueTicks {
		y := yScale.scale(tick)
		lines = append(lines, GridLine{X1: 0, Y1: y, X2: width, Y2: y})
	}

	return lines
}

// Helper functions (same tick generation logic as before)
func (sg *SVGGenerator) generateTimeTicks(start time.Time, duration time.Duration, maxTicks int) []time.Time {
	var ticks []time.Time

	if duration <= 0 {
		return ticks
	}

	var interval time.Duration
	if duration <= time.Hour {
		interval = duration / time.Duration(maxTicks)
		interval = time.Duration(math.Round(float64(interval)/float64(time.Minute))) * time.Minute
		if interval < time.Minute {
			interval = time.Minute
		}
	} else if duration <= 24*time.Hour {
		interval = time.Hour
	} else {
		interval = 24 * time.Hour
	}

	for t := start; t.Before(start.Add(duration)) || t.Equal(start.Add(duration)); t = t.Add(interval) {
		ticks = append(ticks, t)
	}

	return ticks
}

func (sg *SVGGenerator) generateValueTicks(min, max float64, maxTicks int) []float64 {
	if min >= max {
		return []float64{min}
	}

	range_ := max - min
	rawStep := range_ / float64(maxTicks-1)

	magnitude := math.Pow(10, math.Floor(math.Log10(rawStep)))
	normalizedStep := rawStep / magnitude

	var step float64
	if normalizedStep <= 1 {
		step = magnitude
	} else if normalizedStep <= 2 {
		step = 2 * magnitude
	} else if normalizedStep <= 5 {
		step = 5 * magnitude
	} else {
		step = 10 * magnitude
	}

	start := math.Floor(min/step) * step
	var ticks []float64

	for tick := start; tick <= max+step/2; tick += step {
		if tick >= min-step/2 {
			ticks = append(ticks, tick)
		}
	}

	return ticks
}

// Example usage
func plot(outfile string, points []DataPoint, xlabel, ylabel, title string) string {
	data := PlotData{
		Data: points,
		Metadata: PlotMetadata{
			XLabel: xlabel,
			YLabel: ylabel,
			Title:  title,
		},
	}

	// Generate SVG
	config := DefaultPlotConfig()
	// config.YAxisMode = YAxisNice
	// config.YAxisMode = YAxisTight
	config.YAxisMode = YAxisZeroBased
	generator := NewSVGGenerator(config)
	svg, err := generator.GenerateSVG(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	err = os.WriteFile(outfile, []byte(svg), 0666)
	if err != nil {
		log.Printf("Error writing to file %s: %v", outfile, err)
	} else {
		log.Printf("Successfully written '%s' to %s", title, outfile)
	}
	return svg
}

/*
// Example usage with different precision requirements
func main() {
	now := time.Now()

	// Test case 1: Small decimal values (0.0001 - 0.0010)
	smallDecimalData := PlotData{
		Data: []DataPoint{
			{X: now.Add(-20 * time.Minute).UnixMilli(), Y: 0.0001},
			{X: now.Add(-18 * time.Minute).UnixMilli(), Y: 0.0003},
			{X: now.Add(-16 * time.Minute).UnixMilli(), Y: 0.0007},
			{X: now.Add(-14 * time.Minute).UnixMilli(), Y: 0.0009},
			{X: now.Add(-12 * time.Minute).UnixMilli(), Y: 0.0010},
		},
		Metadata: PlotMetadata{
			XLabel: "Time",
			YLabel: "Error Rate",
			Title:  "Small Decimal Values (Auto Precision)",
		},
	}

	// Test case 2: Large integers
	largeIntData := PlotData{
		Data: []DataPoint{
			{X: now.Add(-20 * time.Minute).UnixMilli(), Y: 1000},
			{X: now.Add(-18 * time.Minute).UnixMilli(), Y: 1050},
			{X: now.Add(-16 * time.Minute).UnixMilli(), Y: 1100},
			{X: now.Add(-14 * time.Minute).UnixMilli(), Y: 1150},
			{X: now.Add(-12 * time.Minute).UnixMilli(), Y: 1200},
		},
		Metadata: PlotMetadata{
			XLabel: "Time",
			YLabel: "Requests",
			Title:  "Large Integer Values",
		},
	}

	// Test case 3: Mixed precision values
	mixedData := PlotData{
		Data: []DataPoint{
			{X: now.Add(-20 * time.Minute).UnixMilli(), Y: 45.123},
			{X: now.Add(-18 * time.Minute).UnixMilli(), Y: 45.456},
			{X: now.Add(-16 * time.Minute).UnixMilli(), Y: 45.789},
			{X: now.Add(-14 * time.Minute).UnixMilli(), Y: 46.012},
			{X: now.Add(-12 * time.Minute).UnixMilli(), Y: 46.345},
		},
		Metadata: PlotMetadata{
			XLabel: "Time",
			YLabel: "Latency (ms)",
			Title:  "Mixed Precision Values",
		},
	}

	fmt.Println("=== SMALL DECIMAL VALUES (0.0001-0.0010) ===")
	generator := NewSVGGenerator(DefaultPlotConfig())
	svg, _ := generator.GenerateSVG(smallDecimalData)
	fmt.Println("Y-axis will show: 0.0001, 0.0002, 0.0004, 0.0006, 0.0008, 0.001")

	fmt.Println("\n=== LARGE INTEGER VALUES (1000-1200) ===")
	svg, _ = generator.GenerateSVG(largeIntData)
	fmt.Println("Y-axis will show: 1000, 1050, 1100, 1150, 1200")

	fmt.Println("\n=== MIXED PRECISION VALUES (45.123-46.345) ===")
	svg, _ = generator.GenerateSVG(mixedData)
	fmt.Println("Y-axis will show: 45.1, 45.2, 45.4, 45.6, 45.8, 46, 46.2, 46.4")

	fmt.Println("\n" + svg)
}
*/
