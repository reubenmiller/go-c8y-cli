package run

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

type CmdRun struct {
	device        []string
	listen        string
	configuration string

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
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listener address. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration")

	completion.WithOptions(
		cmd,
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

func (n *CmdRun) RunE(cmd *cobra.Command, args []string) error {
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
		run := ""
		runArgs := []string{}

		dashIdx := cmd.ArgsLenAtDash()
		if dashIdx > -1 {
			run = args[dashIdx]
			if dashIdx+1 < len(args) {
				// Allow users to access port via variables (similar to what ssh does)
				for _, v := range args[dashIdx+1:] {
					expandedValue := v
					expandedValue = strings.ReplaceAll(expandedValue, "%h", host)
					expandedValue = strings.ReplaceAll(expandedValue, "%p", port)
					runArgs = append(runArgs, expandedValue)
				}
			}
		}

		if run == "" {
			return nil, cmderrors.NewUserError("Missing run command")
		}

		runCmd := exec.CommandContext(context.Background(), run, runArgs...)

		// Add target and port to the environment variables so it can be easily accessed from more
		// complex scripts
		runCmd.Env = append(runCmd.Env, fmt.Sprintf("PORT=%s", port))
		runCmd.Env = append(runCmd.Env, fmt.Sprintf("TARGET=%s", host))
		runCmd.Env = append(runCmd.Env, fmt.Sprintf("DEVICE=%s", device))

		// Support WSL environments and expose variables to WSL
		runCmd.Env = append(runCmd.Env, "WSLENV=PORT/u:TARGET/u:DEVICE/u:C8Y_HOST/u")

		runCmd.Stdout = n.factory.IOStreams.Out
		runCmd.Stdin = n.factory.IOStreams.In
		runCmd.Stderr = n.factory.IOStreams.ErrOut

		log.Infof("Executing command: %s %s\n", run, strings.Join(runArgs, " "))

		cs := n.factory.IOStreams.ColorScheme()
		fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting external command on %s (%s)\n", device, strings.TrimRight(client.BaseURL.String(), "/"))))

		start := time.Now()
		sshErr := runCmd.Run()
		duration := time.Since(start).Truncate(time.Millisecond)
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "Duration: %s\n", duration)

		return nil, sshErr
	})
}
