package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type lineChart struct {
	widget.BaseWidget
	data []float64
	max  float64
}

func newLineChart(max float64) *lineChart {
	c := &lineChart{max: max, data: make([]float64, 0)}
	c.ExtendBaseWidget(c)
	return c
}

func (c *lineChart) appendData(val float64) {
	c.data = append(c.data, val)
	if len(c.data) > 50 {
		c.data = c.data[1:]
	}
	c.Refresh()
}

func (c *lineChart) CreateRenderer() fyne.WidgetRenderer {
	return &chartRenderer{chart: c}
}

type chartRenderer struct {
	chart *lineChart
	lines []*canvas.Line
}

func (r *chartRenderer) Destroy() {}

func (r *chartRenderer) Layout(s fyne.Size) {
	if len(r.chart.data) < 2 {
		return
	}
	stepX := float32(s.Width) / float32(len(r.chart.data)-1)
	for i := 1; i < len(r.chart.data) && i-1 < len(r.lines); i++ {
		r.positionLine(i, stepX, s)
	}
}

func (r *chartRenderer) positionLine(i int, stepX float32, s fyne.Size) {
	h, max := float32(s.Height), float32(r.chart.max)
	x1 := float32(i-1) * stepX
	y1 := h - (float32(r.chart.data[i-1]) / max * h)
	x2 := float32(i) * stepX
	y2 := h - (float32(r.chart.data[i]) / max * h)
	r.lines[i-1].Position1 = fyne.NewPos(x1, y1)
	r.lines[i-1].Position2 = fyne.NewPos(x2, y2)
}

func (r *chartRenderer) MinSize() fyne.Size { return fyne.NewSize(100, 50) }

func (r *chartRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, len(r.lines))
	for i, l := range r.lines {
		objs[i] = l
	}
	return objs
}

func (r *chartRenderer) Refresh() {
	for len(r.lines) < len(r.chart.data)-1 {
		l := canvas.NewLine(color.NRGBA{R: 0, G: 255, B: 0, A: 255})
		l.StrokeWidth = 2
		r.lines = append(r.lines, l)
	}
	r.Layout(r.chart.Size())
}
