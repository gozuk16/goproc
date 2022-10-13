package goproc

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"

	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/inhies/go-bytesize"
	"github.com/mattn/go-shellwords"
	"github.com/shirou/gopsutil/v3/process"
)

// 子プロセス情報
type ChildrenProcess struct {
	Name       string  `json:"name"`
	Cmdline    string  `json:"cmdline"`
	Pid        int     `json:"pid"`
	CpuPercent float64 `json:"cpuPercent"`
	Vms        string  `json:"vms"`
	Rss        string  `json:"rss"`
	Swap       string  `json:"swap"`
}

// プロセス情報
type Process struct {
	Name          string            `json:"name"`
	CpuPercent    float64           `json:"cpuPercent"`
	CpuTotal      float64           `json:"cpuTotal"`
	CpuUser       float64           `json:"cpuUser"`
	CpuSystem     float64           `json:"cpuSystem"`
	CpuIdle       float64           `json:"cpuIdle"`
	CpuIowait     float64           `json:"cpuIowait"`
	Vms           string            `json:"vms"`
	Rss           string            `json:"rss"`
	Swap          string            `json:"swap"`
	Cmdline       string            `json:"cmdline"`
	Exe           string            `json:"exe"`
	Cwd           string            `json:"cwd"`
	CreateTime    string            `json:"createTime"`
	Exist         bool              `json:"exist"`
	Status        string            `json:"status"`
	Pid           int               `json:"pid"`
	Ppid          int               `json:"ppid"`
	Children      []ChildrenProcess `json:"children"`
	SumCpuPercent float64           `json:"sumCpuPercent"`
	SumRss        string            `json:"sumRss"`
}

type Processes []Process

// プロセス起動・停止に必要な情報
type ProcessParam struct {
	Env        []string `json:"env"`
	CurrentDir string   `json:"currentDir"`
	Command    string   `json:"command"`
	Args       string   `json:"args"`
	RecordPid  bool     `json:"recordPid"`
	PidFile    string   `json:"pidFile"`
}

const timeformat = "2006/01/02 15:04:05"

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

	// 名前も取れないようなら以後の処理をスキップ(そうじゃないとPanicになる)
	ret.Name, err = p.Name()
	if err != nil {
		log.Printf("error: get process.Name: %v", err)
		return ret, err
	}

	// 死活をチェックして以後の処理をスキップ
	ret.Exist, err = process.PidExists(int32(pid))
	if err != nil {
		log.Printf("error: %v, get process.PidExists: %v", ret.Name, err)
		return ret, err
	}

	cpupercent, err := p.CPUPercent()
	if err != nil {
		log.Printf("error: %v, get process.CPUPercent: %v", ret.Name, err)
		ret.CpuPercent = 0
	} else {
		ret.CpuPercent = math.Round(cpupercent*10) / 10
	}

	cputime, err := p.Times()
	if err != nil {
		log.Printf("error: %v, get process.Time: %v", ret.Name, err)
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
		log.Printf("error: %v, get process.MemoryInfo: %v", ret.Name, err)
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
		log.Printf("error: %v, get process.Cmdline: %v", ret.Name, err)
	}
	ret.Exe, err = p.Exe()
	if err != nil {
		log.Printf("error: %v, get process.Exe: %v", ret.Name, err)
		ret.Exe = err.Error()
	}
	ret.Cwd, err = p.Cwd()
	if err != nil {
		// Winだとcannot read current working directoryになるプロセスがいる（規則性は不明）。MacはOK
		//log.Printf("error: %v, get process.Cwd: %v", ret.Name, err)
		ret.Cwd = err.Error()
	}

	createtime, err := p.CreateTime()
	if err != nil {
		log.Printf("error: %v, get process.CreateTime: %v", ret.Name, err)
	}
	ret.CreateTime = time.Unix(createtime/1000, 0).Format(timeformat)

	// Winだとnot implemented yetとなるプロセスがいる（規則性が不明）。MacはOKなのでこの値は取らない
	/*
		statuses, err := p.Status()
		if err != nil {
			log.Printf("error: %v, get process.Status: %v", ret.Name, err)
		}
		ret.Status = strings.Join(statuses, ", ")
	*/

	ret.Pid = int(p.Pid)

	ppid, err := p.Ppid()
	if err != nil {
		log.Printf("error: %v, get process.Ppid: %v", ret.Name, err)
	}
	ret.Ppid = int(ppid)

	cp := []ChildrenProcess{}
	children, err := p.Children()
	if err != nil {
		return ret, nil
	}
	sumcpu := cpupercent
	sumrss := memory.RSS
	for _, c := range children {
		cname, err := c.Name()
		if err != nil {
			log.Printf("error: get process.Children.Name: %v", err)
			cname = err.Error()
		}

		// Winだとnot implemented yetとなるプロセスがいる（規則性が不明）。MacはOKなのでこの値は取らない
		/*
			_, err = c.Status()
			if err != nil {
				log.Printf("error: %v, get process.Children.Status: %v", cname, err)
			}
		*/

		ccmd, err := c.Cmdline()
		if err != nil {
			log.Printf("error: %v, get process.Children.Cmdline: %v", cname, err)
			ccmd = err.Error()
		}
		var ccpu float64
		ccpupercent, err := c.CPUPercent()
		if err != nil {
			log.Printf("error: %v, get process.Children.CPUPercent: %v", cname, err)
			ccpu = 0
		} else {
			ccpu = math.Round(ccpupercent*10) / 10
			sumcpu = sumcpu + ccpupercent
		}
		cmemory, err := c.MemoryInfo()
		var cvms, crss, cswap string
		if err != nil {
			log.Printf("error: %v, get process.Children.MemoryInfo: %v", cname, err)
			cvms = err.Error()
			crss = err.Error()
			cswap = err.Error()
		} else {
			cvms = bytesize.New(float64(cmemory.VMS)).String()
			crss = bytesize.New(float64(cmemory.RSS)).String()
			cswap = bytesize.New(float64(cmemory.Swap)).String()
			sumrss = sumrss + cmemory.RSS
		}

		cp = append(cp, ChildrenProcess{cname, ccmd, int(c.Pid), ccpu, cvms, crss, cswap})
	}
	ret.Children = cp
	ret.SumCpuPercent = math.Round(sumcpu*10) / 10
	ret.SumRss = bytesize.New(float64(sumrss)).String()

	return ret, nil
}

