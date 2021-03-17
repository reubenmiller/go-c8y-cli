package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/zipUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type newHostedApplicationCmd struct {
	*baseCmd

	file           string
	name           string
	key            string
	availability   string
	contextPath    string
	resourcesURL   string
	skipActivation bool
	skipUpload     bool
}

func NewNewHostedApplicationCmd() *newHostedApplicationCmd {
	ccmd := &newHostedApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "createHostedApplication",
		Short: "Create hosted application",
		Long:  `Create a new hosted web application or update the binary of an existing hosted application`,
		Example: heredoc.Doc(`
$ c8y applications createHostedApplication --file ./myapp.zip

Create new hosted application from a given zip file
		`),
		PreRunE: validateCreateMode,
		RunE:    ccmd.doProcedure,
	}

	cmd.SilenceUsage = true

	addDataFlag(cmd)
	cmd.Flags().StringVar(&ccmd.file, "file", "", "Web application zip file to be uploaded")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application. Defaults to the name")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application")
	cmd.Flags().StringVar(&ccmd.resourcesURL, "resourcesUrl", "/", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipActivation, "skipActivation", false, "Skip microservice subscription when creating the new microservice")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Skip uploading the binary to the platform")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newHostedApplicationCmd) getApplicationDetails() *c8y.Application {

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

	app.Key = app.Name + "-application-key"
	if n.key != "" {
		app.Key = n.key
	}

	app.Type = "HOSTED"

	if n.availability != "" {
		app.Availability = n.availability
	}

	app.ContextPath = app.Name
	if n.contextPath != "" {
		app.ContextPath = n.contextPath
	}

	app.ResourcesURL = "/"
	if n.resourcesURL != "" {
		app.ResourcesURL = n.resourcesURL
	}
	return &app
}

// packageWebApplication zips the given folder path to a zip
func (n *newHostedApplicationCmd) packageWebApplication(src string) (string, error) {
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
func (n *newHostedApplicationCmd) packageAppIfRequired(src string) (zipFile string, err error) {
	f, err := os.Stat(src)

	if err != nil {
		return
	}

	if !f.IsDir() {
		// it is already a zip
		zipFile = src
		return
	}

	Logger.Infof("zipping folder %s", src)
	zipFile, err = n.packageWebApplication(src)

	if err != nil {
		err = fmt.Errorf("failed to package web application. %s", err)
	}
	return
}

func (n *newHostedApplicationCmd) doProcedure(cmd *cobra.Command, args []string) error {
	var application *c8y.Application
	var response *c8y.Response
	var applicationID string
	var err error

	appDetails := n.getApplicationDetails()

	// TODO: Use the default name value from n.Name rather then reading it from the args again.
	Logger.Infof("application name: %s", appDetails.Name)
	if appDetails.Name != "" {
		refs, err := findHostedApplications([]string{appDetails.Name}, true)

		if err != nil {
			return fmt.Errorf("Failed to find hosted application. %s", err)
		}

		if err == nil && len(refs) > 0 {
			applicationID = refs[0].ID
		}
	}

	if applicationID == "" {
		// Create the application
		Logger.Info("Creating new application")
		application, response, err = client.Application.Create(context.Background(), appDetails)

		if err != nil {
			return fmt.Errorf("Failed to create microservice. %s", err)
		}
		applicationID = application.ID
	} else {
		// Get existing application
		Logger.Infof("Getting existing application. id=%s", applicationID)
		application, response, err = client.Application.GetApplication(context.Background(), applicationID)

		if err != nil {
			return fmt.Errorf("Failed to get microservice. %s", err)
		}
	}

	skipUpload := n.skipUpload || n.file == ""

	// Upload binary
	applicationBinaryID := ""
	if !skipUpload {

		zipfile, err := n.packageAppIfRequired(n.file)
		if err != nil {
			Logger.Errorf("Failed to package file. %s", err)
			return fmt.Errorf("failed to package app. %s", err)
		}

		Logger.Infof("uploading binary [app=%s]", application.ID)
		resp, err := client.Application.CreateBinary(context.Background(), zipfile, application.ID)

		if err != nil {
			// handle error
			n.cmd.PrintErrf("failed to upload file. %s", err)
		} else {
			applicationBinaryID = resp.JSON.Get("id").String()
		}
	}

	// App activation (only if a new version was uploaded)
	if !skipUpload && !n.skipActivation {
		Logger.Infof("Activating application")

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
			if resp != nil && resp.StatusCode == 409 {
				Logger.Infof("application is already enabled")
			} else {
				return fmt.Errorf("failed to activate application. %s", err)
			}
		}

		// use the updated application json
		response = resp
	}

	commonOptions, err := cliConfig.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	_, err = processResponse(response, err, commonOptions)
	return err
}
