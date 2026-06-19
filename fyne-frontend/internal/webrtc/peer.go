package webrtc

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type SigMsg struct {
	Type    string `json:"type"`
	Sender  string `json:"sender,omitempty"`
	Target  string `json:"target"`
	Payload string `json:"payload"`
}

type Controller struct {
	pc      *webrtc.PeerConnection
	ws      *websocket.Conn
	target  string
	isHost  bool
	onFrame func([]byte)
	onError func(error)
}

func NewController(url, tgt string, host bool, cb func([]byte), errCb func(error)) (*Controller, error) {
	ws, err := dialWebsocket(url)
	if err != nil {
		return nil, err
	}
	c := &Controller{ws: ws, target: tgt, isHost: host, onFrame: cb, onError: errCb}
	return initController(c)
}

func initController(c *Controller) (*Controller, error) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		c.ws.Close()
		return nil, err
	}
	c.pc = pc
	setupStateChange(c)
	setupIceHandler(c)
	setupDataChannel(c)
	return c, nil
}

func setupStateChange(c *Controller) {
	c.pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateFailed && c.onError != nil {
			c.onError(fmt.Errorf("state: %s", s))
		}
	})
}

func dialWebsocket(url string) (*websocket.Conn, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	return ws, err
}

func setupDataChannel(c *Controller) {
	if c.isHost {
		setupDataChannelHost(c)
	} else {
		setupDataChannelViewer(c)
	}
}

func setupIceHandler(c *Controller) {
	c.pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		b, _ := json.Marshal(i.ToJSON())
		sendSig(c.ws, "candidate", c.target, string(b))
	})
}

func sendSig(ws *websocket.Conn, t, target, payload string) {
	b, _ := json.Marshal(SigMsg{Type: t, Target: target, Payload: payload})
	ws.WriteMessage(websocket.TextMessage, b)
}

func (c *Controller) StartSignaling() {
	go func() {
		defer c.ws.Close()
		for {
			_, msg, err := c.ws.ReadMessage()
			if err != nil {
				break
			}
			handleSigMsg(c, msg)
		}
	}()
	if c.isHost {
		createOffer(c)
	}
}

func createOffer(c *Controller) {
	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		return
	}
	c.pc.SetLocalDescription(offer)
	b, _ := json.Marshal(offer)
	sendSig(c.ws, "offer", c.target, string(b))
}

func handleSigMsg(c *Controller, msg []byte) {
	var s SigMsg
	if err := json.Unmarshal(msg, &s); err != nil {
		return
	}
	if s.Type == "offer" {
		handleOffer(c, s.Payload)
	}
	if s.Type == "answer" {
		handleAnswer(c, s.Payload)
	}
	if s.Type == "candidate" {
		handleCandidate(c, s.Payload)
	}
}

func handleOffer(c *Controller, p string) {
	var sdp webrtc.SessionDescription
	json.Unmarshal([]byte(p), &sdp)
	c.pc.SetRemoteDescription(sdp)
	ans, _ := c.pc.CreateAnswer(nil)
	c.pc.SetLocalDescription(ans)
	b, _ := json.Marshal(ans)
	sendSig(c.ws, "answer", c.target, string(b))
}

func handleAnswer(c *Controller, p string) {
	var sdp webrtc.SessionDescription
	json.Unmarshal([]byte(p), &sdp)
	c.pc.SetRemoteDescription(sdp)
}

func handleCandidate(c *Controller, p string) {
	var ice webrtc.ICECandidateInit
	json.Unmarshal([]byte(p), &ice)
	c.pc.AddICECandidate(ice)
}
