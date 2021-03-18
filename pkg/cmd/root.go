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
	"github.com/reubenmiller/go-c8y-cli/pkg/clierrors"
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

var BootstrapLogger *logger.Logger

var (
	// TODO: Delete these globals!
	cliConfig *config.Config
	client    *c8y.Client
)

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
	log.Println("DEBUG REFACTOR: Running initialize")
	cmd, err := Initialize()
	if err != nil {
		log.Fatalf("DEBUG REFACTOR: Initialize failed. %s", err)
		os.Exit(int(clierrors.ExitError))
	}

	// Expand any aliases
	expandedArgs, err := setArgs(cmd.Command)
	if err != nil {
		Logger.Errorf("Could not expand aliases. %s", err)
	}
	Logger.Debugf("Expanded args: %v", expandedArgs)
	cmd.SetArgs(expandedArgs)

	log.Println("DEBUG REFACTOR: Calling execute")
	if err := cmd.Execute(); err != nil {

		CheckCommandError(cmd.Command, cmd.Factory, err)

		if cErr, ok := err.(cmderrors.CommandError); ok {
			os.Exit(cErr.ExitCode)
		}
		if errors.Is(err, clierrors.ErrNoMatchesFound) {
			// 404
			os.Exit(4)
		}
		os.Exit(100)
	}
}

func CheckCommandError(cmd *cobra.Command, f *cmdutil.Factory, err error) {
	cfg, configErr := f.Config()
	if configErr != nil {
		// todo
	}
	logg, logErr := f.Logger()
	if logErr != nil {

	}
	w := ioutil.Discard
	if cfg != nil && cfg.WithError() {
		w = cmd.OutOrStdout()
	}

	if errors.Is(err, clierrors.ErrNoMatchesFound) {
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
					return nil, cmderrors.NewUserErrorWithExitCode(ee.ExitCode(), ee)
				}

				Logger.Errorf("failed to run external command: %s", err)
				return nil, err
			}
		}
	}
	return expandedArgs, nil
}

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
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
		log.Printf("DEBUG REFACTOR: Read config files failed. Ignoring error. %s", err)
		// return nil, err
	}

	// init logger
	BootstrapLogger = logger.NewLogger(module, logger.Options{
		Level: zapcore.WarnLevel,
		Debug: false,
	})
	logHandler = BootstrapLogger

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
	cmdFactory := factory.New(buildVersion, configFunc, clientFunc, loggerFunc, activityLoggerFunc, dataViewFunc, consoleFunc)
	rootCmd := root.NewCmdRoot(cmdFactory, buildVersion, "")
	consoleHandler = console.NewConsole(rootCmd.OutOrStdout(), func(s []string) []byte {
		return getOutputHeaders(consoleHandler, configHandler, s)
	})

	log.Printf("DEBUG REFACTOR: Settings: %v", configHandler.AllSettings())

	// config file
	cobra.OnInitialize(func() {
		log.Println("DEBUG REFACTOR: cobra.OnInitialize")
		log.Printf("DEBUG REFACTOR: Settings: %v", configHandler.AllSettings())
		logOptions := logger.Options{
			// Level: zapcore.DebugLevel,
			Level: zapcore.WarnLevel,
			Color: !configHandler.DisableColor(),
			Debug: configHandler.Debug(),
		}
		if configHandler.ShowProgress() {
			logOptions.Silent = true
		} else {
			if configHandler.Verbose() {
				logOptions.Level = zapcore.InfoLevel
			}
			if configHandler.Debug() {
				logOptions.Level = zapcore.DebugLevel
			}
		}

		logHandler = logger.NewLogger(module, logOptions)
		c8y.Logger = logHandler
		configHandler.SetLogger(logHandler)
		log.Println("DEBUG REFACTOR: leaving cobra.OnInitialize")
	})

	return rootCmd, nil

}

func initConfig(cfg *config.Config) {

	// only parse env variables if no explict config file is given
	// if globalFlagUseEnv {
	// 	Logger.Println("C8Y_USE_ENVIRONMENT is set. Environment variables can be used to override config settings")
	// 	viper.AutomaticEnv()
	// }

	// set output format
	// Console.Format = cliConfig.GetOutputFormat()
	// Console.Colorized = !cliConfig.DisableColor()
	// Console.Compact = cliConfig.CompactJSON()
	// Console.Disabled = cliConfig.ShowProgress() && isTerminal()

	// Proxy settings
	// Either use explicit proxy, ignore proxy, or use existing env variables
	// --proxy "http://10.0.0.1:8080"
	// --noProxy
	// HTTP_PROXY=http://10.0.0.1:8080
	// NO_PROXY=localhost,127.0.0.1
	proxy := cfg.Proxy()
	noProxy := cfg.IgnoreProxy()
	if noProxy {
		Logger.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
	} else {
		if proxy != "" {
			Logger.Debugf("using explicit proxy [%s]", proxy)

			os.Setenv("HTTP_PROXY", proxy)
			os.Setenv("HTTPS_PROXY", proxy)
			os.Setenv("http_proxy", proxy)
			os.Setenv("https_proxy", proxy)

		} else {
			proxyVars := []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "NO_PROXY", "no_proxy"}

			var proxySettings strings.Builder

			for _, name := range proxyVars {
				if v := os.Getenv(name); v != "" {
					proxySettings.WriteString(fmt.Sprintf(" %s [%s]", name, v))
				}
			}
			if proxySettings.Len() > 0 {
				Logger.Debugf("Using existing env variables.%s", proxySettings)
			}

		}
	}
}
