package terminal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"os/exec"

	"github.com/coder/websocket"
	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
)

// Handler handles websocket requests for terminal access.
func Handler(c *gin.Context) {
	if _, exists := c.Get("ts_node"); !exists {
		log.Println("Terminal failed: ts_node not found in context")
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to accept websocket: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "terminal closed")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	f, cmd, err := startPty(ctx)
	if err != nil {
		log.Printf("failed to start pty: %v", err)
		return
	}
	defer func() {
		_ = f.Close()
		_ = cmd.Process.Kill()
	}()

	go pumpWebsocketToPty(ctx, conn, f)
	pumpPtyToWebsocket(ctx, conn, f)
}

func startPty(ctx context.Context) (*os.File, *exec.Cmd, error) {
	shell := getShell()
	cmd := exec.CommandContext(ctx, shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	f, err := pty.Start(cmd)
	return f, cmd, err
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

func pumpWebsocketToPty(ctx context.Context, conn *websocket.Conn, f *os.File) {
	defer f.Close()
	for {
		_, b, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(b, &msg); err != nil {
			log.Printf("Failed to unmarshal terminal message: %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			decoded, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				log.Printf("Failed to decode base64 input data: %v", err)
				continue
			}
			f.Write(decoded)
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				pty.Setsize(f, &pty.Winsize{
					Rows: msg.Rows,
					Cols: msg.Cols,
				})
			}
		default:
			log.Printf("Unknown terminal message type: %s", msg.Type)
		}
	}
}

func pumpPtyToWebsocket(ctx context.Context, conn *websocket.Conn, f *os.File) {
	buf := make([]byte, 8192)
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			if err := conn.Write(ctx, websocket.MessageBinary, buf[:n]); err != nil {
				break
			}
		}
	}
}

func getShell() string {
	if _, err := os.Stat("/bin/bash"); err == nil {
		return "/bin/bash"
	}
	return "/bin/sh"
}
