package monitor

import (
	"context"
	"testing"
)

func TestGetSystemStatus(t *testing.T) {
	status, err := GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSystemStatus failed: %v", err)
	}

	if status.MemTotal == 0 {
		t.Log("Warning: MemTotal is 0, might be unsupported on this OS")
	}

	t.Logf("System Status: %+v", status)
}
