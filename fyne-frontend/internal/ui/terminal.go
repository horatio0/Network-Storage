package ui

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
	"github.com/gorilla/websocket"
)

type wsWrapper struct {
	conn *websocket.Conn
	r    io.Reader
}

func (w *wsWrapper) Read(p []byte) (int, error) {
	if w.r == nil {
		_, r, err := w.conn.NextReader()
		if err != nil {
			return 0, err
		}
		w.r = r
	}
	n, err := w.r.Read(p)
	if err == io.EOF {
		w.r = nil
		return n, nil
	}
	return n, err
}

func (w *wsWrapper) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

func (w *wsWrapper) Close() error {
	return w.conn.Close()
}

func createTerminalView(a fyne.App, w fyne.Window) fyne.CanvasObject {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return container.NewCenter(widget.NewLabel("No IP"))
	}

	url := fmt.Sprintf("ws://%s:%s/api/v1/terminal/ws", ip, port)
	return connectTerminal(url)
}

func connectTerminal(url string) fyne.CanvasObject {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return container.NewCenter(widget.NewLabel(err.Error()))
	}

	term := terminal.New()
	ws := &wsWrapper{conn: conn}
	go func() {
		_ = term.RunWithConnection(ws, ws)
	}()
	return container.NewMax(term)
}
