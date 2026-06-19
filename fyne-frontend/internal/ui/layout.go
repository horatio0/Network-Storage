package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"reverseproxy-poc/internal/client"
)

type SidebarTab struct {
	Container *fyne.Container
	Bg        *canvas.Rectangle
	Lbl       *canvas.Text
	Dot       *canvas.Circle
}

var (
	sidebarTabs []*SidebarTab
	currentTab  int
)

func SetupMainWindow(a fyne.App, w fyne.Window, c *client.HTTPClient) {
	contentArea := container.NewMax()
	sidebar := createSidebar(a, contentArea, c, w)
	setupLogCallback(a, contentArea, c, w)
	handleSidebarSelect(0, a, contentArea, c, w)
	mainLayout := container.NewBorder(nil, nil, sidebar, nil, contentArea)
	w.SetContent(mainLayout)
}

func setupLogCallback(a fyne.App, contentArea *fyne.Container, c *client.HTTPClient, w fyne.Window) {
	OnLogAdded = func() {
		fyne.Do(func() {
			updateSidebarState(currentTab)
			if currentTab == 1 {
				contentArea.Objects = nil
				loadViewForTab(1, a, contentArea, c, w)
				contentArea.Refresh()
			}
		})
	}
}

func createSidebar(a fyne.App, cArea *fyne.Container, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	titles := []string{"Main", "Logs", "Files", "Terminal", "Screen", "Settings"}
	sidebarTabs = make([]*SidebarTab, 6)
	for i, t := range titles {
		sidebarTabs[i] = buildTabItem(t, i, a, cArea, c, w)
	}
	topBox := container.NewVBox(sidebarTabs[0].Container, sidebarTabs[1].Container, sidebarTabs[2].Container, sidebarTabs[3].Container, sidebarTabs[4].Container)
	botBox := container.NewVBox(widget.NewSeparator(), sidebarTabs[5].Container)
	sidebar := container.NewBorder(nil, botBox, nil, nil, topBox)
	return container.NewBorder(nil, nil, nil, widget.NewSeparator(), sidebar)
}

func buildTabItem(t string, idx int, a fyne.App, cArea *fyne.Container, c *client.HTTPClient, w fyne.Window) *SidebarTab {
	bg := canvas.NewRectangle(color.Transparent)
	lbl := canvas.NewText(t, color.White)

	dot := canvas.NewCircle(color.NRGBA{R: 255, A: 255})
	dotCont := container.NewGridWrap(fyne.NewSize(6, 6), dot)
	hbox := container.NewHBox(lbl, container.NewCenter(dotCont))

	tap := newTappable(func() { handleSidebarSelect(idx, a, cArea, c, w) })
	cont := container.NewStack(bg, container.NewPadded(hbox), tap)
	return &SidebarTab{Container: cont, Bg: bg, Lbl: lbl, Dot: dot}
}

func updateSidebarState(selected int) {
	for i, tab := range sidebarTabs {
		applyTabStyle(i, selected, tab.Bg, tab.Lbl, tab.Dot)
		tab.Container.Refresh()
	}
}

func applyTabStyle(i, selected int, bg *canvas.Rectangle, lbl *canvas.Text, dot *canvas.Circle) {
	dot.Hide()
	bg.FillColor = color.Transparent
	lbl.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if i == selected {
		bg.FillColor = color.NRGBA{R: 50, G: 50, B: 50, A: 255}
		lbl.Color = color.White
	}
	if i == 1 && HasNewLogs {
		dot.Show()
	}
}

func handleSidebarSelect(idx int, a fyne.App, cArea *fyne.Container, c *client.HTTPClient, w fyne.Window) {
	currentTab = idx
	if idx == 1 {
		HasNewLogs = false
	}
	updateSidebarState(idx)
	cArea.Objects = nil
	loadViewForTab(idx, a, cArea, c, w)
	cArea.Refresh()
}

func loadViewForTab(idx int, a fyne.App, cArea *fyne.Container, c *client.HTTPClient, w fyne.Window) {
	if idx == 0 {
		cArea.Add(createMainView(a, c, w))
		return
	}
	if idx == 1 {
		cArea.Add(createLogsView(a, nil))
		return
	}
	if idx == 2 {
		cArea.Add(createFilesView(a, c, w))
		return
	}
	loadViewForTabRight(idx, a, cArea, c, w)
}

func loadViewForTabRight(idx int, a fyne.App, cArea *fyne.Container, c *client.HTTPClient, w fyne.Window) {
	if idx == 3 {
		cArea.Add(createTerminalView(a, w))
		return
	}
	if idx == 4 {
		cArea.Add(createScreenView(a, c))
		return
	}
	if idx == 5 {
		cArea.Add(createSettingsView(a))
		return
	}
}
