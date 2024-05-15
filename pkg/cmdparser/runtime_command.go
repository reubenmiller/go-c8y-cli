package cmdparser

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/cli/safeexec"
	"github.com/kballard/go-shellquote"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/run"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type RuntimeCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	options *CmdOptions
}

// NewRuntimeCmd creates a command which is created at runtime
func NewRuntimeCmd(f *cmdutil.Factory, options *CmdOptions) *RuntimeCmd {
	ccmd := &RuntimeCmd{
		factory: f,
		options: options,
	}
	cmd := options.Command
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Mode checks
		switch options.Spec.GetMethod() {
		case "POST":
			return f.CreateModeEnabled()
		case "PUT":
			return f.UpdateModeEnabled()
		case "DELETE":
			return f.DeleteModeEnabled()
		}
		return nil
	}
	cmd.RunE = ccmd.RunE
	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		options.Completion...,
	)

	flags.WithOptions(
		cmd,
		options.Runtime...,
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)
	return ccmd
}

func (n *RuntimeCmd) Prepare(args []string) error {
	item := n.options.Spec
	subcmd := n.options
	factory := n.factory

	cfg, err := factory.Config()
	if err != nil {
		return err
	}

	// Presets
	if subcmd.Spec.HasPreset() {
		var values *[]flags.GetOption
		switch subcmd.Spec.Preset.Type {
		case PresetGetIdentity:
			values = &subcmd.QueryParameter
		}
		if values != nil {
			for _, p := range subcmd.Spec.Preset.Extensions {
				*values = append(*values, GetOption(subcmd, &p, factory, args)...)
			}
		}
	}

	// path
	for _, p := range item.PathParameters {
		subcmd.Path.Options = append(subcmd.Path.Options, GetOption(subcmd, &p, factory, args)...)
	}
	subcmd.Path.Template = item.Path

	// header
	subcmd.Header = append(subcmd.Header, flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"))
	for _, p := range item.HeaderParameters {
		subcmd.Header = append(subcmd.Header, GetOption(subcmd, &p, factory, args)...)
	}

	if subcmd.Spec.ContentType != "" {
		subcmd.Header = append(subcmd.Header, flags.WithStaticStringValue("Content-Type", subcmd.Spec.ContentType))
	}

	if subcmd.Spec.Accept != "" {
		subcmd.Header = append(subcmd.Header, flags.WithStaticStringValue("Accept", subcmd.Spec.Accept))
	}

	if item.SupportsProcessingMode() {
		subcmd.Header = append(subcmd.Header, flags.WithProcessingModeValue())
	}

	// query
	subcmd.QueryParameter = append(subcmd.QueryParameter, flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"))

	for _, p := range item.QueryParameters {
		subcmd.QueryParameter = append(subcmd.QueryParameter, GetOption(subcmd, &p, factory, args)...)

		// Support Cumulocity Query builder
		if len(p.Children) > 0 {
			queryOptions := []flags.GetOption{}
			for _, child := range p.Children {
				// Ignore special in-built values as these are handled separately
				if child.Name == "queryTemplate" || child.Name == "orderBy" {
					continue
				}
				queryOptions = append(queryOptions, GetOption(subcmd, &child, factory, args)...)
			}

			if subcmd.Spec.HasPreset() {
				switch subcmd.Spec.Preset.Type {
				case PresetQueryInventory:
					for _, p := range subcmd.Spec.Preset.Extensions {
						queryOptions = append(queryOptions, GetOption(subcmd, &p, factory, args)...)
					}
				case PresetQueryInventoryChildren:
					for _, p := range subcmd.Spec.Preset.Extensions {
						queryOptions = append(queryOptions, GetOption(subcmd, &p, factory, args)...)
					}
				}
			}

			subcmd.QueryParameter = append(subcmd.QueryParameter, flags.WithCumulocityQuery(queryOptions, p.GetTargetProperty()))
		}
	}

	// body
	requiredBodyKeys := []string{}
	requiredBodyKeys = append(requiredBodyKeys, item.BodyRequiredKeys...)
	for _, p := range item.Body {
		if p.IsRequired() {
			requiredBodyKeys = append(requiredBodyKeys, p.GetTargetProperty())
		}
	}
	if len(requiredBodyKeys) > 0 {
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithRequiredProperties(requiredBodyKeys...))
	}

	if len(item.Body) > 0 {
		if item.Method == "PUT" || item.Method == "POST" {
			subcmd.Body.Initialize = true
		}
	}

	supportsFormData := false
	switch item.GetBodyContentType() {
	case "binary":
		subcmd.Body.IsBinary = true
		supportsFormData = true
	case "formdata":
		supportsFormData = true
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	case "jsonarray":
		subcmd.Body.DefaultValue = []byte("[]")
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	case "jsonobject":
		subcmd.Body.DefaultValue = []byte("{}")
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	default:
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	}
	for _, p := range item.Body {
		switch p.Type {
		case "file", "attachment":
			subcmd.Body.UploadProgressSource = p.Name
			subcmd.FormData = append(subcmd.FormData, GetOption(subcmd, &p, factory, args)...)
		case "fileContents":
			subcmd.Body.UploadProgressSource = p.Name
			fallthrough
		default:
			if supportsFormData {
				subcmd.FormData = append(subcmd.FormData, GetOption(subcmd, &p, factory, args)...)
			} else {
				subcmd.Body.Options = append(subcmd.Body.Options, GetOption(subcmd, &p, factory, args)...)
			}
		}
	}

	subcmd.Body.Options = append(subcmd.Body.Options, cmdutil.WithTemplateValue(factory))
	subcmd.Body.Options = append(subcmd.Body.Options, flags.WithTemplateVariablesValue())

	for _, bodyTemplate := range item.BodyTemplates {
		if bodyTemplate.Type == "jsonnet" {
			if bodyTemplate.ApplyLast {
				subcmd.Body.Options = append(subcmd.Body.Options, flags.WithRequiredTemplateString(bodyTemplate.Template))
			} else {
				subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDefaultTemplateString(bodyTemplate.Template))
			}
		}
	}

	return nil
}

