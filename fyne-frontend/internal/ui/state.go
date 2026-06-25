package ui

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"time"
)

type LogEntry struct {
	Time     string `json:"time"`
	Message  string `json:"message"`
	Level    string `json:"level"`
	Command  string `json:"command,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
}

var HasNewLogs bool
var OnLogAdded func()

func AddInfoLog(a fyne.App, msg string) {
	addLogInternal(a, LogEntry{Time: time.Now().Format("15:04:05"), Message: msg, Level: "info"})
}

func AddWarnLog(a fyne.App, msg string) {
	addLogInternal(a, LogEntry{Time: time.Now().Format("15:04:05"), Message: msg, Level: "warn"})
}

func AddErrorLog(a fyne.App, msg string, cmd string, stderr string, exitCode int) {
	addLogInternal(a, LogEntry{Time: time.Now().Format("15:04:05"), Message: msg, Level: "error", Command: cmd, Stderr: stderr, ExitCode: exitCode})
}

func addLogInternal(a fyne.App, entry LogEntry) {
	logs := append(LoadLogs(a), entry)
	if len(logs) > 50 {
		logs = logs[len(logs)-50:]
	}
	b, _ := json.Marshal(logs)
	a.Preferences().SetString("system_logs", string(b))
	if currentTab != 1 && currentTab != 0 && entry.Level != "info" {
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
