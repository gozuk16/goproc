package goproc

import (
	"syscall"
	"os/exec"
)

// overwritten with os.Interrupt on windows environment (see goproc_windows.go)
var stopSignal = syscall.SIGTERM

// setService Group PidとSession idを親プロセスから分離する
func setService(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
