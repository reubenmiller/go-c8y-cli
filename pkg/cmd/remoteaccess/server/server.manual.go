package server

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/remoteaccess"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type CmdServer struct {
	device        []string
	listen        string
	configuration string
	open          bool

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
		Long: `
		Start a local proxy server

		You can add use the remote access local proxy within your ssh config file, to use it to
		connect to your device with ssh without having to manually launch the proxy yourself!

		To do this add the following configuration to your device.

		---
		Host <device>
			User <device_username>
			PreferredAuthentications publickey
			IdentityFile <identify_file>
			ServerAliveInterval 120
			StrictHostKeyChecking no
			UserKnownHostsFile /dev/null
			ProxyCommand c8y remoteaccess server --device %n --listen -
		---
		`,
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
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listen. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration. Accepts wildcards")
	cmd.Flags().BoolVar(&ccmd.open, "browser", false, "Open the endpoint in a browser (if available)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdServer) RunE(cmd *cobra.Command, args []string) error {
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	// Disable if stdio mode is being used
	if n.listen == "-" {
		log.Debug("Disabling pipeline stdin parsing when in stdio mode")
		cfg.SetDisableStdin(true)
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	body := mapbuilder.NewInitializedMapBuilder(true)
	body.SetApplyTemplateOnMarshalPreference(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "device", "device"),
	)
	if err != nil {
		return err
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
		item := gjson.ParseBytes(j.Value.([]byte))
		device := item.Get("device").String()

		craConfig, err := c8yfetcher.DetectRemoteAccessConfiguration(client, device, n.configuration)
		if err != nil {
			return nil, err
		}

		log.Debugf("Using remote access configuration: id=%s, name=%s", craConfig.ID, craConfig.Name)

		// Lookup configuration
		craClient := remoteaccess.NewRemoteAccessClient(client, remoteaccess.RemoteAccessOptions{
			ManagedObjectID: device,
			RemoteAccessID:  craConfig.ID,
		})

		if n.listen == "-" {
			log.Debugf("Listening to request from stdin")
			serverErr := craClient.ListenServe(n.factory.IOStreams.In, n.factory.IOStreams.Out)
			return nil, serverErr
		}

		// TCP / socket listener
		if err := craClient.Listen(n.listen); err != nil {
			return nil, err
		}

		localAddress := craClient.GetListenerAddress()
		host, port, _ := strings.Cut(localAddress, ":")
		if host == "" {
			host = "127.0.0.1"
		}

		type ServerInfo struct {
			Port          string
			Host          string
			Device        string
			LocalAddress  string
			User          string
			CumulocityURL string
		}

		messageTmpl := heredoc.Doc(`
			Listening for device ({{.Device}}) {{.CumulocityURL}} on {{.LocalAddress}}
	
			Proxy:     {{.LocalAddress}}
	
			Example clients:
	
				SSH:        ssh -p {{.Port}} {{.User}}@{{.Host}}
	
				Website:    http://{{.LocalAddress}}
	
			Press ctrl-c to shutdown the server
		`)

		t := template.Must(template.New("message").Parse(messageTmpl))
		t.Execute(n.factory.IOStreams.ErrOut, ServerInfo{
			Host:          host,
			Port:          port,
			CumulocityURL: strings.TrimRight(client.BaseURL.String(), "/"),
			LocalAddress:  localAddress,
			Device:        device,
			User:          "<device_username>",
		})

		if n.open {
			go func() {
				targetURL := fmt.Sprintf("http://%s:%s", host, port)
				if err := n.factory.Browser.Browse(targetURL); err != nil {
					cfg.Logger.Warnf("%s", err)
				}
			}()
		}

		serverErr := craClient.Serve()
		return nil, serverErr
	})
}
