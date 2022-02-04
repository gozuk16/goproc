package goproc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/inhies/go-bytesize"
	"github.com/shirou/gopsutil/v3/process"
)

// overwritten with os.Interrupt on windows environment (see main_windows.go)
const stopSignal = syscall.SIGTERM

const timeformat = "2006/01/02 15:04:05"

// 子プロセス情報
type ChildrenProcess struct {
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	Pid     int    `json:"pid"`
	Vms     string `json:"vms"`
	Rss     string `json:"rss"`
	Swap    string `json:"swap"`
}

// プロセス情報
type Process struct {
	Name       string            `json:"name"`
	CpuPercent float64           `json:"cpuPercent"`
	CpuTotal   float64           `json:"cpuTotal"`
	CpuUser    float64           `json:"cpuUser"`
	CpuSystem  float64           `json:"cpuSystem"`
	CpuIdle    float64           `json:"cpuIdle"`
	CpuIowait  float64           `json:"cpuIowait"`
	Vms        string            `json:"vms"`
	Rss        string            `json:"rss"`
	Swap       string            `json:"swap"`
	Cmdline    string            `json:"cmdline"`
	Exe        string            `json:"exe"`
	Cwd        string            `json:"cwd"`
	CreateTime string            `json:"createTime"`
	Exist      bool              `json:"exist"`
	Status     string            `json:"status"`
	Pid        int               `json:"pid"`
	Ppid       int               `json:"ppid"`
	Children   []ChildrenProcess `json:"children"`
}

type Processes []Process

// プロセス起動・停止に必要な情報
type ProcessParam struct {
	Env        []string `json:"env"`
	CurrentDir string   `json:"currentDir"`
	Command    string   `json:"command"`
	Args       string   `json:"args"`
}

var ErrInterrupt = errors.New("interrupt signal accepted.")

// GetProcesses 指定されたPIDのプロセス情報をまとめて返す
func GetProcesses(pids []int) (Processes, error) {
	ret := []Process{}
	for _, pid := range pids {
		p, err := GetProcess(pid)
		if err != nil {
			// errorならスキップする(全部エラーなら0個返す)
			continue
		}
		ret = append(ret, *p)
	}

	return ret, nil
}

// GetProcess 指定されたPIDのプロセス情報を返す
func GetProcess(pid int) (*Process, error) {
	ret := &Process{}

	// 渡されたpidがマイナス、0、1の時はエラーで返す(そうじゃないとPanicになる)
	if pid <= 1 {
		return nil, fmt.Errorf("Don't get process, when pid is %d", pid)
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}

	ret.Name, err = p.Name()
	if err != nil {
		log.Printf("error: get process.Name: %v", err)
	}

	cpupercent, err := p.CPUPercent()
	if err != nil {
		log.Printf("error: get process.CPUPercent: %v", err)
		ret.CpuPercent = 0
	} else {
		ret.CpuPercent = math.Round(cpupercent*10) / 10
	}

	cputime, err := p.Times()
	if err != nil {
		log.Printf("error: get process.Time: %v", err)
		ret.CpuTotal = 0
		ret.CpuUser = 0
		ret.CpuSystem = 0
		ret.CpuIdle = 0
		ret.CpuIowait = 0
	} else {
		ret.CpuTotal = math.Round(cputime.Total()*100) / 100
		ret.CpuUser = cputime.User
		ret.CpuSystem = cputime.System
		ret.CpuIdle = cputime.Idle
		ret.CpuIowait = cputime.Iowait
	}

	memory, err := p.MemoryInfo()
	if err != nil {
		log.Printf("error: get process.MemoryInfo: %v", err)
		ret.Vms = err.Error()
		ret.Rss = err.Error()
		ret.Swap = err.Error()
	} else {
		ret.Vms = bytesize.New(float64(memory.VMS)).String()
		ret.Rss = bytesize.New(float64(memory.RSS)).String()
		ret.Swap = bytesize.New(float64(memory.Swap)).String()
	}

	ret.Cmdline, err = p.Cmdline()
	if err != nil {
		log.Printf("error: get process.Cmdline: %v", err)
	}
	ret.Exe, err = p.Exe()
	if err != nil {
		//log.Printf("error: get process.Exe: %v", err)
		ret.Exe = err.Error()
	}
	ret.Cwd, err = p.Cwd()
	if err != nil {
		//log.Printf("error: get process.Cwd: %v", err)
		// TODO:Macだとnot implemented yetとなる。Winの挙動を調べて分岐するかエラー処理ではじくか？
		ret.Cwd = err.Error()
	}

	createtime, err := p.CreateTime()
	if err != nil {
		log.Printf("error: get process.CreateTime: %v", err)
	}
	ret.CreateTime = time.Unix(createtime/1000, 0).Format(timeformat)

	// TODO:これで死活をチェックして以後の処理をスキップした方がいい？
	ret.Exist, err = process.PidExists(int32(pid))
	if err != nil {
		log.Printf("error: get process.PidExists: %v", err)
	}

	statuses, err := p.Status()
	if err != nil {
		log.Printf("error: get process.Status: %v", err)
	}
	ret.Status = strings.Join(statuses, ", ")

	ret.Pid = int(p.Pid)

	ppid, err := p.Ppid()
	if err != nil {
		log.Printf("error: get process.Ppid: %v", err)
	}
	ret.Ppid = int(ppid)

	cp := []ChildrenProcess{}
	children, err := p.Children()
	if err != nil {
		return ret, nil
	}
	for _, c := range children {
		cname, err := c.Name()
		if err != nil {
			log.Printf("error: get process.Children.Name: %v", err)
			cname = err.Error()
		}
		ccmd, err := c.Cmdline()
		if err != nil {
			log.Printf("error: get process.Children.Cmdline: %v", err)
			ccmd = err.Error()
		}
		cmemory, err := c.MemoryInfo()
		var cvms, crss, cswap string
		if err != nil {
			log.Printf("error: get process.Children.MemoryInfo: %v", err)
			cvms = err.Error()
			crss = err.Error()
			cswap = err.Error()
		} else {
			cvms = bytesize.New(float64(cmemory.VMS)).String()
			crss = bytesize.New(float64(cmemory.RSS)).String()
			cswap = bytesize.New(float64(cmemory.Swap)).String()
		}
		cp = append(cp, ChildrenProcess{cname, ccmd, int(c.Pid), cvms, crss, cswap})
	}
	ret.Children = cp

	return ret, nil
}

