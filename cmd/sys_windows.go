//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func setDaemonAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}

func attachConsole() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole := modkernel32.NewProc("AttachConsole")
	const ATTACH_PARENT_PROCESS = uint32(0xFFFFFFFF)
	procAttachConsole.Call(uintptr(ATTACH_PARENT_PROCESS))

	hout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if hout != syscall.InvalidHandle {
		os.Stdout = os.NewFile(uintptr(hout), "stdout")
	}
	herr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	if herr != syscall.InvalidHandle {
		os.Stderr = os.NewFile(uintptr(herr), "stderr")
	}
}
