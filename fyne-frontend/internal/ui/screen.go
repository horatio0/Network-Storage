package ui

import (
	"fmt"
	"image/color"
	"strings"

	"network-storage-client/internal/client"
	"network-storage-client/internal/webrtc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var rtc *webrtc.Controller
var hostMap map[string]string

func createScreenView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 10, G: 10, B: 10, A: 255})
	viewer := canvas.NewImageFromResource(nil)
	viewer.FillMode = canvas.ImageFillContain

	sel := widget.NewSelect([]string{"Loading..."}, nil)
	go loadScreenDevices(a, c, sel)
	return buildScreenLayout(a, sel, bg, viewer, w)
}

func buildScreenLayout(a fyne.App, s *widget.Select, bg *canvas.Rectangle, v *canvas.Image, w fyne.Window) fyne.CanvasObject {
	selBox := container.NewGridWrap(fyne.NewSize(300, 36), s)
	btn := widget.NewButton("View Remote Screen", func() {
		startRtc(a, w, hostMap[s.Selected], false, buildFrameCb(v))
	})
	top := container.NewHBox(widget.NewLabel("Target:"), selBox, btn)
	return container.NewBorder(top, nil, nil, nil, container.NewStack(bg, v))
}

func loadScreenDevices(a fyne.App, c *client.HTTPClient, sel *widget.Select) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return
	}
	devs, err := client.FetchDevices(c, ip, port)
	if err != nil {
		return
	}
	updateDeviceSelect(sel, devs)
}

func updateDeviceSelect(sel *widget.Select, devs []client.Device) {
	hostMap = make(map[string]string)
	var opts []string
	for _, d := range devs {
		parseDeviceOption(d, &opts)
	}
	fyne.Do(func() { sel.Options = opts; sel.Refresh() })
}

func parseDeviceOption(d client.Device, opts *[]string) {
	os := strings.ToLower(d.OS)
	if os == "windows" || os == "android" {
		lbl := fmt.Sprintf("%s [%s]", d.Name, d.IPs[0])
		*opts = append(*opts, lbl)
		hostMap[lbl] = d.Name
	}
}

func buildFrameCb(v *canvas.Image) func([]byte) {
	return func(data []byte) {
		fyne.Do(func() {
			v.Resource = fyne.NewStaticResource("frame.jpg", data)
			v.Refresh()
		})
	}
}

func startRtc(a fyne.App, w fyne.Window, target string, isHost bool, onFrame func([]byte)) {
	if rtc != nil || target == "" {
		return
	}
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return
	}
	startRtcController(a, w, ip, port, target, isHost, onFrame)
}

func startRtcController(a fyne.App, w fyne.Window, ip, port, tgt string, host bool, cb func([]byte)) {
	url := fmt.Sprintf("ws://%s:%s/api/v1/signaling/ws", ip, port)
	errCb := func(err error) {
		fyne.Do(func() {
			AddLog(a, "WebRTC Err: "+err.Error())
			dialog.ShowInformation("Connection Error", "WebRTC:\n"+err.Error(), w)
			rtc = nil
		})
	}
	c, err := webrtc.NewController(url, tgt, host, cb, errCb)
	if err != nil {
		errCb(err)
		return
	}
	rtc = c
	rtc.StartSignaling()
	fyne.Do(func() { AddLog(a, "WebRTC Started") })
}
