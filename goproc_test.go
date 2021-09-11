package goproc_test

import (
	"fmt"
	"testing"

	"github.com/gozuk16/goproc"
)

/*
type TestCase struct {
	in   string
	except bool
	msg    string
}
*/

func TestGetProcess(t *testing.T) {
	cases := []struct {
		in  int
		msg string
	}{
		{0, "0はエラーで返す"},
		{1, "1はエラーで返す"},
		{-1, "マイナスはエラーで返す"},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			p, err := goproc.GetProcess(c.in)
			if err == nil {
				// マイナス 0 1 がエラーで返らない場合はFail。エラー返さないとPanicになる
				t.Errorf("GetProcess = %s, Failed", err)
			}
			fmt.Println(p)
		})
	}
}
