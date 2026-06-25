package ui

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"io"
	"sync"
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
	mu   sync.Mutex
}

func (w *wsWrapper) Read(p []byte) (int, error) {
	if w.r == nil {
		_, r, err := w.conn.NextReader()
		if err != nil {
			// 웹소켓 연결이 끊어지거나 실패한 경우 io.EOF를 반환하여 루프 탈출
			return 0, io.EOF
		}
		w.r = r
	}
	n, err := w.r.Read(p)
	if err == io.EOF {
		w.r = nil
		return n, nil
	}
	if err != nil {
		w.r = nil
		return n, io.EOF
	}
	return n, nil
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

func (w *wsWrapper) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	msg := TerminalMessage{
		Type: "input",
		Data: base64.StdEncoding.EncodeToString(p),
	}
	err := w.conn.WriteJSON(msg)
	return len(p), err
}

func (w *wsWrapper) Close() error {
	return w.conn.Close()
}

type termLayout struct{}

func (l *termLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (l *termLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 100) // Allow shrinking smaller than default minimum size
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
	stack := container.NewStack()
	connectTerminal(a, url, stack)
	cachedTerminal = stack
	return cachedTerminal
}

func connectTerminal(a fyne.App, url string, stack *fyne.Container) {
	go func() {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			fyne.Do(func() {
				AddErrorLog(a, "Terminal error: "+err.Error(), "WS "+url, err.Error(), 0)
				
				errMsg := widget.NewLabel("Connection Failed: " + err.Error())
				retryBtn := widget.NewButton("Retry", func() {
					stack.Objects = []fyne.CanvasObject{container.NewCenter(widget.NewLabel("Retrying..."))}
					stack.Refresh()
					go func() {
						time.Sleep(100 * time.Millisecond)
						connectTerminal(a, url, stack)
					}()
				})
				vbox := container.NewVBox(errMsg, retryBtn)
				stack.Objects = []fyne.CanvasObject{container.NewCenter(vbox)}
				stack.Refresh()
			})
			return
		}

		term := terminal.New()
		ws := &wsWrapper{conn: conn}

		configChan := make(chan terminal.Config)
		term.AddListener(configChan)

		go func() {
			for cfg := range configChan {
				ws.mu.Lock()
				msg := TerminalMessage{
					Type: "resize",
					Cols: uint16(cfg.Columns),
					Rows: uint16(cfg.Rows),
				}
				ws.conn.WriteJSON(msg)
				ws.mu.Unlock()
			}
		}()

		go func() {
			_ = term.RunWithConnection(ws, ws)
		}()
		go func() {
			time.Sleep(200 * time.Millisecond)
			ws.Write([]byte("\r\n"))
		}()
		bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		fyne.Do(func() {
			stack.Objects = []fyne.CanvasObject{container.New(&termLayout{}, bg, term)}
			stack.Refresh()
		})
	}()
}
