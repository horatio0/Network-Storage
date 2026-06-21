package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"central-control-backend/internal/config"
	"tailscale.com/client/tailscale"
)

func TestTailscaleAuthEnforced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := log.Default()
	cfg := config.AppConfig{
		Tailscale: config.TailscaleConfig{
			Enabled: false, // 설정과 무관하게 강제되어야 함
		},
	}

	tsClient := &tailscale.LocalClient{}

	router := gin.New()
	router.Use(TailscaleAuth(logger, cfg, tsClient))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
