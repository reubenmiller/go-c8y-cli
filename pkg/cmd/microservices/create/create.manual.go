package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/zipUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

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
		flags.WithData(),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("file")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdCreate) getApplicationDetails() *c8y.Application {

	app := c8y.Application{}

	// set default name to the file name
	baseFileName := filepath.Base(n.file)
	baseFileName = baseFileName[0 : len(baseFileName)-len(filepath.Ext(baseFileName))]
	versionRegex := regexp.MustCompile(`(-v?\d+\.\d+\.\d+(-SNAPSHOT)?)?$`)
	appNameFromFile := versionRegex.ReplaceAllString(baseFileName, "")

	// Set application properties

	app.Name = appNameFromFile
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
	return &app
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

	dryRun := cfg.DryRun()
	applicationDetails := n.getApplicationDetails()

	if applicationDetails != nil {
		applicationName = applicationDetails.Name
	}

	if applicationName == "" {
		return cmderrors.NewUserError("Could not detect application name for the given input")
	}

	if applicationName != "" {

		refs, err := c8yfetcher.FindMicroservices(client, []string{applicationName}, true)

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
		application, response, err = client.Application.Create(context.Background(), applicationDetails)

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
		var requiredRoles []string
		requiredRoles = make([]string, 0)
		var manifestContents map[string]interface{}
		// manifestContents = make(map[string]interface{})

		var manifestFile string

		if strings.HasSuffix(n.file, ".json") {
			// user provided just a manifest file
			manifestFile = n.file
		} else if strings.HasSuffix(n.file, ".zip") {
			if val, err := GetManifestFile(n.file); err != nil {
				log.Warningf("failed to get manifest file from microservice. %s", err)
			} else {
				manifestFile = val
			}
		}

		if v, err := jsonUtilities.DecodeJSONFile(manifestFile); err == nil {
			manifestContents = v
		} else {
			log.Warningf("failed to decode manifest file. file=%s, err=%s", manifestFile, err)
			return cmderrors.NewUserError(fmt.Sprintf("invalid manifest file. Only json files are accepted. %s", strings.TrimSpace(err.Error())))
		}

		if roles, ok := manifestContents["requiredRoles"].([]interface{}); ok {
			for _, val := range roles {
				if role, typeOk := val.(string); typeOk {
					requiredRoles = append(requiredRoles, role)
				}
			}
		} else {
			log.Warningf("Failed to read requiredRoles. contents=%v, type=%T", manifestContents, roles)
		}

		// Read the Cumulocity.json file, and upload
		log.Infof("updating application details [id=%s], requiredRoles=%s", application.ID, strings.Join(requiredRoles, ","))
		if !dryRun {
			_, response, err = client.Application.Update(context.Background(), application.ID, &c8y.Application{
				RequiredRoles: requiredRoles,
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

// GetManifestFile extracts the Cumulocity microservice manifest file from a given zip file
func GetManifestFile(zipFilename string) (string, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "c8ygo-")

	if err != nil {
		return "", fmt.Errorf("cannot create temporary file. %s", err)
	}

	files, err := zipUtilities.UnzipFile(zipFilename, tempDir, []string{"Cumulocity.json"})
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("missing Cumulocity.json file")
	}
	return files[0], err
}
