package main

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/pingcap/dm/dm/config"
	bf "github.com/pingcap/tidb-tools/pkg/binlog-filter"
	"github.com/pingcap/tidb-tools/pkg/filter"
	router "github.com/pingcap/tidb-tools/pkg/table-router"
	"github.com/spf13/cobra"
)

// NewLoadConfigFileCmd loads a task config file
func NewLoadConfigFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load <config_file>",
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

	var (
		bwList       = make(map[string]*filter.Filter)
		binlogFilter = make(map[string]*bf.BinlogEvent)
		tableRouter  = make(map[string]*router.Table)
	)

	for _, instance := range cfg.MySQLInstances {
		tableRouter[instance.SourceID], _ = router.NewTableRouter(cfg.CaseSensitive, []*router.TableRule{})
		for _, name := range instance.RouteRules {
			if tableRouter[instance.SourceID].AddRule(cfg.Routes[name]) != nil {
				PrintLines("invalid table route %+v rule of instance %s :\n%v", cfg.Routes[name], instance.SourceID, errors.ErrorStack(err))
				return
			}
		}

		filterRules := make([]*bf.BinlogEventRule, len(instance.FilterRules))
		for j, name := range instance.FilterRules {
			filterRules[j] = cfg.Filters[name]
		}
		binlogFilter[instance.SourceID], err = bf.NewBinlogEvent(cfg.CaseSensitive, filterRules)
		if err != nil {
			PrintLines("invalid binlog event filter rule of instance %s :\n%v", instance.SourceID, errors.ErrorStack(err))
			return
		}

		bwList[instance.SourceID] = filter.New(cfg.CaseSensitive, cfg.BWList[instance.BWListName])
	}

	TaskCfg = cfg
	BWList = bwList
	BinlogFilter = binlogFilter
	TableRouter = tableRouter
	PrintLines("load new task config %s successfuly", cfg.Name)
}

// NewCheckBWListCmd filter schema and table using black white list of task config
func NewCheckBWListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bwlist <source-id> <schema> [table]",
		Short: "filter schema and table using black white list of task config",
		Run:   checkBWListFunc,
	}
	return cmd
}

func checkBWListFunc(cmd *cobra.Command, _ []string) {
	if len(cmd.Flags().Args()) < 2 {
		fmt.Println(cmd.Usage())
		return
	}

	var (
		sourceID = cmd.Flags().Arg(0)
		schema   = cmd.Flags().Arg(1)
		table    = ""
	)
	if len(cmd.Flags().Args()) > 1 {
		table = cmd.Flags().Arg(1)
	}

	tables := []*filter.Table{
		{schema, table},
	}

	bw, ok := BWList[sourceID]
	if !ok {
		PrintLines("not found black whitle list of MySQL %s", sourceID)
		return
	}

	resTables := bw.ApplyOn(tables)
	if len(resTables) == 1 {
		PrintLines("replicated")
	} else {
		PrintLines("ignored")
	}
}

// NewShowTaskConfigCmd displays task config
func NewShowTaskConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "displays task config",
		Run:   showTaskConfigFunc,
	}
	return cmd
}

func showTaskConfigFunc(cmd *cobra.Command, _ []string) {
	if TaskCfg == nil {
		PrintLines("not found task config file, please load")
	} else {
		PrintLines("task config %+v", TaskCfg)
	}
}
