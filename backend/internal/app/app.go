package app

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"central-control-backend/internal/config"
	"central-control-backend/internal/files"
	"central-control-backend/internal/middleware"
	"central-control-backend/internal/monitor"
	"central-control-backend/internal/signaling"
	"central-control-backend/internal/tailscale"
	"central-control-backend/internal/terminal"
)

type App struct {
	logger       *log.Logger
	configPath   string
	router       *gin.Engine
	server       *http.Server
	signalingHub *signaling.Hub
}

func New(cfg config.AppConfig, configPath string, logger *log.Logger) (*App, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	if cfg.MountPath != "" {
		if err := os.MkdirAll(cfg.MountPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create mount path: %w", err)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// App-level middlewares: Logger, Recovery, TailscaleAuth
	router.Use(middleware.Logger(logger), gin.Recovery())
	router.Use(middleware.TailscaleAuth(logger, cfg))

	sigHub := signaling.NewHub()
	setupRoutes(router, cfg, sigHub)

	app := &App{
		logger:       logger,
		configPath:   configPath,
		router:       router,
		signalingHub: sigHub,
	}

	app.server = newServer(cfg.ListenAddr, router)

	return app, nil
}

func setupRoutes(r *gin.Engine, cfg config.AppConfig, sigHub *signaling.Hub) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.File("internal/app/static/hello.html")
	})

	r.GET("/api/v1/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from Gin Harness!",
		})
	})

	r.GET("/api/v1/monitor", func(c *gin.Context) {
		status, err := monitor.GetSystemStatus(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get system status"})
			return
		}
		c.JSON(http.StatusOK, status)
	})

	r.POST("/api/v1/files/upload", files.UploadHandler(cfg.MountPath))
	r.GET("/api/v1/files/download", files.DownloadHandler(cfg.MountPath))
	r.GET("/api/v1/files/list", files.ListHandler(cfg.MountPath))

	r.GET("/api/v1/terminal/ws", terminal.Handler)
	r.GET("/api/v1/signaling/ws", signaling.Handler(sigHub))
	r.GET("/api/v1/tailscale/devices", tailscale.GetDevicesHandler)
}
