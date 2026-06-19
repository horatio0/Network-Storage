package client

import (
	"fmt"
	"os/exec"
	"runtime"
)

func MountDrive(ip, share, local string) error {
	if runtime.GOOS == "windows" {
		remote := fmt.Sprintf(`\\%s\%s`, ip, share)
		cmd := exec.Command("cmd", "/c", "mklink", "/D", local, remote)
		return cmd.Run()
	}
	remote := fmt.Sprintf("//%s/%s", ip, share)
	cmd := exec.Command("sudo", "mount", "-t", "cifs", "-o", "guest", remote, local)
	return cmd.Run()
}

func UnmountDrive(local string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "rmdir", local)
		return cmd.Run()
	}
	cmd := exec.Command("sudo", "umount", local)
	return cmd.Run()
}
