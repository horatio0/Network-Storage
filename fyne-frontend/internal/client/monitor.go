package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SystemStatus represents the JSON response from /api/v1/monitor.
type SystemStatus struct {
	CPUPercent float64 `json:"cpuPercent"`
	MemTotal   uint64  `json:"memTotal"`
	MemUsed    uint64  `json:"memUsed"`
	MemPercent float64 `json:"memPercent"`
	Temp       float64 `json:"temp"`
}

// FetchSystemStatus calls the monitor API.
func FetchSystemStatus(c *HTTPClient, ip, port string) (*SystemStatus, error) {
	url := fmt.Sprintf("http://%s:%s/api/v1/monitor", ip, port)
	req, _ := http.NewRequest("GET", url, nil)
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var s SystemStatus
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

type Device struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
	OS   string   `json:"os"`
}

type deviceResponse struct {
	Devices []Device `json:"devices"`
}

func FetchDevices(c *HTTPClient, ip, port string) ([]Device, error) {
	url := fmt.Sprintf("http://%s:%s/api/v1/tailscale/devices", ip, port)
	req, _ := http.NewRequest("GET", url, nil)
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var s deviceResponse
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, err
	}
	return s.Devices, nil
}
