package remoteaccess

import (
	"context"
	"fmt"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/rasession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdServer struct {
	device        []string
	listen        string
	configuration string
	open          bool
	browserScheme string
	multiplex     bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdServer(f *cmdutil.Factory) *CmdServer {
	ccmd := &CmdServer{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a local proxy server",
		Long: heredoc.Docf(`
			Start a local proxy server

			You can add use the remote access local proxy within your ssh config file, to use it to
			connect to your device with ssh without having to manually launch the proxy yourself!

			To do this add the following configuration to your device.

			%[1]s
			Host [device]
				User [device_username]
				PreferredAuthentications publickey
				IdentityFile [identify_file]
				ServerAliveInterval 120
				StrictHostKeyChecking no
				UserKnownHostsFile /dev/null
				ProxyCommand c8y remoteaccess server --device %%n --listen -
			%[1]s

			Note: When using the "--browser" flag, by default the URL scheme (e.g. http, https) will be
			auto detected based on the Remote Access configuration's name. For example, if the configuration
			name has the "https:" (or e.g. "https+mux:") prefix, then https will be used, otherwise http will be used. This aligns
			with the naming convention used in https://github.com/Cumulocity-IoT/cumulocity-remote-access-cloud-http-proxy
		`, "````"),
		Example: heredoc.Doc(`
			$ c8y remoteaccess server --device 12345
			Start a local proxy server on a random local port

			$ c8y remoteaccess server --device 12345 --listen -
			Start a local proxy server reading from stdin and writing to stdout (useful for usage with the ProxyCommand in ssh)

			$ c8y remoteaccess server --device 12345 --listen unix:///run/example.sock
			Start a local proxy using a UNIX socket

			$ c8y remoteaccess server --device 12345 --listen 127.0.0.1:22022
			Start a local proxy using a local TCP server on a fixed port 22022

			$ c8y remoteaccess server --device 12345 --configuration "*rugpi*" --browser
			Start a local proxy and match on the configuration using wildcards, then open the browser to the endpoint

			$ c8y remoteaccess server --device 12345 --configuration "*rugpi*" --browser --scheme https
			Start a local proxy and match on the configuration using wildcards, then open the browser to the endpoint and force usage of https

			$ c8y remoteaccess server --device 12345 --configuration "http:*" --browser --multiplex
			Start a local proxy which carries all connections over a single remote access session (if supported by the device), e.g. to speed up loading a web application

			$ c8y remoteaccess server --device 12345 --configuration "https+mux:*" --browser
			Start a local proxy for a configuration named with the mux option, which is multiplexed automatically
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listen. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration. Accepts wildcards")
	cmd.Flags().BoolVar(&ccmd.open, "browser", false, "Open the endpoint in a browser (if available)")
	cmd.Flags().StringVar(&ccmd.browserScheme, "scheme", "auto", "URL scheme to use when opening the address in a browser, e.g. http, https or auto")
	cmd.Flags().BoolVar(&ccmd.multiplex, "multiplex", false, "Carry all local connections over a single remote access session (yamux), which avoids the setup cost of a new session per connection. Requires a device supporting it (e.g. thin-edge.io), otherwise one session per connection is used. Enabled by default for configurations named with the mux option, e.g. \"http+mux:example\"")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("scheme", "http", "https", "auto"),
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithRemoteAccessPassthroughConfiguration("configuration", "device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE starts a local proxy to each device's remote access passthrough
// configuration, and serves it until the command is stopped.
func (n *CmdServer) RunE(cmd *cobra.Command, args []string) error {
	// In stdio mode stdin carries the proxied stream, not pipeline input
	if n.listen == "-" {
		cfg, err := n.factory.Config()
		if err != nil {
			return err
		}
		cfg.SetDisableStdin(true)
	}

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputFlag("device"); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	opts := rasession.Options{Configuration: n.configuration}
	if cmd.Flags().Changed("multiplex") {
		opts.Multiplex = &n.multiplex
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		deviceID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}
		opts := opts
		opts.DeviceID = deviceID
		config, err := rasession.Detect(in.ResolveContext(), client, opts)
		if err != nil {
			return nil, err
		}

		return func(ctx context.Context) any {
			session := rasession.New(client, opts, config)
			log.Debugf("Using remote access configuration: id=%s, name=%s, multiplex=%v", config.ID(), config.Name(), session.Multiplex)

			if n.listen == "-" {
				log.Debugf("Listening to request from stdin")
				return session.ListenServe(n.factory.IOStreams.In, n.factory.IOStreams.Out)
			}

			// TCP / socket listener
			if err := session.Listen(n.listen); err != nil {
				return err
			}
			host, port := session.HostPort()
			scheme := session.BrowserScheme(n.browserScheme)

			t := template.Must(template.New("message").Parse(serverMessage))
			_ = t.Execute(n.factory.IOStreams.ErrOut, serverInfo{
				Host:          host,
				Port:          port,
				CumulocityURL: rasession.CumulocityURL(client),
				LocalAddress:  session.GetListenerAddress(),
				Device:        deviceID,
				User:          "<device_username>",
				Scheme:        scheme,
			})

			if n.open {
				go func() {
					if err := n.factory.Browser.Browse(fmt.Sprintf("%s://%s:%s", scheme, host, port)); err != nil {
						log.Warnf("%s", err)
					}
				}()
			}

			return session.Serve()
		}, nil
	})
}

type serverInfo struct {
	Port          string
	Host          string
	Device        string
	LocalAddress  string
	User          string
	CumulocityURL string
	Scheme        string
}

var serverMessage = heredoc.Doc(`
	Listening for device ({{.Device}}) {{.CumulocityURL}} on {{.LocalAddress}}

	Proxy:     {{.LocalAddress}}

	Example clients:

		SSH:        ssh -p {{.Port}} {{.User}}@{{.Host}}

		Website:    {{.Scheme}}://{{.LocalAddress}}

	Press ctrl-c to shutdown the server
`)
