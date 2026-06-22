package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SystemStatus represents the JSON response from /api/v1/monitor.
type SystemStatus struct {
	CPUPercent float64 `json:"cpuPercent"`
	MemTotal   uint64  `json:"memTotal"`
	MemUsed    uint64  `json:"memUsed"`
	MemPercent float64 `json:"memPercent"`
	Temp       float64 `json:"temp"`
}



// MonitorStream connects to the monitor SSE stream.
func MonitorStream(ctx context.Context, ip, port string, onData func(*SystemStatus), onError func(error)) {
	url := fmt.Sprintf("http://%s:%s/api/v1/monitor/stream", ip, port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		onError(err)
		return
	}
	req.Header.Set("Accept", "text/event-stream")

	// Use a client without a timeout for streaming
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		onError(fmt.Errorf("stream network error: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		onError(fmt.Errorf("stream server error: %d", resp.StatusCode))
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if dataStr == "" {
				continue
			}
			var s SystemStatus
			if err := json.Unmarshal([]byte(dataStr), &s); err != nil {
				continue // ignore parse errors on stream
			}
			onData(&s)
		}
	}
	if err := scanner.Err(); err != nil && err != context.Canceled {
		onError(err)
	}
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
