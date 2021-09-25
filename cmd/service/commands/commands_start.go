package commands

import (
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use: "start",
		Run: startCommand,
	}
)

func startCommand(cmd *cobra.Command, args []string) {
	if err := startAction(); err != nil {
		Exit(err, 1)
	}
}

func startAction() (err error) {
	// start service
	return nil
}

func init() {
	RootCmd.AddCommand(startCmd)
}
