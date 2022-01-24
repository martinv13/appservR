//go:build windows

package appserver

import (
	"os/exec"
	"strconv"
)

// noop on Windows
func configCmd(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}

// Kill process and children on Windows
func killCmd(cmd *exec.Cmd) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	return kill.Run()
}
