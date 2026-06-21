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
				AddLog(a, "Terminal error: "+err.Error())
				
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
		go func() {
			_ = term.RunWithConnection(ws, ws)
		}()
		go func() {
			time.Sleep(200 * time.Millisecond)
			ws.Write([]byte("\r\n"))
		}()
		bg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		fyne.Do(func() {
			stack.Objects = []fyne.CanvasObject{container.NewMax(bg, term)}
			stack.Refresh()
		})
	}()
}
