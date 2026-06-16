package config

import "fmt"

func Validate(cfg AppConfig) error {
	if cfg.ListenAddr == "" {
		return fmt.Errorf("ListenAddr is required")
	}
	return nil
}
