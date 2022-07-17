package cmdutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mode"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type Browser interface {
	Browse(string) error
}

type Factory struct {
	IOStreams      *iostreams.IOStreams
	Browser        Browser
	Client         func() (*c8y.Client, error)
	Config         func() (*config.Config, error)
	Logger         func() (*logger.Logger, error)
	ActivityLogger func() (*activitylogger.ActivityLogger, error)
	Console        func() (*console.Console, error)
	DataView       func() (*dataview.DataView, error)

	BuildVersion string
	BuildBranch  string

	// Executable is the path to the currently invoked binary
	Executable string
}

// CreateModeEnabled create mode is enabled
func (f *Factory) CreateModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	)

	if cfg.DryRun() {
		return nil
	}
	return mode.ValidateCreateMode(cfg)
}

// ValidateUpdateMode update mode is enabled
func (f *Factory) UpdateModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	)

	if cfg.DryRun() {
		return nil
	}
	return mode.ValidateUpdateMode(cfg)
}

// ValidateDeleteMode delete mode is enabled
func (f *Factory) DeleteModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	)

	if cfg.DryRun() {
		return nil
	}
	return mode.ValidateDeleteMode(cfg)
}

func (f *Factory) GetRequestHandler() (*request.RequestHandler, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	log, err := f.Logger()
	if err != nil {
		return nil, err
	}

	activityLogger, err := f.ActivityLogger()
	if err != nil {
		return nil, err
	}
	consol, err := f.Console()
	if err != nil {
		return nil, err
	}
	dataview, err := f.DataView()
	if err != nil {
		return nil, err
	}
	client, err := f.Client()
	if err != nil {
		return nil, err
	}

	handler := &request.RequestHandler{
		IsTerminal:     f.IOStreams.IsStdoutTTY(),
		IO:             f.IOStreams,
		Client:         client,
		Config:         cfg,
		Logger:         log,
		DataView:       dataview,
		Console:        consol,
		ActivityLogger: activityLogger,
		HideSensitive:  cfg.HideSensitiveInformationIfActive,
	}
	return handler, nil
}

func (f *Factory) RunWithWorkers(client *c8y.Client, cmd *cobra.Command, req *c8y.RequestOptions, inputIterators *flags.RequestInputIterators) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	log, err := f.Logger()
	if err != nil {
		return err
	}

	activityLogger, err := f.ActivityLogger()
	if err != nil {
		return err
	}
	consol, err := f.Console()
	if err != nil {
		return err
	}
	dataview, err := f.DataView()
	if err != nil {
		return err
	}

	handler := &request.RequestHandler{
		IsTerminal:     f.IOStreams.IsStdoutTTY(),
		IO:             f.IOStreams,
		Client:         client,
		Config:         cfg,
		Logger:         log,
		DataView:       dataview,
		Console:        consol,
		ActivityLogger: activityLogger,
		HideSensitive:  cfg.HideSensitiveInformationIfActive,
	}
	w, err := worker.NewWorker(log, cfg, f.IOStreams, client, activityLogger, handler.ProcessRequestAndResponse, f.CheckPostCommandError)

	if err != nil {
		return err
	}
	return w.ProcessRequestAndResponse(cmd, req, inputIterators)
}

// GetViewProperties Look up the view properties to display
func (f *Factory) GetViewProperties(cfg *config.Config, cmd *cobra.Command, output []byte) ([]string, error) {
	dataview, err := f.DataView()
	if err != nil {
		return nil, err
	}
	log, err := f.Logger()
	if err != nil {
		return nil, err
	}

	view := cfg.ViewOption()
	showRaw := cfg.RawOutput() || cfg.WithTotalPages()

	if showRaw {
		return []string{"**"}, nil
	}
	viewProperties := []string{}
	switch strings.ToLower(view) {
	case config.ViewsOff:
		// dont apply a view
		return []string{"**"}, nil
	case config.ViewsAuto:
		jsonResponse := gjson.ParseBytes(output)
		props, err := dataview.GetView(&jsonResponse, "")

		if err != nil || len(props) == 0 {
			if err != nil {
				log.Infof("No matching view detected. defaulting to '**'. %s", err)
			} else {
				log.Info("No matching view detected. defaulting to '**'")
			}
			viewProperties = append(viewProperties, "**")
		} else {
			log.Infof("Detected view: %s", strings.Join(props, ", "))
			viewProperties = append(viewProperties, props...)
		}
	default:
		// manual view
		props, err := dataview.GetViewByName(view)
		if err != nil || len(props) == 0 {
			if err != nil {
				cfg.Logger.Warnf("no matching view found. %s, name=%s", err, view)
			} else {
				cfg.Logger.Warnf("no matching view found. name=%s", view)
			}
			viewProperties = append(viewProperties, "**")
		} else {
			cfg.Logger.Infof("Detected view: %s", strings.Join(props, ", "))
			viewProperties = append(viewProperties, props...)
		}
	}
	return viewProperties, nil
}

