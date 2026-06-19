package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"reverseproxy-poc/internal/client"
)

// SetupMainWindow configures the split layout with a sidebar.
func SetupMainWindow(a fyne.App, w fyne.Window, c *client.HTTPClient) {
	contentArea := container.NewMax()
	contentArea.Objects = []fyne.CanvasObject{
		createDashboardView(a, c),
	}

	sidebar := createSidebar(a, contentArea, c)
	split := container.NewHSplit(sidebar, contentArea)
	split.Offset = 0.2

	w.SetContent(split)
}

func createSidebar(a fyne.App, contentArea *fyne.Container, c *client.HTTPClient) fyne.CanvasObject {
	menuList := widget.NewList(
		func() int { return 2 },
		func() fyne.CanvasObject { return widget.NewLabel("Menu Item") },
		func(i widget.ListItemID, o fyne.CanvasObject) { updateSidebarItem(i, o) },
	)
	menuList.OnSelected = func(id widget.ListItemID) { handleSidebarSelect(a, id, contentArea, c) }
	menuList.Select(0)

	title := widget.NewLabelWithStyle("Control Panel", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	return container.NewBorder(title, nil, nil, nil, menuList)
}

func updateSidebarItem(i widget.ListItemID, o fyne.CanvasObject) {
	lbl := o.(*widget.Label)
	switch i {
	case 0:
		lbl.SetText("Dashboard")
	case 1:
		lbl.SetText("Settings")
	}
}

func handleSidebarSelect(a fyne.App, id widget.ListItemID, contentArea *fyne.Container, c *client.HTTPClient) {
	contentArea.Objects = nil
	switch id {
	case 0:
		contentArea.Add(createDashboardView(a, c))
	case 1:
		contentArea.Add(createSettingsView(a))
	}
	contentArea.Refresh()
}
