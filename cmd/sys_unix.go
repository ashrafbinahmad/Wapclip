//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func setDaemonAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

func attachConsole() {}
