package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type ExtensionReference struct {
	ContextPath string
	Version     string
	Modules     []string
}

func buildUIRemotes(out map[string]ExtensionReference, ext gjson.Result, versionOrTag string) error {

	// Find version (to verify it exists)
	matchVersion := ""

	if versionOrTag == "" {
		versionOrTag = "latest"
	}

	// Check if version exists
	ext.Get("applicationVersions").ForEach(func(key, value gjson.Result) bool {
		// Match by Version
		if iVersion := value.Get("version"); iVersion.String() == versionOrTag {
			matchVersion = iVersion.String()
			return false
		}

		// Match by Tag
		tags := value.Get("tags")
		if tags.IsArray() {
			tags.ForEach(func(i, tag gjson.Result) bool {
				if tag.String() == versionOrTag {
					matchVersion = value.Get("version").String()
					return false
				}
				return true
			})
		}

		return true
	})

	if matchVersion == "" {
		return fmt.Errorf("not match version found in extension. extension=%s, version=%s", ext.Get("name").String(), versionOrTag)
	}

	remote := ExtensionReference{
		ContextPath: ext.Get("contextPath").String(),
		Version:     matchVersion,
		Modules:     []string{},
	}

	remotes := ext.Get("manifest.remotes")
	if remotes.IsObject() {
		ext.Get("manifest.remotes").ForEach(func(key, value gjson.Result) bool {
			if value.IsArray() {
				for _, v := range value.Array() {
					if moduleName := v.String(); moduleName != "" {
						remote.Modules = append(remote.Modules, moduleName)
					}
				}
			}
			return true
		})
	}

	out[remote.ContextPath] = remote
	return nil
}

func formatApplicationRemotes(extensions map[string]ExtensionReference) map[string][]string {
	out := make(map[string][]string)
	for _, ext := range extensions {
		out[fmt.Sprintf("%s@%s", ext.ContextPath, ext.Version)] = ext.Modules
	}
	return out
}

type CmdInstall struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	application string
	extension   []string
	replaceAll  bool
	updateAll   bool
}

func NewCmdInstall(f *cmdutil.Factory) *CmdInstall {
	ccmd := &CmdInstall{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install/update UI extensions to an application",
		Long: heredoc.Doc(`
			Install or update UI extensions assigned to an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui extensions install --application devicemanagement --extension myext@latest --extension someother@1.2.3
			Install multiple UI extensions to the devicemanagement application

			$ c8y ui extensions install --application devicemanagement --extension myext@latest --replace
			Replace existing UI extensions with the extensions provided

			$ c8y ui extensions install --application devicemanagement --replace
			Remove all UI extension from an application (and adding none)

			$ c8y ui extensions install --application devicemanagement --update-versions
			Update to the latest versions for all of the existing UI extensions that are installed in the application
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSliceVar(&ccmd.extension, "extension", []string{}, "UI Extension to be installed into an application")
	cmd.Flags().StringVar(&ccmd.application, "application", "", "Application where the UI extension will be applied to")
	cmd.Flags().BoolVar(&ccmd.replaceAll, "replace", false, "Replace all UI extensions with the extensions provided")
	cmd.Flags().BoolVar(&ccmd.updateAll, "update-versions", false, "Update to the latest version for extensions already installed in an application")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUIExtensionWithVersions("extension", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("application")

	return ccmd
}

func (n *CmdInstall) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	body := mapbuilder.NewInitializedMapBuilder(true)

	// Lookup application where the extension will be applied to
	// TODO: Check if the user is trying to update an extension that is owned by another tenant
	refs, err := c8yfetcher.FindHostedApplications(n.factory, []string{n.application}, true, "", true)
	if err != nil || len(refs) == 0 {
		return fmt.Errorf("failed to find hosted application. %s", err)
	}

	var app gjson.Result
	if v, ok := refs[0].Data.Value.(gjson.Result); ok {
		app = v
	} else {
		_, resp, err := client.Application.GetApplication(context.Background(), refs[0].ID)
		if err != nil {
			return err
		}
		app = resp.JSON()
	}

	remotes := make(map[string]ExtensionReference)

	if !n.replaceAll {
		// Get existing remotes
		if existingRemotes := app.Get("config.remotes"); existingRemotes.Exists() && existingRemotes.IsObject() {
			existingRemotes.ForEach(func(key, value gjson.Result) bool {
				name, version, valid := strings.Cut(key.String(), "@")

				if !valid {
					return true
				}

				remote := ExtensionReference{
					ContextPath: name,
					Version:     version,
					Modules:     []string{},
				}

				if value.IsArray() {
					value.ForEach(func(idx, value gjson.Result) bool {
						remote.Modules = append(remote.Modules, value.String())
						return true
					})
				}
				if len(remote.Modules) > 0 {
					remotes[remote.ContextPath] = remote
				}

				if n.updateAll {
					n.extension = append(n.extension, fmt.Sprintf("%s@latest", name))
				}

				return true
			})
		}
	}

	// Find extensions
	for _, nameVersion := range n.extension {
		name, versionOrTag, _ := strings.Cut(nameVersion, "@")
		matches, err := c8yfetcher.FindUIExtensions(n.factory, []string{name}, true, "", true)
		if err != nil {
			return err
		}
		for _, ref := range matches {
			if ext, ok := ref.Data.Value.(gjson.Result); ok {
				if err := buildUIRemotes(remotes, ext, versionOrTag); err != nil {
					return err
				}
			}
		}
	}

	// Get existing config
	if v := app.Get("config"); v.Exists() && v.IsObject() {
		body = mapbuilder.NewMapBuilderWithInit([]byte(app.Get("config").Str))
	}

	body.Set("id", app.Get("id").String())
	body.Set("config.remotes", formatApplicationRemotes(remotes))

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
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

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
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
		flags.WithProcessingModeValue(),
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
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("config.remotes"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications/" + refs[0].ID)
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
