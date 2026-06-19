package tailscale

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"tailscale.com/client/tailscale"
)

type Device struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
	OS   string   `json:"os"`
}

func GetDevicesHandler(c *gin.Context) {
	var lc tailscale.LocalClient
	status, err := lc.Status(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tailscale status"})
		return
	}

	var devices []Device

	for _, peer := range status.Peer {
		var ips []string
		for _, ip := range peer.TailscaleIPs {
			ips = append(ips, ip.String())
		}
		
		name := peer.HostName
		if peer.DNSName != "" {
			name = peer.DNSName
		}

		devices = append(devices, Device{
			Name: name,
			IPs:  ips,
			OS:   peer.OS,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}