// WriteJSONToConsole writes given json output to the console supporting the common options of select, output etc.
func (f *Factory) WriteJSONToConsole(cfg *config.Config, cmd *cobra.Command, property string, output []byte) error {
	consol, err := f.Console()
	if err != nil {
		return err
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	if len(commonOptions.Filters.Pluck) == 0 {
		// don't fail if view properties fail
		props, _ := f.GetViewProperties(cfg, cmd, output)
		if len(props) > 0 {
			commonOptions.Filters.Pluck = props
		}
	}
	output, filterErr := commonOptions.Filters.Apply(string(output), property, false, consol.SetHeaderFromInput)
	if filterErr != nil {
		return filterErr
	}

	output = bytes.ReplaceAll(output, []byte("\\u003c"), []byte("<"))
	output = bytes.ReplaceAll(output, []byte("\\u003e"), []byte(">"))
	output = bytes.ReplaceAll(output, []byte("\\u0026"), []byte("&"))

	jsonformatter.WithOutputFormatters(
		consol,
		output,
		false,
		jsonformatter.WithFileOutput(commonOptions.OutputFile != "", commonOptions.OutputFile, false),
		jsonformatter.WithTrimSpace(true),
		jsonformatter.WithJSONStreamOutput(true, consol.IsJSONStream(), consol.IsCSV()),
		jsonformatter.WithSuffix(len(output) > 0, "\n"),
	)
	return nil
}

func (f *Factory) CheckPostCommandError(err error) error {
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

	// cfg must be
	if cfg != nil && cfg.WithError() {
		w = f.IOStreams.Out
	}

	// consol, consolErr := f.Console()
	// if consolErr != nil {
	// 	log.Fatalf("Could not configure console. %s", consolErr)
	// }

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

	outErr := err
	if cfg != nil && cfg.GetSilentExit() {
		outErr = nil
	}
	printLogEntries := !cfg.ShowProgress()

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
			if printLogEntries {
				logg.Errorf("%s", cErr)
			}
			fmt.Fprintf(w, "%s\n", cErr.JSONString())

			cErr.Processed = true
			outErr = cErr
		}
	} else {
		// unexpected error
		cErr := cmderrors.NewSystemErrorF("%s", err)
		cErr.ExitCode = cmderrors.ExitUserError
		if printLogEntries {
			logg.Errorf("%s", cErr)
		}
		logg.Debugf("Processing unexpected error. %s, exitCode=%d", err, cErr.ExitCode)
		fmt.Fprintf(w, "%s\n", cErr.JSONString())
		cErr.Processed = true
		outErr = cErr
	}

	return outErr
}

// NewRequestInputIterators create a request iterator based on pipe line configuration
func NewRequestInputIterators(cmd *cobra.Command, cfg *config.Config) (*flags.RequestInputIterators, error) {
	pipeOpts, err := flags.GetPipeOptionsFromAnnotation(cmd)

	if cfg != nil {
		pipeOpts.Disabled = cfg.DisableStdin()
		pipeOpts.EmptyPipe = cfg.AllowEmptyPipe()
	}
	inputIter := &flags.RequestInputIterators{
		PipeOptions: pipeOpts,
	}
	return inputIter, err
}

func (f *Factory) NewProgress(cmd *cobra.Command, cfg *config.Config) (*flags.RequestInputIterators, error) {

	pipeOpts, err := flags.GetPipeOptionsFromAnnotation(cmd)

	if cfg != nil {
		pipeOpts.Disabled = cfg.DisableStdin()
		pipeOpts.EmptyPipe = cfg.AllowEmptyPipe()
	}
	inputIter := &flags.RequestInputIterators{
		PipeOptions: pipeOpts,
	}
	return inputIter, err
}
