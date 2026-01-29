package enroll

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/progressbar"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/certutil"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// DeviceEnrollCmd command
type DeviceEnrollCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeviceEnrollCmd enrolls a device with Cumulocity using the Certificate Authority feature
func NewDeviceEnrollCmd(f *cmdutil.Factory) *DeviceEnrollCmd {
	ccmd := &DeviceEnrollCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "enroll",
		Short: "Enroll a device using the Cumulocity Certificate Authority",
		Long: heredoc.Doc(`
			Register a device using the Cumulocity Certificate Authority which repeatedly tries to download the
			device's certificate by submitting a Certificate Signing Request via the EST protocol.

			The registration url and QR code is printed on the console to enable users to register the device
			via a web browser.

			This feature requires the feature toggle, "certificate-authority"
		`),
		Example: heredoc.Doc(`
			$ c8y devices enroll --id "ASDF098SD1J10912UD92JDLCNCU8"
			Enroll a new device with a randomized one-time password

			$ c8y devices enroll --id "ASDF098SD1J10912UD92JDLCNCU8" --one-time-password "RqzwJeTusABlk4)KmtIc"
			Enroll a new device and provide the one-time-password to be used for enrollment

			$ c8y devices enroll --id "ASDF098SD1J10912UD92JDLCNCU8" --host example.cumulocity.com
			Enroll a new device and specify a host name so a session does not need to be set

			$ c8y devices enroll --id "ASDF098SD1J10912UD92JDLCNCU8" --key myname.key --cert myname.crt
			Enroll a new device and specify the names of the private key and public certificate to use

			$ c8y util repeat 3 | c8y devices enroll --template "{id: 'device' + input.index}"
			Enroll 2 devices and create unique private key and certificate per device

			$ DEVICE_ID=example
			$ c8y devices enroll --id "$DEVICE_ID"
			$ mosquitto_sub --key "${DEVICE_ID}.key" --cert "${DEVICE_ID}.crt" -t 's/ds' -i "$DEVICE_ID" -h $C8Y_DOMAIN -p 8883 --cafile "$(brew --prefix)/etc/ca-certificates/cert.pem" --debug
			$ mosquitto_sub --key "${DEVICE_ID}.key" --cert "${DEVICE_ID}.crt" --cafile "$(brew --prefix)/etc/ca-certificates/cert.pem" -i "$DEVICE_ID" -h $C8Y_DOMAIN -p 9883 -t 'custom/topic' --debug
			Enroll a device and use the certificate to connect to Cumulocity via MQTT (with mosquitto_sub)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)")
	cmd.Flags().String("type", "", "Device type to register (only works when pre-registering the device)")
	cmd.Flags().String("one-time-password", "", "One Time Password used for initial enrollment. Leave blank for a randomly generated password")
	cmd.Flags().Duration("retry-every", 5*time.Second, "Polling interval to try to download the device certificate")

	cmd.Flags().Bool("overwrite", false, "Overwrite any existing device key and certificate")
	cmd.Flags().Bool("show-qr", false, "Show QR Code with the registration url")
	cmd.Flags().Bool("show-url", true, "Show URL with the registration url")
	cmd.Flags().String("mode", "", "Registration mode")
	cmd.Flags().String("key", "", "Device's private certificate. If it does not exist it will be created")
	cmd.Flags().String("cert", "", "Path to write the downloaded certificate to")
	cmd.Flags().String("csr", "", "Use the given certificate signing request instead generating one")
	cmd.Flags().String("host", "", "Custom Cumulocity host")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("mode", "auto\tTry pre-registering the device if credentials are found", "manual\tForce manual registration via the registration url/QR Code"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "externalId", "name", "id"),

		// Don't require prompts
		flags.WithSemanticMethod("GET"),
	)

	// Required flags

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeviceEnrollCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	llog, err := n.factory.Logger()
	if err != nil {
		return err
	}
	_ = llog

	c8yclient, err := n.factory.Client()
	if err != nil {
		return err
	}

	consol, err := n.factory.Console()
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
		flags.WithStringValue("id"),
		flags.WithStringValue("host"),
		flags.WithStringValue("key"),
		flags.WithStringValue("cert"),
		flags.WithStringValue("csr"),
		flags.WithStringValue("type"),
		flags.WithStringValue("mode"),
		flags.WithBoolValue("overwrite"),
		flags.WithDefaultBoolValue("show-qr"),
		flags.WithDefaultBoolValue("show-url"),
		flags.WithStringValue("one-time-password"),
		flags.WithDuration("retry-every"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
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

	showUserMessage := func(format string, a ...any) {
		cfg.Logger.Infof(format, a...)
	}

	prog := progressbar.NewCountdownProgressBar(n.factory.IOStreams.ErrOut, cfg.RequestTimeout(), "Enroll", true)

	return n.factory.RunWithGenericWorkers(cmd, inputIterators, iter, func(j worker.Job) (any, error) {
		options := gjson.ParseBytes(j.Value.([]byte))

		deviceID := options.Get("id").String()
		if cmd.Flags().Changed("id") {
			if v, err := cmd.Flags().GetString("id"); err == nil {
				deviceID = v
			}
		}

		host := options.Get("host").String()
		deviceType := options.Get("type").String()
		mode := options.Get("mode").String()
		if mode == "" {
			mode = "auto"
		}
		overwrite := options.Get("overwrite").Bool()
		showQRCode := options.Get("show-qr").Bool()
		showURL := options.Get("show-url").Bool()
		keyFile := options.Get("key").String()
		csrFile := options.Get("csr").String()
		certFile := options.Get("cert").String()
		oneTimePassword := options.Get("one-time-password").String()

		if keyFile == "" {
			keyFile = fmt.Sprintf("%s.key", deviceID)
		}
		if certFile == "" {
			certFile = fmt.Sprintf("%s.crt", deviceID)
		}

		if oneTimePassword == "" {
			if v, err := c8yclient.DeviceEnrollment.GenerateOneTimePassword(); err == nil {
				// URL's with "." may not be clickable on some consoles so avoid it
				oneTimePassword = strings.ReplaceAll(v, ".", "_")
			}
		}

		if host == "" {
			host = c8yclient.GetHostname()
		}

		if host == "" {
			return "", fmt.Errorf("host is not set")
		}

		// TODO: Allow users to change the host name
		if err := c8yclient.SetBaseURL(host); err != nil {
			return "", err
		}

		clientHasCredentials := c8yclient.GetHostname() != "" && (c8yclient.Token != "" || (c8yclient.Username != "" && c8yclient.Password != ""))
		preRegister := (mode == "auto" && clientHasCredentials)

		if clientHasCredentials {
			if resp, err := c8yclient.DeviceCredentials.Delete(context.Background(), deviceID); err != nil {
				if !resp.IsDryRun() && (resp.StatusCode() != 401 && resp.StatusCode() != 403 && resp.StatusCode() != 404) {
					cfg.Logger.Infof("Could not remove existing device registration request. externalID=%s", deviceID)
				}
			} else {
				cfg.Logger.Infof("Removed existing device registration request. externalID=%s", deviceID)
			}
		}

		if preRegister {
			// Check if the correct credentials are available or not
			buf := bytes.NewBufferString("")
			c8y.BulkRegistrationRecordWriter(buf,
				c8y.BulkRegistrationRecord{
					ID:            deviceID,
					Name:          deviceID,
					Type:          deviceType,
					AuthType:      c8y.BulkRegistrationAuthTypeCertificates,
					EnrollmentOTP: oneTimePassword,
					IsAgent:       true,
				},
			)
			if _, resp, err := c8yclient.DeviceCredentials.CreateBulk(context.Background(), buf); err != nil {
				preRegister = false
				if resp != nil {
					if resp.StatusCode() != 401 && resp.StatusCode() != 403 && resp.StatusCode() != 404 {
						cfg.Logger.Warnf("Could not pre-register device. statusCode=%s, error=%s", resp.Status(), err)
					}
				} else {
					cfg.Logger.Warnf("Could not pre-register device. error=%s", err)
				}
			}
		}

		timeout := cfg.RequestTimeout()
		retryEvery := time.Duration(options.Get("retry-every").Int())

		// Create client that does not use any authentication
		client := c8y.NewClientFromOptions(nil, c8y.ClientOptions{
			BaseURL:       host,
			Realtime:      false,
			ShowSensitive: true,
		})

		client.SetRequestOptions(c8y.DefaultRequestOptions{
			DryRun: cfg.DryRun(),
			DryRunHandler: func(options *c8y.RequestOptions, req *http.Request) {
				handler := &request.RequestHandler{
					IsTerminal: n.factory.IOStreams.IsStdoutTTY(),
					IO:         n.factory.IOStreams,
					Client:     client,
					Config:     cfg,
					Logger:     llog,
					Console:    consol,
					HideSensitive: func(c *c8y.Client, s string) string {
						return s
					},
				}
				handler.DryRunHandler(n.factory.IOStreams, options, req)
			},
		})

		if overwrite {
			cfg.Logger.Info("Removing any existing private key, or certificate")
			if err := os.Remove(keyFile); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					cfg.Logger.Warnf("Failed to remove private key. file=%s, error=%s", keyFile, err)
				}
			}
			if err := os.Remove(certFile); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					cfg.Logger.Warnf("Failed to remove certificate. file=%s, error=%s", certFile, err)
				}
			}
		}

		// Create private key
		keyPem, keyWasGenerated, err := certutil.LoadOrGenerateKeyFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load or create private key. %w", err)
		}
		if keyWasGenerated {
			cfg.Logger.Infof("Created a new private key. file=%s", keyFile)
		} else {
			cfg.Logger.Infof("Loaded an existing private key. file=%s", keyFile)
		}

		key, err := certutil.ParsePrivateKeyPEM(keyPem)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key. %w", err)
		}

		wasExisting := true

		var cert *x509.Certificate
		if _, err := os.Stat(certFile); overwrite || errors.Is(err, os.ErrNotExist) {

			showUserMessage("\n📣 Starting device enrollment: externalID=%s\n", deviceID)

			// Create/Load CSR
			csr, err := LoadCertificateSigningRequest(
				LoadCSRFromFile(csrFile),
				GenerateCSRFromKey(client, deviceID, key),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create certificate signing request. %w", err)
			}

			ctx := context.Background()

			initDelay := 2 * time.Second
			if cfg.DryRun() || preRegister {
				initDelay = 0
			}

			// Enroll device
			if !preRegister {
				prog.Start()
			}

			result := <-client.DeviceEnrollment.PollEnroll(ctx, c8y.DeviceEnrollmentOption{
				ExternalID:      deviceID,
				OneTimePassword: oneTimePassword, // Generate random one-time password if empty

				// Initial delay before the first download attempt
				InitDelay: initDelay,

				// Check every 5 seconds
				Interval: retryEvery,

				// Give up after 10 minutes
				Timeout: timeout,

				// Print enrollment information
				Banner: &c8y.DeviceEnrollmentBannerOptions{
					Enable:     !preRegister,
					ShowQRCode: showQRCode,
					ShowURL:    showURL,
				},

				CertificateSigningRequest: csr,
			})
			if result.Err != nil {
				showUserMessage("🚫 Failed to download the device's certificate\n")
				return nil, result.Err
			}

			if result.Certificate != nil {
				showUserMessage("✅ Successfully downloaded the device's certificate. key=%s, cert=%s\n", keyFile, certFile)
				cert = result.Certificate
				certPEM := certutil.MarshalCertificateToPEM(cert.Raw)
				os.WriteFile(certFile, certPEM, 0644)
			}
			wasExisting = false
		} else {
			certPEM, err := os.ReadFile(certFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read certificate file. %w", err)
			}

			cert, err = certutil.ParseCertificatePEM(certPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate. %w", err)
			}
			showUserMessage("\n📣 Using existing device certificate: externalID=%s, key=%s, cert=%s\n", cert.Subject.CommonName, keyFile, certFile)
		}

		result := EnrollmentResult{
			Certificate:       certFile,
			Key:               keyFile,
			ID:                deviceID,
			DeviceType:        deviceType,
			AlreadyRegistered: wasExisting,
		}

		outB, outErr := json.Marshal(result)
		if outErr != nil {
			return nil, outErr
		}

		n.factory.WriteOutput(outB, cmdutil.OutputContext{
			Input: j.Input,
		}, &commonOptions)

		return nil, nil
	})
}

type EnrollmentResult struct {
	Certificate       string `json:"certificate,omitempty"`
	Key               string `json:"key,omitempty"`
	ID                string `json:"id,omitempty"`
	DeviceType        string `json:"deviceType,omitempty"`
	AlreadyRegistered bool   `json:"alreadyRegistered"`
}

type CSRSource func() (*x509.CertificateRequest, bool, error)

func LoadCertificateSigningRequest(opts ...CSRSource) (*x509.CertificateRequest, error) {
	for _, opt := range opts {
		csr, ok, err := opt()
		if ok {
			return csr, err
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("no csr was generated")
}

func LoadCSRFromFile(path string) CSRSource {
	return func() (*x509.CertificateRequest, bool, error) {
		if path == "" {
			return nil, false, nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, false, err
		}
		block, _ := pem.Decode(contents)
		if block == nil {
			return nil, false, err
		}
		csr, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			return nil, false, err
		}
		return csr, true, nil
	}
}

func GenerateCSRFromKey(client *c8y.Client, deviceID string, key any) CSRSource {
	return func() (*x509.CertificateRequest, bool, error) {
		csr, err := client.DeviceEnrollment.CreateCertificateSigningRequest(deviceID, key)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create certificate signing request. %w", err)
		}
		return csr, true, nil
	}
}
