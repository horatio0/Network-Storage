package middleware

import (
	"log"
	"net/http"

	"central-control-backend/internal/config"
	"github.com/gin-gonic/gin"
	"tailscale.com/client/tailscale"
)

func TailscaleAuth(logger *log.Logger, cfg config.AppConfig, tsClient *tailscale.LocalClient) gin.HandlerFunc {
	return func(c *gin.Context) {

		remoteAddr := c.Request.RemoteAddr
		whois, err := tsClient.WhoIs(c.Request.Context(), remoteAddr)
		if err != nil {
			logger.Printf("Tailscale Auth failed for %s: %v", remoteAddr, err)
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
				logger.Printf("Tailscale Auth forbidden for user %s (%s)", whois.UserProfile.LoginName, remoteAddr)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: User not allowed"})
				return
			}
		}

		c.Set("ts_user", whois.UserProfile.LoginName)
		c.Set("ts_node", whois.Node.Name)
		c.Next()
	}
}
