package goproc_test

import (
	"fmt"
	"os"
	"os/user"
	"testing"

	"github.com/gozuk16/goproc"
)

func TestGetProcessError(t *testing.T) {
	// マイナス 0 1 がエラーで返らない場合はFail
	cases := []struct {
		in  int
		msg string
	}{
		{0, "0はエラーで返す"},
		{1, "1はエラーで返す"},
		{-1, "マイナスはエラーで返す"},
		{99999, "存在しないPIDはエラーで返す(99999はたいていない想定)"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			_, err := goproc.GetProcess(c.in)
			if err == nil {
				//エラーを返さないとPanicになる
				t.Errorf("GetProcess = %s, Failed", err)
			}
			fmt.Println(err)
		})
	}
}

func TestGetProcesses(t *testing.T) {
	ins := [][]int{
		{0, 1, -1},
		{99999},
		{0},
		{0, 0},
		{0, 0},
	}
	ins[2][0] = os.Getpid()
	ins[3][1] = os.Getpid()
	ins[4][0] = os.Getpid()

	cases := []struct {
		in     []int
		except int
		msg    string
	}{
		{ins[0], 0, "0 1 -1は無視される"},
		{ins[1], 0, "存在しないPID(99999)は無視される(99999はたいていない想定)"},
		{ins[2], 1, "自分自身のPIDを渡す"},
		{ins[3], 1, "0と自分自身のPIDを渡す(0だけ無視される)"},
		{ins[4], 1, "自分自身のPIDと0を渡す(0だけ無視される)"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			p, err := goproc.GetProcesses(c.in)
			if err != nil {
				t.Errorf("GetProcesses = %s, Failed", err)
			}
			if c.except != len(p) {
				t.Errorf("Process num = %v, Failed", len(p))
			}
		})
	}
}

func TestStartProcess(t *testing.T) {
	usr, _ := user.Current()
	p := []goproc.ProcessParam{
		{CurrentDir: usr.HomeDir, Command: "ls", Args: "-l .."},
		{CurrentDir: "/Users/xxx", Command: "ls"},
		{Command: "top"},
	}

	cases := []struct {
		param  goproc.ProcessParam
		except bool
		msg    string
	}{
		{p[0], true, "ls起動出来る"},
		{p[1], false, "存在しないディレクトリをセットしたらエラー"},
		{p[2], true, "常駐プロセス(top)"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			pid, err := goproc.StartProcess(c.param)
			if c.except && err != nil {
				t.Errorf("StartProcess = %s(%d), Failed", err, pid)
			} else if !c.except && err == nil {
				t.Errorf("StartProcess nothing err, Failed")
			} else {
				fmt.Printf("%s(pid: %d)\n", c.param.Command, pid)
			}
		})
	}

}

func TestRunProcess(t *testing.T) {
	usr, _ := user.Current()
	p := []goproc.ProcessParam{
		{CurrentDir: usr.HomeDir, Command: "ls", Args: "-l .."},
		{CurrentDir: "/Users/xxx", Command: "ls"},
		{Command: "top"},
	}

	cases := []struct {
		param  goproc.ProcessParam
		except bool
		msg    string
	}{
		{p[0], true, "ls起動出来る(エラーがなければ内容は目視で確認)"},
		{p[1], false, "存在しないディレクトリをセットしたらエラー"},
		{p[2], true, "常駐プロセス(top)"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			err := goproc.RunProcess(c.param)
			if err != nil && err.Error() == "interrupt signal accepted." {
				// TODO: errors.IS()で書き直すには元でError() stringを実装した型を書く
				fmt.Println("interrup is normal.")
			} else if c.except && err != nil {
				t.Errorf("StartProcess = %s, Failed", err)
			} else if !c.except && err == nil {
				t.Errorf("StartProcess nothing err, Failed")
			} else {
				fmt.Printf("%s\n", c.param.Command)
			}
		})
	}

}