func (n *RuntimeCmd) RunE(cmd *cobra.Command, args []string) error {
	switch n.options.Spec.CommandType() {
	case models.CommandTypeCommand:
		return n.ExecuteCommand(cmd, args)
	default:
		return n.ExecuteAPI(cmd, args)
	}
}

// RunE executes the command
func (n *RuntimeCmd) ExecuteAPI(cmd *cobra.Command, args []string) error {
	if err := n.Prepare(args); err != nil {
		return err
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		n.options.QueryParameter...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParametersWithMapping(query, n.options.Spec.FlagMapping)

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		n.options.Header...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
		n.options.FormData...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	var body *mapbuilder.MapBuilder
	if len(n.options.Body.DefaultValue) > 0 {
		body = mapbuilder.NewMapBuilderWithInit(n.options.Body.DefaultValue)
	} else {
		body = mapbuilder.NewInitializedMapBuilder(n.options.Body.Initialize)
	}
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		n.options.Body.Options...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate(n.options.Path.Template)
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		n.options.Path.Options...,
	)
	if err != nil {
		return err
	}

	var req *c8y.RequestOptions
	if n.options.Body.IsBinary {
		req = &c8y.RequestOptions{
			Method:       n.options.Spec.Method,
			Path:         path.GetTemplate(),
			Query:        queryValue,
			Body:         body.GetFileContents(),
			FormData:     formData,
			Header:       headers,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}
	} else {
		req = &c8y.RequestOptions{
			Method:       n.options.Spec.Method,
			Path:         path.GetTemplate(),
			Query:        queryValue,
			Body:         body,
			FormData:     formData,
			Header:       headers,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}
	}

	// add upload progress bar
	if n.options.Body.UploadProgressSource != "" {
		req.PrepareRequest = c8ybinary.AddProgress(
			cmd,
			n.options.Body.UploadProgressSource,
			cfg.GetProgressBar(n.factory.IOStreams.ErrOut, n.factory.IOStreams.IsStderrTTY()),
		)
	}

	return n.factory.RunWithWorkers(client, cmd, req, inputIterators)
}

