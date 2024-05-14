package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/remoteaccess"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type CmdSSH struct {
	device         []string
	externalID     []string
	externalIDType string
	listen         string
	user           string
	configuration  string

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
		`),
		Example: heredoc.Doc(`
$ c8y remoteaccess ssh --device 12345
Start an interactive SSH session on the device

$ c8y remoteaccess ssh --device 12345 --user admin
Start an interactive SSH session on the device with a given ssh user

$ c8y remoteaccess ssh --device 12345 --user admin -- systemctl status
Use a non-interactive session to execute a single command and print the result
		`),
		RunE: ccmd.RunE,
	}

	// Flags
	cmd.Flags().StringSliceVar(&ccmd.device, "device", []string{}, "Device")
	// cmd.Flags().StringSliceVar(&ccmd.externalID, "external-id", []string{}, "Device external identity")
	// cmd.Flags().StringVar(&ccmd.externalIDType, "external-type", "c8y_Serial", "Device external identity")
	cmd.Flags().StringVar(&ccmd.listen, "listen", "127.0.0.1:0", "Listener address. unix:///run/example.sock")
	cmd.Flags().StringVar(&ccmd.user, "user", "", "Default ssh user")
	cmd.Flags().StringVar(&ccmd.configuration, "configuration", "", "Remote Access Configuration")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	// cmd.MarkFlagsMutuallyExclusive("device", "external-id")
	// cmd.MarkFlagsMutuallyExclusive("device", "external-type")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
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

	// TODO: This is overly complicated. Refactor once the generic work is done
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

	out, err := body.MarshalJSON()
	if err != nil {
		return err
	}
	op := gjson.ParseBytes(out)
	device := op.Get("device").String()

	craConfig, err := c8yfetcher.DetectRemoteAccessConfiguration(client, device, n.configuration)
	if err != nil {
		return err
	}

	log.Debugf("Using remote access configuration: id=%s, name=%s", craConfig.ID, craConfig.Name)

	// Lookup configuration
	craClient := remoteaccess.NewRemoteAccessClient(client, remoteaccess.RemoteAccessOptions{
		ManagedObjectID: device,
		RemoteAccessID:  craConfig.ID,
	}, log)

	if n.listen == "-" {
		log.Debugf("Listening to request from stdin")
		return craClient.ListenServe(n.factory.IOStreams.In, n.factory.IOStreams.Out)
	}

	// TCP / socket listener
	if err := craClient.Listen(n.listen); err != nil {
		return err
	}

	localAddress := craClient.GetListenerAddress()
	host, port, _ := strings.Cut(localAddress, ":")
	if host == "" {
		host = "127.0.0.1"
	}

	// Start in background
	// TODO: Work out how to shut it down cleanly
	go craClient.Serve()

	// Build ssh command
	sshArgs := []string{
		"-o", "ServerAliveInterval=120",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}

	sshTarget := host
	if n.user != "" {
		sshTarget = fmt.Sprintf("%s@%s", n.user, host)
	}
	sshArgs = append(sshArgs, "-p", port, sshTarget)

	dashIdx := cmd.ArgsLenAtDash()
	if dashIdx > -1 {
		sshArgs = append(sshArgs, "--")
		sshArgs = append(sshArgs, args[dashIdx:]...)
	}

	sshCmd := exec.CommandContext(context.Background(), "ssh", sshArgs...)
	if n.factory.IOStreams.IsStdoutTTY() {
		sshCmd.Stdout = os.Stdout
		sshCmd.Stdin = os.Stdin
		sshCmd.Stderr = os.Stderr
	}

	log.Debugf("Executing command: ssh %s\n", strings.Join(sshArgs, " "))

	cs := n.factory.IOStreams.ColorScheme()
	fmt.Fprintln(n.factory.IOStreams.ErrOut, cs.Green(fmt.Sprintf("Starting interactive ssh session with %s (%s)\n", device, strings.TrimRight(client.BaseURL.String(), "/"))))

	start := time.Now()
	sshErr := sshCmd.Run()
	duration := time.Since(start).Truncate(time.Millisecond)
	fmt.Fprintf(n.factory.IOStreams.ErrOut, "Duration: %s\n", duration)
	return sshErr
}
