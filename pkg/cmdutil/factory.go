package cmdutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mode"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
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

	// Extension
	ExtensionManager func() extensions.ExtensionManager

	// Executable is the path to the currently invoked binary
	Executable string

	// Command
	Command *cobra.Command
}

// Set reference to the cobra command
func (f *Factory) SetCommand(cmd *cobra.Command) *Factory {
	f.Command = cmd
	return f
}

// CreateModeEnabled create mode is enabled
func (f *Factory) CreateModeEnabled() error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	if err := cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	); err != nil {
		return err
	}

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

	if err := cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	); err != nil {
		return err
	}

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

	if err := cfg.WithOptions(
		config.WithBindEnv(config.SettingsDryRun, false),
	); err != nil {
		return err
	}

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

func (f *Factory) RunWithGenericWorkers(cmd *cobra.Command, inputIterators *flags.RequestInputIterators, iter iterator.Iterator, runFunc worker.Runner) error {
	client, err := f.Client()
	if err != nil {
		return err
	}
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
	// consol, err := f.Console()
	// if err != nil {
	// 	return err
	// }
	// dataview, err := f.DataView()
	// if err != nil {
	// 	return err
	// }

	w, err := worker.NewGenericWorker(log, cfg, f.IOStreams, client, activityLogger, runFunc, f.CheckPostCommandError)

	if err != nil {
		return err
	}
	return w.Run(cmd, iter, inputIterators)
}

func (f *Factory) RunSequentiallyWithGenericWorkers(cmd *cobra.Command, iter iterator.Iterator, runFunc worker.Runner, inputIterators *flags.RequestInputIterators) error {
	client, err := f.Client()
	if err != nil {
		return err
	}
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

	w, err := worker.NewGenericWorker(log, cfg, f.IOStreams, client, activityLogger, runFunc, f.CheckPostCommandError)

	if err != nil {
		return err
	}
	return w.RunSequentially(cmd, iter, inputIterators)
}

