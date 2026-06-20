package client

import "errors"

var (
	// ErrPasswordRequired is returned when a sudo password is required but not provided.
	ErrPasswordRequired = errors.New("sudo password required")

	sudoPassword string
)

// SetSudoPassword caches the sudo password in memory.
func SetSudoPassword(password string) {
	sudoPassword = password
}

// GetSudoPassword retrieves the cached sudo password.
func GetSudoPassword() string {
	return sudoPassword
}
