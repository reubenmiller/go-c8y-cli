package ssh

import (
	"context"
	"fmt"
	"io"
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
	batch          bool
	fallbackUser   string

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

			If no user is given (via --user or the setting), ssh uses your local user name. Use
			--fallback-user (or the remoteaccess.sshFallbackUser setting) to try another user, e.g. root,
			if the authentication with the local user name fails:

				c8y settings update remoteaccess.sshFallbackUser root

			Use --batch when running ssh from a script or an AI agent. ssh then never prompts for a
			password or passphrase, only key-based authentication is used, and the connection fails
			instead of waiting for input.

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

			$ c8y remoteaccess connect ssh --device 12345 --batch -- systemctl status
			Execute a single command using key-based authentication only, failing instead of prompting for a password

			$ c8y remoteaccess connect ssh --device 12345 --fallback-user root -- systemctl status
			Execute a single command with the local user name, or as root if the authentication fails
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listener address. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.user, "user", "", "Default ssh user")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration")
	cmd.Flags().StringVar(&ccmd.preferredAuth, "preferred-auth", "", "Set the preferred authentication for the ssh connection. This will add '-o PreferredAuthentications=<value>' to the ssh command")
	cmd.Flags().BoolVar(&ccmd.batch, "batch", false, "Never prompt for a password or passphrase. Only key-based authentication is used, and the connection fails instead of waiting for input")
	cmd.Flags().StringVar(&ccmd.fallbackUser, "fallback-user", "", "ssh user to try if no user is given and the authentication with the local user name fails, e.g. root. Defaults to the remoteaccess.sshFallbackUser setting")
	cmd.Flags().StringVarP(&ccmd.portForwarding, "port-forward", "L", "", "SSH Port-Forwarding option in the format [bind_address:]port:host:hostport. It also accepts a custom short form, <local>[:<remote>]. The value is passed to the ssh -L option, so read the ssh man page for more info")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithRemoteAccessPassthroughConfiguration("configuration", "device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("preferred-auth", "password", "publickey"),
		completion.WithValidateSet("fallback-user", "root"),
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

		if n.batch && n.preferredAuth != "" && n.preferredAuth != "publickey" {
			return nil, cmderrors.NewUserError(fmt.Sprintf("--preferred-auth %s can't be used with --batch, as batch mode only uses key-based authentication", n.preferredAuth))
		}

		user := n.user
		if user == "" {
			// Use default user (if set)
			user = cfg.GetRemoteAccessDefaultSSHUser()
		}
		fallbackUser := n.fallbackUser
		if !cmd.Flags().Changed("fallback-user") {
			fallbackUser = cfg.GetRemoteAccessSSHFallbackUser()
		}
		users := SSHUsers(user, fallbackUser)

		portForwarding := n.GetPortForwarding()
		var command []string
		if dashIdx := cmd.ArgsLenAtDash(); dashIdx > -1 {
			command = args[dashIdx:]
		}

		cs := n.factory.IOStreams.ColorScheme()
		if portForwarding != "" {
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting interactive ssh session with %s (%s) with port-forwarding %s\n", device, strings.TrimRight(client.BaseURL.String(), "/"), portForwarding)))
		} else {
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting interactive ssh session with %s (%s)\n", device, strings.TrimRight(client.BaseURL.String(), "/"))))
		}

		exitCode, sshErr := ConnectWithFallback(users, func(u string) (int, string, error) {
			sshArgs := BuildSSHArgs(SSHOptions{
				Host:           host,
				Port:           port,
				User:           u,
				PreferredAuth:  n.preferredAuth,
				Batch:          n.batch,
				PortForwarding: portForwarding,
				Command:        command,
			})

			stderr := &tailWriter{max: 4096}
			sshCmd := exec.CommandContext(context.Background(), "ssh", sshArgs...)
			sshCmd.Stdout = n.factory.IOStreams.Out
			sshCmd.Stdin = n.factory.IOStreams.In
			sshCmd.Stderr = io.MultiWriter(n.factory.IOStreams.ErrOut, stderr)
			log.Infof("Executing command: ssh %s\n", strings.Join(sshArgs, " "))

			start := time.Now()
			err := sshCmd.Run()
			duration := time.Since(start).Truncate(time.Millisecond)
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "Duration: %s\n", duration)

			exitCode := 0
			if sshCmd.ProcessState != nil {
				exitCode = sshCmd.ProcessState.ExitCode()
			}
			return exitCode, stderr.String(), err
		}, func(next string) {
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Yellow(fmt.Sprintf("Authentication failed with the default user. Retrying with the %q user", next)))
		})

		// Use exit code from the command
		if exitCode != 0 {
			return nil, cmderrors.NewErrorWithExitCode(
				cmderrors.ExitCode(exitCode),
				sshErr,
			)
		}

		return nil, sshErr
	}, nil)
}