func (n *RuntimeCmd) ExecuteCommand(cmd *cobra.Command, args []string) error {
	// Ignore arguments provided by the user after a "--"
	// e.g. c8y myext cli do_something -- ls -la
	// everything after the " -- " should not be included in args
	// otherwise this will affect the pipeline processing
	switch i := cmd.Flags().ArgsLenAtDash(); i {
	case -1:
		// do nothing
	case 0:
		args = []string{}
	default:
		args = args[0:i]
	}

	if err := n.Prepare(args); err != nil {
		return err
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	console, err := n.factory.Console()
	if err != nil {
		return err
	}

	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// body
	var body *mapbuilder.MapBuilder
	if len(n.options.Body.DefaultValue) > 0 {
		body = mapbuilder.NewMapBuilderWithInit(n.options.Body.DefaultValue)
	} else {
		body = mapbuilder.NewInitializedMapBuilder(n.options.Body.Initialize)
	}
	err = flags.WithCLI(
		cmd,
		body,
		inputIterators,
		n.options.Body.Options...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	body.SetApplyTemplateOnMarshalPreference(true)

	if len(n.options.Spec.Command) == 0 {
		return fmt.Errorf("invalid spec. requires at least one command to be defined")
	}

	var iter iterator.Iterator
	if inputIterators.Total > 0 {
		iter = mapbuilder.NewMapBuilderIterator(body)
	} else {
		iter = iterator.NewBoundIterator(mapbuilder.NewMapBuilderIterator(body), 1)
	}

	return n.factory.RunWithGenericWorkers(cmd, inputIterators, iter, func(j worker.Job) (any, error) {
		log.Warnf("Doing work. %s", j.Value)

		var b []byte
		switch v := j.Value.(type) {
		case []byte:
			b = v
		}

		if len(b) == 0 {
			return nil, io.EOF
		}

		_, cmdArgs, err := n.PrepareRuntimeCommand(b)
		if err != nil {
			return nil, IgnoreEOF(err)
		}

		if len(n.options.Spec.Command) == 0 {
			return nil, fmt.Errorf("invalid spec. requires at least one command to be defined")
		}

		exe, err := safeexec.LookPath(n.options.Spec.Command[0])
		if err != nil {
			return nil, err
		}

		totalArgs := make([]string, 0)
		if len(n.options.Spec.Command) > 1 {
			totalArgs = append(totalArgs, n.options.Spec.Command[1:]...)
		}
		// TODO: the api spec should control if the args should be quoted or not
		// TODO: allow users to control where this value is added
		shouldEscape := true
		if shouldEscape {
			// totalArgs = append(totalArgs, shellquote.Join(cmdArgs...))
			totalArgs = append(totalArgs, strings.Join(cmdArgs, " "))
		} else {
			totalArgs = append(totalArgs, cmdArgs...)
		}

		if argsAtDash := cmd.ArgsLenAtDash(); argsAtDash > -1 {
			totalArgs = append(totalArgs, "--")
			totalArgs = append(totalArgs, cmd.Flags().Args()[argsAtDash:]...)
		}

		if cfg.ShouldUseDryRun(cmd.CommandPath()) {
			if shouldEscape {
				customArgs := make([]string, 0, len(totalArgs))
				for _, v := range totalArgs {
					if strings.Contains(v, "'") {
						customArgs = append(customArgs, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "\\'")))
					} else {
						customArgs = append(customArgs, v)
					}
				}
				// customArgs := strings.Join(totalArgs, " ")
				fmt.Fprintf(n.factory.IOStreams.Out, "DRY: Executing command: %s %s\n", shellquote.Join(n.options.Spec.Command[0]), strings.Join(customArgs, " "))

				cmdstr := shellescape.QuoteCommand(totalArgs)
				fmt.Fprintf(n.factory.IOStreams.Out, "DRY: Executing command: %s %s\n", shellquote.Join(n.options.Spec.Command[0]), cmdstr)
			} else {
				fmt.Fprintf(n.factory.IOStreams.Out, "DRY: Executing command: %s %s\n", shellquote.Join(n.options.Spec.Command[0]), shellquote.Join(totalArgs...))

				cmdstr := shellescape.QuoteCommand(cmdArgs)
				fmt.Fprintf(n.factory.IOStreams.Out, "DRY: Executing command: %s %s\n", shellquote.Join(n.options.Spec.Command[0]), cmdstr)
			}
			return nil, nil
		}

		log.Info("Executing command: %s %s", shellquote.Join(n.options.Spec.Command[0]), shellquote.Join(totalArgs...))

		externalCmd := exec.Command(exe, totalArgs...)
		// TODO: Only map stdout/stderr to command when not running in the background
		var outb, errb bytes.Buffer

		// TODO: Must be interactive to have a session
		isInteractive := n.factory.IOStreams.IsStdinTTY()

		if isInteractive {
			log.Warn("Using interactive console")
			externalCmd.Stderr = os.Stderr
			externalCmd.Stdout = os.Stdout
			externalCmd.Stdin = os.Stdin
		} else {
			externalCmd.Stdout = &outb
			// Option 1: Combine output by reusing the same stdout buffer
			externalCmd.Stderr = &outb
			// externalCmd.Stderr = &errb
			externalCmd.Stdin = nil
		}
		preparedCmd := run.PrepareCmd(externalCmd)

		err = preparedCmd.Run()

		if !isInteractive {
			err := n.factory.WriteOutput(outb.Bytes(), cmdutil.OutputContext{
				Input: j.Input,
			}, &j.CommonOptions)

			if err != nil {
				return nil, err
			}

			// output, err := n.factory.ExecuteOutputTemplate(outb.Bytes(), cmdutil.OutputContext{
			// 	Input: j.Input,
			// }, j.CommonOptions)

			// if err != nil {
			// 	return nil, err
			// }

			// if _, writeErr := console.Write(output); writeErr != nil {
			// 	return nil, writeErr
			// }
			if _, writeErr := console.Write(errb.Bytes()); writeErr != nil {
				return nil, writeErr
			}
		}

		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				return nil, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCode(ee.ExitCode()), ee)
			}

			return nil, err
		}

		return nil, nil
	})

}

func RunCommand() func(j worker.Job) (any, error) {
	return func(j worker.Job) (any, error) {
		return nil, nil
	}
}

func IgnoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

func (n *RuntimeCmd) PrepareRuntimeCommand(b []byte) (path string, args []string, err error) {
	data := gjson.ParseBytes(b)

	if v := data.Get("command.path"); v.Exists() {
		path = v.String()
	}

	args = make([]string, 0)
	if v := data.Get("command.args"); v.Exists() && v.IsArray() {
		v.ForEach(func(i, item gjson.Result) bool {
			if item.IsObject() {
				item.ForEach(func(cmdKey, cmdValue gjson.Result) bool {
					switch cmdValue.Type {
					case gjson.String:
						args = append(args, cmdKey.String())
						args = append(args, cmdValue.String())
					case gjson.Number:
						args = append(args, cmdValue.Str)
					case gjson.True:
						args = append(args, cmdKey.String())
					case gjson.False:
						args = append(args, fmt.Sprintf("%s=%v", cmdKey.String(), cmdKey.Bool()))
					case gjson.Null:
						// ignore null values
					default:
						// Ignore array and object types
					}
					return true
				})
			} else if item.Type == gjson.String {
				args = append(args, item.String())
			}
			return true
		})
	}
	return
}
