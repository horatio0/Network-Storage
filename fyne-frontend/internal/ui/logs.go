package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createLogsView(a fyne.App, unused interface{}) fyne.CanvasObject {
	logs := LoadLogs(a)
	vbox := container.NewVBox()
	for _, l := range logs {
		vbox.Add(buildLogText(l))
	}
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, container.NewPadded(container.NewScroll(vbox)))

	title := widget.NewLabelWithStyle("System Logs", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	clearBtn := widget.NewButton("Clear", func() { ClearLogs(a) })
	copyBtn := widget.NewButton("Copy", func() {
		logs := LoadLogs(a)
		var sb strings.Builder
		for _, l := range logs {
			sb.WriteString(l.Time + " - " + l.Message + "\n")
		}
		if len(a.Driver().AllWindows()) > 0 {
			a.Driver().AllWindows()[0].Clipboard().SetContent(sb.String())
		}
	})
	buttons := container.NewHBox(copyBtn, clearBtn)
	topBar := container.NewBorder(nil, nil, nil, buttons, title)

	return container.NewBorder(topBar, nil, nil, nil, stack)
}

func buildLogText(l LogEntry) *canvas.Text {
	t := canvas.NewText(l.Time+" - "+l.Message, color.White)
	t.TextSize = 12
	t.TextStyle = fyne.TextStyle{Monospace: true}
	return t
}
