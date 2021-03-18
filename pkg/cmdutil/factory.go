package cmdutil

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/mode"
	"github.com/reubenmiller/go-c8y-cli/pkg/request"
	"github.com/reubenmiller/go-c8y-cli/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type Factory struct {
	IOStreams      *iostreams.IOStreams
	Client         func() (*c8y.Client, error)
	Config         func() (*config.Config, error)
	Logger         func() (*logger.Logger, error)
	ActivityLogger func() (*activitylogger.ActivityLogger, error)
	Console        func() (*console.Console, error)
	DataView       func() (*dataview.DataView, error)

	// Executable is the path to the currently invoked binary
	Executable string
}

// CreateModeEnabled create mode is enabled
func (f *Factory) CreateModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	return mode.ValidateCreateMode(cfg)
}

// ValidateUpdateMode update mode is enabled
func (f *Factory) UpdateModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	return mode.ValidateUpdateMode(cfg)
}

// ValidateDeleteMode delete mode is enabled
func (f *Factory) DeleteModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
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
		Client:         client,
		Config:         cfg,
		Logger:         log,
		DataView:       dataview,
		Console:        consol,
		ActivityLogger: activityLogger,
		HideSensitive:  config.HideSensitiveInformationIfActive,
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
		Client:         client,
		Config:         cfg,
		Logger:         log,
		DataView:       dataview,
		Console:        consol,
		ActivityLogger: activityLogger,
		HideSensitive:  config.HideSensitiveInformationIfActive,
	}
	w, err := worker.NewWorker(log, cfg, f.IOStreams, client, activityLogger, handler.ProcessRequestAndResponse)

	if err != nil {
		return err
	}
	return w.ProcessRequestAndResponse(cmd, req, inputIterators)
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
	output = commonOptions.Filters.Apply(string(output), property, false, consol.SetHeaderFromInput)

	jsonformatter.WithOutputFormatters(
		consol,
		output,
		false,
		jsonformatter.WithTrimSpace(true),
		jsonformatter.WithJSONStreamOutput(true, consol.IsJSONStream(), consol.IsCSV()),
		jsonformatter.WithSuffix(len(output) > 0, "\n"),
	)
	return nil
}
