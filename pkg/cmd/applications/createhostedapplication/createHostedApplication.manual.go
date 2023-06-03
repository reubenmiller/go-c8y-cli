package createhostedapplication

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/zipUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

const CumulocityManifestFile = "cumulocity.json"

type Application struct {
	c8y.Application
	Manifest Manifest
}

type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Key         string `json:"key"`
	Version     string `json:"version"`
	ContextPath string `json:"contextPath"`
}

type CmdCreateHostedApplication struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	file           string
	name           string
	key            string
	availability   string
	contextPath    string
	resourcesURL   string
	skipActivation bool
	skipUpload     bool
}

func NewCmdCreateHostedApplication(f *cmdutil.Factory) *CmdCreateHostedApplication {
	ccmd := &CmdCreateHostedApplication{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "createHostedApplication",
		Short: "Create hosted application",
		Long: heredoc.Doc(`
			Create a new hosted web application or update the binary of an existing hosted application

			If the zip file or folder contains a 'cumulocity.json' manifest file, then the key will be automatically read from it.
		`),
		Example: heredoc.Doc(`
			$ c8y applications createHostedApplication --file ./myapp.zip
			Create new hosted application from a given zip file. The application will be called "myapp". If the application placeholder does not exist then it will be created

			$ c8y applications createHostedApplication --file simple-helloworld/build --name custom_helloworld
			Create/update hosted web application from a build folder and specify a custom application name

			$ c8y applications createHostedApplication --file myapp.zip --skipActivation
			Create/update hosted web application but don't activate it, so the current version (if any) will be untouched
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "File or Folder of the web application. It should contain a index.html file in the root folder/ or zip file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application. Defaults to the value inside the cumulocity.json file (if present)")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application")
	cmd.Flags().StringVar(&ccmd.resourcesURL, "resourcesUrl", "/", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipActivation, "skipActivation", false, "Don't activate to the application after it has been created and uploaded")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Don't uploaded the web app binary. Only the application placeholder will be created")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("availability", "MARKET", "PRIVATE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdCreateHostedApplication) getApplicationDetails(log *logger.Logger) (*Application, error) {

	app := Application{}

	// set default name to the file name
	baseFileName := filepath.Base(n.file)
	baseFileName = baseFileName[0 : len(baseFileName)-len(filepath.Ext(baseFileName))]
	versionRegex := regexp.MustCompile(`(-v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	appNameFromFile := versionRegex.ReplaceAllString(baseFileName, "")

	// Set application properties

	if strings.EqualFold(filepath.Ext(n.file), ".zip") {
		// Try loading manifest file directly from the zip (without unzipping it)
		log.Infof("Trying to detect manifest from a zip file. path=%s", n.file)
		if err := GetManifestContents(n.file, &app.Manifest); err != nil {
			log.Infof("Could not find manifest file. Expected %s to contain %s. %s", n.file, CumulocityManifestFile, err)
		}
	} else if n.file != "" {
		// Assume json (regardless of file type)
		manifestPath := filepath.Join(n.file, CumulocityManifestFile)
		log.Infof("Assuming file is json (regardless of file extension). path=%s", manifestPath)

		if _, err := os.Stat(manifestPath); err == nil {
			jsonFile, err := os.Open(manifestPath)
			if err != nil {
				return nil, err
			}
			byteValue, _ := ioutil.ReadAll(jsonFile)

			if err := json.Unmarshal(byteValue, &app.Manifest); err != nil {
				log.Warnf("invalid manifest file. Only json or zip files are accepted. %s", strings.TrimSpace(err.Error()))
			}
		}
	}

	if app.Manifest.Name != "" {
		app.Name = app.Manifest.Name
	}

	app.Name = appNameFromFile
	if app.Manifest.Name != "" {
		app.Name = app.Manifest.Name
	}
	if n.name != "" {
		app.Name = n.name
	}

	app.Key = app.Name + "-application-key"
	if app.Manifest.Key != "" {
		app.Key = app.Manifest.Key
	}
	if n.key != "" {
		app.Key = n.key
	}

	app.Type = "HOSTED"

	if n.availability != "" {
		app.Availability = n.availability
	}

	app.ContextPath = app.Name
	if app.Manifest.ContextPath != "" {
		app.ContextPath = app.Manifest.ContextPath
	}
	if n.contextPath != "" {
		app.ContextPath = n.contextPath
	}

	app.ResourcesURL = "/"
	if n.resourcesURL != "" {
		app.ResourcesURL = n.resourcesURL
	}
	return &app, nil
}

// packageWebApplication zips the given folder path to a zip
func (n *CmdCreateHostedApplication) packageWebApplication(src string) (string, error) {
	dir, err := ioutil.TempDir("", "c8y-packer")
	if err != nil {
		return "", fmt.Errorf("failed to create temp folder. %w", err)
	}

	dstZip := filepath.Join(dir, filepath.Base(src)+".zip")
	// zip folder
	if err := zipUtilities.ZipDirectoryContents(src, dstZip); err != nil {
		return "", fmt.Errorf("failed to zip source. %w", err)
	}

	return dstZip, nil
}

// packageAppIfRequired normalizes the handling of the given app. If src is a zip file, then nothing is done.
// If the src is a folder, then the folder contents are zipped and the path to the zip file are returned.
func (n *CmdCreateHostedApplication) packageAppIfRequired(src string) (zipFile string, err error) {
	f, err := os.Stat(src)

	if err != nil {
		return
	}

	if !f.IsDir() {
		// it is already a zip
		zipFile = src
		return
	}

	if log, err := n.factory.Logger(); log != nil && err == nil {
		log.Infof("zipping folder %s", src)
	}
	zipFile, err = n.packageWebApplication(src)

	if err != nil {
		err = fmt.Errorf("failed to package web application. %s", err)
	}
	return
}

func (n *CmdCreateHostedApplication) RunE(cmd *cobra.Command, args []string) error {
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
	var application *c8y.Application
	var response *c8y.Response
	var applicationID string

	// note: use POST when checking if it should use try run or not, even though it could actually be PUT as well
	dryRun := cfg.ShouldUseDryRun(cmd.CommandPath())
	appDetails, err := n.getApplicationDetails(log)

	if err != nil {
		return err
	}

	// TODO: Use the default name value from n.Name rather then reading it from the args again.
	log.Infof("application name: %s", appDetails.Name)
	if appDetails.Name != "" {
		refs, err := c8yfetcher.FindHostedApplications(n.factory, []string{appDetails.Name}, true, "", true)

		if err != nil {
			return fmt.Errorf("failed to find hosted application. %s", err)
		}

		if err == nil && len(refs) > 0 {
			applicationID = refs[0].ID
		}
	}

	if applicationID == "" {
		// Create the application
		log.Info("Creating new application")
		application, response, err = client.Application.Create(context.Background(), &appDetails.Application)

		if err != nil {
			return fmt.Errorf("failed to create hosted application. %s", err)
		}
		applicationID = application.ID
	} else {
		// Get existing application
		log.Infof("Getting existing application. id=%s", applicationID)
		application, response, err = client.Application.GetApplication(
			c8yfetcher.WithDisabledDryRunContext(client),
			applicationID,
		)

		if err != nil {
			return fmt.Errorf("failed to get hosted application. %s", err)
		}
	}

	skipUpload := n.skipUpload || n.file == ""

	// Upload binary
	applicationBinaryID := ""
	if !skipUpload {
		if !dryRun {
			zipfile, err := n.packageAppIfRequired(n.file)
			if err != nil {
				log.Errorf("Failed to package file. %s", err)
				return fmt.Errorf("failed to package app. %s", err)
			}

			log.Infof("uploading binary [app=%s]", application.ID)
			progress := n.factory.IOStreams.ProgressIndicator()
			resp, err := c8ybinary.CreateBinaryWithProgress(
				context.Background(),
				client,
				"/application/applications/"+application.ID+"/binaries",
				zipfile,
				nil,
				progress,
			)
			n.factory.IOStreams.WaitForProgressIndicator()

			if err != nil {
				// handle error
				n.SubCommand.GetCommand().PrintErrf("failed to upload file. %s", err)
			} else {
				applicationBinaryID = resp.JSON("id").String()
			}
		}
	}

	// App activation (only if a new version was uploaded)
	if !skipUpload && !n.skipActivation {
		if !dryRun {
			log.Infof("Activating application")

			if applicationBinaryID == "" {
				return fmt.Errorf("failed to activate new application version because binary id is empty")
			}

			_, resp, err := client.Application.Update(
				context.Background(),
				applicationID,
				&c8y.Application{
					ActiveVersionID: applicationBinaryID,
				},
			)

			if err != nil {
				if resp != nil && resp.StatusCode() == 409 {
					log.Infof("application is already enabled")
				} else {
					return fmt.Errorf("failed to activate application. %s", err)
				}
			}

			// use the updated application json
			response = resp
		}
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	handler, err := n.factory.GetRequestHandler()
	if err != nil {
		return err
	}
	_, err = handler.ProcessResponse(response, err, nil, commonOptions)
	return err
}

func GetManifestContents(zipFilename string, contents interface{}) error {
	reader, err := zip.OpenReader(zipFilename)
	if err != nil {
		return err
	}

	defer reader.Close()

	for _, file := range reader.File {
		// check if the file matches the name for application portfolio xml
		if strings.EqualFold(file.Name, CumulocityManifestFile) {
			rc, err := file.Open()
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(rc); err != nil {
				return err
			}

			defer rc.Close()

			// Unmarshal bytes
			if err := json.Unmarshal(buf.Bytes(), &contents); err != nil {
				return err
			}
		}
	}
	return nil
}
