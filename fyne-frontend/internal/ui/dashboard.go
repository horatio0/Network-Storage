package ui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"network-storage-client/internal/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var updateMainLogs func()
var dashboardCancel context.CancelFunc

var (
	lastMonErr  string
	lastLogTime time.Time
	cpuChart    *lineChart
	memChart    *lineChart
	tempChart   *lineChart
)

func createMainView(a fyne.App, c *client.HTTPClient, w fyne.Window) fyne.CanvasObject {
	cpuLbl := widget.NewLabel("CPU: - %")
	memLbl := widget.NewLabel("Mem: - GB")
	tempLbl := widget.NewLabel("Temp: - °C")
	var oldCPUData, oldMemData, oldTempData []float64
	if cpuChart != nil {
		oldCPUData = cpuChart.data
	}
	if memChart != nil {
		oldMemData = memChart.data
	}
	if tempChart != nil {
		oldTempData = tempChart.data
	}

	cpuChart = newLineChart(100)
	cpuChart.data = oldCPUData
	memChart = newLineChart(100)
	memChart.data = oldMemData
	tempChart = newLineChart(100)
	tempChart.data = oldTempData

	mBtn := widget.NewButton("", nil)
	updateMountBtn(a, mBtn)

	devBox := container.NewVBox()

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		ip := a.Preferences().StringWithFallback("server_ip", "")
		port := a.Preferences().StringWithFallback("server_port", "8080")
		if ip != "" {
			go fetchAndUpdateDevs(a, c, ip, port, devBox, w)
		}
	})

	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip != "" {
		go fetchAndUpdateDevs(a, c, ip, port, devBox, w)
	}

	if dashboardCancel != nil {
		dashboardCancel()
	}
	var ctx context.Context
	ctx, dashboardCancel = context.WithCancel(context.Background())

	go startDashboardLoop(ctx, a, c, cpuLbl, memLbl, tempLbl, devBox, w)
	return buildMainUI(a, cpuLbl, memLbl, tempLbl, mBtn, devBox, refreshBtn)
}

func buildMainUI(a fyne.App, cpu, mem, temp *widget.Label, mBtn *widget.Button, devBox *fyne.Container, refreshBtn *widget.Button) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Main Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	cpuBox := container.NewBorder(cpu, nil, nil, nil, cpuChart)
	memBox := container.NewBorder(mem, nil, nil, nil, memChart)
	tempBox := container.NewBorder(temp, nil, nil, nil, tempChart)
	grid := container.NewGridWithColumns(3, cpuBox, memBox, tempBox)
	card := widget.NewCard("Status", "", grid)

	bg := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	devScroll := container.NewPadded(container.NewScroll(devBox))
	
	devTitle := container.NewHBox(
		widget.NewLabelWithStyle("Connected Devices", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		refreshBtn,
	)
	devContent := container.NewBorder(devTitle, nil, nil, nil, container.NewStack(bg, devScroll))
	
	devCard := widget.NewCard("", "", devContent)
	
	logsContainer := container.NewStack(createLogsView(a, nil))
	logsCard := widget.NewCard("", "", logsContainer)
	updateMainLogs = func() {
		logsContainer.Objects = []fyne.CanvasObject{createLogsView(a, nil)}
		logsContainer.Refresh()
	}

	bottomGrid := container.NewGridWithColumns(2, devCard, logsCard)

	top := container.NewVBox(title, mBtn)
	return container.NewPadded(container.NewBorder(top, card, nil, nil, bottomGrid))
}

