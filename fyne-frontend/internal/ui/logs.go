package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var logFilter = struct {
	All   bool
	Info  bool
	Warn  bool
	Error bool
}{All: true}

type logItemLayout struct{}

func (l *logItemLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[0].MinSize()
}

func (l *logItemLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	t := objects[0]
	btn := objects[1]

	btnMin := btn.MinSize()
	btnH := btnMin.Height
	if btnH > 14 {
		btnH = 14
	}
	btn.Resize(fyne.NewSize(btnMin.Width, btnH))
	btn.Move(fyne.NewPos(size.Width-btnMin.Width, (size.Height-btnH)/2))

	t.Resize(fyne.NewSize(size.Width-btnMin.Width-5, size.Height))
	t.Move(fyne.NewPos(0, 0))
}

type mouseTrackingButton struct {
	widget.Button
	onTappedWithEvent func(e *fyne.PointEvent)
}

func newMouseTrackingButton(text string, tapped func(e *fyne.PointEvent)) *mouseTrackingButton {
	b := &mouseTrackingButton{
		onTappedWithEvent: tapped,
	}
	b.Text = text
	b.ExtendBaseWidget(b)
	return b
}

func (b *mouseTrackingButton) Tapped(e *fyne.PointEvent) {
	if b.onTappedWithEvent != nil {
		b.onTappedWithEvent(e)
	} else {
		b.Button.Tapped(e)
	}
}

func createLogsView(a fyne.App, unused interface{}) fyne.CanvasObject {
	var filtered []LogEntry
	var list *widget.List

	updateFiltered := func() {
		logs := LoadLogs(a)
		filtered = nil
		for _, l := range logs {
			if logFilter.All || (logFilter.Info && l.Level == "info") || (logFilter.Warn && l.Level == "warn") || (logFilter.Error && l.Level == "error") {
				filtered = append(filtered, l)
			}
		}
	}
	
	updateFiltered()

	list = widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject {
			t := canvas.NewText("", color.White)
			t.TextSize = 12
			t.TextStyle = fyne.TextStyle{Monospace: true}
			
			btn := widget.NewButton("v", nil)
			
			return container.New(&logItemLayout{}, t, btn)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			c := o.(*fyne.Container)
			t := c.Objects[0].(*canvas.Text)
			btn := c.Objects[1].(*widget.Button)
			
			l := filtered[i]
			prefix := ""
			switch l.Level {
			case "info":
				prefix = "[Info] "
			case "warn":
				prefix = "[Warn] "
			case "error":
				prefix = "[Error] "
			default:
				if len(l.Level) > 0 {
					prefix = "[" + strings.ToUpper(l.Level[:1]) + strings.ToLower(l.Level[1:]) + "] "
				} else {
					prefix = "[Log] "
				}
			}
			t.Text = prefix + l.Time + " - " + l.Message
			if l.Level == "error" {
				t.Color = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
				btn.Show()
				btn.OnTapped = func() { showErrorDetails(a, l) }
			} else if l.Level == "warn" {
				t.Color = color.NRGBA{R: 255, G: 200, B: 50, A: 255}
				btn.Hide()
			} else {
				t.Color = color.White
				btn.Hide()
			}
			t.Refresh()
		},
	)
	list.HideSeparators = true
	
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, list)

	title := widget.NewLabelWithStyle("System Logs", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	clearBtn := widget.NewButton("Clear", func() { 
		ClearLogs(a)
		updateFiltered()
		list.Refresh() 
	})
	copyBtn := widget.NewButton("Copy", func() {
		var sb strings.Builder
		for _, l := range filtered {
			sb.WriteString(l.Time + " - " + l.Message + "\n")
		}
		a.Clipboard().SetContent(sb.String())
	})
	
	// Filtering buttons
	allCheck := widget.NewCheck("ALL", nil)
	infoCheck := widget.NewCheck("Info", nil)
	warnCheck := widget.NewCheck("Warn", nil)
	errCheck := widget.NewCheck("Error", nil)

	updating := false
	
	var filterBtn *mouseTrackingButton
	
	updateFilterText := func() {
		if allCheck.Checked {
			filterBtn.SetText("Filter: ALL \u25be")
		} else {
			var sel []string
			if infoCheck.Checked { sel = append(sel, "Info") }
			if warnCheck.Checked { sel = append(sel, "Warn") }
			if errCheck.Checked { sel = append(sel, "Error") }
			if len(sel) == 0 {
				filterBtn.SetText("Filter: None \u25be")
			} else {
				filterBtn.SetText("Filter: " + strings.Join(sel, ", ") + " \u25be")
			}
		}
	}

	filterBtn = newMouseTrackingButton("Filter: ALL \u25be", func(e *fyne.PointEvent) {
		content := container.NewVBox(allCheck, infoCheck, warnCheck, errCheck)
		popup := widget.NewPopUp(content, a.Driver().AllWindows()[0].Canvas())
		// Position it just below the mouse click point
		popup.ShowAtPosition(fyne.NewPos(e.AbsolutePosition.X, e.AbsolutePosition.Y+10))
	})

	allCheck.OnChanged = func(checked bool) {
		if updating {
			return
		}
		updating = true
		if checked {
			infoCheck.SetChecked(false)
			warnCheck.SetChecked(false)
			errCheck.SetChecked(false)
		}
		logFilter.All = checked
		updating = false
		updateFilterText()
		updateFiltered()
		list.Refresh()
	}

	updateOthers := func() {
		if updating {
			return
		}
		updating = true
		if infoCheck.Checked || warnCheck.Checked || errCheck.Checked {
			allCheck.SetChecked(false)
			logFilter.All = false
		} else {
			allCheck.SetChecked(true)
			logFilter.All = true
		}
		logFilter.Info = infoCheck.Checked
		logFilter.Warn = warnCheck.Checked
		logFilter.Error = errCheck.Checked
		updating = false
		updateFilterText()
		updateFiltered()
		list.Refresh()
	}

	infoCheck.OnChanged = func(b bool) { updateOthers() }
	warnCheck.OnChanged = func(b bool) { updateOthers() }
	errCheck.OnChanged = func(b bool) { updateOthers() }

	updating = true
	allCheck.SetChecked(logFilter.All)
	infoCheck.SetChecked(logFilter.Info)
	warnCheck.SetChecked(logFilter.Warn)
	errCheck.SetChecked(logFilter.Error)
	updating = false
	updateFilterText()
	
	filtersBox := container.NewHBox(filterBtn)
	buttonsBox := container.NewHBox(copyBtn, clearBtn)
	
	topBar := container.NewBorder(nil, nil, filtersBox, buttonsBox, title)

	return container.NewBorder(topBar, nil, nil, nil, stack)
}

