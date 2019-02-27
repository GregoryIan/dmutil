package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/juju/errors"
	"github.com/pingcap/dm/dm/config"
	bf "github.com/pingcap/tidb-tools/pkg/binlog-filter"
	"github.com/pingcap/tidb-tools/pkg/filter"
	"github.com/pingcap/tidb-tools/pkg/table-router"
	"github.com/spf13/cobra"
)

var (
	// TaskCfg is config of task
	TaskCfg *config.TaskConfig
	// BWList is black white list of task
	BWList map[string]*filter.Filter
	// BinlogFilter is binlog event filter list of task
	BinlogFilter map[string]*bf.BinlogEvent
	// TableRouter is table router of task
	TableRouter map[string]*router.Table
)

// Start starts running a command
func Start(args []string) {
	rootCmd := &cobra.Command{
		Use:   "dmutil",
		Short: "DM Util Tools",
	}

	rootCmd.AddCommand(
		NewLoadConfigFileCmd(),
		NewCheckBWListCmd(),
		NewShowTaskConfigCmd(),
		//NewCheckTableRouteCmd(),
		//NewCheckBinlogEventFilterCmd(),
	)

	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(rootCmd.UsageString())
	}
}

// PrintLines adds a wrap to support `\n` within `chzyer/readline`
func PrintLines(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
}

// GetFileContent reads and returns file's content
func GetFileContent(fpath string) ([]byte, error) {
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return content, nil
}

func main() {
	fmt.Println("Welcome dm util tool!")

	//log.SetLevelByString("error")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		fmt.Printf("got signal [%v] to exit", sig)
		switch sig {
		case syscall.SIGTERM:
			os.Exit(0)
		default:
			os.Exit(1)
		}
	}()

	loop()

	fmt.Println("dmctl exit")
}

func loop() {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/dmctlreadline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				break
			} else if err == io.EOF {
				break
			}
			continue
		}
		if line == "exit" {
			os.Exit(0)
		} else if line == "" {
			continue
		}

		args := strings.Fields(line)
		Start(args)
	}
}