func showSudoDialog(a fyne.App, btn *widget.Button, errMsg string, isMount bool) {
	pwdEntry := widget.NewPasswordEntry()
	pwdEntry.PlaceHolder = "Enter Sudo Password"

	var vbox *fyne.Container
	if errMsg != "" {
		errLbl := widget.NewLabelWithStyle(errMsg, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		errLbl.Importance = widget.DangerImportance
		vbox = container.NewVBox(errLbl, pwdEntry)
	} else {
		vbox = container.NewVBox(pwdEntry)
	}

	content := container.NewPadded(vbox)

	win := a.Driver().AllWindows()[0]
	var d dialog.Dialog
	
	submitFunc := func() {
		client.SetSudoPassword(pwdEntry.Text)
		if isMount {
			executeMount(a, btn)
		} else {
			executeUnmount(a, btn)
		}
		if d != nil {
			d.Hide()
		}
	}
	
	pwdEntry.OnSubmitted = func(s string) {
		submitFunc()
	}

	d = dialog.NewCustomConfirm("Sudo Password Required", "OK", "Cancel", content, func(ok bool) {
		if ok {
			submitFunc()
		}
	}, win)
	
	d.Resize(fyne.NewSize(400, 150))
	d.Show()
	win.Canvas().Focus(pwdEntry)
}

func updateMountBtn(a fyne.App, btn *widget.Button) {
	isMounted := a.Preferences().BoolWithFallback("is_mounted", false)
	if isMounted {
		btn.SetText("Unmount Filesystem")
		btn.OnTapped = func() { executeUnmount(a, btn) }
	} else {
		btn.SetText("Mount Filesystem")
		btn.OnTapped = func() { executeMount(a, btn) }
	}
}

func executeMount(a fyne.App, btn *widget.Button) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	share := a.Preferences().StringWithFallback("share_name", "/NS/share")
	local := a.Preferences().StringWithFallback("mount_path", "")

	if runtime.GOOS == "windows" {
		matched, _ := regexp.MatchString(`^[a-zA-Z]:$`, local)
		if !matched {
			win := a.Driver().AllWindows()[0]
			dialog.ShowError(errors.New("윈도우에서는 Z:와 같은 드라이브 문자를 사용해야 합니다"), win)
			return
		}
	}
	
	go func() {
		err := client.MountDrive(ip, share, local)
		if err == client.ErrPasswordRequired {
			fyne.Do(func() {
				showSudoDialog(a, btn, "", true)
			})
			return
		}

		fyne.Do(func() {
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "incorrect password") {
					client.SetSudoPassword("")
					showSudoDialog(a, btn, "Incorrect password. Please try again.", true)
				} else {
					cmdStr := ""
					if runtime.GOOS == "windows" {
						cmdStr = fmt.Sprintf(`net use %s \\%s\%s`, local, ip, share)
					} else {
						cmdStr = fmt.Sprintf("sudo -S mount -t nfs %s:%s %s", ip, share, local)
					}
					AddErrorLog(a, "Mount Error: "+err.Error(), cmdStr, err.Error(), 0)
				}
			} else {
				AddInfoLog(a, "Mounted to "+local)
				a.Preferences().SetBool("is_mounted", true)
				updateMountBtn(a, btn)
			}
		})
	}()
}

func executeUnmount(a fyne.App, btn *widget.Button) {
	local := a.Preferences().StringWithFallback("mount_path", "")
	
	go func() {
		err := client.UnmountDrive(local)
		if err == client.ErrPasswordRequired {
			fyne.Do(func() {
				showSudoDialog(a, btn, "", false)
			})
			return
		}

		fyne.Do(func() {
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "incorrect password") {
					client.SetSudoPassword("")
					showSudoDialog(a, btn, "Incorrect password. Please try again.", false)
				} else {
					cmdStr := ""
					if runtime.GOOS == "windows" {
						cmdStr = fmt.Sprintf(`net use %s /delete /y`, local)
					} else {
						cmdStr = fmt.Sprintf("sudo -S umount -l %s", local)
					}
					AddErrorLog(a, "Unmount Error: "+err.Error(), cmdStr, err.Error(), 0)
				}
			} else {
				AddInfoLog(a, "Unmounted from "+local)
				a.Preferences().SetBool("is_mounted", false)
				updateMountBtn(a, btn)
			}
		})
	}()
}

