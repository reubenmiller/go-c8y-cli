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
	"github.com/reubenmiller/go-c8y-cli/v2/internal/run"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alias/expand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/root"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// Logger is used to record the log messages which should be visible to the user when using the verbose flag
var Logger *logger.Logger

// Build data
// These variables should be set using the -ldflags "-X github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd.version=1.0.0" when running go build
var buildVersion string
var buildBranch string

const (
	module = "c8yapi"
)

func init() {
	Logger = logger.NewLogger(module, logger.Options{})

	// set seed for random generation
	rand.Seed(time.Now().UTC().UnixNano())

	// Enable case insensitive matches
	cobra.EnableCaseInsensitive = true
}

// Execute runs the root command
func MainRun() {
	rootCmd, err := Initialize()
	if err != nil {
		os.Exit(int(cmderrors.ExitError))
	}

	// Expand any aliases
	expandedArgs, err := setArgs(rootCmd.Command)
	if err != nil {
		Logger.Errorf("Could not expand aliases. %s", err)
		os.Exit(int(cmderrors.ExitInvalidAlias))
	}

	// Completions for aliases
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		v := viper.GetViper()
		aliases := v.GetStringMapString(config.SettingsCommonAliases)
		for key, value := range v.GetStringMapString(config.SettingsAliases) {
			aliases[key] = value
		}

		var results []string
		for aliasName, aliasValue := range aliases {
			if strings.HasPrefix(aliasName, toComplete) {
				var s string
				if strings.HasPrefix(aliasValue, "!") {
					s = fmt.Sprintf("%s\tShell alias", aliasName)
				} else {
					if len(aliasValue) > 80 {
						aliasValue = aliasValue[:80] + "..."
					}
					s = fmt.Sprintf("%s\tAlias for %s", aliasName, aliasValue)
				}
				results = append(results, s)
			}
		}
		return results, cobra.ShellCompDirectiveNoFileComp
	}

	Logger.Debugf("Expanded args: %v", expandedArgs)
	rootCmd.SetArgs(expandedArgs)

	if err := rootCmd.Execute(); err != nil {
		err = CheckCommandError(rootCmd.Command, rootCmd.Factory, err)

		// Help is not really error, just a way to exit early
		// after displaying help to ther user
		if errors.Is(err, cmderrors.ErrHelp) {
			os.Exit(int(cmderrors.ExitOK))
		}

		if cErr, ok := err.(cmderrors.CommandError); ok {
			os.Exit(int(cErr.ExitCode))
		}
		if errors.Is(err, cmderrors.ErrNoMatchesFound) {
			os.Exit(int(cmderrors.ExitNotFound404))
		}
		os.Exit(int(cmderrors.ExitError))
	}
}

func CheckCommandError(cmd *cobra.Command, f *cmdutil.Factory, err error) error {
	cfg, configErr := f.Config()
	if configErr != nil {
		log.Fatalf("Could not load configuration. %s", configErr)
	}
	logg, logErr := f.Logger()
	if logErr != nil {
		log.Fatalf("Could not configure logger. %s", logErr)
	}
	w := ioutil.Discard

	if errors.Is(err, cmderrors.ErrHelp) {
		return err
	}

	if iterator.IsEmptyPipeInputError(err) && !cfg.AllowEmptyPipe() {
		// Ignore empty pipe errors
		logg.Debug("detected empty piped data")
		return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitOK)
	}

	// read directly from flags, as unknown flag errors are thrown before the config is read
	if localWithError, fErr := cmd.Flags().GetBool("withError"); localWithError && fErr == nil {
		w = cmd.OutOrStdout()
	} else if cfg != nil && cfg.WithError() {
		w = cmd.OutOrStdout()
	}

	if errors.Is(err, cmderrors.ErrNoMatchesFound) {
		// Simulate a 404 error
		customErr := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitNotFound404, err)
		customErr.StatusCode = 404
		err = customErr
	}

	if errors.Is(err, encrypt.ErrDecryptFailed) {
		// Decryption error
		customErr := cmderrors.NewUserErrorWithExitCode(cmderrors.ExitDecryption, err)
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

			if !cErr.Processed {
				logg.Errorf("%s", cErr)
				fmt.Fprintf(w, "%s\n", cErr.JSONString())
			} else {
				logg.Debugf("Error has already been logged. %s", cErr)
			}
		}
	} else {
		// unexpected error
		cErr := cmderrors.NewSystemErrorF("%s", err)
		logg.Errorf("%s", cErr)
		fmt.Fprintf(w, "%s\n", cErr.JSONString())
	}
	return err
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
		aliases := v.GetStringMapString(config.SettingsCommonAliases)
		for name, value := range v.GetStringMapString(config.SettingsAliases) {
			aliases[name] = value
		}
		expandedArgs, isShell, err = expand.ExpandAlias(aliases, os.Args, nil)
		if err != nil {
			return nil, err
		}

		Logger.Debugf("%v -> %v", originalArgs, expandedArgs)

		if isShell {
			exe, err := safeexec.LookPath(expandedArgs[0])
			if err != nil {
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

				return nil, err
			}
			os.Exit(int(cmderrors.ExitOK))
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

	// init logger
	logHandler = logger.NewLogger(module, logger.Options{
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
	cmdFactory := factory.New(buildVersion, buildBranch, configFunc, clientFunc, loggerFunc, activityLoggerFunc, dataViewFunc, consoleFunc)
	rootCmd := root.NewCmdRoot(cmdFactory, buildVersion, "")

	tableOptions := &console.TableOptions{
		MinColumnWidth:           configHandler.ViewColumnMinWidth(),
		MaxColumnWidth:           configHandler.ViewColumnMaxWidth(),
		MinEmptyValueColumnWidth: configHandler.ViewColumnEmptyValueMinWidth(),
		ColumnPadding:            configHandler.ViewColumnPadding(),
		RowMode:                  configHandler.ViewRowMode(),
	}
	consoleHandler = console.NewConsole(rootCmd.OutOrStdout(), tableOptions, func(s []string) []byte {
		return getOutputHeaders(consoleHandler, configHandler, s)
	})

	return rootCmd, nil
}
