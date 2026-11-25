package connect

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/remoteaccess/rasession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

type CmdRun struct {
	device        []string
	listen        string
	configuration string
	multiplex     bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdRun(f *cmdutil.Factory) *CmdRun {
	ccmd := &CmdRun{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Connect to a device and run a custom command",
		Long: heredoc.Doc(`
			Connect to a device using the local proxy, then run a given command which makes use of the proxy.

			You can run any command you want by providing all arguments after the "--", for example:

				c8y remoteaccess connect run --device mydevice -- ./run-something.sh %h %p

			The local proxy server port and target can be accessed either by using variable references:
				* Target - Use '%h' or the environment variable '$TARGET'
				* Port - Use '%p' or the environment variable '$PORT'

			It is recommended to use the "%h" (target) "%p" (port) to refer to the local proxy server
			as this avoid potential problems with variable expansions and removes the need to have to
			escape variable etc. (e.g. \$HOST). Below shows an example using a custom ssh command:

				c8y remoteaccess connect run --device mydevice -- ssh -p %p root@%h

			Or you can use a shell to do something more complex and use environment variable references, though not the
			single quotes around the shell command to prevent the variable expansion in your current shell.

				c8y remoteaccess connect run --device mydevice -- sh -c 'ssh -p $PORT root@$TARGET'
		`),
		Example: heredoc.Doc(`
			$ c8y remoteaccess connect run --device 12345 -- ssh -p %p root@%h
			Start an interactive SSH session on the device with a given ssh user

			$ c8y remoteaccess connect run --device rpi5-abcdef01 --configuration passthrough -- ssh -p %p -L 1885:127.0.0.1:1883 -o ServerAliveInterval=120 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@127.0.0.1
			Start an SSH session to setup port-forwarding to map the remote's 127.0.0.1:1883 port to your machine's 1885 port

			$ c8y remoteaccess connect run --device 12345 --configuration "http:*" --multiplex -- curl -s http://%h:%p/
			Run a command against a web server on the device, carrying all its connections over a single remote access session (if supported by the device)
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listener address. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration")
	cmd.Flags().BoolVar(&ccmd.multiplex, "multiplex", false, "Carry all local connections over a single remote access session (yamux), which avoids the setup cost of a new session per connection. Requires a device supporting it (e.g. thin-edge.io), otherwise one session per connection is used. Enabled by default for configurations named with the mux option, e.g. \"http+mux:example\"")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", ccmd.factory.ClientV2),
		completion.WithRemoteAccessPassthroughConfiguration("configuration", "device", ccmd.factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE starts a local proxy to each device's remote access passthrough
// configuration, and runs the command given after "--" against it.
func (n *CmdRun) RunE(cmd *cobra.Command, args []string) error {
	// Build the command before connecting, so a missing one fails fast
	run := ""
	var runArgs []string
	if dashIdx := cmd.ArgsLenAtDash(); dashIdx > -1 && dashIdx < len(args) {
		run = args[dashIdx]
		runArgs = args[dashIdx+1:]
	}
	if run == "" {
		return cmderrors.NewUserError("Missing run command")
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

			// TCP / socket listener, served in the background
			if err := session.Listen(n.listen); err != nil {
				return err
			}
			go session.Serve()
			host, port := session.HostPort()

			// Allow users to access the port via variables (similar to what ssh does)
			expandedArgs := make([]string, 0, len(runArgs))
			for _, v := range runArgs {
				v = strings.ReplaceAll(v, "%h", host)
				v = strings.ReplaceAll(v, "%p", port)
				expandedArgs = append(expandedArgs, v)
			}

			runCmd := exec.CommandContext(ctx, run, expandedArgs...)
			// Add the target and port to the environment, so they can be used by
			// more complex scripts. WSLENV exposes them to WSL
			runCmd.Env = append(os.Environ(),
				"PORT="+port,
				"TARGET="+host,
				"DEVICE="+deviceID,
				"WSLENV=PORT/u:TARGET/u:DEVICE/u:C8Y_HOST/u",
			)
			runCmd.Stdout = n.factory.IOStreams.Out
			runCmd.Stdin = n.factory.IOStreams.In
			runCmd.Stderr = n.factory.IOStreams.ErrOut

			log.Infof("Executing command: %s %s\n", run, strings.Join(expandedArgs, " "))

			cs := n.factory.IOStreams.ColorScheme()
			fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting external command on %s (%s)\n", deviceID, rasession.CumulocityURL(client))))

			start := time.Now()
			runErr := runCmd.Run()
			fmt.Fprintf(n.factory.IOStreams.ErrOut, "Duration: %s\n", time.Since(start).Truncate(time.Millisecond))

			// Use the exit code of the command
			if runCmd.ProcessState != nil && runCmd.ProcessState.ExitCode() != 0 {
				return cmderrors.NewErrorWithExitCode(cmderrors.ExitCode(runCmd.ProcessState.ExitCode()), runErr)
			}
			return runErr
		}, nil
	})
}
