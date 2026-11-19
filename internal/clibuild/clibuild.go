// Package clibuild constructs the full c8y root command tree offline, without a
// live Cumulocity session. It is the single bootstrap shared by every generator
// that projects from the command tree (docs, man pages, PowerShell cmdlets,
// tests). Keeping one builder guarantees every projection sees the identical
// tree. See proposals/CLI_CODEGEN_INVERSION.md.
package clibuild

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/root"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// NewRootCommand builds the complete c8y command tree with a factory whose
// dependencies (client, console, dataview, …) are unavailable. The tree is
// fully wired (every command registered, every flag and annotation set) so it
// is safe to walk for code generation; it is NOT safe to execute commands.
func NewRootCommand() (*root.CmdRoot, error) {
	var client *c8y.Client
	var dataView *dataview.DataView
	var consoleHandler *console.Console
	var activityLoggerHandler *activitylogger.ActivityLogger

	logHandler := logger.NewLogger("", logger.Options{
		Level: zapcore.WarnLevel,
		Debug: false,
	})

	configHandler := config.NewConfig(viper.GetViper())
	if _, err := configHandler.ReadConfigFiles(nil); err != nil {
		logHandler.Infof("Failed to read configuration. Trying to proceed anyway. %s", err)
	}

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
	return root.NewCmdRoot(cmdFactory, "", ""), nil
}
