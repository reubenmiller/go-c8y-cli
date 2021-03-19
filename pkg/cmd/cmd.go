package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/reubenmiller/go-c8y-cli/internal/run"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias/expand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/root"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// Logger is used to record the log messages which should be visible to the user when using the verbose flag
var Logger *logger.Logger

// Build data
// These variables should be set using the -ldflags "-X github.com/reubenmiller/go-c8y-cli/pkg/cmd.version=1.0.0" when running go build
var buildVersion string
var buildBranch string

const (
	module = "c8yapi"
)

func init() {
	Logger = logger.NewLogger(module, logger.Options{})

	// set seed for random generation
	rand.Seed(time.Now().UTC().UnixNano())
}

// Execute runs the root command
func MainRun() {
	cmd, err := Initialize()
	if err != nil {
		os.Exit(int(cmderrors.ExitError))
	}

	// Expand any aliases
	expandedArgs, err := setArgs(cmd.Command)
	if err != nil {
		Logger.Errorf("Could not expand aliases. %s", err)
	}
	Logger.Debugf("Expanded args: %v", expandedArgs)
	cmd.SetArgs(expandedArgs)

	if err := cmd.Execute(); err != nil {
		CheckCommandError(cmd.Command, cmd.Factory, err)

		if cErr, ok := err.(cmderrors.CommandError); ok {
			os.Exit(int(cErr.ExitCode))
		}
		if errors.Is(err, cmderrors.ErrNoMatchesFound) {
			os.Exit(int(cmderrors.ExitNotFound404))
		}
		os.Exit(int(cmderrors.ExitError))
	}
}

func CheckCommandError(cmd *cobra.Command, f *cmdutil.Factory, err error) {
	cfg, configErr := f.Config()
	if configErr != nil {
		// TODO: Handle error or at least log it
	}
	logg, logErr := f.Logger()
	if logErr != nil {
		// TODO: Handle error or at least log it
	}
	w := ioutil.Discard
	if cfg != nil && cfg.WithError() {
		w = cmd.OutOrStdout()
	}

	if errors.Is(err, cmderrors.ErrNoMatchesFound) {
		// Simulate a 404 error
		customErr := cmderrors.CommandError{}
		customErr.StatusCode = 404
		customErr.ExitCode = 4
		customErr.Message = err.Error()
		err = customErr
	}

	if cErr, ok := err.(cmderrors.CommandError); ok {
		if cErr.StatusCode == 403 || cErr.StatusCode == 401 {
			logg.Error(fmt.Sprintf("Authentication failed (statusCode=%d). Try to run set-session again, or check the password", cErr.StatusCode))
		}

		// format errors as json messages
		// only log users errors
		silentStatusCodes := ""
		if cfg != nil {
			silentStatusCodes = cfg.GetSilentStatusCodes()
		}
		if !cErr.IsSilent() && !strings.Contains(silentStatusCodes, fmt.Sprintf("%d", cErr.StatusCode)) {
			logg.Errorf("%s", cErr)
			fmt.Fprintf(w, "%s\n", cErr.JSONString())
		}
	} else {
		// unexpected error
		cErr := cmderrors.NewSystemErrorF("%s", err)
		logg.Errorf("%s", cErr)
		fmt.Fprintf(w, "%s\n", cErr.JSONString())
	}
}

func setArgs(cmd *cobra.Command) ([]string, error) {
	expandedArgs := []string{}
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}
	cmd, _, err := cmd.Traverse(expandedArgs)
	if err != nil || cmd == cmd.Root() {
		originalArgs := expandedArgs
		isShell := false

		v := viper.GetViper()
		aliases := v.GetStringMapString("settings.aliases")
		expandedArgs, isShell, err = expand.ExpandAlias(aliases, os.Args, nil)
		if err != nil {
			Logger.Errorf("failed to process aliases:  %s", err)
			return nil, err
		}

		Logger.Debugf("%v -> %v", originalArgs, expandedArgs)

		if isShell {
			exe, err := safeexec.LookPath(expandedArgs[0])
			if err != nil {
				Logger.Errorf("failed to run external command: %s", err)
				return nil, err
			}

			externalCmd := exec.Command(exe, expandedArgs[1:]...)
			externalCmd.Stderr = os.Stderr
			externalCmd.Stdout = os.Stdout
			externalCmd.Stdin = os.Stdin
			preparedCmd := run.PrepareCmd(externalCmd)

			err = preparedCmd.Run()
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					return nil, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCode(ee.ExitCode()), ee)
				}

				Logger.Errorf("failed to run external command: %s", err)
				return nil, err
			}
		}
	}
	return expandedArgs, nil
}

func getOutputHeaders(c *console.Console, cfg *config.Config, input []string) (headers []byte) {
	if !c.IsCSV() || !c.WithCSVHeader() || len(input) == 0 {
		Logger.Debugf("Ignoring csv headers: isCSV=%v, WithHeader=%v", c.IsCSV(), c.WithCSVHeader())
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

// Initialize initializes the configuration manager and c8y client
func Initialize() (*root.CmdRoot, error) {

	var client *c8y.Client
	var dataView *dataview.DataView
	var consoleHandler *console.Console
	var logHandler *logger.Logger
	var activityLoggerHandler *activitylogger.ActivityLogger
	var configHandler = config.NewConfig(viper.GetViper())
	if _, err := configHandler.ReadConfigFiles(nil); err != nil {
		log.Printf("WARNING: Failed to read configuration. Trying to proceed anyway. %s", err)
		// return nil, err
	}

	// init logger
	logHandler = logger.NewLogger(module, logger.Options{
		Level: zapcore.WarnLevel,
		Debug: false,
	})

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
	cmdFactory := factory.New(buildVersion, buildBranch, configFunc, clientFunc, loggerFunc, activityLoggerFunc, dataViewFunc, consoleFunc)
	rootCmd := root.NewCmdRoot(cmdFactory, buildVersion, "")
	consoleHandler = console.NewConsole(rootCmd.OutOrStdout(), func(s []string) []byte {
		return getOutputHeaders(consoleHandler, configHandler, s)
	})

	return rootCmd, nil
}
