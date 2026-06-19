package ui

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"time"
)

type LogEntry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
}

var HasNewLogs bool
var OnLogAdded func()

func AddLog(a fyne.App, msg string) {
	logs := append([]LogEntry{{Time: time.Now().Format("15:04:05"), Message: msg}}, LoadLogs(a)...)
	if len(logs) > 50 {
		logs = logs[:50]
	}
	b, _ := json.Marshal(logs)
	a.Preferences().SetString("system_logs", string(b))
	if currentTab != 1 {
		HasNewLogs = true
	}
	if OnLogAdded != nil {
		OnLogAdded()
	}
}

func LoadLogs(a fyne.App) []LogEntry {
	b := a.Preferences().StringWithFallback("system_logs", "[]")
	var logs []LogEntry
	json.Unmarshal([]byte(b), &logs)
	return logs
}

func ClearLogs(a fyne.App) {
	a.Preferences().SetString("system_logs", "[]")
	if OnLogAdded != nil {
		OnLogAdded()
	}
}
