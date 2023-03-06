package goproc

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

// overwritten with os.Interrupt on windows environment (see goproc_windows.go)
var stopSignal = syscall.SIGTERM

// setService Group PidとSession idを親プロセスから分離する
func setService(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func GetEnviron(p *process.Process) ([]string, error) {
	result, err := exec.Command("ps", "-p", strconv.Itoa(int(p.Pid)), "-Eww", "-o", "command").Output()
	if err != nil {
		log.Printf("getEnviron: %v", err)
		return nil, nil
	}
	s := strings.Split(string(result), "\n")
	envs := getEnvironFromPsCommand(s[1])
	return envs, nil
}

func getEnvironFromPsCommand(str string) []string {
	//log.Printf("= %d, %s", strings.Count(str, "="), str)
	var envs []string
	var appendEnv string
	var isAppend, isEnv bool
	var i int
	// 環境変数を前から1つずつスペースで分割して処理していく（スペースや=が値に入ってくることを考慮）
	for {
		before, after, ok := strings.Cut(str, " ")
		//log.Printf("%d: %s, %s, %v: %s", i, before, after, ok, appendEnv)
		i++
		// 1つ目はコマンドなので必ず飛ばす
		if i == 1 {
			str = after
			continue
		}
		if ok {
			// 2つ目以降にコマンドの引数があれば飛ばす
			if !isEnv {
				// -で始まるか、=が入ってない、javaで始まる場合は引数と見なす
				if strings.HasPrefix(before, "-") || strings.HasPrefix(before, "java") || strings.Count(before, "=") == 0 {
					str = after
					continue
				}
				// 引数じゃない値が出現したら以降は環境変数と見なす
				isEnv = true
			}

			// afterをスペースで分割して、前側に=が入ってなければ最初の候補にスペース入りの値があったと見なす
			b, a, ok := strings.Cut(after, " ")
			if strings.Count(b, "=") == 0 {
				if appendEnv == "" {
					appendEnv = before + " " + b
				} else {
					appendEnv += " " + before + " " + b
				}
				if ok {
					str = a
					// 後側の1つめに=が入っていればスペース入り環境変数の終わりと見なす
					if strings.Count(strings.Split(a, " ")[0], "=") > 0 {
						isAppend = true
					}
				} else {
					envs = append(envs, appendEnv)
					break
				}
			} else {
				if appendEnv == "" {
					appendEnv = before
				} else {
					appendEnv += " " + before
				}
				envs = append(envs, appendEnv)
				str = after
				appendEnv = ""
			}
		} else {
			// 最後の1つなら抜ける
			// =が無かった場合は前の変数のオプションと見なす
			if strings.Count(before, "=") == 0 {
				before = appendEnv + " " + before
			}
			envs = append(envs, before)
			break
		}
		if isAppend {
			envs = append(envs, appendEnv)
			appendEnv = ""
			isAppend = false
		}
	}
	//log.Printf("%#v", envs)
	return envs
}
