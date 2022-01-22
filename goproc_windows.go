package goproc

import (
	"os/exec"
)

// setService Group PidとSession idを親プロセスから分離する Windowsのやり方が分かるまで空にしておく
func setService(cmd *exec.Cmd) {
	return
}
