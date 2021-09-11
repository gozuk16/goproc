package goproc

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/inhies/go-bytesize"
)

const timeformat = "2006/01/02 15:04:05"

type ChildrenProcess struct {
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	Pid     int    `json:"pid"`
	Vms     string `json:"vms"`
	Rss     string `json:"rss"`
	Swap    string `json:"swap"`
}

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

// GetProcesses 指定されたPIDのプロセス情報をまとめて返す
func GetProcesses(pids []int) (Processes, error) {
	ret := []Process{}
	for _, pid := range pids {
		p, err := GetProcess(pid)
		if err != nil {
			// TODO 渡されたpids全部がerrorの時だけエラーにする
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
