package client

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func MountDrive(ip, share, local string) error {
	if runtime.GOOS == "windows" {
		remote := fmt.Sprintf(`\\%s\%s`, ip, share)
		// Windows에서는 관리자 권한 없이 사용자 세션에 마운트(네트워크 드라이브)
		cmd := exec.Command("net", "use", local, remote)
		return cmd.Run()
	}

	pwd := GetSudoPassword()
	if pwd == "" {
		return ErrPasswordRequired
	}

	remote := fmt.Sprintf("%s:/%s", ip, share)
	cmd := exec.Command("sudo", "-S", "mount", "-t", "nfs", remote, local)
	cmd.Stdin = strings.NewReader(pwd + "\n")
	return cmd.Run()
}

func UnmountDrive(local string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("net", "use", local, "/delete")
		return cmd.Run()
	}

	pwd := GetSudoPassword()
	if pwd == "" {
		return ErrPasswordRequired
	}

	cmd := exec.Command("sudo", "-S", "umount", local)
	cmd.Stdin = strings.NewReader(pwd + "\n")
	return cmd.Run()
}
