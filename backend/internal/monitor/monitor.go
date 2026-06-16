package monitor

import (
	"context"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
)

type SystemStatus struct {
	CPUPercent float64 `json:"cpuPercent"`
	MemTotal   uint64  `json:"memTotal"`
	MemUsed    uint64  `json:"memUsed"`
	MemPercent float64 `json:"memPercent"`
	Temp       float64 `json:"temp"` // Celsius
}

func GetSystemStatus(ctx context.Context) (SystemStatus, error) {
	var status SystemStatus

	cpuPercents, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuPercents) > 0 {
		status.CPUPercent = cpuPercents[0]
	}

	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		status.MemTotal = vmStat.Total
		status.MemUsed = vmStat.Used
		status.MemPercent = vmStat.UsedPercent
	}

	temps, err := sensors.TemperaturesWithContext(ctx)
	if err == nil {
		for _, t := range temps {
			if t.Temperature > status.Temp {
				status.Temp = t.Temperature
			}
		}
	}

	return status, nil
}