func startDashboardLoop(mainCtx context.Context, a fyne.App, c *client.HTTPClient, cpu, mem, temp *widget.Label, devBox *fyne.Container, w fyne.Window) {
	for {
		select {
		case <-mainCtx.Done():
			return
		default:
		}

		ip := a.Preferences().StringWithFallback("server_ip", "")
		port := a.Preferences().StringWithFallback("server_port", "8080")
		if ip == "" {
			logMonitorErr(a, "Please set Server IP", "")
			select {
			case <-mainCtx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}

		streamCtx, cancelStream := context.WithCancel(mainCtx)
		
		// Watch for IP/Port changes to cancel stream
		go func(currentIP, currentPort string) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-streamCtx.Done():
					return
				case <-ticker.C:
					newIP := a.Preferences().StringWithFallback("server_ip", "")
					newPort := a.Preferences().StringWithFallback("server_port", "8080")
					if newIP != currentIP || newPort != currentPort {
						cancelStream()
						return
					}
				}
			}
		}(ip, port)

		client.MonitorStream(streamCtx, ip, port, func(s *client.SystemStatus) {
			lastMonErr = ""
			updateLabels(s, cpu, mem, temp)
		}, func(err error) {
			cmdStr := fmt.Sprintf("GET http://%s:%s/api/v1/monitor/stream", ip, port)
			logMonitorErr(a, err.Error(), cmdStr)
		})
		
		cancelStream()
		select {
		case <-mainCtx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func fetchAndUpdateDevs(a fyne.App, c *client.HTTPClient, ip, port string, devBox *fyne.Container, w fyne.Window) {
	devs, _ := client.FetchDevices(c, ip, port)
	if devs != nil {
		sort.Slice(devs, func(i, j int) bool {
			ipI, ipJ := "", ""
			if len(devs[i].IPs) > 0 { ipI = devs[i].IPs[0] }
			if len(devs[j].IPs) > 0 { ipJ = devs[j].IPs[0] }
			return ipI < ipJ
		})
		updateDevices(a, devBox, devs, w)
	}
}

func logMonitorErr(a fyne.App, e, cmd string) {
	fyne.Do(func() {
		if lastMonErr != e {
			AddErrorLog(a, "Monitor Error: "+e, cmd, e, 0)
			lastLogTime = time.Now()
			noti := fyne.NewNotification("Error", "에러가 발생했습니다. Logs탭에서 확인해 주세요")
			go a.SendNotification(noti)
			lastMonErr = e
		} else if time.Since(lastLogTime) > 10*time.Second {
			AddErrorLog(a, "Monitor Error: "+e, cmd, e, 0)
			lastLogTime = time.Now()
		}
	})
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

func updateDevices(a fyne.App, devBox *fyne.Container, devs []client.Device, w fyne.Window) {
	fyne.Do(func() {
		var objects []fyne.CanvasObject
		for i, d := range devs {
			objects = append(objects, buildDeviceRow(a, i, d, w))
		}
		devBox.Objects = objects
		devBox.Refresh()
	})
}

func buildDeviceRow(a fyne.App, i int, d client.Device, w fyne.Window) fyne.CanvasObject {
	ipStr := ""
	if len(d.IPs) > 0 {
		ipStr = d.IPs[0]
	}
	key := "alias_" + ipStr
	
	defaultAlias := d.Name
	if parts := strings.Split(defaultAlias, "."); len(parts) > 0 {
		defaultAlias = parts[0]
	}
	if defaultAlias == "" {
		defaultAlias = fmt.Sprintf("device%d", i+1)
	}
	alias := a.Preferences().StringWithFallback(key, defaultAlias)

	lbl := canvas.NewText(fmt.Sprintf("%s [%s] %s", d.OS, ipStr, alias), color.White)
	btn := widget.NewButton("✏️", nil)
	btn.OnTapped = func() { promptAlias(a, key, alias, w) }

	return container.NewBorder(nil, nil, nil, btn, lbl)
}

func promptAlias(a fyne.App, key, alias string, w fyne.Window) {
	entry := widget.NewEntry()
	entry.SetText(alias)
	dialog.ShowCustomConfirm("Edit Alias", "Save", "Cancel", entry, func(b bool) {
		if b {
			a.Preferences().SetString(key, entry.Text)
		}
	}, w)
}
