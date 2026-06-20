package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"network-storage-client/internal/client"
)

var currentPath string = "/"

func createFilesView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	vbox := container.NewVBox()
	pathLbl := widget.NewLabel("Path: " + currentPath)

	upBtn := widget.NewButton("Upload", func() { triggerUpload(a, c, w) })
	newBtn := widget.NewButton("New Directory", func() { promptMkdir(a, c, w, vbox, pathLbl) })

	go refreshFileList(a, c, vbox, pathLbl, w)

	top := container.NewHBox(upBtn, newBtn, pathLbl)
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, container.NewScroll(container.NewPadded(vbox)))
	return container.NewBorder(top, nil, nil, nil, stack)
}

func refreshFileList(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return
	}
	files, err := client.ListFiles(c, ip, port, currentPath)
	if err != nil {
		fyne.Do(func() { AddLog(a, "File List Error: "+err.Error()) })
		return
	}
	buildFileBrowser(a, c, vbox, pathLbl, files, w)
}

func buildFileBrowser(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, files []client.FileInfo, w fyne.Window) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir && !files[j].IsDir {
			return true
		}
		if !files[i].IsDir && files[j].IsDir {
			return false
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	var items []client.FileInfo
	if currentPath != "/" {
		items = append(items, client.FileInfo{Name: "..", IsDir: true})
	}
	items = append(items, files...)

	fyne.Do(func() {
		pathLbl.SetText("Path: " + currentPath)
		configureFileList(a, c, vbox, pathLbl, items, w)
		vbox.Refresh()
	})
}

func configureFileList(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, items []client.FileInfo, w fyne.Window) {
	vbox.Objects = nil
	for _, i := range items {
		vbox.Add(createFileRow(a, c, vbox, pathLbl, i, w))
	}
}

func createFileRow(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, i client.FileInfo, w fyne.Window) fyne.CanvasObject {
	t := canvas.NewText(i.Name, color.White)
	
	var display fyne.CanvasObject = t
	if i.IsDir {
		img := canvas.NewImageFromFile("resources/directory.png")
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(20, 20))
		display = container.NewHBox(img, t)
	}
	
	tap := newTappable(func() { handleFileClick(a, c, vbox, pathLbl, i, w) })
	
	if i.Name == ".." {
		return container.NewStack(display, tap)
	}
	delBtn := widget.NewButton("Delete", func() { promptDelete(a, c, w, vbox, pathLbl, i.Name) })
	if !i.IsDir {
		dlBtn := widget.NewButton("Download", func() { triggerDownload(a, c, i.Name) })
		actions := container.NewHBox(dlBtn, delBtn)
		return container.NewBorder(nil, nil, nil, actions, container.NewStack(display, tap))
	}
	return container.NewBorder(nil, nil, nil, delBtn, container.NewStack(display, tap))
}

func handleFileClick(a fyne.App, c *client.HTTPClient, vbox *fyne.Container, pathLbl *widget.Label, f client.FileInfo, w fyne.Window) {
	if !f.IsDir {
		return
	}
	updateCurrentPath(f.Name)
	go refreshFileList(a, c, vbox, pathLbl, w)
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
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		go uploadWorker(a, c, ip, port, reader.URI().Path(), currentPath)
	}, w)
	d.Resize(fyne.NewSize(800, 600))
	d.Show()
}

func uploadWorker(a fyne.App, c *client.HTTPClient, ip, port, path, targetDir string) {
	err := client.UploadFile(c, ip, port, path, targetDir)
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
	target := currentPath
	if target == "/" {
		target = "/" + filename
	} else {
		target = target + "/" + filename
	}
	go downloadWorker(a, c, ip, port, target, savePath)
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

func promptMkdir(a fyne.App, c *client.HTTPClient, w fyne.Window, vbox *fyne.Container, pathLbl *widget.Label) {
	entry := widget.NewEntry()
	var d dialog.Dialog
	
	submitFunc := func() {
		if entry.Text != "" {
			target := currentPath
			if target == "/" {
				target = "/" + entry.Text
			} else {
				target = target + "/" + entry.Text
			}
			go executeMkdir(a, c, target, vbox, pathLbl, w)
		}
		if d != nil {
			d.Hide()
		}
	}

	entry.OnSubmitted = func(s string) {
		submitFunc()
	}

	d = dialog.NewCustomConfirm("New Directory", "Create", "Cancel", entry, func(b bool) {
		if b {
			submitFunc()
		}
	}, w)
	d.Resize(fyne.NewSize(400, 150))
	d.Show()
	w.Canvas().Focus(entry)
}

func executeMkdir(a fyne.App, c *client.HTTPClient, target string, vbox *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if err := client.Mkdir(c, ip, port, target); err != nil {
		fyne.Do(func() { AddLog(a, "Mkdir Error: "+err.Error()) })
	} else {
		fyne.Do(func() { AddLog(a, "Created Dir: "+target) })
		refreshFileList(a, c, vbox, pathLbl, w)
	}
}

func promptDelete(a fyne.App, c *client.HTTPClient, w fyne.Window, vbox *fyne.Container, pathLbl *widget.Label, name string) {
	dialog.ShowConfirm("Delete", "Delete "+name+"?", func(b bool) {
		if b {
			target := currentPath
			if target == "/" {
				target = "/" + name
			} else {
				target = target + "/" + name
			}
			go executeDelete(a, c, target, vbox, pathLbl, w)
		}
	}, w)
}

func executeDelete(a fyne.App, c *client.HTTPClient, target string, vbox *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if err := client.DeletePath(c, ip, port, target); err != nil {
		fyne.Do(func() { AddLog(a, "Delete Error: "+err.Error()) })
	} else {
		fyne.Do(func() { AddLog(a, "Deleted: "+target) })
		refreshFileList(a, c, vbox, pathLbl, w)
	}
}
