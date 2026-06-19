package webrtc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var hostController *Controller

func StartBackgroundListener(ip, port string) {
	if ip == "" {
		return
	}
	url := fmt.Sprintf("ws://%s:%s/api/v1/signaling/ws", ip, port)
	go autoListenLoop(url)
}

func autoListenLoop(url string) {
	for {
		ws, err := dialWebsocket(url)
		if err == nil {
			listenWs(ws)
		}
		time.Sleep(5 * time.Second)
	}
}

func listenWs(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		handleAutoSig(ws, msg)
	}
}

func handleAutoSig(ws *websocket.Conn, msg []byte) {
	var s SigMsg
	if err := json.Unmarshal(msg, &s); err != nil {
		return
	}
	if s.Type == "offer" {
		acceptOffer(ws, &s)
	} else if s.Type == "candidate" && hostController != nil {
		handleCandidate(hostController, s.Payload)
	}
}

func acceptOffer(ws *websocket.Conn, s *SigMsg) {
	c := &Controller{ws: ws, target: s.Sender, isHost: true}
	c.pc, _ = webrtc.NewPeerConnection(webrtc.Configuration{})
	setupIceHandler(c)
	c.pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() { go sendScreenFrames(dc) })
	})
	hostController = c
	handleOffer(c, s.Payload)
}
