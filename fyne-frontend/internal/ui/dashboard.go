package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"reverseproxy-poc/internal/client"
)

var (
	lastMonErr     string
	lastMonErrTime time.Time
	cpuChart       *lineChart
	memChart       *lineChart
	tempChart      *lineChart
)

func createMainView(a fyne.App, c *client.HTTPClient) fyne.CanvasObject {
	cpuLbl := widget.NewLabel("CPU: - %")
	memLbl := widget.NewLabel("Mem: - GB")
	tempLbl := widget.NewLabel("Temp: - °C")
	cpuChart = newLineChart(100)
	memChart = newLineChart(100)
	tempChart = newLineChart(100)

	mBtn := widget.NewButton("", nil)
	updateMountBtn(a, mBtn)

	go startDashboardLoop(a, c, cpuLbl, memLbl, tempLbl)
	return buildMainUI(cpuLbl, memLbl, tempLbl, mBtn)
}

func buildMainUI(cpu, mem, temp *widget.Label, mBtn *widget.Button) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Main Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	cpuBox := container.NewBorder(cpu, nil, nil, nil, cpuChart)
	memBox := container.NewBorder(mem, nil, nil, nil, memChart)
	tempBox := container.NewBorder(temp, nil, nil, nil, tempChart)
	grid := container.NewGridWithColumns(3, cpuBox, memBox, tempBox)
	card := widget.NewCard("Status", "", grid)
	return container.NewPadded(container.NewVBox(title, mBtn, card))
}

func updateMountBtn(a fyne.App, btn *widget.Button) {
	isMounted := a.Preferences().BoolWithFallback("is_mounted", false)
	if isMounted {
		btn.SetText("Unmount Server")
		btn.OnTapped = func() { executeUnmount(a, btn) }
	} else {
		btn.SetText("Mount Server")
		btn.OnTapped = func() { executeMount(a, btn) }
	}
}

func executeMount(a fyne.App, btn *widget.Button) {
	ip, share, local := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("share_name", "shared"), a.Preferences().StringWithFallback("mount_path", "")
	if err := client.MountDrive(ip, share, local); err != nil {
		AddLog(a, "Mount Error: "+err.Error())
	} else {
		AddLog(a, "Mounted to "+local)
		a.Preferences().SetBool("is_mounted", true)
		updateMountBtn(a, btn)
	}
}

func executeUnmount(a fyne.App, btn *widget.Button) {
	local := a.Preferences().StringWithFallback("mount_path", "")
	if err := client.UnmountDrive(local); err != nil {
		AddLog(a, "Unmount Error: "+err.Error())
	} else {
		AddLog(a, "Unmounted "+local)
		a.Preferences().SetBool("is_mounted", false)
		updateMountBtn(a, btn)
	}
}

func startDashboardLoop(a fyne.App, c *client.HTTPClient, cpu, mem, temp *widget.Label) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		updateDashboardData(a, c, cpu, mem, temp)
	}
}

func updateDashboardData(a fyne.App, c *client.HTTPClient, cpu, mem, temp *widget.Label) {
	ip, port := a.Preferences().StringWithFallback("server_ip", ""), a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		logMonitorErr(a, "Please set Server IP")
		return
	}
	s, err := client.FetchSystemStatus(c, ip, port)
	if err != nil {
		logMonitorErr(a, err.Error())
		return
	}
	lastMonErr = ""
	updateLabels(s, cpu, mem, temp)
}

func logMonitorErr(a fyne.App, e string) {
	if lastMonErr != e || time.Since(lastMonErrTime) > 10*time.Second {
		fyne.Do(func() { AddLog(a, "Monitor Error: "+e) })
		lastMonErr = e
		lastMonErrTime = time.Now()
	}
}

func updateLabels(s *client.SystemStatus, cpu, mem, temp *widget.Label) {
	memGB := float64(s.MemUsed) / (1024 * 1024 * 1024)
	totGB := float64(s.MemTotal) / (1024 * 1024 * 1024)
	fyne.Do(func() {
		cpu.SetText(fmt.Sprintf("CPU: %.1f %%", s.CPUPercent))
		mem.SetText(fmt.Sprintf("Mem: %.1f / %.1f GB (%.1f %%)", memGB, totGB, s.MemPercent))
		temp.SetText(fmt.Sprintf("Temp: %.1f °C", s.Temp))
		cpuChart.appendData(s.CPUPercent)
		memChart.appendData(s.MemPercent)
		tempChart.appendData(s.Temp)
	})
}
