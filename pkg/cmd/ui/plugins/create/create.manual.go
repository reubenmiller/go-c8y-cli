package create

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
			return f.CreateModeEnabled()
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

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("availability", "SHARED", "PRIVATE", "MARKET"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("file")

	return ccmd
}

func (n *CmdCreate) getApplicationDetails(client *c8y.Client, log *logger.Logger) (*c8y.UIExtension, error) {

	// set default name to the file name
	baseFileName := filepath.Base(n.file)
	fileExt := filepath.Ext(baseFileName)
	baseFileName = baseFileName[0 : len(baseFileName)-len(fileExt)]
	versionRegex := regexp.MustCompile(`(-v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	appNameFromFile := versionRegex.ReplaceAllString(baseFileName, "")

	// Set application properties
	app, err := client.UIExtension.NewUIExtensionFromFile(n.file)
	if err != nil {
		return nil, err
	}

	// Check if it is really a plugin!
	if app.ManifestFile.Package != "plugin" {
		return nil, fmt.Errorf("invalid file. The given file is not a UI plugin (e.g. the manifest.package != 'plugin'). file=%s", n.file)
	}

	// Set application name using the following preferences (first match wins)
	// 1. Explicit name
	// 2. Name from file (if the given file is not a json file) - as this allows
	//    overriding the app name by just changing the file name (and not requiring to edit it)
	// 3. Name from manifest file
	if app.ManifestFile.Name != "" {
		app.Name = app.ManifestFile.Name
	}

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
		app.ManifestFile.Version = n.version
	}

	return app, nil
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
	if downloadErr := fileutilities.DownloadFile(u, fTmpFile); downloadErr != nil {
		return "", fmt.Errorf("could not download plugin. %w", downloadErr)
	}
	return tmpFile, nil
}

func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	handler, err := n.factory.GetRequestHandler()
	if err != nil {
		return err
	}

	// Check if file is a url
	if ShouldDownload(n.file) {
		if cfg.DryRun() {
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "DRY: Downloading plugin from url: %s\n", n.file)
		} else {
			localFile, downloadErr := DownloadFile(n.file, log)
			if downloadErr != nil {
				return fmt.Errorf("could not download plugin. %w", downloadErr)
			}
			log.Infof("Downloaded plugin to %s", localFile)
			n.file = localFile

			defer func() {
				if err := os.Remove(localFile); err != nil {
					log.Warnf("could not delete downloaded file. %w", err)
				}
			}()
		}

	}

	dryRun := cfg.ShouldUseDryRun(cmd.CommandPath())
	application, err := n.getApplicationDetails(client, log)
	if err != nil {
		return err
	}

	// The file is not allowed to have any spaces in the filename
	// as it will be rejected by the server
	// Let's help users out and create a temp file fixing the names
	filename := filepath.Base(n.file)
	if !IsValidFilename(filename) {
		tmpFile := filepath.Join(os.TempDir(), strings.ReplaceAll(filename, " ", "_"))
		log.Warnf("Plugin file contains a space, so creating a temp file without a space to avoid being rejected by the server. tmpfile=%s", tmpFile)

		if err := fileutilities.CopyFile(tmpFile, n.file); err != nil {
			return fmt.Errorf("could not copy file to tmp directory. %w", err)
		}
		defer os.Remove(tmpFile)
		n.file = tmpFile
	}

	ctx := c8y.WithCommonOptionsContext(context.Background(), c8y.CommonOptions{
		DryRun: dryRun,
	})

	file, fileErr := os.Open(n.file)
	if fileErr != nil {
		return fileErr
	}
	progress := n.factory.IOStreams.ProgressIndicator()
	bar, barErr := c8ybinary.NewProgressBar(progress, n.file)
	if barErr != nil {
		return barErr
	}

	var fileReader io.ReadCloser
	fileReader = file
	if bar != nil {
		fileReader = bar.ProxyReader(file)
	}

	// Activate the tag if a tag includes 'latest'
	for _, tag := range n.tags {
		if strings.EqualFold(tag, "latest") {
			n.activate = true
			break
		}
	}

	_, response, err := client.UIExtension.CreateExtension(ctx, &application.Application, fileReader, c8y.UpsertOptions{
		SkipActivation: !n.activate,
		Version: &c8y.ApplicationVersion{
			Version: application.ManifestFile.Version,
			Tags:    n.tags,
		},
	})
	if err != nil {
		return err
	}
	n.factory.IOStreams.WaitForProgressIndicator()
	commonOptions.ResultProperty = "-"
	_, err = handler.ProcessResponse(response, err, nil, commonOptions)
	return err
}