// StartService 非同期サービスを起動し、PIDを知らせる
func StartService(param ProcessParam) (int, error) {
	startArgs := strings.Fields(param.Args)

	// 先に環境変数を展開して反映しておかないと修正したPATHがexec.Commandに適用されない
	env := []string{}
	if len(param.Env) > 0 {
		env = setExpandEnv(param.Env)
	}

	cmd := exec.Command(param.Command, startArgs...)
	cmd.Dir = param.CurrentDir
	if len(env) > 0 {
		cmd.Env = env
	}

	setService(cmd)
	err := cmd.Start()
	if err != nil {
		return -1, err
	} else {
		pid := cmd.Process.Pid
		return pid, nil
	}
}

// RunProcess プロセス起動して終了まで待つ
func RunProcess(param ProcessParam) error {
	// Ctrl+Cを受け取る
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	done := make(chan error, 1)
	go newProcess(done, param)

	select {
	case <-quit:
		return ErrInterrupt
	case err := <-done:
		if err != nil {
			return err
		}
	}
	return nil
}

// newProcess goroutineでプロセスを起動
func newProcess(done chan<- error, param ProcessParam) {
	defer close(done)

	startArgs := strings.Fields(param.Args)

	// 先に環境変数を展開して反映しておかないと修正したPATHがexec.Commandに適用されない
	env := []string{}
	if len(param.Env) > 0 {
		env = setExpandEnv(param.Env)
	}

	cmd := exec.Command(param.Command, startArgs...)
	cmd.Dir = param.CurrentDir
	if len(param.Env) > 0 {
		cmd.Env = env
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	stdoutStderr := io.MultiReader(stdout, stderr)

	setService(cmd)
	err := cmd.Start()
	if err != nil {
		done <- err
	} else {
		scanner := bufio.NewScanner(stdoutStderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}

	done <- nil
}

// StopService サービス停止コマンドを起動し、サービスが終了するまで待つ
// timeoutあり
func StopService(param ProcessParam) error {
	if err := RunProcess(param); err != nil {
		return err
	}

	return nil

}

// StopServiceByPid PIDでプロセスを識別してシグナルを送信して終了する
/*
func StopServiceByPid(pid int) error {
	if err := stopProcessByPid(pid); err != nil {
		return err
	}

	return nil
}
*/

// StopServiceByPid PIDでプロセスを識別してシグナルを送信して終了する
//func stopProcessByPid(pid int) error {
func StopServiceByPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	//stopSignal := syscall.SIGTERM
	err = p.Signal(stopSignal)
	if err != nil {
		return err
	}

	return nil
}

// setExpandEnv 渡された環境変数に変数があれば固定値に展開して返す
func setExpandEnv(orgEnv []string) []string {
	var env []string

	for _, e := range orgEnv {
		expandEnv := os.ExpandEnv(e)
		env := strings.Split(expandEnv, "=")
		if err := os.Setenv(env[0], env[1]); err != nil {
			log.Println(err)
		}
		log.Printf("env: %v to expandEnv: %v\n", e, expandEnv)
		env = append(env, expandEnv)
	}
	return env
}