func showErrorDetails(a fyne.App, l LogEntry) {
	w := a.Driver().AllWindows()[0]
	
	timeLbl := widget.NewLabel("Time: " + l.Time)
	
	cmdLbl := widget.NewLabel(l.Command)
	cmdLbl.Wrapping = fyne.TextWrapWord
	
	stderrLbl := widget.NewLabel(l.Stderr)
	stderrLbl.Wrapping = fyne.TextWrapWord
	
	exitLbl := widget.NewLabel(fmt.Sprintf("Exit Code: %d", l.ExitCode))
	
	var d dialog.Dialog
	
	copyBtn := widget.NewButton("Copy", func() {
		text := fmt.Sprintf("Time: %s\nCommand: %s\nStderr: %s\nExit Code: %d", l.Time, l.Command, l.Stderr, l.ExitCode)
		a.Clipboard().SetContent(text)
	})
	closeBtn := widget.NewButton("Close", func() {
		if d != nil {
			d.Hide()
		}
	})
	
	btnGrid := container.NewGridWithColumns(2, copyBtn, closeBtn)

	top := container.NewVBox(
		timeLbl,
		widget.NewLabel("Command:"),
		cmdLbl,
		widget.NewLabel("Stderr:"),
	)
	bottom := container.NewVBox(
		exitLbl,
		btnGrid,
	)

	scroll := container.NewVScroll(stderrLbl)

	content := container.NewBorder(top, bottom, nil, nil, scroll)
	d = dialog.NewCustomWithoutButtons("Error Details", content, w)

	targetSize := fyne.NewSize(
		fyne.Max(top.MinSize().Width, fyne.Max(bottom.MinSize().Width, stderrLbl.MinSize().Width)),
		top.MinSize().Height+bottom.MinSize().Height+stderrLbl.MinSize().Height+40,
	)
	
	winSize := w.Canvas().Size()
	maxW := winSize.Width * 0.8
	maxH := winSize.Height * 0.8
	
	if targetSize.Width > maxW {
		targetSize.Width = maxW
	}
	if targetSize.Height > maxH {
		targetSize.Height = maxH
	}
	if targetSize.Width < 300 {
		targetSize.Width = 300
	}
	if targetSize.Height < 200 {
		targetSize.Height = 200
	}

	d.Resize(targetSize)
	d.Show()
}
