package goproc

import (
	"os"
	"os/exec"
)

var stopSignal = os.Interrupt

// setService Group PidとSession idを親プロセスから分離する Windowsのやり方が分かるまで空にしておく
func setService(cmd *exec.Cmd) {
	return
}
