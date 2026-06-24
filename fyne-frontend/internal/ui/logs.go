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
	
	list := widget.NewList(
		func() int { return len(logs) },
		func() fyne.CanvasObject {
			t := canvas.NewText("", color.White)
			t.TextSize = 12
			t.TextStyle = fyne.TextStyle{Monospace: true}
			return t
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			t := o.(*canvas.Text)
			t.Text = logs[i].Time + " - " + logs[i].Message
			t.Refresh()
		},
	)
	list.HideSeparators = true
	
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, list)

	title := widget.NewLabelWithStyle("System Logs", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	clearBtn := widget.NewButton("Clear", func() { 
		ClearLogs(a)
		logs = LoadLogs(a)
		list.Refresh() 
	})
	copyBtn := widget.NewButton("Copy", func() {
		logs := LoadLogs(a)
		var sb strings.Builder
		for _, l := range logs {
			sb.WriteString(l.Time + " - " + l.Message + "\n")
		}
		a.Clipboard().SetContent(sb.String())
	})
	buttons := container.NewHBox(copyBtn, clearBtn)
	topBar := container.NewBorder(nil, nil, nil, buttons, title)

	return container.NewBorder(topBar, nil, nil, nil, stack)
}
