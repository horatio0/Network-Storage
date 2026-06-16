package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"reverseproxy-poc/internal/config"
	"tailscale.com/client/tailscale"
)

func TailscaleAuth(logger *log.Logger, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Tailscale.Enabled {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		whois, err := tailscale.WhoIs(c.Request.Context(), clientIP)
		if err != nil {
			logger.Printf("Tailscale Auth failed for %s: %v", clientIP, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Not on Tailscale"})
			return
		}

		if len(cfg.Tailscale.AllowedUsers) > 0 {
			allowed := false
			for _, u := range cfg.Tailscale.AllowedUsers {
				if whois.UserProfile.LoginName == u {
					allowed = true
					break
				}
			}
			if !allowed {
				logger.Printf("Tailscale Auth forbidden for user %s (%s)", whois.UserProfile.LoginName, clientIP)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: User not allowed"})
				return
			}
		}

		c.Set("ts_user", whois.UserProfile.LoginName)
		c.Set("ts_node", whois.Node.Name)
		c.Next()
	}
}
