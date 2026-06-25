package main
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
)
type menuButton struct {
	widget.Button
	menu *fyne.Menu
}
func (b *menuButton) Tapped(e *fyne.PointEvent) {
	// widget.Button's tap will be overridden
	widget.ShowPopUpMenuAtPosition(b.menu, fyne.CurrentApp().Driver().AllWindows()[0].Canvas(), e.AbsolutePosition)
}
func main() {}
