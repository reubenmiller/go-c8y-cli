package enroll

import (
	"bytes"
	"cmp"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/mdp/qrterminal/v3"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/progressbar"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices/enrollment"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/certutil"
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
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputFlag("id"); err != nil {
		return err
	}
	if err := r.Body(
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
	); err != nil {
		return cmderrors.NewUserError(err)
	}
	session, err := r.Client()
	if err != nil {
		return err
	}
	cfg := r.Config
	errOut := n.factory.IOStreams.ErrOut
	prog := progressbar.NewCountdownProgressBar(errOut, cfg.RequestTimeout(), "Enroll", n.factory.IOStreams.IsStderrTTY())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		raw, err := in.Body()
		if err != nil {
			return nil, err
		}
		options := gjson.ParseBytes(raw)

		// --id wins over the template, which wins over the piped value
		deviceID := in.String("id")
		if v := options.Get("id").String(); v != "" && !cmd.Flags().Changed("id") {
			deviceID = v
		}
		if deviceID == "" {
			return nil, cmderrors.NewUserError("missing device id")
		}

		keyFile := cmp.Or(options.Get("key").String(), deviceID+".key")
		certFile := cmp.Or(options.Get("cert").String(), deviceID+".crt")
		csrFile := options.Get("csr").String()
		// the files are read or written, so they can't be in a denied location
		for _, p := range []string{keyFile, csrFile, certFile} {
			if err := fileguard.Check(p); err != nil {
				return nil, err
			}
		}
		overwrite := options.Get("overwrite").Bool()

		// A device with a certificate is enrolled already
		if !overwrite {
			if certPEM, err := os.ReadFile(certFile); err == nil {
				cert, err := certutil.ParseCertificatePEM(certPEM)
				if err != nil {
					return nil, fmt.Errorf("failed to parse certificate. file=%s, %w", certFile, err)
				}
				cfg.Logger.Infof("Using the existing device certificate. externalID=%s, key=%s, cert=%s", cert.Subject.CommonName, keyFile, certFile)
				result := enrollmentResult(deviceID, options.Get("type").String(), keyFile, certFile, true)
				return func(context.Context) any { return result }, nil
			}
		}

		client, enrollOpt, err := enrollOptions(session, cfg.GetHost() != "", deviceID, options, cfg.RequestTimeout())
		if err != nil {
			return nil, err
		}

		// The certificate signing request is given, or created with the
		// device's key (an existing one, or a new one)
		var newKey []byte
		if csrFile != "" {
			if enrollOpt.CSR, err = loadCSR(csrFile); err != nil {
				return nil, err
			}
		} else {
			keyPEM, err := os.ReadFile(keyFile)
			if overwrite || errors.Is(err, os.ErrNotExist) {
				if keyPEM, err = certutil.MakeEllipticPrivateKeyPEM(); err != nil {
					return nil, err
				}
				newKey = keyPEM
			} else if err != nil {
				return nil, fmt.Errorf("failed to read private key. file=%s, %w", keyFile, err)
			}
			key, err := certutil.ParsePrivateKeyPEM(keyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key. file=%s, %w", keyFile, err)
			}
			if enrollOpt.CSR, err = client.Devices.Enrollment.CreateCertificateSigningRequest(deviceID, key); err != nil {
				return nil, fmt.Errorf("failed to create certificate signing request. %w", err)
			}
		}

		banner := enrollment.NewDeviceEnrollmentBannerOptions(options.Get("show-qr").Bool(), options.Get("show-url").Bool(), renderQRCode)
		enrollOpt.OnRegistrationRequired = func(req devices.RegistrationRequired) {
			if req.Err != nil {
				logRegistrationError(cfg.Logger, req.Err)
			}
			if err := client.Devices.Enrollment.WriteBanner(errOut, req.ExternalID, req.OneTimePassword, *banner); err != nil {
				cfg.Logger.Warnf("Could not show the registration url. %s", err)
			}
			prog.Start()
		}
		enrollOpt.OnAttemptError = func(err error) {
			cfg.Logger.Debugf("The device's certificate is not available yet. externalID=%s, error=%s", deviceID, err)
		}

		return func(ctx context.Context) any {
			cfg.Logger.Infof("Starting device enrollment. externalID=%s", deviceID)
			result := client.Devices.Enroll(ctx, enrollOpt)
			if result.Err != nil || result.Data.Certificate == nil {
				// an error, or a dry run
				return result
			}
			// The key is saved with the certificate, so a failed enrollment leaves no files
			if newKey != nil {
				if err := certutil.WriteKey(keyFile, newKey); err != nil {
					return fmt.Errorf("failed to write private key. file=%s, %w", keyFile, err)
				}
			}
			if err := os.WriteFile(certFile, certutil.MarshalCertificateToPEM(result.Data.Raw), 0644); err != nil {
				return fmt.Errorf("failed to write certificate. file=%s, %w", certFile, err)
			}
			cfg.Logger.Infof("Downloaded the device's certificate. key=%s, cert=%s", keyFile, certFile)
			if csrFile != "" && options.Get("key").String() == "" {
				// The key of a given signing request is unknown
				keyFile = ""
			}
			return enrollmentResult(deviceID, enrollOpt.Type, keyFile, certFile, false)
		}, nil
	})
}

