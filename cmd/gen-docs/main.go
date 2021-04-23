package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/internal/docs"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/root"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

func main() {
	var flagError pflag.ErrorHandling

	docCmd := pflag.NewFlagSet("", flagError)
	manPage := docCmd.BoolP("man-page", "", false, "Generate manual pages")
	website := docCmd.BoolP("website", "", false, "Generate website pages")
	dir := docCmd.StringP("doc-path", "", "", "Path directory where you want generate doc files")
	help := docCmd.BoolP("help", "h", false, "Help about any command")

	if err := docCmd.Parse(os.Args); err != nil {
		os.Exit(1)
	}

	if *help {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", os.Args[0], docCmd.FlagUsages())
		if err != nil {
			fatal(err)
		}
		os.Exit(1)
	}

	if *dir == "" {
		fatal("no dir set")
	}

	rootCmd := createCmdRoot()
	rootCmd.InitDefaultHelpCmd()

	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		fatal(err)
	}

	if *website {
		err = docs.GenMarkdownTreeCustom(rootCmd.Command, *dir, filePrepender, linkHandler)
		if err != nil {
			fatal(err)
		}
	}

	if *manPage {

		header := &docs.GenManHeader{
			Title:   "c8y",
			Section: "1",
			Source:  "",
			Manual:  "",
		}
		err = docs.GenManTree(rootCmd.Command, header, *dir)
		if err != nil {
			fatal(err)
		}
	}
}

func filePrepender(filename string, opts ...string) string {
	var category, fullCommand string
	if len(opts) >= 3 {
		category = opts[0]
		// title = opts[1]
		fullCommand = opts[2]
	}
	header := fmt.Sprintf(`---
category: %s
title: %s
---
`, category, fullCommand)
	return header
}

func linkHandler(name string, opts ...string) string {
	return fmt.Sprintf("./%s", strings.TrimSuffix(name, ".md"))
}

func fatal(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func createCmdRoot() *root.CmdRoot {
	var client *c8y.Client
	var dataView *dataview.DataView
	var consoleHandler *console.Console
	var logHandler *logger.Logger
	var activityLoggerHandler *activitylogger.ActivityLogger
	var configHandler = config.NewConfig(viper.GetViper())

	// init logger
	logHandler = logger.NewLogger("", logger.Options{
		Level: zapcore.WarnLevel,
		Debug: false,
	})

	if _, err := configHandler.ReadConfigFiles(nil); err != nil {
		logHandler.Infof("Failed to read configuration. Trying to proceed anyway. %s", err)
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
	loggerFunc := func() (*logger.Logger, error) {
		if logHandler == nil {
			return nil, fmt.Errorf("logger is missing")
		}
		return logHandler, nil
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
			return nil, fmt.Errorf("console is missing")
		}
		return consoleHandler, nil
	}
	cmdFactory := factory.New("", "", configFunc, clientFunc, loggerFunc, activityLoggerFunc, dataViewFunc, consoleFunc)

	return root.NewCmdRoot(cmdFactory, "", "")
}
