package goproc_test

import (
	"fmt"
	"testing"

	"github.com/gozuk16/goproc"
)

func TestGetProcess(t *testing.T) {
	// TODO マイナス 0 1 はエラーになる。テーブルドリブンに書き換える
	p, err := goproc.GetProcess(1)
	if err == nil {
		t.Errorf("GetProcess = %s, Failed", err)
	}
	fmt.Println(p)
}
