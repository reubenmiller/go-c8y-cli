package cmdutil

import (
	"strings"

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
	"github.com/tidwall/gjson"
)

type Factory struct {
	IOStreams      *iostreams.IOStreams
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
	case config.ViewsNone:
		// dont apply a view
		return []string{"**"}, nil
	case config.ViewsAll:
		jsonResponse := gjson.ParseBytes(output)
		props, err := dataview.GetView(&jsonResponse, "")

		if err != nil || len(props) == 0 {
			if err != nil {
				log.Warnf("Failed to detect view. defaulting to '**'. %s", err)
			} else {
				log.Warn("Failed to detect view. defaulting to '**'")
			}
			viewProperties = append(viewProperties, "**")
		} else {
			log.Infof("Detected view: %s", strings.Join(props, ", "))
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
