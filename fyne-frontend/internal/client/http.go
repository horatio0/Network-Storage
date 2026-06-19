package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
)

// HTTPClient wraps the standard http.Client to provide common 401/403 handling.
type HTTPClient struct {
	client *http.Client
	app    fyne.App
}

// NewHTTPClient creates a new tailored HTTPClient.
func NewHTTPClient(a fyne.App) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 10 * time.Second},
		app:    a,
	}
}

// DoRequest performs an HTTP request and handles errors automatically.
func (c *HTTPClient) DoRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if err := checkAuthError(resp.StatusCode); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func checkAuthError(status int) error {
	if status == http.StatusUnauthorized {
		return fmt.Errorf("auth error (401): Tailscale not running")
	}
	if status == http.StatusForbidden {
		return fmt.Errorf("auth error (403): Unauthorized device")
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("server error: %d", status)
	}
	return nil
}
