package goproc

import (
	"math"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var stopSignal = os.Interrupt

// setService Group PidとSession idを親プロセスから分離する Windowsのやり方が分かるまで空にしておく
func setService(cmd *exec.Cmd) {
	return
}

func getCPUPercent(p *process.Process) (float64, error) {
	// CPUPercent()はタスクマネージャーやtopと違う。同じような値はPercent()で取れる(https://github.com/shirou/gopsutil/issues/1006)
	// Winの標準はたぶん1秒更新なので合わせる
	cpupercent, err := p.Percent(1 * time.Second)
	if err != nil {
		return 0, err
	} else {
		// Winのタスクマネージャーは全コア合計のCPU使用率が出るのでコア数で割る
		// Winのタスクマネージャーは小数点以下切り捨てだが、小数点以下一位を四捨五入で出す
		return math.Round(cpupercent/float64(runtime.NumCPU())*10) / 10, nil
	}
}

// GetEnviron MacでEnviron()が動かないので独自実装。Winでは単なるWrapper
func GetEnviron(p *process.Process) ([]string, error) {
	envs, err := p.Environ()
	if err != nil {
		return nil, err
	}
	return envs, nil
}
