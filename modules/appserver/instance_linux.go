//go:build linux

package appserver

import (
	"os/exec"

	"golang.org/x/sys/unix"
)

// Configure cmd to kill all children on Linux
func configCmd(cmd *exec.Cmd) *exec.Cmd {
	cmd.SysProcAttr = &unix.SysProcAttr{Setsid: true}
	return cmd
}

// Kill process and children on Linux
func killCmd(cmd *exec.Cmd) error {
	return unix.Kill(-cmd.Process.Pid, unix.SIGKILL)
}