// enrollOptions returns the client which enrolls the device and the options
// to enroll it. A device is enrolled with the session's tenant, or with --host.
// The session's credentials are only used for the session's host: there they
// remove an old device request and, in auto mode, register the device.
func enrollOptions(session *apiv2.Client, hasSession bool, deviceID string, options gjson.Result, timeout time.Duration) (*apiv2.Client, devices.EnrollOptions, error) {
	opt := devices.EnrollOptions{
		ExternalID:      deviceID,
		Type:            options.Get("type").String(),
		OneTimePassword: options.Get("one-time-password").String(),
		Interval:        time.Duration(options.Get("retry-every").Int()),
		Timeout:         timeout,
	}

	client := session
	if host := options.Get("host").String(); host != "" {
		u, err := url.Parse(apiv2.FormatBaseURL(host))
		if err != nil || u.Host == "" {
			return nil, opt, cmderrors.NewUserError(fmt.Sprintf("invalid host. %s", host))
		}
		if !hasSession || u.Host != session.BaseURL.Host {
			// Another tenant: enroll without credentials
			client = apiv2.NewClient(apiv2.ClientOptions{BaseURL: u.String(), Timeout: opt.Timeout})
			hasSession = false
		}
	} else if !hasSession {
		return nil, opt, cmderrors.NewUserError("host is not set. Use a session or --host")
	}

	auth := session.Auth
	hasCredentials := hasSession && (auth.Token != "" || (auth.Username != "" && auth.Password != ""))
	opt.RemoveRequest = hasCredentials
	opt.Register = hasCredentials && cmp.Or(options.Get("mode").String(), "auto") == "auto"
	return client, opt, nil
}

// logRegistrationError logs why the device could not be registered, so a user
// has to register it. Missing permissions are expected.
func logRegistrationError(log interface {
	Infof(string, ...any)
	Warnf(string, ...any)
}, err error) {
	var apiErr *core.Error
	if errors.As(err, &apiErr) && (apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusForbidden || apiErr.Code == http.StatusNotFound) {
		log.Infof("Could not register the device. %s", err)
		return
	}
	log.Warnf("Could not register the device. %s", err)
}

func renderQRCode(url string) (string, error) {
	var b bytes.Buffer
	qrterminal.GenerateHalfBlock(url, qrterminal.L, &b)
	return b.String(), nil
}

func loadCSR(path string) (*x509.CertificateRequest, error) {
	contents, err := fileguard.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(contents)
	if block == nil {
		return nil, fmt.Errorf("the certificate signing request is not PEM encoded. file=%s", path)
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate signing request. file=%s, %w", path, err)
	}
	return csr, nil
}

func enrollmentResult(id, deviceType, key, cert string, existing bool) output.Seq {
	b, _ := json.Marshal(EnrollmentResult{
		Certificate:       cert,
		Key:               key,
		ID:                id,
		DeviceType:        deviceType,
		AlreadyRegistered: existing,
	})
	return c8ystream.FromDocs(jsondoc.New(b))
}

type EnrollmentResult struct {
	Certificate       string `json:"certificate,omitempty"`
	Key               string `json:"key,omitempty"`
	ID                string `json:"id,omitempty"`
	DeviceType        string `json:"deviceType,omitempty"`
	AlreadyRegistered bool   `json:"alreadyRegistered"`
}
