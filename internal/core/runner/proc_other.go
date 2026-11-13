//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

func shellCommand(command string) *exec.Cmd {
	return exec.Command("sh", "-c", command)
}

func prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killTree(cmd *exec.Cmd, pid int) {
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		_ = cmd.Process.Kill()
	}
}
