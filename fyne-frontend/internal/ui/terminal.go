package ui

import (
	"fmt"
	"image/color"
	"io"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

var (
	cachedTerminal fyne.CanvasObject
	cachedTerminalURL string
)

func createTerminalView(a fyne.App, w fyne.Window) fyne.CanvasObject {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip == "" {
		return container.NewCenter(widget.NewLabel("No IP"))
	}

	url := fmt.Sprintf("ws://%s:%s/api/v1/terminal/ws", ip, port)
	
	if cachedTerminal != nil && cachedTerminalURL == url {
		return cachedTerminal
	}

	cachedTerminalURL = url
	cachedTerminal = connectTerminal(url)
	return cachedTerminal
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
	go func() {
		time.Sleep(200 * time.Millisecond)
		ws.Write([]byte("\r\n"))
	}()
	bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	return container.NewMax(bg, term)
}
