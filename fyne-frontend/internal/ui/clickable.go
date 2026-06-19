package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tappable struct {
	widget.BaseWidget
	onTapped func()
}

func newTappable(f func()) *tappable {
	t := &tappable{onTapped: f}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappable) Tapped(_ *fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *tappable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(widget.NewLabel(""))
}
