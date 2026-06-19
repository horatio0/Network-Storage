package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"reverseproxy-poc/internal/client"
)

func createFilesView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	statusLbl := widget.NewLabel("Status: Ready")
	mBtn := widget.NewButton("Mount", func() { executeMount(a, statusLbl) })
	uBtn := widget.NewButton("Unmount", func() { executeUnmount(a, statusLbl) })
	upBtn := widget.NewButton("Upload File", func() { triggerUpload(a, c, w, statusLbl) })
	dlEntry := widget.NewEntry()
	dlEntry.SetPlaceHolder("filename.txt")
	dlBtn := widget.NewButton("Download", func() { triggerDownload(a, c, dlEntry.Text, statusLbl) })
	return buildFilesForm(statusLbl, container.NewHBox(mBtn, uBtn), upBtn, dlEntry, dlBtn)
}

func buildFilesForm(statusLbl *widget.Label, mBox fyne.CanvasObject, upBtn *widget.Button, dlEntry *widget.Entry, dlBtn *widget.Button) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("File Manager", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	dlBox := container.NewHBox(widget.NewLabel("Download:"), dlEntry, dlBtn)
	card := widget.NewCard("Actions", "", container.NewVBox(mBox, upBtn, dlBox))
	return container.NewPadded(container.NewVBox(title, card, statusLbl))
}

func executeMount(a fyne.App, statusLbl *widget.Label) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	share := a.Preferences().StringWithFallback("share_name", "shared")
	local := a.Preferences().StringWithFallback("mount_path", "")
	err := client.MountDrive(ip, share, local)
	if err != nil {
		statusLbl.SetText("Mount Failed: " + err.Error())
	} else {
		statusLbl.SetText("Mounted to " + local)
	}
}

func executeUnmount(a fyne.App, statusLbl *widget.Label) {
	local := a.Preferences().StringWithFallback("mount_path", "")
	err := client.UnmountDrive(local)
	if err != nil {
		statusLbl.SetText("Unmount Failed: " + err.Error())
	} else {
		statusLbl.SetText("Unmounted " + local)
	}
}

func triggerUpload(a fyne.App, c *client.HTTPClient, w fyne.Window, statusLbl *widget.Label) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		go uploadWorker(c, ip, port, reader.URI().Path(), statusLbl)
	}, w)
}

func uploadWorker(c *client.HTTPClient, ip, port, path string, lbl *widget.Label) {
	err := client.UploadFile(c, ip, port, path)
	fyne.Do(func() {
		if err != nil {
			lbl.SetText("Upload Error: " + err.Error())
		} else {
			lbl.SetText("Uploaded: " + filepath.Base(path))
		}
	})
}

func triggerDownload(a fyne.App, c *client.HTTPClient, filename string, lbl *widget.Label) {
	if filename == "" {
		return
	}
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	home, _ := os.UserHomeDir()
	savePath := filepath.Join(home, "Downloads", filename)
	go downloadWorker(c, ip, port, filename, savePath, lbl)
}

func downloadWorker(c *client.HTTPClient, ip, port, file, save string, lbl *widget.Label) {
	err := client.DownloadFile(c, ip, port, file, save)
	fyne.Do(func() {
		if err != nil {
			lbl.SetText("Download Error: " + err.Error())
		} else {
			lbl.SetText("Downloaded to: " + save)
		}
	})
}
