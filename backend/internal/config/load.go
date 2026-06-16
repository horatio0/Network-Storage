package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func Load(path string) (AppConfig, error) {
	var cfg AppConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config file: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config file: %w", err)
	}
	return cfg, nil
}
