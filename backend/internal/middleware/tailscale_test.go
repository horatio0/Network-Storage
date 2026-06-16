package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"reverseproxy-poc/internal/config"
)

func TestTailscaleAuthDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := log.Default()
	cfg := config.AppConfig{
		Tailscale: config.TailscaleConfig{
			Enabled: false,
		},
	}

	router := gin.New()
	router.Use(TailscaleAuth(logger, cfg))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
