package ui

import (
	"image/color"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"network-storage-client/internal/client"
)

var currentPath string = "/"

var filesData []client.FileInfo

func createFilesView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	listContainer := container.NewStack()
	pathLbl := widget.NewLabel("Path: " + currentPath)

	upBtn := widget.NewButton("Upload", func() { triggerUpload(a, c, listContainer, pathLbl, w) })
	newBtn := widget.NewButton("New Directory", func() { promptMkdir(a, c, w, listContainer, pathLbl) })

	go refreshFileList(a, c, listContainer, pathLbl, w)

	top := container.NewHBox(upBtn, newBtn, pathLbl)
	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	stack := container.NewStack(bg, listContainer)
	return container.NewBorder(top, nil, nil, nil, stack)
}

func refreshFileList(a fyne.App, c *client.HTTPClient, listContainer *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return
	}
	files, err := client.ListFiles(c, ip, port, currentPath)
	if err != nil {
		fyne.Do(func() { AddLog(a, "File List Error: "+err.Error()) })
		return
	}
	buildFileBrowser(a, c, listContainer, pathLbl, files, w)
}

func buildFileBrowser(a fyne.App, c *client.HTTPClient, listContainer *fyne.Container, pathLbl *widget.Label, files []client.FileInfo, w fyne.Window) {
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
		filesData = items
		if len(listContainer.Objects) == 0 {
			listContainer.Objects = []fyne.CanvasObject{createFileWidgetList(a, c, listContainer, pathLbl, w)}
		} else {
			listContainer.Objects[0].(*widget.List).Refresh()
		}
		listContainer.Refresh()
	})
}

func createFileWidgetList(a fyne.App, c *client.HTTPClient, listContainer *fyne.Container, pathLbl *widget.Label, w fyne.Window) *widget.List {
	list := widget.NewList(
		func() int { return len(filesData) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.FolderIcon())
			label := canvas.NewText("", color.White)
			dlBtn := widget.NewButton("Download", nil)
			delBtn := widget.NewButton("Delete", nil)
			btnBox := container.NewHBox(dlBtn, delBtn)
			return container.NewBorder(nil, nil, icon, btnBox, label)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			item := filesData[id]
			border := o.(*fyne.Container)
			
			var icon *widget.Icon
			var label *canvas.Text
			var btnBox *fyne.Container
			for _, obj := range border.Objects {
				switch v := obj.(type) {
				case *widget.Icon:
					icon = v
				case *canvas.Text:
					label = v
				case *fyne.Container:
					btnBox = v
				}
			}

			label.Text = item.Name
			label.Refresh()

			if item.IsDir {
				icon.SetResource(theme.FolderIcon())
				icon.Show()
			} else {
				icon.Hide()
			}

			if item.Name == ".." {
				btnBox.Hide()
			} else {
				btnBox.Show()
				dlBtn := btnBox.Objects[0].(*widget.Button)
				delBtn := btnBox.Objects[1].(*widget.Button)
				
				if item.IsDir {
					dlBtn.Hide()
				} else {
					dlBtn.Show()
					dlBtn.OnTapped = func() { triggerDownload(a, c, item.Name) }
				}
				
				delBtn.OnTapped = func() { promptDelete(a, c, w, listContainer, pathLbl, item.Name) }
			}
		},
	)
	list.HideSeparators = true
	
	list.OnSelected = func(id widget.ListItemID) {
		list.Unselect(id)
		handleFileClick(a, c, listContainer, pathLbl, filesData[id], w)
	}
	
	return list
}

func handleFileClick(a fyne.App, c *client.HTTPClient, listContainer *fyne.Container, pathLbl *widget.Label, f client.FileInfo, w fyne.Window) {
	if !f.IsDir {
		return
	}
	updateCurrentPath(f.Name)
	go refreshFileList(a, c, listContainer, pathLbl, w)
}

func updateCurrentPath(name string) {
	currentPath = path.Join(currentPath, name)
}

func triggerUpload(a fyne.App, c *client.HTTPClient, listContainer *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		go uploadWorker(a, c, ip, port, reader.URI().Path(), currentPath, listContainer, pathLbl, w)
	}, w)
	d.Resize(fyne.NewSize(800, 600))
	d.Show()
}

func uploadWorker(a fyne.App, c *client.HTTPClient, ip, port, path, targetDir string, listContainer *fyne.Container, pathLbl *widget.Label, w fyne.Window) {
	err := client.UploadFile(c, ip, port, path, targetDir)
	fyne.Do(func() {
		if err != nil {
			AddLog(a, "Upload Error: "+err.Error())
		} else {
			AddLog(a, "Uploaded: "+filepath.Base(path))
		}
	})
	refreshFileList(a, c, listContainer, pathLbl, w)
}

func triggerDownload(a fyne.App, c *client.HTTPClient, filename string) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	home, _ := os.UserHomeDir()
	savePath := filepath.Join(home, "Downloads", filename)
	target := path.Join(currentPath, filename)
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
			target := path.Join(currentPath, entry.Text)
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
			target := path.Join(currentPath, name)
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
