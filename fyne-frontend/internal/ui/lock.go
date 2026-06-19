package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createLockView(a fyne.App, w fyne.Window, main fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 15, G: 15, B: 15, A: 255})
	pwd := widget.NewPasswordEntry()
	pwd.SetPlaceHolder("Enter Password")
	pwdWrap := container.NewGridWrap(fyne.NewSize(300, 36), pwd)

	errLbl := canvas.NewText("", color.NRGBA{R: 255, G: 50, B: 50, A: 255})
	errLbl.Alignment = fyne.TextAlignCenter

	btn := widget.NewButton("Unlock", func() { handleUnlock(a, w, main, pwd, errLbl) })
	box := container.NewVBox(widget.NewLabelWithStyle("App Locked", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), pwdWrap, btn, errLbl)
	return container.NewStack(bg, container.NewCenter(box))
}

func handleUnlock(a fyne.App, w fyne.Window, main fyne.CanvasObject, pwd *widget.Entry, errLbl *canvas.Text) {
	saved := a.Preferences().StringWithFallback("app_password", "0000")
	if pwd.Text == saved {
		w.SetContent(main)
	} else {
		pwd.SetText("")
		errLbl.Text = "Incorrect Password!"
		errLbl.Refresh()
	}
}
