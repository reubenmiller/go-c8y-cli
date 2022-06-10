package create

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
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

const CumulocityManifestFile = "cumulocity.json"

type Application struct {
	c8y.Application
	Manifest Manifest
}

type Manifest struct {
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	RequiredRoles []string `json:"requiredRoles"`
	Roles         []string `json:"roles"`
}

type CmdCreate struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	file             string
	name             string
	key              string
	availability     string
	contextPath      string
	resourceURL      string
	skipSubscription bool
	skipUpload       bool
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create microservice",
		Long:  `Create a new microservice or update the application binary of an existing microservice`,
		Example: heredoc.Doc(`
$ c8y microservices create --file ./myapp.zip
Create new microservice

$ c8y microservices create --name my-application --file ./myapp.zip
Create or update a microservice using an explicit name
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "Microservice file to be uploaded (or Cumulocity.json) file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().StringVar(&ccmd.resourceURL, "resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipSubscription, "skipSubscription", false, "Skip microservice subscription when creating the new microservice")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Skip uploading the binary to the platform")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("file")

	return ccmd
}

func (n *CmdCreate) getApplicationDetails(log *logger.Logger) (*Application, error) {

	app := Application{}

	// set default name to the file name
	baseFileName := filepath.Base(n.file)
	fileExt := filepath.Ext(baseFileName)
	baseFileName = baseFileName[0 : len(baseFileName)-len(fileExt)]
	versionRegex := regexp.MustCompile(`(-v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	appNameFromFile := versionRegex.ReplaceAllString(baseFileName, "")

	// Set application properties

	if strings.HasSuffix(n.file, ".zip") {
		// Try loading manifest file directly from the zip (without unzipping it)
		log.Infof("Trying to detect manifest from a zip file. path=%s", n.file)
		if err := GetManifestContents(n.file, &app.Manifest); err != nil {
			log.Infof("Could not find manifest file. Expected %s to contain %s. %s", n.file, CumulocityManifestFile, err)
		}
	} else if n.file != "" {
		// Assume json (regardless of file type)
		log.Infof("Assuming file is json (regardless of file extension). path=%s", n.file)
		jsonFile, err := os.Open(n.file)
		if err != nil {
			return nil, err
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)

		if err := json.Unmarshal(byteValue, &app.Manifest); err != nil {
			log.Warnf("invalid manifest file. Only json or zip files are accepted. %s", strings.TrimSpace(err.Error()))
		}
	}

	// Set application name using the following preferences (first match wins)
	// 1. Explicit name
	// 2. Name from file (if the given file is not a json file) - as this allows
	//    overriding the app name by just changing the file name (and not requiring to edit it)
	// 3. Name from manifest file
	if app.Manifest.Name != "" {
		app.Name = app.Manifest.Name
	}

	if !strings.EqualFold(fileExt, ".json") && strings.EqualFold(fileExt, ".zip") {
		app.Name = appNameFromFile
	}

	if n.name != "" {
		app.Name = n.name
	}

	app.Key = app.Name
	if n.key != "" {
		app.Key = n.key
	}

	app.Type = "MICROSERVICE"

	if n.availability != "" {
		app.Availability = n.availability
	}

	app.ContextPath = app.Name
	if n.contextPath != "" {
		app.ContextPath = n.contextPath
	}

	app.ResourcesURL = "/"
	if n.resourceURL != "" {
		app.ResourcesURL = n.resourceURL
	}
	return &app, nil
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
	var application *c8y.Application
	var response *c8y.Response
	var applicationID string
	var applicationName string

	dryRun := cfg.ShouldUseDryRun(cmd.CommandPath())
	applicationDetails, err := n.getApplicationDetails(log)

	if err != nil {
		return err
	}

	if applicationDetails != nil {
		applicationName = applicationDetails.Name
	}

	if applicationName == "" {
		return cmderrors.NewUserError("Could not detect application name for the given input")
	}

	if applicationName != "" {

		refs, err := c8yfetcher.FindMicroservices(client, []string{applicationName}, true, "")

		if err != nil {
			return cmderrors.NewUserError(err)
		}

		idValue, _ := c8yfetcher.GetFetchedResultsAsString(refs)

		for _, item := range idValue {
			if item != "" {
				// Get first result (for consistency with c8y microservice get --id)
				applicationID = c8yfetcher.NewIDValue(item).GetID()
				break
			}
		}
	}

	if err != nil {
		return err
	}

	if applicationID == "" {
		// Create the application
		log.Info("Creating new application")
		application, response, err = client.Application.Create(context.Background(), &applicationDetails.Application)

		if err != nil {
			return fmt.Errorf("failed to create microservice. %s", err)
		}
	} else {
		// Get existing application
		log.Infof("Getting existing application. id=%s", applicationID)
		application, response, err = client.Application.GetApplication(context.Background(), applicationID)

		if err != nil {
			return fmt.Errorf("failed to get microservice. %s", err)
		}
	}

	skipUpload := n.skipUpload

	if _, err := os.Stat(n.file); err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("could not read manifest file. %s. error=%s", n.file, err))
	}

	// Only upload zip files
	if !strings.HasSuffix(n.file, ".zip") {
		log.Info("Skipping microservice binary upload")
		skipUpload = true
	}

	// Upload binary
	if !skipUpload {
		log.Infof("uploading binary [id=%s]", application.ID)
		if !dryRun {
			_, err := client.Application.CreateBinary(context.Background(), n.file, application.ID)

			if err != nil {
				// handle error
				n.SubCommand.GetCommand().PrintErrf("failed to upload file. %s", err)
			}
		}
	} else {
		//
		// Upload information from the cumulocity manifest file
		// because the zip file is not being uploaded because the app
		// will be hosted outside of the platform
		//
		// Read the Cumulocity.json file, and upload
		log.Infof("updating application details [id=%s], requiredRoles=%s", application.ID, strings.Join(applicationDetails.Manifest.RequiredRoles, ","))
		if !dryRun {
			_, response, err = client.Application.Update(context.Background(), application.ID, &c8y.Application{
				RequiredRoles: applicationDetails.Manifest.RequiredRoles,
			})
			if err != nil {
				return err
			}
		}
	}

	// App subscription
	if !n.skipSubscription {
		log.Infof("Subscribing to application")
		if !dryRun {
			_, resp, err := client.Tenant.AddApplicationReference(context.Background(), client.TenantName, application.Self)

			if err != nil {
				if resp != nil && resp.StatusCode == 409 {
					log.Infof("microservice is already enabled")
				} else {
					return fmt.Errorf("Failed to subscribe to application. %s", err)
				}
			}
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
	_, err = handler.ProcessResponse(response, err, commonOptions)
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
