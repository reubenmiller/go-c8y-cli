package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
	availability string
	contextPath  string
	resourceURL  string
	// TODO: Check if a new binary version needs to be activated or not
	skipActivate bool
	tags         []string
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create/Update UI extension and version",
		Long: heredoc.Doc(`
			Create a new UI extension or update the binary of an existing UI extension

			Notes
			* If the given file has a space in the filename, then it will be copied to a temp location
			  and the spaces will be replaced with underscores otherwise the server will reject the
			  file upload
		`),
		Example: heredoc.Doc(`
			$ c8y ui extensions create --file ./myapp.zip
			Create new ui extension from a file

			$ c8y ui extensions create --name my-extension --file ./myapp.zip
			Create or update a ui extension using an explicit name
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "UI Extension file to be uploaded file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().StringVar(&ccmd.resourceURL, "resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")
	cmd.Flags().BoolVar(&ccmd.skipActivate, "skipActivate", false, "Skip activating the new version of the application")
	cmd.Flags().StringSliceVar(&ccmd.tags, "tags", []string{}, "Tags")

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
		return nil, fmt.Errorf("invalid file. The given file is not a UI extension (e.g. the manifest.package != 'plugin'). file=%s", n.file)
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

	return app, nil
}

func IsValidFilename(v string) bool {
	return !strings.ContainsAny(v, " ")
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

	// dryRun := cfg.ShouldUseDryRun(cmd.CommandPath())
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
		log.Warnf("Extension file contains a space, so creating a temp file without a space to avoid being rejected by the server. tmpfile=%s", tmpFile)

		if err := fileutilities.CopyFile(tmpFile, n.file); err != nil {
			return fmt.Errorf("could not copy file to tmp directory. %w", err)
		}
		defer os.Remove(tmpFile)
		n.file = tmpFile
	}

	_, response, err := client.UIExtension.CreateExtension(context.Background(), &application.Application, n.file, c8y.UpsertOptions{
		SkipActivation: n.skipActivate,
		Version: &c8y.ApplicationVersion{
			Version: application.ManifestFile.Version,
			Tags:    n.tags,
		},
	})

	_, err = handler.ProcessResponse(response, err, nil, commonOptions)
	return err
}
