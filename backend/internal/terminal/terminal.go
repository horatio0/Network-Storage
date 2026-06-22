package terminal

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/coder/websocket"
	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
)

// Handler handles websocket requests for terminal access.
func Handler(c *gin.Context) {
	tsNodeRaw, exists := c.Get("ts_node")
	if !exists {
		log.Println("Terminal failed: ts_node not found in context")
		return
	}
	_ = tsNodeRaw.(string)

	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to accept websocket: %v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "closing")

	ctx := c.Request.Context()
	shell := getShell()
	cmd := exec.CommandContext(ctx, shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.Start(cmd)
	if err != nil {
		log.Printf("failed to start pty: %v", err)
		return
	}
	defer func() {
		_ = f.Close()
		_ = cmd.Process.Kill()
	}()

	wsConn := websocket.NetConn(ctx, conn, websocket.MessageBinary)
	go io.Copy(f, wsConn)
	_, _ = io.Copy(wsConn, f)
}

func getShell() string {
	if _, err := os.Stat("/bin/bash"); err == nil {
		return "/bin/bash"
	}
	return "/bin/sh"
}
