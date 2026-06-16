package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"reverseproxy-poc/internal/config"
	"reverseproxy-poc/internal/middleware"
	"reverseproxy-poc/internal/monitor"
)

type App struct {
	logger     *log.Logger
	configPath string
	router     *gin.Engine
	server     *http.Server
}

func New(cfg config.AppConfig, configPath string, logger *log.Logger) (*App, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// App-level middlewares: Logger, Recovery, TailscaleAuth
	router.Use(middleware.Logger(logger), gin.Recovery())
	router.Use(middleware.TailscaleAuth(logger, cfg))

	setupRoutes(router)

	app := &App{
		logger:     logger,
		configPath: configPath,
		router:     router,
	}

	app.server = newServer(cfg.ListenAddr, router)

	return app, nil
}

func setupRoutes(r *gin.Engine) {
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
}