// ConnectWithFallback runs connect for each user until it succeeds, or fails for
// another reason than the authentication. It returns the result of the last attempt
func ConnectWithFallback(users []string, connect func(user string) (int, string, error), onRetry func(next string)) (int, error) {
	var exitCode int
	var err error
	for i, u := range users {
		var stderr string
		exitCode, stderr, err = connect(u)
		if i < len(users)-1 && IsAuthFailure(exitCode, stderr) {
			if onRetry != nil {
				onRetry(users[i+1])
			}
			continue
		}
		break
	}
	return exitCode, err
}

// SSHUsers returns the users to try in order. The fallback user is only tried if no
// user was given, as ssh then uses the local user name, which often doesn't exist on
// the device
func SSHUsers(user, fallbackUser string) []string {
	users := []string{user}
	if user == "" && fallbackUser != "" {
		users = append(users, fallbackUser)
	}
	return users
}

// SSHOptions options to build the ssh command
type SSHOptions struct {
	Host           string
	Port           string
	User           string
	PreferredAuth  string
	Batch          bool
	PortForwarding string
	Command        []string
}

// BuildSSHArgs returns the arguments for the ssh command
func BuildSSHArgs(opts SSHOptions) []string {
	sshArgs := []string{
		"-o", "ServerAliveInterval=120",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}

	if opts.Batch {
		// never prompt for a password or passphrase, and don't wait forever
		sshArgs = append(sshArgs,
			"-o", "BatchMode=yes",
			"-o", "PreferredAuthentications=publickey",
			"-o", "ConnectTimeout=30",
		)
	} else if opts.PreferredAuth != "" {
		sshArgs = append(sshArgs, "-o", fmt.Sprintf("PreferredAuthentications=%s", opts.PreferredAuth))
	}

	sshTarget := opts.Host
	if opts.User != "" {
		sshTarget = fmt.Sprintf("%s@%s", opts.User, opts.Host)
	}
	sshArgs = append(sshArgs, "-p", opts.Port, sshTarget)

	if opts.PortForwarding != "" {
		sshArgs = append(sshArgs, "-L", opts.PortForwarding)
	}

	if len(opts.Command) > 0 {
		sshArgs = append(sshArgs, "--")
		sshArgs = append(sshArgs, opts.Command...)
	}
	return sshArgs
}

// IsAuthFailure returns true if ssh failed to authenticate. ssh uses the exit
// code 255 for its own errors, and prints "Permission denied (<methods>)" when
// all authentication methods failed
func IsAuthFailure(exitCode int, stderr string) bool {
	return exitCode == 255 && strings.Contains(stderr, "Permission denied (")
}

// tailWriter keeps the last bytes written to it
type tailWriter struct {
	max int
	buf []byte
}

func (w *tailWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	if len(w.buf) > w.max {
		w.buf = w.buf[len(w.buf)-w.max:]
	}
	return len(p), nil
}

func (w *tailWriter) String() string {
	return string(w.buf)
}
