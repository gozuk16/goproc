package goproc_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"testing"
	"time"

	"github.com/gozuk16/goproc"
)

// setLogger ログ出力の調整
func setLogger() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestMain(m *testing.M) {
	setLogger()

	/*
		currentdir, _ = os.Getwd()
		IsmHome = filepath.Clean(filepath.Join(currentdir, testIsmHome))
		IsmLog = filepath.Clean(filepath.Join(IsmHome, testIsmLog))
		fmt.Println(IsmHome)
	*/

	m.Run()
}

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

func TestStopService(t *testing.T) {
	usr, _ := user.Current()
	p := []goproc.ProcessParam{
		{WorkingDir: usr.HomeDir, Command: "ls", Args: "-l .."},
		{WorkingDir: "/Users/xxx", Command: "ls"},
		{Command: "top"},
		{SetEnv: []string{"JAVA_HOME=/Users/gozu/.jenv/versions/1.8.0.212", "PATH=$JAVA_HOME/bin:$PATH"}, Command: "java", Args: "-version"},
		{WorkingDir: "/Users/gozu/INFOCOM/ism/service/jetty/demo-base", SetEnv: []string{"JAVA_HOME=/Users/gozu/.jenv/versions/1.8.0.212", "PATH=${JAVA_HOME}/bin:$PATH"}, Command: "java", Args: "-jar ../start.jar STOP.PORT=28282 STOP.KEY=secret jetty.http.port=8081 jetty.ssl.port=8444"},
	}

	cases := []struct {
		param  goproc.ProcessParam
		except bool
		msg    string
	}{
		{p[0], true, "ls起動出来る(エラーがなければ内容は目視で確認)"},
		{p[1], false, "存在しないディレクトリをセットしたらエラー"},
		//{p[2], true, "常駐プロセス(top)"},
		{p[3], true, "環境変数の展開"},
		{p[4], true, "環境変数でJavaを切り替える"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			err := goproc.StopService(c.param)
			if err != nil && err.Error() == "interrupt signal accepted." {
				// TODO: errors.IS()で書き直すには元でError() stringを実装した型を書く
				fmt.Println("interrup is normal.")
			} else if c.except && err != nil {
				t.Errorf("StopProcess = %s, Failed", err)
			} else if !c.except && err == nil {
				t.Errorf("StopProcess nothing err, Failed")
			} else {
				fmt.Printf("%s\n", c.param.Command)
			}
		})
	}

}

func TestStartService(t *testing.T) {
	usr, _ := user.Current()
	p := []goproc.ProcessParam{
		{WorkingDir: usr.HomeDir, Command: "ls", Args: "-l .."},
		{WorkingDir: "/Users/xxx", Command: "ls"},
		{Command: "top"},
		{SetEnv: []string{"JAVA_HOME=/Users/gozu/.jenv/versions/1.8.0.212", "PATH=$JAVA_HOME/bin:$PATH"}, Command: "java", Args: "-version"},
		{WorkingDir: "/Users/gozu/INFOCOM/ism/service/jetty/demo-base", SetEnv: []string{"JAVA_HOME=/Users/gozu/.jenv/versions/1.8.0.212", "PATH=${JAVA_HOME}/bin:$PATH"}, Command: "java", Args: "-jar ../start.jar STOP.PORT=28282 STOP.KEY=secret jetty.http.port=8081 jetty.ssl.port=8444"},
		{WorkingDir: usr.HomeDir, Command: "sh", Args: "-c \"sleep 1 && ls -l\""},
	}

	cases := []struct {
		param  goproc.ProcessParam
		except bool
		msg    string
	}{
		{p[0], true, "ls起動出来る(エラーがなければ内容は目視で確認)"},
		{p[1], false, "存在しないディレクトリをセットしたらエラー"},
		//{p[2], true, "常駐プロセス(top)"},
		{p[3], true, "環境変数の展開"},
		{p[4], true, "環境変数でJavaを切り替える"},
		{p[5], true, "sh経由で起動する"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			// Ctrl+Cを受け取る
			quit := make(chan os.Signal)
			signal.Notify(quit, os.Interrupt)
			done := make(chan error, 1)
			go goproc.StartService(done, c.param)
			// ちょっと待ってからエラーチェック
			time.Sleep(1100 * time.Millisecond)
			select {
			case <-quit:
				fmt.Println("interrup is normal.")
			case err := <-done:
				if err != nil {
					fmt.Println(err)
					if c.except {
						t.Errorf("StartProcess = %s, Failed", err)
					}
				}
			default:
				// 何も無ければ起動したので処理を継続
				if !c.except {
					t.Errorf("StartProcess nothing err, Failed")
				}
			}
		})
	}

}
