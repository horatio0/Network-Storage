package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createSettingsView(a fyne.App) fyne.CanvasObject {
	ip, port, share, local, pwd := createSettingsEntries(a)

	saveBtn := widget.NewButton("Save", func() {
		a.Preferences().SetString("server_ip", ip.Text)
		a.Preferences().SetString("server_port", port.Text)
		a.Preferences().SetString("share_name", share.Text)
		a.Preferences().SetString("mount_path", local.Text)
		a.Preferences().SetString("app_password", pwd.Text)
	})

	return buildSettingsForm(ip, port, share, local, pwd, saveBtn)
}

func newPrefEntry(a fyne.App, key, def string, isPwd bool) *widget.Entry {
	e := widget.NewEntry()
	if isPwd {
		e = widget.NewPasswordEntry()
	}
	e.SetText(a.Preferences().StringWithFallback(key, def))
	return e
}

func createSettingsEntries(a fyne.App) (*widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry) {
	ip := newPrefEntry(a, "server_ip", "", false)
	port := newPrefEntry(a, "server_port", "8080", false)
	share := newPrefEntry(a, "share_name", "/NS/share", false)
	local := newPrefEntry(a, "mount_path", getDefaultMountPath(), false)
	pwd := newPrefEntry(a, "app_password", "0000", true)
	return ip, port, share, local, pwd
}

func buildSettingsForm(ip, port, share, local, pwd *widget.Entry, saveBtn *widget.Button) fyne.CanvasObject {
	return container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Server Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Tailscale IP:"), ip,
		widget.NewLabel("Port:"), port,
		widget.NewLabel("Share Name:"), share,
		widget.NewLabel("Local Mount Path:"), local,
		widget.NewLabel("App Lock Password:"), pwd,
		saveBtn,
	))
}

func getDefaultMountPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "HomeNAS")
}