// GetViewProperties Look up the view properties to display
func (f *Factory) GetViewProperties(cfg *config.Config, cmd *cobra.Command, output []byte) ([]string, error) {
	dataView, err := f.DataView()
	if err != nil {
		return nil, err
	}
	log, err := f.Logger()
	if err != nil {
		return nil, err
	}

	view := cfg.ViewOption()
	showRaw := cfg.RawOutput() || cfg.WithTotalPages() || cfg.WithTotalElements()

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
		props, err := dataView.GetView(&dataview.ViewData{
			ResponseBody: &jsonResponse,
		})

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
		props, err := dataView.GetViewByName(view)
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

func (f *Factory) CheckPostCommandError(err error) error {
	cfg, configErr := f.Config()
	if configErr != nil {
		log.Fatalf("Could not load configuration. %s", configErr)
	}
	logg, logErr := f.Logger()
	if logErr != nil {
		log.Fatalf("Could not configure logger. %s", logErr)
	}
	w := io.Discard

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

func (f *Factory) ResolveTemplates(pattern string, withFullPath bool) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	paths := cfg.GetTemplatePaths()

	allMatches := []string{}

	// Filter
	matches, err := pathresolver.ResolvePaths(paths, "*", []string{".jsonnet"}, "ignore")
	if err != nil {
		return []string{"jsonnet"}, err
	}

	// Apply full matches
	for _, m := range matches {
		option := filepath.Base(m)
		if matched, _ := filepath.Match(pattern, option); matched {
			if withFullPath {
				allMatches = append(allMatches, m)
			} else {
				allMatches = append(allMatches, option)
			}
		}
	}

	// Extensions
	for _, ext := range f.ExtensionManager().List() {
		extTemplatePath := ext.TemplatePath()
		if extTemplatePath == "" {
			continue
		}
		matches, err := pathresolver.ResolvePaths([]string{extTemplatePath}, "*", []string{".jsonnet"}, "ignore")
		if err != nil {
			return []string{"jsonnet"}, err
		}

		// Apply full matches
		for _, m := range matches {
			option := fmt.Sprintf("%s::%s", ext.Name(), filepath.Base(m))
			if matched, _ := filepath.Match(pattern, option); matched {
				if withFullPath {
					allMatches = append(allMatches, m)
				} else {
					allMatches = append(allMatches, option)
				}
			}
		}

	}

	return allMatches, nil
}

func (f *Factory) GetTenant() string {
	client, err := f.Client()
	if err != nil {
		return ""
	}
	return client.TenantName
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

type OutputContext struct {
	Input    any
	Response *http.Response
	Duration time.Duration
}

func (f *Factory) ExecuteOutputTemplate(output []byte, params OutputContext, commonOptions *config.CommonCommandOptions) ([]byte, error) {
	if commonOptions.OutputTemplate == "" {
		return output, nil
	}

	outputBuilder := mapbuilder.NewInitializedMapBuilder(true)

	if err := outputBuilder.AddLocalTemplateVariable("flags", commonOptions.CommandFlags); err != nil {
		return nil, err
	}

	requestData := make(map[string]interface{})
	responseData := make(map[string]interface{})

	// Add request/response variables
	if params.Response != nil {
		resp := params.Response
		requestData["path"] = resp.Request.URL.Path
		requestData["pathEncoded"] = strings.Replace(resp.Request.URL.String(), resp.Request.URL.Scheme+"://"+resp.Request.URL.Host, "", 1)
		requestData["host"] = resp.Request.URL.Host
		requestData["url"] = resp.Request.URL.String()
		requestData["query"] = request.TryUnescapeURL(resp.Request.URL.RawQuery)
		requestData["queryParams"] = request.FlattenArrayMap(resp.Request.URL.Query())
		requestData["method"] = resp.Request.Method
		// requestData["header"] = resp.Response.Request.Header

		// TODO: Add a response variable to included the status code, content type,
		responseData["statusCode"] = resp.StatusCode
		responseData["status"] = resp.Status
		responseData["duration"] = params.Duration.Milliseconds()
		responseData["contentLength"] = resp.ContentLength
		responseData["contentType"] = resp.Header.Get("Content-Type")
		responseData["header"] = request.FlattenArrayMap(resp.Header)
		responseData["proto"] = resp.Proto
		responseData["body"] = string(output)
	}

	if err := outputBuilder.AddLocalTemplateVariable("request", requestData); err != nil {
		return nil, err
	}

	if err := outputBuilder.AddLocalTemplateVariable("response", responseData); err != nil {
		return nil, err
	}

	outputJSON := make(map[string]any)
	if parseErr := jsonUtilities.ParseJSON(string(output), outputJSON); parseErr == nil {
		if err := outputBuilder.AddLocalTemplateVariable("output", outputJSON); err != nil {
			return nil, err
		}
	} else {
		if err := outputBuilder.AddLocalTemplateVariable("output", string(output)); err != nil {
			return nil, err
		}
	}

	outputBuilder.AppendTemplate(commonOptions.OutputTemplate)
	out, outErr := outputBuilder.MarshalJSONWithInput(params.Input)

	if outErr != nil {
		return out, outErr
	}
	return out, nil
}

func (f *Factory) WriteOutputWithoutPropertyGuess(output []byte, params OutputContext) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	commonOptions, err := cfg.GetOutputCommonOptions(f.Command)
	if err != nil {
		return err
	}

	_, err = f.WriteOutputWithRows(output, params, commonOptions.DisableResultPropertyDetection())
	return err
}

func (f *Factory) WriteOutput(output []byte, params OutputContext, commonOptions *config.CommonCommandOptions) error {
	_, err := f.WriteOutputWithRows(output, params, commonOptions)
	return err
}

func (f *Factory) WriteOutputWithRows(output []byte, params OutputContext, commonOptions *config.CommonCommandOptions) (int, error) {
	consol, err := f.Console()
	if err != nil {
		return 0, err
	}

	cfg, err := f.Config()
	if err != nil {
		return 0, err
	}

	dataView, err := f.DataView()
	if err != nil {
		return 0, err
	}

	logg, err := f.Logger()
	if err != nil {
		return 0, err
	}

	if commonOptions == nil {
		if f.Command == nil {
			return 0, fmt.Errorf("command output options are mandatory")
		}
		commonOptions = cfg.MustGetOutputCommonOptions(f.Command)
	}

	unfilteredSize := 0
	outputJSON := gjson.ParseBytes(output)

	if len(output) > 0 || commonOptions.HasOutputTemplate() {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		if params.Response != nil {
			PrintResponseSize(logg, params.Response, output)
		}

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON(output)

		dataProperty := ""
		showRaw := cfg.RawOutput() || cfg.WithTotalPages() || cfg.WithTotalElements()

		dataProperty = commonOptions.ResultProperty
		if dataProperty == "" {
			dataProperty = f.GuessDataProperty(outputJSON)
		} else if dataProperty == "-" {
			dataProperty = ""
		}

		if v := outputJSON.Get(dataProperty); v.Exists() && v.IsArray() {
			unfilteredSize = len(v.Array())
			logg.Infof("Unfiltered array size. len=%d", unfilteredSize)
		}

		// Apply output template (before the data is processed as the template can transform text to json or other way around)
		if commonOptions.HasOutputTemplate() {
			var tempBody []byte
			if showRaw || dataProperty == "" {
				tempBody = output
			} else {
				tempBody = []byte(outputJSON.Get(dataProperty).Raw)
			}
			dataProperty = ""

			tmplOutput, tmplErr := f.ExecuteOutputTemplate(tempBody, params, commonOptions)
			if tmplErr != nil {
				return unfilteredSize, tmplErr
			}

			if jsonUtilities.IsValidJSON(tmplOutput) {
				isJSONResponse = true
				output = pretty.Ugly(tmplOutput)
				outputJSON = gjson.ParseBytes(output)
			} else {
				isJSONResponse = false
				// TODO: Is removing the quotes doing too much, what happens if someone is building csv, and it using quotes around some fields?
				// e.g. `"my value",100`, that would get transformed to `my value",100`
				// Trim any quotes wrapping the values
				tmplOutput = bytes.TrimSpace(tmplOutput)

				output = pretty.Ugly(bytes.Trim(tmplOutput, "\""))
				outputJSON = gjson.ParseBytes([]byte(""))
			}
		}

		if isJSONResponse && commonOptions.Filters != nil {
			if showRaw {
				dataProperty = ""
			}

			if cfg.RawOutput() {
				logg.Infof("Raw mode active. In raw mode the following settings are forced, view=off, output=json")
			}
			view := cfg.ViewOption()
			logg.Infof("View mode: %s", view)

			// Detect view (if no filters are given)
			if len(commonOptions.Filters.Pluck) == 0 {
				if len(output) > 0 && dataView != nil {
					inputData := outputJSON
					if dataProperty != "" {
						inputData = outputJSON.Get(dataProperty)
					}

					switch strings.ToLower(view) {
					case config.ViewsOff:
						// dont apply a view
						if !showRaw {
							commonOptions.Filters.Pluck = []string{"**"}
						}
					case config.ViewsAuto:
						viewData := &dataview.ViewData{
							ResponseBody: &inputData,
						}

						if params.Response != nil {
							viewData.ContentType = params.Response.Header.Get("Content-Type")
							viewData.Request = params.Response.Request
						}

						props, err := dataView.GetView(viewData)

						if err != nil || len(props) == 0 {
							if err != nil {
								logg.Infof("No matching view detected. defaulting to '**'. %s", err)
							} else {
								logg.Info("No matching view detected. defaulting to '**'")
							}
							commonOptions.Filters.Pluck = []string{"**"}
						} else {
							logg.Infof("Detected view: %s", strings.Join(props, ", "))
							commonOptions.Filters.Pluck = props
						}
					default:
						props, err := dataView.GetViewByName(view)
						if err != nil || len(props) == 0 {
							if err != nil {
								logg.Warnf("no matching view found. %s, name=%s", err, view)
							} else {
								logg.Warnf("no matching view found. name=%s", view)
							}
							commonOptions.Filters.Pluck = []string{"**"}
						} else {
							logg.Infof("Detected view: %s", strings.Join(props, ", "))
							commonOptions.Filters.Pluck = props
						}
					}
				}
			} else {
				logg.Debugf("using existing pluck values. %v", commonOptions.Filters.Pluck)
			}

			if filterOutput, filterErr := commonOptions.Filters.Apply(string(output), dataProperty, false, consol.SetHeaderFromInput); filterErr != nil {
				logg.Warnf("filter error. %s", filterErr)
				responseText = filterOutput
			} else {
				responseText = filterOutput
			}

			emptyArray := []byte("[]\n")

			if !showRaw {
				if len(responseText) == len(emptyArray) && bytes.Equal(responseText, emptyArray) {
					logg.Info("No matching results found. Empty response will be omitted")
					responseText = []byte{}
				}
			}

		} else {
			responseText = output
		}

		// replace special escaped unicode sequences
		responseText = bytes.ReplaceAll(responseText, []byte("\\u003c"), []byte("<"))
		responseText = bytes.ReplaceAll(responseText, []byte("\\u003e"), []byte(">"))
		responseText = bytes.ReplaceAll(responseText, []byte("\\u0026"), []byte("&"))

		// Wait for progress bar to finish before printing to console
		// to prevent overriding the output
		f.IOStreams.WaitForProgressIndicator()

		jsonformatter.WithOutputFormatters(
			consol,
			responseText,
			!isJSONResponse,
			jsonformatter.WithFileOutput(commonOptions.OutputFile != "", commonOptions.OutputFile, false),
			jsonformatter.WithTrimSpace(true),
			jsonformatter.WithJSONStreamOutput(isJSONResponse, consol.IsJSONStream(), consol.IsTextOutput()),
			jsonformatter.WithSuffix(len(responseText) > 0, "\n"),
		)
	}
	return unfilteredSize, nil
}

func (f *Factory) GuessDataProperty(output gjson.Result) string {
	property := ""
	arrayPropertes := []string{}
	totalKeys := 0

	logg, err := f.Logger()
	if err != nil {
		panic(err)
	}

	if v := output.Get("id"); !v.Exists() {
		// Find the property which is an array
		output.ForEach(func(key, value gjson.Result) bool {
			totalKeys++
			if value.IsArray() {
				arrayPropertes = append(arrayPropertes, key.String())
			}
			return true
		})
	}

	if len(arrayPropertes) > 1 {
		logg.Debugf("Could not detect property as more than 1 array like property detected: %v", arrayPropertes)
		return ""
	}
	logg.Debugf("Array properties: %v", arrayPropertes)

	if len(arrayPropertes) == 0 {
		return ""
	}

	property = arrayPropertes[0]

	// if total keys is a high number, than it is most likely not an array of data
	// i.e. for the /tenant/statistics
	if property != "" && totalKeys > 10 {
		return ""
	}

	if property != "" && totalKeys < 10 {
		logg.Debugf("Data property: %s", property)
	}
	return property
}

func PrintResponseSize(l *logger.Logger, resp *http.Response, output []byte) {
	if resp.ContentLength > -1 {
		l.Infof("Response Length: %0.1fKB", float64(resp.ContentLength)/1024)
	} else {
		if resp.Uncompressed {
			l.Infof("Response Length: %0.1fKB (uncompressed)", float64(len(output))/1024)
		} else {
			l.Infof("Response Length: %0.1fKB", float64(len(output))/1024)
		}
	}
}
