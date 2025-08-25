package root

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	module = "c8yapi"
)

func init() {
	// Enable case insensitive matches
	cobra.EnableCaseInsensitive = true
}

// GetInitLoggerOptions create a simple logger with a best-guess log level based on the given arguments
// It will not activate on any environment variables, however it will do a simplistic parsing of
// the common logging options before the configuration has been read which enables debugging around
// the configuration and extensions etc.
func GetInitLoggerOptions(args []string) logger.Options {
	color := true
	level := slog.LevelWarn
	debug := false

	for _, item := range args {
		switch item {
		case "--debug", "--debug=true":
			level = slog.LevelDebug
			debug = true
		case "--verbose", "-v", "--verbose=true":
			level = slog.LevelInfo
		case "--noColor", "--noColor=true", "-M", "-M=true":
			color = false
		}
	}

	return logger.Options{
		Level: level,
		Debug: debug,
		Color: color,
	}
}

func getOutputHeaders(c *console.Console, cfg *config.Config, input []string) (headers []byte) {
	if !c.IsCSV() || !c.WithCSVHeader() || len(input) == 0 {
		slog.Debug("Ignoring csv headers", "isCSV", c.IsCSV(), "withHeader", c.WithCSVHeader())
		return
	}
	if len(input) > 0 {
		return []byte(input[0] + "\n")
	}

	// TODO: improve detection by parsing more lines to find column names (if more lines are available)
	columns := make([][]byte, 0)
	for _, v := range cfg.GetJSONSelect() {
		for _, column := range strings.Split(v, ",") {

			if i := strings.Index(column, ":"); i > -1 {
				columns = append(columns, []byte(column[0:i]))
			} else {
				columns = append(columns, []byte(column))
			}
		}
	}
	return append(bytes.Join(columns, []byte(",")), []byte("\n")...)
}

func ShouldIgnoreSessionFile(args []string) bool {
	cmdStr := strings.Join(args, " ")
	return strings.Contains(cmdStr, "c8y sessions login") || strings.Contains(cmdStr, "c8y sessions set")
}

// Initialize initializes the configuration manager and c8y client
func NewCommand(buildVersion, buildBranch string) (*CmdRoot, error) {
	if buildVersion == "" {
		buildVersion = "0.0.0"
	}
	if buildBranch == "" {
		buildBranch = "unknown"
	}

	var client *c8y.Client
	var dataView *dataview.DataView
	var consoleHandler *console.Console
	var activityLoggerHandler *activitylogger.ActivityLogger
	var configHandler = config.NewConfig(viper.GetViper())

	// Note: Loading of the session should be deferred until the commands are loaded
	if _, err := configHandler.ReadConfigFiles(nil, ShouldIgnoreSessionFile(os.Args)); err != nil {
		slog.Info("Failed to read configuration. Trying to proceed anyway", "err", err)
	}

	// cmd factory
	configFunc := func() (*config.Config, error) {
		if configHandler == nil {
			return nil, fmt.Errorf("config is missing")
		}
		return configHandler, nil
	}
	clientFunc := func() (*c8y.Client, error) {
		if client == nil {
			return nil, fmt.Errorf("client is missing")
		}
		return client, nil
	}
	activityLoggerFunc := func() (*activitylogger.ActivityLogger, error) {
		if activityLoggerHandler == nil {
			return nil, fmt.Errorf("activityLogger is missing")
		}
		return activityLoggerHandler, nil
	}
	dataViewFunc := func() (*dataview.DataView, error) {
		if dataView == nil {
			return nil, fmt.Errorf("dataView is missing")
		}
		return dataView, nil
	}

	consoleFunc := func() (*console.Console, error) {
		if consoleHandler == nil {
			// return nil, fmt.Errorf("console is missing")
		}
		return consoleHandler, nil
	}
	cmdFactory := factory.New(buildVersion, buildBranch, configFunc, clientFunc, activityLoggerFunc, dataViewFunc, consoleFunc)

	// Register the template resolver so the configuration can lookup values as needed
	configHandler.RegisterTemplateResolver(cmdutil.NewTemplateResolver(cmdFactory))

	rootCmd := NewCmdRoot(cmdFactory, buildVersion, "")

	// Add reference to root command
	cmdFactory.SetCommand(rootCmd.Command)

	tableOptions := &console.TableOptions{
		MinColumnWidth:           configHandler.ViewColumnMinWidth(),
		MaxColumnWidth:           configHandler.ViewColumnMaxWidth(),
		MinEmptyValueColumnWidth: configHandler.ViewColumnEmptyValueMinWidth(),
		ColumnPadding:            configHandler.ViewColumnPadding(),
		RowMode:                  configHandler.ViewRowMode(),
		NumberFormatter:          configHandler.GetTableViewNumberFormatter(),
	}
	consoleHandler = console.NewConsole(rootCmd.OutOrStdout(), tableOptions, func(s []string) []byte {
		return getOutputHeaders(consoleHandler, configHandler, s)
	})

	return rootCmd, nil
}
