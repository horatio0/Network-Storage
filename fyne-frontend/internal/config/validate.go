package config

import "errors"

func Validate(cfg AppConfig) error {
	if cfg.ListenAddr == "" {
		return errors.New("listen address is required")
	}

	return nil
}
