package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/zipUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type newMicroserviceCmd struct {
	*baseCmd

	file             string
	name             string
	key              string
	availability     string
	contextPath      string
	resourceURL      string
	skipSubscription bool
	skipUpload       bool
}

func newNewMicroserviceCmd() *newMicroserviceCmd {
	ccmd := &newMicroserviceCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "New microservice",
		Long:  ``,
		Example: `
$ c8y microservices create --file ./myapp.zip
Create new microservice
		`,
		RunE: ccmd.doProcedure,
	}

	cmd.SilenceUsage = true

	addDataFlag(cmd)
	cmd.Flags().StringVar(&ccmd.file, "file", "", "Microservice file to be uploaded (or Cumulocity.json) file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().StringVar(&ccmd.resourceURL, "resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipSubscription, "skipSubscription", false, "Skip microservice subscription when creating the new microservice")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Skip uploading the binary to the platform")

	// Required flags
	cmd.MarkFlagRequired("file")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newMicroserviceCmd) getApplicationDetails() *c8y.Application {

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

	app.Key = app.Name + "-microservice-key"
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

func (n *newMicroserviceCmd) doProcedure(cmd *cobra.Command, args []string) error {
	var application *c8y.Application
	var response *c8y.Response
	var applicationID string
	var applicationName string
	var err error

	applicationDetails := n.getApplicationDetails()

	if applicationDetails != nil {
		applicationName = applicationDetails.Name
	}

	if applicationName == "" {
		return newUserError("Could not detect application name for the given input")
	}

	if applicationName != "" {

		refs, err := findMicroservices([]string{applicationName}, true)

		if err != nil {
			return newUserError(err)
		}

		idValue, _ := getFetchedResultsAsString(refs)

		/* if err != nil {
			return newUserError("no matching microservices found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching microservices found", idInputValues)
		} */
		if idValue != nil {
			for _, item := range idValue {
				if item != "" {
					applicationID = newIDValue(item).GetID()
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if applicationID == "" {
		// Create the application
		Logger.Info("Creating new application")
		application, response, err = client.Application.Create(context.Background(), applicationDetails)

		if err != nil {
			return fmt.Errorf("failed to create microservice. %s", err)
		}
	} else {
		// Get existing application
		Logger.Infof("Getting existing application. id=%s", applicationID)
		application, response, err = client.Application.GetApplication(context.Background(), applicationID)

		if err != nil {
			return fmt.Errorf("failed to get microservice. %s", err)
		}
	}

	skipUpload := n.skipUpload

	// Only upload zip files
	if !strings.HasSuffix(n.file, ".zip") {
		Logger.Info("Skipping microservice binary upload")
		skipUpload = true
	}

	// Upload binary
	if !skipUpload {
		Logger.Infof("uploading binary [id=%s]", application.ID)
		if !globalFlagDryRun {
			_, err := client.Application.CreateBinary(context.Background(), n.file, application.ID)

			if err != nil {
				// handle error
				n.cmd.PrintErrf("failed to upload file. %s", err)
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
				Logger.Warningf("failed to get manifest file from microservice. %s", err)
			} else {
				manifestFile = val
			}
		}

		if v, err := jsonUtilities.DecodeJSONFile(manifestFile); err == nil {
			manifestContents = v
		} else {
			Logger.Warningf("failed to decode manifest file. file=%s, err=%s", manifestFile, err)
			return newUserError(fmt.Sprintf("invalid manifest file. Only json files are accepted. %s", err))
		}

		if values, ok := manifestContents["requiredRoles"].([]string); ok {
			requiredRoles = append(requiredRoles, values...)
		}

		// Read the Cumulocity.json file, and upload
		Logger.Infof("updating application details [id=%s], requiredRoles=%s", application.ID, strings.Join(requiredRoles, ","))
		if !globalFlagDryRun {
			_, response, err = client.Application.Update(context.Background(), application.ID, &c8y.Application{
				RequiredRoles: requiredRoles,
				// manifest: manifestContents,
			})
		}
	}

	// App subscription
	if !n.skipSubscription {
		Logger.Infof("Subscribing to application")
		if !globalFlagDryRun {
			_, resp, err := client.Tenant.AddApplicationReference(context.Background(), client.TenantName, application.Self)

			if err != nil {
				if resp != nil && resp.StatusCode == 409 {
					Logger.Infof("microservice is already enabled")
				} else {
					return fmt.Errorf("Failed to subscribe to application. %s", err)
				}
			}
		}
	}

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	return processResponse(response, err, commonOptions)
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
