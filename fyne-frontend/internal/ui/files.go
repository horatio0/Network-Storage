package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"reverseproxy-poc/internal/client"
)

var currentPath string = "/"

func createFilesView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	upBtn := widget.NewButton("Upload File", func() { triggerUpload(a, c, w) })
	pathLbl := widget.NewLabel("Path: " + currentPath)
	vbox := container.NewVBox()

	go refreshFileList(a, c, vbox, pathLbl)

	top := container.NewHBox(upBtn, pathLbl)
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, container.NewScroll(container.NewPadded(vbox)))
	return container.NewBorder(top, nil, nil, nil, stack)
}

func refreshFileList(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return
	}
	files, err := client.ListFiles(c, ip, port, currentPath)
	if err != nil {
		fyne.Do(func() { AddLog(a, "File List Error: "+err.Error()) })
		return
	}
	buildFileBrowser(a, c, vbox, pathLbl, files)
}

func buildFileBrowser(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, files []client.FileInfo) {
	var items []client.FileInfo
	if currentPath != "/" {
		items = append(items, client.FileInfo{Name: "..", IsDir: true})
	}
	items = append(items, files...)

	fyne.Do(func() {
		pathLbl.SetText("Path: " + currentPath)
		configureFileList(a, c, vbox, pathLbl, items)
		vbox.Refresh()
	})
}

func configureFileList(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, items []client.FileInfo) {
	vbox.Objects = nil
	for _, item := range items {
		i := item
		prefix := "📄 "
		if i.IsDir {
			prefix = "📁 "
		}
		t := canvas.NewText(prefix+i.Name, color.White)
		tap := newTappable(func() { handleFileClick(a, c, vbox, pathLbl, i) })
		vbox.Add(container.NewPadded(container.NewStack(t, tap)))
	}
}

func handleFileClick(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, f client.FileInfo) {
	if !f.IsDir {
		triggerDownload(a, c, f.Name)
		return
	}
	updateCurrentPath(f.Name)
	go refreshFileList(a, c, vbox, pathLbl)
}

func updateCurrentPath(name string) {
	if name == ".." {
		moveUpDirectory()
		return
	}
	if currentPath == "/" {
		currentPath = "/" + name
	} else {
		currentPath = currentPath + "/" + name
	}
}

func moveUpDirectory() {
	parts := strings.Split(strings.TrimSuffix(currentPath, "/"), "/")
	currentPath = "/" + strings.Join(parts[:len(parts)-1], "/")
	if currentPath == "" || currentPath == "//" {
		currentPath = "/"
	}
}

func triggerUpload(a fyne.App, c *client.HTTPClient, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		go uploadWorker(a, c, ip, port, reader.URI().Path())
	}, w)
}

func uploadWorker(a fyne.App, c *client.HTTPClient, ip, port, path string) {
	err := client.UploadFile(c, ip, port, path)
	fyne.Do(func() {
		if err != nil {
			AddLog(a, "Upload Error: "+err.Error())
		} else {
			AddLog(a, "Uploaded: "+filepath.Base(path))
		}
	})
}

func triggerDownload(a fyne.App, c *client.HTTPClient, filename string) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	home, _ := os.UserHomeDir()
	savePath := filepath.Join(home, "Downloads", filename)
	go downloadWorker(a, c, ip, port, filename, savePath)
}

func downloadWorker(a fyne.App, c *client.HTTPClient, ip, port, file, save string) {
	err := client.DownloadFile(c, ip, port, file, save)
	fyne.Do(func() {
		if err != nil {
			AddLog(a, "Download Error: "+err.Error())
		} else {
			AddLog(a, "Downloaded to: "+save)
		}
	})
}
