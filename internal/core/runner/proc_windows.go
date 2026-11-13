//go:build windows

package runner

import (
	"os/exec"
	"strconv"
	"syscall"
)

func shellCommand(command string) *exec.Cmd {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `/S /C "` + command + `"`}
	return cmd
}

func prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr.CreationFlags |= syscall.CREATE_NEW_PROCESS_GROUP
}

func killTree(cmd *exec.Cmd, pid int) {
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run(); err != nil {
		_ = cmd.Process.Kill()
	}
}
