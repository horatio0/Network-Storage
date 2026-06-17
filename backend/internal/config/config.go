package config

type TailscaleConfig struct {
	Enabled      bool     `json:"enabled"`
	AllowedUsers []string `json:"allowedUsers"`
}

type AppConfig struct {
	ListenAddr string          `json:"listenAddr"`
	MountPath  string          `json:"mountPath"`
	Tailscale  TailscaleConfig `json:"tailscale"`
}
