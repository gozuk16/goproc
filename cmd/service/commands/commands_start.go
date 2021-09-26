package commands

import (
	"errors"
	"fmt"

	"github.com/gozuk16/goproc"
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use: "start",
		Run: startCommand,
	}
)

func startCommand(cmd *cobra.Command, args []string) {
	var p goproc.ProcessParam
	if len(args) > 0 {
		p.StartCmd = args[0]
	} else {
		Exit(errors.New("parameter not found"), 1)
	}
	if err := startAction(p); err != nil {
		Exit(err, 1)
	}
}

func startAction(p goproc.ProcessParam) (err error) {
	// start service
	pid, err := goproc.StartProcess(p)
	if err != nil {
		return err
	}
	fmt.Println(pid)

	return nil
}

func init() {
	RootCmd.AddCommand(startCmd)
}
