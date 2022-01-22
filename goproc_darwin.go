package goproc

import (
	"os/exec"
	"syscall"
)

// setService Group PidとSession idを親プロセスから分離する
func setService(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
