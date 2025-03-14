package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// RegisterESTCmd command
type RegisterESTCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRegisterESTCmd creates a command to Register device via the Cumulocity CA
func NewRegisterESTCmd(f *cmdutil.Factory) *RegisterESTCmd {
	ccmd := &RegisterESTCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "register-ca",
		Short: "Register device via the Cumulocity CA (private preview feature)",
		Long: heredoc.Doc(`
			Register a new device using the bulk registration api

			This feature requires the private preview feature toggle, "certificate-authority"
		`),
		Example: heredoc.Doc(`
			$ c8y deviceregistration bulk register-ca --id "ASDF098SD1J10912UD92JDLCNCU8"
			Register a new device using BASIC authentication and generate a random password (printed on the console)

			$ c8y deviceregistration bulk register-ca --id "ASDF098SD1J10912UD92JDLCNCU8" --one-time-password "example"
			Register a new device and provide the one-time-password to be used for enrollment

			$ echo -e "device1\ndevice2" | c8y deviceregistration bulk register-ca --type linux --template "{name: input.value}"
			Register 2 devices, and set the names based on their external id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)")
	cmd.Flags().String("name", "", "Device name. Defaults to the external id")
	cmd.Flags().String("type", "thin-edge.io", "Device type")
	cmd.Flags().String("iccid", "", "The ICCID of the device (SIM card number). If the ICCID appears in file, the import adds a fragment c8y_Mobile.iccid.")
	cmd.Flags().String("external-type", "c8y_Serial", "The type of the external ID. If IDTYPE doesn't appear in the file, the default value is used. The default value is c8y_Serial")
	cmd.Flags().String("one-time-password", "", "One Time Password used for initial enrollment. Leave blank for a randomly generated password")
	cmd.Flags().String("tenant", "", "The ID of the tenant for which the registration is executed (only allowed for the management tenant)")
	cmd.Flags().String("group", "", "The path in the groups hierarchy where the device is added. PATH contains the name of each group separated by /, that is: main_group/sub_group/.../last_sub_group. If a group does not exist, the import creates the group")

	completion.WithOptions(
		cmd,
		completion.WithRootDeviceGroup("group", func() (*c8y.Client, error) { return f.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "externalId", "name", "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *RegisterESTCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	llog, err := n.factory.Logger()
	if err != nil {
		return err
	}

	c8yclient, err := n.factory.Client()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithOverrideValue("id", "id"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("external-type", "external-type"),
		flags.WithStringValue("iccid", "iccid"),
		flags.WithStringValue("one-time-password", "otp"),
		flags.WithStringValue("tenant", "tenant"),
		flags.WithStringValue("group", "group"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithStringValue("id", "id", ""),
	)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	var iter iterator.Iterator
	if inputIterators.Total > 0 {
		iter = mapbuilder.NewMapBuilderIterator(body)
	} else {
		iter = iterator.NewBoundIterator(mapbuilder.NewMapBuilderIterator(body), 1)
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	commonOptions.DisableResultPropertyDetection()

	return n.factory.RunWithGenericWorkers(cmd, inputIterators, iter, func(j worker.Job) (any, error) {

		options := gjson.ParseBytes(j.Value.([]byte))

		formData := make(map[string]io.Reader)
		b := bytes.NewBufferString("")

		externalID := options.Get("id").String()
		externalIDType := options.Get("external-type").String()
		iccid := options.Get("iccid").String()
		tenant := options.Get("tenant").String()
		groupPath := options.Get("group").String()

		deviceName := options.Get("name").String()
		if deviceName == "" {
			deviceName = externalID
		}
		deviceType := options.Get("type").String()
		deviceCredentials := options.Get("otp").String()

		showPassword := false
		if deviceCredentials == "" {
			// Show the password to the user (as they need this when connecting the device)
			deviceCredentials = randdata.Password(32)
			showPassword = true
		}

		passwordWarningChars := "\""
		if strings.ContainsAny(deviceCredentials, passwordWarningChars) {
			llog.Warnf("Device password contains some unsupported characters [%s]. Please avoid using any of them", passwordWarningChars)
		}

		// Cumulocity CA uses a one-time password header
		writeCSV(b, []KeyValuePair{
			{"ID", externalID},
			{"AUTH_TYPE", "CERTIFICATES"},
			{"ENROLLMENT_OTP", deviceCredentials},
			{"NAME", deviceName},
			{"TYPE", deviceType},
			{"IDTYPE", externalIDType},
			{"ICCID", iccid},
			{"TENANT", tenant},
			{"PATH", groupPath},
			{"com_cumulocity_model_Agent.active", "true"},
		})

		formData["file"] = b

		req := c8y.RequestOptions{
			Method:       http.MethodPost,
			Path:         "devicecontrol/bulkNewDeviceRequests",
			Accept:       "application/json",
			FormData:     formData,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}

		response, responseErr := c8yclient.SendRequest(context.Background(), req)
		if responseErr != nil {
			return response, responseErr
		}

		// dry run
		if response == nil {
			return "", nil
		}

		// TODO: inspect the response to see if the bulk registration was successful
		body := response.Body()
		totalFailed := gjson.GetBytes(body, "numberOfFailed").Int()
		if totalFailed != 0 {
			llog.Infof("Response: %v", response)
			failuresReasons := make([]string, 0)
			response.JSON("failedCreationList").ForEach(func(key, value gjson.Result) bool {
				if v := value.Get("failureReason"); v.Exists() {
					if reason := v.String(); reason != "" {
						failuresReasons = append(failuresReasons, fmt.Sprintf("id=%s, reason=%s", value.Get("deviceId").String(), reason))
					}
				}
				return true
			})
			return response, cmderrors.NewUserError(fmt.Sprintf("bulk registration has some failures. failed=%d, reasons=%v", totalFailed, failuresReasons))
		}
		llog.Infof("Bulk registration was successful. %v", response)

		// Lookup device id so that the command can be piped to downstream items
		identity, _, identityErr := c8yclient.Identity.GetExternalID(context.Background(), externalIDType, externalID)
		if identityErr != nil {
			return "", identityErr
		}

		// Build a response to return to the user (this is not the response receive from c8y)
		output := map[string]any{}
		output["id"] = identity.ManagedObject.ID
		output["externalId"] = externalID
		output["name"] = deviceName
		output["username"] = fmt.Sprintf("device_%s", externalID)
		if showPassword {
			// Use password so the output is normalized with the register-basic output
			output["password"] = deviceCredentials
		}
		output["type"] = deviceType

		outB, jsonErr := json.Marshal(output)
		if jsonErr != nil {
			return "", jsonErr
		}

		contentType := response.Response.Header.Get("Content-Type")
		llog.Infof("API Content-Type: %s", contentType)

		err := n.factory.WriteOutput(outB, cmdutil.OutputContext{
			Input:    j.Input,
			Response: response.Response,
		}, &commonOptions)
		return nil, err
	})
}
