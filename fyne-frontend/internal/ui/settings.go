package ui

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"network-storage-client/internal/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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

	debugMenu := fyne.NewMenu("",
		&fyne.MenuItem{
			Label: "Test Error",
			ChildMenu: fyne.NewMenu("",
				fyne.NewMenuItem("404 Not Found", func() {
					go func() {
						AddInfoLog(a, "Running Test Error...")
						ipStr := a.Preferences().StringWithFallback("server_ip", "")
						portStr := a.Preferences().StringWithFallback("server_port", "8080")
						url := "http://" + ipStr + ":" + portStr + "/abcdef"
						req, _ := http.NewRequest("GET", url, nil)
						c := client.NewHTTPClient(a)
						_, err := c.DoRequest(req)
						if err != nil {
							AddErrorLog(a, "404 Not Found test failed: "+err.Error(), "GET "+url, err.Error(), 1)
						}
					}()
				}),
				fyne.NewMenuItem("Mount Error", func() {
					go func() {
						AddInfoLog(a, "Running Test Error...")
						ip := "255.255.255.255"
						share := "invalid_share"
						local := "/tmp/invalid_mount"
						err := client.MountDrive(ip, share, local)
						if err != nil {
							cmdStr := ""
							if runtime.GOOS == "windows" {
								cmdStr = fmt.Sprintf(`net use %s \\%s\%s`, local, ip, share)
							} else {
								cmdStr = fmt.Sprintf("sudo -S mount -t nfs %s:%s %s", ip, share, local)
							}
							AddErrorLog(a, "Mount Error test failed: "+err.Error(), cmdStr, err.Error(), 1)
						}
					}()
				}),
				fyne.NewMenuItem("Network Error", func() {
					go func() {
						AddInfoLog(a, "Running Test Error...")
						url := "http://255.255.255.255:8080/test"
						req, _ := http.NewRequest("GET", url, nil)
						c := client.NewHTTPClient(a)
						_, err := c.DoRequest(req)
						if err != nil {
							AddErrorLog(a, "Network Error test failed: "+err.Error(), "GET "+url, err.Error(), 1)
						}
					}()
				}),
			),
		},
	)
	debugBtn := newMenuButton("Debug Mode \u25be", debugMenu)

	return buildSettingsForm(ip, port, share, local, pwd, saveBtn, debugBtn)
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

func buildSettingsForm(ip, port, share, local, pwd *widget.Entry, saveBtn *widget.Button, debugBtn *menuButton) fyne.CanvasObject {
	header := container.NewHBox(
		widget.NewLabelWithStyle("Server Config", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		debugBtn,
	)
	return container.NewPadded(container.NewVBox(
		header,
		widget.NewLabel("Tailscale IP:"), ip,
		widget.NewLabel("Port:"), port,
		widget.NewLabel("Share Name:"), share,
		widget.NewLabel("Local Mount Path:"), local,
		widget.NewLabel("App Lock Password:"), pwd,
		saveBtn,
	))
}

func getDefaultMountPath() string {
	if runtime.GOOS == "windows" {
		return "Z:"
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "HomeNAS")
}

type menuButton struct {
	widget.Button
	menu *fyne.Menu
}

func (b *menuButton) Tapped(e *fyne.PointEvent) {
	widget.ShowPopUpMenuAtPosition(b.menu, fyne.CurrentApp().Driver().AllWindows()[0].Canvas(), e.AbsolutePosition)
}

func newMenuButton(label string, menu *fyne.Menu) *menuButton {
	b := &menuButton{menu: menu}
	b.Text = label
	b.ExtendBaseWidget(b)
	return b
}
