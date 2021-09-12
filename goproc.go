package goproc

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/inhies/go-bytesize"
)

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
	Env        string `json:"env"`
	CurrentDir string `json:"currentDir"`
	StartCmd   string `json:"startCmd"`
	StartArgs  string `json:"startArgs"`
	StopCmd    string `json:"stopCmd"`
	StopArgs   string `json:"stopArgs"`
}

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

	ret.Name, _ = p.Name()

	cpupercent, _ := p.CPUPercent()
	ret.CpuPercent = math.Round(cpupercent*1000) / 10

	cputime, _ := p.Times()
	ret.CpuTotal = math.Round(cputime.Total()*100) / 100
	ret.CpuUser = cputime.User
	ret.CpuSystem = cputime.System
	ret.CpuIdle = cputime.Idle
	ret.CpuIowait = cputime.Iowait

	memory, _ := p.MemoryInfo()
	ret.Vms = bytesize.New(float64(memory.VMS)).String()
	ret.Rss = bytesize.New(float64(memory.RSS)).String()
	ret.Swap = bytesize.New(float64(memory.Swap)).String()

	ret.Cmdline, _ = p.Cmdline()
	ret.Exe, _ = p.Exe()
	ret.Cwd, _ = p.Cwd()

	createtime, _ := p.CreateTime()
	ret.CreateTime = time.Unix(createtime/1000, 0).Format(timeformat)

	ret.Exist, _ = process.PidExists(int32(pid))

	statuses, _ := p.Status()
	ret.Status = strings.Join(statuses, ", ")

	ret.Pid = int(p.Pid)

	ppid, _ := p.Ppid()
	ret.Ppid = int(ppid)

	cp := []ChildrenProcess{}
	children, _ := p.Children()
	for _, c := range children {
		cname, _ := c.Name()
		ccmd, _ := c.Cmdline()
		cmemory, _ := c.MemoryInfo()
		cvms := bytesize.New(float64(cmemory.VMS)).String()
		crss := bytesize.New(float64(cmemory.RSS)).String()
		cswap := bytesize.New(float64(cmemory.Swap)).String()
		cp = append(cp, ChildrenProcess{cname, ccmd, int(c.Pid), cvms, crss, cswap})
	}
	ret.Children = cp

	return ret, nil
}

// StartProcess プロセス起動
func StartProcess(param ProcessParam) error {
	// Ctrl+Cを受け取る
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	done := make(chan error, 1)
	go newProcess(done, param)

	select {
	case <-quit:
		fmt.Println("interrup signal accepted.")
	case err := <-done:
		if err != nil {
			fmt.Println("exit.", err)
			return err
		}
		fmt.Println("exit.")
	}
	return nil
}

// newProcess goroutineでプロセスを1つづつ起動
func newProcess(done chan<- error, param ProcessParam) {

	// process start
	startArgs := strings.Fields(param.StartArgs)
	cmd := exec.Command(param.StartCmd, startArgs...)
	cmd.Dir = param.CurrentDir
	// cmd.Env = startEnv
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		//log.Fatal(err)
		done <- err
	}

	fmt.Println("--- stderr ---")
	scanner2 := bufio.NewScanner(stderr)
	for scanner2.Scan() {
		fmt.Println(scanner2.Text())
	}

	fmt.Println("--- stdout ---")
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	done <- nil
	close(done)
}
