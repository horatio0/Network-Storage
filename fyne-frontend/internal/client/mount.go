package client

import (
	"fmt"
	"os"
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

	if err := os.MkdirAll(local, 0755); err != nil {
		// Ignore error; if it fails, mount will fail too, or the dir might already exist.
	}

	remote := fmt.Sprintf("%s:%s", ip, share)
	cmd := exec.Command("sudo", "-S", "mount", "-t", "nfs", remote, local)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stdin = strings.NewReader(pwd + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func UnmountDrive(local string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("net", "use", local, "/delete", "/y")
		err := cmd.Run()
		if err == nil {
			os.Remove(local)
		}
		return err
	}

	pwd := GetSudoPassword()
	if pwd == "" {
		return ErrPasswordRequired
	}

	cmd := exec.Command("sudo", "-S", "umount", "-l", local)
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	cmd.Stdin = strings.NewReader(pwd + "\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		os.Remove(local)
	} else {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return err
}
