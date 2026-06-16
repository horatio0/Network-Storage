package config

type AppConfig struct {
	ListenAddr string `json:"listenAddr"`
}

func Default() AppConfig {
	return AppConfig{
		ListenAddr: ":8080",
	}
}
