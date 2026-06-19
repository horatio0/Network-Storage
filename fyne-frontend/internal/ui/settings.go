package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createSettingsView(a fyne.App) fyne.CanvasObject {
	ip, port, share, local := createSettingsEntries(a)

	saveBtn := widget.NewButton("Save", func() {
		a.Preferences().SetString("server_ip", ip.Text)
		a.Preferences().SetString("server_port", port.Text)
		a.Preferences().SetString("share_name", share.Text)
		a.Preferences().SetString("mount_path", local.Text)
	})

	return buildSettingsForm(ip, port, share, local, saveBtn)
}

func createSettingsEntries(a fyne.App) (*widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry) {
	ip := widget.NewEntry()
	ip.SetText(a.Preferences().StringWithFallback("server_ip", ""))
	port := widget.NewEntry()
	port.SetText(a.Preferences().StringWithFallback("server_port", "8080"))

	share := widget.NewEntry()
	share.SetText(a.Preferences().StringWithFallback("share_name", "shared"))

	local := widget.NewEntry()
	local.SetText(a.Preferences().StringWithFallback("mount_path", getDefaultMountPath()))

	return ip, port, share, local
}

func buildSettingsForm(ip, port, share, local *widget.Entry, saveBtn *widget.Button) fyne.CanvasObject {
	return container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Server Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Tailscale IP:"), ip,
		widget.NewLabel("Port:"), port,
		widget.NewLabel("Share Name:"), share,
		widget.NewLabel("Local Mount Path:"), local,
		saveBtn,
	))
}

func getDefaultMountPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "HomeNAS")
}
