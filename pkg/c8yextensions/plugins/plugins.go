package plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type PluginCmd struct {
	*subcommand.SubCommand

	Factory *cmdutil.Factory

	Application   string
	Remove        []string
	Add           []string
	ReplaceAll    bool
	UpdateAll     bool
	RemoveInvalid bool
}

type ExtensionReference struct {
	ContextPath string
	Version     string
	Modules     []string
}

func (r *ExtensionReference) Key() string {
	if r.Version == "" {
		return fmt.Sprintf("%s@%s", r.ContextPath, "latest")
	}
	return fmt.Sprintf("%s@%s", r.ContextPath, r.Version)
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
		return cmderrors.NewUserError(fmt.Sprintf("plugin version not found. plugin=%s, version=%s", ext.Get("name").String(), versionOrTag))
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

func NewPluginRunner(cmd *cobra.Command, args []string, f *cmdutil.Factory, managerOptions *PluginCmd, opts ...flags.GetOption) func() error {
	return func() error {
		cfg, err := f.Config()
		if err != nil {
			return err
		}

		client, err := f.Client()
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
		body := mapbuilder.NewInitializedMapBuilder(true)
		body.SetAppendTemplatePreference(true)
		err = flags.WithBody(
			cmd,
			body,
			inputIterators,
			flags.WithDataFlagValue(),
			cmdutil.WithTemplateValue(f),
			flags.WithTemplateVariablesValue(),
		)
		if err != nil {
			return cmderrors.NewUserError(err)
		}

		// path parameters
		path := flags.NewStringTemplate("/application/applications/{application}")
		err = flags.WithPathParameters(
			cmd,
			path,
			inputIterators,
			c8yfetcher.WithHostedApplicationByNameFirstMatch(f, args, "application", "application"),
		)
		if err != nil {
			return err
		}

		req := c8y.RequestOptions{
			Method:                 "PUT",
			Path:                   path.GetTemplate(),
			Query:                  queryValue,
			Body:                   body,
			FormData:               formData,
			Header:                 headers,
			PrepareRequest:         prepareBody(f, managerOptions),
			PrepareRequestOnDryRun: true,
			IgnoreAccept:           cfg.IgnoreAcceptHeader(),
			DryRun:                 cfg.ShouldUseDryRun(cmd.CommandPath()),
		}

		return f.RunWithWorkers(client, cmd, &req, inputIterators)
	}
}

func buildBody(f *cmdutil.Factory, applicationID string, managerOptions *PluginCmd, inputBody io.ReadCloser) (io.Reader, error) {

	client, err := f.Client()
	if err != nil {
		return nil, err
	}

	log, err := f.Logger()
	if err != nil {
		return nil, err
	}

	// Lookup application where the extension will be applied to
	// TODO: Check if the user is trying to update an extension that is owned by another tenant
	refs, err := c8yfetcher.FindHostedApplications(f, []string{applicationID}, true, "", true)
	if err != nil || len(refs) == 0 {
		return nil, fmt.Errorf("failed to find hosted application. %s", err)
	}

	var app gjson.Result
	if v, ok := refs[0].Data.Value.(gjson.Result); ok {
		app = v
	} else {
		_, resp, err := client.Application.GetApplication(context.Background(), refs[0].ID)
		if err != nil {
			return nil, err
		}
		app = resp.JSON()
	}

	remotes := make(map[string]ExtensionReference)

	extensions := make([]string, 0)
	extensions = append(extensions, managerOptions.Add...)

	bodyErrors := make([]error, 0)

	if !managerOptions.ReplaceAll {
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

				// Remove invalid plugins which have either been removed or the version no longer exists
				if managerOptions.UpdateAll {
					extensions = append(extensions, fmt.Sprintf("%s@latest", remote.ContextPath))
				} else if managerOptions.RemoveInvalid {
					// Add to the extensions list as these will be validated afterwards
					extensions = append(extensions, remote.Key())
				} else {
					remotes[remote.ContextPath] = remote
				}
				return true
			})
		}
	}

	// Add plugins (only valid)
	for _, nameVersion := range extensions {
		name, versionOrTag, _ := strings.Cut(nameVersion, "@")
		matches, err := c8yfetcher.FindUIPlugins(f, []string{name}, true, "", true)
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			for _, ref := range matches {
				if ext, ok := ref.Data.Value.(gjson.Result); ok {
					if err := buildUIRemotes(remotes, ext, versionOrTag); err != nil {
						if managerOptions.RemoveInvalid {
							// ignore error
							log.Warnf("Removing reference to revoked plugin version: name=%s, version=%s, remote=%s", ext.Get("name").String(), versionOrTag, nameVersion)
							continue
						}
						return nil, err
					}
				}
			}
		} else if managerOptions.RemoveInvalid {
			// ignore error
			log.Warnf("Removing reference to orphaned plugin: remote=%s", nameVersion)
			continue
		} else {
			bodyErrors = append(bodyErrors, fmt.Errorf("plugin does not exist. name=%s, tag/version=%s, remote=%s", name, versionOrTag, nameVersion))
		}
	}

	// Remove plugins
	for _, nameVersion := range managerOptions.Remove {
		name, _, _ := strings.Cut(nameVersion, "@")
		matches, err := c8yfetcher.FindUIPlugins(f, []string{name}, true, "", true)
		if err != nil {
			return nil, err
		}
		for _, ref := range matches {
			if ext, ok := ref.Data.Value.(gjson.Result); ok {
				delete(remotes, ext.Get("contextPath").String())
			}
		}
	}

	// Create a new body which merges the retrieved plugin info, and any given templates
	currentBody, err := io.ReadAll(inputBody)
	if err != nil {
		return nil, err
	}

	body := mapbuilder.NewInitializedMapBuilder(true)

	// Get existing config
	if v := app.Get("config"); v.Exists() && v.IsObject() {
		// body.AppendTemplate(app.Get("config").Str)
		body = mapbuilder.NewMapBuilderWithInit([]byte(app.Get("config").Str))
	}

	// HACK: Perform a deep merge by using the jsonnet "+:" operator
	// This is required because the original template has already be converted to json
	// thus losing information
	body.AppendTemplate(string(convertToJsonnetDeepMerge(currentBody)))
	body.SetAppendTemplatePreference(true)

	body.Set("id", app.Get("id").String())
	body.Set("config.remotes", formatApplicationRemotes(remotes))
	body.SetRequiredKeys("config.remotes")

	tmpBody, err := body.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(tmpBody), wrapErrors(bodyErrors)
}

func wrapErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return cmderrors.NewUserError("Invalid user input. ", errs[0])
	default:
		return cmderrors.NewUserError("Invalid user input. ", errors.Join(errs...))
	}
}

func convertToJsonnetDeepMerge(s []byte) []byte {
	re := regexp.MustCompile(`:(\s*){`)
	return re.ReplaceAll(s, []byte(`+:$1{`))
}

func prepareBody(f *cmdutil.Factory, managerOptions *PluginCmd) func(r *http.Request) (*http.Request, error) {
	return func(r *http.Request) (*http.Request, error) {
		var applicationID string
		if i := strings.LastIndex(r.URL.Path, "/"); i > -1 {
			applicationID = r.URL.Path[i+1:]
		}

		body, err := buildBody(f, applicationID, managerOptions, r.Body)
		if err != nil {
			return r, err
		}
		r.Body = io.NopCloser(body)
		return r, nil
	}
}
