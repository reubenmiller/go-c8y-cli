package create

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/artifact"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications/versions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/ui/plugins"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

type CmdCreate struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	file         string
	name         string
	key          string
	version      string
	availability string
	contextPath  string
	resourceURL  string
	activate     bool
	tags         []string
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create/Update UI plugin and version",
		Long: heredoc.Doc(`
			Create a new UI plugin or update the binary of an existing UI plugin

			Notes
			* If the given file has a space in the filename, then it will be copied to a temp location
			  and the spaces will be replaced with underscores otherwise the server will reject the
			  file upload
		`),
		Example: heredoc.Doc(`
			$ c8y ui plugins create --file ./myapp.zip --tag latest
			Create new ui plugin from a file and tag it as the latest

			$ c8y ui plugins create --name my-plugin --file ./myapp.zip
			Create or update a ui plugin using an explicit name

			$ c8y ui plugins create --file ./myapp.zip --name my-plugin --version 1.0.0 --tag latest --tag stable-v1
			Create/update a ui plugin and override the name and version

			$ c8y ui plugins create --file "https://github.com/SoftwareAG/cumulocity-remote-access-cloud-http-proxy/releases/download/v2.5.0/cloud-http-proxy-ui.zip"
			Create/update a ui plugin from a URL
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "UI plugin file to be uploaded file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application")
	cmd.Flags().StringVar(&ccmd.version, "version", "", "Plugin version")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, SHARED, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().StringVar(&ccmd.resourceURL, "resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")
	cmd.Flags().StringSliceVar(&ccmd.tags, "tags", []string{}, "Tags. Include 'latest' to change the activeVersionId of the application")
	cmd.Flags().StringSliceVar(&ccmd.tags, "tag", []string{}, "Tags. Include 'latest' to change the activeVersionId of the application")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("availability", "SHARED", "PRIVATE", "MARKET"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.MarkRenamed("tags", "tag"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("file")

	return ccmd
}

// getApplicationDetails returns the details of the plugin and the version of
// the given file, which are read from the manifest of the plugin (and can be
// overridden by the flags)
func (n *CmdCreate) getApplicationDetails(file string) (*plugins.Plugin, *plugins.ManifestFile, error) {

	// set default name to the file name
	appNameFromFile := artifact.ParseName(file)
	fileExt := filepath.Ext(file)

	// Set application properties
	manifestFile := &plugins.ManifestFile{}
	if err := plugins.GetManifestContents(file, manifestFile); err != nil {
		return nil, nil, err
	}

	// Check if it is really a plugin!
	allowedTypes := []string{"plugin", "blueprint"}
	if !slices.Contains(allowedTypes, strings.ToLower(manifestFile.Package)) {
		return nil, nil, fmt.Errorf("invalid file. The given file is not a UI plugin or blueprint (e.g. the manifest.package is not one of [%v]). file=%s", allowedTypes, file)
	}

	// The manifest is not part of the plugin which is created (it is read from
	// the uploaded file by the platform)
	app := &plugins.Plugin{
		Name:        manifestFile.Name,
		Key:         manifestFile.Key,
		Type:        plugins.ApplicationTypeHosted,
		ContextPath: manifestFile.ContextPath,
	}

	// Set application name using the following preferences (first match wins)
	// 1. Explicit name
	// 2. Name from file (if the given file is not a json file) - as this allows
	//    overriding the app name by just changing the file name (and not requiring to edit it)
	// 3. Name from manifest file
	if manifestFile.Name != "" {
		app.Name = manifestFile.Name
	}

	// The file is always a local file here (a url was downloaded first), so the
	// name of a downloaded plugin is derived from the name of the downloaded file
	if !strings.EqualFold(fileExt, ".json") && strings.EqualFold(fileExt, ".zip") {
		app.Name = appNameFromFile
	}

	if n.name != "" {
		app.Name = n.name
	}

	app.Key = app.Name + "-key"
	if n.key != "" {
		app.Key = n.key
	}

	if n.availability != "" {
		app.Availability = n.availability
	}

	app.ContextPath = app.Name
	if n.contextPath != "" {
		app.ContextPath = n.contextPath
	}

	if n.version != "" {
		manifestFile.Version = n.version
	}

	return app, manifestFile, nil
}

func IsValidFilename(v string) bool {
	return !strings.ContainsAny(v, " ")
}

func ShouldDownload(v string) bool {
	return strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")
}

func DownloadFile(u string, log *logger.Logger) (string, error) {
	fileURL, urlErr := url.Parse(u)
	if urlErr != nil {
		return "", fmt.Errorf("invalid url format. %w", urlErr)
	}
	tmpFilename := filepath.Base(fileURL.Path)
	if filepath.Ext(tmpFilename) != ".zip" {
		tmpFilename = tmpFilename + ".zip"
	}
	tmpFile := filepath.Join(os.TempDir(), tmpFilename)
	log.Debugf("Downloading %s to %s", fileURL.String(), tmpFile)
	fTmpFile, fileErr := os.OpenFile(tmpFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return "", fileErr
	}
	// Don't use the name from a file in this case. tedge-container-plugin-ui_1.0.2?
	if downloadErr := fileutilities.DownloadFile(u, fTmpFile); downloadErr != nil {
		return "", fmt.Errorf("could not download plugin. %w", downloadErr)
	}
	return tmpFile, nil
}

// RunE uploads the plugin as a new version of the plugin, and creates the
// plugin first if it does not exist yet. The new version is activated when it
// is tagged as the latest version. The result is the new version.
func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	// v2.59.0 sent the requests of this command without a page size
	r.NoDefaultPageSize()
	client, err := r.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	progress := n.factory.IOStreams.ProgressIndicator()

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		plan, err := n.prepare(in, client, log)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) any {
			return plan.upload(ctx, client, progress)
		}, nil
	})
}

// uploadPlan is everything which is needed to upload the plugin
type uploadPlan struct {
	file     string
	plugin   *plugins.Plugin
	existing string // id of the existing plugin
	version  string
	tags     []string
	activate bool
	cleanup  []func()
}

func (p *uploadPlan) close() {
	for _, fn := range slices.Backward(p.cleanup) {
		fn()
	}
	p.cleanup = nil
}

// prepare gets the plugin file (downloading it when it is a url), reads its
// manifest and looks up the existing plugin
func (n *CmdCreate) prepare(in *c8ystream.Resolver, client *apiv2.Client, log *logger.Logger) (*uploadPlan, error) {
	plan := &uploadPlan{}
	if err := n.prepareFile(in, client, log, plan); err != nil {
		plan.close()
		return nil, err
	}
	return plan, nil
}

// fileError is a problem with the plugin file (it is missing, denied, not a
// plugin, could not be downloaded). It exits with 100 (unexpected error), as in v1.
func fileError(err error) error {
	return cmderrors.NewSystemError(err)
}

func (n *CmdCreate) prepareFile(in *c8ystream.Resolver, client *apiv2.Client, log *logger.Logger, plan *uploadPlan) error {
	file := n.file

	// Check if file is a url
	if ShouldDownload(file) {
		localFile, downloadErr := DownloadFile(file, log)
		if downloadErr != nil {
			return fileError(fmt.Errorf("could not download plugin. %w", downloadErr))
		}
		log.Infof("Downloaded plugin to %s", localFile)
		file = localFile
		plan.cleanup = append(plan.cleanup, func() {
			if err := os.Remove(localFile); err != nil {
				log.Warnf("could not delete downloaded file. %s", err)
			}
		})
	} else if err := fileguard.Check(file); err != nil {
		return fileError(err)
	}

	application, manifest, err := n.getApplicationDetails(file)
	if err != nil {
		return fileError(err)
	}

	// The file is not allowed to have any spaces in the filename
	// as it will be rejected by the server
	// Let's help users out and create a temp file fixing the names
	filename := filepath.Base(file)
	if !IsValidFilename(filename) {
		tmpFile := filepath.Join(os.TempDir(), strings.ReplaceAll(filename, " ", "_"))
		log.Warnf("Plugin file contains a space, so creating a temp file without a space to avoid being rejected by the server. tmpfile=%s", tmpFile)

		if err := fileutilities.CopyFile(tmpFile, file); err != nil {
			return fileError(fmt.Errorf("could not copy file to tmp directory. %w", err))
		}
		plan.cleanup = append(plan.cleanup, func() { _ = os.Remove(tmpFile) })
		file = tmpFile
	}

	// Activate the tag if a tag includes 'latest'
	plan.activate = plugins.HasTag(n.tags, plugins.ApplicationTagLatest)
	plan.tags = append([]string{}, n.tags...)

	// Only the plugins which are owned by the current tenant can be managed
	matches, err := client.UIPlugins.List(in.ResolveContext(), plugins.ListOptions{
		Name:        application.Name,
		Owner:       client.Auth.Tenant,
		HasVersions: true,
	}).Collect()
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		plan.existing = matches[0].ID()
	} else {
		// A new plugin must have its first version activated
		plan.activate = true
		if !plugins.HasTag(plan.tags, plugins.ApplicationTagLatest) {
			plan.tags = append(plan.tags, plugins.ApplicationTagLatest)
		}
	}

	plan.file = file
	plan.plugin = application
	plan.version = manifest.Version
	return nil
}

// upload creates the plugin (if it does not exist), uploads the new version
// and activates it. Only the new version is returned.
func (p *uploadPlan) upload(ctx context.Context, client *apiv2.Client, progress *iostreams.Progress) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		defer p.close()
		fail := func(err error) { yield(jsondoc.Empty(), err) }

		appID := p.existing
		if appID == "" {
			created := client.UIPlugins.Create(ctx, p.plugin)
			if created.Err != nil {
				fail(created.Err)
				return
			}
			if appID = created.Data.ID(); appID == "" {
				// dry run: the plugin was not created, so there is no version to upload
				return
			}
		} else if p.plugin.Availability != "" {
			// Update the existing plugin
			updated := client.UIPlugins.Update(ctx, appID, map[string]any{"availability": p.plugin.Availability})
			if updated.Err != nil {
				fail(updated.Err)
				return
			}
		}

		file, err := fileguard.Open(p.file)
		if err != nil {
			fail(err)
			return
		}
		defer file.Close()

		opts := versions.CreateOptions{
			Version: p.version,
			Tags:    p.tags,
			File:    versions.UploadFileOptions{FilePath: p.file, Reader: file},
		}
		finish := func() {}
		if progress.IsEnabled() {
			fn, done, err := c8ybinary.NewUploadProgress(progress, p.file)
			if err != nil {
				fail(err)
				return
			}
			opts.File.Progress = fn
			finish = done
		}
		version := client.Applications.Versions.Create(ctx, appID, opts)
		finish()
		progress.Wait()
		if version.Err != nil {
			fail(version.Err)
			return
		}

		// Activate the version
		if binaryID := version.Data.BinaryID(); p.activate && binaryID != "" {
			if activated := client.UIPlugins.Activate(ctx, appID, binaryID); activated.Err != nil {
				fail(activated.Err)
				return
			}
		}

		for doc, err := range version.Documents() {
			if !yield(doc, err) {
				return
			}
		}
	}
}
