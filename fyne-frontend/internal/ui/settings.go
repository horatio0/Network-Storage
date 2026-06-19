package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createSettingsView(a fyne.App) fyne.CanvasObject {
	p := a.Preferences()
	ip, port := createSettingsEntries(p)
	btn := widget.NewButton("Save", func() { saveSettings(p, ip.Text, port.Text) })
	btn.Importance = widget.HighImportance

	return container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Tailscale IP:"), ip,
		widget.NewLabel("Port:"), port, btn,
	))
}

func createSettingsEntries(p fyne.Preferences) (*widget.Entry, *widget.Entry) {
	ip := widget.NewEntry()
	ip.SetPlaceHolder("100.x.x.x")
	ip.SetText(p.StringWithFallback("server_ip", ""))
	port := widget.NewEntry()
	port.SetPlaceHolder("8080")
	port.SetText(p.StringWithFallback("server_port", "8080"))
	return ip, port
}

func saveSettings(p fyne.Preferences, ip, port string) {
	p.SetString("server_ip", ip)
	p.SetString("server_port", port)
}
