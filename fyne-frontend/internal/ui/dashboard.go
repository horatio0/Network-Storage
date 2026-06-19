package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"reverseproxy-poc/internal/client"
)

func createDashboardView(a fyne.App, c *client.HTTPClient) fyne.CanvasObject {
	cpuLbl := widget.NewLabel("CPU: - %")
	memLbl := widget.NewLabel("Mem: - GB")
	tempLbl := widget.NewLabel("Temp: - °C")
	errLbl := widget.NewLabel("")

	go startDashboardLoop(a, c, cpuLbl, memLbl, tempLbl, errLbl)

	return buildDashboardUI(cpuLbl, memLbl, tempLbl, errLbl)
}

func buildDashboardUI(cpu, mem, temp, errLbl *widget.Label) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Server Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	card := widget.NewCard("Status", "", container.NewVBox(cpu, mem, temp))
	return container.NewPadded(container.NewVBox(title, card, errLbl))
}

func startDashboardLoop(a fyne.App, c *client.HTTPClient, cpu, mem, temp, errLbl *widget.Label) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		updateDashboardData(a, c, cpu, mem, temp, errLbl)
	}
}

func updateDashboardData(a fyne.App, c *client.HTTPClient, cpu, mem, temp, errLbl *widget.Label) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		errLbl.SetText("Error: Please set Server IP in Settings")
		return
	}
	s, err := client.FetchSystemStatus(c, ip, port)
	if err != nil {
		errLbl.SetText(fmt.Sprintf("Error: %v", err))
		return
	}
	updateLabels(s, cpu, mem, temp, errLbl)
}

func updateLabels(s *client.SystemStatus, cpu, mem, temp, errLbl *widget.Label) {
	cpu.SetText(fmt.Sprintf("CPU: %.1f %%", s.CPUPercent))
	memGB := float64(s.MemUsed) / (1024 * 1024 * 1024)
	totGB := float64(s.MemTotal) / (1024 * 1024 * 1024)
	mem.SetText(fmt.Sprintf("Mem: %.1f / %.1f GB (%.1f %%)", memGB, totGB, s.MemPercent))
	temp.SetText(fmt.Sprintf("Temp: %.1f °C", s.Temp))
	errLbl.SetText("")
}
