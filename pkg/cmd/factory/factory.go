package factory

import (
	"os"

	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func New(appVersion string, buildBranch string, configFunc func() (*config.Config, error), clientFunc func() (*c8y.Client, error), loggerFunc func() (*logger.Logger, error), activityLoggerFunc func() (*activitylogger.ActivityLogger, error), dataViewFunc func() (*dataview.DataView, error), consoleFunc func() (*console.Console, error)) *cmdutil.Factory {
	io := iostreams.System(false, true)

	c8yExecutable := "c8y"
	if exe, err := os.Executable(); err == nil {
		c8yExecutable = exe
	}

	f := &cmdutil.Factory{
		IOStreams:      io,
		Config:         configFunc,
		Client:         clientFunc,
		Executable:     c8yExecutable,
		Logger:         loggerFunc,
		ActivityLogger: activityLoggerFunc,
		DataView:       dataViewFunc,
		Console:        consoleFunc,
		BuildVersion:   appVersion,
		BuildBranch:    buildBranch,
	}
	f.Browser = browser(f)
	return f
}

func browser(f *cmdutil.Factory) cmdutil.Browser {
	io := f.IOStreams
	return cmdutil.NewBrowser(browserLauncher(f), io.Out, io.ErrOut)
}

// Browser precedence
// 1. browser from config
// 2. BROWSER
func browserLauncher(f *cmdutil.Factory) string {
	cfg, err := f.Config()
	if err == nil {

		if cfgBrowser := cfg.Browser(); cfgBrowser != "" {
			return cfgBrowser
		}
	}

	return os.Getenv("BROWSER")
}
