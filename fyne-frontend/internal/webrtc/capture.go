package webrtc

import (
	"bytes"
	"image/jpeg"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/pion/webrtc/v3"
)

func setupDataChannelHost(c *Controller) {
	dc, err := c.pc.CreateDataChannel("video", nil)
	if err != nil {
		return
	}
	dc.OnOpen(func() {
		go sendScreenFrames(dc)
	})
}

func sendScreenFrames(dc *webrtc.DataChannel) {
	ticker := time.NewTicker(time.Second / 5) // 5 FPS
	defer ticker.Stop()
	for range ticker.C {
		if dc.ReadyState() != webrtc.DataChannelStateOpen {
			break
		}
		bounds := screenshot.GetDisplayBounds(0)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
		dc.Send(buf.Bytes())
	}
}

func setupDataChannelViewer(c *Controller) {
	c.pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if c.onFrame != nil {
				c.onFrame(msg.Data)
			}
		})
	})
}