// StartService 非同期サービスを起動し、PIDを知らせる
func StartService(done chan<- error, param ProcessParam) {
	defer close(done)
	startArgs, err := shellwords.Parse(param.Args)
	if err != nil {
		log.Println(err)
		done <- err
	}
	// param.Commandが空ならstartArgsに全部入っていると見なす
	if param.Command == "" {
		if len(startArgs) == 1 {
			// startArgsが1つならparam.Commandに詰めて空にする
			param.Command = startArgs[0]
			startArgs = nil
		} else if len(startArgs) > 1 {
			// startArgsが2つ以上なら1つ目をparam.Commandに詰めて2つ目以降のパラメーターをstartArgsに詰め直す
			param.Command = startArgs[0]
			startArgs = startArgs[1:]
		} else {
			done <- err
		}
	}

	// 先に環境変数を展開して反映しておかないと修正したPATHがexec.Commandに適用されない
	env := []string{}
	if len(param.Env) > 0 {
		env = setExpandEnv(param.Env)
	}

	var cmd *exec.Cmd
	if len(startArgs) > 0 {
		cmd = exec.Command(param.Command, startArgs...)
	} else {
		cmd = exec.Command(param.Command)
	}
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	stdoutStderr := io.MultiReader(stdout, stderr)

	cmd.Dir = param.CurrentDir
	if len(env) > 0 {
		cmd.Env = env
	}

	setService(cmd)
	if err := cmd.Start(); err != nil {
		done <- err
	} else {
		if param.RecordPid {
			if err := createPidFile(cmd.Process.Pid, param.PidFile); err != nil {
				cmd.Wait()
				done <- err
			}
		}
	}

	scanner := bufio.NewScanner(stdoutStderr)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		done <- err
	}

	done <- nil
}

// StopService サービス停止コマンドを起動し、サービスが終了するまで待つ
func StopService(param ProcessParam) error {

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

	err := cmd.Start()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdoutStderr)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// StopServiceByPid PIDでプロセスを識別してシグナルを送信して終了する
func StopServiceByPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
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
		//log.Printf("env len: %d, env: %v to expandEnv: %v\n", len(env), e, expandEnv)
		if len(env) < 2 {
			// key=value になってない場合はスキップ
			log.Println("env format error, %v", env)
			continue
		}
		if err := os.Setenv(env[0], env[1]); err != nil {
			log.Println(err)
			continue
		}
		//log.Printf("env: %v to expandEnv: %v\n", e, expandEnv)
		env = append(env, expandEnv)
	}
	return env
}

// isExistFile ファイル存在判定
func isExistFile(file string) bool {
	if f, err := os.Stat(file); os.IsNotExist(err) || f.IsDir() {
		return false
	} else {
		return true
	}
}

// isExistDir ディレクトリ存在判定
func isExistDir(dir string) bool {
	if f, err := os.Stat(dir); os.IsNotExist(err) || !f.IsDir() {
		return false
	} else {
		return true
	}
}

// createPidFile pidと書き込むファイル名を受け取ってPIDファイルを作成する
func createPidFile(pid int, pidfile string) error {
	if isExistFile(pidfile) {
		return errors.New("is exist pidfile")
	}

	// dirが無かったら作る
	dir, _ := filepath.Split(pidfile)
	dir = filepath.Clean(dir)
	if !isExistDir(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}

	// PIDファイル作成
	fp, err := os.Create(pidfile)
	if err != nil {
		return err
	}
	defer fp.Close()

	fp.WriteString(fmt.Sprint(pid))

	return nil
}
