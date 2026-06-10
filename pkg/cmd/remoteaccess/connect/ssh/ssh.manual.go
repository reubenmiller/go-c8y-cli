package ssh

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
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

type CmdSSH struct {
	device         []string
	listen         string
	user           string
	configuration  string
	portForwarding string
	preferredAuth  string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdSSH(f *cmdutil.Factory) *CmdSSH {
	ccmd := &CmdSSH{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Connect to a device via ssh",
		Long: heredoc.Doc(`
			Connect to a device via ssh

			Additional arguments can be passed to the ssh shell by using the "--" convention where everything
			after the "--" will be passed untouched to the ssh shell. In this mode, the shell will not be
			interactive, and it will return upon completion of the command.

			You can set the default ssh user to use for all ssh connections for your current c8y session file
			using:

				c8y settings update remoteaccess.sshuser root

		`),
		Example: heredoc.Doc(`
			$ c8y remoteaccess connect ssh --device 12345
			Start an interactive SSH session on the device

			$ c8y remoteaccess connect ssh --device 12345 --user admin
			Start an interactive SSH session on the device with a given ssh user

			$ c8y remoteaccess connect ssh --device 12345 --user admin --preferred-auth password
			Start an interactive SSH session on the device with a given ssh user and force password authentication

			$ c8y remoteaccess connect ssh --device 12345 --user admin -L 1883:127.0.0.1:1883
			Start an interactive SSH session and configure port-forward by mapping the remote's 127.0.0.1:1883 to your machine's port 1883

			$ c8y remoteaccess connect ssh --device 12345 --user admin -L 1883
			Start an interactive SSH session and configure port-forward: 127.0.0.11883 (remote) => 1883 (local)

			$ c8y remoteaccess connect ssh --device 12345 --user admin -L 1884:1883
			Start an interactive SSH session and configure port-forward: 127.0.0.11883 (remote) => 1884 (local)

			$ c8y remoteaccess connect ssh --device 12345 --user admin -- systemctl status
			Use a non-interactive session to execute a single command and print the result

			$ c8y remoteaccess connect ssh --device 12345 --user admin -- "sh -c 'cat /etc/os-release'"
			use a non-interactive session to execute a custom shell command (notice the surrounding double quotes on the command!)
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listener address. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.user, "user", "", "Default ssh user")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration")
	cmd.Flags().StringVar(&ccmd.preferredAuth, "preferred-auth", "", "Set the preferred authentication for the ssh connection. This will add '-o PreferredAuthentications=<value>' to the ssh command")
	cmd.Flags().StringVarP(&ccmd.portForwarding, "port-forward", "L", "", "SSH Port-Forwarding option in the format [bind_address:]port:host:hostport. It also accepts a custom short form, <local>[:<remote>]. The value is passed to the ssh -L option, so read the ssh man page for more info")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithRemoteAccessPassthroughConfiguration("configuration", "device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("preferred-auth", "password", "publickey"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdSSH) GetPortForwarding() string {
	// convenience functions to mirror docker port mapping options
	portMapping := n.portForwarding

	if portMapping == "" {
		return portMapping
	}

	portForwardingParts := strings.Split(portMapping, ":")
	switch len(portForwardingParts) {
	case 1:
		// -L 1883 => -L 1883:127.0.0.1:1883
		portMapping = fmt.Sprintf("%s:127.0.0.1:%s", portForwardingParts[0], portForwardingParts[0])
	case 2:
		// -L 1884:1883 => -L 1884:127.0.0.1:1883
		portMapping = fmt.Sprintf("%s:127.0.0.1:%s", portForwardingParts[0], portForwardingParts[1])
	}
	return portMapping
}
func (n *CmdSSH) RunE(cmd *cobra.Command, args []string) error {
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

		// TCP / socket listener
		if err := craClient.Listen(n.listen); err != nil {
			return nil, err
		}

		localAddress := craClient.GetListenerAddress()
		host, port, _ := strings.Cut(localAddress, ":")
		if host == "" {
			host = "127.0.0.1"
		}

		// Start in background
		go craClient.Serve()

		// Build ssh command
		sshArgs := []string{
			"-o", "ServerAliveInterval=120",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
		}

		if n.preferredAuth != "" {
			sshArgs = append(sshArgs, "-o", fmt.Sprintf("PreferredAuthentications=%s", n.preferredAuth))
		}

		sshTarget := host
		if n.user == "" {
			// Use default user (if set)
			n.user = cfg.GetRemoteAccessDefaultSSHUser()
		}
		if n.user != "" {
			sshTarget = fmt.Sprintf("%s@%s", n.user, host)
		}
		sshArgs = append(sshArgs, "-p", port, sshTarget)

		portForwarding := n.GetPortForwarding()
		if portForwarding != "" {
			sshArgs = append(sshArgs, "-L", portForwarding)
		}

		dashIdx := cmd.ArgsLenAtDash()
		if dashIdx > -1 {
			sshArgs = append(sshArgs, "--")
			sshArgs = append(sshArgs, args[dashIdx:]...)
		}

		sshCmd := exec.CommandContext(context.Background(), "ssh", sshArgs...)
		sshCmd.Stdout = n.factory.IOStreams.Out
		sshCmd.Stdin = n.factory.IOStreams.In
		sshCmd.Stderr = n.factory.IOStreams.ErrOut
		log.Infof("Executing command: ssh %s\n", strings.Join(sshArgs, " "))

		cs := n.factory.IOStreams.ColorScheme()
		if portForwarding != "" {
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting interactive ssh session with %s (%s) with port-forwarding %s\n", device, strings.TrimRight(client.BaseURL.String(), "/"), portForwarding)))
		} else {
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting interactive ssh session with %s (%s)\n", device, strings.TrimRight(client.BaseURL.String(), "/"))))
		}

		start := time.Now()
		sshErr := sshCmd.Run()
		duration := time.Since(start).Truncate(time.Millisecond)
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "Duration: %s\n", duration)

		// Use exit code from the command
		if sshCmd.ProcessState != nil {
			if sshCmd.ProcessState.ExitCode() != 0 {
				return nil, cmderrors.NewErrorWithExitCode(
					cmderrors.ExitCode(sshCmd.ProcessState.ExitCode()),
					sshErr,
				)
			}
		}

		return nil, sshErr
	}, nil)
}
