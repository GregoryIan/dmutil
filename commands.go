package main

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/pingcap/dm/dm/config"
	"github.com/spf13/cobra"
)

// NewLoadConfigFileCmd loads a task config file
func NewLoadConfigFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-config-file <config_file>",
		Short: "load a task config file",
		Run:   loadConfigFileFunc,
	}
	return cmd
}

// loadConfigFileFunc loads a task config file
func loadConfigFileFunc(cmd *cobra.Command, _ []string) {
	if len(cmd.Flags().Args()) != 1 {
		fmt.Println(cmd.Usage())
		return
	}
	content, err := GetFileContent(cmd.Flags().Arg(0))
	if err != nil {
		PrintLines("get confile file content error:\n%v", errors.ErrorStack(err))
		return
	}

	cfg := config.NewTaskConfig()
	err = cfg.Decode(string(content))
	if err != nil {
		PrintLines("decode task config:\n%v", errors.ErrorStack(err))
		return
	}

	PrintLines("load new task config %s", cfg)
}
